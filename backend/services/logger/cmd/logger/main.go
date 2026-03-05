package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/example/appfoundrylab/backend/pkg/env"
	"github.com/example/appfoundrylab/backend/services/logger/internal/incidents"
	"github.com/example/appfoundrylab/backend/services/logger/internal/ingest"
	"github.com/example/appfoundrylab/backend/services/logger/internal/queue"
	"github.com/example/appfoundrylab/backend/services/logger/internal/requestlogs"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	allowUnsignedIngest := env.GetWithDefault("LOGGER_ALLOW_UNSIGNED_INGEST", "false") == "true"
	ingestSecret := os.Getenv("LOGGER_SHARED_SECRET")
	if !allowUnsignedIngest && ingestSecret == "" {
		log.Fatal("LOGGER_SHARED_SECRET is required when LOGGER_ALLOW_UNSIGNED_INGEST=false")
	}

	queueSize := env.GetIntWithDefault("LOGGER_QUEUE_SIZE", 2048)
	workerCount := env.GetIntWithDefault("LOGGER_WORKERS", 4)
	retryMax := env.GetIntWithDefault("LOGGER_RETRY_MAX", 1)
	dropAlertThresholdPct := float64(env.GetIntWithDefault("LOGGER_DROP_ALERT_THRESHOLD_PCT", 5))
	q := queue.New(queueSize, workerCount, retryMax)
	q.SetDropAlertThresholdPct(dropAlertThresholdPct)
	q.StartWorkers(ctx)

	r := chi.NewRouter()
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(2 * time.Second))

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	r.Get("/metrics", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(q.Stats())
	})
	r.Get("/metrics/prometheus", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		_, _ = w.Write([]byte(q.PrometheusMetrics()))
	})
	r.Get("/incident-events", func(w http.ResponseWriter, r *http.Request) {
		limit := parseListLimit(r.URL.Query().Get("limit"), 20, 100)

		queryCtx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		events, err := incidents.ListRecent(queryCtx, limit)
		if err != nil {
			http.Error(w, "failed to load incident events", http.StatusServiceUnavailable)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"items": events})
	})
	r.Get("/incident-events/summary", func(w http.ResponseWriter, r *http.Request) {
		queryCtx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		summary, err := incidents.Summarize(queryCtx)
		if err != nil {
			http.Error(w, "failed to load incident summary", http.StatusServiceUnavailable)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(summary)
	})
	r.Get("/request-logs", func(w http.ResponseWriter, r *http.Request) {
		limit := parseListLimit(r.URL.Query().Get("limit"), 20, 100)

		queryCtx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		items, err := requestlogs.ListRecent(queryCtx, limit, r.URL.Query().Get("traceId"))
		if err != nil {
			http.Error(w, "failed to load request logs", http.StatusServiceUnavailable)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"items": items})
	})

	r.Post("/ingest", func(w http.ResponseWriter, r *http.Request) {
		maxBody := int64(1 << 20)
		if v := os.Getenv("MAX_REQUEST_BODY_BYTES"); v != "" {
			if parsed, err := strconv.ParseInt(v, 10, 64); err == nil && parsed > 0 {
				maxBody = parsed
			}
		}
		r.Body = http.MaxBytesReader(w, r.Body, maxBody)
		if err := verifyIngestAuth(r, ingestSecret, allowUnsignedIngest); err != nil {
			http.Error(w, "unauthorized ingest request", http.StatusUnauthorized)
			return
		}

		var payload ingest.RequestLog
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		if ok := q.Enqueue(payload); !ok {
			w.WriteHeader(http.StatusAccepted)
			_, _ = w.Write([]byte(`{"status":"dropped"}`))
			return
		}
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"status":"queued"}`))
	})
	r.Post("/incident-events", func(w http.ResponseWriter, r *http.Request) {
		maxBody := int64(1 << 20)
		r.Body = http.MaxBytesReader(w, r.Body, maxBody)
		if err := verifyIngestAuth(r, ingestSecret, allowUnsignedIngest); err != nil {
			http.Error(w, "unauthorized incident request", http.StatusUnauthorized)
			return
		}

		var payload incidents.Event
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}

		insertCtx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := incidents.Insert(insertCtx, payload); err != nil {
			http.Error(w, "failed to persist incident event", http.StatusServiceUnavailable)
			return
		}

		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"status":"stored"}`))
	})

	addr := ":" + env.GetWithDefault("LOGGER_PORT", "8090")
	server := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	stopCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		<-stopCtx.Done()
		shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelShutdown()
		cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	log.Printf("logger service listening on %s", addr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
	q.Wait()
}

func verifyIngestAuth(r *http.Request, secret string, allowUnsignedIngest bool) error {
	if secret == "" && allowUnsignedIngest {
		return nil
	}
	if secret == "" {
		return errors.New("missing LOGGER_SHARED_SECRET")
	}

	timestamp := r.Header.Get("X-Logger-Timestamp")
	signature := r.Header.Get("X-Logger-Signature")
	if timestamp == "" || signature == "" {
		return errors.New("missing auth headers")
	}

	parsedTimestamp, err := time.Parse(time.RFC3339Nano, timestamp)
	if err != nil {
		return errors.New("invalid timestamp")
	}

	now := time.Now().UTC()
	if parsedTimestamp.Before(now.Add(-5*time.Minute)) || parsedTimestamp.After(now.Add(30*time.Second)) {
		return errors.New("timestamp outside replay window")
	}

	bodyBytes, err := readBody(r)
	if err != nil {
		return err
	}
	expected := signLogPayload(secret, timestamp, bodyBytes)
	if !hmac.Equal([]byte(signature), []byte(expected)) {
		return errors.New("signature mismatch")
	}

	r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	return nil
}

func readBody(r *http.Request) ([]byte, error) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	return bodyBytes, nil
}

func signLogPayload(secret, timestamp string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp))
	mac.Write([]byte("\n"))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

func parseListLimit(raw string, defaultLimit, maxLimit int64) int64 {
	if raw == "" {
		return defaultLimit
	}

	parsed, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || parsed <= 0 {
		return defaultLimit
	}
	if parsed > maxLimit {
		return maxLimit
	}
	return parsed
}

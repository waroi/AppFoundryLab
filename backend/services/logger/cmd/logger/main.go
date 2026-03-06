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
	"github.com/example/appfoundrylab/backend/pkg/runtimeknobs"
	"github.com/example/appfoundrylab/backend/services/logger/internal/incidents"
	"github.com/example/appfoundrylab/backend/services/logger/internal/ingest"
	"github.com/example/appfoundrylab/backend/services/logger/internal/mongo"
	"github.com/example/appfoundrylab/backend/services/logger/internal/queue"
	"github.com/example/appfoundrylab/backend/services/logger/internal/requestlogs"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

const (
	loggerHandlerTimeout        = 2 * time.Second
	loggerQueryTimeout          = 2 * time.Second
	loggerMaxBodyBytes    int64 = 1 << 20
	loggerShutdownTimeout       = 5 * time.Second
	loggerReadTimeout           = 5 * time.Second
	loggerWriteTimeout          = 5 * time.Second
	loggerIdleTimeout           = 30 * time.Second
)

type jsonErrorEnvelope struct {
	Error jsonErrorBody `json:"error"`
}

type jsonErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

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
	retryBackoffBase := env.GetIntWithDefault("LOGGER_RETRY_BACKOFF_BASE_MS", 100)
	retryBackoffMax := env.GetIntWithDefault("LOGGER_RETRY_BACKOFF_MAX_MS", 1000)
	dropAlertThresholdPct := float64(env.GetIntWithDefault("LOGGER_DROP_ALERT_THRESHOLD_PCT", 5))
	q := queue.New(queueSize, workerCount, retryMax)
	q.SetRetryBackoff(time.Duration(retryBackoffBase)*time.Millisecond, time.Duration(retryBackoffMax)*time.Millisecond)
	q.SetDropAlertThresholdPct(dropAlertThresholdPct)
	q.StartWorkers(ctx)

	r := chi.NewRouter()
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(loggerHandlerTimeout))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		healthCtx, cancel := context.WithTimeout(r.Context(), healthCheckTimeout())
		defer cancel()

		httpStatus, payload := loggerHealthPayload(mongo.Health(healthCtx))
		writeJSON(w, httpStatus, payload)
	})
	r.Get("/metrics", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, q.Stats())
	})
	r.Get("/metrics/prometheus", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		if _, err := w.Write([]byte(q.PrometheusMetrics())); err != nil {
			log.Printf("logger service failed to write prometheus metrics: %v", err)
		}
	})
	r.Get("/incident-events", func(w http.ResponseWriter, r *http.Request) {
		limit := parseListLimit(r.URL.Query().Get("limit"), 20, 100)

		queryCtx, cancel := context.WithTimeout(r.Context(), loggerQueryTimeout)
		defer cancel()

		events, err := incidents.ListRecent(queryCtx, limit)
		if err != nil {
			writeJSONError(w, http.StatusServiceUnavailable, "incident_events_unavailable", "failed to load incident events")
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{"items": events})
	})
	r.Get("/incident-events/summary", func(w http.ResponseWriter, r *http.Request) {
		queryCtx, cancel := context.WithTimeout(r.Context(), loggerQueryTimeout)
		defer cancel()

		summary, err := incidents.Summarize(queryCtx)
		if err != nil {
			writeJSONError(w, http.StatusServiceUnavailable, "incident_summary_unavailable", "failed to load incident summary")
			return
		}

		writeJSON(w, http.StatusOK, summary)
	})
	r.Get("/request-logs", func(w http.ResponseWriter, r *http.Request) {
		limit := parseListLimit(r.URL.Query().Get("limit"), 20, 100)

		queryCtx, cancel := context.WithTimeout(r.Context(), loggerQueryTimeout)
		defer cancel()

		items, err := requestlogs.ListRecent(queryCtx, limit, r.URL.Query().Get("traceId"))
		if err != nil {
			writeJSONError(w, http.StatusServiceUnavailable, "request_logs_unavailable", "failed to load request logs")
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{"items": items})
	})

	r.Post("/ingest", func(w http.ResponseWriter, r *http.Request) {
		maxBody := loggerMaxBodyBytes
		if v := os.Getenv("MAX_REQUEST_BODY_BYTES"); v != "" {
			if parsed, err := strconv.ParseInt(v, 10, 64); err == nil && parsed > 0 {
				maxBody = parsed
			}
		}
		r.Body = http.MaxBytesReader(w, r.Body, maxBody)
		if err := verifyIngestAuth(r, ingestSecret, allowUnsignedIngest); err != nil {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized_ingest_request", "unauthorized ingest request")
			return
		}

		var payload ingest.RequestLog
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid_json", "invalid json")
			return
		}
		if ok := q.Enqueue(payload); !ok {
			writeJSON(w, http.StatusAccepted, map[string]string{"status": "dropped"})
			return
		}
		writeJSON(w, http.StatusAccepted, map[string]string{"status": "queued"})
	})
	r.Post("/incident-events", func(w http.ResponseWriter, r *http.Request) {
		maxBody := loggerMaxBodyBytes
		r.Body = http.MaxBytesReader(w, r.Body, maxBody)
		if err := verifyIngestAuth(r, ingestSecret, allowUnsignedIngest); err != nil {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized_incident_request", "unauthorized incident request")
			return
		}

		var payload incidents.Event
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid_json", "invalid json")
			return
		}

		insertCtx, cancel := context.WithTimeout(r.Context(), loggerQueryTimeout)
		defer cancel()
		if err := incidents.Insert(insertCtx, payload); err != nil {
			writeJSONError(w, http.StatusServiceUnavailable, "incident_event_persist_failed", "failed to persist incident event")
			return
		}

		writeJSON(w, http.StatusAccepted, map[string]string{"status": "stored"})
	})

	addr := ":" + env.GetWithDefault("LOGGER_PORT", "8090")
	server := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  loggerReadTimeout,
		WriteTimeout: loggerWriteTimeout,
		IdleTimeout:  loggerIdleTimeout,
	}

	stopCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		<-stopCtx.Done()
		shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), loggerShutdownTimeout)
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

func loggerHealthPayload(err error) (int, map[string]any) {
	payload := map[string]any{
		"status": "ok",
		"checks": map[string]string{"mongo": "up"},
	}
	if err == nil {
		return http.StatusOK, payload
	}

	payload["status"] = "degraded"
	payload["checks"] = map[string]string{"mongo": "down"}
	payload["lastError"] = err.Error()
	return http.StatusServiceUnavailable, payload
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
	if parsedTimestamp.Before(now.Add(-ingestReplayMaxAge())) || parsedTimestamp.After(now.Add(ingestReplayMaxFutureSkew())) {
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

func healthCheckTimeout() time.Duration {
	return runtimeknobs.LoggerHealthTimeout()
}

func ingestReplayMaxAge() time.Duration {
	return runtimeknobs.LoggerIngestTimestampMaxAge()
}

func ingestReplayMaxFutureSkew() time.Duration {
	return runtimeknobs.LoggerIngestTimestampMaxFutureSkew()
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("logger service failed to encode json response: %v", err)
	}
}

func writeJSONError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, jsonErrorEnvelope{
		Error: jsonErrorBody{
			Code:    code,
			Message: message,
		},
	})
}

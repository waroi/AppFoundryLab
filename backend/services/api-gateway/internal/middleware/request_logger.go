package middleware

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/example/appfoundrylab/backend/pkg/env"
	"github.com/example/appfoundrylab/backend/services/api-gateway/pkg/httpx"
)

type RequestLog struct {
	Path       string `json:"path"`
	Method     string `json:"method"`
	IP         string `json:"ip"`
	TraceID    string `json:"traceId"`
	DurationMS int64  `json:"durationMs"`
	StatusCode int    `json:"statusCode"`
	OccurredAt string `json:"occurredAt"`
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

type asyncLogSender struct {
	endpoint    string
	secret      string
	client      *http.Client
	workers     int
	retryMax    int
	backoffBase time.Duration
	backoffMax  time.Duration

	queue chan RequestLog

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	dropped atomic.Uint64
}

type AsyncLogSenderStats struct {
	Enabled       bool   `json:"enabled"`
	Endpoint      string `json:"endpoint"`
	QueueDepth    int    `json:"queueDepth"`
	QueueCapacity int    `json:"queueCapacity"`
	Workers       int    `json:"workers"`
	RetryMax      int    `json:"retryMax"`
	DroppedTotal  uint64 `json:"droppedTotal"`
}

var currentAsyncLogSender atomic.Pointer[asyncLogSender]

func newAsyncLogSenderFromEnv() *asyncLogSender {
	s := newAsyncLogSender(
		os.Getenv("LOGGER_ENDPOINT"),
		os.Getenv("LOGGER_SHARED_SECRET"),
		env.GetIntWithDefault("LOGGER_QUEUE_SIZE", 2048),
		env.GetIntWithDefault("LOGGER_WORKERS", 4),
		env.GetIntWithDefault("LOGGER_RETRY_MAX", 1),
		&http.Client{Timeout: 800 * time.Millisecond},
	)
	if s == nil {
		currentAsyncLogSender.Store(nil)
		return nil
	}

	backoffBase := env.GetIntWithDefault("LOGGER_RETRY_BACKOFF_BASE_MS", 50)
	backoffMax := env.GetIntWithDefault("LOGGER_RETRY_BACKOFF_MAX_MS", 500)
	if backoffBase <= 0 {
		backoffBase = 50
	}
	if backoffMax <= 0 {
		backoffMax = 500
	}
	if backoffMax < backoffBase {
		backoffMax = backoffBase
	}
	s.backoffBase = time.Duration(backoffBase) * time.Millisecond
	s.backoffMax = time.Duration(backoffMax) * time.Millisecond
	currentAsyncLogSender.Store(s)
	return s
}

func newAsyncLogSender(endpoint, secret string, queueSize, workers, retryMax int, client *http.Client) *asyncLogSender {
	if endpoint == "" {
		return nil
	}
	if queueSize <= 0 {
		queueSize = 2048
	}
	if workers <= 0 {
		workers = 4
	}
	if retryMax < 0 {
		retryMax = 0
	}
	if client == nil {
		client = &http.Client{Timeout: 800 * time.Millisecond}
	}

	ctx, cancel := context.WithCancel(context.Background())
	s := &asyncLogSender{
		endpoint:    endpoint,
		secret:      secret,
		client:      client,
		workers:     workers,
		retryMax:    retryMax,
		backoffBase: 50 * time.Millisecond,
		backoffMax:  500 * time.Millisecond,
		queue:       make(chan RequestLog, queueSize),
		ctx:         ctx,
		cancel:      cancel,
	}

	for i := 0; i < workers; i++ {
		s.wg.Add(1)
		go s.worker()
	}
	return s
}

func (s *asyncLogSender) Close() {
	s.cancel()
	s.wg.Wait()
}

func (s *asyncLogSender) Enqueue(entry RequestLog) {
	select {
	case s.queue <- entry:
	default:
		dropped := s.dropped.Add(1)
		if dropped%100 == 0 {
			log.Printf("gateway async logger queue drops=%d", dropped)
		}
	}
}

func (s *asyncLogSender) DroppedTotal() uint64 {
	return s.dropped.Load()
}

func (s *asyncLogSender) Stats() AsyncLogSenderStats {
	if s == nil {
		return AsyncLogSenderStats{}
	}
	return AsyncLogSenderStats{
		Enabled:       true,
		Endpoint:      s.endpoint,
		QueueDepth:    len(s.queue),
		QueueCapacity: cap(s.queue),
		Workers:       s.workers,
		RetryMax:      s.retryMax,
		DroppedTotal:  s.dropped.Load(),
	}
}

func CurrentAsyncLogSenderStats() AsyncLogSenderStats {
	return currentAsyncLogSender.Load().Stats()
}

func (s *asyncLogSender) worker() {
	defer s.wg.Done()

	for {
		select {
		case <-s.ctx.Done():
			return
		case entry := <-s.queue:
			s.sendWithRetry(entry)
		}
	}
}

func (s *asyncLogSender) sendWithRetry(entry RequestLog) {
	for attempt := 0; attempt <= s.retryMax; attempt++ {
		if err := s.sendOnce(entry); err == nil {
			return
		}
		if attempt < s.retryMax {
			wait := s.retryBackoffForAttempt(attempt + 1)
			timer := time.NewTimer(wait)
			select {
			case <-s.ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
			}
		}
	}
}

func (s *asyncLogSender) retryBackoffForAttempt(retryAttempt int) time.Duration {
	if retryAttempt <= 0 {
		return 0
	}

	delay := s.backoffBase
	for i := 1; i < retryAttempt; i++ {
		delay *= 2
		if delay >= s.backoffMax {
			return s.backoffMax
		}
	}
	if delay > s.backoffMax {
		return s.backoffMax
	}
	return delay
}

func (s *asyncLogSender) sendOnce(logEntry RequestLog) error {
	body, err := json.Marshal(logEntry)
	if err != nil {
		return err
	}

	timestamp := time.Now().UTC().Format(time.RFC3339Nano)
	req, err := http.NewRequestWithContext(s.ctx, http.MethodPost, s.endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if logEntry.TraceID != "" {
		req.Header.Set(httpx.TraceIDHeader, logEntry.TraceID)
	}
	if s.secret != "" {
		req.Header.Set("X-Logger-Timestamp", timestamp)
		req.Header.Set("X-Logger-Signature", signLogPayload(s.secret, timestamp, body))
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	_ = resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("logger ingest non-2xx status: %d", resp.StatusCode)
	}
	return nil
}

func AsyncRequestLogger(next http.Handler) http.Handler {
	sender := newAsyncLogSenderFromEnv()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(recorder, r)

		if sender == nil {
			return
		}

		entry := RequestLog{
			Path:       r.URL.Path,
			Method:     r.Method,
			IP:         clientIP(r),
			TraceID:    httpx.TraceIDFromContext(r.Context()),
			DurationMS: time.Since(start).Milliseconds(),
			StatusCode: recorder.statusCode,
			OccurredAt: time.Now().UTC().Format(time.RFC3339Nano),
		}
		sender.Enqueue(entry)
	})
}

func signLogPayload(secret, timestamp string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp))
	mac.Write([]byte("\n"))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

func clientIP(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		return xff
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

package middleware

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAsyncLogSenderDeliversEntries(t *testing.T) {
	received := make(chan struct{}, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
		_ = r.Body.Close()
		select {
		case received <- struct{}{}:
		default:
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	sender := newAsyncLogSender(server.URL, "secret", 8, 1, 0, &http.Client{Timeout: 400 * time.Millisecond})
	if sender == nil {
		t.Fatal("expected sender to be initialized")
	}
	defer sender.Close()

	sender.Enqueue(RequestLog{Path: "/api/v1/test", Method: http.MethodGet})

	select {
	case <-received:
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for async sender delivery")
	}
}

func TestAsyncLogSenderDropsWhenQueueIsFull(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		_, _ = io.ReadAll(r.Body)
		_ = r.Body.Close()
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	sender := newAsyncLogSender(server.URL, "", 1, 1, 0, &http.Client{Timeout: 500 * time.Millisecond})
	if sender == nil {
		t.Fatal("expected sender to be initialized")
	}
	defer sender.Close()

	for i := 0; i < 50; i++ {
		sender.Enqueue(RequestLog{Path: "/api/v1/drop", Method: http.MethodPost})
	}

	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) && sender.DroppedTotal() == 0 {
		time.Sleep(10 * time.Millisecond)
	}
	if sender.DroppedTotal() == 0 {
		t.Fatal("expected queue drop count to be greater than zero")
	}
}

func TestNewAsyncLogSenderReturnsNilWhenEndpointMissing(t *testing.T) {
	sender := newAsyncLogSender("", "", 1, 1, 0, &http.Client{Timeout: 100 * time.Millisecond})
	if sender != nil {
		t.Fatal("expected nil sender when endpoint is empty")
	}
}

func TestSignLogPayloadDeterministic(t *testing.T) {
	s1 := signLogPayload("secret", "2026-02-27T00:00:00Z", []byte(`{"a":1}`))
	s2 := signLogPayload("secret", "2026-02-27T00:00:00Z", []byte(`{"a":1}`))
	if s1 == "" || s1 != s2 {
		t.Fatalf("expected deterministic signature, got %q and %q", s1, s2)
	}
}

func TestSenderStopsWithContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
		_ = r.Body.Close()
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	sender := newAsyncLogSender(server.URL, "", 8, 1, 0, &http.Client{Timeout: 200 * time.Millisecond})
	if sender == nil {
		t.Fatal("expected sender to be initialized")
	}

	cancelDone := make(chan struct{})
	go func() {
		sender.Close()
		close(cancelDone)
	}()

	select {
	case <-cancelDone:
	case <-time.After(1 * time.Second):
		t.Fatal("sender close timed out")
	}
}

func TestSendOnceReturnsErrorOnCanceledContext(t *testing.T) {
	sender := newAsyncLogSender("http://127.0.0.1:1/ingest", "", 1, 1, 0, &http.Client{Timeout: 50 * time.Millisecond})
	if sender == nil {
		t.Fatal("expected sender to be initialized")
	}
	defer sender.Close()

	sender.cancel()
	err := sender.sendOnce(RequestLog{Path: "/api/v1/test", Method: http.MethodGet})
	if err == nil {
		t.Fatal("expected sendOnce to fail when context is canceled")
	}
	if sender.ctx.Err() != context.Canceled {
		t.Fatalf("expected sender context canceled, got %v", sender.ctx.Err())
	}
}

func TestSendOnceReturnsErrorOnNon2xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
		_ = r.Body.Close()
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	sender := newAsyncLogSender(server.URL, "", 1, 1, 0, &http.Client{Timeout: 100 * time.Millisecond})
	if sender == nil {
		t.Fatal("expected sender to be initialized")
	}
	defer sender.Close()

	err := sender.sendOnce(RequestLog{Path: "/api/v1/test", Method: http.MethodGet})
	if err == nil {
		t.Fatal("expected sendOnce to fail on non-2xx response")
	}
}

func TestRetryBackoffForAttempt(t *testing.T) {
	sender := newAsyncLogSender("http://127.0.0.1:1/ingest", "", 1, 1, 3, &http.Client{Timeout: 50 * time.Millisecond})
	if sender == nil {
		t.Fatal("expected sender to be initialized")
	}
	defer sender.Close()

	sender.backoffBase = 25 * time.Millisecond
	sender.backoffMax = 80 * time.Millisecond

	if got := sender.retryBackoffForAttempt(1); got != 25*time.Millisecond {
		t.Fatalf("attempt 1 backoff mismatch: got=%v want=%v", got, 25*time.Millisecond)
	}
	if got := sender.retryBackoffForAttempt(2); got != 50*time.Millisecond {
		t.Fatalf("attempt 2 backoff mismatch: got=%v want=%v", got, 50*time.Millisecond)
	}
	if got := sender.retryBackoffForAttempt(3); got != 80*time.Millisecond {
		t.Fatalf("attempt 3 backoff mismatch: got=%v want=%v", got, 80*time.Millisecond)
	}
}

func TestNewAsyncLogSenderFromEnvBackoffConfig(t *testing.T) {
	t.Setenv("LOGGER_ENDPOINT", "http://127.0.0.1:1/ingest")
	t.Setenv("LOGGER_SHARED_SECRET", "")
	t.Setenv("LOGGER_QUEUE_SIZE", "4")
	t.Setenv("LOGGER_WORKERS", "1")
	t.Setenv("LOGGER_RETRY_MAX", "2")
	t.Setenv("LOGGER_RETRY_BACKOFF_BASE_MS", "120")
	t.Setenv("LOGGER_RETRY_BACKOFF_MAX_MS", "350")

	sender := newAsyncLogSenderFromEnv()
	if sender == nil {
		t.Fatal("expected sender from env to be initialized")
	}
	defer sender.Close()

	if sender.backoffBase != 120*time.Millisecond {
		t.Fatalf("backoff base mismatch: got=%v want=%v", sender.backoffBase, 120*time.Millisecond)
	}
	if sender.backoffMax != 350*time.Millisecond {
		t.Fatalf("backoff max mismatch: got=%v want=%v", sender.backoffMax, 350*time.Millisecond)
	}
}

func TestNewAsyncLogSenderFromEnvBackoffFallback(t *testing.T) {
	t.Setenv("LOGGER_ENDPOINT", "http://127.0.0.1:1/ingest")
	t.Setenv("LOGGER_SHARED_SECRET", "")
	t.Setenv("LOGGER_QUEUE_SIZE", "4")
	t.Setenv("LOGGER_WORKERS", "1")
	t.Setenv("LOGGER_RETRY_MAX", "2")
	t.Setenv("LOGGER_RETRY_BACKOFF_BASE_MS", "-1")
	t.Setenv("LOGGER_RETRY_BACKOFF_MAX_MS", "10")

	sender := newAsyncLogSenderFromEnv()
	if sender == nil {
		t.Fatal("expected sender from env to be initialized")
	}
	defer sender.Close()

	if sender.backoffBase != 50*time.Millisecond {
		t.Fatalf("invalid base should fallback to default, got=%v", sender.backoffBase)
	}
	if sender.backoffMax != 50*time.Millisecond {
		t.Fatalf("max below base should be clamped to base, got=%v", sender.backoffMax)
	}
}

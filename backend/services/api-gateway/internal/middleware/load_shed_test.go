package middleware

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/example/appfoundrylab/backend/services/api-gateway/internal/metrics"
)

func TestLoadSheddingRejectsWhenInFlightLimitExceeded(t *testing.T) {
	store := metrics.NewStore()
	blocker := make(chan struct{})

	h := LoadShedding(store, 1, []string{"/health"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-blocker
		w.WriteHeader(http.StatusOK)
	}))

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
		res := httptest.NewRecorder()
		h.ServeHTTP(res, req)
	}()

	time.Sleep(30 * time.Millisecond)

	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	res2 := httptest.NewRecorder()
	h.ServeHTTP(res2, req2)
	if res2.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 when in-flight limit exceeded, got %d", res2.Code)
	}
	if res2.Header().Get("Retry-After") == "" {
		t.Fatal("expected Retry-After header on shed response")
	}

	close(blocker)
	wg.Wait()

	snapshot := store.Snapshot()
	if snapshot.LoadShedTotal != 1 {
		t.Fatalf("expected load shed total=1, got %d", snapshot.LoadShedTotal)
	}
}

func TestLoadSheddingSkipsExemptPaths(t *testing.T) {
	store := metrics.NewStore()
	h := LoadShedding(store, 1, []string{"/health"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("expected exempt endpoint to pass through, got %d", res.Code)
	}

	if snapshot := store.Snapshot(); snapshot.LoadShedTotal != 0 {
		t.Fatalf("expected load shed total=0 for exempt path, got %d", snapshot.LoadShedTotal)
	}
}

func TestLoadSheddingDisabledWhenLimitNonPositive(t *testing.T) {
	h := LoadShedding(metrics.NewStore(), 0, nil)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)
	if res.Code != http.StatusNoContent {
		t.Fatalf("expected passthrough when disabled, got %d", res.Code)
	}
}

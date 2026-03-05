package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/example/appfoundrylab/backend/services/api-gateway/internal/metrics"
)

func TestHTTPMetricsCollectsRequests(t *testing.T) {
	store := metrics.NewStore()
	h := HTTPMetrics(store)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)

	snapshot := store.Snapshot()
	if snapshot.RequestsTotal != 1 {
		t.Fatalf("expected requests_total=1, got %d", snapshot.RequestsTotal)
	}
	if snapshot.RequestErrors != 1 {
		t.Fatalf("expected request_errors=1, got %d", snapshot.RequestErrors)
	}
}

func TestHTTPMetricsSkipsMetricsEndpoint(t *testing.T) {
	store := metrics.NewStore()
	h := HTTPMetrics(store)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)

	snapshot := store.Snapshot()
	if snapshot.RequestsTotal != 0 {
		t.Fatalf("expected requests_total=0 for /metrics scrape, got %d", snapshot.RequestsTotal)
	}
}

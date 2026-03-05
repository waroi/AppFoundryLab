package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/example/appfoundrylab/backend/services/api-gateway/internal/metrics"
)

func TestMetricsHandlerExposesCoreSeries(t *testing.T) {
	store := metrics.NewStore()
	store.Observe(http.StatusOK, 10*time.Millisecond)
	store.Observe(http.StatusInternalServerError, 250*time.Millisecond)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	res := httptest.NewRecorder()

	MetricsHandler(store).ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.Code)
	}

	body := res.Body.String()
	expected := []string{
		"api_gateway_requests_total 2",
		"api_gateway_request_errors_total 1",
		"api_gateway_request_error_rate 0.500000",
		"api_gateway_request_duration_ms_bucket",
		"api_gateway_request_duration_ms_sum",
		"api_gateway_request_duration_ms_count 2",
		"api_gateway_load_shed_total 0",
		"api_gateway_inflight_requests 0",
		"api_gateway_inflight_requests_peak 0",
	}
	for _, token := range expected {
		if !strings.Contains(body, token) {
			t.Fatalf("expected metrics output to contain %q, got: %s", token, body)
		}
	}
}

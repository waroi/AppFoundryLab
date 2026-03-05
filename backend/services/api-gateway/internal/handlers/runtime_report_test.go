package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/example/appfoundrylab/backend/services/api-gateway/internal/metrics"
)

func TestRuntimeReportHandler(t *testing.T) {
	store := metrics.NewStore()
	store.Observe(http.StatusOK, 15*time.Millisecond)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/runtime-report", nil)
	res := httptest.NewRecorder()

	RuntimeReportHandler(
		RuntimeConfigSummary{Profile: "standard"},
		store,
		RuntimeMetricsOptions{},
	).ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.Code)
	}

	var payload RuntimeReportSummary
	if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
		t.Fatalf("expected valid json, got %v", err)
	}
	if payload.GeneratedAt == "" {
		t.Fatal("expected generatedAt to be set")
	}
	if payload.ReportVersion == "" {
		t.Fatal("expected reportVersion to be set")
	}
	if payload.Config.Profile != "standard" {
		t.Fatalf("expected profile standard, got %s", payload.Config.Profile)
	}
	if payload.Metrics.RequestsTotal != 1 {
		t.Fatalf("expected requestsTotal=1, got %d", payload.Metrics.RequestsTotal)
	}
	if payload.Metrics.Alerts.HighestSeverity == "" {
		t.Fatal("expected alerts summary to be included")
	}
	if payload.Incident.RecommendedSeverity == "" {
		t.Fatal("expected incident summary to be included")
	}
}

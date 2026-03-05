package incidents

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/example/appfoundrylab/backend/services/api-gateway/internal/handlers"
)

func TestDispatchEventWebhookSink(t *testing.T) {
	t.Helper()

	var received handlers.RuntimeIncidentEventRecord
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST request, got %s", r.Method)
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("failed to decode payload: %v", err)
		}
		if r.Header.Get("X-Incident-Event-Signature") == "" {
			t.Fatal("expected webhook signature header to be set")
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	monitor := NewMonitor(
		func() handlers.RuntimeReportSummary { return handlers.RuntimeReportSummary{} },
		"",
		server.URL,
		"",
		"webhook-secret",
		"webhook",
		[]string{"127.0.0.1"},
		time.Second,
		time.Minute,
		server.Client(),
	)
	if monitor == nil {
		t.Fatal("expected monitor to be created")
	}

	event := handlers.RuntimeIncidentEventRecord{
		ID:        "evt-1",
		AlertCode: "gateway.error_rate",
		Status:    "active",
	}
	monitor.dispatchEvent(event)

	if received.ID != event.ID {
		t.Fatalf("expected webhook payload id %q, got %q", event.ID, received.ID)
	}
	if received.AlertCode != event.AlertCode {
		t.Fatalf("expected webhook payload alert code %q, got %q", event.AlertCode, received.AlertCode)
	}
}

func TestDeriveLoggerBaseURL(t *testing.T) {
	t.Helper()

	got := deriveLoggerBaseURL("http://logger:8090/ingest")
	if got != "http://logger:8090" {
		t.Fatalf("expected logger base URL without /ingest, got %q", got)
	}
}

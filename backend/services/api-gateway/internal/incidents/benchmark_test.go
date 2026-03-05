package incidents

import (
	"testing"
	"time"
    "net/http"
    "log"
    "io"

	"github.com/example/appfoundrylab/backend/services/api-gateway/internal/handlers"
)

// A stub round tripper to mock HTTP requests
type stubRoundTripper struct{}
func (rt *stubRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
    return &http.Response{
        StatusCode: 200,
        Body:       http.NoBody,
        Header:     make(http.Header),
    }, nil
}

func BenchmarkDispatchEvent(b *testing.B) {
    // Disable logging for benchmark
    log.SetOutput(io.Discard)

	// Create a dummy event
	now := time.Now().UTC().Format(time.RFC3339Nano)
	event := handlers.RuntimeIncidentEventRecord{
		ID:                  "test-id",
		EventType:           "opened",
		AlertCode:           "test.alert",
		Severity:            "high",
		Status:              "active",
		Source:              "runtime",
		Title:               "Test Incident",
		Summary:             "This is a test incident",
		Message:             "Test message",
		RecommendedAction:   "Test action",
		RecommendedSeverity: "high",
		TriggeredAt:         now,
		FirstSeenAt:         now,
		LastSeenAt:          now,
		BreachCount:         1,
		ReportGeneratedAt:   now,
		ReportVersion:       "1.0",
	}

	m := &Monitor{
		sink: "logger,stdout,webhook",
		loggerEndpoint: "http://example.com/logger",
		webhookURL: "http://example.com/webhook",
        client: &http.Client{
            Transport: &stubRoundTripper{},
        },
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.dispatchEvent(event)
	}
}

package handlers

import (
	"context"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/example/appfoundrylab/backend/services/api-gateway/internal/metrics"
	"github.com/example/appfoundrylab/backend/services/api-gateway/internal/worker"
)

func TestRuntimeDiagnosticsCacheReusesSnapshotWithinTTL(t *testing.T) {
	store := metrics.NewStore()
	store.Observe(http.StatusOK, 10*time.Millisecond)

	var readinessCalls atomic.Int32
	ready := &ReadyEndpoints{
		cache:        newReadyCache(),
		cacheTTL:     time.Second,
		staleIfError: time.Second,
		readinessProbe: func(_ context.Context, _ *worker.Client) (int, map[string]any, bool) {
			readinessCalls.Add(1)
			return http.StatusOK, map[string]any{
				"status": "ok",
				"checks": map[string]string{
					"postgres": "up",
					"redis":    "up",
					"worker":   "up",
				},
			}, false
		},
	}

	var loggerCalls atomic.Int32
	loggerClient := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			loggerCalls.Add(1)
			switch req.URL.Path {
			case "/health":
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader(`{"status":"ok"}`)),
				}, nil
			case "/metrics":
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body: io.NopCloser(strings.NewReader(`{
						"queueDepth":1,
						"queueCapacity":64,
						"workers":4,
						"enqueuedTotal":2,
						"droppedTotal":0,
						"processedTotal":2,
						"failedTotal":0,
						"retriedTotal":0,
						"inflightWorkers":0,
						"dropRatio":0,
						"dropAlertThresholdPct":5,
						"dropAlertThresholdHit":false
					}`)),
				}, nil
			case "/incident-events/summary":
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader(`{"totalEvents":0,"activeEvents":0,"latestEventAt":"","lastEventStatus":""}`)),
				}, nil
			default:
				return &http.Response{
					StatusCode: http.StatusNotFound,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader(`{}`)),
				}, nil
			}
		}),
	}

	cache := NewRuntimeDiagnosticsCache(
		RuntimeConfigSummary{Profile: "standard"},
		store,
		RuntimeMetricsOptions{
			ReadyEndpoints:   ready,
			LoggerEndpoint:   "http://logger.local/ingest",
			LoggerHTTPClient: loggerClient,
		},
		time.Minute,
	)

	first := cache.Report()
	second := cache.Report()
	metricsSnapshot := cache.Metrics()

	if first.GeneratedAt == "" || second.GeneratedAt == "" {
		t.Fatal("expected generatedAt to be set")
	}
	if first.GeneratedAt != second.GeneratedAt {
		t.Fatalf("expected cached report to reuse generatedAt, got %s vs %s", first.GeneratedAt, second.GeneratedAt)
	}
	if metricsSnapshot.RequestsTotal != first.Metrics.RequestsTotal {
		t.Fatalf("expected cached metrics to match report metrics, got %d vs %d", metricsSnapshot.RequestsTotal, first.Metrics.RequestsTotal)
	}
	if got := readinessCalls.Load(); got != 1 {
		t.Fatalf("expected readiness probe to run once, got %d", got)
	}
	if got := loggerCalls.Load(); got != 2 {
		t.Fatalf("expected logger health+metrics to run once each, got %d", got)
	}
}

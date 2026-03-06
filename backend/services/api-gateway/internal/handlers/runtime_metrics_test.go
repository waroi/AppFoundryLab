package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/example/appfoundrylab/backend/services/api-gateway/internal/metrics"
	"github.com/example/appfoundrylab/backend/services/api-gateway/internal/middleware"
	"github.com/example/appfoundrylab/backend/services/api-gateway/internal/worker"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func TestRuntimeMetricsHandler(t *testing.T) {
	store := metrics.NewStore()
	store.Observe(http.StatusOK, 20*time.Millisecond)
	store.Observe(http.StatusInternalServerError, 40*time.Millisecond)
	store.ObserveLoadShed()
	store.IncInflight()

	loggerClient := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
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
						"queueDepth":3,
						"queueCapacity":128,
						"workers":4,
						"enqueuedTotal":12,
						"droppedTotal":1,
						"processedTotal":11,
						"failedTotal":0,
						"retriedTotal":2,
						"inflightWorkers":1,
						"dropRatio":0.08,
						"dropAlertThresholdPct":5,
						"dropAlertThresholdHit":true
					}`)),
				}, nil
			case "/incident-events/summary":
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body: io.NopCloser(strings.NewReader(`{
						"totalEvents":8,
						"activeEvents":2,
						"latestEventAt":"2026-03-01T12:05:00Z",
						"lastEventStatus":"active"
					}`)),
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

	ready := &ReadyEndpoints{
		cache:        newReadyCache(),
		cacheTTL:     time.Second,
		staleIfError: 10 * time.Second,
		readinessProbe: func(_ context.Context, _ *worker.Client) (int, map[string]any, bool) {
			return http.StatusServiceUnavailable, map[string]any{
				"status": "degraded",
				"checks": map[string]string{
					"postgres": "up",
					"redis":    "up",
					"worker":   "down",
				},
			}, true
		},
	}
	ready.cache.Set(readySnapshot{
		httpStatus: http.StatusOK,
		payload: map[string]any{
			"status": "ok",
		},
		createdAt: time.Now().Add(-250 * time.Millisecond),
		expiresAt: time.Now().Add(750 * time.Millisecond),
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/runtime-metrics", nil)
	res := httptest.NewRecorder()

	RuntimeMetricsHandler(store, RuntimeMetricsOptions{
		ReadyEndpoints:   ready,
		LoggerEndpoint:   "http://logger.local/ingest",
		LoggerHTTPClient: loggerClient,
		RequestLoggerStatsProvider: func() middleware.AsyncLogSenderStats {
			return middleware.AsyncLogSenderStats{
				Enabled:       true,
				Endpoint:      "http://logger.local/ingest",
				QueueDepth:    2,
				QueueCapacity: 64,
				Workers:       4,
				RetryMax:      1,
				DroppedTotal:  2,
			}
		},
		IncidentEmitterStatsProvider: func() RuntimeIncidentEmitterStats {
			return RuntimeIncidentEmitterStats{
				Enabled:          true,
				Sink:             "logger+stdout",
				LastDispatchAt:   "2026-03-01T12:05:00Z",
				DispatchFailures: 0,
			}
		},
	}).ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.Code)
	}

	var payload RuntimeMetricsSummary
	if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
		t.Fatalf("expected valid json, got %v", err)
	}

	if payload.RequestsTotal != 2 {
		t.Fatalf("expected requestsTotal=2, got %d", payload.RequestsTotal)
	}
	if payload.LoadShedTotal != 1 {
		t.Fatalf("expected loadShedTotal=1, got %d", payload.LoadShedTotal)
	}
	if payload.InflightCurrent != 1 {
		t.Fatalf("expected inflightCurrent=1, got %d", payload.InflightCurrent)
	}
	if payload.LatencyAverageMS <= 0 {
		t.Fatalf("expected positive latency average, got %f", payload.LatencyAverageMS)
	}
	if payload.Health.Status != "degraded" {
		t.Fatalf("expected degraded health, got %s", payload.Health.Status)
	}
	if payload.Trace.ResponseHeader != "X-Trace-Id" {
		t.Fatalf("expected trace header X-Trace-Id, got %s", payload.Trace.ResponseHeader)
	}
	if !payload.LoggerService.Reachable {
		t.Fatal("expected logger service to be reachable")
	}
	if payload.LoggerService.QueueDepth != 3 {
		t.Fatalf("expected logger queueDepth=3, got %d", payload.LoggerService.QueueDepth)
	}
	if payload.GatewayLogger.DroppedTotal != 2 {
		t.Fatalf("expected gateway logger droppedTotal=2, got %d", payload.GatewayLogger.DroppedTotal)
	}
	if len(payload.RecentHistory) == 0 {
		t.Fatal("expected recent history to be populated")
	}
	if payload.Alerts.ActiveCount == 0 {
		t.Fatal("expected active alerts to be populated")
	}
	if payload.Alerts.HighestSeverity != "critical" {
		t.Fatalf("expected highest severity critical, got %s", payload.Alerts.HighestSeverity)
	}
	if len(payload.Alerts.Items) == 0 {
		t.Fatal("expected alert items to be populated")
	}
	if !payload.IncidentJournal.Enabled {
		t.Fatal("expected incident journal to be enabled")
	}
	if payload.IncidentJournal.TotalEvents != 8 {
		t.Fatalf("expected incident total=8, got %d", payload.IncidentJournal.TotalEvents)
	}
	if len(payload.Warnings) == 0 {
		t.Fatal("expected runtime warnings to be populated")
	}
}

func TestRuntimeMetricsLoggerHealthDegradedStillCountsAsReachable(t *testing.T) {
	loggerClient := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Path {
			case "/health":
				return &http.Response{
					StatusCode: http.StatusServiceUnavailable,
					Header:     make(http.Header),
					Body: io.NopCloser(strings.NewReader(`{
						"status":"degraded",
						"checks":{"mongo":"down"},
						"lastError":"mongo down"
					}`)),
				}, nil
			case "/metrics":
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body: io.NopCloser(strings.NewReader(`{
						"queueDepth":1,
						"queueCapacity":32,
						"workers":2,
						"enqueuedTotal":5,
						"droppedTotal":0,
						"processedTotal":4,
						"failedTotal":1,
						"retriedTotal":1,
						"inflightWorkers":0,
						"dropRatio":0.0,
						"dropAlertThresholdPct":5,
						"dropAlertThresholdHit":false
					}`)),
				}, nil
			case "/incident-events/summary":
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader(`{"totalEvents":0,"activeEvents":0}`)),
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

	payload := BuildRuntimeMetricsSummary(metrics.NewStore(), RuntimeMetricsOptions{
		LoggerEndpoint:   "http://logger.local/ingest",
		LoggerHTTPClient: loggerClient,
	})

	if !payload.LoggerService.Reachable {
		t.Fatal("expected logger service endpoint to be reachable")
	}
	if payload.LoggerService.HealthStatus != "degraded" {
		t.Fatalf("expected degraded logger health, got %s", payload.LoggerService.HealthStatus)
	}
	if payload.LoggerService.LastError != "mongo down" {
		t.Fatalf("expected logger last error to be propagated, got %s", payload.LoggerService.LastError)
	}

	found := false
	for _, warning := range payload.Warnings {
		if warning == "logger service health is degraded" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected warning for degraded logger health")
	}

	found = false
	for _, item := range payload.Alerts.Items {
		if item.Code == "logger.health_degraded" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected logger.health_degraded alert item")
	}
}

func TestRuntimeMetricsHandlerTracksDegradedLoggerHealth(t *testing.T) {
	loggerClient := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Path {
			case "/health":
				return &http.Response{
					StatusCode: http.StatusServiceUnavailable,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader(`{"status":"degraded","checks":{"mongo":"down"}}`)),
				}, nil
			case "/metrics":
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader(`{"queueDepth":0,"queueCapacity":64,"workers":2}`)),
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

	payload := BuildRuntimeMetricsSummary(metrics.NewStore(), RuntimeMetricsOptions{
		LoggerEndpoint:   "http://logger.local/ingest",
		LoggerHTTPClient: loggerClient,
		RequestLoggerStatsProvider: func() middleware.AsyncLogSenderStats {
			return middleware.AsyncLogSenderStats{}
		},
	})

	if !payload.LoggerService.Reachable {
		t.Fatal("expected logger service to be reachable even when health is degraded")
	}
	if payload.LoggerService.HealthStatus != "degraded" {
		t.Fatalf("expected degraded health status, got %s", payload.LoggerService.HealthStatus)
	}
	if len(payload.Warnings) == 0 {
		t.Fatal("expected warnings for degraded logger health")
	}
	if payload.Alerts.ActiveCount == 0 {
		t.Fatal("expected alerts for degraded logger health")
	}
}

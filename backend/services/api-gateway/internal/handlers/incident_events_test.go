package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRuntimeIncidentEventsHandler(t *testing.T) {
	t.Run("proxies incident events", func(t *testing.T) {
		client := &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if req.URL.Path != "/incident-events" {
					return &http.Response{
						StatusCode: http.StatusNotFound,
						Header:     make(http.Header),
						Body:       io.NopCloser(strings.NewReader(`{}`)),
					}, nil
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body: io.NopCloser(strings.NewReader(`{
						"items":[
							{
								"id":"evt-1",
								"eventType":"opened",
								"alertCode":"gateway.error_rate",
								"severity":"critical",
								"status":"active",
								"source":"api-gateway",
								"title":"Critical runtime incident detected",
								"summary":"1 active alert(s), highest severity critical, health degraded, 1 runbook(s) mapped",
								"message":"gateway error rate stayed above 5% in recent samples",
								"recommendedAction":"inspect recent 5xx responses and dependent services",
								"recommendedSeverity":"sev-1",
								"triggeredAt":"2026-03-01T12:05:00Z",
								"firstSeenAt":"2026-03-01T12:00:00Z",
								"lastSeenAt":"2026-03-01T12:05:00Z",
								"breachCount":2,
								"traceId":"",
								"reportGeneratedAt":"2026-03-01T12:05:00Z",
								"reportVersion":"2026-03-incident-v1",
								"runbooks":[]
							}
						]
					}`)),
				}, nil
			}),
		}

		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/incident-events", nil)
		res := httptest.NewRecorder()
		RuntimeIncidentEventsHandler("http://logger.local/ingest", client).ServeHTTP(res, req)
		if res.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", res.Code)
		}
		if !strings.Contains(res.Body.String(), `"alertCode":"gateway.error_rate"`) {
			t.Fatalf("expected incident event payload, got %s", res.Body.String())
		}
	})

	t.Run("returns structured error when logger fails", func(t *testing.T) {
		client := &http.Client{
			Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusInternalServerError,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader(`{}`)),
				}, nil
			}),
		}

		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/incident-events", nil)
		res := httptest.NewRecorder()
		RuntimeIncidentEventsHandler("http://logger.local/ingest", client).ServeHTTP(res, req)
		if res.Code != http.StatusServiceUnavailable {
			t.Fatalf("expected 503, got %d", res.Code)
		}
		if !strings.Contains(res.Body.String(), `"code":"logger_unavailable"`) {
			t.Fatalf("expected logger_unavailable code, got %s", res.Body.String())
		}
	})
}

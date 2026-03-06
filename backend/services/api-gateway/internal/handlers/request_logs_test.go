package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRuntimeRequestLogsHandler(t *testing.T) {
	t.Run("proxies trace query to logger", func(t *testing.T) {
		client := &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if req.URL.Path != "/request-logs" {
					t.Fatalf("expected /request-logs, got %s", req.URL.Path)
				}
				if req.URL.Query().Get("traceId") != "trace-123" {
					t.Fatalf("expected traceId to be forwarded")
				}
				if req.URL.Query().Get("limit") != "5" {
					t.Fatalf("expected limit to be forwarded")
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body: io.NopCloser(strings.NewReader(`{
						"items":[
							{
								"path":"/restore-drill/trace-123",
								"method":"GET",
								"ip":"127.0.0.1",
								"traceId":"trace-123",
								"durationMs":12,
								"statusCode":200,
								"occurredAt":"2026-03-01T12:05:00Z"
							}
						]
					}`)),
				}, nil
			}),
		}

		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/request-logs?traceId=trace-123&limit=5", nil)
		res := httptest.NewRecorder()
		RuntimeRequestLogsHandler("http://logger.local/ingest", client).ServeHTTP(res, req)
		if res.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", res.Code)
		}
		if !strings.Contains(res.Body.String(), `"traceId":"trace-123"`) {
			t.Fatalf("expected proxied request log payload, got %s", res.Body.String())
		}
	})

	t.Run("returns unavailable when logger fails", func(t *testing.T) {
		client := &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusInternalServerError,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader(`{}`)),
				}, nil
			}),
		}

		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/request-logs?traceId=trace-123&limit=5", nil)
		res := httptest.NewRecorder()
		RuntimeRequestLogsHandler("http://logger.local/ingest", client).ServeHTTP(res, req)
		if res.Code != http.StatusServiceUnavailable {
			t.Fatalf("expected 503, got %d", res.Code)
		}
		if !strings.Contains(res.Body.String(), `"code":"logger_unavailable"`) {
			t.Fatalf("expected logger_unavailable code, got %s", res.Body.String())
		}
	})

	t.Run("clamps oversized limit before proxying", func(t *testing.T) {
		client := &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if req.URL.Query().Get("limit") != "100" {
					t.Fatalf("expected oversized limit to clamp to 100, got %s", req.URL.Query().Get("limit"))
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader(`{"items":[]}`)),
				}, nil
			}),
		}

		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/request-logs?limit=5000", nil)
		res := httptest.NewRecorder()
		RuntimeRequestLogsHandler("http://logger.local/ingest", client).ServeHTTP(res, req)
		if res.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", res.Code)
		}
	})

	t.Run("returns bad request for invalid limit", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/request-logs?limit=not-a-number", nil)
		res := httptest.NewRecorder()
		RuntimeRequestLogsHandler("http://logger.local/ingest", nil).ServeHTTP(res, req)
		if res.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", res.Code)
		}
		if !strings.Contains(res.Body.String(), `"code":"invalid_query_limit"`) {
			t.Fatalf("expected invalid_query_limit code, got %s", res.Body.String())
		}
	})
}

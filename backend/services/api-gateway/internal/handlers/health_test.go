package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/example/appfoundrylab/backend/services/api-gateway/internal/worker"
)

func TestLiveHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	res := httptest.NewRecorder()

	LiveHandler(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", res.Code)
	}

	var payload map[string]any
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if payload["status"] != "ok" {
		t.Errorf("expected status 'ok', got %v", payload["status"])
	}

	if payload["service"] != "api-gateway" {
		t.Errorf("expected service 'api-gateway', got %v", payload["service"])
	}
}

func TestReadyEndpointUsesFreshCache(t *testing.T) {
	now := time.Unix(100, 0)
	probeCalls := 0
	endpoints := &ReadyEndpoints{
		cache:        newReadyCache(),
		cacheTTL:     time.Second,
		staleIfError: 10 * time.Second,
		now:          func() time.Time { return now },
		readinessProbe: func(_ context.Context, _ *worker.Client) (int, map[string]any, bool) {
			probeCalls++
			return http.StatusOK, map[string]any{"status": "ok"}, false
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	res := httptest.NewRecorder()
	endpoints.Ready(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("expected first call 200, got %d", res.Code)
	}
	if probeCalls != 1 {
		t.Fatalf("expected first call to probe once, got %d", probeCalls)
	}
	if res.Header().Get("X-Ready-Cache") != "miss" {
		t.Fatalf("expected X-Ready-Cache=miss, got %s", res.Header().Get("X-Ready-Cache"))
	}

	res2 := httptest.NewRecorder()
	endpoints.Ready(res2, req)
	if probeCalls != 1 {
		t.Fatalf("expected second call to use cache, probe calls=%d", probeCalls)
	}
	if res2.Header().Get("X-Ready-Cache") != "fresh" {
		t.Fatalf("expected X-Ready-Cache=fresh, got %s", res2.Header().Get("X-Ready-Cache"))
	}
}

func TestReadyEndpointServesStaleOnError(t *testing.T) {
	now := time.Unix(200, 0)
	shouldFail := false
	endpoints := &ReadyEndpoints{
		cache:        newReadyCache(),
		cacheTTL:     time.Second,
		staleIfError: 30 * time.Second,
		now:          func() time.Time { return now },
		readinessProbe: func(_ context.Context, _ *worker.Client) (int, map[string]any, bool) {
			if shouldFail {
				return http.StatusServiceUnavailable, map[string]any{"status": "degraded"}, true
			}
			return http.StatusOK, map[string]any{"status": "ok"}, false
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	res1 := httptest.NewRecorder()
	endpoints.Ready(res1, req)
	if res1.Code != http.StatusOK {
		t.Fatalf("expected prime call 200, got %d", res1.Code)
	}

	now = now.Add(2 * time.Second)
	shouldFail = true
	res2 := httptest.NewRecorder()
	endpoints.Ready(res2, req)
	if res2.Code != http.StatusOK {
		t.Fatalf("expected stale response 200, got %d", res2.Code)
	}
	if res2.Header().Get("X-Ready-Cache") != "stale" {
		t.Fatalf("expected X-Ready-Cache=stale, got %s", res2.Header().Get("X-Ready-Cache"))
	}
}

func TestReadyInvalidate(t *testing.T) {
	probeCalls := 0
	endpoints := &ReadyEndpoints{
		cache:        newReadyCache(),
		cacheTTL:     time.Second,
		staleIfError: 10 * time.Second,
		now:          time.Now,
		readinessProbe: func(_ context.Context, _ *worker.Client) (int, map[string]any, bool) {
			probeCalls++
			return http.StatusOK, map[string]any{"status": "ok"}, false
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	res := httptest.NewRecorder()
	endpoints.Ready(res, req)
	if probeCalls != 1 {
		t.Fatalf("expected one probe call, got %d", probeCalls)
	}

	ireq := httptest.NewRequest(http.MethodPost, "/health/ready/invalidate", nil)
	ires := httptest.NewRecorder()
	endpoints.Invalidate(ires, ireq)
	if ires.Code != http.StatusOK {
		t.Fatalf("expected invalidate 200, got %d", ires.Code)
	}

	nowRes := httptest.NewRecorder()
	endpoints.Ready(nowRes, req)
	if probeCalls != 2 {
		t.Fatalf("expected probe call after invalidation, got %d", probeCalls)
	}
}

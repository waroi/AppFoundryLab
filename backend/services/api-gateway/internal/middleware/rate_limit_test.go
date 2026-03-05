package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiterAllowAndLimit(t *testing.T) {
	rl := NewRateLimiter(2, time.Minute)

	allowed, _, _ := rl.Allow("ip:/api")
	if !allowed {
		t.Fatalf("first request should be allowed")
	}
	allowed, _, _ = rl.Allow("ip:/api")
	if !allowed {
		t.Fatalf("second request should be allowed")
	}
	allowed, _, retryAfter := rl.Allow("ip:/api")
	if allowed {
		t.Fatalf("third request should be limited")
	}
	if retryAfter <= 0 {
		t.Fatalf("retry after should be positive")
	}
}

func TestRateLimitMiddlewareReturns429(t *testing.T) {
	mw := RateLimitByIP(1, time.Minute)
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req1 := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	req1.RemoteAddr = "127.0.0.1:1111"
	res1 := httptest.NewRecorder()
	h.ServeHTTP(res1, req1)
	if res1.Code != http.StatusOK {
		t.Fatalf("first request expected 200, got %d", res1.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	req2.RemoteAddr = "127.0.0.1:1111"
	res2 := httptest.NewRecorder()
	h.ServeHTTP(res2, req2)
	if res2.Code != http.StatusTooManyRequests {
		t.Fatalf("second request expected 429, got %d", res2.Code)
	}
	if res2.Header().Get("Retry-After") == "" {
		t.Fatalf("expected Retry-After header")
	}
}

func TestDistributedRateLimitMiddlewareFailOpenWhenRedisUnavailable(t *testing.T) {
	mw := RateLimitByIPDistributed(nil, "api", 1, time.Minute)
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	req.RemoteAddr = "127.0.0.1:1111"
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected fail-open 200 response when redis is unavailable, got %d", res.Code)
	}
}

func TestDistributedRateLimitMiddlewareFailClosedWhenRedisUnavailable(t *testing.T) {
	mw := RateLimitByIPDistributedWithFailureMode(nil, "api", 1, time.Minute, "closed")
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	req.RemoteAddr = "127.0.0.1:1111"
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)

	if res.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected fail-closed 503 response when redis is unavailable, got %d", res.Code)
	}
}

func TestNormalizeRedisFailureMode(t *testing.T) {
	if got := normalizeRedisFailureMode("closed"); got != "closed" {
		t.Fatalf("expected closed mode, got %q", got)
	}
	if got := normalizeRedisFailureMode("invalid"); got != "open" {
		t.Fatalf("expected invalid mode to fallback open, got %q", got)
	}
}

func TestToInt64(t *testing.T) {
	v, err := toInt64(int64(42))
	if err != nil || v != 42 {
		t.Fatalf("expected int64 conversion success, got v=%d err=%v", v, err)
	}

	v, err = toInt64(7)
	if err != nil || v != 7 {
		t.Fatalf("expected int conversion success, got v=%d err=%v", v, err)
	}

	if _, err := toInt64("bad"); err == nil {
		t.Fatal("expected conversion error for string input")
	}
}

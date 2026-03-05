package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDeprecatedAPIVersionHeaders(t *testing.T) {
	mw := DeprecatedAPIVersion("/api/v1", "Fri, 27 Feb 2026 00:00:00 GMT", "Tue, 30 Jun 2026 23:59:59 GMT")

	h := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)

	if res.Header().Get("Deprecation") == "" {
		t.Fatal("expected Deprecation header")
	}
	if res.Header().Get("Sunset") == "" {
		t.Fatal("expected Sunset header")
	}
	if res.Header().Get("Link") != "</api/v1>; rel=\"successor-version\"" {
		t.Fatalf("unexpected Link header: %s", res.Header().Get("Link"))
	}
	if res.Header().Get("Warning") == "" {
		t.Fatal("expected Warning header")
	}
}

package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRuntimeConfigHandler(t *testing.T) {
	summary := RuntimeConfigSummary{
		Profile: "standard",
		HTTP: RuntimeHTTPSummary{
			LegacyAPIEnabled: true,
		},
		Security: RuntimeSecuritySummary{
			StrictDependencies: true,
		},
		Operations: RuntimeOperationsSummary{
			RateLimitStore: "memory",
		},
		Warnings: []string{"default bootstrap credentials are still active"},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/runtime-config", nil)
	res := httptest.NewRecorder()

	RuntimeConfigHandler(summary).ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.Code)
	}

	var payload RuntimeConfigSummary
	if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
		t.Fatalf("expected valid json, got %v", err)
	}
	if payload.Profile != "standard" {
		t.Fatalf("expected profile standard, got %q", payload.Profile)
	}
	if len(payload.Warnings) != 1 {
		t.Fatalf("expected one warning, got %d", len(payload.Warnings))
	}
}

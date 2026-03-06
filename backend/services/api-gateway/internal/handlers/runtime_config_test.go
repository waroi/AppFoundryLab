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
			RequestLogging: RuntimeRequestLoggingSummary{
				TrustedProxyCIDRs: []string{"127.0.0.1/32"},
			},
			LoggerTiming: RuntimeLoggerTimingSummary{
				HealthTimeoutMS:                     1500,
				IngestTimestampMaxAgeSeconds:        300,
				IngestTimestampMaxFutureSkewSeconds: 5,
			},
			DependencyPolicies: []RuntimeDependencyPolicy{
				{
					Route:           "GET /api/v1/users",
					Dependency:      "postgres",
					StrictMode:      "fail-fast",
					NonStrictMode:   "continue",
					RuntimeBehavior: "503 users_unavailable",
				},
			},
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
	if len(payload.Operations.DependencyPolicies) != 1 {
		t.Fatalf("expected one dependency policy, got %d", len(payload.Operations.DependencyPolicies))
	}
	if payload.Operations.DependencyPolicies[0].Dependency != "postgres" {
		t.Fatalf("expected postgres dependency policy, got %q", payload.Operations.DependencyPolicies[0].Dependency)
	}
	if got := payload.Operations.RequestLogging.TrustedProxyCIDRs[0]; got != "127.0.0.1/32" {
		t.Fatalf("expected normalized trusted proxy cidr, got %q", got)
	}
	if payload.Operations.LoggerTiming.HealthTimeoutMS != 1500 {
		t.Fatalf("expected logger health timeout 1500ms, got %d", payload.Operations.LoggerTiming.HealthTimeoutMS)
	}
}

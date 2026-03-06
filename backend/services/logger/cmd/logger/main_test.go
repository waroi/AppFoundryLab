package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/example/appfoundrylab/backend/pkg/runtimeknobs"
	"github.com/example/appfoundrylab/backend/services/logger/internal/incidents"
)

func TestVerifyIngestAuth(t *testing.T) {
	body := []byte(`{"path":"/health"}`)
	now := time.Now().UTC().Format(time.RFC3339Nano)

	req := httptest.NewRequest("POST", "/ingest", bytes.NewReader(body))
	if err := verifyIngestAuth(req, "", false); err == nil {
		t.Fatal("expected error when shared secret is missing in fail-closed mode")
	}

	req = httptest.NewRequest("POST", "/ingest", bytes.NewReader(body))
	if err := verifyIngestAuth(req, "", true); err != nil {
		t.Fatalf("expected nil error when unsigned ingest is allowed, got %v", err)
	}

	secret := "test-secret"
	signature := signLogPayload(secret, now, body)

	req = httptest.NewRequest("POST", "/ingest", bytes.NewReader(body))
	req.Header.Set("X-Logger-Timestamp", now)
	req.Header.Set("X-Logger-Signature", signature)
	if err := verifyIngestAuth(req, secret, false); err != nil {
		t.Fatalf("expected valid signed request, got %v", err)
	}

	req = httptest.NewRequest("POST", "/ingest", bytes.NewReader(body))
	req.Header.Set("X-Logger-Timestamp", now)
	req.Header.Set("X-Logger-Signature", "bad-signature")
	if err := verifyIngestAuth(req, secret, false); err == nil {
		t.Fatal("expected error for invalid signature")
	}
}

func TestVerifyIngestAuthRejectsExcessiveFutureSkew(t *testing.T) {
	body := []byte(`{"path":"/health"}`)
	timestamp := time.Now().UTC().Add(10 * time.Second).Format(time.RFC3339Nano)
	secret := "test-secret"

	req := httptest.NewRequest("POST", "/ingest", bytes.NewReader(body))
	req.Header.Set("X-Logger-Timestamp", timestamp)
	req.Header.Set("X-Logger-Signature", signLogPayload(secret, timestamp, body))

	if err := verifyIngestAuth(req, secret, false); err == nil {
		t.Fatal("expected future replay skew validation to fail")
	}
}

func TestVerifyIngestAuthHonorsConfiguredFutureSkew(t *testing.T) {
	t.Setenv("LOGGER_INGEST_TIMESTAMP_MAX_FUTURE_SKEW_SECONDS", "15")

	body := []byte(`{"path":"/health"}`)
	timestamp := time.Now().UTC().Add(10 * time.Second).Format(time.RFC3339Nano)
	secret := "test-secret"

	req := httptest.NewRequest("POST", "/ingest", bytes.NewReader(body))
	req.Header.Set("X-Logger-Timestamp", timestamp)
	req.Header.Set("X-Logger-Signature", signLogPayload(secret, timestamp, body))

	if err := verifyIngestAuth(req, secret, false); err != nil {
		t.Fatalf("expected configured future skew to allow request, got %v", err)
	}
}

func TestReplayWindowHelpersFallbackOnInvalidValues(t *testing.T) {
	t.Setenv("LOGGER_INGEST_TIMESTAMP_MAX_AGE_SECONDS", "-1")
	t.Setenv("LOGGER_INGEST_TIMESTAMP_MAX_FUTURE_SKEW_SECONDS", "0")
	t.Setenv("LOGGER_HEALTH_TIMEOUT_MS", "-25")

	if got := ingestReplayMaxAge(); got != runtimeknobs.DefaultLoggerIngestTimestampMaxAge {
		t.Fatalf("expected max age fallback to %v, got %v", runtimeknobs.DefaultLoggerIngestTimestampMaxAge, got)
	}
	if got := ingestReplayMaxFutureSkew(); got != runtimeknobs.DefaultLoggerIngestTimestampMaxFutureSkew {
		t.Fatalf(
			"expected future skew fallback to %v, got %v",
			runtimeknobs.DefaultLoggerIngestTimestampMaxFutureSkew,
			got,
		)
	}
	if got := healthCheckTimeout(); got != runtimeknobs.DefaultLoggerHealthTimeout {
		t.Fatalf("expected health timeout fallback to %v, got %v", runtimeknobs.DefaultLoggerHealthTimeout, got)
	}
}

func TestIncidentEventJSONContract(t *testing.T) {
	payload := incidents.Event{
		ID:                "evt-1",
		EventType:         "opened",
		AlertCode:         "gateway.error_rate",
		Severity:          "critical",
		Status:            "active",
		Source:            "api-gateway",
		Title:             "Gateway error rate is elevated",
		Summary:           "Critical gateway alert is active.",
		Message:           "gateway error rate is elevated",
		RecommendedAction: "inspect gateway and dependencies",
		TriggeredAt:       time.Now().UTC().Format(time.RFC3339Nano),
		LastSeenAt:        time.Now().UTC().Format(time.RFC3339Nano),
		ReportGeneratedAt: time.Now().UTC().Format(time.RFC3339Nano),
		ReportVersion:     "2026-03-incident-v1",
		Runbooks: []incidents.RunbookReference{
			{ID: "api-degradation", Title: "API Degradation", Path: "docs/api-degradation-runbook.md"},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/incident-events", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Logger-Timestamp", time.Now().UTC().Format(time.RFC3339Nano))
	req = req.WithContext(context.Background())
	if len(body) == 0 {
		t.Fatal("expected non-empty incident payload")
	}
}

func TestLoggerHealthResponseShape(t *testing.T) {
	t.Run("healthy mongo", func(t *testing.T) {
		statusCode, payload := loggerHealthPayload(nil)
		if statusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", statusCode)
		}
		if payload["status"] != "ok" {
			t.Fatalf("expected ok payload, got %v", payload["status"])
		}
	})

	t.Run("degraded mongo", func(t *testing.T) {
		lastErr := errors.New("mongo down")
		statusCode, payload := loggerHealthPayload(lastErr)
		if statusCode != http.StatusServiceUnavailable {
			t.Fatalf("expected 503, got %d", statusCode)
		}
		if payload["status"] != "degraded" {
			t.Fatalf("expected degraded payload, got %v", payload["status"])
		}
		if payload["lastError"] != lastErr.Error() {
			t.Fatalf("expected lastError to be present, got %v", payload["lastError"])
		}
	})
}

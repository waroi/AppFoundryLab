package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

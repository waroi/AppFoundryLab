package handlers

import "testing"

func TestBuildRuntimeIncidentSummaryMapsRunbooksAndSeverity(t *testing.T) {
	config := RuntimeConfigSummary{
		Security: RuntimeSecuritySummary{
			DefaultCredentialsInUse: true,
		},
	}
	metrics := RuntimeMetricsSummary{
		Alerts: RuntimeAlertSummary{
			ActiveCount:      2,
			HighestSeverity:  runtimeAlertSeverityCritical,
			RecentlyBreached: true,
			Items: []RuntimeAlertItem{
				{
					Code:              "gateway.error_rate",
					Severity:          runtimeAlertSeverityCritical,
					Status:            "active",
					Source:            "gateway",
					Message:           "gateway error rate stayed high",
					RecommendedAction: "inspect 5xx responses",
					LastTriggeredAt:   "2026-03-01T12:00:00Z",
				},
				{
					Code:              "health.degraded",
					Severity:          runtimeAlertSeverityCritical,
					Status:            "active",
					Source:            "health",
					Message:           "health degraded",
					RecommendedAction: "inspect dependencies",
					LastTriggeredAt:   "2026-03-01T12:01:00Z",
				},
			},
		},
		Health: RuntimeHealthSummary{
			Status:     "degraded",
			HTTPStatus: 503,
		},
	}

	incident := BuildRuntimeIncidentSummary(config, metrics)
	if incident.RecommendedSeverity != "sev-1" {
		t.Fatalf("expected sev-1, got %s", incident.RecommendedSeverity)
	}
	if len(incident.NextActions) == 0 {
		t.Fatal("expected next actions to be populated")
	}
	if len(incident.TriggeredAlerts) != 2 {
		t.Fatalf("expected 2 triggered alerts, got %d", len(incident.TriggeredAlerts))
	}
	runbooks := BuildRuntimeRunbookReferences(metrics.Alerts)
	if len(runbooks) < 2 {
		t.Fatalf("expected at least 2 runbooks, got %d", len(runbooks))
	}
}

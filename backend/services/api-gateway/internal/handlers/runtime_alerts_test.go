package handlers

import (
	"testing"

	"github.com/example/appfoundrylab/backend/services/api-gateway/internal/metrics"
)

func TestBuildRuntimeAlertSummaryUsesRecentHistoryAndRuntimeSignals(t *testing.T) {
	snapshot := metrics.Snapshot{
		RecentHistory: []metrics.TrendPoint{
			{
				RecordedAt:       "2026-03-01T10:00:00Z",
				RequestsTotal:    24,
				RequestErrors:    2,
				ErrorRate:        0.08,
				LatencyAverageMS: 620,
				LoadShedTotal:    1,
			},
			{
				RecordedAt:       "2026-03-01T10:05:00Z",
				RequestsTotal:    28,
				RequestErrors:    3,
				ErrorRate:        0.11,
				LatencyAverageMS: 710,
				LoadShedTotal:    1,
			},
		},
	}

	alerts := buildRuntimeAlertSummary(
		snapshot,
		RuntimeHealthSummary{Status: "degraded", LastCheckedAt: "2026-03-01T10:05:30Z"},
		RuntimeGatewayLoggerSummary{Enabled: true, DroppedTotal: 2},
		RuntimeLoggerServiceSummary{Configured: true, Reachable: false},
	)

	if alerts.ActiveCount < 3 {
		t.Fatalf("expected multiple active alerts, got %d", alerts.ActiveCount)
	}
	if alerts.HighestSeverity != runtimeAlertSeverityCritical {
		t.Fatalf("expected critical highest severity, got %s", alerts.HighestSeverity)
	}
	if !alerts.RecentlyBreached {
		t.Fatal("expected recently breached to be true")
	}
	if len(alerts.Items) < 4 {
		t.Fatalf("expected at least 4 alert items, got %d", len(alerts.Items))
	}
	if alerts.Items[0].LastTriggeredAt == "" {
		t.Fatal("expected alert item to track last triggered time")
	}
}

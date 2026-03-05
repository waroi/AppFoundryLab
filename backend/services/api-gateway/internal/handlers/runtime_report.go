package handlers

import (
	"net/http"
	"time"

	"github.com/example/appfoundrylab/backend/services/api-gateway/internal/metrics"
	"github.com/example/appfoundrylab/backend/services/api-gateway/pkg/httpx"
)

type RuntimeReportSummary struct {
	GeneratedAt   string                    `json:"generatedAt"`
	ReportVersion string                    `json:"reportVersion"`
	Config        RuntimeConfigSummary      `json:"config"`
	Metrics       RuntimeMetricsSummary     `json:"metrics"`
	Runbooks      []RuntimeRunbookReference `json:"runbooks"`
	Incident      RuntimeIncidentSummary    `json:"incident"`
}

type MetricsSnapshotProvider interface {
	Snapshot() metrics.Snapshot
}

func BuildRuntimeReportSummary(config RuntimeConfigSummary, store MetricsSnapshotProvider, options RuntimeMetricsOptions) RuntimeReportSummary {
	return buildRuntimeReportSummaryAt(
		config,
		store,
		options,
		time.Now().UTC().Format(time.RFC3339Nano),
	)
}

func RuntimeReportHandler(config RuntimeConfigSummary, store MetricsSnapshotProvider, options RuntimeMetricsOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		httpx.WriteJSON(w, http.StatusOK, BuildRuntimeReportSummary(config, store, options))
	}
}

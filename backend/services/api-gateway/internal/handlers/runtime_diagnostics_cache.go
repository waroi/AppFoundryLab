package handlers

import (
	"net/http"
	"sync"
	"time"

	"github.com/example/appfoundrylab/backend/services/api-gateway/pkg/httpx"
)

type runtimeDiagnosticsSnapshot struct {
	expiresAt time.Time
	report    RuntimeReportSummary
}

type RuntimeDiagnosticsCache struct {
	mu sync.Mutex

	config  RuntimeConfigSummary
	store   MetricsSnapshotProvider
	options RuntimeMetricsOptions
	ttl     time.Duration

	snapshot runtimeDiagnosticsSnapshot
}

func NewRuntimeDiagnosticsCache(
	config RuntimeConfigSummary,
	store MetricsSnapshotProvider,
	options RuntimeMetricsOptions,
	ttl time.Duration,
) *RuntimeDiagnosticsCache {
	return &RuntimeDiagnosticsCache{
		config:  config,
		store:   store,
		options: normalizeRuntimeMetricsOptions(options),
		ttl:     ttl,
	}
}

func (c *RuntimeDiagnosticsCache) Metrics() RuntimeMetricsSummary {
	return c.Report().Metrics
}

func (c *RuntimeDiagnosticsCache) Report() RuntimeReportSummary {
	if c == nil {
		return RuntimeReportSummary{}
	}

	now := time.Now()
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.snapshot.report.GeneratedAt != "" && (c.ttl <= 0 || now.Before(c.snapshot.expiresAt)) {
		return c.snapshot.report
	}

	report := buildRuntimeReportSummaryAt(
		c.config,
		c.store,
		c.options,
		now.UTC().Format(time.RFC3339Nano),
	)
	c.snapshot = runtimeDiagnosticsSnapshot{
		expiresAt: now.Add(c.ttl),
		report:    report,
	}
	return report
}

func buildRuntimeReportSummaryAt(
	config RuntimeConfigSummary,
	store MetricsSnapshotProvider,
	options RuntimeMetricsOptions,
	generatedAt string,
) RuntimeReportSummary {
	options = normalizeRuntimeMetricsOptions(options)
	metricsSummary := BuildRuntimeMetricsSummary(store, options)
	runbooks := BuildRuntimeRunbookReferences(metricsSummary.Alerts)
	return RuntimeReportSummary{
		GeneratedAt:   generatedAt,
		ReportVersion: runtimeReportVersion,
		Config:        config,
		Metrics:       metricsSummary,
		Runbooks:      runbooks,
		Incident:      BuildRuntimeIncidentSummary(config, metricsSummary),
	}
}

func RuntimeMetricsHandlerWithCache(cache *RuntimeDiagnosticsCache) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		httpx.WriteJSON(w, http.StatusOK, cache.Metrics())
	}
}

func RuntimeReportHandlerWithCache(cache *RuntimeDiagnosticsCache) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		httpx.WriteJSON(w, http.StatusOK, cache.Report())
	}
}

func RuntimeIncidentReportHandlerWithCache(cache *RuntimeDiagnosticsCache) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		httpx.WriteJSON(w, http.StatusOK, cache.Report())
	}
}

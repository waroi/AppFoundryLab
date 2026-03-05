package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/example/appfoundrylab/backend/services/api-gateway/internal/metrics"
	"github.com/example/appfoundrylab/backend/services/api-gateway/internal/middleware"
	"github.com/example/appfoundrylab/backend/services/api-gateway/pkg/httpx"
)

type RuntimeTraceSummary struct {
	Enabled           bool   `json:"enabled"`
	ResponseHeader    string `json:"responseHeader"`
	ForwardedToLogger bool   `json:"forwardedToLogger"`
	StoredOnLoggerAs  string `json:"storedOnLoggerAs"`
	StorageField      string `json:"storageField"`
}

type RuntimeHealthSummary struct {
	Status            string `json:"status"`
	HTTPStatus        int    `json:"httpStatus"`
	Postgres          string `json:"postgres"`
	Redis             string `json:"redis"`
	Worker            string `json:"worker"`
	CacheState        string `json:"cacheState"`
	CacheAgeMS        int64  `json:"cacheAgeMs"`
	CacheTTLMS        int64  `json:"cacheTtlMs"`
	StaleIfErrorTTLMS int64  `json:"staleIfErrorTtlMs"`
	LastCheckedAt     string `json:"lastCheckedAt"`
}

type RuntimeGatewayLoggerSummary struct {
	Enabled       bool   `json:"enabled"`
	Endpoint      string `json:"endpoint"`
	QueueDepth    int    `json:"queueDepth"`
	QueueCapacity int    `json:"queueCapacity"`
	Workers       int    `json:"workers"`
	RetryMax      int    `json:"retryMax"`
	DroppedTotal  uint64 `json:"droppedTotal"`
}

type RuntimeLoggerServiceSummary struct {
	Configured            bool    `json:"configured"`
	Reachable             bool    `json:"reachable"`
	EndpointBase          string  `json:"endpointBase"`
	HealthStatus          string  `json:"healthStatus"`
	QueueDepth            int     `json:"queueDepth"`
	QueueCapacity         int     `json:"queueCapacity"`
	Workers               int     `json:"workers"`
	EnqueuedTotal         uint64  `json:"enqueuedTotal"`
	DroppedTotal          uint64  `json:"droppedTotal"`
	ProcessedTotal        uint64  `json:"processedTotal"`
	FailedTotal           uint64  `json:"failedTotal"`
	RetriedTotal          uint64  `json:"retriedTotal"`
	InflightWorkers       int64   `json:"inflightWorkers"`
	DropRatio             float64 `json:"dropRatio"`
	DropAlertThresholdPct float64 `json:"dropAlertThresholdPct"`
	DropAlertThresholdHit bool    `json:"dropAlertThresholdHit"`
	LastError             string  `json:"lastError"`
}

type RuntimeIncidentJournalSummary struct {
	Enabled           bool   `json:"enabled"`
	Sink              string `json:"sink"`
	Configured        bool   `json:"configured"`
	Reachable         bool   `json:"reachable"`
	TotalEvents       uint64 `json:"totalEvents"`
	ActiveEvents      uint64 `json:"activeEvents"`
	LatestEventAt     string `json:"latestEventAt"`
	LastEventStatus   string `json:"lastEventStatus"`
	DispatchFailures  uint64 `json:"dispatchFailures"`
	LastDispatchAt    string `json:"lastDispatchAt"`
	LastDispatchError string `json:"lastDispatchError"`
}

type RuntimeMetricsSummary struct {
	RequestsTotal    uint64                        `json:"requestsTotal"`
	RequestErrors    uint64                        `json:"requestErrors"`
	ErrorRate        float64                       `json:"errorRate"`
	LatencyCount     uint64                        `json:"latencyCount"`
	LatencyAverageMS float64                       `json:"latencyAverageMs"`
	LoadShedTotal    uint64                        `json:"loadShedTotal"`
	InflightCurrent  int64                         `json:"inflightCurrent"`
	InflightPeak     int64                         `json:"inflightPeak"`
	RecentHistory    []metrics.TrendPoint          `json:"recentHistory"`
	Alerts           RuntimeAlertSummary           `json:"alerts"`
	Health           RuntimeHealthSummary          `json:"health"`
	Trace            RuntimeTraceSummary           `json:"trace"`
	GatewayLogger    RuntimeGatewayLoggerSummary   `json:"gatewayLogger"`
	LoggerService    RuntimeLoggerServiceSummary   `json:"loggerService"`
	IncidentJournal  RuntimeIncidentJournalSummary `json:"incidentJournal"`
	Warnings         []string                      `json:"warnings"`
}

type RuntimeAlertSummary struct {
	ActiveCount      int                `json:"activeCount"`
	HighestSeverity  string             `json:"highestSeverity"`
	RecentlyBreached bool               `json:"recentlyBreached"`
	Items            []RuntimeAlertItem `json:"items"`
}

type RuntimeAlertItem struct {
	Code              string `json:"code"`
	Severity          string `json:"severity"`
	Status            string `json:"status"`
	Source            string `json:"source"`
	Message           string `json:"message"`
	RecommendedAction string `json:"recommendedAction"`
	BreachCount       int    `json:"breachCount"`
	LastTriggeredAt   string `json:"lastTriggeredAt"`
}

type RuntimeMetricsOptions struct {
	ReadyEndpoints               *ReadyEndpoints
	LoggerEndpoint               string
	LoggerHTTPClient             *http.Client
	RequestLoggerStatsProvider   func() middleware.AsyncLogSenderStats
	IncidentEmitterStatsProvider func() RuntimeIncidentEmitterStats
}

type loggerHealthPayload struct {
	Status string `json:"status"`
}

type loggerMetricsPayload struct {
	QueueDepth            int     `json:"queueDepth"`
	QueueCapacity         int     `json:"queueCapacity"`
	Workers               int     `json:"workers"`
	EnqueuedTotal         uint64  `json:"enqueuedTotal"`
	DroppedTotal          uint64  `json:"droppedTotal"`
	ProcessedTotal        uint64  `json:"processedTotal"`
	FailedTotal           uint64  `json:"failedTotal"`
	RetriedTotal          uint64  `json:"retriedTotal"`
	InflightWorkers       int64   `json:"inflightWorkers"`
	DropRatio             float64 `json:"dropRatio"`
	DropAlertThresholdPct float64 `json:"dropAlertThresholdPct"`
	DropAlertThresholdHit bool    `json:"dropAlertThresholdHit"`
}

type loggerIncidentSummaryPayload struct {
	TotalEvents     uint64 `json:"totalEvents"`
	ActiveEvents    uint64 `json:"activeEvents"`
	LatestEventAt   string `json:"latestEventAt"`
	LastEventStatus string `json:"lastEventStatus"`
}

func RuntimeMetricsHandler(store *metrics.Store, options RuntimeMetricsOptions) http.HandlerFunc {
	options = normalizeRuntimeMetricsOptions(options)

	return func(w http.ResponseWriter, _ *http.Request) {
		httpx.WriteJSON(w, http.StatusOK, BuildRuntimeMetricsSummary(store, options))
	}
}

func normalizeRuntimeMetricsOptions(options RuntimeMetricsOptions) RuntimeMetricsOptions {
	if options.RequestLoggerStatsProvider == nil {
		options.RequestLoggerStatsProvider = middleware.CurrentAsyncLogSenderStats
	}
	if options.IncidentEmitterStatsProvider == nil {
		options.IncidentEmitterStatsProvider = func() RuntimeIncidentEmitterStats {
			return RuntimeIncidentEmitterStats{}
		}
	}
	if options.LoggerHTTPClient == nil {
		options.LoggerHTTPClient = &http.Client{Timeout: 800 * time.Millisecond}
	}
	return options
}

func BuildRuntimeMetricsSummary(store MetricsSnapshotProvider, options RuntimeMetricsOptions) RuntimeMetricsSummary {
	options = normalizeRuntimeMetricsOptions(options)
	snapshot := store.Snapshot()
	latencyAverage := 0.0
	if snapshot.LatencyCount > 0 {
		latencyAverage = snapshot.LatencySumMS / float64(snapshot.LatencyCount)
	}

	gatewayLogger := runtimeGatewayLoggerSummary(options.RequestLoggerStatsProvider())
	incidentEmitter := options.IncidentEmitterStatsProvider()
	health, loggerService, incidentJournal := collectRuntimeDiagnosticsSummaries(options, incidentEmitter)
	trace := RuntimeTraceSummary{
		Enabled:           true,
		ResponseHeader:    httpx.TraceIDHeader,
		ForwardedToLogger: gatewayLogger.Enabled && loggerService.Configured,
		StoredOnLoggerAs:  httpx.TraceIDHeader,
		StorageField:      "traceId",
	}

	warnings := buildRuntimeMetricsWarnings(health, gatewayLogger, loggerService, incidentJournal)
	alerts := buildRuntimeAlertSummary(snapshot, health, gatewayLogger, loggerService)

	return RuntimeMetricsSummary{
		RequestsTotal:    snapshot.RequestsTotal,
		RequestErrors:    snapshot.RequestErrors,
		ErrorRate:        snapshot.ErrorRate,
		LatencyCount:     snapshot.LatencyCount,
		LatencyAverageMS: latencyAverage,
		LoadShedTotal:    snapshot.LoadShedTotal,
		InflightCurrent:  snapshot.InflightCurrent,
		InflightPeak:     snapshot.InflightPeak,
		RecentHistory:    snapshot.RecentHistory,
		Alerts:           alerts,
		Health:           health,
		Trace:            trace,
		GatewayLogger:    gatewayLogger,
		LoggerService:    loggerService,
		IncidentJournal:  incidentJournal,
		Warnings:         warnings,
	}
}

func collectRuntimeDiagnosticsSummaries(
	options RuntimeMetricsOptions,
	incidentEmitter RuntimeIncidentEmitterStats,
) (RuntimeHealthSummary, RuntimeLoggerServiceSummary, RuntimeIncidentJournalSummary) {
	var (
		health          RuntimeHealthSummary
		loggerService   RuntimeLoggerServiceSummary
		incidentJournal RuntimeIncidentJournalSummary
		wg              sync.WaitGroup
	)

	wg.Add(3)

	go func() {
		defer wg.Done()
		health = runtimeHealthSummary(options.ReadyEndpoints)
	}()

	go func() {
		defer wg.Done()
		loggerService = runtimeLoggerServiceSummary(options.LoggerHTTPClient, options.LoggerEndpoint)
	}()

	go func() {
		defer wg.Done()
		incidentJournal = runtimeIncidentJournalSummary(options.LoggerHTTPClient, options.LoggerEndpoint, incidentEmitter)
	}()

	wg.Wait()
	return health, loggerService, incidentJournal
}

func runtimeGatewayLoggerSummary(stats middleware.AsyncLogSenderStats) RuntimeGatewayLoggerSummary {
	return RuntimeGatewayLoggerSummary{
		Enabled:       stats.Enabled,
		Endpoint:      stats.Endpoint,
		QueueDepth:    stats.QueueDepth,
		QueueCapacity: stats.QueueCapacity,
		Workers:       stats.Workers,
		RetryMax:      stats.RetryMax,
		DroppedTotal:  stats.DroppedTotal,
	}
}

func runtimeHealthSummary(ready *ReadyEndpoints) RuntimeHealthSummary {
	summary := RuntimeHealthSummary{
		Status:            "unknown",
		HTTPStatus:        http.StatusServiceUnavailable,
		Postgres:          "unknown",
		Redis:             "unknown",
		Worker:            "unknown",
		CacheState:        "empty",
		CacheTTLMS:        0,
		StaleIfErrorTTLMS: 0,
	}
	if ready == nil {
		return summary
	}

	now := time.Now()
	summary.CacheTTLMS = ready.cacheTTL.Milliseconds()
	summary.StaleIfErrorTTLMS = ready.staleIfError.Milliseconds()

	ready.cache.mu.Lock()
	cached := ready.cache.snap
	ready.cache.mu.Unlock()
	if cached.payload != nil {
		summary.CacheAgeMS = now.Sub(cached.createdAt).Milliseconds()
		summary.LastCheckedAt = cached.createdAt.UTC().Format(time.RFC3339Nano)
		if now.Before(cached.expiresAt) {
			summary.CacheState = "fresh"
		} else if ready.staleIfError > 0 && now.Sub(cached.expiresAt) <= ready.staleIfError {
			summary.CacheState = "stale"
		} else {
			summary.CacheState = "expired"
		}
	}

	if ready.readinessProbe == nil {
		return summary
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer cancel()

	httpStatus, payload, _ := ready.readinessProbe(ctx, ready.workerClient)
	summary.HTTPStatus = httpStatus
	if status, _ := payload["status"].(string); status != "" {
		summary.Status = status
	}
	if checks, ok := payload["checks"].(map[string]string); ok {
		summary.Postgres = checks["postgres"]
		summary.Redis = checks["redis"]
		summary.Worker = checks["worker"]
		return summary
	}
	if checks, ok := payload["checks"].(map[string]any); ok {
		if postgres, _ := checks["postgres"].(string); postgres != "" {
			summary.Postgres = postgres
		}
		if redis, _ := checks["redis"].(string); redis != "" {
			summary.Redis = redis
		}
		if worker, _ := checks["worker"].(string); worker != "" {
			summary.Worker = worker
		}
	}

	return summary
}

func runtimeLoggerServiceSummary(client *http.Client, loggerEndpoint string) RuntimeLoggerServiceSummary {
	summary := RuntimeLoggerServiceSummary{
		Configured: false,
		Reachable:  false,
	}
	baseURL := deriveLoggerBaseURL(loggerEndpoint)
	if baseURL == "" {
		return summary
	}

	summary.Configured = true
	summary.EndpointBase = baseURL

	ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer cancel()

	healthReq, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/health", nil)
	if err != nil {
		summary.LastError = err.Error()
		return summary
	}
	healthResp, err := client.Do(healthReq)
	if err != nil {
		summary.LastError = err.Error()
		return summary
	}
	defer healthResp.Body.Close()

	if healthResp.StatusCode < http.StatusOK || healthResp.StatusCode >= http.StatusMultipleChoices {
		summary.LastError = "logger health returned non-2xx"
		return summary
	}

	var healthPayload loggerHealthPayload
	if err := json.NewDecoder(healthResp.Body).Decode(&healthPayload); err != nil {
		summary.LastError = err.Error()
		return summary
	}

	summary.Reachable = true
	summary.HealthStatus = healthPayload.Status

	metricsReq, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/metrics", nil)
	if err != nil {
		summary.LastError = err.Error()
		return summary
	}
	metricsResp, err := client.Do(metricsReq)
	if err != nil {
		summary.LastError = err.Error()
		return summary
	}
	defer metricsResp.Body.Close()

	if metricsResp.StatusCode < http.StatusOK || metricsResp.StatusCode >= http.StatusMultipleChoices {
		summary.LastError = "logger metrics returned non-2xx"
		return summary
	}

	var metricsPayload loggerMetricsPayload
	if err := json.NewDecoder(metricsResp.Body).Decode(&metricsPayload); err != nil {
		summary.LastError = err.Error()
		return summary
	}

	summary.QueueDepth = metricsPayload.QueueDepth
	summary.QueueCapacity = metricsPayload.QueueCapacity
	summary.Workers = metricsPayload.Workers
	summary.EnqueuedTotal = metricsPayload.EnqueuedTotal
	summary.DroppedTotal = metricsPayload.DroppedTotal
	summary.ProcessedTotal = metricsPayload.ProcessedTotal
	summary.FailedTotal = metricsPayload.FailedTotal
	summary.RetriedTotal = metricsPayload.RetriedTotal
	summary.InflightWorkers = metricsPayload.InflightWorkers
	summary.DropRatio = metricsPayload.DropRatio
	summary.DropAlertThresholdPct = metricsPayload.DropAlertThresholdPct
	summary.DropAlertThresholdHit = metricsPayload.DropAlertThresholdHit
	return summary
}

func deriveLoggerBaseURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return ""
	}

	path := strings.TrimSuffix(parsed.Path, "/")
	path = strings.TrimSuffix(path, "/ingest")
	parsed.Path = path
	parsed.RawPath = ""
	parsed.RawQuery = ""
	parsed.Fragment = ""
	base := strings.TrimSuffix(parsed.String(), "/")
	return base
}

func buildRuntimeMetricsWarnings(
	health RuntimeHealthSummary,
	gatewayLogger RuntimeGatewayLoggerSummary,
	loggerService RuntimeLoggerServiceSummary,
	incidentJournal RuntimeIncidentJournalSummary,
) []string {
	warnings := make([]string, 0, 5)
	if health.Status != "ok" {
		warnings = append(warnings, "runtime health is degraded")
	}
	if gatewayLogger.Enabled && gatewayLogger.DroppedTotal > 0 {
		warnings = append(warnings, "gateway async logger queue has dropped entries")
	}
	if loggerService.Configured && !loggerService.Reachable {
		warnings = append(warnings, "logger service is configured but unreachable")
	}
	if loggerService.DropAlertThresholdHit {
		warnings = append(warnings, "logger queue drop alert threshold is hit")
	}
	if incidentJournal.DispatchFailures > 0 {
		warnings = append(warnings, "incident journal dispatch has recent failures")
	}
	return warnings
}

func runtimeIncidentJournalSummary(client *http.Client, loggerEndpoint string, emitter RuntimeIncidentEmitterStats) RuntimeIncidentJournalSummary {
	summary := RuntimeIncidentJournalSummary{
		Enabled:           emitter.Enabled,
		Sink:              emitter.Sink,
		DispatchFailures:  emitter.DispatchFailures,
		LastDispatchAt:    emitter.LastDispatchAt,
		LastDispatchError: emitter.LastDispatchError,
	}
	if !strings.Contains(emitter.Sink, "logger") {
		return summary
	}

	baseURL := deriveLoggerBaseURL(loggerEndpoint)
	if baseURL == "" {
		return summary
	}
	summary.Configured = true

	ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/incident-events/summary", nil)
	if err != nil {
		summary.LastDispatchError = err.Error()
		return summary
	}
	resp, err := client.Do(req)
	if err != nil {
		summary.LastDispatchError = err.Error()
		return summary
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		summary.LastDispatchError = "incident journal summary returned non-2xx"
		return summary
	}

	var payload loggerIncidentSummaryPayload
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		summary.LastDispatchError = err.Error()
		return summary
	}
	summary.Reachable = true
	summary.TotalEvents = payload.TotalEvents
	summary.ActiveEvents = payload.ActiveEvents
	summary.LatestEventAt = payload.LatestEventAt
	summary.LastEventStatus = payload.LastEventStatus
	return summary
}

const (
	runtimeAlertSeverityInfo     = "info"
	runtimeAlertSeverityWarning  = "warning"
	runtimeAlertSeverityCritical = "critical"
)

func buildRuntimeAlertSummary(
	snapshot metrics.Snapshot,
	health RuntimeHealthSummary,
	gatewayLogger RuntimeGatewayLoggerSummary,
	loggerService RuntimeLoggerServiceSummary,
) RuntimeAlertSummary {
	items := make([]RuntimeAlertItem, 0, 6)

	if item := metricAlertFromHistory(
		"gateway.error_rate",
		"gateway",
		"error rate stayed above 5% in recent samples",
		"inspect recent 5xx responses and dependent services",
		runtimeAlertSeverityCritical,
		snapshot.RecentHistory,
		func(point metrics.TrendPoint) bool { return point.ErrorRate >= 0.05 && point.RequestsTotal >= 20 },
	); item.Code != "" {
		items = append(items, item)
	}

	if item := metricAlertFromHistory(
		"gateway.latency",
		"gateway",
		"average latency stayed above 500 ms in recent samples",
		"check downstream latency, retries, and load profile",
		runtimeAlertSeverityWarning,
		snapshot.RecentHistory,
		func(point metrics.TrendPoint) bool { return point.LatencyAverageMS >= 500 && point.RequestsTotal >= 10 },
	); item.Code != "" {
		items = append(items, item)
	}

	if item := metricAlertFromHistory(
		"gateway.load_shed",
		"gateway",
		"load shedding was observed in recent samples",
		"review max in-flight limits and profile capacity settings",
		runtimeAlertSeverityCritical,
		snapshot.RecentHistory,
		func(point metrics.TrendPoint) bool { return point.LoadShedTotal > 0 },
	); item.Code != "" {
		items = append(items, item)
	}

	if item := healthAlertItem(health); item.Code != "" {
		items = append(items, item)
	}
	if item := gatewayLoggerAlertItem(gatewayLogger); item.Code != "" {
		items = append(items, item)
	}
	if item := loggerServiceAlertItem(loggerService); item.Code != "" {
		items = append(items, item)
	}

	highest := runtimeAlertSeverityInfo
	activeCount := 0
	recentlyBreached := false
	for _, item := range items {
		if item.Status == "active" {
			activeCount++
			highest = higherSeverity(highest, item.Severity)
			recentlyBreached = true
			continue
		}
		if item.BreachCount > 0 {
			recentlyBreached = true
		}
	}

	return RuntimeAlertSummary{
		ActiveCount:      activeCount,
		HighestSeverity:  highest,
		RecentlyBreached: recentlyBreached,
		Items:            items,
	}
}

func metricAlertFromHistory(
	code string,
	source string,
	message string,
	recommendedAction string,
	severity string,
	history []metrics.TrendPoint,
	matches func(metrics.TrendPoint) bool,
) RuntimeAlertItem {
	breachCount := 0
	lastTriggeredAt := ""
	for _, point := range history {
		if !matches(point) {
			continue
		}
		breachCount++
		lastTriggeredAt = point.RecordedAt
	}
	if breachCount == 0 {
		return RuntimeAlertItem{}
	}

	status := "recent"
	if len(history) > 0 && matches(history[len(history)-1]) {
		status = "active"
	}

	return RuntimeAlertItem{
		Code:              code,
		Severity:          severity,
		Status:            status,
		Source:            source,
		Message:           message,
		RecommendedAction: recommendedAction,
		BreachCount:       breachCount,
		LastTriggeredAt:   lastTriggeredAt,
	}
}

func healthAlertItem(health RuntimeHealthSummary) RuntimeAlertItem {
	if health.Status == "ok" {
		return RuntimeAlertItem{}
	}
	return RuntimeAlertItem{
		Code:              "health.degraded",
		Severity:          runtimeAlertSeverityCritical,
		Status:            "active",
		Source:            "health",
		Message:           "runtime health probe is degraded",
		RecommendedAction: "inspect postgres, redis, and worker readiness checks",
		BreachCount:       1,
		LastTriggeredAt:   health.LastCheckedAt,
	}
}

func gatewayLoggerAlertItem(gatewayLogger RuntimeGatewayLoggerSummary) RuntimeAlertItem {
	if !gatewayLogger.Enabled || gatewayLogger.DroppedTotal == 0 {
		return RuntimeAlertItem{}
	}
	return RuntimeAlertItem{
		Code:              "gateway.logger_drop",
		Severity:          runtimeAlertSeverityWarning,
		Status:            "active",
		Source:            "gateway-logger",
		Message:           "gateway async logger queue has dropped entries",
		RecommendedAction: "increase queue capacity or inspect logger backpressure",
		BreachCount:       1,
	}
}

func loggerServiceAlertItem(loggerService RuntimeLoggerServiceSummary) RuntimeAlertItem {
	if loggerService.Configured && !loggerService.Reachable {
		return RuntimeAlertItem{
			Code:              "logger.unreachable",
			Severity:          runtimeAlertSeverityCritical,
			Status:            "active",
			Source:            "logger-service",
			Message:           "logger service is configured but unreachable",
			RecommendedAction: "verify logger container health, route, and shared secret settings",
			BreachCount:       1,
		}
	}
	if loggerService.DropAlertThresholdHit {
		return RuntimeAlertItem{
			Code:              "logger.drop_threshold",
			Severity:          runtimeAlertSeverityCritical,
			Status:            "active",
			Source:            "logger-service",
			Message:           "logger queue drop alert threshold is hit",
			RecommendedAction: "inspect logger throughput and queue pressure immediately",
			BreachCount:       1,
		}
	}
	return RuntimeAlertItem{}
}

func higherSeverity(current string, candidate string) string {
	order := map[string]int{
		runtimeAlertSeverityInfo:     0,
		runtimeAlertSeverityWarning:  1,
		runtimeAlertSeverityCritical: 2,
	}
	if order[candidate] > order[current] {
		return candidate
	}
	return current
}

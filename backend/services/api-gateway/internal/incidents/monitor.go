package incidents

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/example/appfoundrylab/backend/services/api-gateway/internal/handlers"
)

type ReportBuilder func() handlers.RuntimeReportSummary

type alertState struct {
	FirstSeenAt string
	LastSeenAt  string
	LastStatus  string
	BreachCount int
	LastEmitAt  time.Time
}

type emitterStatsSnapshot struct {
	Enabled           bool
	Sink              string
	DispatchFailures  uint64
	LastDispatchAt    string
	LastDispatchError string
}

type Monitor struct {
	reportBuilder       ReportBuilder
	loggerEndpoint      string
	webhookURL          string
	loggerSharedSecret  string
	webhookHMACSecret   string
	sink                string
	enabledSinks        map[string]bool
	webhookAllowedHosts map[string]struct{}
	interval            time.Duration
	dedupeWindow        time.Duration
	client              *http.Client

	cancel context.CancelFunc
	wg     sync.WaitGroup

	mu     sync.Mutex
	states map[string]alertState

	dispatchFailures  atomic.Uint64
	statsValue        atomic.Value
	lastDispatchAt    string
	lastDispatchError string
}

func NewMonitor(reportBuilder ReportBuilder, loggerEndpoint, webhookURL, loggerSharedSecret, webhookHMACSecret, sink string, webhookAllowedHosts []string, interval, dedupeWindow time.Duration, client *http.Client) *Monitor {
	if sink == "" || sink == "disabled" || reportBuilder == nil {
		stats := emitterStatsSnapshot{Enabled: false, Sink: sink}
		currentEmitterStats.Store(stats)
		return nil
	}
	if interval <= 0 {
		interval = 10 * time.Second
	}
	if dedupeWindow <= 0 {
		dedupeWindow = 5 * time.Minute
	}
	if client == nil {
		client = &http.Client{Timeout: 1500 * time.Millisecond}
	}

	monitor := &Monitor{
		reportBuilder:       reportBuilder,
		loggerEndpoint:      loggerEndpoint,
		webhookURL:          webhookURL,
		loggerSharedSecret:  loggerSharedSecret,
		webhookHMACSecret:   webhookHMACSecret,
		sink:                sink,
		enabledSinks:        make(map[string]bool),
		interval:            interval,
		dedupeWindow:        dedupeWindow,
		client:              client,
		states:              make(map[string]alertState),
		webhookAllowedHosts: make(map[string]struct{}, len(webhookAllowedHosts)),
	}
	for _, host := range webhookAllowedHosts {
		host = strings.ToLower(strings.TrimSpace(host))
		if host == "" {
			continue
		}
		monitor.webhookAllowedHosts[host] = struct{}{}
	}
	for _, item := range strings.FieldsFunc(sink, func(r rune) bool { return r == '+' || r == ',' }) {
		monitor.enabledSinks[strings.TrimSpace(item)] = true
	}
	monitor.publishStats()
	return monitor
}

func (m *Monitor) Start(parent context.Context) {
	if m == nil {
		return
	}
	ctx, cancel := context.WithCancel(parent)
	m.cancel = cancel
	m.wg.Add(1)
	go m.loop(ctx)
}

func (m *Monitor) Close() {
	if m == nil {
		return
	}
	if m.cancel != nil {
		m.cancel()
	}
	m.wg.Wait()
}

func CurrentEmitterStats() handlers.RuntimeIncidentEmitterStats {
	stats, _ := currentEmitterStats.Load().(emitterStatsSnapshot)
	return handlers.RuntimeIncidentEmitterStats{
		Enabled:           stats.Enabled,
		Sink:              stats.Sink,
		DispatchFailures:  stats.DispatchFailures,
		LastDispatchAt:    stats.LastDispatchAt,
		LastDispatchError: stats.LastDispatchError,
	}
}

var currentEmitterStats atomic.Value

func init() {
	currentEmitterStats.Store(emitterStatsSnapshot{})
}

func (m *Monitor) loop(ctx context.Context) {
	defer m.wg.Done()
	m.process()
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.process()
		}
	}
}

func (m *Monitor) process() {
	report := m.reportBuilder()
	now := time.Now().UTC()

	m.mu.Lock()
	defer m.mu.Unlock()

	seenCodes := make(map[string]bool, len(report.Metrics.Alerts.Items))
	for _, alert := range report.Metrics.Alerts.Items {
		seenCodes[alert.Code] = true
		state := m.states[alert.Code]
		previousBreachCount := state.BreachCount
		if state.FirstSeenAt == "" {
			state.FirstSeenAt = firstNonEmpty(alert.LastTriggeredAt, now.Format(time.RFC3339Nano))
		}
		state.LastSeenAt = firstNonEmpty(alert.LastTriggeredAt, now.Format(time.RFC3339Nano))

		eventType := ""
		if alert.Status == "active" && state.LastStatus != "active" {
			eventType = "opened"
		} else if alert.Status == "active" && now.Sub(state.LastEmitAt) >= m.dedupeWindow && alert.BreachCount > previousBreachCount {
			eventType = "updated"
		} else if alert.Status == "recent" && state.LastStatus == "" {
			eventType = "opened"
		}

		if eventType != "" {
			m.dispatchEvent(buildIncidentEventRecord(report, alert, state, eventType))
			state.LastEmitAt = now
		}
		state.BreachCount = alert.BreachCount
		state.LastStatus = alert.Status
		m.states[alert.Code] = state
	}

	for code, state := range m.states {
		if seenCodes[code] {
			continue
		}
		if state.LastStatus == "active" || state.LastStatus == "recent" {
			resolvedAlert := handlers.RuntimeAlertItem{
				Code:              code,
				Status:            "resolved",
				Severity:          "info",
				Source:            "runtime",
				Message:           "alert is no longer active in the latest runtime snapshot",
				RecommendedAction: "confirm the related service has stabilized and close the incident if no regressions remain",
				BreachCount:       state.BreachCount,
				LastTriggeredAt:   state.LastSeenAt,
			}
			m.dispatchEvent(buildIncidentEventRecord(report, resolvedAlert, state, "resolved"))
			state.LastStatus = "resolved"
			state.LastEmitAt = now
			m.states[code] = state
		}
	}
}

func buildIncidentEventRecord(
	report handlers.RuntimeReportSummary,
	alert handlers.RuntimeAlertItem,
	state alertState,
	eventType string,
) handlers.RuntimeIncidentEventRecord {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	return handlers.RuntimeIncidentEventRecord{
		ID:                  buildIncidentEventID(alert.Code, eventType, now),
		EventType:           eventType,
		AlertCode:           alert.Code,
		Severity:            alert.Severity,
		Status:              alert.Status,
		Source:              alert.Source,
		Title:               report.Incident.Title,
		Summary:             report.Incident.Summary,
		Message:             alert.Message,
		RecommendedAction:   alert.RecommendedAction,
		RecommendedSeverity: report.Incident.RecommendedSeverity,
		TriggeredAt:         now,
		FirstSeenAt:         firstNonEmpty(state.FirstSeenAt, now),
		LastSeenAt:          firstNonEmpty(alert.LastTriggeredAt, now),
		BreachCount:         alert.BreachCount,
		ReportGeneratedAt:   report.GeneratedAt,
		ReportVersion:       report.ReportVersion,
		Runbooks:            handlers.RunbookReferencesForAlertCode(alert.Code),
	}
}

func buildIncidentEventID(code, eventType, now string) string {
	safeCode := strings.ReplaceAll(code, ".", "-")
	return safeCode + "-" + eventType + "-" + strings.ReplaceAll(strings.ReplaceAll(now, ":", "-"), ".", "-")
}

func (m *Monitor) dispatchEvent(event handlers.RuntimeIncidentEventRecord) {
	if len(m.sink) == 0 {
		return
	}

	payload, err := json.Marshal(event)
	if err != nil {
		m.dispatchFailures.Add(1)
		m.lastDispatchError = err.Error()
		return
	}

	if sinkIncludes(m.sink, "logger") {
		if err := m.dispatchToLogger(payload); err != nil {
			m.dispatchFailures.Add(1)
			m.lastDispatchError = err.Error()
		}
	}
	if sinkIncludes(m.sink, "stdout") {
		log.Printf("incident_event %s", payload)
	}
	if sinkIncludes(m.sink, "webhook") {
		if err := m.dispatchToWebhook(payload); err != nil {
			m.dispatchFailures.Add(1)
			m.lastDispatchError = err.Error()
		}
	}
	m.lastDispatchAt = time.Now().UTC().Format(time.RFC3339Nano)
	m.publishStats()
}

func (m *Monitor) dispatchToLogger(body []byte) error {
	baseURL := deriveLoggerBaseURL(m.loggerEndpoint)
	if baseURL == "" {
		return nil
	}
	timestamp := time.Now().UTC().Format(time.RFC3339Nano)
	req, err := http.NewRequest(http.MethodPost, baseURL+"/incident-events", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if m.loggerSharedSecret != "" {
		req.Header.Set("X-Logger-Timestamp", timestamp)
		req.Header.Set("X-Logger-Signature", signLoggerPayload(m.loggerSharedSecret, timestamp, body))
	}
	resp, err := m.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return &dispatchError{status: resp.StatusCode}
	}
	return nil
}

func (m *Monitor) dispatchToWebhook(body []byte) error {
	if strings.TrimSpace(m.webhookURL) == "" {
		return nil
	}
	parsed, err := url.Parse(m.webhookURL)
	if err != nil {
		return err
	}
	if len(m.webhookAllowedHosts) > 0 {
		if _, ok := m.webhookAllowedHosts[strings.ToLower(parsed.Hostname())]; !ok {
			return errors.New("incident webhook host is not allowed")
		}
	}
	req, err := http.NewRequest(http.MethodPost, m.webhookURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if m.webhookHMACSecret != "" {
		timestamp := time.Now().UTC().Format(time.RFC3339Nano)
		req.Header.Set("X-Incident-Event-Timestamp", timestamp)
		req.Header.Set("X-Incident-Event-Signature", signLoggerPayload(m.webhookHMACSecret, timestamp, body))
	}
	resp, err := m.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return &dispatchError{status: resp.StatusCode}
	}
	return nil
}

type dispatchError struct {
	status int
}

func (e *dispatchError) Error() string {
	return "incident dispatch failed with status " + http.StatusText(e.status)
}

func (m *Monitor) publishStats() {
	currentEmitterStats.Store(emitterStatsSnapshot{
		Enabled:           true,
		Sink:              m.sink,
		DispatchFailures:  m.dispatchFailures.Load(),
		LastDispatchAt:    m.lastDispatchAt,
		LastDispatchError: m.lastDispatchError,
	})
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
	return strings.TrimSuffix(parsed.String(), "/")
}

func signLoggerPayload(secret, timestamp string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp))
	mac.Write([]byte("\n"))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

func firstNonEmpty(candidates ...string) string {
	for _, candidate := range candidates {
		if strings.TrimSpace(candidate) != "" {
			return candidate
		}
	}
	return ""
}

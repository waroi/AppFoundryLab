package handlers

import (
	"fmt"
	"strings"
)

const runtimeReportVersion = "2026-03-incident-v1"

type RuntimeRunbookReference struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Path     string `json:"path"`
	Reason   string `json:"reason"`
	Priority string `json:"priority"`
}

type RuntimeIncidentEvidence struct {
	Kind   string `json:"kind"`
	Label  string `json:"label"`
	Value  string `json:"value"`
	Source string `json:"source"`
}

type RuntimeIncidentSummary struct {
	RecommendedSeverity string                    `json:"recommendedSeverity"`
	Category            string                    `json:"category"`
	Title               string                    `json:"title"`
	Summary             string                    `json:"summary"`
	SuspectedSystems    []string                  `json:"suspectedSystems"`
	TriggeredAlerts     []string                  `json:"triggeredAlerts"`
	NextActions         []string                  `json:"nextActions"`
	Evidence            []RuntimeIncidentEvidence `json:"evidence"`
}

type RuntimeIncidentEventRecord struct {
	ID                  string                    `json:"id"`
	EventType           string                    `json:"eventType"`
	AlertCode           string                    `json:"alertCode"`
	Severity            string                    `json:"severity"`
	Status              string                    `json:"status"`
	Source              string                    `json:"source"`
	Title               string                    `json:"title"`
	Summary             string                    `json:"summary"`
	Message             string                    `json:"message"`
	RecommendedAction   string                    `json:"recommendedAction"`
	RecommendedSeverity string                    `json:"recommendedSeverity"`
	TriggeredAt         string                    `json:"triggeredAt"`
	FirstSeenAt         string                    `json:"firstSeenAt"`
	LastSeenAt          string                    `json:"lastSeenAt"`
	BreachCount         int                       `json:"breachCount"`
	TraceID             string                    `json:"traceId"`
	ReportGeneratedAt   string                    `json:"reportGeneratedAt"`
	ReportVersion       string                    `json:"reportVersion"`
	Runbooks            []RuntimeRunbookReference `json:"runbooks"`
}

type RuntimeIncidentEventsResponse struct {
	Items []RuntimeIncidentEventRecord `json:"items"`
}

type RuntimeIncidentEmitterStats struct {
	Enabled           bool   `json:"enabled"`
	Sink              string `json:"sink"`
	DispatchFailures  uint64 `json:"dispatchFailures"`
	LastDispatchAt    string `json:"lastDispatchAt"`
	LastDispatchError string `json:"lastDispatchError"`
}

func BuildRuntimeIncidentSummary(config RuntimeConfigSummary, metrics RuntimeMetricsSummary) RuntimeIncidentSummary {
	runbooks := BuildRuntimeRunbookReferences(metrics.Alerts)
	nextActions := make([]string, 0, len(metrics.Alerts.Items)+1)
	triggered := make([]string, 0, len(metrics.Alerts.Items))
	suspected := make([]string, 0, 4)
	evidence := make([]RuntimeIncidentEvidence, 0, 6)

	for _, alert := range metrics.Alerts.Items {
		triggered = append(triggered, alert.Code)
		nextActions = appendIfMissing(nextActions, alert.RecommendedAction)
		suspected = appendIfMissing(suspected, suspectedSystemForAlert(alert.Code))
		if alert.LastTriggeredAt != "" {
			evidence = append(evidence, RuntimeIncidentEvidence{
				Kind:   "alert",
				Label:  alert.Code,
				Value:  fmt.Sprintf("%s at %s", alert.Status, alert.LastTriggeredAt),
				Source: alert.Source,
			})
		}
	}
	if metrics.Health.Status != "" {
		evidence = append(evidence, RuntimeIncidentEvidence{
			Kind:   "health",
			Label:  "runtime-health",
			Value:  fmt.Sprintf("%s (%d)", metrics.Health.Status, metrics.Health.HTTPStatus),
			Source: "health",
		})
	}
	if metrics.LoggerService.LastError != "" {
		evidence = append(evidence, RuntimeIncidentEvidence{
			Kind:   "logger",
			Label:  "logger-last-error",
			Value:  metrics.LoggerService.LastError,
			Source: "logger-service",
		})
	}

	if len(runbooks) > 0 {
		nextActions = appendIfMissing(nextActions, "open the mapped runbook and execute the first response checklist")
	}
	if metrics.Alerts.ActiveCount == 0 && len(metrics.Warnings) == 0 {
		nextActions = appendIfMissing(nextActions, "no active incident response is required; keep monitoring recent trend history")
	}
	if config.Security.DefaultCredentialsInUse {
		nextActions = appendIfMissing(nextActions, "rotate bootstrap credentials before promoting this profile")
	}

	return RuntimeIncidentSummary{
		RecommendedSeverity: recommendedIncidentSeverity(metrics.Alerts),
		Category:            incidentCategory(metrics.Alerts),
		Title:               incidentTitle(metrics.Alerts),
		Summary:             incidentSummaryText(metrics, runbooks),
		SuspectedSystems:    suspected,
		TriggeredAlerts:     triggered,
		NextActions:         nextActions,
		Evidence:            evidence,
	}
}

func BuildRuntimeRunbookReferences(alerts RuntimeAlertSummary) []RuntimeRunbookReference {
	refs := make([]RuntimeRunbookReference, 0, len(alerts.Items))
	for _, alert := range alerts.Items {
		for _, ref := range RunbookReferencesForAlertCode(alert.Code) {
			if containsRunbookReference(refs, ref.ID) {
				continue
			}
			refs = append(refs, ref)
		}
	}
	return refs
}

func RunbookReferencesForAlertCode(code string) []RuntimeRunbookReference {
	switch code {
	case "gateway.load_shed":
		return []RuntimeRunbookReference{{
			ID:       "load-shedding",
			Title:    "Load Shedding Runbook",
			Path:     "docs/load-shedding-runbook.md",
			Reason:   "Gateway is actively rejecting requests under pressure.",
			Priority: "high",
		}}
	case "gateway.error_rate":
		return []RuntimeRunbookReference{{
			ID:       "api-degradation",
			Title:    "API Degradation Runbook",
			Path:     "docs/api-degradation-runbook.md",
			Reason:   "Gateway 5xx error rate crossed the incident threshold.",
			Priority: "high",
		}}
	case "gateway.latency":
		return []RuntimeRunbookReference{{
			ID:       "latency-regression",
			Title:    "Latency Regression Runbook",
			Path:     "docs/latency-regression-runbook.md",
			Reason:   "Gateway latency stayed above the warning threshold.",
			Priority: "medium",
		}}
	case "health.degraded":
		return []RuntimeRunbookReference{{
			ID:       "dependency-degradation",
			Title:    "Dependency Degradation Runbook",
			Path:     "docs/dependency-degradation-runbook.md",
			Reason:   "Readiness checks report a degraded dependency.",
			Priority: "high",
		}}
	case "logger.unreachable", "logger.drop_threshold", "gateway.logger_drop":
		return []RuntimeRunbookReference{{
			ID:       "logger-pipeline",
			Title:    "Logger Pipeline Runbook",
			Path:     "docs/logger-pipeline-runbook.md",
			Reason:   "The logging pipeline is unreachable or dropping events.",
			Priority: "high",
		}}
	default:
		return nil
	}
}

func recommendedIncidentSeverity(alerts RuntimeAlertSummary) string {
	if alerts.ActiveCount == 0 && !alerts.RecentlyBreached {
		return "sev-4"
	}
	if alerts.HighestSeverity == runtimeAlertSeverityCritical {
		return "sev-1"
	}
	if alerts.ActiveCount > 0 {
		return "sev-2"
	}
	return "sev-3"
}

func incidentCategory(alerts RuntimeAlertSummary) string {
	for _, alert := range alerts.Items {
		switch alert.Code {
		case "gateway.error_rate", "gateway.latency", "gateway.load_shed":
			return "api-runtime"
		case "health.degraded":
			return "dependency-health"
		case "logger.unreachable", "logger.drop_threshold", "gateway.logger_drop":
			return "observability"
		}
	}
	return "informational"
}

func incidentTitle(alerts RuntimeAlertSummary) string {
	if alerts.ActiveCount == 0 && !alerts.RecentlyBreached {
		return "Runtime snapshot has no active incidents"
	}
	if alerts.HighestSeverity == runtimeAlertSeverityCritical {
		return "Critical runtime incident detected"
	}
	if alerts.ActiveCount > 0 {
		return "Runtime warning requires operator review"
	}
	return "Recent runtime breach requires follow-up"
}

func incidentSummaryText(metrics RuntimeMetricsSummary, runbooks []RuntimeRunbookReference) string {
	parts := []string{
		fmt.Sprintf("%d active alert(s)", metrics.Alerts.ActiveCount),
		fmt.Sprintf("highest severity %s", metrics.Alerts.HighestSeverity),
		fmt.Sprintf("health %s", metrics.Health.Status),
	}
	if len(runbooks) > 0 {
		parts = append(parts, fmt.Sprintf("%d runbook(s) mapped", len(runbooks)))
	}
	return strings.Join(parts, ", ")
}

func suspectedSystemForAlert(code string) string {
	switch code {
	case "gateway.error_rate", "gateway.latency", "gateway.load_shed":
		return "api-gateway"
	case "health.degraded":
		return "dependencies"
	case "logger.unreachable", "logger.drop_threshold", "gateway.logger_drop":
		return "logger-service"
	default:
		return "runtime"
	}
}

func containsRunbookReference(refs []RuntimeRunbookReference, id string) bool {
	for _, ref := range refs {
		if ref.ID == id {
			return true
		}
	}
	return false
}

func appendIfMissing(items []string, item string) []string {
	item = strings.TrimSpace(item)
	if item == "" {
		return items
	}
	for _, existing := range items {
		if existing == item {
			return items
		}
	}
	return append(items, item)
}

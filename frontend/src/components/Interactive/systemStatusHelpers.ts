import type { RuntimeIncidentEventsResponse, RuntimeReportResponse } from "@/lib/api/types";
import { type getCopy, renderTemplate } from "@/lib/ui/copy";

export type SystemStatusCopy = ReturnType<typeof getCopy>;

export function availableTraceIds(events: RuntimeIncidentEventsResponse["items"]): string[] {
  return [...new Set(events.map((event) => event.traceId).filter(Boolean))].slice(0, 6) as string[];
}

export function buildIncidentSummaryLines(
  copy: SystemStatusCopy,
  runtimeReport: RuntimeReportResponse | null,
): string[] {
  if (!runtimeReport) {
    return [];
  }

  return [
    `${copy.systemStatus.clipboardIncident}: ${runtimeReport.incident.title}`,
    `${copy.common.severity}: ${runtimeReport.incident.recommendedSeverity}`,
    `${copy.common.category}: ${runtimeReport.incident.category}`,
    `${copy.common.summary}: ${runtimeReport.incident.summary}`,
    `${copy.systemStatus.clipboardAlerts}: ${runtimeReport.incident.triggeredAlerts.join(", ") || copy.common.none}`,
    `${copy.systemStatus.clipboardRunbooks}: ${runtimeReport.runbooks.map((runbook) => `${runbook.title} (${runbook.path})`).join(", ") || copy.common.none}`,
  ];
}

export function requestStatusTone(statusCode: number): string {
  if (statusCode >= 500) {
    return "status-badge status-danger";
  }
  if (statusCode >= 400) {
    return "status-badge status-warning";
  }
  if (statusCode >= 300) {
    return "status-badge status-neutral";
  }
  return "status-badge status-success";
}

export function traceStateText(copy: SystemStatusCopy, activeTraceQuery: string): string {
  if (activeTraceQuery) {
    return `${copy.systemStatus.traceFilter}: ${activeTraceQuery}`;
  }
  return copy.systemStatus.latestTraceSummary;
}

export function useTraceLabel(copy: SystemStatusCopy, traceId: string): string {
  return renderTemplate(copy.systemStatus.useTrace, { traceId });
}

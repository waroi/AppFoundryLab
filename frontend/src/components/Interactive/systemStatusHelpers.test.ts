import { getCopy } from "@/lib/ui/copy";
import { describe, expect, it } from "vitest";
import {
  availableTraceIds,
  buildIncidentSummaryLines,
  requestStatusTone,
  traceStateText,
} from "./systemStatusHelpers";

const copy = getCopy("en");

describe("systemStatusHelpers", () => {
  it("deduplicates trace ids and ignores blanks", () => {
    const traceIds = availableTraceIds([
      { traceId: "trace-a" },
      { traceId: "trace-a" },
      { traceId: "trace-b" },
      { traceId: "" },
    ] as never);

    expect(traceIds).toEqual(["trace-a", "trace-b"]);
  });

  it("builds incident summary lines for clipboard export", () => {
    const lines = buildIncidentSummaryLines(copy, {
      generatedAt: "2026-03-01T00:00:00Z",
      reportVersion: "v1",
      config: {} as never,
      metrics: {} as never,
      runbooks: [
        {
          id: "runbook-1",
          title: "Runbook",
          path: "docs/runbook.md",
          reason: "Reason",
          priority: "high",
        },
      ],
      incident: {
        title: "Gateway issue",
        recommendedSeverity: "sev-2",
        category: "api-runtime",
        summary: "Summary",
        suspectedSystems: [],
        triggeredAlerts: ["gateway.error_rate"],
        nextActions: [],
        evidence: [],
      },
    });

    expect(lines).toContain("Incident: Gateway issue");
    expect(lines).toContain("Severity: sev-2");
    expect(lines).toContain("Alerts: gateway.error_rate");
    expect(lines).toContain("Runbooks: Runbook (docs/runbook.md)");
  });

  it("returns trace state text for active filter or latest mode", () => {
    expect(traceStateText(copy, "")).toBe(copy.systemStatus.latestTraceSummary);
    expect(traceStateText(copy, "trace-123")).toBe("Trace filter: trace-123");
  });

  it("maps request status codes to tone classes", () => {
    expect(requestStatusTone(200)).toBe("status-badge status-success");
    expect(requestStatusTone(302)).toBe("status-badge status-neutral");
    expect(requestStatusTone(404)).toBe("status-badge status-warning");
    expect(requestStatusTone(503)).toBe("status-badge status-danger");
  });
});

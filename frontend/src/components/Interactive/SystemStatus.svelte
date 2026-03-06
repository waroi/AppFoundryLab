<script lang="ts">
import { onMount } from "svelte";
import AuthPanel from "@/components/Interactive/SystemStatus/AuthPanel.svelte";
import AdminDiagnosticsPanel from "@/components/Interactive/SystemStatus/AdminDiagnosticsPanel.svelte";
import FibonacciPanel from "@/components/Interactive/SystemStatus/FibonacciPanel.svelte";
import HealthSummary from "@/components/Interactive/SystemStatus/HealthSummary.svelte";
import ProtectedUsersPanel from "@/components/Interactive/SystemStatus/ProtectedUsersPanel.svelte";
import { fetchTyped } from "@/lib/api/fetcher";
import type {
  HealthResponse,
  RuntimeConfigResponse,
  RuntimeIncidentEventsResponse,
  RuntimeMetricsResponse,
  RuntimeReportResponse,
  RuntimeRequestLogRecord,
  UsersResponse,
} from "@/lib/api/types";
import {
  isFibonacciResponse,
  isHealthResponse,
  isRuntimeIncidentEventsResponse,
  isRuntimeReportResponse,
  isRuntimeRequestLogsResponse,
  isTokenResponse,
  isUsersResponse,
} from "@/lib/api/validators";
import {
  formatDateTime as formatLocalizedDateTime,
  getCopy,
  translateError,
} from "@/lib/ui/copy";
import { DEFAULT_LOCALE, type Locale, locale } from "@/lib/ui/preferences";
import { buildIncidentSummaryLines } from "./systemStatusHelpers";

export let initialLocale: Locale = DEFAULT_LOCALE;

let health: HealthResponse | null = null;
let users: UsersResponse["data"] = [];
let loading = true;
let healthError = "";
let authError = "";
let usersError = "";
let username = "developer";
let password = "";
let token = "";
let role = "";
let runtimeConfig: RuntimeConfigResponse | null = null;
let runtimeMetrics: RuntimeMetricsResponse | null = null;
let runtimeReport: RuntimeReportResponse | null = null;
let incidentEvents: RuntimeIncidentEventsResponse["items"] = [];
let requestLogs: RuntimeRequestLogRecord[] = [];
let requestLogsLoading = false;
let requestLogsError = "";
let traceQuery = "";
let activeTraceQuery = "";
let runtimeError = "";
let fibInput = 40;
let fibResult = "";
let fibError = "";

const REQUEST_LOG_LIMIT = 10;

$: activeLocale = typeof window === "undefined" ? initialLocale : $locale;
$: copy = getCopy(activeLocale);

onMount(async () => {
  try {
    health = await fetchTyped("/health", isHealthResponse);
  } catch (err) {
    healthError = err instanceof Error ? err.message : "unknown_error";
  } finally {
    loading = false;
  }
});

function resetAuthenticatedState(): void {
  authError = "";
  usersError = "";
  runtimeError = "";
  runtimeConfig = null;
  runtimeMetrics = null;
  runtimeReport = null;
  incidentEvents = [];
  requestLogs = [];
  requestLogsLoading = false;
  requestLogsError = "";
  traceQuery = "";
  activeTraceQuery = "";
  token = "";
  role = "";
  users = [];
}

async function loginAndLoad(): Promise<void> {
  resetAuthenticatedState();

  try {
    const auth = await fetchTyped("/api/v1/auth/token", isTokenResponse, {
      method: "POST",
      body: JSON.stringify({ username, password }),
    });

    token = auth.accessToken;
    role = auth.role;
    const usersPromise = fetchTyped("/api/v1/users", isUsersResponse, undefined, auth.accessToken);

    if (auth.role === "admin") {
      const [usersData, runtimeReportPayload, incidentEventsPayload] = await Promise.all([
        usersPromise,
        fetchTyped(
          "/api/v1/admin/runtime-report",
          isRuntimeReportResponse,
          undefined,
          auth.accessToken,
        ),
        fetchTyped(
          "/api/v1/admin/incident-events",
          isRuntimeIncidentEventsResponse,
          undefined,
          auth.accessToken,
        ),
      ]);
      users = usersData.data;
      runtimeReport = runtimeReportPayload;
      runtimeConfig = runtimeReportPayload.config;
      runtimeMetrics = runtimeReportPayload.metrics;
      incidentEvents = incidentEventsPayload.items;
      void loadRequestLogs("", auth.accessToken);
      return;
    }

    const usersData = await usersPromise;
    users = usersData.data;
  } catch (err) {
    const message = err instanceof Error ? err.message : "unknown_error";
    if (!token) {
      authError = message;
      return;
    }
    if (role === "admin") {
      runtimeError = message;
      return;
    }
    usersError = message;
  }
}

async function loadRequestLogs(traceId = "", authToken = token): Promise<void> {
  if (!authToken) {
    requestLogs = [];
    requestLogsError = "login_required";
    return;
  }

  const normalizedTraceId = traceId.trim();
  const params = new URLSearchParams({ limit: String(REQUEST_LOG_LIMIT) });
  if (normalizedTraceId) {
    params.set("traceId", normalizedTraceId);
  }

  requestLogsLoading = true;
  requestLogsError = "";
  activeTraceQuery = normalizedTraceId;
  try {
    const payload = await fetchTyped(
      `/api/v1/admin/request-logs?${params.toString()}`,
      isRuntimeRequestLogsResponse,
      undefined,
      authToken,
    );
    requestLogs = payload.items;
  } catch (err) {
    requestLogs = [];
    requestLogsError = err instanceof Error ? err.message : "unknown_error";
  } finally {
    requestLogsLoading = false;
  }
}

async function runTraceSearch(): Promise<void> {
  await loadRequestLogs(traceQuery);
}

async function clearTraceSearch(): Promise<void> {
  traceQuery = "";
  await loadRequestLogs("");
}

async function useIncidentTrace(traceId: string): Promise<void> {
  traceQuery = traceId;
  await loadRequestLogs(traceId);
}

function translateErrorMessage(value: string): string {
  return translateError(activeLocale, value);
}

async function runFibonacci(): Promise<void> {
  fibError = "";
  fibResult = "";

  if (!token) {
    fibError = "login_required";
    return;
  }

  try {
    const payload = await fetchTyped(
      "/api/v1/compute/fibonacci",
      isFibonacciResponse,
      {
        method: "POST",
        body: JSON.stringify({ n: fibInput }),
      },
      token,
    );
    fibResult = String(payload.value);
  } catch (err) {
    fibError = err instanceof Error ? err.message : "unknown_error";
  }
}

function downloadRuntimeReport(): void {
  if (!runtimeReport) {
    return;
  }

  const blob = new Blob([JSON.stringify(runtimeReport, null, 2)], { type: "application/json" });
  const url = URL.createObjectURL(blob);
  const anchor = document.createElement("a");
  anchor.href = url;
  anchor.download = `runtime-report-${runtimeReport.generatedAt.replaceAll(":", "-")}.json`;
  anchor.click();
  URL.revokeObjectURL(url);
}

function downloadIncidentReport(): void {
  if (!runtimeReport) {
    return;
  }

  const payload = {
    generatedAt: runtimeReport.generatedAt,
    reportVersion: runtimeReport.reportVersion,
    incident: runtimeReport.incident,
    runbooks: runtimeReport.runbooks,
    metrics: {
      alerts: runtimeReport.metrics.alerts,
      incidentJournal: runtimeReport.metrics.incidentJournal,
      warnings: runtimeReport.metrics.warnings,
    },
  };
  const blob = new Blob([JSON.stringify(payload, null, 2)], { type: "application/json" });
  const url = URL.createObjectURL(blob);
  const anchor = document.createElement("a");
  anchor.href = url;
  anchor.download = `incident-report-${runtimeReport.generatedAt.replaceAll(":", "-")}.json`;
  anchor.click();
  URL.revokeObjectURL(url);
}

async function copyIncidentSummary(): Promise<void> {
  if (!runtimeReport || typeof navigator === "undefined" || !navigator.clipboard) {
    return;
  }

  await navigator.clipboard.writeText(
    buildIncidentSummaryLines(copy, runtimeReport).join("\n"),
  );
}

function formatDateTime(value: string): string {
  return formatLocalizedDateTime(activeLocale, value);
}
</script>

<section class="card mt-6" data-testid="system-status-root">
  <h2 class="text-xl font-semibold">{copy.systemStatus.title}</h2>

  {#if loading}
    <p class="mt-3 text-body">{copy.systemStatus.loading}</p>
  {:else if healthError}
    <p class="mt-3 text-sm" style={`color: var(--danger-text);`} data-testid="health-error">
      {copy.common.error}: {translateErrorMessage(healthError)}
    </p>
  {:else if health}
    <HealthSummary {copy} {health} />

    <AuthPanel
      bind:username
      bind:password
      {copy}
      {activeLocale}
      {token}
      {role}
      {authError}
      {translateErrorMessage}
      onLogin={loginAndLoad}
    />

    {#if role === "admin"}
      <AdminDiagnosticsPanel
        bind:traceQuery
        {copy}
        {activeLocale}
        {runtimeError}
        {runtimeConfig}
        {runtimeMetrics}
        {runtimeReport}
        {incidentEvents}
        {requestLogs}
        {requestLogsLoading}
        {requestLogsError}
        {activeTraceQuery}
        onDownloadRuntimeReport={downloadRuntimeReport}
        onDownloadIncidentReport={downloadIncidentReport}
        onCopyIncidentSummary={copyIncidentSummary}
        onSearchTrace={runTraceSearch}
        onLatestTrace={clearTraceSearch}
        onUseTrace={useIncidentTrace}
      />
    {/if}

    <ProtectedUsersPanel {copy} {users} {usersError} {translateErrorMessage} />

    <FibonacciPanel
      bind:fibInput
      {copy}
      {fibResult}
      {fibError}
      {translateErrorMessage}
      onRunFibonacci={runFibonacci}
    />
  {/if}
</section>

<script lang="ts">
import { onMount } from "svelte";
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
  booleanLabel,
  formatDurationMs,
  formatDateTime as formatLocalizedDateTime,
  formatPercent,
  formatRole,
  getCopy,
  renderTemplate,
  translateError,
} from "@/lib/ui/copy";
import { DEFAULT_LOCALE, type Locale, locale } from "@/lib/ui/preferences";
export let initialLocale: Locale = DEFAULT_LOCALE;

let health: HealthResponse | null = null;
let users: UsersResponse["data"] = [];
let loading = true;
let error = "";
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
    error = err instanceof Error ? err.message : "unknown_error";
  } finally {
    loading = false;
  }
});

async function loginAndLoad(): Promise<void> {
  error = "";
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
    error = err instanceof Error ? err.message : "unknown_error";
    runtimeError = err instanceof Error ? err.message : "unknown_error";
    requestLogsLoading = false;
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

function availableTraceIds(): string[] {
  return [...new Set(incidentEvents.map((event) => event.traceId).filter(Boolean))].slice(
    0,
    6,
  ) as string[];
}

function translateErrorMessage(value: string): string {
  return translateError(activeLocale, value);
}

function formatDateTime(value: string): string {
  return formatLocalizedDateTime(activeLocale, value);
}

function requestStatusTone(statusCode: number): string {
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

function boolLabel(
  value: boolean,
  mode: "yesNo" | "enabledDisabled" | "onOff" | "requiredOptional",
): string {
  return booleanLabel(activeLocale, value, mode);
}

function useTraceLabel(traceId: string): string {
  return renderTemplate(copy.systemStatus.useTrace, { traceId });
}

function traceStateText(): string {
  if (activeTraceQuery) {
    return `${copy.systemStatus.traceFilter}: ${activeTraceQuery}`;
  }
  return copy.systemStatus.latestTraceSummary;
}

function copySummaryLines(): string[] {
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

  await navigator.clipboard.writeText(copySummaryLines().join("\n"));
}
</script>

<section class="card mt-6" data-testid="system-status-root">
  <h2 class="text-xl font-semibold">{copy.systemStatus.title}</h2>

  {#if loading}
    <p class="mt-3 text-body">{copy.systemStatus.loading}</p>
  {:else if error}
    <p class="mt-3 text-sm" style={`color: var(--danger-text);`}>
      {copy.common.error}: {translateErrorMessage(error)}
    </p>
  {:else if health}
    <div class="mt-3 grid gap-2 sm:grid-cols-3">
      <div class="panel-subtle">
        <p class="text-sm text-muted">{copy.systemStatus.gatewayState}</p>
        <p class="font-semibold uppercase">{health.status}</p>
      </div>
      <div class="panel-subtle">
        <p class="text-sm text-muted">{copy.systemStatus.postgresRedis}</p>
        <p class="font-semibold">{health.checks.postgres} / {health.checks.redis}</p>
      </div>
      <div class="panel-subtle">
        <p class="text-sm text-muted">{copy.systemStatus.grpcWorker}</p>
        <p class="font-semibold">{health.checks.worker}</p>
      </div>
    </div>

    <div class="mt-5 panel">
      <h3 class="font-semibold">{copy.systemStatus.authTitle}</h3>
      <p class="mt-2 text-sm text-soft">{copy.systemStatus.authHelper}</p>
      <div class="mt-2 flex flex-wrap gap-2">
        <input
          class="field"
          bind:value={username}
          placeholder={copy.systemStatus.username}
          data-testid="login-username"
        />
        <input
          class="field"
          type="password"
          bind:value={password}
          placeholder={copy.systemStatus.password}
          data-testid="login-password"
        />
        <button class="button-accent" type="button" on:click={loginAndLoad} data-testid="login-submit">
          {copy.systemStatus.loginButton}
        </button>
      </div>
      {#if token}
        <p
          class="mt-2 text-sm"
          style={`color: var(--success-text);`}
          data-testid="auth-role"
          data-role={role}
        >
          {copy.systemStatus.authenticatedAsRole}: <strong>{formatRole(activeLocale, role)}</strong>
        </p>
      {/if}
    </div>

    {#if role === "admin"}
      <div class="mt-5 panel" style={`background: var(--accent-panel-bg); border-color: var(--accent-panel-border);`}>
        <h3 class="font-semibold">{copy.systemStatus.adminTitle}</h3>
        {#if runtimeError}
          <p class="mt-2 text-sm" style={`color: var(--danger-text);`}>
            {copy.common.error}: {translateErrorMessage(runtimeError)}
          </p>
        {:else if runtimeConfig}
          {#if runtimeReport}
            <div class="mt-3 panel-strong text-sm">
              <div class="flex flex-wrap items-center justify-between gap-2">
                <p>
                  <strong>{copy.systemStatus.exportSnapshot}:</strong> {formatDateTime(runtimeReport.generatedAt)}
                </p>
                <div class="flex flex-wrap gap-2">
                  <button class="button-primary" type="button" on:click={downloadRuntimeReport}>
                    {copy.systemStatus.downloadRuntimeReport}
                  </button>
                  <button class="button-secondary" type="button" on:click={downloadIncidentReport}>
                    {copy.systemStatus.downloadIncidentReport}
                  </button>
                  <button class="button-ghost" type="button" on:click={copyIncidentSummary}>
                    {copy.systemStatus.copyIncidentSummary}
                  </button>
                </div>
              </div>
            </div>
          {/if}

          <div class="mt-3 grid gap-2 sm:grid-cols-3">
            <div class="panel-strong">
              <p class="text-sm text-muted">{copy.systemStatus.profile}</p>
              <p class="font-semibold uppercase">{runtimeConfig.profile}</p>
            </div>
            <div class="panel-strong">
              <p class="text-sm text-muted">{copy.systemStatus.rateLimitStore}</p>
              <p class="font-semibold">{runtimeConfig.operations.rateLimitStore}</p>
            </div>
            <div class="panel-strong">
              <p class="text-sm text-muted">{copy.systemStatus.workerTls}</p>
              <p class="font-semibold">{runtimeConfig.security.workerTlsMode}</p>
            </div>
          </div>

          <div class="mt-3 grid gap-3 sm:grid-cols-2">
            <div class="panel-strong text-sm">
              <p><strong>{copy.systemStatus.legacyApi}:</strong> {boolLabel(runtimeConfig.http.legacyApiEnabled, "enabledDisabled")}</p>
              <p><strong>{copy.systemStatus.strictDependencies}:</strong> {boolLabel(runtimeConfig.security.strictDependencies, "onOff")}</p>
              <p><strong>{copy.systemStatus.localAuthMode}:</strong> {runtimeConfig.security.localAuthMode}</p>
              <p><strong>{copy.systemStatus.signedLoggerIngest}:</strong> {boolLabel(runtimeConfig.security.loggerSignedIngestEnabled, "requiredOptional")}</p>
              <p>
                <strong>{copy.systemStatus.loadShedding}:</strong>
                {#if runtimeConfig.http.maxInFlightRequests > 0}
                  {" "}{runtimeConfig.http.maxInFlightRequests}
                {:else}
                  {" "}{copy.common.disabled}
                {/if}
              </p>
            </div>

            <div class="panel-strong text-sm">
              <p><strong>{copy.systemStatus.autoMigrate}:</strong> {boolLabel(runtimeConfig.operations.autoMigrate, "onOff")}</p>
              <p><strong>{copy.systemStatus.redisFailureMode}:</strong> {runtimeConfig.operations.redisFailureMode}</p>
              <p><strong>{copy.systemStatus.incidentSink}:</strong> {runtimeConfig.operations.incidentEventSink}</p>
              <p>
                <strong>{copy.systemStatus.incidentWebhook}:</strong>
                {boolLabel(runtimeConfig.operations.incidentEventWebhookConfigured, "onOff")}
              </p>
              <p>
                <strong>{copy.systemStatus.incidentRetention}:</strong>
                {` ${runtimeConfig.operations.incidentEventRetentionDays} ${copy.common.days}`}
              </p>
              <p>
                <strong>{copy.systemStatus.diagnosticsCache}:</strong>
                {` ${runtimeConfig.operations.runtimeDiagnosticsCacheTtlMs} ${copy.common.milliseconds}`}
              </p>
              <p>
                <strong>{copy.systemStatus.readyCache}:</strong>
                {` ${runtimeConfig.http.readyCacheTtlMs} ${copy.common.milliseconds}`}
              </p>
              <p>
                <strong>{copy.systemStatus.staleIfError}:</strong>
                {` ${runtimeConfig.http.readyStaleIfErrorTtlMs} ${copy.common.milliseconds}`}
              </p>
            </div>
          </div>

          {#if runtimeMetrics}
            {#if runtimeReport}
              <div class="mt-3 panel-incident text-sm">
                <div class="flex flex-wrap items-center justify-between gap-2">
                  <div>
                    <p class="font-semibold">{runtimeReport.incident.title}</p>
                    <p class="text-muted">
                      {copy.common.severity}: <strong>{runtimeReport.incident.recommendedSeverity}</strong>
                      {" | "}
                      {copy.common.category}: <strong>{runtimeReport.incident.category}</strong>
                    </p>
                  </div>
                  <p class="text-muted">
                    {copy.systemStatus.reportVersion}: <strong>{runtimeReport.reportVersion}</strong>
                  </p>
                </div>
                <p class="mt-2 text-body">{runtimeReport.incident.summary}</p>
                {#if runtimeReport.incident.nextActions.length > 0}
                  <div class="mt-3">
                    <p class="font-semibold">{copy.systemStatus.nextActions}</p>
                    <ul class="mt-1 list-disc pl-5 text-body">
                      {#each runtimeReport.incident.nextActions as action}
                        <li>{action}</li>
                      {/each}
                    </ul>
                  </div>
                {/if}
                {#if runtimeReport.runbooks.length > 0}
                  <div class="mt-3">
                    <p class="font-semibold">{copy.systemStatus.runbookMapping}</p>
                    <div class="mt-2 grid gap-2 sm:grid-cols-2">
                      {#each runtimeReport.runbooks as runbook}
                        <div class="panel-strong">
                          <p class="font-semibold">{runbook.title}</p>
                          <p><strong>{copy.common.priority}:</strong> {runbook.priority}</p>
                          <p><strong>{copy.common.path}:</strong> <code>{runbook.path}</code></p>
                          <p class="mt-1 text-body">{runbook.reason}</p>
                        </div>
                      {/each}
                    </div>
                  </div>
                {/if}
              </div>
            {/if}

            {#if runtimeMetrics.alerts.items.length > 0}
              <div class="mt-3 panel-alert text-sm">
                <div class="flex flex-wrap items-center justify-between gap-2">
                  <p class="font-semibold">{copy.systemStatus.operationalAlerts}</p>
                  <p class="text-muted">
                    {copy.systemStatus.active}: <strong>{runtimeMetrics.alerts.activeCount}</strong>
                    {" | "}
                    {copy.systemStatus.highest}: <strong>{runtimeMetrics.alerts.highestSeverity}</strong>
                  </p>
                </div>
                <div class="mt-3 grid gap-2 sm:grid-cols-2">
                  {#each runtimeMetrics.alerts.items as alert}
                    <div class="panel-strong">
                      <p class="font-semibold">{alert.code}</p>
                      <p><strong>{copy.common.status}:</strong> {alert.status}</p>
                      <p><strong>{copy.common.severity}:</strong> {alert.severity}</p>
                      <p><strong>{copy.common.source}:</strong> {alert.source}</p>
                      <p><strong>{copy.systemStatus.breaches}:</strong> {alert.breachCount}</p>
                      {#if alert.lastTriggeredAt}
                        <p><strong>{copy.systemStatus.lastTriggered}:</strong> {formatDateTime(alert.lastTriggeredAt)}</p>
                      {/if}
                      <p class="mt-2 text-body">{alert.message}</p>
                      <p class="mt-1 text-soft"><strong>{copy.common.action}:</strong> {alert.recommendedAction}</p>
                    </div>
                  {/each}
                </div>
              </div>
            {/if}

            <div class="mt-3 grid gap-2 sm:grid-cols-4">
              <div class="panel-strong">
                <p class="text-sm text-muted">{copy.systemStatus.requests}</p>
                <p class="font-semibold">{runtimeMetrics.requestsTotal}</p>
              </div>
              <div class="panel-strong">
                <p class="text-sm text-muted">{copy.systemStatus.serverErrors}</p>
                <p class="font-semibold">{runtimeMetrics.requestErrors}</p>
              </div>
              <div class="panel-strong">
                <p class="text-sm text-muted">{copy.systemStatus.averageLatency}</p>
                <p class="font-semibold">{formatDurationMs(activeLocale, runtimeMetrics.latencyAverageMs, 1)}</p>
              </div>
              <div class="panel-strong">
                <p class="text-sm text-muted">{copy.systemStatus.loadShed}</p>
                <p class="font-semibold">{runtimeMetrics.loadShedTotal}</p>
              </div>
            </div>

            <div class="mt-3 panel-strong text-sm" data-testid="runtime-metrics-summary">
              <p><strong>{copy.systemStatus.errorRate}:</strong> {formatPercent(activeLocale, runtimeMetrics.errorRate, 2)}</p>
              <p><strong>{copy.systemStatus.inflightCurrent}:</strong> {runtimeMetrics.inflightCurrent}</p>
              <p><strong>{copy.systemStatus.inflightPeak}:</strong> {runtimeMetrics.inflightPeak}</p>
              <p><strong>{copy.systemStatus.latencySamples}:</strong> {runtimeMetrics.latencyCount}</p>
            </div>

            {#if runtimeMetrics.recentHistory.length > 0}
              <div class="mt-3 panel-strong text-sm">
                <p class="font-semibold">{copy.systemStatus.recentTrend}</p>
                <div class="mt-3 grid gap-2 sm:grid-cols-4">
                  {#each runtimeMetrics.recentHistory.slice(-4) as point}
                    <div class="panel-subtle">
                      <p class="text-xs text-muted">{formatDateTime(point.recordedAt)}</p>
                      <p><strong>{copy.systemStatus.trendRequests}:</strong> {point.requestsTotal}</p>
                      <p><strong>{copy.systemStatus.trendErrors}:</strong> {formatPercent(activeLocale, point.errorRate, 1)}</p>
                      <p><strong>{copy.systemStatus.trendLatency}:</strong> {formatDurationMs(activeLocale, point.latencyAverageMs, 1)}</p>
                    </div>
                  {/each}
                </div>
              </div>
            {/if}

            <div class="mt-3 grid gap-3 sm:grid-cols-3">
              <div class="panel-strong text-sm">
                <p class="font-semibold">{copy.systemStatus.healthCorrelation}</p>
                <p><strong>{copy.common.status}:</strong> {runtimeMetrics.health.status}</p>
                <p><strong>{copy.systemStatus.checks}:</strong> {runtimeMetrics.health.postgres} / {runtimeMetrics.health.redis} / {runtimeMetrics.health.worker}</p>
                <p><strong>{copy.systemStatus.httpStatus}:</strong> {runtimeMetrics.health.httpStatus}</p>
                <p><strong>{copy.systemStatus.cache}:</strong> {runtimeMetrics.health.cacheState}</p>
                <p><strong>{copy.systemStatus.cacheAge}:</strong> {formatDurationMs(activeLocale, runtimeMetrics.health.cacheAgeMs)}</p>
              </div>

              <div class="panel-strong text-sm">
                <p class="font-semibold">{copy.systemStatus.traceFlow}</p>
                <p><strong>{copy.systemStatus.header}:</strong> {runtimeMetrics.trace.responseHeader}</p>
                <p><strong>{copy.systemStatus.enabled}:</strong> {boolLabel(runtimeMetrics.trace.enabled, "yesNo")}</p>
                <p><strong>{copy.systemStatus.forwardedToLogger}:</strong> {boolLabel(runtimeMetrics.trace.forwardedToLogger, "yesNo")}</p>
                <p><strong>{copy.systemStatus.loggerHeader}:</strong> {runtimeMetrics.trace.storedOnLoggerAs}</p>
                <p><strong>{copy.systemStatus.logField}:</strong> {runtimeMetrics.trace.storageField}</p>
              </div>

              <div class="panel-strong text-sm">
                <p class="font-semibold">{copy.systemStatus.gatewayLoggerQueue}</p>
                <p><strong>{copy.systemStatus.enabled}:</strong> {boolLabel(runtimeMetrics.gatewayLogger.enabled, "yesNo")}</p>
                <p><strong>{copy.systemStatus.queue}:</strong> {runtimeMetrics.gatewayLogger.queueDepth} / {runtimeMetrics.gatewayLogger.queueCapacity}</p>
                <p><strong>{copy.systemStatus.workers}:</strong> {runtimeMetrics.gatewayLogger.workers}</p>
                <p><strong>{copy.systemStatus.retryMax}:</strong> {runtimeMetrics.gatewayLogger.retryMax}</p>
                <p><strong>{copy.systemStatus.dropped}:</strong> {runtimeMetrics.gatewayLogger.droppedTotal}</p>
              </div>
            </div>

            <div class="mt-3 panel-strong text-sm" data-testid="trace-lookup-panel">
              <div class="flex flex-wrap items-center justify-between gap-2">
                <div>
                  <p class="font-semibold">{copy.systemStatus.traceLookup}</p>
                  <p
                    class="text-muted"
                    data-testid="trace-lookup-state"
                    data-mode={activeTraceQuery ? "filtered" : "latest"}
                    data-trace-id={activeTraceQuery || ""}
                  >
                    {traceStateText()}
                  </p>
                </div>
                <div class="flex flex-wrap gap-2">
                  <input
                    class="field"
                    bind:value={traceQuery}
                    placeholder={copy.systemStatus.tracePlaceholder}
                    data-testid="trace-query-input"
                  />
                  <button
                    class="button-primary"
                    type="button"
                    on:click={runTraceSearch}
                    data-testid="trace-search-button"
                  >
                    {copy.systemStatus.search}
                  </button>
                  <button
                    class="button-ghost"
                    type="button"
                    on:click={clearTraceSearch}
                    data-testid="trace-latest-button"
                  >
                    {copy.systemStatus.latest}
                  </button>
                </div>
              </div>

              {#if availableTraceIds().length > 0}
                <div class="mt-3 flex flex-wrap gap-2">
                  {#each availableTraceIds() as traceId}
                    <button
                      class="button-ghost"
                      type="button"
                      on:click={() => void useIncidentTrace(traceId)}
                      data-testid={`incident-trace-${traceId}`}
                    >
                      {useTraceLabel(traceId)}
                    </button>
                  {/each}
                </div>
              {/if}

              {#if requestLogsLoading}
                <p class="mt-3 text-muted">{copy.systemStatus.loadingRequestLogs}</p>
              {:else if requestLogsError}
                <p class="mt-3 text-sm" style={`color: var(--danger-text);`}>
                  {copy.common.error}: {translateErrorMessage(requestLogsError)}
                </p>
              {:else if requestLogs.length === 0}
                <p class="mt-3 text-muted">
                  {#if activeTraceQuery}
                    {copy.systemStatus.noTraceMatch}
                  {:else}
                    {copy.systemStatus.noRequestLogs}
                  {/if}
                </p>
              {:else}
                <div class="mt-3 space-y-2">
                  {#each requestLogs as entry}
                    <div class="panel-subtle" data-testid="request-log-row">
                      <div class="flex flex-wrap items-center justify-between gap-2">
                        <p class="font-semibold">
                          <span class="status-badge status-neutral mr-2">{entry.method}</span>
                          {entry.path}
                        </p>
                        <span class={requestStatusTone(entry.statusCode)}>
                          {entry.statusCode}
                        </span>
                      </div>
                      <p class="mt-2 text-soft">
                        {formatDurationMs(activeLocale, entry.durationMs)}
                        {" | "}
                        {formatDateTime(entry.occurredAt)}
                        {" | "}
                        {entry.ip}
                      </p>
                      {#if entry.traceId}
                        <p class="mt-1 text-muted">
                          {copy.common.trace}: <code>{entry.traceId}</code>
                        </p>
                      {/if}
                    </div>
                  {/each}
                </div>
              {/if}
            </div>

            <div class="mt-3 panel-strong text-sm">
              <p class="font-semibold">{copy.systemStatus.loggerService}</p>
              <div class="mt-2 grid gap-2 sm:grid-cols-3">
                <p><strong>{copy.systemStatus.configured}:</strong> {boolLabel(runtimeMetrics.loggerService.configured, "yesNo")}</p>
                <p><strong>{copy.systemStatus.reachable}:</strong> {boolLabel(runtimeMetrics.loggerService.reachable, "yesNo")}</p>
                <p><strong>{copy.systemStatus.health}:</strong> {runtimeMetrics.loggerService.healthStatus || copy.common.unknown}</p>
                <p><strong>{copy.systemStatus.queue}:</strong> {runtimeMetrics.loggerService.queueDepth} / {runtimeMetrics.loggerService.queueCapacity}</p>
                <p><strong>{copy.systemStatus.processed}:</strong> {runtimeMetrics.loggerService.processedTotal}</p>
                <p><strong>{copy.systemStatus.dropped}:</strong> {runtimeMetrics.loggerService.droppedTotal}</p>
                <p><strong>{copy.systemStatus.retried}:</strong> {runtimeMetrics.loggerService.retriedTotal}</p>
                <p><strong>{copy.systemStatus.inflightWorkers}:</strong> {runtimeMetrics.loggerService.inflightWorkers}</p>
                <p><strong>{copy.systemStatus.dropRatio}:</strong> {formatPercent(activeLocale, runtimeMetrics.loggerService.dropRatio, 2)}</p>
              </div>
              {#if runtimeMetrics.loggerService.lastError}
                <p class="mt-2 text-sm" style={`color: var(--danger-text);`}>
                  <strong>{copy.systemStatus.loggerError}:</strong> {runtimeMetrics.loggerService.lastError}
                </p>
              {/if}
            </div>

            <div class="mt-3 panel-strong text-sm">
              <p class="font-semibold">{copy.systemStatus.incidentJournal}</p>
              <div class="mt-2 grid gap-2 sm:grid-cols-3">
                <p><strong>{copy.systemStatus.enabled}:</strong> {boolLabel(runtimeMetrics.incidentJournal.enabled, "yesNo")}</p>
                <p><strong>{copy.systemStatus.sink}:</strong> {runtimeMetrics.incidentJournal.sink || copy.common.disabled}</p>
                <p><strong>{copy.systemStatus.reachable}:</strong> {boolLabel(runtimeMetrics.incidentJournal.reachable, "yesNo")}</p>
                <p><strong>{copy.systemStatus.totalEvents}:</strong> {runtimeMetrics.incidentJournal.totalEvents}</p>
                <p><strong>{copy.systemStatus.activeEvents}:</strong> {runtimeMetrics.incidentJournal.activeEvents}</p>
                <p><strong>{copy.systemStatus.lastEventStatus}:</strong> {runtimeMetrics.incidentJournal.lastEventStatus || copy.common.notAvailable}</p>
                <p><strong>{copy.systemStatus.dispatchFailures}:</strong> {runtimeMetrics.incidentJournal.dispatchFailures}</p>
                <p><strong>{copy.systemStatus.lastDispatch}:</strong> {runtimeMetrics.incidentJournal.lastDispatchAt ? formatDateTime(runtimeMetrics.incidentJournal.lastDispatchAt) : copy.common.notAvailable}</p>
                <p><strong>{copy.systemStatus.latestEvent}:</strong> {runtimeMetrics.incidentJournal.latestEventAt ? formatDateTime(runtimeMetrics.incidentJournal.latestEventAt) : copy.common.notAvailable}</p>
              </div>
              {#if runtimeMetrics.incidentJournal.lastDispatchError}
                <p class="mt-2 text-sm" style={`color: var(--danger-text);`}>
                  <strong>{copy.systemStatus.incidentJournalError}:</strong> {runtimeMetrics.incidentJournal.lastDispatchError}
                </p>
              {/if}
            </div>

            <div class="mt-3 panel-strong text-sm">
              <p class="font-semibold">{copy.systemStatus.recentIncidentEvents}</p>
              {#if incidentEvents.length === 0}
                <p class="mt-2 text-muted">{copy.systemStatus.noIncidentEvents}</p>
              {:else}
                <div class="mt-3 grid gap-2 sm:grid-cols-2">
                  {#each incidentEvents as event}
                    <div class="panel-subtle">
                      <p class="font-semibold">{event.alertCode}</p>
                      <p><strong>{copy.common.event}:</strong> {event.eventType}</p>
                      <p><strong>{copy.common.severity}:</strong> {event.recommendedSeverity}</p>
                      <p><strong>{copy.common.status}:</strong> {event.status}</p>
                      <p><strong>{copy.systemStatus.lastSeen}:</strong> {formatDateTime(event.lastSeenAt)}</p>
                      {#if event.traceId}
                        <p><strong>{copy.common.trace}:</strong> <code>{event.traceId}</code></p>
                      {/if}
                      <p class="mt-2 text-body">{event.message}</p>
                      <p class="mt-1 text-soft"><strong>{copy.common.action}:</strong> {event.recommendedAction}</p>
                      {#if event.traceId}
                        <button class="button-ghost mt-2" type="button" on:click={() => void useIncidentTrace(event.traceId)}>
                          {copy.systemStatus.useTraceButton}
                        </button>
                      {/if}
                    </div>
                  {/each}
                </div>
              {/if}
            </div>
          {/if}

          {#if runtimeConfig.warnings.length > 0 || (runtimeMetrics && runtimeMetrics.warnings.length > 0)}
            <div class="mt-3 panel-danger">
              <p class="font-semibold" style={`color: var(--danger-text);`}>{copy.systemStatus.warnings}</p>
              <ul class="mt-2 list-disc pl-5 text-sm" style={`color: var(--danger-text);`}>
                {#each runtimeConfig.warnings as warning}
                  <li>{warning}</li>
                {/each}
                {#if runtimeMetrics}
                  {#each runtimeMetrics.warnings as warning}
                    <li>{warning}</li>
                  {/each}
                {/if}
              </ul>
            </div>
          {/if}
        {:else}
          <p class="mt-2 text-sm text-muted">{copy.systemStatus.diagnosticsPending}</p>
        {/if}
      </div>
    {/if}

    <h3 class="mt-4 font-semibold">{copy.systemStatus.protectedUsers}</h3>
    <ul class="mt-2 space-y-2">
      {#if users.length === 0}
        <li class="text-sm text-muted">{copy.systemStatus.noUsers}</li>
      {:else}
        {#each users as user}
          <li class="panel text-sm">
            <strong>{user.name}</strong> ({user.email})
          </li>
        {/each}
      {/if}
    </ul>

    <div class="mt-5 panel">
      <h3 class="font-semibold">{copy.systemStatus.fibonacciTitle}</h3>
      <div class="mt-2 flex flex-wrap items-center gap-2">
        <input
          class="field"
          type="number"
          min="0"
          max="93"
          bind:value={fibInput}
        />
        <button class="button-accent" type="button" on:click={runFibonacci}>
          {copy.systemStatus.compute}
        </button>
      </div>
      {#if fibResult}
        <p class="mt-2 text-sm">{copy.systemStatus.result}: <strong>{fibResult}</strong></p>
      {/if}
      {#if fibError}
        <p class="mt-2 text-sm" style={`color: var(--danger-text);`}>
          {copy.common.error}: {translateErrorMessage(fibError)}
        </p>
      {/if}
    </div>
  {/if}
</section>

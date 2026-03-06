<script lang="ts">
import TraceLookupPanel from "@/components/Interactive/SystemStatus/TraceLookupPanel.svelte";
import type {
  RuntimeConfigResponse,
  RuntimeIncidentEventsResponse,
  RuntimeMetricsResponse,
  RuntimeReportResponse,
  RuntimeRequestLogRecord,
} from "@/lib/api/types";
import {
  booleanLabel,
  formatDurationMs,
  formatDateTime,
  formatPercent,
  getCopy,
  translateError,
  type Locale,
} from "@/lib/ui/copy";

export let activeLocale: Locale;
export let copy: ReturnType<typeof getCopy>;
export let runtimeConfig: RuntimeConfigResponse | null = null;
export let runtimeMetrics: RuntimeMetricsResponse | null = null;
export let runtimeReport: RuntimeReportResponse | null = null;
export let incidentEvents: RuntimeIncidentEventsResponse["items"] = [];
export let requestLogs: RuntimeRequestLogRecord[] = [];
export let requestLogsLoading = false;
export let requestLogsError = "";
export let traceQuery = "";
export let activeTraceQuery = "";
export let runtimeError = "";
export let onDownloadRuntimeReport: () => void;
export let onDownloadIncidentReport: () => void;
export let onCopyIncidentSummary: () => Promise<void> | void;
export let onSearchTrace: () => Promise<void> | void;
export let onLatestTrace: () => Promise<void> | void;
export let onUseTrace: (traceId: string) => Promise<void> | void;

function boolLabel(
  value: boolean,
  mode: "yesNo" | "enabledDisabled" | "onOff" | "requiredOptional",
): string {
  return booleanLabel(activeLocale, value, mode);
}

function availableTraceIds(): string[] {
  return [...new Set(incidentEvents.map((event) => event.traceId).filter(Boolean))].slice(
    0,
    6,
  ) as string[];
}

</script>

<div class="mt-5 panel" style={`background: var(--accent-panel-bg); border-color: var(--accent-panel-border);`}>
  <h3 class="font-semibold">{copy.systemStatus.adminTitle}</h3>
  {#if runtimeError}
    <p class="mt-2 text-sm" style={`color: var(--danger-text);`} data-testid="runtime-error">
      {copy.common.error}: {translateError(activeLocale, runtimeError)}
    </p>
  {:else if runtimeConfig}
    {#if runtimeReport}
      <div class="mt-3 panel-strong text-sm">
        <div class="flex flex-wrap items-center justify-between gap-2">
          <p>
            <strong>{copy.systemStatus.exportSnapshot}:</strong> {formatDateTime(activeLocale, runtimeReport.generatedAt)}
          </p>
          <div class="flex flex-wrap gap-2">
            <button class="button-primary" type="button" on:click={onDownloadRuntimeReport}>
              {copy.systemStatus.downloadRuntimeReport}
            </button>
            <button class="button-secondary" type="button" on:click={onDownloadIncidentReport}>
              {copy.systemStatus.downloadIncidentReport}
            </button>
            <button class="button-ghost" type="button" on:click={onCopyIncidentSummary}>
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
                  <p><strong>{copy.systemStatus.lastTriggered}:</strong> {formatDateTime(activeLocale, alert.lastTriggeredAt)}</p>
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
                <p class="text-xs text-muted">{formatDateTime(activeLocale, point.recordedAt)}</p>
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

      <TraceLookupPanel
        bind:traceQuery
        {activeLocale}
        {copy}
        {activeTraceQuery}
        {requestLogsLoading}
        {requestLogsError}
        {requestLogs}
        traceIds={availableTraceIds()}
        onSearch={onSearchTrace}
        onLatest={onLatestTrace}
        onUseTrace={onUseTrace}
      />

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
          <p><strong>{copy.systemStatus.lastDispatch}:</strong> {runtimeMetrics.incidentJournal.lastDispatchAt ? formatDateTime(activeLocale, runtimeMetrics.incidentJournal.lastDispatchAt) : copy.common.notAvailable}</p>
          <p><strong>{copy.systemStatus.latestEvent}:</strong> {runtimeMetrics.incidentJournal.latestEventAt ? formatDateTime(activeLocale, runtimeMetrics.incidentJournal.latestEventAt) : copy.common.notAvailable}</p>
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
                <p><strong>{copy.systemStatus.lastSeen}:</strong> {formatDateTime(activeLocale, event.lastSeenAt)}</p>
                {#if event.traceId}
                  <p><strong>{copy.common.trace}:</strong> <code>{event.traceId}</code></p>
                {/if}
                <p class="mt-2 text-body">{event.message}</p>
                <p class="mt-1 text-soft"><strong>{copy.common.action}:</strong> {event.recommendedAction}</p>
                {#if event.traceId}
                  <button class="button-ghost mt-2" type="button" on:click={() => void onUseTrace(event.traceId)}>
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
    <p class="mt-2 text-sm text-muted" data-testid="runtime-pending">{copy.systemStatus.diagnosticsPending}</p>
  {/if}
</div>

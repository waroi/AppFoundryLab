<script lang="ts">
import type { RuntimeRequestLogRecord } from "@/lib/api/types";
import {
  formatDurationMs,
  formatDateTime,
  getCopy,
  renderTemplate,
  translateError,
  type Locale,
} from "@/lib/ui/copy";

export let activeLocale: Locale;
export let copy: ReturnType<typeof getCopy>;
export let traceQuery = "";
export let activeTraceQuery = "";
export let requestLogsLoading = false;
export let requestLogsError = "";
export let requestLogs: RuntimeRequestLogRecord[] = [];
export let traceIds: string[] = [];
export let onSearch: () => Promise<void> | void;
export let onLatest: () => Promise<void> | void;
export let onUseTrace: (traceId: string) => Promise<void> | void;

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

function useTraceLabel(traceId: string): string {
  return renderTemplate(copy.systemStatus.useTrace, { traceId });
}

function traceStateText(): string {
  if (activeTraceQuery) {
    return `${copy.systemStatus.traceFilter}: ${activeTraceQuery}`;
  }
  return copy.systemStatus.latestTraceSummary;
}
</script>

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
      <label class="field-stack">
        <span class="field-label">{copy.common.trace}</span>
        <input
          class="field"
          bind:value={traceQuery}
          placeholder={copy.systemStatus.tracePlaceholder}
          aria-label={copy.common.trace}
          data-testid="trace-query-input"
        />
      </label>
      <button class="button-primary" type="button" on:click={onSearch} data-testid="trace-search-button">
        {copy.systemStatus.search}
      </button>
      <button class="button-ghost" type="button" on:click={onLatest} data-testid="trace-latest-button">
        {copy.systemStatus.latest}
      </button>
    </div>
  </div>

  {#if traceIds.length > 0}
    <div class="mt-3 flex flex-wrap gap-2">
      {#each traceIds as traceId}
        <button
          class="button-ghost"
          type="button"
          on:click={() => void onUseTrace(traceId)}
          data-testid={`incident-trace-${traceId}`}
        >
          {useTraceLabel(traceId)}
        </button>
      {/each}
    </div>
  {/if}

  {#if requestLogsLoading}
    <p class="mt-3 text-muted" data-testid="trace-loading">{copy.systemStatus.loadingRequestLogs}</p>
  {:else if requestLogsError}
    <p class="mt-3 text-sm" style={`color: var(--danger-text);`} data-testid="trace-error">
      {copy.common.error}: {translateError(activeLocale, requestLogsError)}
    </p>
  {:else if requestLogs.length === 0}
    <p class="mt-3 text-muted" data-testid="trace-empty-state">
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
            {formatDateTime(activeLocale, entry.occurredAt)}
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

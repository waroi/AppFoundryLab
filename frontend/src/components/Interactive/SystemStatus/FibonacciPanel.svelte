<script lang="ts">
import { getCopy } from "@/lib/ui/copy";

export let copy: ReturnType<typeof getCopy>;
export let fibInput = 40;
export let fibResult = "";
export let fibError = "";
export let onRunFibonacci: () => Promise<void> | void;
export let translateErrorMessage: (value: string) => string;
</script>

<div class="mt-5 panel">
  <h3 class="font-semibold">{copy.systemStatus.fibonacciTitle}</h3>
  <div class="mt-2 flex flex-wrap items-center gap-2">
    <label class="field-stack">
      <span class="field-label">{copy.systemStatus.fibonacciTitle}</span>
      <input
        class="field"
        type="number"
        min="0"
        max="93"
        bind:value={fibInput}
        aria-label={copy.systemStatus.fibonacciTitle}
        data-testid="fibonacci-input"
      />
    </label>
    <button class="button-accent" type="button" on:click={onRunFibonacci}>
      {copy.systemStatus.compute}
    </button>
  </div>
  {#if fibResult}
    <p class="mt-2 text-sm" data-testid="fibonacci-result">
      {copy.systemStatus.result}: <strong>{fibResult}</strong>
    </p>
  {/if}
  {#if fibError}
    <p class="mt-2 text-sm" style={`color: var(--danger-text);`} data-testid="fibonacci-error">
      {copy.common.error}: {translateErrorMessage(fibError)}
    </p>
  {/if}
</div>

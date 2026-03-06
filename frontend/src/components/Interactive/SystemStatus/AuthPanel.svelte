<script lang="ts">
import { formatRole, getCopy, type Locale } from "@/lib/ui/copy";

export let activeLocale: Locale;
export let copy: ReturnType<typeof getCopy>;
export let username = "";
export let password = "";
export let token = "";
export let role = "";
export let authError = "";
export let onLogin: () => Promise<void> | void;
export let translateErrorMessage: (value: string) => string;
</script>

<div class="mt-5 panel">
  <h3 class="font-semibold">{copy.systemStatus.authTitle}</h3>
  <p class="mt-2 text-sm text-soft">{copy.systemStatus.authHelper}</p>
  <div class="mt-2 flex flex-wrap gap-2">
    <label class="field-stack">
      <span class="field-label">{copy.systemStatus.username}</span>
      <input
        class="field"
        bind:value={username}
        placeholder={copy.systemStatus.username}
        aria-label={copy.systemStatus.username}
        data-testid="login-username"
      />
    </label>
    <label class="field-stack">
      <span class="field-label">{copy.systemStatus.password}</span>
      <input
        class="field"
        type="password"
        bind:value={password}
        placeholder={copy.systemStatus.password}
        aria-label={copy.systemStatus.password}
        data-testid="login-password"
      />
    </label>
    <button class="button-accent" type="button" on:click={onLogin} data-testid="login-submit">
      {copy.systemStatus.loginButton}
    </button>
  </div>
  {#if authError}
    <p class="mt-2 text-sm" style={`color: var(--danger-text);`} data-testid="auth-error">
      {copy.common.error}: {translateErrorMessage(authError)}
    </p>
  {/if}
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

<script lang="ts">
import { getCopy, translateError } from "@/lib/ui/copy";
import { DEFAULT_LOCALE, locale, type Locale } from "@/lib/ui/preferences";
import { onMount } from "svelte";

type VerificationPayload = {
  marker: string;
  status: string;
  usersExpected: number;
  usersFound: number;
  requestLogsExpected: number;
  requestLogsFound: number;
};

let verification: VerificationPayload | null = null;
let manifestLines: string[] = [];
let error = "";
export let initialLocale: Locale = DEFAULT_LOCALE;

$: activeLocale = typeof window === "undefined" ? initialLocale : $locale;
$: copy = getCopy(activeLocale);

function isVerificationPayload(value: unknown): value is VerificationPayload {
  if (typeof value !== "object" || value === null) {
    return false;
  }

  const candidate = value as Record<string, unknown>;
  return (
    typeof candidate.marker === "string" &&
    typeof candidate.status === "string" &&
    typeof candidate.usersExpected === "number" &&
    typeof candidate.usersFound === "number" &&
    typeof candidate.requestLogsExpected === "number" &&
    typeof candidate.requestLogsFound === "number"
  );
}

onMount(async () => {
  try {
    const [verificationResponse, manifestResponse] = await Promise.all([
      fetch("/fixtures/restore-drill/fixture-verification-sample.json"),
      fetch("/fixtures/restore-drill/fixture-manifest-sample.txt"),
    ]);

    const verificationPayload: unknown = await verificationResponse.json();
    if (!verificationResponse.ok || !isVerificationPayload(verificationPayload)) {
      throw new Error("invalid_restore_drill_verification");
    }

    const manifestText = await manifestResponse.text();
    if (!manifestResponse.ok) {
      throw new Error("invalid_restore_drill_manifest");
    }

    verification = verificationPayload;
    manifestLines = manifestText
      .split("\n")
      .map((line) => line.trim())
      .filter(Boolean);
  } catch (currentError) {
    error = currentError instanceof Error ? currentError.message : "unknown_error";
  }
});
</script>

<section class="card mt-6" data-testid="restore-drill-preview">
  <h2 class="text-xl font-semibold">{copy.restoreDrill.title}</h2>
  <p class="mt-2 text-body">
    {copy.restoreDrill.description}
  </p>

  {#if error}
    <p class="mt-3 text-sm" style={`color: var(--danger-text);`}>
      {copy.common.error}: {translateError(activeLocale, error)}
    </p>
  {:else if verification}
    <div class="mt-3 grid gap-2 text-sm sm:grid-cols-4">
      <div class="panel-subtle">
        <p class="text-muted">{copy.restoreDrill.marker}</p>
        <p class="font-semibold" data-testid="restore-drill-marker">{verification.marker}</p>
      </div>
      <div class="panel-subtle">
        <p class="text-muted">{copy.restoreDrill.status}</p>
        <p
          class="font-semibold uppercase"
          data-testid="restore-drill-status"
          data-status={verification.status}
        >
          {verification.status}
        </p>
      </div>
      <div class="panel-subtle">
        <p class="text-muted">{copy.restoreDrill.users}</p>
        <p class="font-semibold">{verification.usersFound} / {verification.usersExpected}</p>
      </div>
      <div class="panel-subtle">
        <p class="text-muted">{copy.restoreDrill.requestLogs}</p>
        <p class="font-semibold">{verification.requestLogsFound} / {verification.requestLogsExpected}</p>
      </div>
    </div>

    <div class="mt-3 panel-strong text-sm">
      <p class="font-semibold">{copy.restoreDrill.fixtureManifest}</p>
      <ul class="mt-2 space-y-1 text-soft">
        {#each manifestLines as line}
          <li data-testid="restore-drill-manifest-line"><code>{line}</code></li>
        {/each}
      </ul>
    </div>
  {:else}
    <p class="mt-3 text-muted">{copy.restoreDrill.loading}</p>
  {/if}
</section>

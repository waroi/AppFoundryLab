<script lang="ts">
import type { HealthResponse } from "@/lib/api/types";
import { getCopy } from "@/lib/ui/copy";

export let copy: ReturnType<typeof getCopy>;
export let health: HealthResponse;
</script>

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

<script lang="ts">
import type { UsersResponse } from "@/lib/api/types";
import { getCopy } from "@/lib/ui/copy";

export let copy: ReturnType<typeof getCopy>;
export let users: UsersResponse["data"] = [];
export let usersError = "";
export let translateErrorMessage: (value: string) => string;
</script>

<h3 class="mt-4 font-semibold">{copy.systemStatus.protectedUsers}</h3>
{#if usersError}
  <p class="mt-2 text-sm" style={`color: var(--danger-text);`} data-testid="users-error">
    {copy.common.error}: {translateErrorMessage(usersError)}
  </p>
{/if}
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

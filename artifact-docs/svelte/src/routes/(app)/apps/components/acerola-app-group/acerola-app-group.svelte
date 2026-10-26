<script lang="ts" module>
  import type { AppGroup } from '$lib/hooks/use-app-hub/use-app-hub.svelte';

  /**
   * Um tipo de app e os apps dele, em grade. O título é um `h2` com a contagem: na lista de
   * tipos a pessoa acha o grupo pelo nome antes de ler os cartões.
   */
  export type AcerolaAppGroupProps = {
    data: { group: AppGroup };
    ui?: { className?: string };
  };
</script>

<script lang="ts">
  import AppLinkCard from '../acerola-app-link-card/acerola-app-link-card.svelte';
  import { cn } from '$lib/utils/cn';

  let { data, ui }: AcerolaAppGroupProps = $props();
</script>

<section class={cn('flex flex-col gap-3', ui?.className)} aria-labelledby="app-group-{data.group.id}">
  <h2 id="app-group-{data.group.id}" class="flex items-baseline gap-2 text-lg">
    {data.group.name}
    <span class="text-muted-foreground text-sm">{data.group.apps.length}</span>
  </h2>
  <div class="card-grid">
    {#each data.group.apps as app (app.id)}
      <AppLinkCard data={{ app }} />
    {/each}
  </div>
</section>

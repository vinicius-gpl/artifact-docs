<script lang="ts" module>
  import type { AppGroup } from '$lib/hooks/use-app-hub/use-app-hub.svelte';

  /**
   * O corpo do hub de apps, na ordem dos estados (skill `ui-standards` §5): carregando → erro
   * → nenhum app cadastrado → busca escondeu tudo → os grupos.
   */
  export type AcerolaAppHubViewProps = {
    data: { groups: AppGroup[]; total: number };
    state: { isLoading: boolean; error: string | null; search: string };
    actions: { onRetry: () => void; onClearSearch: () => void };
  };
</script>

<script lang="ts">
  import AppWindow from '@lucide/svelte/icons/app-window';
  import SearchX from '@lucide/svelte/icons/search-x';

  import AppGroupSection from '../acerola-app-group/acerola-app-group.svelte';
  import ActionButton from '$lib/components/acerola-action-button/acerola-action-button.svelte';
  import EmptyState from '$lib/components/acerola-empty-state/acerola-empty-state.svelte';
  import ErrorState from '$lib/components/acerola-error-state/acerola-error-state.svelte';
  import { Skeleton } from '$lib/components/acerola-skeleton/acerola-skeleton';

  let { data, state, actions }: AcerolaAppHubViewProps = $props();
</script>

{#if state.isLoading}
  <div class="card-grid" aria-busy="true" aria-label="Carregando">
    {#each { length: 4 } as _, index (index)}
      <Skeleton class="rounded-surface h-28" />
    {/each}
  </div>
{:else if state.error}
  <ErrorState data={{ title: 'Os apps não carregaram', message: state.error }} actions={{ onRetry: actions.onRetry }} />
{:else if data.total === 0}
  <EmptyState
    data={{ title: 'Nenhum app cadastrado', description: 'Os apps cadastrados aparecem aqui, agrupados por tipo.' }}
    ui={{ icon: AppWindow }}
  />
{:else if data.groups.length === 0}
  <EmptyState
    data={{ title: 'Nenhum app encontrado', description: `Nada combina com “${state.search}”.` }}
    ui={{ icon: SearchX }}
  >
    <ActionButton
      data={{ label: 'Limpar busca' }}
      ui={{ variant: 'secondary', size: 'lg' }}
      actions={{ onClick: actions.onClearSearch }}
    />
  </EmptyState>
{:else}
  <div class="flex flex-col gap-8">
    {#each data.groups as group (group.id)}
      <AppGroupSection data={{ group }} />
    {/each}
  </div>
{/if}

<script lang="ts" module>
  import type { PaletteView } from '$lib/hooks/use-brand-palettes/use-brand-palettes.svelte';

  /** O corpo da casa da marca, na ordem dos estados: carregando → erro → vazio → paletas. */
  export type AcerolaPaletteListProps = {
    data: { palettes: PaletteView[] };
    state: { isLoading: boolean; error: string | null };
    actions: { onRetry: () => void };
  };
</script>

<script lang="ts">
  import Palette from '@lucide/svelte/icons/palette';

  import PaletteCard from '../acerola-palette-card/acerola-palette-card.svelte';
  import EmptyState from '$lib/components/acerola-empty-state/acerola-empty-state.svelte';
  import ErrorState from '$lib/components/acerola-error-state/acerola-error-state.svelte';
  import { Skeleton } from '$lib/components/acerola-skeleton/acerola-skeleton';

  let { data, state, actions }: AcerolaPaletteListProps = $props();
</script>

{#if state.isLoading}
  <div class="flex flex-col gap-5" aria-busy="true" aria-label="Carregando">
    <Skeleton class="rounded-surface h-56" />
    <Skeleton class="rounded-surface h-56" />
  </div>
{:else if state.error}
  <ErrorState data={{ title: 'As paletas não carregaram', message: state.error }} actions={{ onRetry: actions.onRetry }} />
{:else if data.palettes.length === 0}
  <EmptyState
    data={{
      title: 'Nenhuma paleta ainda',
      description: 'Suba um .css com as cores base para gerar a primeira paleta da marca.',
    }}
    ui={{ icon: Palette }}
  />
{:else}
  <div class="flex flex-col gap-5">
    {#each data.palettes as palette (palette.id)}
      <PaletteCard data={{ palette }} />
    {/each}
  </div>
{/if}

<script lang="ts" module>
  import type { LucideIcon } from '@lucide/svelte';

  import type { PreviewItem } from '$lib/components/acerola-preview-card/acerola-preview-card.svelte';
  import type { ShortcutListItem } from '$lib/components/acerola-shortcut-list/acerola-shortcut-list.svelte';

  /**
   * O CORPO de uma tela de galeria: a galeria à esquerda e, ao lado, "Fixados" e "Recentes" —
   * o arranjo da tela de Artefatos do Claude (briefing, módulo 2).
   *
   * Existe porque Artefatos, Planejamentos e Ideias têm o mesmo corpo, e uma rota não pode
   * importar componente de outra: a ordem dos estados (skill `ui-standards` §5) mora aqui, uma
   * vez, e não copiada em três telas que cedo ou tarde divergiriam.
   *
   * A ordem, com early return: carregando → erro → vazio de verdade → filtro escondeu tudo →
   * conteúdo. A lateral só aparece com conteúdo: num erro ou no esqueleto ela seria ruído.
   */
  export type AcerolaGalleryViewProps = {
    data: {
      items: PreviewItem[];
      total: number;
      pinned: ShortcutListItem[];
      recent: ShortcutListItem[];
      caption: string;
      emptyTitle: string;
      emptyDescription: string;
    };
    ui?: { emptyIcon?: LucideIcon; className?: string };
    state: {
      isLoading: boolean;
      error: string | null;
      hasFilters: boolean;
      forceCards?: boolean;
      pendingPinId?: string | null;
    };
    actions: { onRetry: () => void; onClearFilters: () => void; onTogglePin: (id: string) => void };
  };

  /** Quantos cartões de esqueleto: uma fileira cheia no monitor, sem rolar no celular. */
  export const SKELETON_CARDS = 4;
</script>

<script lang="ts">
  import Clock from '@lucide/svelte/icons/clock';
  import Pin from '@lucide/svelte/icons/pin';
  import SearchX from '@lucide/svelte/icons/search-x';

  import ActionButton from '$lib/components/acerola-action-button/acerola-action-button.svelte';
  import EmptyState from '$lib/components/acerola-empty-state/acerola-empty-state.svelte';
  import ErrorState from '$lib/components/acerola-error-state/acerola-error-state.svelte';
  import PreviewGallery from '$lib/components/acerola-preview-gallery/acerola-preview-gallery.svelte';
  import ShortcutList from '$lib/components/acerola-shortcut-list/acerola-shortcut-list.svelte';
  import { Skeleton } from '$lib/components/acerola-skeleton/acerola-skeleton';
  import { cn } from '$lib/utils/cn';

  let { data, ui, state, actions }: AcerolaGalleryViewProps = $props();
</script>

{#if state.isLoading}
  <div class={cn('card-grid', ui?.className)} aria-busy="true" aria-label="Carregando">
    {#each { length: SKELETON_CARDS } as _, index (index)}
      <Skeleton class="rounded-surface aspect-[4/5]" />
    {/each}
  </div>
{:else if state.error}
  <ErrorState
    data={{ title: 'A galeria não carregou', message: state.error }}
    ui={{ className: ui?.className }}
    actions={{ onRetry: actions.onRetry }}
  />
{:else if data.total === 0}
  <EmptyState
    data={{ title: data.emptyTitle, description: data.emptyDescription }}
    ui={{ icon: ui?.emptyIcon, className: ui?.className }}
  />
{:else}
  <div class={cn('grid gap-5 lg:grid-cols-[minmax(0,1fr)_16rem]', ui?.className)}>
    {#if data.items.length === 0}
      <EmptyState
        data={{
          title: 'Nada encontrado',
          description: 'Nenhum item combina com a busca e os filtros. Tente outro termo.',
        }}
        ui={{ icon: SearchX }}
      >
        <ActionButton
          data={{ label: 'Limpar filtros' }}
          ui={{ variant: 'secondary', size: 'lg' }}
          actions={{ onClick: actions.onClearFilters }}
        />
      </EmptyState>
    {:else}
      <PreviewGallery
        data={{ items: data.items, caption: data.caption }}
        state={{ forceCards: state.forceCards, pendingPinId: state.pendingPinId }}
        actions={{ onTogglePin: actions.onTogglePin }}
      />
    {/if}
    <aside class="flex flex-col gap-3" aria-label="Atalhos">
      <ShortcutList
        data={{ title: 'Fixados', items: data.pinned, emptyLabel: 'Fixe um item para ele aparecer aqui.' }}
        ui={{ icon: Pin }}
      />
      <ShortcutList
        data={{ title: 'Recentes', items: data.recent, emptyLabel: 'Abra um item para ele aparecer aqui.' }}
        ui={{ icon: Clock }}
      />
    </aside>
  </div>
{/if}

<script lang="ts" module>
  import type { PreviewItem } from '$lib/components/acerola-preview-card/acerola-preview-card.svelte';

  /**
   * A galeria de itens com preview, nos DOIS formatos da casa (skill `ui-standards` §6): grade
   * de cartões com a miniatura, e lista em tabela para comparar título, selos e data.
   *
   * Quem escolhe o formato é a preferência do sistema (`use-table-view`), que a tela passa em
   * `state.forceCards`: abaixo de `xl` é sempre cartão; acima, tabela — ou cartão, se a pessoa
   * pediu "ver sempre em cards". A troca é por CSS, os dois blocos existem e um fica escondido.
   *
   * Só desenha o conteúdo: carregando, erro e vazio são da tela, que decide antes de chegar aqui.
   */
  export type AcerolaPreviewGalleryProps = {
    data: { items: PreviewItem[]; caption: string };
    ui?: { className?: string };
    state?: { forceCards?: boolean; pendingPinId?: string | null };
    actions?: { onTogglePin?: (id: string) => void };
  };

  export function countLabel(count: number): string {
    return count === 1 ? '1 item' : `${count} itens`;
  }
</script>

<script lang="ts">
  import Lock from '@lucide/svelte/icons/lock';
  import Pin from '@lucide/svelte/icons/pin';
  import PinOff from '@lucide/svelte/icons/pin-off';

  import ActionButton from '$lib/components/acerola-action-button/acerola-action-button.svelte';
  import PreviewCard from '$lib/components/acerola-preview-card/acerola-preview-card.svelte';
  import StatusBadge from '$lib/components/acerola-status-badge/acerola-status-badge.svelte';
  import {
    Table,
    TableActions,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
  } from '$lib/components/acerola-table/acerola-table';
  import { cn } from '$lib/utils/cn';

  let { data, ui, state, actions }: AcerolaPreviewGalleryProps = $props();

  const forceCards = $derived(Boolean(state?.forceCards));
</script>

<div class={cn('flex flex-col gap-3', ui?.className)}>
  <div class={cn('card-grid', !forceCards && 'xl:hidden')} data-slot="preview-cards-mobile">
    {#each data.items as item (item.id)}
      <PreviewCard
        data={{ item }}
        state={{ isPinPending: state?.pendingPinId === item.id }}
        actions={{ onTogglePin: actions?.onTogglePin }}
      />
    {/each}
  </div>

  <div
    class={cn('overflow-x-auto', forceCards ? 'hidden' : 'hidden xl:block')}
    data-slot="preview-table-desktop"
  >
    <Table class="min-w-[48rem]">
      <TableHeader>
        <TableRow>
          <TableHead class="min-w-80">Título</TableHead>
          <TableHead class="min-w-48">Detalhes</TableHead>
          <TableHead class="min-w-36">Editado</TableHead>
          <TableHead class="w-16"><span class="sr-only">Ações</span></TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {#each data.items as item (item.id)}
          <TableRow>
            <TableCell>
              <div class="flex items-center gap-2">
                <a href={item.href} class="font-medium hover:underline">{item.title}</a>
                {#if item.isLocked}
                  <Lock class="text-muted-foreground size-3.5" aria-label="Privado" role="img" />
                {/if}
              </div>
              {#if item.description}
                <p class="text-muted-foreground line-clamp-1 text-xs">{item.description}</p>
              {/if}
            </TableCell>
            <TableCell>
              <div class="flex flex-wrap gap-1.5">
                {#each item.badges as badge (badge.label)}
                  <StatusBadge data={{ label: badge.label }} ui={{ tone: badge.tone, size: 'sm' }} />
                {/each}
              </div>
            </TableCell>
            <TableCell class="text-muted-foreground text-sm">{item.meta}</TableCell>
            <TableCell>
              <TableActions>
                <ActionButton
                  data={{ label: item.isPinned ? `Desafixar ${item.title}` : `Fixar ${item.title}` }}
                  ui={{
                    variant: 'ghost',
                    size: 'lg',
                    icon: item.isPinned ? PinOff : Pin,
                    isIconOnly: true,
                  }}
                  state={{
                    isPressed: item.isPinned,
                    isLoading: state?.pendingPinId === item.id,
                  }}
                  actions={{ onClick: () => actions?.onTogglePin?.(item.id) }}
                />
              </TableActions>
            </TableCell>
          </TableRow>
        {/each}
      </TableBody>
      {#snippet footer()}
        <span>{data.caption}</span>
        <span>{countLabel(data.items.length)}</span>
      {/snippet}
    </Table>
  </div>
</div>

<script lang="ts" module>
  import type { AppTile } from '$lib/hooks/use-app-hub/use-app-hub.svelte';

  /**
   * Um app do hub: logo, nome, endereço e se está no ar.
   *
   * O cartão inteiro é o link, e abre em outra aba (`target="_blank"`): o hub é ponto de
   * partida, e quem clica em "Jellyfin" não quer perder a lista que estava vendo.
   * `rel="noopener"` impede o app aberto de mexer nesta aba por `window.opener`.
   *
   * A logo pode não existir (favicon que o site não tem, endereço fora do ar): aí entram as
   * iniciais, no mesmo quadrado — nunca o ícone de imagem quebrada do navegador.
   */
  export type AcerolaAppLinkCardProps = {
    data: { app: AppTile };
    ui?: { className?: string };
    /** `onEdit` liga o botão de editar, no canto de baixo — FORA do link: botão dentro de
     * `<a>` é HTML inválido, e o clique abriria o app em vez de editar. */
    actions?: { onEdit?: () => void };
  };

  export function initialsOf(name: string): string {
    return name
      .split(/\s+/)
      .filter(Boolean)
      .slice(0, 2)
      .map((word) => word[0]?.toLocaleUpperCase('pt-BR') ?? '')
      .join('');
  }
</script>

<script lang="ts">
  import ExternalLink from '@lucide/svelte/icons/external-link';
  import Pencil from '@lucide/svelte/icons/pencil';

  import ActionButton from '$lib/components/acerola-action-button/acerola-action-button.svelte';

  import StatusBadge from '$lib/components/acerola-status-badge/acerola-status-badge.svelte';
  import { fillColorOf, fillFromPointer } from '$lib/motion/hover-fill';
  import { hoverLift } from '$lib/motion/motion';
  import { cn } from '$lib/utils/cn';

  let { data, ui, actions }: AcerolaAppLinkCardProps = $props();

  /* Estado puramente visual: a imagem falhou, mostra as iniciais. */
  let hasIconFailed = $state(false);

  const app = $derived(data.app);
</script>

<div class={cn('relative', ui?.className)}>
<a
  href={app.url}
  target="_blank"
  rel="noopener noreferrer"
  class="glass hover-fill rounded-surface flex min-h-10 flex-col gap-3 p-4"
  style:--fill-color={fillColorOf(app.status.tone === 'danger' ? 'danger' : 'neutral')}
  use:fillFromPointer
  use:hoverLift
>
  <div class="flex items-center gap-3">
    <span class="bg-muted rounded-box flex size-10 shrink-0 items-center justify-center overflow-hidden">
      {#if hasIconFailed}
        <span class="text-primary-strong text-sm font-semibold" aria-hidden="true">
          {initialsOf(app.name)}
        </span>
      {:else}
        <img src={app.iconUrl} alt="" class="size-6" onerror={() => (hasIconFailed = true)} />
      {/if}
    </span>
    <div class="min-w-0 flex-1">
      <p class="truncate font-medium">{app.name}</p>
      <p class="text-muted-foreground truncate text-xs">{app.host}</p>
    </div>
    <ExternalLink class="text-muted-foreground size-4 shrink-0" aria-hidden="true" />
  </div>
  <div class={cn('flex items-center justify-between gap-2', actions?.onEdit && 'pr-10')}>
    <StatusBadge data={{ label: app.status.label }} ui={{ tone: app.status.tone, size: 'sm' }} />
    <span class="text-muted-foreground text-xs">{app.checkedLabel}</span>
  </div>
</a>
{#if actions?.onEdit}
  <ActionButton
    data={{ label: `Editar ${app.name}` }}
    ui={{ variant: 'ghost', size: 'sm', icon: Pencil, isIconOnly: true, className: 'absolute right-2 bottom-2' }}
    actions={{ onClick: actions.onEdit }}
  />
{/if}
</div>

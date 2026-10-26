<script lang="ts" module>
  import type { StatusBadgeTone } from '$lib/components/acerola-status-badge/acerola-status-badge.svelte';

  /**
   * Um item de galeria COM PREVIEW: a página aparece em miniatura no alto, como na tela de
   * Artefatos do Claude, e embaixo o título, os selos e "Editado há X".
   *
   * Genérico de propósito: não sabe o que é artefato, planejamento ou ideia. Quem traduz a
   * entidade para `PreviewItem` é o hook da tela — por isso as três galerias do hub usam o
   * mesmo cartão e não três parecidos.
   *
   * O PREVIEW: com miniatura (`thumbUrl`), uma imagem; sem ela, a própria página num
   * `<iframe>` reduzido à metade (briefing, "Preview dos cards"). O iframe é decorativo — sem
   * foco, sem clique, `sandbox` sem `allow-same-origin` — e só carrega quando chega perto da
   * tela (`loading="lazy"`).
   *
   * O cartão inteiro abre o item: o link do título se estica sobre ele (`after:inset-0`), e o
   * botão de fixar fica por cima (`z-10`). Assim há UM link por cartão para o leitor de tela,
   * e não um por pedaço.
   */
  export type PreviewBadge = { label: string; tone: StatusBadgeTone };

  export type PreviewItem = {
    id: string;
    title: string;
    description: string | null;
    href: string;
    /** O endereço da página, para o preview em iframe quando não há miniatura. */
    previewUrl: string;
    thumbUrl: string | null;
    /** A linha do rodapé: "Editado há 2 dias". */
    meta: string;
    badges: PreviewBadge[];
    isLocked: boolean;
    isPinned: boolean;
  };

  export type AcerolaPreviewCardProps = {
    data: { item: PreviewItem };
    ui?: { className?: string };
    state?: { isPinPending?: boolean };
    actions?: { onTogglePin?: (id: string) => void };
  };
</script>

<script lang="ts">
  import Lock from '@lucide/svelte/icons/lock';
  import Pin from '@lucide/svelte/icons/pin';
  import PinOff from '@lucide/svelte/icons/pin-off';

  import ActionButton from '$lib/components/acerola-action-button/acerola-action-button.svelte';
  import StatusBadge from '$lib/components/acerola-status-badge/acerola-status-badge.svelte';
  import { fillColorOf, fillFromPointer } from '$lib/motion/hover-fill';
  import { hoverLift } from '$lib/motion/motion';
  import { cn } from '$lib/utils/cn';

  let { data, ui, state, actions }: AcerolaPreviewCardProps = $props();

  const item = $derived(data.item);
</script>

<article
  class={cn('glass hover-fill rounded-surface flex flex-col', ui?.className)}
  style:--fill-color={fillColorOf('brand')}
  use:fillFromPointer
  use:hoverLift
  data-slot="preview-card"
>
  <div class="bg-background relative aspect-[16/10] overflow-hidden border-b" aria-hidden="true">
    {#if item.thumbUrl}
      <img src={item.thumbUrl} alt="" class="size-full object-cover object-top" loading="lazy" />
    {:else}
      <iframe
        src={item.previewUrl}
        title=""
        tabindex="-1"
        loading="lazy"
        sandbox="allow-scripts"
        class="pointer-events-none absolute top-0 left-0 h-[200%] w-[200%] origin-top-left scale-50 border-0"
      ></iframe>
    {/if}
  </div>

  <div class="flex flex-col gap-2 p-4">
    <div class="flex items-start justify-between gap-2">
      <h3 class="min-w-0 font-medium break-words">
        <a
          href={item.href}
          class="after:absolute after:inset-0 after:content-[''] focus-visible:outline-none focus-visible:after:outline-2 focus-visible:after:outline-ring"
        >
          {item.title}
        </a>
      </h3>
      {#if item.isLocked}
        <Lock class="text-muted-foreground mt-1 size-4 shrink-0" aria-label="Privado" role="img" />
      {/if}
    </div>
    {#if item.description}
      <p class="text-muted-foreground line-clamp-2 text-sm">{item.description}</p>
    {/if}
    {#if item.badges.length > 0}
      <div class="flex flex-wrap gap-1.5">
        {#each item.badges as badge (badge.label)}
          <StatusBadge data={{ label: badge.label }} ui={{ tone: badge.tone, size: 'sm' }} />
        {/each}
      </div>
    {/if}
  </div>

  <footer class="border-border/60 flex items-center justify-between gap-2 border-t py-1 pr-1 pl-4">
    <span class="text-muted-foreground text-xs">{item.meta}</span>
    <ActionButton
      data={{ label: item.isPinned ? `Desafixar ${item.title}` : `Fixar ${item.title}` }}
      ui={{
        variant: 'ghost',
        size: 'lg',
        icon: item.isPinned ? PinOff : Pin,
        isIconOnly: true,
        className: 'relative z-10',
      }}
      state={{ isPressed: item.isPinned, isLoading: state?.isPinPending }}
      actions={{ onClick: () => actions?.onTogglePin?.(item.id) }}
    />
  </footer>
</article>

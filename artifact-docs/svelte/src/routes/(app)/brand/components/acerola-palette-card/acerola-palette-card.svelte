<script lang="ts" module>
  import type { PaletteView } from '$lib/hooks/use-brand-palettes/use-brand-palettes.svelte';

  /**
   * Uma paleta gerada: as amostras de cada token, nos dois temas, e o `.css` para baixar.
   *
   * Cada amostra é dividida ao meio — claro à esquerda, escuro à direita — porque a pergunta
   * de quem olha é "essa cor funciona nos dois?", e duas grades separadas obrigariam a comparar
   * de memória. A cor entra por `style`, não por classe: ela é DADO que veio da API, não token
   * desta interface.
   *
   * Baixar é um link (`<a download>`), não um botão com script: o navegador já sabe salvar um
   * arquivo, e o link continua funcionando com o botão do meio ou "Salvar link como".
   */
  export type AcerolaPaletteCardProps = {
    data: { palette: PaletteView };
    ui?: { className?: string };
  };
</script>

<script lang="ts">
  import Download from '@lucide/svelte/icons/download';

  import { actionButton } from '$lib/components/acerola-action-button/acerola-action-button.svelte';
  import StatusBadge from '$lib/components/acerola-status-badge/acerola-status-badge.svelte';
  import { cn } from '$lib/utils/cn';

  let { data, ui }: AcerolaPaletteCardProps = $props();

  const palette = $derived(data.palette);
</script>

<article class={cn('glass rounded-surface flex flex-col gap-4 p-5', ui?.className)}>
  <header class="flex flex-wrap items-start justify-between gap-3">
    <div class="min-w-0">
      <h2 class="flex flex-wrap items-center gap-2 text-lg">
        {palette.name}
        {#if palette.isBrand}
          <StatusBadge data={{ label: 'Marca oficial' }} ui={{ tone: 'brand', size: 'sm' }} />
        {/if}
      </h2>
      <p class="text-muted-foreground text-xs">{palette.createdLabel}</p>
    </div>
    <a
      href={palette.downloadUrl}
      download={palette.downloadName}
      class={cn(actionButton({ variant: 'secondary', size: 'lg' }), 'inline-flex items-center gap-2 text-sm')}
    >
      <Download class="size-4" aria-hidden="true" />
      Baixar CSS
    </a>
  </header>

  <ul class="grid grid-cols-2 gap-3 sm:grid-cols-4">
    {#each palette.colors as color (color.token)}
      <li class="flex flex-col gap-1.5">
        <div class="rounded-box flex h-16 overflow-hidden border" aria-hidden="true">
          <span class="flex-1" style:background-color={color.light}></span>
          <span class="flex-1" style:background-color={color.dark}></span>
        </div>
        <span class="font-mono text-xs">{color.token}</span>
        <span class="text-muted-foreground font-mono text-xs">
          {color.light} · {color.dark}
        </span>
      </li>
    {/each}
  </ul>
</article>

<script lang="ts" module>
  import type { LucideIcon } from '@lucide/svelte';

  /**
   * Uma lista curta de atalhos com título — "Fixados", "Recentes" — na lateral da galeria.
   *
   * Cada linha é um link de verdade (`<a href>`), com 40px de altura: abre em nova aba, o leitor
   * de tela anuncia como link, e o alvo é do tamanho mínimo da marca (Lei de Fitts).
   *
   * Lista vazia não some: mostra a frase de `emptyLabel`, para a pessoa saber que o bloco existe
   * e como ele se enche ("Fixe um item para ele aparecer aqui").
   */
  export type ShortcutListItem = { id: string; label: string; href: string; meta: string };

  export type AcerolaShortcutListProps = {
    data: { title: string; items: ShortcutListItem[]; emptyLabel: string };
    ui?: { icon?: LucideIcon; className?: string };
  };
</script>

<script lang="ts">
  import { cn } from '$lib/utils/cn';

  let { data, ui }: AcerolaShortcutListProps = $props();

  const Icon = $derived(ui?.icon);
</script>

<section class={cn('glass rounded-surface flex flex-col gap-1 p-3', ui?.className)}>
  <h2 class="text-muted-foreground flex items-center gap-2 px-2 py-1 text-xs font-medium tracking-wide uppercase">
    {#if Icon}
      <Icon class="size-3.5" aria-hidden="true" />
    {/if}
    {data.title}
  </h2>
  {#if data.items.length === 0}
    <p class="text-muted-foreground px-2 py-2 text-sm">{data.emptyLabel}</p>
  {:else}
    <ul class="flex flex-col">
      {#each data.items as item (item.id)}
        <li>
          <a
            href={item.href}
            class="hover:bg-muted rounded-box flex min-h-10 flex-col justify-center px-2 py-1.5 transition-colors"
          >
            <span class="truncate text-sm">{item.label}</span>
            <span class="text-muted-foreground text-xs">{item.meta}</span>
          </a>
        </li>
      {/each}
    </ul>
  {/if}
</section>

<script lang="ts" module>
  import type { NavItem } from '$lib/navigation/navigation';

  /**
   * A BARRA DO TOPO — a navegação do hub inteiro: a marca à esquerda, as telas no meio e o
   * tema à direita, como um site e não como um painel de empresa.
   *
   * Substitui o menu lateral que veio do acerola-ticket: aquele existia para dezenas de telas,
   * perfil, papéis e "Sair". O hub tem cinco telas e ninguém para identificar (não tem login),
   * então uma linha só basta e devolve a largura inteira para a galeria.
   *
   * Fica presa no topo (`sticky`) com o acabamento Vidro: o conteúdo passa por baixo desfocado.
   * A tela aberta é marcada com `aria-current="page"` e um traço na cor da marca embaixo — o
   * indicador do Carbon.
   *
   * No celular os itens não cabem em linha: viram um botão "Menu" que abre a lista embaixo da
   * barra. Abrir e fechar é estado visual, mora aqui; clicar num item fecha a lista.
   */
  export type AcerolaTopNavProps = {
    data: { items: readonly NavItem[]; homeHref: string };
    state: { activeKey: string | undefined };
    ui?: { className?: string };
  };
</script>

<script lang="ts">
  import Menu from '@lucide/svelte/icons/menu';
  import X from '@lucide/svelte/icons/x';

  import ActionButton from '$lib/components/acerola-action-button/acerola-action-button.svelte';
  import { BrandMark } from '$lib/components/acerola-brand-mark/acerola-brand-mark';
  import ThemeToggle from '$lib/components/acerola-theme-toggle/acerola-theme-toggle.svelte';
  import { cn } from '$lib/utils/cn';

  let { data, state: navState, ui }: AcerolaTopNavProps = $props();

  let isMenuOpen = $state(false);

  function closeMenu() {
    isMenuOpen = false;
  }
</script>

<header
  class={cn(
    'border-border/70 bg-background/70 sticky top-0 z-40 border-b backdrop-blur-xl',
    ui?.className,
  )}
>
  <div class="mx-auto flex h-16 w-full max-w-6xl items-center gap-6 px-4 sm:px-6">
    <a href={data.homeHref} class="flex min-h-10 shrink-0 items-center gap-2" onclick={closeMenu}>
      <BrandMark ui={{ size: 'sm' }} />
      <span class="text-lg font-medium">artifact-docs</span>
    </a>

    <nav aria-label="Principal" class="hidden md:block">
      <ul class="flex items-center gap-1">
        {#each data.items as item (item.key)}
          {@const isActive = item.key === navState.activeKey}
          <li>
            <a
              href={item.to}
              aria-current={isActive ? 'page' : undefined}
              class={cn(
                'flex h-16 items-center border-b-2 px-3 text-sm transition-colors',
                isActive
                  ? 'border-primary text-foreground'
                  : 'text-muted-foreground hover:text-foreground border-transparent',
              )}
            >
              {item.label}
            </a>
          </li>
        {/each}
      </ul>
    </nav>

    <div class="ml-auto flex items-center gap-1">
      <ThemeToggle />
      <div class="md:hidden">
        <ActionButton
          data={{ label: isMenuOpen ? 'Fechar menu' : 'Abrir menu' }}
          ui={{ variant: 'ghost', size: 'lg', icon: isMenuOpen ? X : Menu, isIconOnly: true }}
          state={{ isPressed: isMenuOpen }}
          actions={{ onClick: () => (isMenuOpen = !isMenuOpen) }}
        />
      </div>
    </div>
  </div>

  {#if isMenuOpen}
    <nav aria-label="Principal" class="border-border/70 border-t md:hidden" data-slot="top-nav-menu">
      <ul class="mx-auto flex max-w-6xl flex-col px-4 py-2">
        {#each data.items as item (item.key)}
          {@const isActive = item.key === navState.activeKey}
          <li>
            <a
              href={item.to}
              aria-current={isActive ? 'page' : undefined}
              onclick={closeMenu}
              class={cn(
                'flex min-h-12 items-center gap-3 border-l-2 px-3 text-base',
                isActive
                  ? 'border-primary bg-primary-soft text-foreground'
                  : 'text-muted-foreground border-transparent',
              )}
            >
              <item.icon class="size-5" aria-hidden="true" />
              {item.label}
            </a>
          </li>
        {/each}
      </ul>
    </nav>
  {/if}
</header>

<script lang="ts" module>
  import type { NavItem } from '$lib/navigation/navigation';

  /**
   * A BARRA DO TOPO — a navegação do hub inteiro: a marca à esquerda, as telas no meio e o
   * tema à direita, como um site e não como um painel de empresa.
   *
   * Substitui o menu lateral que veio do acerola-ticket: aquele existia para dezenas de telas,
   * perfil e papéis. O hub tem quatro telas e uma pessoa só, então uma linha basta e devolve a
   * largura inteira para a galeria. Sem avatar nem perfil: o hub é de uma pessoa só, e de
   * identidade basta o "Sair".
   *
   * Fica presa no topo (`sticky`) com o acabamento Vidro: o conteúdo passa por baixo desfocado.
   * A tela aberta é marcada com `aria-current="page"` e um traço na cor da marca embaixo — o
   * indicador do Carbon. O traço é UM só e desliza de item em item (GSAP, `moveIndicator`):
   * o olho acompanha para onde a navegação foi, em vez de ver um traço sumir e outro surgir.
   *
   * No celular os itens não cabem em linha: viram um botão "Menu" que abre a lista embaixo da
   * barra. Abrir e fechar é estado visual, mora aqui; clicar num item fecha a lista.
   */
  export type AcerolaTopNavProps = {
    data: {
      items: readonly NavItem[];
      homeHref: string;
    };
    state: { activeKey: string | undefined; isSigningOut?: boolean };
    actions: { onSignOut: () => void };
    ui?: { className?: string };
  };
</script>

<script lang="ts">
  import LogOut from '@lucide/svelte/icons/log-out';
  import Menu from '@lucide/svelte/icons/menu';
  import X from '@lucide/svelte/icons/x';

  import ActionButton from '$lib/components/acerola-action-button/acerola-action-button.svelte';
  import { tick } from 'svelte';

  import { BrandMark } from '$lib/components/acerola-brand-mark/acerola-brand-mark';
  import ThemeToggle from '$lib/components/acerola-theme-toggle/acerola-theme-toggle.svelte';
  import { moveIndicator, slideDown } from '$lib/motion/motion';
  import { startsGroup } from '$lib/navigation/navigation';
  import { cn } from '$lib/utils/cn';

  let { data, state: navState, actions, ui }: AcerolaTopNavProps = $props();

  let isMenuOpen = $state(false);

  /* Estado visual do marcador: a lista onde ele corre e se já foi posicionado uma vez. */
  let list = $state<HTMLUListElement>();
  let indicator = $state<HTMLSpanElement>();
  let isPlaced = false;

  function placeIndicator(immediate: boolean) {
    const active = list?.querySelector<HTMLElement>('[aria-current="page"]') ?? null;
    moveIndicator(indicator ?? null, active, { immediate });
  }

  $effect(() => {
    void navState.activeKey;
    void tick().then(() => {
      placeIndicator(!isPlaced);
      isPlaced = true;
    });
  });

  function closeMenu() {
    isMenuOpen = false;
  }
</script>

<!-- Janela mudou de largura: os itens mudam de lugar, o marcador vai junto, sem viajar. -->
<svelte:window onresize={() => placeIndicator(true)} />

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
      <ul bind:this={list} class="relative flex items-center gap-1">
        {#each data.items as item, index (item.key)}
          {@const isActive = item.key === navState.activeKey}
          <!-- A linha entre grupos: Apps de um lado, as abas de publicação do outro. -->
          {#if startsGroup(data.items, index)}
            <li class="bg-border mx-2 h-5 w-px" aria-hidden="true" data-slot="top-nav-divider"></li>
          {/if}
          <li>
            <a
              href={item.to}
              aria-current={isActive ? 'page' : undefined}
              title={item.group}
              class={cn(
                'flex h-16 items-center px-3 text-sm transition-colors',
                isActive ? 'text-foreground' : 'text-muted-foreground hover:text-foreground',
              )}
            >
              {item.label}
            </a>
          </li>
        {/each}
        <span
          bind:this={indicator}
          class="bg-primary pointer-events-none absolute bottom-0 left-0 h-0.5 w-0 opacity-0"
          aria-hidden="true"
          data-slot="top-nav-indicator"
        ></span>
      </ul>
    </nav>

    <div class="ml-auto flex items-center gap-1">
      <ThemeToggle />
      <ActionButton
        data={{ label: 'Sair', loadingLabel: 'Saindo…' }}
        ui={{ variant: 'ghost', size: 'lg', icon: LogOut, isIconOnly: true }}
        state={{ isLoading: navState.isSigningOut }}
        actions={{ onClick: actions.onSignOut }}
      />
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
    <nav
      aria-label="Principal"
      class="border-border/70 overflow-hidden border-t md:hidden"
      data-slot="top-nav-menu"
      use:slideDown
    >
      <ul class="mx-auto flex max-w-6xl flex-col px-4 py-2">
        {#each data.items as item, index (item.key)}
          {@const isActive = item.key === navState.activeKey}
          {#if startsGroup(data.items, index) && item.group}
            <li
              class="text-muted-foreground border-border/70 mt-2 border-t px-3 pt-3 pb-1 text-xs font-medium tracking-wide uppercase"
              aria-hidden="true"
            >
              {item.group}
            </li>
          {/if}
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

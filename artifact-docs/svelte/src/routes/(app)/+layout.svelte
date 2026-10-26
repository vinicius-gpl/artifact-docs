<script lang="ts">
  import type { Snippet } from 'svelte';

  import type { LayoutData } from './$types';

  import TopNav from '$lib/components/acerola-top-nav/acerola-top-nav.svelte';
  import { usePageTransition } from '$lib/hooks/use-page-transition/use-page-transition.svelte';
  import { useTopNavModel } from '$lib/hooks/use-top-nav/use-top-nav.svelte';

  /**
   * A CASCA DO HUB — a barra no topo e a tela embaixo.
   *
   * `data.user` vem da guarda de sessão (`+layout.ts`): quando qualquer coisa daqui para baixo
   * roda, a sessão já foi conferida. A margem do conteúdo é daqui; cada tela define só a
   * largura máxima (`mx-auto w-full max-w-6xl`).
   */
  let { data, children }: { data: LayoutData; children: Snippet } = $props();

  const nav = useTopNavModel({ user: data.user });

  /* A troca de página anima este contêiner: a tela sai, a nova entra em cascata. */
  let main = $state<HTMLElement>();
  usePageTransition(() => main);
</script>

<TopNav data={nav.data} state={nav.state} actions={nav.actions} />

<main bind:this={main} class="min-w-0 px-4 pt-8 pb-16 sm:px-6">
  {@render children()}
</main>

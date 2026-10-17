<script lang="ts">
  import type { Snippet } from 'svelte';

  import AppShell from '$lib/components/acerola-app-shell/acerola-app-shell.svelte';
  import { useAppShellModel } from '$lib/hooks/use-app-shell/use-app-shell.svelte';
  import { NAV_ITEMS } from '$lib/navigation/navigation';
  import type { LayoutData } from './$types';

  /**
   * A CASCA DO HUB — menu à esquerda, tela à direita.
   *
   * Uma casca só, porque o hub é um módulo só. `data.user` e `data.sidebarOpen` vêm da guarda
   * de sessão em `+layout.ts`: quando qualquer coisa daqui para baixo roda, a sessão já foi
   * conferida.
   */
  let { data, children }: { data: LayoutData; children: Snippet } = $props();

  const shell = useAppShellModel({ user: data.user, items: NAV_ITEMS });
</script>

<AppShell
  initialSidebarOpen={data.sidebarOpen}
  data={shell.data}
  ui={shell.ui}
  state={shell.state}
  actions={shell.actions}
>
  {@render children()}
</AppShell>

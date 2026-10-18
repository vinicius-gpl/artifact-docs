<script lang="ts">
  import { QueryClient, setQueryClientContext } from '@tanstack/svelte-query';

  import { type SessionUser } from '$lib/contracts/user.schema';
  import { type NavItem } from '$lib/navigation/navigation';
  import { useAppShellModel, type AppShellModel } from './use-app-shell.svelte';

  /**
   * O hook usa `useQueryClient()` (para `onLogout` limpar o cache) — só existe dentro de um
   * componente, com um `QueryClient` no contexto. Este apoio monta o model e o entrega ao
   * teste, do mesmo jeito que os outros `*-harness.test.svelte` dos hooks.
   */
  let {
    user,
    items,
    onReady,
  }: {
    user: SessionUser;
    items: readonly NavItem[];
    onReady: (model: AppShellModel) => void;
  } = $props();

  setQueryClientContext(new QueryClient({ defaultOptions: { queries: { retry: false } } }));

  onReady(useAppShellModel({ user, items }));
</script>

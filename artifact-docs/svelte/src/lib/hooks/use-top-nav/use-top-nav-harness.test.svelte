<script lang="ts">
  import { type QueryClient, setQueryClientContext } from '@tanstack/svelte-query';

  import { type SessionUser } from '$lib/contracts/user.schema';
  import { useTopNavModel, type TopNavModel } from './use-top-nav.svelte';

  /** O hook usa `useQueryClient()` (sair limpa o cache): monta dentro de um componente. */
  let {
    user,
    queryClient,
    onReady,
  }: { user: SessionUser; queryClient: QueryClient; onReady: (model: TopNavModel) => void } =
    $props();

  setQueryClientContext(queryClient);

  onReady(useTopNavModel({ user }));
</script>

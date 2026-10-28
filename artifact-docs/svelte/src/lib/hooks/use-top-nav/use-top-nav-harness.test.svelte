<script lang="ts">
  import { type QueryClient, setQueryClientContext } from '@tanstack/svelte-query';

  import { useTopNavModel, type TopNavModel } from './use-top-nav.svelte';

  /** O hook usa `useQueryClient()` (sair limpa o cache): monta dentro de um componente. */
  let { queryClient, onReady }: { queryClient: QueryClient; onReady: (model: TopNavModel) => void } =
    $props();

  setQueryClientContext(queryClient);

  onReady(useTopNavModel());
</script>

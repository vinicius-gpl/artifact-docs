<script lang="ts">
  import { QueryClient, setQueryClientContext } from '@tanstack/svelte-query';

  import { useAppHubModel, type AppHubModel } from './use-app-hub.svelte';

  /** O hook usa o svelte-query: monta dentro de um componente, com um `QueryClient`. */
  let { onReady }: { onReady: (model: AppHubModel) => void } = $props();

  setQueryClientContext(new QueryClient({ defaultOptions: { queries: { retry: false } } }));

  onReady(useAppHubModel());
</script>

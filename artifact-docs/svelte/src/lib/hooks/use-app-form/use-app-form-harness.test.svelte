<script lang="ts">
  import { QueryClient, setQueryClientContext } from '@tanstack/svelte-query';

  import { useAppFormModel, type AppFormModel } from './use-app-form.svelte';

  /** O hook usa o svelte-query: só existe dentro de um componente, com um `QueryClient`. */
  let { onReady }: { onReady: (model: AppFormModel) => void } = $props();

  setQueryClientContext(
    new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } }),
  );

  onReady(useAppFormModel());
</script>

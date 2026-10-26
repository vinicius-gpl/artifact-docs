<script lang="ts">
  import { QueryClient, setQueryClientContext } from '@tanstack/svelte-query';

  import { useArtifactDetailModel, type ArtifactDetailModel } from './use-artifact-detail.svelte';

  /** O hook usa o svelte-query: monta dentro de um componente, com um `QueryClient`. */
  let { label, onReady }: { label: string; onReady: (model: ArtifactDetailModel) => void } = $props();

  setQueryClientContext(new QueryClient({ defaultOptions: { queries: { retry: false } } }));

  onReady(useArtifactDetailModel({ label }));
</script>

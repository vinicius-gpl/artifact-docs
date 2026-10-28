<script lang="ts">
  import { type QueryClient, setQueryClientContext } from '@tanstack/svelte-query';

  import { useArtifactEditModel, type ArtifactEditModel } from './use-artifact-edit.svelte';

  /** O hook usa o svelte-query: só existe dentro de um componente, com um `QueryClient`. */
  let {
    label,
    queryClient,
    onReady,
  }: { label: string; queryClient: QueryClient; onReady: (model: ArtifactEditModel) => void } =
    $props();

  setQueryClientContext(queryClient);

  onReady(useArtifactEditModel({ label }));
</script>

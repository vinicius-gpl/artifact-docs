<script lang="ts">
  import { QueryClient, setQueryClientContext } from '@tanstack/svelte-query';

  import { type ArtifactKind } from '$lib/contracts/artifact.schema';
  import { useArtifactPublishModel, type ArtifactPublishModel } from './use-artifact-publish.svelte';

  /** O hook usa o svelte-query: só existe dentro de um componente, com um `QueryClient`. */
  let {
    kind,
    label,
    onReady,
  }: { kind?: ArtifactKind; label?: string; onReady: (model: ArtifactPublishModel) => void } = $props();

  setQueryClientContext(new QueryClient({ defaultOptions: { mutations: { retry: false } } }));

  onReady(useArtifactPublishModel({ kind, label }));
</script>

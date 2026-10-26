<script lang="ts">
  import { QueryClient, setQueryClientContext } from '@tanstack/svelte-query';

  import { type ArtifactKind } from '$lib/contracts/artifact.schema';
  import { useArtifactGalleryModel, type ArtifactGalleryModel } from './use-artifact-gallery.svelte';

  /**
   * O hook usa o svelte-query — só existe dentro de um componente, com um `QueryClient` no
   * contexto. Este apoio monta o model e o entrega ao teste.
   */
  let {
    kind,
    onReady,
  }: { kind?: ArtifactKind; onReady: (model: ArtifactGalleryModel) => void } = $props();

  setQueryClientContext(new QueryClient({ defaultOptions: { queries: { retry: false } } }));

  onReady(useArtifactGalleryModel({ kind }));
</script>

<script lang="ts">
  import ArtifactViewer from '../components/acerola-artifact-viewer/acerola-artifact-viewer.svelte';
  import { useArtifactDetailModel } from '$lib/hooks/use-artifact-detail/use-artifact-detail.svelte';

  /* Liga o hook ao visualizador. Mora num slot para a página poder remontá-lo quando a label
     muda: o hook busca o artefato da label que recebeu ao nascer. */
  let { label }: { label: string } = $props();

  const detail = useArtifactDetailModel({ label });
</script>

<svelte:head>
  <title>{detail.data.artifact?.title ?? 'Artefato'}</title>
</svelte:head>

<ArtifactViewer data={detail.data} state={detail.state} actions={detail.actions} />

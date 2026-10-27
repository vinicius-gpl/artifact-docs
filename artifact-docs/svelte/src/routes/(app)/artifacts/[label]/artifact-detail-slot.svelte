<script lang="ts">
  import ArtifactPublishDialog from '../components/acerola-artifact-publish-dialog/acerola-artifact-publish-dialog.svelte';
  import ArtifactViewer from '../components/acerola-artifact-viewer/acerola-artifact-viewer.svelte';
  import ConfirmDialog from '$lib/components/acerola-confirm-dialog/acerola-confirm-dialog.svelte';
  import { useArtifactDetailModel } from '$lib/hooks/use-artifact-detail/use-artifact-detail.svelte';
  import { useArtifactPublishModel } from '$lib/hooks/use-artifact-publish/use-artifact-publish.svelte';

  /* Liga os hooks ao visualizador, ao diálogo de versão nova e à confirmação de apagar. Mora
     num slot para a página poder remontá-lo quando a label muda: os hooks buscam o artefato da
     label que receberam ao nascer. */
  let { label }: { label: string } = $props();

  const detail = useArtifactDetailModel({ label });
  const publish = useArtifactPublishModel({ label });
</script>

<svelte:head>
  <title>{detail.data.artifact?.title ?? 'Artefato'}</title>
</svelte:head>

<ArtifactViewer
  data={detail.data}
  state={detail.state}
  actions={{
    onRetry: detail.actions.onRetry,
    onPublishVersion: publish.actions.onOpen,
    onTogglePrivate: detail.actions.onTogglePrivate,
    onRestore: detail.actions.onRestore,
    onDelete: detail.actions.onDeleteRequest,
  }}
/>

<ArtifactPublishDialog data={publish.data} ui={publish.ui} state={publish.state} actions={publish.actions} />

<ConfirmDialog
  data={{
    title: 'Apagar este artefato?',
    description: `"${detail.data.artifact?.title ?? label}" e todas as versões dele saem do hub, e o link para de abrir. Não dá para desfazer.`,
    confirmLabel: 'Apagar artefato',
    confirmingLabel: 'Apagando…',
  }}
  state={{ isOpen: detail.state.isDeleteOpen, isConfirming: detail.state.isDeleting, error: detail.state.deleteError }}
  actions={{ onConfirm: detail.actions.onDeleteConfirm, onCancel: detail.actions.onDeleteCancel }}
/>

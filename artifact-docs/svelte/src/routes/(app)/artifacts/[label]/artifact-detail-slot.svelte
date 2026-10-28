<script lang="ts">
  import ArtifactEditDialog from '../components/acerola-artifact-edit-dialog/acerola-artifact-edit-dialog.svelte';
  import ArtifactViewer from '../components/acerola-artifact-viewer/acerola-artifact-viewer.svelte';
  import ConfirmDialog from '$lib/components/acerola-confirm-dialog/acerola-confirm-dialog.svelte';
  import PublishDialog from '$lib/components/acerola-publish-dialog/acerola-publish-dialog.svelte';
  import { useArtifactDetailModel } from '$lib/hooks/use-artifact-detail/use-artifact-detail.svelte';
  import { useArtifactEditModel } from '$lib/hooks/use-artifact-edit/use-artifact-edit.svelte';
  import { useArtifactPublishModel } from '$lib/hooks/use-artifact-publish/use-artifact-publish.svelte';
  import { markTab } from '$lib/navigation/active-tab.svelte';

  /* Liga os hooks ao visualizador, aos diálogos de editar e de versão nova e à confirmação de
     apagar. Mora num slot para a página poder remontá-lo quando a label muda: os hooks buscam
     o artefato da label que receberam ao nascer. */
  let { label }: { label: string } = $props();

  const detail = useArtifactDetailModel({ label });
  const publish = useArtifactPublishModel({ label });
  const edit = useArtifactEditModel({ label });

  /* A barra acende a aba do tipo (uma ideia acende Ideias), e não a do endereço
     `/artifacts`. Solta a marca ao sair da tela. */
  $effect(() => {
    const navKey = detail.data.back.navKey;
    if (navKey) return markTab(navKey);
  });
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
    onEdit: edit.actions.onOpen,
    onTogglePrivate: detail.actions.onTogglePrivate,
    onRestore: detail.actions.onRestore,
    onDelete: detail.actions.onDeleteRequest,
  }}
/>

<PublishDialog data={publish.data} ui={publish.ui} state={publish.state} actions={publish.actions} />

<ArtifactEditDialog data={edit.data} ui={edit.ui} state={edit.state} actions={edit.actions} />

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

<script lang="ts">
  import FileText from '@lucide/svelte/icons/file-text';
  import Upload from '@lucide/svelte/icons/upload';

  import ActionButton from '$lib/components/acerola-action-button/acerola-action-button.svelte';
  import GalleryView from '$lib/components/acerola-gallery-view/acerola-gallery-view.svelte';
  import PageHeader from '$lib/components/acerola-page-header/acerola-page-header.svelte';
  import PublishDialog from '$lib/components/acerola-publish-dialog/acerola-publish-dialog.svelte';
  import TableViewToggle from '$lib/components/acerola-table-view-toggle/acerola-table-view-toggle.svelte';
  import TextField from '$lib/components/acerola-text-field/acerola-text-field.svelte';
  import { useArtifactGalleryModel } from '$lib/hooks/use-artifact-gallery/use-artifact-gallery.svelte';
  import { useArtifactPublishModel } from '$lib/hooks/use-artifact-publish/use-artifact-publish.svelte';
  import { useTableViewModel } from '$lib/hooks/use-table-view/use-table-view.svelte';

  /* Documentos: os artefatos de `kind` document. Planejamentos e Ideias têm aba própria — cada
     aba mostra e publica só o seu tipo. Documento não tem situação, então o filtro é só a busca. */
  const gallery = useArtifactGalleryModel({ kind: 'document' });
  const tableView = useTableViewModel();
  const publish = useArtifactPublishModel({ kind: 'document' });
</script>

<svelte:head>
  <title>Documentos</title>
</svelte:head>

<div class="mx-auto flex w-full max-w-6xl flex-col gap-5">
  <PageHeader
    data={{ title: 'Documentos', description: 'Documentação, páginas e PDFs publicados, com preview.' }}
  >
    <ActionButton
      data={{ label: 'Publicar documento' }}
      ui={{ icon: Upload }}
      actions={{ onClick: publish.actions.onOpen }}
    />
  </PageHeader>

  <div class="flex flex-wrap items-end gap-x-4 gap-y-3">
    <TextField
      data={{ label: 'Buscar', name: 'search', value: gallery.state.filters.search, autoComplete: 'off' }}
      ui={{ type: 'search', className: 'w-full sm:w-72' }}
      actions={{ onChange: gallery.actions.onSearch }}
    />
    <div class="ml-auto">
      <TableViewToggle />
    </div>
  </div>

  <GalleryView
    data={{
      ...gallery.data,
      caption: 'Documentos',
      emptyTitle: 'Nenhum documento ainda',
      emptyDescription: 'Publique o primeiro: HTML, um site em .zip, Markdown ou PDF.',
    }}
    ui={{ emptyIcon: FileText }}
    state={{ ...gallery.state, forceCards: tableView.forceCards }}
    actions={gallery.actions}
  />
</div>

<PublishDialog data={publish.data} ui={publish.ui} state={publish.state} actions={publish.actions} />

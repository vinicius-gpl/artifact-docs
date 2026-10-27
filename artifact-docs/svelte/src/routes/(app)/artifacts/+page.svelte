<script lang="ts">
  import LayoutGrid from '@lucide/svelte/icons/layout-grid';
  import Upload from '@lucide/svelte/icons/upload';

  import ArtifactPublishDialog from './components/acerola-artifact-publish-dialog/acerola-artifact-publish-dialog.svelte';
  import ActionButton from '$lib/components/acerola-action-button/acerola-action-button.svelte';

  import FilterField from '$lib/components/acerola-filter-field/acerola-filter-field.svelte';
  import GalleryView from '$lib/components/acerola-gallery-view/acerola-gallery-view.svelte';
  import OptionPicker from '$lib/components/acerola-option-picker/acerola-option-picker.svelte';
  import PageHeader from '$lib/components/acerola-page-header/acerola-page-header.svelte';
  import TableViewToggle from '$lib/components/acerola-table-view-toggle/acerola-table-view-toggle.svelte';
  import TextField from '$lib/components/acerola-text-field/acerola-text-field.svelte';
  import { useArtifactGalleryModel } from '$lib/hooks/use-artifact-gallery/use-artifact-gallery.svelte';
  import { useArtifactPublishModel } from '$lib/hooks/use-artifact-publish/use-artifact-publish.svelte';
  import { useTableViewModel } from '$lib/hooks/use-table-view/use-table-view.svelte';

  /* A galeria de tudo: documentos, planejamentos e ideias, com filtro por tipo. */
  const gallery = useArtifactGalleryModel();
  const tableView = useTableViewModel();
  const publish = useArtifactPublishModel();
</script>

<svelte:head>
  <title>Artefatos</title>
</svelte:head>

<div class="mx-auto flex w-full max-w-6xl flex-col gap-5">
  <PageHeader
    data={{ title: 'Artefatos', description: 'Documentos, páginas e PDFs publicados, com preview.' }}
  >
    <ActionButton
      data={{ label: 'Publicar artefato' }}
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
    <FilterField data={{ label: 'Tipo' }}>
      <OptionPicker
        data={{ value: gallery.state.filters.kind, options: gallery.ui.kindOptions }}
        ui={{ ariaLabel: 'Tipo', allLabel: 'Todos' }}
        actions={{ onChange: gallery.actions.onKindChange }}
      />
    </FilterField>
    <div class="ml-auto">
      <TableViewToggle />
    </div>
  </div>

  <GalleryView
    data={{
      ...gallery.data,
      caption: 'Artefatos',
      emptyTitle: 'Nenhum artefato ainda',
      emptyDescription: 'Os artefatos publicados pela interface ou pela API aparecem aqui, com preview.',
    }}
    ui={{ emptyIcon: LayoutGrid }}
    state={{ ...gallery.state, forceCards: tableView.forceCards }}
    actions={gallery.actions}
  />
</div>

<ArtifactPublishDialog data={publish.data} ui={publish.ui} state={publish.state} actions={publish.actions} />

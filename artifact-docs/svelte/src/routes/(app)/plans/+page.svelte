<script lang="ts">
  import ListTodo from '@lucide/svelte/icons/list-todo';

  import Upload from '@lucide/svelte/icons/upload';

  import ActionButton from '$lib/components/acerola-action-button/acerola-action-button.svelte';
  import FilterField from '$lib/components/acerola-filter-field/acerola-filter-field.svelte';
  import GalleryView from '$lib/components/acerola-gallery-view/acerola-gallery-view.svelte';
  import OptionPicker from '$lib/components/acerola-option-picker/acerola-option-picker.svelte';
  import PageHeader from '$lib/components/acerola-page-header/acerola-page-header.svelte';
  import PublishDialog from '$lib/components/acerola-publish-dialog/acerola-publish-dialog.svelte';
  import TableViewToggle from '$lib/components/acerola-table-view-toggle/acerola-table-view-toggle.svelte';
  import TextField from '$lib/components/acerola-text-field/acerola-text-field.svelte';
  import { useArtifactGalleryModel } from '$lib/hooks/use-artifact-gallery/use-artifact-gallery.svelte';
  import { useArtifactPublishModel } from '$lib/hooks/use-artifact-publish/use-artifact-publish.svelte';
  import { useTableViewModel } from '$lib/hooks/use-table-view/use-table-view.svelte';

  /* Planejamentos são artefatos de `kind` plan: a mesma galeria, com filtro por situação. */
  const gallery = useArtifactGalleryModel({ kind: 'plan' });
  const tableView = useTableViewModel();
  const publish = useArtifactPublishModel({ kind: 'plan' });
</script>

<svelte:head>
  <title>Planejamentos</title>
</svelte:head>

<div class="mx-auto flex w-full max-w-6xl flex-col gap-5">
  <PageHeader
    data={{ title: 'Planejamentos', description: 'Planos em andamento, com versões e situação.' }}
  >
    <ActionButton
      data={{ label: 'Publicar planejamento' }}
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
    <FilterField data={{ label: 'Situação' }}>
      <OptionPicker
        data={{ value: gallery.state.filters.status, options: gallery.ui.statusOptions }}
        ui={{ ariaLabel: 'Situação', allLabel: 'Todas' }}
        actions={{ onChange: gallery.actions.onStatusChange }}
      />
    </FilterField>
    <div class="ml-auto">
      <TableViewToggle />
    </div>
  </div>

  <GalleryView
    data={{
      ...gallery.data,
      caption: 'Planejamentos',
      emptyTitle: 'Nenhum planejamento ainda',
      emptyDescription: 'Publique o primeiro planejamento: HTML, um site em .zip, Markdown ou PDF.',
    }}
    ui={{ emptyIcon: ListTodo }}
    state={{ ...gallery.state, forceCards: tableView.forceCards }}
    actions={gallery.actions}
  />
</div>

<PublishDialog data={publish.data} ui={publish.ui} state={publish.state} actions={publish.actions} />

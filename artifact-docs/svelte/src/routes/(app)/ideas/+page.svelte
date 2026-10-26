<script lang="ts">
  import Lightbulb from '@lucide/svelte/icons/lightbulb';

  import FilterField from '$lib/components/acerola-filter-field/acerola-filter-field.svelte';
  import GalleryView from '$lib/components/acerola-gallery-view/acerola-gallery-view.svelte';
  import OptionPicker from '$lib/components/acerola-option-picker/acerola-option-picker.svelte';
  import PageHeader from '$lib/components/acerola-page-header/acerola-page-header.svelte';
  import TableViewToggle from '$lib/components/acerola-table-view-toggle/acerola-table-view-toggle.svelte';
  import TextField from '$lib/components/acerola-text-field/acerola-text-field.svelte';
  import { useArtifactGalleryModel } from '$lib/hooks/use-artifact-gallery/use-artifact-gallery.svelte';
  import { useTableViewModel } from '$lib/hooks/use-table-view/use-table-view.svelte';

  /* Ideias são artefatos de `kind` idea: a mesma galeria, com filtro por situação. */
  const gallery = useArtifactGalleryModel({ kind: 'idea' });
  const tableView = useTableViewModel();
</script>

<svelte:head>
  <title>Ideias</title>
</svelte:head>

<div class="mx-auto flex w-full max-w-6xl flex-col gap-5">
  <PageHeader
    data={{ title: 'Ideias', description: 'O banco de ideias: do rascunho ao que já virou projeto.' }}
  />

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
      caption: 'Ideias',
      emptyTitle: 'Nenhuma ideia ainda',
      emptyDescription: 'Publique um artefato com o tipo Ideia para ela aparecer aqui.',
    }}
    ui={{ emptyIcon: Lightbulb }}
    state={{ ...gallery.state, forceCards: tableView.forceCards }}
    actions={gallery.actions}
  />
</div>

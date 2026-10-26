<script lang="ts">
  import AppHubView from './components/acerola-app-hub-view/acerola-app-hub-view.svelte';
  import PageHeader from '$lib/components/acerola-page-header/acerola-page-header.svelte';
  import TextField from '$lib/components/acerola-text-field/acerola-text-field.svelte';
  import { useAppHubModel } from '$lib/hooks/use-app-hub/use-app-hub.svelte';

  const hub = useAppHubModel();
</script>

<svelte:head>
  <title>Apps</title>
</svelte:head>

<div class="mx-auto flex w-full max-w-6xl flex-col gap-5">
  <PageHeader
    data={{
      title: 'Apps',
      description: 'Atalhos para os apps self-hosted e da VPS, agrupados por tipo.',
    }}
  />

  <div class="flex flex-wrap items-end gap-x-4 gap-y-3">
    <TextField
      data={{ label: 'Buscar', name: 'search', value: hub.state.search, autoComplete: 'off' }}
      ui={{ type: 'search', className: 'w-full sm:w-72' }}
      actions={{ onChange: hub.actions.onSearch }}
    />
  </div>

  <AppHubView
    data={hub.data}
    state={hub.state}
    actions={{ onRetry: hub.actions.onRetry, onClearSearch: () => hub.actions.onSearch('') }}
  />
</div>

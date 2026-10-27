<script lang="ts">
  import Plus from '@lucide/svelte/icons/plus';

  import AppDialog from './components/acerola-app-dialog/acerola-app-dialog.svelte';
  import AppHubView from './components/acerola-app-hub-view/acerola-app-hub-view.svelte';
  import ActionButton from '$lib/components/acerola-action-button/acerola-action-button.svelte';
  import ConfirmDialog from '$lib/components/acerola-confirm-dialog/acerola-confirm-dialog.svelte';
  import PageHeader from '$lib/components/acerola-page-header/acerola-page-header.svelte';
  import TextField from '$lib/components/acerola-text-field/acerola-text-field.svelte';
  import { useAppFormModel } from '$lib/hooks/use-app-form/use-app-form.svelte';
  import { useAppHubModel } from '$lib/hooks/use-app-hub/use-app-hub.svelte';

  const hub = useAppHubModel();
  const form = useAppFormModel();
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
  >
    <ActionButton
      data={{ label: 'Cadastrar app' }}
      ui={{ icon: Plus }}
      actions={{ onClick: form.actions.onCreate }}
    />
  </PageHeader>

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
    actions={{
      onRetry: hub.actions.onRetry,
      onClearSearch: () => hub.actions.onSearch(''),
      onEdit: form.actions.onEdit,
      onCreate: form.actions.onCreate,
    }}
  />
</div>

<AppDialog
  data={form.data}
  ui={form.ui}
  state={form.state}
  actions={{
    onClose: form.actions.onClose,
    onChange: form.actions.onChange,
    onBlur: form.actions.onBlur,
    onSubmit: form.actions.onSubmit,
    onDelete: form.actions.onDeleteRequest,
  }}
/>

<ConfirmDialog
  data={{
    title: 'Apagar este app?',
    description: `"${form.data.values.name}" sai do hub. O app continua no ar; some só o atalho. Não dá para desfazer.`,
    confirmLabel: 'Apagar app',
    confirmingLabel: 'Apagando…',
  }}
  state={{
    isOpen: form.state.isDeleteOpen,
    isConfirming: form.state.isDeleting,
    error: form.state.deleteError,
  }}
  actions={{ onConfirm: form.actions.onDeleteConfirm, onCancel: form.actions.onDeleteCancel }}
/>

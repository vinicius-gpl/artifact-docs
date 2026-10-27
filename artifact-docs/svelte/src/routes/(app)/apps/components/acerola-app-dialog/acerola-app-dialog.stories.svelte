<script module lang="ts">
  import { defineMeta } from '@storybook/addon-svelte-csf';
  import { fn } from 'storybook/test';

  import AppDialog from './acerola-app-dialog.svelte';

  const typeSuggestions = ['Self-hosted', 'Na VPS', 'Ferramentas'];
  const empty = { name: '', url: '', typeName: '' };
  const jellyfin = { name: 'Jellyfin', url: 'https://jellyfin.exemplo.com', typeName: 'Self-hosted' };
  const actions = { onClose: fn(), onChange: fn(), onBlur: fn(), onSubmit: fn(), onDelete: fn() };
  const idle = { isOpen: true, isSubmitting: false, error: null, fieldErrors: {} };

  const { Story } = defineMeta({
    title: 'Features/Apps/AcerolaAppDialog',
    component: AppDialog,
    args: { data: { values: empty, typeSuggestions }, ui: { mode: 'new' }, state: idle, actions },
  });
</script>

<Story name="Default" />

<Story name="Edit" args={{ data: { values: jellyfin, typeSuggestions }, ui: { mode: 'edit' } }} />

<Story name="FirstApp" args={{ data: { values: empty, typeSuggestions: [] } }} />

<Story
  name="Saving"
  args={{ data: { values: jellyfin, typeSuggestions }, state: { ...idle, isSubmitting: true } }}
/>

<Story
  name="WithErrors"
  args={{
    state: {
      ...idle,
      error: 'Não consegui cadastrar o app. Tente de novo em instantes.',
      fieldErrors: {
        name: 'Informe o nome do app',
        url: 'Informe o endereço completo, começando com https://',
        typeName: 'Informe o tipo do app',
      },
    },
  }}
/>

<script module lang="ts">
  import { defineMeta } from '@storybook/addon-svelte-csf';
  import { fn } from 'storybook/test';

  import ArtifactEditDialog from './acerola-artifact-edit-dialog.svelte';

  const values = {
    title: 'Abertura do CNPJ',
    description: 'Natureza jurídica, contador, prazos e custos.',
    tags: 'cnpj, contador',
    status: 'active',
    visibility: 'private' as const,
  };
  const idle = { isOpen: true, isSubmitting: false, error: null, fieldErrors: {} };

  const { Story } = defineMeta({
    title: 'Features/Artifacts/AcerolaArtifactEditDialog',
    component: ArtifactEditDialog,
    args: {
      data: { values },
      ui: { hasStatus: true },
      state: idle,
      actions: { onClose: fn(), onChange: fn(), onBlur: fn(), onSubmit: fn() },
    },
  });
</script>

<Story name="Plan" />

<Story
  name="Document"
  args={{ data: { values: { ...values, status: '' } }, ui: { hasStatus: false } }}
/>

<Story name="Saving" args={{ state: { ...idle, isSubmitting: true } }} />

<!-- Caso limite: título apagado e a recusa do servidor. -->
<Story
  name="Errors"
  args={{
    data: { values: { ...values, title: '' } },
    state: {
      ...idle,
      error: 'Não consegui salvar a alteração. Tente de novo em instantes.',
      fieldErrors: { title: 'Informe o título' },
    },
  }}
/>

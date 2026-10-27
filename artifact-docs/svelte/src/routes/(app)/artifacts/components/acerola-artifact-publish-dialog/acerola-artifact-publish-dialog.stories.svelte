<script module lang="ts">
  import { defineMeta } from '@storybook/addon-svelte-csf';
  import { fn } from 'storybook/test';

  import PublishDialog from './acerola-artifact-publish-dialog.svelte';

  const kindOptions = [
    { value: 'document', label: 'Documento' },
    { value: 'plan', label: 'Planejamento' },
    { value: 'idea', label: 'Ideia' },
  ];
  const empty = { file: null, title: '', label: '', kind: 'document' as const, description: '', tags: '' };
  const filled = {
    file: new File(['<h1>Guia</h1>'], 'guia-de-instalacao.zip'),
    title: 'Guia de instalação',
    label: 'guia-de-instalacao',
    kind: 'document' as const,
    description: 'Passo a passo para subir o hub numa VPS nova.',
    tags: 'hub, deploy',
  };
  const actions = { onClose: fn(), onChange: fn(), onFileChange: fn(), onBlur: fn(), onSubmit: fn() };
  const idle = { isOpen: true, isSubmitting: false, error: null, fieldErrors: {} };

  const { Story } = defineMeta({
    title: 'Features/Artifacts/AcerolaArtifactPublishDialog',
    component: PublishDialog,
    args: {
      data: { values: empty, kindOptions, accept: '.html,.zip,.md,.pdf' },
      ui: { mode: 'new' },
      state: idle,
      actions,
    },
  });
</script>

<Story name="Default" />

<Story name="Filled" args={{ data: { values: filled, kindOptions, accept: '.html,.zip,.md,.pdf' } }} />

<Story
  name="NewVersion"
  args={{
    data: { values: { ...empty, label: 'abertura-cnpj', kind: 'plan' }, kindOptions, accept: '.html,.zip,.md,.pdf' },
    ui: { mode: 'version' },
  }}
/>

<Story
  name="Publishing"
  args={{ data: { values: filled, kindOptions, accept: '.html,.zip,.md,.pdf' }, state: { ...idle, isSubmitting: true } }}
/>

<Story
  name="WithErrors"
  args={{
    state: {
      ...idle,
      error: 'Outra publicação desta label terminou antes. Envie de novo.',
      fieldErrors: { file: 'Escolha o arquivo a publicar', title: 'Informe o título', label: 'Informe a label' },
    },
  }}
/>

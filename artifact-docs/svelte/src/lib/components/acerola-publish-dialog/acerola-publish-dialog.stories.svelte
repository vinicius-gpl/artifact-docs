<script module lang="ts">
  import { defineMeta } from '@storybook/addon-svelte-csf';
  import { fn } from 'storybook/test';

  import PublishDialog from './acerola-publish-dialog.svelte';

  const copy = {
    title: 'Publicar documento',
    description: 'O arquivo vira uma página servida pelo hub, com preview em Documentos.',
    submitLabel: 'Publicar documento',
    titlePlaceholder: 'Guia de instalação',
  };
  const accept = '.html,.zip,.md,.pdf';
  const empty = {
    file: null,
    title: '',
    label: '',
    description: '',
    tags: '',
    status: '',
    visibility: 'private' as const,
  };
  const filled = {
    file: new File(['<h1>Guia</h1>'], 'guia-de-instalacao.zip'),
    title: 'Guia de instalação',
    label: 'guia-de-instalacao',
    description: 'Passo a passo para subir o hub numa VPS nova.',
    tags: 'hub, deploy',
    status: '',
    visibility: 'private' as const,
  };
  const actions = {
    onClose: fn(),
    onChange: fn(),
    onFileChange: fn(),
    onBlur: fn(),
    onSubmit: fn(),
  };
  const idle = { isOpen: true, isSubmitting: false, error: null, fieldErrors: {} };

  const { Story } = defineMeta({
    title: 'Components/AcerolaPublishDialog',
    component: PublishDialog,
    args: { data: { values: empty, accept, copy }, ui: { mode: 'new' }, state: idle, actions },
  });
</script>

<Story name="Default" />

<Story name="Filled" args={{ data: { values: filled, accept, copy } }} />

<Story
  name="Plan"
  args={{
    data: {
      values: { ...empty, status: 'draft' },
      accept,
      copy: {
        title: 'Publicar planejamento',
        description: 'O plano vira uma página servida pelo hub, com preview em Planejamentos.',
        submitLabel: 'Publicar planejamento',
        titlePlaceholder: 'Abertura do CNPJ',
      },
    },
    ui: { mode: 'new', hasStatus: true },
  }}
/>

<Story
  name="NewVersion"
  args={{
    data: {
      values: { ...empty, label: 'abertura-cnpj' },
      accept,
      copy: {
        title: 'Publicar nova versão',
        description: 'A versão nova entra no ar no lugar da atual.',
        submitLabel: 'Publicar versão',
        titlePlaceholder: '',
      },
    },
    ui: { mode: 'version' },
  }}
/>

<Story
  name="Publishing"
  args={{ data: { values: filled, accept, copy }, state: { ...idle, isSubmitting: true } }}
/>

<Story
  name="WithErrors"
  args={{
    state: {
      ...idle,
      error: 'Outra publicação desta label terminou antes. Envie de novo.',
      fieldErrors: {
        file: 'Escolha o arquivo a publicar',
        title: 'Informe o título',
        label: 'Informe a label',
      },
    },
  }}
/>

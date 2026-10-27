<script module lang="ts">
  import { defineMeta } from '@storybook/addon-svelte-csf';
  import { fn } from 'storybook/test';

  import FileField from './acerola-file-field.svelte';

  const { Story } = defineMeta({
    title: 'Components/AcerolaFileField',
    component: FileField,
    args: {
      data: {
        label: 'Arquivo',
        name: 'file',
        accept: '.html,.zip,.md,.pdf',
        fileName: null,
        hint: 'HTML, site em .zip, Markdown ou PDF — até 100 MB.',
      },
      actions: { onChange: fn(), onBlur: fn() },
    },
  });
</script>

<Story name="Default" />

<Story
  name="Chosen"
  args={{ data: { label: 'Arquivo', name: 'file', fileName: 'guia-de-instalacao.zip' } }}
/>

<Story
  name="Required"
  args={{ data: { label: 'Arquivo', name: 'file', fileName: null, isRequired: true } }}
/>

<Story
  name="WithError"
  args={{
    data: { label: 'Arquivo', name: 'file', fileName: 'planilha.xlsx' },
    state: { error: 'Envie um arquivo .html, .zip, .md ou .pdf' },
  }}
/>

<Story
  name="Disabled"
  args={{ data: { label: 'Arquivo', name: 'file', fileName: 'guia.html' }, state: { isDisabled: true } }}
/>

<Story
  name="LongFileName"
  args={{
    data: {
      label: 'Arquivo',
      name: 'file',
      fileName: 'documentacao-completa-do-projeto-acerola-reader-versao-final-revisada-2026.zip',
    },
    ui: { className: 'w-64' },
  }}
/>

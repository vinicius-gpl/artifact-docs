<script module lang="ts">
  import { defineMeta } from '@storybook/addon-svelte-csf';
  import { fn } from 'storybook/test';

  import { type ArtifactView } from '$lib/hooks/use-artifact-detail/use-artifact-detail.svelte';
  import ArtifactViewer from './acerola-artifact-viewer.svelte';

  const artifact: ArtifactView = {
    title: 'Briefing do hub',
    description: 'O que é o hub, os módulos e a stack.',
    url: `data:text/html;charset=utf-8,${encodeURIComponent('<h1 style="font-family:sans-serif">Briefing do hub</h1>')}`,
    badges: [{ label: 'Documento', tone: 'brand' }],
    isPrivate: false,
    facts: [
      { label: 'Label', value: 'briefing-do-hub' },
      { label: 'Formato', value: 'Markdown' },
      { label: 'Tamanho', value: '12 KB' },
    ],
    versions: [
      { number: 2, version: 'v2', detail: '25/10/2026 · 12 KB', isCurrent: true },
      { number: 1, version: 'v1', detail: '18/10/2026 · 11 KB', isCurrent: false },
    ],
  };

  const idle = { isLoading: false, error: null };

  const withActions = {
    onRetry: fn(),
    onPublishVersion: fn(),
    onEdit: fn(),
    onTogglePrivate: fn(),
    onRestore: fn(),
    onDelete: fn(),
  };

  const { Story } = defineMeta({
    title: 'Features/Artifacts/AcerolaArtifactViewer',
    component: ArtifactViewer,
    args: { actions: { onRetry: fn() } },
  });
</script>

<Story name="Default" args={{ data: { artifact }, state: idle }} />

<Story name="Loading" args={{ data: { artifact: null }, state: { ...idle, isLoading: true } }} />

<Story name="LoadFailed" args={{ data: { artifact: null }, state: { ...idle, error: 'Artefato não encontrado.' } }} />

<!-- Caso limite: privado, versão única, sem descrição. -->
<Story
  name="PrivateSingleVersion"
  args={{
    data: { artifact: { ...artifact, isPrivate: true, description: null, versions: artifact.versions.slice(0, 1) } },
    state: idle,
  }}
/>

<Story name="WithActions" args={{ data: { artifact }, state: idle, actions: withActions }} />

<Story
  name="Restoring"
  args={{
    data: { artifact },
    state: { ...idle, pendingAction: { kind: 'restore', version: 1 } },
    actions: withActions,
  }}
/>

<Story
  name="ActionFailed"
  args={{
    data: { artifact },
    state: { ...idle, actionError: 'Versão não encontrada.' },
    actions: withActions,
  }}
/>

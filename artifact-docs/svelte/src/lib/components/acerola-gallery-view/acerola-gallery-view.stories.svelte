<script module lang="ts">
  import { defineMeta } from '@storybook/addon-svelte-csf';
  import { fn } from 'storybook/test';

  import { type PreviewItem } from '$lib/components/acerola-preview-card/acerola-preview-card.svelte';
  import GalleryView from './acerola-gallery-view.svelte';

  const page = (title: string) =>
    `data:text/html;charset=utf-8,${encodeURIComponent(`<h1 style="font-family:sans-serif">${title}</h1>`)}`;

  const items: PreviewItem[] = ['Briefing do hub', 'Manual da marca', 'Custo de IAM'].map(
    (title, index) => ({
      id: `a${index}`,
      title,
      description: 'Texto de exemplo.',
      href: '#',
      previewUrl: page(title),
      thumbUrl: null,
      meta: 'Editado há 2 dias',
      badges: [{ label: 'HTML', tone: 'neutral' }],
      isLocked: index === 1,
      isPinned: index === 0,
    }),
  );

  const data = {
    items,
    total: items.length,
    pinned: [{ id: 'a0', label: 'Briefing do hub', href: '#', meta: 'Documento' }],
    recent: [],
    caption: 'Artefatos',
    emptyTitle: 'Nenhum artefato ainda',
    emptyDescription: 'Publique o primeiro HTML, Markdown ou PDF.',
  };

  const idle = { isLoading: false, error: null, hasFilters: false };

  const { Story } = defineMeta({
    title: 'Components/AcerolaGalleryView',
    component: GalleryView,
    args: { actions: { onRetry: fn(), onClearFilters: fn(), onTogglePin: fn() } },
  });
</script>

<Story name="Default" args={{ data, state: idle }} />

<Story name="Loading" args={{ data, state: { ...idle, isLoading: true } }} />

<Story name="LoadFailed" args={{ data, state: { ...idle, error: 'O banco não respondeu.' } }} />

<Story name="Empty" args={{ data: { ...data, items: [], total: 0 }, state: idle }} />

<!-- Caso limite: o filtro escondeu tudo. -->
<Story name="FilteredOut" args={{ data: { ...data, items: [] }, state: { ...idle, hasFilters: true } }} />

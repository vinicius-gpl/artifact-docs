<script module lang="ts">
  import { defineMeta } from '@storybook/addon-svelte-csf';
  import { fn } from 'storybook/test';

  import { type PreviewItem } from '$lib/components/acerola-preview-card/acerola-preview-card.svelte';
  import PreviewGallery from './acerola-preview-gallery.svelte';

  const page = (title: string) =>
    `data:text/html;charset=utf-8,${encodeURIComponent(`<h1 style="font-family:sans-serif">${title}</h1>`)}`;

  function item(id: string, title: string, overrides: Partial<PreviewItem> = {}): PreviewItem {
    return {
      id,
      title,
      description: 'Texto de exemplo para a descrição.',
      href: '#',
      previewUrl: page(title),
      thumbUrl: null,
      meta: 'Editado há 3 dias',
      badges: [{ label: 'HTML', tone: 'neutral' }],
      isLocked: false,
      isPinned: false,
      ...overrides,
    };
  }

  const items = [
    item('a1', 'Briefing do hub', { isPinned: true }),
    item('a2', 'Manual da marca', { isLocked: true }),
    item('a3', 'Abertura do CNPJ', { badges: [{ label: 'Em andamento', tone: 'info' }] }),
    item('a4', 'Custo de IAM'),
  ];

  const { Story } = defineMeta({
    title: 'Components/AcerolaPreviewGallery',
    component: PreviewGallery,
    args: { actions: { onTogglePin: fn() } },
  });
</script>

<Story name="Default" args={{ data: { items, caption: 'Artefatos' } }} />

<Story name="AlwaysCards" args={{ data: { items, caption: 'Artefatos' }, state: { forceCards: true } }} />

<Story
  name="PinPending"
  args={{ data: { items, caption: 'Artefatos' }, state: { pendingPinId: 'a2' } }}
/>

<!-- Caso limite: um item só. -->
<Story name="SingleItem" args={{ data: { items: [items[0]!], caption: 'Artefatos' } }} />

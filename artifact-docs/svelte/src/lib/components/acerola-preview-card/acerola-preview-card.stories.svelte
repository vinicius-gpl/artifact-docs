<script module lang="ts">
  import { defineMeta } from '@storybook/addon-svelte-csf';
  import type { ComponentProps } from 'svelte';
  import { fn } from 'storybook/test';

  import PreviewCard, { type PreviewItem } from './acerola-preview-card.svelte';

  const page = (title: string) =>
    `data:text/html;charset=utf-8,${encodeURIComponent(`<h1 style="font-family:sans-serif">${title}</h1><p>Conteúdo de exemplo.</p>`)}`;

  const item: PreviewItem = {
    id: 'a1',
    title: 'Briefing do hub',
    description: 'O que é o hub, os módulos e a stack.',
    href: '#',
    previewUrl: page('Briefing do hub'),
    thumbUrl: null,
    meta: 'Editado há 2 dias',
    badges: [{ label: 'Markdown', tone: 'neutral' }],
    isLocked: false,
    isPinned: false,
  };

  const { Story } = defineMeta({
    title: 'Components/AcerolaPreviewCard',
    component: PreviewCard,
    args: { actions: { onTogglePin: fn() } },
  });
</script>

{#snippet narrow(args: ComponentProps<typeof PreviewCard>)}
  <div class="w-72">
    <PreviewCard {...args} />
  </div>
{/snippet}

<Story name="Default" args={{ data: { item } }} template={narrow} />

<Story
  name="PinnedAndPrivate"
  args={{ data: { item: { ...item, isPinned: true, isLocked: true } } }}
  template={narrow}
/>

<Story
  name="WithStatus"
  args={{
    data: {
      item: {
        ...item,
        title: 'Abertura do CNPJ',
        badges: [
          { label: 'Em andamento', tone: 'info' },
          { label: 'Markdown', tone: 'neutral' },
        ],
      },
    },
  }}
  template={narrow}
/>

<Story name="PinPending" args={{ data: { item }, state: { isPinPending: true } }} template={narrow} />

<!-- Caso limite: título longo, sem descrição nem selos. -->
<Story
  name="LongTitleNoDescription"
  args={{
    data: {
      item: {
        ...item,
        title: 'Relatório de custos da VPS com o comparativo de todos os provedores avaliados',
        description: null,
        badges: [],
      },
    },
  }}
  template={narrow}
/>

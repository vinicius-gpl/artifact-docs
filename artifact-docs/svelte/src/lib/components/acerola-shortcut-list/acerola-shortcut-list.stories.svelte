<script module lang="ts">
  import { defineMeta } from '@storybook/addon-svelte-csf';
  import Pin from '@lucide/svelte/icons/pin';

  import ShortcutList from './acerola-shortcut-list.svelte';

  const { Story } = defineMeta({
    title: 'Components/AcerolaShortcutList',
    component: ShortcutList,
  });
</script>

<Story
  name="Default"
  args={{
    data: {
      title: 'Fixados',
      emptyLabel: 'Fixe um item para ele aparecer aqui.',
      items: [
        { id: '1', label: 'Briefing do hub', href: '#', meta: 'Documento' },
        { id: '2', label: 'Abertura do CNPJ', href: '#', meta: 'Planejamento' },
      ],
    },
    ui: { icon: Pin },
  }}
/>

<Story
  name="Empty"
  args={{ data: { title: 'Recentes', emptyLabel: 'Abra um item para ele aparecer aqui.', items: [] } }}
/>

<!-- Caso limite: título comprido numa coluna estreita. -->
<Story
  name="LongLabel"
  args={{
    data: {
      title: 'Recentes',
      emptyLabel: '',
      items: [
        {
          id: '1',
          label: 'Relatório de custos da VPS com o comparativo de todos os provedores',
          href: '#',
          meta: 'Aberto há 2 dias',
        },
      ],
    },
  }}
/>

<script module lang="ts">
  import { defineMeta } from '@storybook/addon-svelte-csf';
  import { fn } from 'storybook/test';

  import { type AppTile } from '$lib/hooks/use-app-hub/use-app-hub.svelte';
  import AppLinkCard from './acerola-app-link-card.svelte';

  const app: AppTile = {
    id: 'app-1',
    name: 'Home Assistant',
    url: 'https://casa.exemplo.com',
    host: 'casa.exemplo.com',
    iconUrl: 'https://casa.exemplo.com/favicon.ico',
    status: { label: 'No ar', tone: 'success' },
    checkedLabel: 'Verificado há 2 minutos',
  };

  const { Story } = defineMeta({
    title: 'Features/Apps/AcerolaAppLinkCard',
    component: AppLinkCard,
  });
</script>

<Story name="Default" args={{ data: { app }, ui: { className: 'w-72' } }} />

<Story
  name="Offline"
  args={{
    data: { app: { ...app, status: { label: 'Fora do ar', tone: 'danger' } } },
    ui: { className: 'w-72' },
  }}
/>

<!-- Caso limite: nome e endereço longos num cartão estreito. -->
<Story
  name="LongName"
  args={{
    data: {
      app: {
        ...app,
        name: 'Painel de monitoramento dos serviços da casa',
        host: 'monitoramento.servicos.casa.exemplo.com',
      },
    },
    ui: { className: 'w-60' },
  }}
/>

<Story name="Editable" args={{ data: { app }, ui: { className: 'w-72' }, actions: { onEdit: fn() } }} />

<script module lang="ts">
  import { defineMeta } from '@storybook/addon-svelte-csf';
  import { fn } from 'storybook/test';

  import { type AppGroup } from '$lib/hooks/use-app-hub/use-app-hub.svelte';
  import AppHubView from './acerola-app-hub-view.svelte';

  const groups: AppGroup[] = [
    {
      id: 't1',
      name: 'Self-hosted',
      apps: ['Jellyfin', 'Nextcloud'].map((name, index) => ({
        id: `a${index}`,
        name,
        url: 'https://exemplo.com',
        host: 'exemplo.com',
        iconUrl: 'https://exemplo.com/favicon.ico',
        status: { label: 'No ar', tone: 'success' },
        checkedLabel: 'Verificado há 1 minuto',
      })),
    },
  ];

  const idle = { isLoading: false, error: null, search: '' };

  const { Story } = defineMeta({
    title: 'Features/Apps/AcerolaAppHubView',
    component: AppHubView,
    args: { actions: { onRetry: fn(), onClearSearch: fn() } },
  });
</script>

<Story name="Default" args={{ data: { groups, total: 2 }, state: idle }} />

<Story name="Loading" args={{ data: { groups, total: 2 }, state: { ...idle, isLoading: true } }} />

<Story name="LoadFailed" args={{ data: { groups, total: 2 }, state: { ...idle, error: 'A API está fora do ar.' } }} />

<Story name="Empty" args={{ data: { groups: [], total: 0 }, state: idle }} />

<!-- Caso limite: a busca escondeu tudo. -->
<Story name="SearchedOut" args={{ data: { groups: [], total: 2 }, state: { ...idle, search: 'zzz' } }} />

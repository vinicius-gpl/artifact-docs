<script module lang="ts">
  import { defineMeta } from '@storybook/addon-svelte-csf';

  import { type AppTile } from '$lib/hooks/use-app-hub/use-app-hub.svelte';
  import AppGroup from './acerola-app-group.svelte';

  function tile(id: string, name: string, online = true): AppTile {
    return {
      id,
      name,
      url: `https://${id}.exemplo.com`,
      host: `${id}.exemplo.com`,
      iconUrl: `https://${id}.exemplo.com/favicon.ico`,
      status: online ? { label: 'No ar', tone: 'success' } : { label: 'Fora do ar', tone: 'danger' },
      checkedLabel: 'Verificado há 1 minuto',
    };
  }

  const { Story } = defineMeta({
    title: 'Features/Apps/AcerolaAppGroup',
    component: AppGroup,
  });
</script>

<Story
  name="Default"
  args={{
    data: {
      group: {
        id: 't1',
        name: 'Self-hosted',
        apps: [tile('jellyfin', 'Jellyfin'), tile('casa', 'Home Assistant'), tile('nuvem', 'Nextcloud', false)],
      },
    },
  }}
/>

<!-- Caso limite: um app só. -->
<Story name="SingleApp" args={{ data: { group: { id: 't2', name: 'Na VPS', apps: [tile('kuma', 'Uptime Kuma')] } } }} />

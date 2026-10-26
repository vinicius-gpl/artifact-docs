<script module lang="ts">
  import { defineMeta } from '@storybook/addon-svelte-csf';
  import { fn } from 'storybook/test';

  import { NAV_ITEMS } from '$lib/navigation/navigation';
  import TopNav from './acerola-top-nav.svelte';

  const data = {
    items: NAV_ITEMS,
    homeHref: '#',
    user: { name: 'Você Mesmo', email: 'voce@exemplo.com', avatarUrl: null },
  };

  const { Story } = defineMeta({
    title: 'Components/AcerolaTopNav',
    component: TopNav,
    args: { actions: { onSignOut: fn() } },
  });
</script>

<Story name="Default" args={{ data, state: { activeKey: 'artifacts' } }} />

<Story name="AnotherScreen" args={{ data, state: { activeKey: 'ideas' } }} />

<Story name="SigningOut" args={{ data, state: { activeKey: 'apps', isSigningOut: true } }} />

<!-- Caso limite: endereço fora do menu, nada aceso. -->
<Story name="NothingActive" args={{ data, state: { activeKey: undefined } }} />

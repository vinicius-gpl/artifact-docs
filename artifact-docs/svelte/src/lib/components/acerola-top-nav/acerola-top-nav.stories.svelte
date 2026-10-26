<script module lang="ts">
  import { defineMeta } from '@storybook/addon-svelte-csf';

  import { NAV_ITEMS } from '$lib/navigation/navigation';
  import TopNav from './acerola-top-nav.svelte';

  const data = { items: NAV_ITEMS, homeHref: '#' };

  const { Story } = defineMeta({
    title: 'Components/AcerolaTopNav',
    component: TopNav,
  });
</script>

<Story name="Default" args={{ data, state: { activeKey: 'artifacts' } }} />

<Story name="AnotherScreen" args={{ data, state: { activeKey: 'brand' } }} />

<!-- Caso limite: endereço fora do menu, nada aceso. -->
<Story name="NothingActive" args={{ data, state: { activeKey: undefined } }} />

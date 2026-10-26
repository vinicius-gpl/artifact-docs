import { describe, expect, it } from 'vitest';

import { HOME_PATH, NAV_ITEMS } from '$lib/navigation/navigation';

import { useTopNavModel } from './use-top-nav.svelte';

const { page } = await import('$app/state');

function openScreen(pathname: string) {
  page.url = new URL(`http://localhost:5010${pathname}`);
}

describe('useTopNavModel', () => {
  // feliz
  it('lights the item of the screen that is open, also in its subpages', () => {
    const model = useTopNavModel();

    openScreen('/artifacts/abertura-cnpj');
    expect(model.state.activeKey).toBe('artifacts');

    openScreen('/apps');
    expect(model.state.activeKey).toBe('apps');
    expect(model.data).toEqual({ items: NAV_ITEMS, homeHref: HOME_PATH });
  });

  // triste
  it('lights nothing on an address outside the menu', () => {
    const model = useTopNavModel();
    openScreen('/nao-existe');

    expect(model.state.activeKey).toBeUndefined();
  });
});

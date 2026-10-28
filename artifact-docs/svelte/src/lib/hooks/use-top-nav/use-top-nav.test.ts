import { QueryClient } from '@tanstack/svelte-query';
import { render } from '@testing-library/svelte';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { HOME_PATH, NAV_ITEMS } from '$lib/navigation/navigation';
import Harness from './use-top-nav-harness.test.svelte';
import { type TopNavModel } from './use-top-nav.svelte';

vi.mock('$lib/api/auth.api', () => ({ authApi: { mockSignOut: vi.fn() } }));
vi.mock('$lib/auth/neon-auth.client', () => ({ signOutOfNeon: vi.fn() }));
const mode = vi.hoisted(() => ({ mocks: true }));
vi.mock('$lib/api/http-client', () => ({
  get USE_MOCKS() {
    return mode.mocks;
  },
}));

const { authApi } = await import('$lib/api/auth.api');
const { signOutOfNeon } = await import('$lib/auth/neon-auth.client');
const { goto } = await import('$app/navigation');
const { page } = await import('$app/state');

function openScreen(pathname: string) {
  page.url = new URL(`http://localhost:5010${pathname}`);
}

function mountModel(queryClient = new QueryClient()): TopNavModel {
  let model!: TopNavModel;
  render(Harness, { props: { queryClient, onReady: (ready) => (model = ready) } });

  return model;
}

beforeEach(() => {
  mode.mocks = true;
  vi.mocked(authApi.mockSignOut).mockReset().mockResolvedValue(undefined);
  vi.mocked(signOutOfNeon).mockReset().mockResolvedValue(undefined);
  vi.mocked(goto).mockClear();
});

describe('useTopNavModel', () => {
  // feliz
  it('lights the item of the screen that is open, also in its subpages', () => {
    const model = mountModel();

    openScreen('/artifacts/abertura-cnpj');
    expect(model.state.activeKey).toBe('artifacts');

    openScreen('/apps');
    expect(model.state.activeKey).toBe('apps');
    expect(model.data).toMatchObject({ items: NAV_ITEMS, homeHref: HOME_PATH });
  });

  it('signs out of the mock session, clears the cached data and goes to the login', async () => {
    const queryClient = new QueryClient();
    queryClient.setQueryData(['artifacts'], { items: [] });
    const model = mountModel(queryClient);

    model.actions.onSignOut();

    await vi.waitFor(() => expect(goto).toHaveBeenCalledWith('/login'));
    expect(authApi.mockSignOut).toHaveBeenCalled();
    expect(queryClient.getQueryData(['artifacts'])).toBeUndefined();
  });

  it('signs out of Neon outside mock mode', async () => {
    mode.mocks = false;
    const model = mountModel();

    model.actions.onSignOut();

    await vi.waitFor(() => expect(signOutOfNeon).toHaveBeenCalled());
  });

  // triste
  it('lights nothing on an address outside the menu', () => {
    const model = mountModel();
    openScreen('/nao-existe');

    expect(model.state.activeKey).toBeUndefined();
  });
});

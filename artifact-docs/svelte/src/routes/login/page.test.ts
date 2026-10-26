import { isRedirect } from '@sveltejs/kit';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { authApi } from '$lib/api/auth.api';
import { ApiError } from '$lib/api/http-client';

import { load } from './+page';

vi.mock('$lib/api/auth.api', () => ({ authApi: { me: vi.fn(), setup: vi.fn() } }));

const me = vi.mocked(authApi.me);
const setup = vi.mocked(authApi.setup);
const at = (path: string) => ({ url: new URL(`https://hub.exemplo.com${path}`) });

describe('a tela de login', () => {
  beforeEach(() => {
    me.mockReset();
    setup.mockReset().mockResolvedValue({ needsSetup: false });
  });

  // feliz
  it('sends who is already signed in straight to where they were going', async () => {
    me.mockResolvedValue({ id: 'u1', email: 'voce@exemplo.com', name: 'Você', image: null });

    const thrown: unknown = await load(at('/login?next=%2Fideas')).catch((error: unknown) => error);

    expect(isRedirect(thrown)).toBe(true);
    expect(thrown).toMatchObject({ location: '/ideas' });
  });

  // triste
  it('shows the login when there is no session', async () => {
    me.mockRejectedValue(new ApiError(401, 'Entre para continuar.'));

    await expect(load(at('/login'))).resolves.toEqual({
      returnTo: '/',
      error: null,
      needsSetup: false,
    });
  });

  it('offers the first access only while the account does not exist', async () => {
    me.mockRejectedValue(new ApiError(401, 'Entre para continuar.'));
    setup.mockResolvedValue({ needsSetup: true });

    await expect(load(at('/login'))).resolves.toMatchObject({ needsSetup: true });
  });

  it('hides the first access when the API cannot say', async () => {
    me.mockRejectedValue(new ApiError(401, 'Entre para continuar.'));
    setup.mockRejectedValue(new ApiError(503, 'A API está fora do ar.'));

    await expect(load(at('/login'))).resolves.toMatchObject({ needsSetup: false });
  });

  it('keeps the denied warning without asking the API again', async () => {
    await expect(load(at('/login?error=forbidden'))).resolves.toEqual({
      returnTo: '/',
      error: 'forbidden',
      needsSetup: false,
    });
    expect(me).not.toHaveBeenCalled();
  });
});

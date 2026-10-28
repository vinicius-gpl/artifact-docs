import { isRedirect } from '@sveltejs/kit';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { authApi } from '$lib/api/auth.api';
import { ApiError } from '$lib/api/http-client';

import { load } from './+page';

vi.mock('$lib/api/auth.api', () => ({ authApi: { me: vi.fn() } }));

const me = vi.mocked(authApi.me);
const at = (path: string) => ({ url: new URL(`https://hub.exemplo.com${path}`) });

describe('a tela de login', () => {
  beforeEach(() => {
    me.mockReset();
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
    });
  });

  it('keeps the denied warning without asking the API again', async () => {
    await expect(load(at('/login?error=forbidden'))).resolves.toEqual({
      returnTo: '/',
      error: 'forbidden',
    });
    expect(me).not.toHaveBeenCalled();
  });
});

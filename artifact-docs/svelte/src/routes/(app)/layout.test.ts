import { isRedirect } from '@sveltejs/kit';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { authApi } from '$lib/api/auth.api';
import { ApiError } from '$lib/api/http-client';

import { load } from './+layout';

vi.mock('$lib/api/auth.api', () => ({ authApi: { me: vi.fn() } }));

const me = vi.mocked(authApi.me);
const at = (path: string) => ({ url: new URL(`https://hub.exemplo.com${path}`) });

async function thrownBy(path: string): Promise<unknown> {
  return load(at(path)).catch((error: unknown) => error);
}

describe('a guarda de sessão', () => {
  beforeEach(() => {
    me.mockReset();
  });

  // feliz
  it('hands the signed-in person to the screens', async () => {
    const user = { id: 'u1', email: 'voce@exemplo.com', name: 'Você', image: null };
    me.mockResolvedValue(user);

    await expect(load(at('/apps'))).resolves.toEqual({ user });
  });

  // triste
  it('sends who has no session to the login, remembering where they were', async () => {
    me.mockRejectedValue(new ApiError(401, 'Entre para continuar.'));

    const thrown = await thrownBy('/artifacts/briefing');

    expect(isRedirect(thrown)).toBe(true);
    expect(thrown).toMatchObject({ location: '/login?next=%2Fartifacts%2Fbriefing' });
  });

  it('sends an account outside the allow list back to the login, with the warning', async () => {
    me.mockRejectedValue(new ApiError(403, 'Sem acesso.'));

    expect(await thrownBy('/apps')).toMatchObject({ location: '/login?error=forbidden' });
  });

  it('lets any other failure reach the error page', async () => {
    me.mockRejectedValue(new ApiError(503, 'A API está fora do ar.'));

    expect(await thrownBy('/apps')).toBeInstanceOf(ApiError);
  });
});

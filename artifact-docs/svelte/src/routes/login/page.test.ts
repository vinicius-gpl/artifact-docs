import { isRedirect } from '@sveltejs/kit';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { authApi } from '$lib/api/auth.api';
import { ApiError } from '$lib/api/http-client';

import { load } from './+page';

vi.mock('$lib/api/auth.api', () => ({ authApi: { me: vi.fn() } }));

const me = vi.mocked(authApi.me);
const at = (path: string) => ({ url: new URL(`https://acerola.exemplo.com${path}`) });

describe('a rota do login', () => {
  beforeEach(() => {
    me.mockReset();
    me.mockRejectedValue(new ApiError(401, 'Unauthorized'));
  });

  // feliz
  it('remembers where to go back to after signing in', async () => {
    await expect(load(at('/login?next=%2Fartifacts%3Fkind%3Dplan'))).resolves.toEqual({
      returnTo: '/artifacts?kind=plan',
    });
  });

  it('sends who is already signed in straight to where they were going', async () => {
    me.mockResolvedValue({ id: 'user-1' } as Awaited<ReturnType<typeof authApi.me>>);

    const thrown: unknown = await load(at('/login?next=%2Fartifacts')).catch((e: unknown) => e);

    expect(isRedirect(thrown)).toBe(true);
    expect(thrown).toMatchObject({ status: 307, location: '/artifacts' });
  });

  // triste
  /* Um `next` de fora faria do login um trampolim para outro site. */
  it('ignores a return address of another site', async () => {
    await expect(load(at('/login?next=https%3A%2F%2Fgolpe.exemplo.com'))).resolves.toEqual({
      returnTo: '/',
    });
  });
});

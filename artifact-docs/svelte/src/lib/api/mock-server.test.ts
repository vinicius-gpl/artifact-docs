import { beforeEach, describe, expect, it } from 'vitest';

import { type AppHub } from '$lib/contracts/app-link.schema';
import { type ArtifactDetail, type ArtifactList } from '$lib/contracts/artifact.schema';

import { ApiError } from './http-client';
import { htmlUrl, MOCK_PASSWORD, MOCK_USER } from './mock-fixtures';
import { resetMockAccount, setMockSession } from './mock-auth';
import { handleMockRequest, matchPath, resetMockDatabase } from './mock-server';

/* Sem atraso: o teste mede a resposta, não a espera. */
const request = <T>(path: string, options = {}) => handleMockRequest<T>(path, options, 0);

beforeEach(() => {
  resetMockDatabase(Date.parse('2026-10-25T12:00:00.000Z'));
  resetMockAccount();
  setMockSession(MOCK_USER.email);
});

describe('matchPath', () => {
  // feliz
  it('reads the params out of the path', () => {
    expect(matchPath('/artifacts/:label', '/artifacts/abertura-cnpj')).toEqual({
      label: 'abertura-cnpj',
    });
  });

  // triste
  it('refuses a path with another shape', () => {
    expect(matchPath('/artifacts/:label', '/artifacts')).toBeNull();
    expect(matchPath('/apps', '/artifacts')).toBeNull();
  });
});

describe('handleMockRequest', () => {
  // feliz
  it('lists artifacts of one kind, newest first', async () => {
    const list = await request<ArtifactList>('/artifacts', { query: { kind: 'plan' } });

    expect(list.items.every((artifact) => artifact.kind === 'plan')).toBe(true);
    expect(list.items[0]?.label).toBe('abertura-cnpj');
  });

  it('records the visit when an artifact is opened', async () => {
    const detail = await request<ArtifactDetail>('/artifacts/erp-pessoal');

    expect(detail.lastViewedAt).not.toBeNull();
    expect(detail.versions).toHaveLength(detail.currentVersion);
  });

  it('pins an artifact and keeps it pinned on the next read', async () => {
    await request('/artifacts/erp-pessoal', { method: 'PATCH', body: { isPinned: true } });

    const detail = await request<ArtifactDetail>('/artifacts/erp-pessoal');
    expect(detail.isPinned).toBe(true);
  });

  it('answers the apps hub in the contract shape', async () => {
    const hub = await request<AppHub>('/apps');

    expect(hub.types.length).toBeGreaterThan(0);
    expect(hub.apps.length).toBeGreaterThan(0);
  });

  // triste
  it('answers 404 for an artifact that does not exist', async () => {
    await expect(request('/artifacts/nao-existe')).rejects.toMatchObject({ status: 404 });
  });

  it('refuses a patch outside the contract', async () => {
    await expect(
      request('/artifacts/erp-pessoal', { method: 'PATCH', body: { isPinned: 'sim' } }),
    ).rejects.toBeInstanceOf(ApiError);
  });

  it('answers 404 for a route the mock does not have, naming it', async () => {
    await expect(request('/tokens', { method: 'POST' })).rejects.toThrow(
      'Rota de mock não encontrada: POST /tokens',
    );
  });

  it('hands out a copy, so the screen cannot change the database by reference', async () => {
    const first = await request<ArtifactList>('/artifacts');
    first.items[0]!.title = 'mexido';

    const second = await request<ArtifactList>('/artifacts');
    expect(second.items[0]?.title).not.toBe('mexido');
  });

  it('answers 401 without the mock session, like the Go API will', async () => {
    setMockSession(null);

    await expect(request('/artifacts')).rejects.toMatchObject({ status: 401 });
  });
});

describe('mock sign-in', () => {
  // feliz
  it('opens the mock session and then answers who is signed in', async () => {
    setMockSession(null);

    await request('/auth/mock/sign-in', {
      method: 'POST',
      body: { email: MOCK_USER.email, password: MOCK_PASSWORD },
    });

    await expect(request('/auth/me')).resolves.toMatchObject({ email: 'voce@exemplo.com' });
  });

  // triste
  it('refuses a wrong password with the same message as a wrong e-mail', async () => {
    setMockSession(null);

    await expect(
      request('/auth/mock/sign-in', {
        method: 'POST',
        body: { email: MOCK_USER.email, password: 'errada' },
      }),
    ).rejects.toThrow('E-mail ou senha incorretos.');
    await expect(request('/auth/me')).rejects.toMatchObject({ status: 401 });
  });

  it('closes the session on sign-out', async () => {
    await request('/auth/mock/sign-out', { method: 'POST' });

    await expect(request('/auth/me')).rejects.toMatchObject({ status: 401 });
  });
});

describe('handleMockRequest — publishing', () => {
  function publishForm(fields: Record<string, string>, file?: File): FormData {
    const form = new FormData();
    if (file) form.append('file', file, file.name);
    for (const [field, value] of Object.entries(fields)) form.append(field, value);

    return form;
  }

  // feliz
  it('publishes a new artifact and then a second version of it', async () => {
    const html = new File(['<h1>Guia</h1>'], 'guia.html', { type: 'text/html' });
    const created = await request<{ currentVersion: number; url: string }>('/artifacts/guia-novo', {
      method: 'PUT',
      body: publishForm({ title: 'Guia novo', kind: 'document' }, html),
    });
    const updated = await request<{ currentVersion: number; title: string }>(
      '/artifacts/guia-novo',
      {
        method: 'PUT',
        body: publishForm({}, new File(['# Guia'], 'guia.md')),
      },
    );

    expect(created.currentVersion).toBe(1);
    expect(created.url).toContain(encodeURIComponent('<h1>Guia</h1>'));
    expect(updated).toMatchObject({ currentVersion: 2, title: 'Guia novo' });
  });

  // triste
  it('refuses a new artifact without title, like the Go', async () => {
    const html = new File(['<p>x</p>'], 'x.html');

    await expect(
      request('/artifacts/sem-titulo', {
        method: 'PUT',
        body: publishForm({ kind: 'idea' }, html),
      }),
    ).rejects.toMatchObject({ status: 400 });
  });

  // feliz
  it('restores an older version and deletes an artifact', async () => {
    const list = await request<ArtifactList>('/artifacts');
    const withVersions = list.items.find((item) => item.currentVersion > 1);
    if (!withVersions) throw new Error('the fixtures need an artifact with versions');

    const restored = await request<{ currentVersion: number }>(
      `/artifacts/${withVersions.label}/restore`,
      { method: 'POST', body: { version: 1 } },
    );
    await request(`/artifacts/${withVersions.label}`, { method: 'DELETE' });

    expect(restored.currentVersion).toBe(1);
    await expect(request(`/artifacts/${withVersions.label}`)).rejects.toMatchObject({
      status: 404,
    });
  });

  // triste
  it('answers 404 when restoring a version that does not exist', async () => {
    const list = await request<ArtifactList>('/artifacts');
    const label = list.items[0]?.label ?? '';

    await expect(
      request(`/artifacts/${label}/restore`, { method: 'POST', body: { version: 999 } }),
    ).rejects.toMatchObject({ status: 404 });
  });
});

describe('handleMockRequest — apps', () => {
  const input = { name: 'Immich', url: 'https://fotos.exemplo.com', typeName: 'Mídia' };

  // feliz
  it('creates an app in a new type, then moves it and drops the empty type', async () => {
    const created = await request<{ id: string; typeId: string }>('/apps', {
      method: 'POST',
      body: input,
    });
    const hub = await request<AppHub>('/apps');
    expect(hub.types.some((type) => type.id === created.typeId && type.name === 'Mídia')).toBe(
      true,
    );

    await request(`/apps/${created.id}`, {
      method: 'PATCH',
      body: { ...input, typeName: hub.types[0]!.name.toUpperCase() },
    });

    const after = await request<AppHub>('/apps');
    expect(after.types.some((type) => type.name === 'Mídia')).toBe(false);
  });

  // triste
  it('refuses an address without http and answers 404 for an app that does not exist', async () => {
    await expect(
      request('/apps', { method: 'POST', body: { ...input, url: 'fotos.exemplo.com' } }),
    ).rejects.toMatchObject({ status: 400 });
    await expect(request('/apps/nao-existe', { method: 'DELETE' })).rejects.toMatchObject({
      status: 404,
    });
  });
});

describe('htmlUrl', () => {
  // feliz
  it('falls back to a data: page where the browser has no blob URLs', () => {
    expect(htmlUrl('<h1>Oi</h1>')).toBe(
      `data:text/html;charset=utf-8,${encodeURIComponent('<h1>Oi</h1>')}`,
    );
  });

  // triste
  it('uses a blob: page when the browser can open it in a new tab', () => {
    const original = URL.createObjectURL;
    URL.createObjectURL = () => 'blob:http://localhost/1';

    expect(htmlUrl('<h1>Oi</h1>')).toBe('blob:http://localhost/1');
    URL.createObjectURL = original;
  });
});

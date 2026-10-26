import { beforeEach, describe, expect, it } from 'vitest';

import { type AppHub } from '$lib/contracts/app-link.schema';
import { type ArtifactDetail, type ArtifactList } from '$lib/contracts/artifact.schema';

import { ApiError } from './http-client';
import { handleMockRequest, matchPath, resetMockDatabase } from './mock-server';

/* Sem atraso: o teste mede a resposta, não a espera. */
const request = <T>(path: string, options = {}) => handleMockRequest<T>(path, options, 0);

beforeEach(() => {
  resetMockDatabase(Date.parse('2026-10-25T12:00:00.000Z'));
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
});

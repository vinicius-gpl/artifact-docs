import { render, waitFor } from '@testing-library/svelte';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { ApiError } from '$lib/api/http-client';
import { createMockDatabase } from '$lib/api/mock-fixtures';
import Harness from './use-app-hub-harness.test.svelte';
import { groupApps, toAppTile, type AppHubModel } from './use-app-hub.svelte';

vi.mock('$lib/api/apps.api', () => ({ appsApi: { hub: vi.fn() } }));

const { appsApi } = await import('$lib/api/apps.api');

const NOW = Date.parse('2026-10-25T12:00:00.000Z');
const hub = () => createMockDatabase(NOW).appHub;

function mountModel(): AppHubModel {
  let model!: AppHubModel;
  render(Harness, { props: { onReady: (ready) => (model = ready) } });

  return model;
}

beforeEach(() => {
  vi.mocked(appsApi.hub).mockReset().mockResolvedValue(hub());
});

describe('toAppTile', () => {
  // feliz
  it('uses the site favicon when there is no logo', () => {
    const tile = toAppTile(hub().apps[0]!, NOW);

    expect(tile).toMatchObject({
      host: 'jellyfin.exemplo.com',
      iconUrl: 'https://jellyfin.exemplo.com/favicon.ico',
      status: { label: 'No ar', tone: 'success' },
      checkedLabel: 'Nunca verificado',
    });
  });

  // triste
  it('marks an app that is down in red, with the last check', () => {
    const app = { ...hub().apps[2]!, checkedAt: '2026-10-25T11:55:00.000Z' };

    expect(toAppTile(app, NOW)).toMatchObject({
      status: { label: 'Fora do ar', tone: 'danger' },
      checkedLabel: 'Verificado há 5 minutos',
    });
  });
});

describe('groupApps', () => {
  // feliz
  it('groups by type in the order the person chose', () => {
    expect(groupApps(hub(), '', NOW).map((group) => group.name)).toEqual([
      'Self-hosted',
      'Na VPS',
      'Ferramentas',
    ]);
  });

  // triste
  it('drops the types left empty by the search', () => {
    const groups = groupApps(hub(), 'kuma', NOW);

    expect(groups.map((group) => group.name)).toEqual(['Na VPS']);
    expect(groups[0]?.apps.map((app) => app.name)).toEqual(['Uptime Kuma']);
  });
});

describe('useAppHubModel', () => {
  // feliz
  it('loads the hub and counts what is down', async () => {
    const model = mountModel();

    await waitFor(() => expect(model.state.isLoading).toBe(false));
    expect(model.data.total).toBe(7);
    expect(model.data.offline).toBe(1);
  });

  // triste
  it('shows the reason when the hub does not load', async () => {
    vi.mocked(appsApi.hub).mockRejectedValue(new ApiError(503, 'A API está fora do ar.'));
    const model = mountModel();

    await waitFor(() => expect(model.state.error).toBe('A API está fora do ar.'));
  });
});

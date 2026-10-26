import { render, waitFor } from '@testing-library/svelte';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { ApiError } from '$lib/api/http-client';
import { createMockDatabase } from '$lib/api/mock-fixtures';
import { type Artifact, type ArtifactKind } from '$lib/contracts/artifact.schema';
import Harness from './use-artifact-gallery-harness.test.svelte';
import {
  filterArtifacts,
  pinnedOf,
  recentOf,
  toPreviewItem,
  type ArtifactGalleryModel,
} from './use-artifact-gallery.svelte';

vi.mock('$lib/api/artifacts.api', () => ({
  artifactsApi: { list: vi.fn(), update: vi.fn() },
}));

const { artifactsApi } = await import('$lib/api/artifacts.api');

const NOW = Date.parse('2026-10-25T12:00:00.000Z');

function artifacts(): Artifact[] {
  return createMockDatabase(NOW).artifacts;
}

function byLabel(label: string): Artifact {
  const found = artifacts().find((artifact) => artifact.label === label);
  if (!found) throw new Error(label);

  return found;
}

function mountModel(kind?: ArtifactKind): ArtifactGalleryModel {
  let model!: ArtifactGalleryModel;
  render(Harness, { props: { kind, onReady: (ready) => (model = ready) } });

  return model;
}

beforeEach(() => {
  vi.mocked(artifactsApi.list).mockReset().mockResolvedValue({ items: artifacts() });
  vi.mocked(artifactsApi.update).mockReset();
});

describe('toPreviewItem', () => {
  // feliz
  it('turns a plan into a card with status, kind and format badges', () => {
    const item = toPreviewItem(byLabel('abertura-cnpj'), true, NOW);

    expect(item).toMatchObject({
      title: 'Abertura do CNPJ',
      href: '/artifacts/abertura-cnpj',
      meta: 'Editado há 1 dia',
      isLocked: true,
      isPinned: true,
    });
    expect(item.badges.map((badge) => badge.label)).toEqual([
      'Em andamento',
      'Planejamento',
      'Markdown',
    ]);
  });

  // triste
  it('leaves the kind out when the gallery already has a fixed kind', () => {
    const item = toPreviewItem(byLabel('abertura-cnpj'), false, NOW);

    expect(item.badges.map((badge) => badge.label)).not.toContain('Planejamento');
  });
});

describe('filterArtifacts', () => {
  // feliz
  it('searches title, description and tags, ignoring case', () => {
    const found = filterArtifacts(artifacts(), { search: 'MARCA', kind: '', status: '' });

    expect(found.map((artifact) => artifact.label)).toContain('manual-da-marca');
    expect(found.map((artifact) => artifact.label)).toContain('gerador-de-paletas');
  });

  it('combines kind and status', () => {
    const found = filterArtifacts(artifacts(), { search: '', kind: 'plan', status: 'active' });

    expect(found.map((artifact) => artifact.label).sort()).toEqual([
      'abertura-cnpj',
      'tunel-wireguard',
    ]);
  });

  // triste
  it('returns nothing when no item matches', () => {
    expect(filterArtifacts(artifacts(), { search: 'zzz', kind: '', status: '' })).toEqual([]);
  });
});

describe('pinnedOf and recentOf', () => {
  // feliz
  it('lists the pinned ones and the recently opened ones', () => {
    expect(pinnedOf(artifacts()).map((link) => link.label)).toContain('Abertura do CNPJ');
    expect(recentOf(artifacts(), NOW).map((link) => link.label)).toEqual([
      'Briefing do artifact-docs',
      'Abertura do CNPJ',
      'Manual da marca Grupo Azuos',
    ]);
  });

  // triste
  it('leaves never-opened items out of the recent list', () => {
    const neverOpened = artifacts().map((artifact) => ({ ...artifact, lastViewedAt: null }));

    expect(recentOf(neverOpened, NOW)).toEqual([]);
  });
});

describe('useArtifactGalleryModel', () => {
  // feliz
  it('loads the list for the fixed kind and maps it to cards', async () => {
    const model = mountModel('idea');

    await waitFor(() => expect(model.state.isLoading).toBe(false));
    expect(artifactsApi.list).toHaveBeenCalledWith({ kind: 'idea' });
    expect(model.data.items.length).toBeGreaterThan(0);
  });

  it('filters as the person types and clears back to everything', async () => {
    const model = mountModel();
    await waitFor(() => expect(model.state.isLoading).toBe(false));

    model.actions.onSearch('wireguard');
    expect(model.data.items.map((item) => item.title)).toEqual(['Túnel WireGuard']);
    expect(model.state.hasFilters).toBe(true);

    model.actions.onClearFilters();
    expect(model.data.items).toHaveLength(model.data.total);
  });

  it('pins by id through the API', async () => {
    vi.mocked(artifactsApi.update).mockResolvedValue(byLabel('erp-pessoal'));
    const model = mountModel();
    await waitFor(() => expect(model.state.isLoading).toBe(false));

    model.actions.onTogglePin(byLabel('erp-pessoal').id);

    await waitFor(() =>
      expect(artifactsApi.update).toHaveBeenCalledWith('erp-pessoal', { isPinned: true }),
    );
  });

  // triste
  it('shows the reason when the list does not load', async () => {
    vi.mocked(artifactsApi.list).mockRejectedValue(new ApiError(500, 'O banco não respondeu.'));
    const model = mountModel();

    await waitFor(() => expect(model.state.error).toBe('O banco não respondeu.'));
  });

  it('ignores an unknown kind or status in the filters', async () => {
    const model = mountModel();
    await waitFor(() => expect(model.state.isLoading).toBe(false));

    model.actions.onKindChange('banana');
    model.actions.onStatusChange('');

    expect(model.state.filters).toMatchObject({ kind: '', status: '' });
  });
});

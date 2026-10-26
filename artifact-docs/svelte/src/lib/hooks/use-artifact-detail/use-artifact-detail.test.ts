import { render, waitFor } from '@testing-library/svelte';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { ApiError } from '$lib/api/http-client';
import { createMockDatabase } from '$lib/api/mock-fixtures';
import { type ArtifactDetail } from '$lib/contracts/artifact.schema';
import Harness from './use-artifact-detail-harness.test.svelte';
import { toArtifactView, type ArtifactDetailModel } from './use-artifact-detail.svelte';

vi.mock('$lib/api/artifacts.api', () => ({ artifactsApi: { get: vi.fn() } }));

const { artifactsApi } = await import('$lib/api/artifacts.api');

function detail(label: string): ArtifactDetail {
  const db = createMockDatabase(Date.parse('2026-10-25T12:00:00.000Z'));
  const artifact = db.artifacts.find((item) => item.label === label)!;

  return { ...artifact, versions: db.versions[artifact.id] ?? [] };
}

function mountModel(label: string): ArtifactDetailModel {
  let model!: ArtifactDetailModel;
  render(Harness, { props: { label, onReady: (ready) => (model = ready) } });

  return model;
}

beforeEach(() => {
  vi.mocked(artifactsApi.get).mockReset();
});

describe('toArtifactView', () => {
  // feliz
  it('lists the versions newest first and marks the current one', () => {
    const view = toArtifactView(detail('briefing-artifact-docs'));

    expect(view.versions.map((version) => version.version)).toEqual(['v4', 'v3', 'v2', 'v1']);
    expect(view.versions[0]?.isCurrent).toBe(true);
    expect(view.facts).toContainEqual({ label: 'Formato', value: 'Markdown' });
  });

  // triste
  it('says a private artifact needs login and has no status badge when it is a document', () => {
    const view = toArtifactView(detail('manual-da-marca'));

    expect(view.facts).toContainEqual({ label: 'Acesso', value: 'Exige login' });
    expect(view.badges.map((badge) => badge.label)).toEqual(['Documento']);
  });
});

describe('useArtifactDetailModel', () => {
  // feliz
  it('loads the artifact by its label', async () => {
    vi.mocked(artifactsApi.get).mockResolvedValue(detail('abertura-cnpj'));
    const model = mountModel('abertura-cnpj');

    await waitFor(() => expect(model.data.artifact?.title).toBe('Abertura do CNPJ'));
    expect(artifactsApi.get).toHaveBeenCalledWith('abertura-cnpj');
  });

  // triste
  it('shows the reason when the artifact does not exist', async () => {
    vi.mocked(artifactsApi.get).mockRejectedValue(new ApiError(404, 'Artefato não encontrado.'));
    const model = mountModel('nao-existe');

    await waitFor(() => expect(model.state.error).toBe('Artefato não encontrado.'));
    expect(model.data.artifact).toBeNull();
  });
});

import { render, waitFor } from '@testing-library/svelte';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { ApiError } from '$lib/api/http-client';
import { createMockDatabase } from '$lib/api/mock-fixtures';
import { type ArtifactDetail } from '$lib/contracts/artifact.schema';
import Harness from './use-artifact-detail-harness.test.svelte';
import { toArtifactView, type ArtifactDetailModel } from './use-artifact-detail.svelte';

vi.mock('$app/navigation', () => ({ goto: vi.fn() }));
vi.mock('$lib/api/artifacts.api', () => ({
  artifactsApi: { get: vi.fn(), update: vi.fn(), restore: vi.fn(), remove: vi.fn() },
}));

const { goto } = await import('$app/navigation');

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
  vi.mocked(artifactsApi.update).mockReset();
  vi.mocked(artifactsApi.restore).mockReset();
  vi.mocked(artifactsApi.remove).mockReset();
  vi.mocked(goto).mockReset();
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
  it('says a private artifact is private and has no status badge when it is a document', () => {
    const view = toArtifactView(detail('manual-da-marca'));

    expect(view.facts).toContainEqual({ label: 'Acesso', value: 'Privado' });
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

describe('useArtifactDetailModel — actions', () => {
  async function loaded(label: string): Promise<ArtifactDetailModel> {
    vi.mocked(artifactsApi.get).mockResolvedValue(detail(label));
    const model = mountModel(label);
    await waitFor(() => expect(model.data.artifact).not.toBeNull());

    return model;
  }

  // feliz
  it('turns a public artifact private', async () => {
    const model = await loaded('abertura-cnpj');
    const current = detail('abertura-cnpj');
    vi.mocked(artifactsApi.update).mockResolvedValue({ ...current, isPrivate: !current.isPrivate });

    model.actions.onTogglePrivate();

    await waitFor(() =>
      expect(artifactsApi.update).toHaveBeenCalledWith('abertura-cnpj', {
        isPrivate: !current.isPrivate,
      }),
    );
  });

  // feliz
  it('restores an older version', async () => {
    const model = await loaded('briefing-artifact-docs');
    vi.mocked(artifactsApi.restore).mockResolvedValue(detail('briefing-artifact-docs'));

    model.actions.onRestore(2);

    await waitFor(() =>
      expect(artifactsApi.restore).toHaveBeenCalledWith('briefing-artifact-docs', 2),
    );
  });

  // feliz
  it('deletes only after the confirmation, then goes back to the gallery', async () => {
    const model = await loaded('abertura-cnpj');
    vi.mocked(artifactsApi.remove).mockResolvedValue(undefined);

    model.actions.onDeleteRequest();
    expect(model.state.isDeleteOpen).toBe(true);
    expect(artifactsApi.remove).not.toHaveBeenCalled();

    model.actions.onDeleteConfirm();

    await waitFor(() => expect(goto).toHaveBeenCalledWith('/artifacts'));
    expect(model.state.isDeleteOpen).toBe(false);
  });

  // triste
  it('keeps the confirmation open with the reason when deleting fails', async () => {
    const model = await loaded('abertura-cnpj');
    vi.mocked(artifactsApi.remove).mockRejectedValue(
      new ApiError(500, 'Não consegui apagar o artefato.'),
    );

    model.actions.onDeleteRequest();
    model.actions.onDeleteConfirm();

    await waitFor(() => expect(model.state.deleteError).toBe('Não consegui apagar o artefato.'));
    expect(model.state.isDeleteOpen).toBe(true);
    expect(goto).not.toHaveBeenCalled();
  });

  // triste
  it('shows the reason when restoring fails', async () => {
    const model = await loaded('briefing-artifact-docs');
    vi.mocked(artifactsApi.restore).mockRejectedValue(new ApiError(404, 'Versão não encontrada.'));

    model.actions.onRestore(9);

    await waitFor(() => expect(model.state.actionError).toBe('Versão não encontrada.'));
  });
});

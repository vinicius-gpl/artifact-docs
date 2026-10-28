import { QueryClient } from '@tanstack/svelte-query';
import { render, waitFor } from '@testing-library/svelte';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { ApiError } from '$lib/api/http-client';
import { createMockDatabase } from '$lib/api/mock-fixtures';
import { type ArtifactDetail } from '$lib/contracts/artifact.schema';
import Harness from './use-artifact-edit-harness.test.svelte';
import {
  editErrorsOf,
  editValuesOf,
  toPatch,
  type ArtifactEditModel,
} from './use-artifact-edit.svelte';

vi.mock('$lib/api/artifacts.api', () => ({ artifactsApi: { update: vi.fn() } }));

const { artifactsApi } = await import('$lib/api/artifacts.api');

const database = createMockDatabase();
const plan = database.artifacts.find((artifact) => artifact.label === 'abertura-cnpj')!;
const PLAN: ArtifactDetail = { ...plan, versions: database.versions[plan.id] ?? [] };

function mountModel(detail: ArtifactDetail | null = PLAN) {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  if (detail) queryClient.setQueryData(['artifacts', 'detail', PLAN.label], detail);

  let model!: ArtifactEditModel;
  render(Harness, {
    props: { label: PLAN.label, queryClient, onReady: (ready) => (model = ready) },
  });

  return { model, queryClient };
}

beforeEach(() => {
  vi.mocked(artifactsApi.update)
    .mockReset()
    .mockImplementation((_label, patch) => Promise.resolve({ ...plan, ...patch } as never));
});

describe('editValuesOf and toPatch', () => {
  // feliz
  it('starts from what is published and sends it back as a full patch', () => {
    const values = editValuesOf(PLAN);

    expect(values).toMatchObject({ title: PLAN.title, status: 'active', visibility: 'private' });
    expect(toPatch({ ...values, tags: ' cnpj, ,cnpj, contador ' }, 'plan')).toMatchObject({
      tags: ['cnpj', 'contador'],
      status: 'active',
      isPrivate: true,
    });
  });

  it('clears the description when it is empty and sends no status for a document', () => {
    const patch = toPatch({ ...editValuesOf(PLAN), description: '  ' }, 'document');

    expect(patch.description).toBeNull();
    expect(patch).not.toHaveProperty('status');
  });

  // triste
  it('refuses an empty title, like publishing does', () => {
    expect(editErrorsOf({ ...editValuesOf(PLAN), title: '  ' }, 'plan')).toEqual({
      title: 'Informe o título',
    });
  });
});

describe('useArtifactEditModel', () => {
  // feliz
  it('opens with the artifact, saves the edit and updates the detail at once', async () => {
    const { model, queryClient } = mountModel();

    model.actions.onOpen();
    expect(model.state.isOpen).toBe(true);
    expect(model.ui.hasStatus).toBe(true);
    model.actions.onChange('title', 'Empresa aberta');
    model.actions.onSubmit();

    await waitFor(() => expect(model.state.isOpen).toBe(false));
    expect(artifactsApi.update).toHaveBeenCalledWith(
      'abertura-cnpj',
      expect.objectContaining({ title: 'Empresa aberta' }),
    );
    const detail = queryClient.getQueryData<ArtifactDetail>(['artifacts', 'detail', PLAN.label]);
    expect(detail?.title).toBe('Empresa aberta');
  });

  // triste
  it('does not open before the artifact has loaded', () => {
    const { model } = mountModel(null);

    model.actions.onOpen();

    expect(model.state.isOpen).toBe(false);
  });

  it('keeps the dialog open with the reason when the server refuses', async () => {
    vi.mocked(artifactsApi.update).mockRejectedValue(new ApiError(400, 'Situação inválida.'));
    const { model } = mountModel();

    model.actions.onOpen();
    model.actions.onSubmit();

    await waitFor(() => expect(model.state.error).toBe('Situação inválida.'));
    expect(model.state.isOpen).toBe(true);
  });

  it('shows the title error only after trying to save, and sends nothing', () => {
    const { model } = mountModel();
    model.actions.onOpen();
    model.actions.onChange('title', '');

    expect(model.state.fieldErrors).toEqual({});
    model.actions.onSubmit();

    expect(model.state.fieldErrors.title).toBe('Informe o título');
    expect(artifactsApi.update).not.toHaveBeenCalled();
  });
});

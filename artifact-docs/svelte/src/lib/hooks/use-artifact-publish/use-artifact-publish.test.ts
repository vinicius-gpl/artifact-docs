import { render, waitFor } from '@testing-library/svelte';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { ApiError } from '$lib/api/http-client';
import { createMockDatabase } from '$lib/api/mock-fixtures';
import { type PublishFormValues } from '$lib/components/acerola-publish-dialog/acerola-publish-dialog.svelte';
import { type ArtifactKind } from '$lib/contracts/artifact.schema';
import Harness from './use-artifact-publish-harness.test.svelte';
import { publishErrorsOf, type ArtifactPublishModel } from './use-artifact-publish.svelte';

vi.mock('$app/navigation', () => ({ goto: vi.fn() }));
vi.mock('$lib/api/artifacts.api', () => ({ artifactsApi: { publish: vi.fn() } }));

const { goto } = await import('$app/navigation');
const { artifactsApi } = await import('$lib/api/artifacts.api');

const published = createMockDatabase().artifacts[0]!;
const html = new File(['<h1>Guia</h1>'], 'guia.html', { type: 'text/html' });

function mountModel(input: { kind?: ArtifactKind; label?: string } = {}): ArtifactPublishModel {
  let model!: ArtifactPublishModel;
  render(Harness, { props: { ...input, onReady: (ready) => (model = ready) } });

  return model;
}

beforeEach(() => {
  vi.mocked(goto).mockReset();
  vi.mocked(artifactsApi.publish)
    .mockReset()
    .mockResolvedValue({ ...published, label: 'guia-rapido' });
});

describe('publishErrorsOf', () => {
  const filled: PublishFormValues = {
    file: html,
    title: 'Guia',
    label: 'guia',
    description: '',
    tags: '',
    status: '',
    visibility: 'private',
  };

  // feliz
  it('has no error for a filled new document', () => {
    expect(publishErrorsOf(filled, 'document')).toEqual({});
  });

  // triste
  it('asks for the file and the title of a new artifact, but only the file of a version', () => {
    const empty = { ...filled, file: null, title: '' };

    expect(publishErrorsOf(empty, 'plan')).toMatchObject({
      file: 'Escolha o arquivo a publicar',
      title: 'Informe o título',
    });
    expect(publishErrorsOf(empty, undefined)).toEqual({ file: 'Escolha o arquivo a publicar' });
  });
});

describe('useArtifactPublishModel', () => {
  // feliz
  it('names what the tab publishes and suggests the label from the title', () => {
    const model = mountModel({ kind: 'idea' });

    model.actions.onChange('title', 'Guia Rápido');
    expect(model.data.copy.submitLabel).toBe('Publicar ideia');
    expect(model.data.values.label).toBe('guia-rapido');

    model.actions.onChange('label', 'meu-guia');
    model.actions.onChange('title', 'Outro título');
    expect(model.data.values.label).toBe('meu-guia');
  });

  // feliz
  it('publishes a new artifact of the tab kind and opens it', async () => {
    const model = mountModel({ kind: 'plan' });
    model.actions.onOpen();
    model.actions.onFileChange(html);
    model.actions.onChange('title', 'Guia rápido');

    model.actions.onSubmit();

    await waitFor(() => expect(goto).toHaveBeenCalledWith('/artifacts/guia-rapido'));
    expect(vi.mocked(artifactsApi.publish).mock.calls[0]?.[0]).toMatchObject({
      file: html,
      label: 'guia-rapido',
      kind: 'plan',
      status: 'draft',
      isPrivate: true,
    });
    expect(model.state.isOpen).toBe(false);
  });

  // feliz
  it('publishes an idea open to the link with the chosen status, and a document without one', async () => {
    const idea = mountModel({ kind: 'idea' });
    idea.actions.onOpen();
    idea.actions.onFileChange(html);
    idea.actions.onChange('title', 'Ideia');
    idea.actions.onChange('status', 'active');
    idea.actions.onChange('visibility', 'public');
    idea.actions.onSubmit();

    await waitFor(() => expect(artifactsApi.publish).toHaveBeenCalled());
    expect(vi.mocked(artifactsApi.publish).mock.calls[0]?.[0]).toMatchObject({
      status: 'active',
      isPrivate: false,
    });

    const document = mountModel({ kind: 'document' });
    expect(document.ui.hasStatus).toBe(false);
    expect(document.data.values.status).toBe('');
  });

  // feliz
  it('publishes a new version without a kind, staying on the artifact', async () => {
    const model = mountModel({ label: 'abertura-cnpj' });
    model.actions.onOpen();
    model.actions.onFileChange(html);

    model.actions.onSubmit();

    await waitFor(() => expect(artifactsApi.publish).toHaveBeenCalled());
    expect(model.ui.mode).toBe('version');
    const sent = vi.mocked(artifactsApi.publish).mock.calls[0]?.[0];
    expect(sent).toMatchObject({ label: 'abertura-cnpj', kind: undefined });
    expect(sent).not.toHaveProperty('status');
    expect(sent).not.toHaveProperty('isPrivate');
    expect(goto).not.toHaveBeenCalled();
  });

  // triste
  it('shows no error before the person leaves a field, and all of them after trying to send', () => {
    const model = mountModel();
    model.actions.onOpen();
    expect(model.state.fieldErrors).toEqual({});

    model.actions.onSubmit();

    expect(model.state.fieldErrors.file).toBe('Escolha o arquivo a publicar');
    expect(artifactsApi.publish).not.toHaveBeenCalled();
  });

  // triste
  it('keeps the dialog open with the server reason when publishing fails', async () => {
    vi.mocked(artifactsApi.publish).mockRejectedValue(
      new ApiError(409, 'Outra publicação desta label terminou antes.'),
    );
    const model = mountModel();
    model.actions.onOpen();
    model.actions.onFileChange(html);
    model.actions.onChange('title', 'Guia');

    model.actions.onSubmit();

    await waitFor(() =>
      expect(model.state.error).toBe('Outra publicação desta label terminou antes.'),
    );
    expect(model.state.isOpen).toBe(true);
    expect(model.data.values.title).toBe('Guia');
  });
});

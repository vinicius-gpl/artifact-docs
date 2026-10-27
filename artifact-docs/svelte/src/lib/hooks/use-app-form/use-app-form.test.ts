import { render, waitFor } from '@testing-library/svelte';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { ApiError } from '$lib/api/http-client';
import { createMockDatabase } from '$lib/api/mock-fixtures';
import Harness from './use-app-form-harness.test.svelte';
import {
  appFormErrorsOf,
  typeSuggestionsOf,
  valuesOf,
  type AppFormModel,
} from './use-app-form.svelte';

vi.mock('$lib/api/apps.api', () => ({
  appsApi: { hub: vi.fn(), create: vi.fn(), update: vi.fn(), remove: vi.fn() },
}));

const { appsApi } = await import('$lib/api/apps.api');

const hub = createMockDatabase().appHub;
const firstApp = hub.apps[0]!;

async function mountModel(): Promise<AppFormModel> {
  let model!: AppFormModel;
  render(Harness, { props: { onReady: (ready) => (model = ready) } });
  await waitFor(() => expect(model.data.typeSuggestions.length).toBeGreaterThan(0));

  return model;
}

beforeEach(() => {
  vi.mocked(appsApi.hub).mockReset().mockResolvedValue(hub);
  vi.mocked(appsApi.create).mockReset().mockResolvedValue(firstApp);
  vi.mocked(appsApi.update).mockReset().mockResolvedValue(firstApp);
  vi.mocked(appsApi.remove).mockReset().mockResolvedValue(undefined);
});

describe('valuesOf and typeSuggestionsOf', () => {
  // feliz
  it('fills the edit form with the type name and suggests the types in hub order', () => {
    const values = valuesOf(hub, firstApp.id);
    const type = hub.types.find((item) => item.id === firstApp.typeId);

    expect(values).toEqual({ name: firstApp.name, url: firstApp.url, typeName: type?.name });
    expect(typeSuggestionsOf(hub)[0]).toBe(
      [...hub.types].sort((a, b) => a.position - b.position)[0]?.name,
    );
  });

  // triste
  it('has nothing to edit for an app that is gone, and no suggestion before loading', () => {
    expect(valuesOf(hub, 'sumiu')).toBeNull();
    expect(typeSuggestionsOf(undefined)).toEqual([]);
    expect(appFormErrorsOf({ name: '', url: 'x', typeName: '' })).toMatchObject({
      name: 'Informe o nome do app',
      typeName: 'Informe o tipo do app',
    });
  });
});

describe('useAppFormModel', () => {
  // feliz
  it('creates an app with a new type and closes', async () => {
    const model = await mountModel();
    model.actions.onCreate();
    model.actions.onChange('name', 'Immich');
    model.actions.onChange('url', 'https://fotos.exemplo.com');
    model.actions.onChange('typeName', 'Mídia');

    model.actions.onSubmit();

    await waitFor(() => expect(model.state.isOpen).toBe(false));
    expect(appsApi.create).toHaveBeenCalledWith({
      name: 'Immich',
      url: 'https://fotos.exemplo.com',
      typeName: 'Mídia',
    });
  });

  // feliz
  it('edits an existing app and deletes it only after the confirmation', async () => {
    const model = await mountModel();
    model.actions.onEdit(firstApp.id);
    expect(model.ui.mode).toBe('edit');

    model.actions.onSubmit();
    await waitFor(() =>
      expect(appsApi.update).toHaveBeenCalledWith(firstApp.id, expect.any(Object)),
    );

    model.actions.onEdit(firstApp.id);
    model.actions.onDeleteRequest();
    expect(appsApi.remove).not.toHaveBeenCalled();
    model.actions.onDeleteConfirm();

    await waitFor(() => expect(model.state.isOpen).toBe(false));
    expect(appsApi.remove).toHaveBeenCalledWith(firstApp.id);
  });

  // triste
  it('shows the errors only after trying to send, and keeps the dialog open on refusal', async () => {
    vi.mocked(appsApi.create).mockRejectedValue(
      new ApiError(400, 'Informe o tipo do app (até 40 caracteres).'),
    );
    const model = await mountModel();
    model.actions.onCreate();
    expect(model.state.fieldErrors).toEqual({});

    model.actions.onSubmit();
    expect(model.state.fieldErrors.name).toBe('Informe o nome do app');
    expect(appsApi.create).not.toHaveBeenCalled();

    model.actions.onChange('name', 'Immich');
    model.actions.onChange('url', 'https://fotos.exemplo.com');
    model.actions.onChange('typeName', 'Mídia');
    model.actions.onSubmit();

    await waitFor(() =>
      expect(model.state.error).toBe('Informe o tipo do app (até 40 caracteres).'),
    );
    expect(model.state.isOpen).toBe(true);
  });
});

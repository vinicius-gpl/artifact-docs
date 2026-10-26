import { render, waitFor } from '@testing-library/svelte';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { ApiError } from '$lib/api/http-client';
import { createMockDatabase } from '$lib/api/mock-fixtures';
import Harness from './use-brand-palettes-harness.test.svelte';
import { fileNameOf, sortPalettes, type BrandPalettesModel } from './use-brand-palettes.svelte';

vi.mock('$lib/api/palettes.api', () => ({ palettesApi: { list: vi.fn() } }));

const { palettesApi } = await import('$lib/api/palettes.api');

const palettes = () => createMockDatabase().palettes;

function mountModel(): BrandPalettesModel {
  let model!: BrandPalettesModel;
  render(Harness, { props: { onReady: (ready) => (model = ready) } });

  return model;
}

beforeEach(() => {
  vi.mocked(palettesApi.list).mockReset().mockResolvedValue({ items: palettes() });
});

describe('fileNameOf', () => {
  // feliz
  it('turns the palette name into a file name', () => {
    expect(fileNameOf('Grupo Azuos · rascunho')).toBe('grupo-azuos-rascunho.css');
  });

  // triste
  it('falls back to a generic name when nothing is left (edge case)', () => {
    expect(fileNameOf('···')).toBe('paleta.css');
  });
});

describe('sortPalettes', () => {
  // feliz
  it('puts the brand palette first', () => {
    const reversed = [...palettes()].reverse();

    expect(sortPalettes(reversed)[0]?.isBrand).toBe(true);
  });

  // triste
  it('keeps the list as it is when there is no brand palette', () => {
    const none = palettes().map((palette) => ({ ...palette, isBrand: false }));

    expect(sortPalettes(none).map((palette) => palette.id)).toEqual(
      none.map((palette) => palette.id),
    );
  });
});

describe('useBrandPalettesModel', () => {
  // feliz
  it('loads the palettes with a download for each one', async () => {
    const model = mountModel();

    await waitFor(() => expect(model.state.isLoading).toBe(false));
    expect(model.data.palettes[0]).toMatchObject({
      isBrand: true,
      downloadName: 'artifact-docs-carbon.css',
    });
    expect(model.data.palettes[0]?.downloadUrl).toMatch(/^data:text\/css/);
  });

  // triste
  it('shows the reason when the palettes do not load', async () => {
    vi.mocked(palettesApi.list).mockRejectedValue(new ApiError(500, 'O R2 não respondeu.'));
    const model = mountModel();

    await waitFor(() => expect(model.state.error).toBe('O R2 não respondeu.'));
  });
});

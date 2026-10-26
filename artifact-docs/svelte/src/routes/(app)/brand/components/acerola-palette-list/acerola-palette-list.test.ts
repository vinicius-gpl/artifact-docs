import { fireEvent, render, screen } from '@testing-library/svelte';
import { describe, expect, it, vi } from 'vitest';

import { type PaletteView } from '$lib/hooks/use-brand-palettes/use-brand-palettes.svelte';
import PaletteList from './acerola-palette-list.svelte';

const PALETTE: PaletteView = {
  id: 'p1',
  name: 'Artifact-docs · Carbon',
  isBrand: true,
  colors: [],
  downloadUrl: 'data:text/css,',
  downloadName: 'artifact-docs-carbon.css',
  createdLabel: 'Gerada em 25/10/2026',
};

const idle = { isLoading: false, error: null };

describe('AcerolaPaletteList', () => {
  // feliz
  it('lists every palette', () => {
    render(PaletteList, {
      props: { data: { palettes: [PALETTE] }, state: idle, actions: { onRetry: vi.fn() } },
    });

    expect(screen.getByRole('heading', { name: /Artifact-docs/ })).toBeInTheDocument();
  });

  // triste
  it('shows the reason and retries when the list fails', async () => {
    const onRetry = vi.fn();
    render(PaletteList, {
      props: {
        data: { palettes: [] },
        state: { ...idle, error: 'O R2 não respondeu.' },
        actions: { onRetry },
      },
    });

    await fireEvent.click(screen.getByRole('button', { name: /Tentar de novo/ }));
    expect(onRetry).toHaveBeenCalled();
  });

  it('tells how to start when there is no palette', () => {
    render(PaletteList, {
      props: { data: { palettes: [] }, state: idle, actions: { onRetry: vi.fn() } },
    });

    expect(screen.getByText('Nenhuma paleta ainda')).toBeInTheDocument();
  });
});

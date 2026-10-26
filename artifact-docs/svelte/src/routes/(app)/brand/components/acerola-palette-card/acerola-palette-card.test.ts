import { render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';

import { type PaletteView } from '$lib/hooks/use-brand-palettes/use-brand-palettes.svelte';
import PaletteCard from './acerola-palette-card.svelte';

const PALETTE: PaletteView = {
  id: 'p1',
  name: 'Artifact-docs · Carbon',
  isBrand: true,
  colors: [{ token: 'primary', light: '#ffa58a', dark: '#ffb59e' }],
  downloadUrl: 'data:text/css,:root{}',
  downloadName: 'artifact-docs-carbon.css',
  createdLabel: 'Gerada em 25/10/2026',
};

describe('AcerolaPaletteCard', () => {
  // feliz
  it('offers the generated CSS as a download with the palette name', () => {
    render(PaletteCard, { props: { data: { palette: PALETTE } } });

    const link = screen.getByRole('link', { name: 'Baixar CSS' });
    expect(link).toHaveAttribute('download', 'artifact-docs-carbon.css');
    expect(link).toHaveAttribute('href', 'data:text/css,:root{}');
    expect(screen.getByText('Marca oficial')).toBeInTheDocument();
    expect(screen.getByText('primary')).toBeInTheDocument();
  });

  // triste
  it('does not call a draft palette the official one', () => {
    render(PaletteCard, { props: { data: { palette: { ...PALETTE, isBrand: false } } } });

    expect(screen.queryByText('Marca oficial')).toBeNull();
  });

  it('draws a palette without colors (edge case)', () => {
    render(PaletteCard, { props: { data: { palette: { ...PALETTE, colors: [] } } } });

    expect(screen.getByRole('heading', { name: /Artifact-docs/ })).toBeInTheDocument();
  });
});

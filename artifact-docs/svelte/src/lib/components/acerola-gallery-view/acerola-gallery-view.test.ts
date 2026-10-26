import { fireEvent, render, screen } from '@testing-library/svelte';
import { describe, expect, it, vi } from 'vitest';

import { type PreviewItem } from '$lib/components/acerola-preview-card/acerola-preview-card.svelte';
import GalleryView, { type AcerolaGalleryViewProps } from './acerola-gallery-view.svelte';

const ITEM: PreviewItem = {
  id: 'a1',
  title: 'Briefing do hub',
  description: null,
  href: '/artifacts/briefing',
  previewUrl: 'data:text/html,<h1>x</h1>',
  thumbUrl: null,
  meta: 'Editado há 1 dia',
  badges: [],
  isLocked: false,
  isPinned: false,
};

function props(overrides: {
  data?: Partial<AcerolaGalleryViewProps['data']>;
  state?: Partial<AcerolaGalleryViewProps['state']>;
} = {}): AcerolaGalleryViewProps {
  return {
    data: {
      items: [ITEM],
      total: 1,
      pinned: [],
      recent: [],
      caption: 'Artefatos',
      emptyTitle: 'Nenhum artefato ainda',
      emptyDescription: 'Publique o primeiro.',
      ...overrides.data,
    },
    state: { isLoading: false, error: null, hasFilters: false, ...overrides.state },
    actions: { onRetry: vi.fn(), onClearFilters: vi.fn(), onTogglePin: vi.fn() },
  };
}

describe('AcerolaGalleryView', () => {
  // feliz
  it('shows the gallery with the pinned and recent shortcuts beside it', () => {
    render(GalleryView, { props: props() });

    expect(screen.getAllByRole('link', { name: 'Briefing do hub' }).length).toBeGreaterThan(0);
    expect(screen.getByRole('heading', { name: 'Fixados' })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: 'Recentes' })).toBeInTheDocument();
  });

  it('draws a skeleton while loading, never "nothing here"', () => {
    render(GalleryView, { props: props({ state: { isLoading: true } }) });

    expect(screen.getByLabelText('Carregando')).toHaveAttribute('aria-busy', 'true');
    expect(screen.queryByText('Nenhum artefato ainda')).toBeNull();
  });

  // triste
  it('shows the reason and retries when the list fails', async () => {
    const view = props({ state: { error: 'O banco não respondeu.' } });
    render(GalleryView, { props: view });

    expect(screen.getByText('O banco não respondeu.')).toBeInTheDocument();
    await fireEvent.click(screen.getByRole('button', { name: /Tentar de novo/ }));
    expect(view.actions.onRetry).toHaveBeenCalled();
  });

  it('tells what is empty when there is nothing at all', () => {
    render(GalleryView, { props: props({ data: { items: [], total: 0 } }) });

    expect(screen.getByText('Nenhum artefato ainda')).toBeInTheDocument();
  });

  it('offers to clear the filters when they hid everything', async () => {
    const view = props({ data: { items: [], total: 3 }, state: { hasFilters: true } });
    render(GalleryView, { props: view });

    await fireEvent.click(screen.getByRole('button', { name: 'Limpar filtros' }));
    expect(view.actions.onClearFilters).toHaveBeenCalled();
  });
});

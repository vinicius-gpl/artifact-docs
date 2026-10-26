import { fireEvent, render, within } from '@testing-library/svelte';
import { describe, expect, it, vi } from 'vitest';

import { type PreviewItem } from '$lib/components/acerola-preview-card/acerola-preview-card.svelte';
import PreviewGallery, { countLabel } from './acerola-preview-gallery.svelte';

function item(id: string, title: string, overrides: Partial<PreviewItem> = {}): PreviewItem {
  return {
    id,
    title,
    description: null,
    href: `/artifacts/${id}`,
    previewUrl: 'data:text/html,<h1>x</h1>',
    thumbUrl: null,
    meta: 'Editado há 1 dia',
    badges: [],
    isLocked: false,
    isPinned: false,
    ...overrides,
  };
}

const ITEMS = [item('a1', 'Briefing'), item('a2', 'Manual da marca', { isPinned: true })];

function slot(container: HTMLElement, name: string): HTMLElement {
  const element = container.querySelector<HTMLElement>(`[data-slot="${name}"]`);
  if (!element) throw new Error(`sem ${name}`);

  return element;
}

describe('AcerolaPreviewGallery', () => {
  // feliz
  it('draws the same items as cards and as table rows', () => {
    const { container } = render(PreviewGallery, {
      props: { data: { items: ITEMS, caption: 'Artefatos' } },
    });

    const cards = slot(container, 'preview-cards-mobile');
    const table = slot(container, 'preview-table-desktop');
    expect(within(cards).getAllByRole('article')).toHaveLength(2);
    expect(within(table).getByRole('link', { name: 'Manual da marca' })).toBeInTheDocument();
    expect(within(table).getByText('2 itens')).toBeInTheDocument();
  });

  it('pins from the table row with the item id', async () => {
    const onTogglePin = vi.fn();
    const { container } = render(PreviewGallery, {
      props: { data: { items: ITEMS, caption: 'Artefatos' }, actions: { onTogglePin } },
    });

    await fireEvent.click(
      within(slot(container, 'preview-table-desktop')).getByRole('button', {
        name: 'Fixar Briefing',
      }),
    );

    expect(onTogglePin).toHaveBeenCalledWith('a1');
  });

  it('hides the table when the person asked to always see cards', () => {
    const { container } = render(PreviewGallery, {
      props: { data: { items: ITEMS, caption: 'Artefatos' }, state: { forceCards: true } },
    });

    expect(slot(container, 'preview-table-desktop').className).toContain('hidden');
    expect(slot(container, 'preview-cards-mobile').className).not.toContain('xl:hidden');
  });

  // triste
  it('locks only the pin of the item being saved', () => {
    const { container } = render(PreviewGallery, {
      props: { data: { items: ITEMS, caption: 'Artefatos' }, state: { pendingPinId: 'a2' } },
    });
    const table = within(slot(container, 'preview-table-desktop'));

    expect(table.getByRole('button', { name: 'Desafixar Manual da marca' })).toBeDisabled();
    expect(table.getByRole('button', { name: 'Fixar Briefing' })).toBeEnabled();
  });

  it('counts a single item in the singular (edge case)', () => {
    expect(countLabel(1)).toBe('1 item');
    expect(countLabel(0)).toBe('0 itens');
  });
});

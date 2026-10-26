import { fireEvent, render, screen } from '@testing-library/svelte';
import { describe, expect, it, vi } from 'vitest';

import PreviewCard, { type PreviewItem } from './acerola-preview-card.svelte';

function item(overrides: Partial<PreviewItem> = {}): PreviewItem {
  return {
    id: 'a1',
    title: 'Briefing do hub',
    description: 'O que é o hub e os módulos.',
    href: '/artifacts/briefing',
    previewUrl: 'data:text/html,<h1>Briefing</h1>',
    thumbUrl: null,
    meta: 'Editado há 2 dias',
    badges: [{ label: 'Markdown', tone: 'neutral' }],
    isLocked: false,
    isPinned: false,
    ...overrides,
  };
}

describe('AcerolaPreviewCard', () => {
  // feliz
  it('links the whole card to the item through its title', () => {
    render(PreviewCard, { props: { data: { item: item() } } });

    expect(screen.getByRole('link', { name: 'Briefing do hub' })).toHaveAttribute(
      'href',
      '/artifacts/briefing',
    );
    expect(screen.getByText('Editado há 2 dias')).toBeInTheDocument();
    expect(screen.getByText('Markdown')).toBeInTheDocument();
  });

  it('shows the page itself, sandboxed, when there is no thumbnail', () => {
    const { container } = render(PreviewCard, { props: { data: { item: item() } } });

    const frame = container.querySelector('iframe');
    expect(frame).toHaveAttribute('sandbox', 'allow-scripts');
    expect(frame).toHaveAttribute('tabindex', '-1');
  });

  it('uses the thumbnail instead of the page when there is one', () => {
    const { container } = render(PreviewCard, {
      props: { data: { item: item({ thumbUrl: '/thumbs/a1.png' }) } },
    });

    expect(container.querySelector('img')).toHaveAttribute('src', '/thumbs/a1.png');
    expect(container.querySelector('iframe')).toBeNull();
  });

  it('asks to pin the item by its id', async () => {
    const onTogglePin = vi.fn();
    render(PreviewCard, { props: { data: { item: item() }, actions: { onTogglePin } } });

    await fireEvent.click(screen.getByRole('button', { name: 'Fixar Briefing do hub' }));

    expect(onTogglePin).toHaveBeenCalledWith('a1');
  });

  it('marks a private item with a lock', () => {
    render(PreviewCard, { props: { data: { item: item({ isLocked: true }) } } });

    expect(screen.getByRole('img', { name: 'Privado' })).toBeInTheDocument();
  });

  // triste
  it('offers to unpin, pressed, when the item is already pinned', () => {
    render(PreviewCard, { props: { data: { item: item({ isPinned: true }) } } });

    expect(screen.getByRole('button', { name: 'Desafixar Briefing do hub' })).toHaveAttribute(
      'aria-pressed',
      'true',
    );
  });

  it('locks the pin button while the change is being saved', () => {
    render(PreviewCard, {
      props: { data: { item: item() }, state: { isPinPending: true } },
    });

    expect(screen.getByRole('button', { name: 'Fixar Briefing do hub' })).toBeDisabled();
  });

  it('draws without description nor badges (edge case)', () => {
    render(PreviewCard, {
      props: { data: { item: item({ description: null, badges: [] }) } },
    });

    expect(screen.getByRole('link', { name: 'Briefing do hub' })).toBeInTheDocument();
  });
});

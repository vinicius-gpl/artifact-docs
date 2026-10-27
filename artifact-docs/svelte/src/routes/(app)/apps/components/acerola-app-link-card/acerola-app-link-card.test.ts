import { fireEvent, render, screen } from '@testing-library/svelte';
import { describe, expect, it, vi } from 'vitest';

import { type AppTile } from '$lib/hooks/use-app-hub/use-app-hub.svelte';
import AppLinkCard, { initialsOf } from './acerola-app-link-card.svelte';

const APP: AppTile = {
  id: 'app-1',
  name: 'Home Assistant',
  url: 'https://casa.exemplo.com',
  host: 'casa.exemplo.com',
  iconUrl: 'https://casa.exemplo.com/favicon.ico',
  status: { label: 'No ar', tone: 'success' },
  checkedLabel: 'Verificado há 2 minutos',
};

describe('AcerolaAppLinkCard', () => {
  // feliz
  it('opens the app in another tab, safely', () => {
    render(AppLinkCard, { props: { data: { app: APP } } });

    const link = screen.getByRole('link', { name: /Home Assistant/ });
    expect(link).toHaveAttribute('href', 'https://casa.exemplo.com');
    expect(link).toHaveAttribute('target', '_blank');
    expect(link).toHaveAttribute('rel', 'noopener noreferrer');
    expect(screen.getByText('No ar')).toBeInTheDocument();
  });

  // triste
  it('falls back to the initials when the icon does not load', async () => {
    const { container } = render(AppLinkCard, { props: { data: { app: APP } } });

    await fireEvent.error(container.querySelector('img')!);

    expect(container.querySelector('img')).toBeNull();
    expect(screen.getByText('HA')).toBeInTheDocument();
  });

  it('builds initials from at most two words (edge case)', () => {
    expect(initialsOf('Uptime Kuma Status')).toBe('UK');
    expect(initialsOf('neon')).toBe('N');
  });

  // feliz
  it('edits outside the link, so the click does not open the app', async () => {
    const onEdit = vi.fn();
    render(AppLinkCard, { props: { data: { app: APP }, actions: { onEdit } } });

    const button = screen.getByRole('button', { name: 'Editar Home Assistant' });
    await fireEvent.click(button);

    expect(onEdit).toHaveBeenCalled();
    expect(button.closest('a')).toBeNull();
  });

  // triste
  it('has no edit button when the screen does not offer editing', () => {
    render(AppLinkCard, { props: { data: { app: APP } } });

    expect(screen.queryByRole('button', { name: /Editar/ })).not.toBeInTheDocument();
  });
});

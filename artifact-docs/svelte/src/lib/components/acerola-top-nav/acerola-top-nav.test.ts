import { fireEvent, render, screen, within } from '@testing-library/svelte';
import { describe, expect, it, vi } from 'vitest';

import { NAV_ITEMS } from '$lib/navigation/navigation';

import TopNav from './acerola-top-nav.svelte';

/* `null` (e não `undefined`) para "nada aceso": `undefined` num parâmetro com padrão faz o
   JavaScript usar o padrão, e o teste de "nada aceso" acabava com "Documentos" aceso. */
function renderNav(active: string | null = 'artifacts', onSignOut = vi.fn()) {
  const activeKey = active ?? undefined;
  return render(TopNav, {
    props: {
      data: { items: NAV_ITEMS, homeHref: '/apps' },
      state: { activeKey },
      actions: { onSignOut },
    },
  });
}

describe('AcerolaTopNav', () => {
  // feliz
  it('links every screen and marks the open one as the current page', () => {
    renderNav('plans');

    const nav = screen.getByRole('navigation', { name: 'Principal' });
    expect(within(nav).getAllByRole('link')).toHaveLength(NAV_ITEMS.length);
    expect(within(nav).getByRole('link', { name: 'Planejamentos' })).toHaveAttribute(
      'aria-current',
      'page',
    );
    expect(within(nav).getByRole('link', { name: 'Documentos' })).not.toHaveAttribute(
      'aria-current',
    );
  });

  // feliz
  it('draws one divider between Apps and the publishing tabs', () => {
    const { container } = renderNav();

    expect(container.querySelectorAll('[data-slot="top-nav-divider"]')).toHaveLength(1);
  });

  it('takes the brand back home and offers the theme switch', () => {
    renderNav();

    expect(screen.getByRole('link', { name: /artifact-docs/ })).toHaveAttribute('href', '/apps');
    expect(screen.getByRole('button', { name: /Mudar para o tema/ })).toBeInTheDocument();
  });

  it('signs out, with no avatar, profile screen nor effects button', async () => {
    const onSignOut = vi.fn();
    renderNav('apps', onSignOut);

    expect(screen.queryByRole('img', { name: /Você/ })).toBeNull();
    await fireEvent.click(screen.getByRole('button', { name: 'Sair' }));
    expect(onSignOut).toHaveBeenCalled();
    expect(screen.queryByRole('button', { name: /Efeitos|Perfil/ })).toBeNull();
  });

  // triste
  it('opens the phone menu and closes it when an item is chosen', async () => {
    const { container } = renderNav();

    await fireEvent.click(screen.getByRole('button', { name: 'Abrir menu' }));
    const menu = container.querySelector<HTMLElement>('[data-slot="top-nav-menu"]');
    expect(menu).not.toBeNull();

    await fireEvent.click(within(menu!).getByRole('link', { name: 'Ideias' }));
    expect(container.querySelector('[data-slot="top-nav-menu"]')).toBeNull();
  });

  it('lights nothing on an address outside the menu (edge case)', () => {
    const { container } = renderNav(null);

    expect(container.querySelector('[aria-current="page"]')).toBeNull();
  });
});

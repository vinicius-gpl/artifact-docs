import { render, waitFor } from '@testing-library/svelte';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { type SessionUser } from '$lib/contracts/user.schema';
import { NAV_ITEMS } from '$lib/navigation/navigation';
import Harness from './use-app-shell-harness.test.svelte';
import { type AppShellModel } from './use-app-shell.svelte';

vi.mock('$lib/auth/neon-auth.client', () => ({
  neonAuth: { signOut: vi.fn() },
  readAuthToken: vi.fn().mockResolvedValue(null),
}));

const { neonAuth } = await import('$lib/auth/neon-auth.client');
const { goto } = await import('$app/navigation');
const { page } = await import('$app/state');

/** A tela aberta, do ponto de vista do model: é dela que sai o item aceso. */
function openScreen(pathname: string) {
  page.url = new URL(`http://localhost:5173${pathname}`);
}

function user(overrides: Partial<SessionUser> = {}): SessionUser {
  return { id: '1', email: 'ana@exemplo.com', name: 'Ana', role: 'admin', ...overrides };
}

function mountModel(overrides: Partial<SessionUser> = {}): AppShellModel {
  let model!: AppShellModel;
  render(Harness, {
    props: { user: user(overrides), items: NAV_ITEMS, onReady: (ready) => (model = ready) },
  });

  return model;
}

beforeEach(() => {
  vi.mocked(neonAuth.signOut)
    .mockReset()
    .mockResolvedValue(undefined as never);
  vi.mocked(goto).mockClear();
  openScreen('/');
});

describe('useAppShellModel', () => {
  // feliz
  it('translates the session user into what the shell shows', () => {
    const model = mountModel({ name: 'Ana Souza', email: 'ana@exemplo.com', role: 'user' });

    expect(model.data.user).toEqual({
      name: 'Ana Souza',
      email: 'ana@exemplo.com',
      role: 'Usuário',
      avatarUrl: null,
    });
  });

  // feliz
  /* A foto da conta vem pelo `/api/auth/me`: sem repassá-la, o rodapé fica nas iniciais. */
  it('passes the account photo on to the shell', () => {
    const model = mountModel({ image: 'https://exemplo.com/fotos/ana.png' });

    expect(model.data.user.avatarUrl).toBe('https://exemplo.com/fotos/ana.png');
  });

  // feliz
  it('lights the menu item of the open screen, also on its subroutes', () => {
    openScreen('/artifacts/briefing');
    const model = mountModel();

    expect(model.ui.items).toBe(NAV_ITEMS);
    expect(model.state.activeKey).toBe('artifacts');
  });

  // triste
  it('lights nothing on a screen outside the menu', () => {
    openScreen('/somewhere-else');
    const model = mountModel();

    expect(model.state.activeKey).toBeUndefined();
  });

  // feliz
  it('opens and closes the profile', () => {
    const model = mountModel();

    model.actions.onOpenProfile();
    expect(model.state.isProfileOpen).toBe(true);

    model.actions.onCloseProfile();
    expect(model.state.isProfileOpen).toBe(false);
  });

  // feliz
  it('logs out at Neon Auth and navigates to /login', async () => {
    const model = mountModel();

    model.actions.onLogout();

    await waitFor(() => expect(neonAuth.signOut).toHaveBeenCalledOnce());
    await waitFor(() => expect(goto).toHaveBeenCalledWith('/login'));
  });

  // triste
  /* Sair não pode depender de rede: quem clicou em "Sair" precisa sair da tela de qualquer
     jeito, e a sessão que sobrar do lado da Neon vence sozinha. */
  it('still navigates to /login even when the sign out request fails', async () => {
    vi.mocked(neonAuth.signOut).mockRejectedValueOnce(new Error('network down'));
    const model = mountModel();

    model.actions.onLogout();

    await waitFor(() => expect(goto).toHaveBeenCalledWith('/login'));
  });
});

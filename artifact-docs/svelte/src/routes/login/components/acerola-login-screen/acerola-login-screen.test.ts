import { render, screen } from '@testing-library/svelte';
import { describe, expect, it, vi } from 'vitest';

import LoginScreen, { type AcerolaLoginScreenProps } from './acerola-login-screen.svelte';

function renderScreen(overrides: { isMockMode?: boolean } = {}) {
  const props: AcerolaLoginScreenProps = {
    data: {
      isMockMode: overrides.isMockMode ?? false,
      signIn: { values: { email: '', password: '' } },
    },
    state: { signIn: { isSubmitting: false, error: null, fieldErrors: {} } },
    actions: { signIn: { onChange: vi.fn(), onSubmit: vi.fn() } },
  };
  render(LoginScreen, { props });

  return props.actions;
}

describe('AcerolaLoginScreen', () => {
  // feliz
  it('opens on the sign-in form', () => {
    renderScreen();

    expect(screen.getByRole('heading', { name: 'Entrar' })).toBeInTheDocument();
    expect(screen.getByLabelText('E-mail')).toBeInTheDocument();
  });

  it('leads to the first access, which is not a sign-up form', () => {
    renderScreen();

    expect(
      screen.getByRole('link', { name: 'Primeiro acesso ou esqueci a senha' }),
    ).toHaveAttribute('href', '/first-access');
    expect(screen.queryByLabelText('Nome')).toBeNull();
  });

  it('says it is local mode in mock mode', () => {
    renderScreen({ isMockMode: true });

    expect(screen.getByText('Modo local')).toBeInTheDocument();
  });

  // triste
  it('hides the local mode badge outside mock mode', () => {
    renderScreen({ isMockMode: false });

    expect(screen.queryByText('Modo local')).toBeNull();
  });
});

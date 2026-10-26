import { fireEvent, render, screen } from '@testing-library/svelte';
import { describe, expect, it, vi } from 'vitest';

import LoginScreen, { type AcerolaLoginScreenProps } from './acerola-login-screen.svelte';

function renderScreen(overrides: { isMockMode?: boolean; needsSetup?: boolean } = {}) {
  const idle = { isSubmitting: false, error: null, fieldErrors: {} };
  const props: AcerolaLoginScreenProps = {
    data: {
      isMockMode: overrides.isMockMode ?? false,
      needsSetup: overrides.needsSetup ?? false,
      signIn: { values: { email: '', password: '' } },
      firstAccess: { values: { name: '', email: '', password: '', passwordConfirmation: '' } },
    },
    state: { signIn: idle, firstAccess: idle },
    actions: {
      signIn: { onChange: vi.fn(), onSubmit: vi.fn() },
      firstAccess: { onChange: vi.fn(), onSubmit: vi.fn() },
    },
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

  it('switches to the first access while there is no account, and back', async () => {
    const actions = renderScreen({ needsSetup: true });

    await fireEvent.click(screen.getByRole('button', { name: 'Primeiro acesso · Criar a conta' }));
    expect(screen.getByRole('heading', { name: 'Primeiro acesso' })).toBeInTheDocument();
    await fireEvent.click(screen.getByRole('button', { name: 'Criar conta e entrar' }));
    expect(actions.firstAccess.onSubmit).toHaveBeenCalled();

    await fireEvent.click(screen.getByRole('button', { name: 'Já tenho conta · Entrar' }));
    expect(screen.getByRole('heading', { name: 'Entrar' })).toBeInTheDocument();
  });

  it('says it is local mode in mock mode', () => {
    renderScreen({ isMockMode: true });

    expect(screen.getByText('Modo local')).toBeInTheDocument();
  });

  // triste
  it('hides the first access once the account exists', () => {
    renderScreen({ needsSetup: false });

    expect(screen.queryByRole('button', { name: /Primeiro acesso/ })).toBeNull();
  });
});

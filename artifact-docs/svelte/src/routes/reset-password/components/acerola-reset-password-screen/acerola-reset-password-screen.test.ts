import { fireEvent, render, screen } from '@testing-library/svelte';
import { describe, expect, it, vi } from 'vitest';

import ResetPasswordScreen, {
  type AcerolaResetPasswordScreenProps,
} from './acerola-reset-password-screen.svelte';

function renderScreen(state: Partial<AcerolaResetPasswordScreenProps['state']> = {}) {
  const props: AcerolaResetPasswordScreenProps = {
    data: { values: { password: '', passwordConfirmation: '' } },
    state: { isSubmitting: false, isLinkValid: true, error: null, fieldErrors: {}, ...state },
    actions: { onChange: vi.fn(), onSubmit: vi.fn() },
  };
  render(ResetPasswordScreen, { props });

  return props.actions;
}

describe('AcerolaResetPasswordScreen', () => {
  // feliz
  it('asks for the password twice and saves it', async () => {
    const actions = renderScreen();

    expect(screen.getByRole('heading', { name: 'Crie sua senha' })).toBeInTheDocument();
    await fireEvent.click(screen.getByRole('button', { name: 'Salvar senha' }));

    expect(actions.onSubmit).toHaveBeenCalled();
  });

  // triste
  it('offers a new link instead of the form when the link is not valid', () => {
    renderScreen({ isLinkValid: false });

    expect(screen.getByRole('link', { name: 'Pedir outro link' })).toHaveAttribute(
      'href',
      '/first-access',
    );
    expect(screen.queryByRole('button', { name: 'Salvar senha' })).toBeNull();
  });

  it('shows the mismatch under the confirmation', () => {
    renderScreen({ fieldErrors: { passwordConfirmation: 'As senhas não são iguais' } });

    expect(screen.getByText('As senhas não são iguais')).toBeInTheDocument();
  });
});

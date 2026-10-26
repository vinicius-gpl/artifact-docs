import { fireEvent, render, screen } from '@testing-library/svelte';
import { describe, expect, it, vi } from 'vitest';

import SignInForm, { type AcerolaSignInFormProps } from './acerola-sign-in-form.svelte';

function renderForm(formState: Partial<AcerolaSignInFormProps['state']> = {}) {
  const props: AcerolaSignInFormProps = {
    data: { values: { email: '', password: '' } },
    state: { isSubmitting: false, error: null, fieldErrors: {}, ...formState },
    actions: { onChange: vi.fn(), onSubmit: vi.fn() },
  };
  render(SignInForm, { props });

  return props.actions;
}

describe('AcerolaSignInForm', () => {
  // feliz
  it('asks only for e-mail and password, and sends the form', async () => {
    const actions = renderForm();

    await fireEvent.input(screen.getByLabelText('E-mail'), {
      target: { value: 'voce@exemplo.com' },
    });
    await fireEvent.click(screen.getByRole('button', { name: 'Entrar' }));

    expect(actions.onChange).toHaveBeenCalledWith('email', 'voce@exemplo.com');
    expect(actions.onSubmit).toHaveBeenCalled();
    expect(screen.getByLabelText('Senha')).toHaveAttribute('autocomplete', 'current-password');
  });

  // triste
  it('shows each field error next to its field', () => {
    renderForm({ fieldErrors: { email: 'Informe o e-mail' } });

    expect(screen.getByText('Informe o e-mail')).toBeInTheDocument();
  });

  it('announces a refused sign-in', () => {
    renderForm({ error: 'E-mail ou senha incorretos.' });

    expect(screen.getByRole('alert')).toHaveTextContent('E-mail ou senha incorretos.');
  });

  it('locks the form while signing in', () => {
    renderForm({ isSubmitting: true });

    expect(screen.getByRole('button', { name: 'Entrando…' })).toBeDisabled();
    expect(screen.getByLabelText('E-mail')).toBeDisabled();
  });
});

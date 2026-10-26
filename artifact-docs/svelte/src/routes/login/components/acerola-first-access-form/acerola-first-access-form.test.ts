import { fireEvent, render, screen } from '@testing-library/svelte';
import { describe, expect, it, vi } from 'vitest';

import FirstAccessForm, {
  type AcerolaFirstAccessFormProps,
} from './acerola-first-access-form.svelte';

function renderForm(formState: Partial<AcerolaFirstAccessFormProps['state']> = {}) {
  const props: AcerolaFirstAccessFormProps = {
    data: { values: { name: '', email: '', password: '', passwordConfirmation: '' } },
    state: { isSubmitting: false, error: null, fieldErrors: {}, ...formState },
    actions: { onChange: vi.fn(), onSubmit: vi.fn() },
  };
  render(FirstAccessForm, { props });

  return props.actions;
}

describe('AcerolaFirstAccessForm', () => {
  // feliz
  it('asks for name, e-mail and a confirmed new password, and sends it', async () => {
    const actions = renderForm();

    await fireEvent.input(screen.getByLabelText('Nome'), { target: { value: 'Você' } });
    await fireEvent.click(screen.getByRole('button', { name: 'Criar conta e entrar' }));

    expect(actions.onChange).toHaveBeenCalledWith('name', 'Você');
    expect(actions.onSubmit).toHaveBeenCalled();
    expect(screen.getByLabelText('Repita a senha')).toHaveAttribute('autocomplete', 'new-password');
  });

  // triste
  it('shows the mismatch on the confirmation field', () => {
    renderForm({ fieldErrors: { passwordConfirmation: 'As senhas não são iguais' } });

    expect(screen.getByText('As senhas não são iguais')).toBeInTheDocument();
  });

  it('announces why the account was not created', () => {
    renderForm({ error: 'A conta do hub já existe.' });

    expect(screen.getByRole('alert')).toHaveTextContent('A conta do hub já existe.');
  });

  it('locks the form while creating the account', () => {
    renderForm({ isSubmitting: true });

    expect(screen.getByRole('button', { name: 'Criando a conta…' })).toBeDisabled();
  });
});

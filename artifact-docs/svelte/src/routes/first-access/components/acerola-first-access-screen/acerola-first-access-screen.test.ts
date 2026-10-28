import { fireEvent, render, screen } from '@testing-library/svelte';
import { describe, expect, it, vi } from 'vitest';

import FirstAccessScreen, {
  type AcerolaFirstAccessScreenProps,
} from './acerola-first-access-screen.svelte';

function renderScreen(state: Partial<AcerolaFirstAccessScreenProps['state']> = {}) {
  const props: AcerolaFirstAccessScreenProps = {
    data: { email: 'dono@exemplo.com' },
    state: { isSubmitting: false, isSent: false, error: null, ...state },
    actions: { onChange: vi.fn(), onSubmit: vi.fn() },
  };
  render(FirstAccessScreen, { props });

  return props.actions;
}

describe('AcerolaFirstAccessScreen', () => {
  // feliz
  it('asks only for the e-mail and sends the link', async () => {
    const actions = renderScreen();

    expect(screen.getByRole('heading', { name: 'Primeiro acesso' })).toBeInTheDocument();
    expect(screen.queryByLabelText('Nome')).toBeNull();
    await fireEvent.click(screen.getByRole('button', { name: 'Enviar link' }));

    expect(actions.onSubmit).toHaveBeenCalled();
  });

  it('swaps the form for the notice once the link is sent', () => {
    renderScreen({ isSent: true });

    expect(screen.getByRole('heading', { name: 'Confira seu e-mail' })).toBeInTheDocument();
    expect(screen.queryByLabelText('E-mail')).toBeNull();
  });

  // triste
  it('shows why the link was not sent', () => {
    renderScreen({ error: 'Muitas tentativas seguidas.' });

    expect(screen.getByText('Muitas tentativas seguidas.')).toBeInTheDocument();
  });
});

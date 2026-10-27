import { render, screen } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';

import AppDialog, { type AcerolaAppDialogProps } from './acerola-app-dialog.svelte';

function props(overrides: Partial<AcerolaAppDialogProps> = {}): AcerolaAppDialogProps {
  return {
    data: {
      values: { name: '', url: '', typeName: '' },
      typeSuggestions: ['Self-hosted', 'Na VPS'],
    },
    ui: { mode: 'new' },
    state: { isOpen: true, isSubmitting: false, error: null, fieldErrors: {} },
    actions: {
      onClose: vi.fn(),
      onChange: vi.fn(),
      onBlur: vi.fn(),
      onSubmit: vi.fn(),
      onDelete: vi.fn(),
    },
    ...overrides,
  };
}

describe('AcerolaAppDialog', () => {
  // feliz
  it('fills the type from a suggestion and submits', async () => {
    const current = props();
    render(AppDialog, { props: current });

    await userEvent.click(screen.getByRole('button', { name: 'Na VPS' }));
    await userEvent.click(screen.getByRole('button', { name: 'Cadastrar app' }));

    expect(current.actions.onChange).toHaveBeenCalledWith('typeName', 'Na VPS');
    expect(current.actions.onSubmit).toHaveBeenCalled();
    expect(screen.queryByRole('button', { name: 'Apagar app' })).not.toBeInTheDocument();
  });

  // feliz
  it('offers to delete when editing', async () => {
    const current = props({ ui: { mode: 'edit' } });
    render(AppDialog, { props: current });

    await userEvent.click(screen.getByRole('button', { name: 'Apagar app' }));

    expect(screen.getByRole('heading', { name: 'Editar app' })).toBeInTheDocument();
    expect(current.actions.onDelete).toHaveBeenCalled();
  });

  // triste
  it('shows the field errors and the server refusal', () => {
    render(AppDialog, {
      props: props({
        state: {
          isOpen: true,
          isSubmitting: false,
          error: 'Não consegui cadastrar o app.',
          fieldErrors: { url: 'Informe o endereço completo, começando com https://' },
        },
      }),
    });

    expect(
      screen.getByText('Informe o endereço completo, começando com https://'),
    ).toBeInTheDocument();
    expect(screen.getByText('Não consegui cadastrar o app.')).toBeInTheDocument();
  });

  // triste
  it('locks the fields while saving', () => {
    render(AppDialog, {
      props: props({ state: { isOpen: true, isSubmitting: true, error: null, fieldErrors: {} } }),
    });

    expect(screen.getByLabelText(/Nome/)).toBeDisabled();
    expect(screen.getByRole('button', { name: 'Cancelar' })).toBeDisabled();
  });
});

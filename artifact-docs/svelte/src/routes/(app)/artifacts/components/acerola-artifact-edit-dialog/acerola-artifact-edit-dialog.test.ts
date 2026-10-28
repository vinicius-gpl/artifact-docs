import { fireEvent, render, screen } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';

import ArtifactEditDialog, {
  type AcerolaArtifactEditDialogProps,
} from './acerola-artifact-edit-dialog.svelte';

function props(
  overrides: Partial<AcerolaArtifactEditDialogProps> = {},
): AcerolaArtifactEditDialogProps {
  return {
    data: {
      values: {
        title: 'Abertura do CNPJ',
        description: '',
        tags: 'cnpj',
        status: 'active',
        visibility: 'private',
      },
    },
    ui: { hasStatus: true },
    state: { isOpen: true, isSubmitting: false, error: null, fieldErrors: {} },
    actions: { onClose: vi.fn(), onChange: vi.fn(), onBlur: vi.fn(), onSubmit: vi.fn() },
    ...overrides,
  };
}

describe('AcerolaArtifactEditDialog', () => {
  // feliz
  it('comes filled, reports the edits and saves', async () => {
    const current = props();
    render(ArtifactEditDialog, { props: current });

    expect(screen.getByLabelText(/Título/)).toHaveValue('Abertura do CNPJ');
    await userEvent.type(screen.getByLabelText(/Tags/), 's');
    await fireEvent.click(screen.getByRole('button', { name: 'Com o link' }));
    await fireEvent.click(screen.getByRole('button', { name: 'Salvar' }));

    expect(current.actions.onChange).toHaveBeenCalledWith('tags', 'cnpjs');
    expect(current.actions.onChange).toHaveBeenCalledWith('visibility', 'public');
    expect(current.actions.onSubmit).toHaveBeenCalled();
  });

  it('never offers to change the label, which is the address', () => {
    render(ArtifactEditDialog, { props: props() });

    expect(screen.queryByLabelText(/Label/)).toBeNull();
  });

  // triste
  it('has no status for a document and shows the errors', () => {
    render(ArtifactEditDialog, {
      props: props({
        ui: { hasStatus: false },
        state: {
          isOpen: true,
          isSubmitting: false,
          error: 'Não consegui salvar a alteração.',
          fieldErrors: { title: 'Informe o título' },
        },
      }),
    });

    expect(screen.queryByText('Situação')).toBeNull();
    expect(screen.getByText('Informe o título')).toBeInTheDocument();
    expect(screen.getByText('Não consegui salvar a alteração.')).toBeInTheDocument();
  });
});

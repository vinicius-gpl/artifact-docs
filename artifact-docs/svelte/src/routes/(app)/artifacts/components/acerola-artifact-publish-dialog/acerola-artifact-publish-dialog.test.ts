import { render, screen } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';

import PublishDialog, {
  type AcerolaArtifactPublishDialogProps,
} from './acerola-artifact-publish-dialog.svelte';

function props(
  overrides: Partial<AcerolaArtifactPublishDialogProps> = {},
): AcerolaArtifactPublishDialogProps {
  return {
    data: {
      values: { file: null, title: '', label: '', kind: 'document', description: '', tags: '' },
      kindOptions: [
        { value: 'document', label: 'Documento' },
        { value: 'plan', label: 'Planejamento' },
        { value: 'idea', label: 'Ideia' },
      ],
      accept: '.html,.zip,.md,.pdf',
    },
    ui: { mode: 'new' },
    state: { isOpen: true, isSubmitting: false, error: null, fieldErrors: {} },
    actions: {
      onClose: vi.fn(),
      onChange: vi.fn(),
      onFileChange: vi.fn(),
      onBlur: vi.fn(),
      onSubmit: vi.fn(),
    },
    ...overrides,
  };
}

describe('AcerolaArtifactPublishDialog', () => {
  // feliz
  it('reports the file, the typed title and the submit', async () => {
    const current = props();
    render(PublishDialog, { props: current });
    const file = new File(['<h1>Guia</h1>'], 'guia.html', { type: 'text/html' });

    await userEvent.upload(screen.getByLabelText('Arquivo'), file);
    await userEvent.type(screen.getByLabelText(/Título/), 'G');
    await userEvent.click(screen.getByRole('button', { name: 'Publicar' }));

    expect(current.actions.onFileChange).toHaveBeenCalledWith(file);
    expect(current.actions.onChange).toHaveBeenCalledWith('title', 'G');
    expect(current.actions.onSubmit).toHaveBeenCalled();
  });

  // feliz
  it('locks the label and hides the kind when publishing a new version', () => {
    render(PublishDialog, {
      props: props({
        ui: { mode: 'version' },
        data: { ...props().data, values: { ...props().data.values, label: 'abertura-cnpj' } },
      }),
    });

    expect(screen.getByRole('heading', { name: 'Publicar nova versão' })).toBeInTheDocument();
    expect(screen.getByLabelText(/Label/)).toBeDisabled();
    expect(screen.queryByText('Tipo')).not.toBeInTheDocument();
  });

  // triste
  it('shows field errors and the server refusal while keeping the form', () => {
    render(PublishDialog, {
      props: props({
        state: {
          isOpen: true,
          isSubmitting: false,
          error: 'Outra publicação desta label terminou antes.',
          fieldErrors: { file: 'Escolha o arquivo a publicar', title: 'Informe o título' },
        },
      }),
    });

    expect(screen.getByText('Escolha o arquivo a publicar')).toBeInTheDocument();
    expect(screen.getByText('Informe o título')).toBeInTheDocument();
    expect(screen.getByText('Outra publicação desta label terminou antes.')).toBeInTheDocument();
  });

  // triste
  it('locks every field and the buttons while publishing', () => {
    render(PublishDialog, {
      props: props({ state: { isOpen: true, isSubmitting: true, error: null, fieldErrors: {} } }),
    });

    expect(screen.getByLabelText('Arquivo')).toBeDisabled();
    expect(screen.getByRole('button', { name: 'Cancelar' })).toBeDisabled();
    expect(screen.getByRole('button', { name: /Publicando/ })).toBeInTheDocument();
  });
});

import { render, screen } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';

import PublishDialog, { type AcerolaPublishDialogProps } from './acerola-publish-dialog.svelte';

function props(overrides: Partial<AcerolaPublishDialogProps> = {}): AcerolaPublishDialogProps {
  return {
    data: {
      values: {
        file: null,
        title: '',
        label: '',
        description: '',
        tags: '',
        status: 'draft',
        visibility: 'private',
      },
      accept: '.html,.zip,.md,.pdf',
      copy: {
        title: 'Publicar documento',
        description: 'O arquivo vira uma página servida pelo hub.',
        submitLabel: 'Publicar documento',
        titlePlaceholder: 'Guia de instalação',
      },
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

describe('AcerolaPublishDialog', () => {
  // feliz
  it('reports the file, the typed title and the submit', async () => {
    const current = props();
    render(PublishDialog, { props: current });
    const file = new File(['<h1>Guia</h1>'], 'guia.html', { type: 'text/html' });

    await userEvent.upload(screen.getByLabelText('Arquivo'), file);
    await userEvent.type(screen.getByLabelText(/Título/), 'G');
    await userEvent.click(screen.getByRole('button', { name: 'Publicar documento' }));

    expect(current.actions.onFileChange).toHaveBeenCalledWith(file);
    expect(current.actions.onChange).toHaveBeenCalledWith('title', 'G');
    expect(current.actions.onSubmit).toHaveBeenCalled();
  });

  // feliz
  it('asks the status of a plan or idea and the access', async () => {
    const current = props({ ui: { mode: 'new', hasStatus: true } });
    render(PublishDialog, { props: current });

    expect(screen.getByText('Situação')).toBeInTheDocument();
    expect(screen.getByRole('group', { name: 'Acesso' })).toHaveTextContent('Privado');
    await userEvent.click(screen.getByRole('button', { name: 'Em andamento' }));
    await userEvent.click(screen.getByRole('button', { name: 'Com o link' }));

    expect(current.actions.onChange).toHaveBeenCalledWith('status', 'active');
    expect(current.actions.onChange).toHaveBeenCalledWith('visibility', 'public');
  });

  // triste
  it('has no status for a document, and neither status nor access on a new version', () => {
    const { unmount } = render(PublishDialog, { props: props() });
    expect(screen.queryByText('Situação')).toBeNull();
    expect(screen.getByText('Acesso')).toBeInTheDocument();
    unmount();

    render(PublishDialog, { props: props({ ui: { mode: 'version', hasStatus: false } }) });
    expect(screen.queryByText('Acesso')).toBeNull();
  });

  // feliz
  it('locks the label when publishing a new version', () => {
    render(PublishDialog, {
      props: props({
        ui: { mode: 'version' },
        data: {
          ...props().data,
          values: { ...props().data.values, label: 'abertura-cnpj' },
          copy: { ...props().data.copy, title: 'Publicar nova versão' },
        },
      }),
    });

    expect(screen.getByRole('heading', { name: 'Publicar nova versão' })).toBeInTheDocument();
    expect(screen.getByLabelText(/Label/)).toBeDisabled();
    expect(screen.getByPlaceholderText('Manter o atual')).toBeInTheDocument();
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

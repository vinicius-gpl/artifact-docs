import { render, screen } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';

import FileField from './acerola-file-field.svelte';

const data = { label: 'Arquivo', name: 'file', accept: '.html,.zip', fileName: null };

describe('AcerolaFileField', () => {
  // feliz
  it('reports the chosen file', async () => {
    const onChange = vi.fn();
    render(FileField, { props: { data, actions: { onChange } } });
    const file = new File(['<h1>oi</h1>'], 'guia.html', { type: 'text/html' });

    await userEvent.upload(screen.getByLabelText('Arquivo'), file);

    expect(onChange).toHaveBeenCalledWith(file);
  });

  // feliz
  it('shows the chosen file name and offers to change it', () => {
    render(FileField, { props: { data: { ...data, fileName: 'guia.html' } } });

    expect(screen.getByText('guia.html')).toBeInTheDocument();
    expect(screen.getByText('Trocar')).toBeInTheDocument();
  });

  // triste
  it('announces the error to assistive technology instead of the hint', () => {
    render(FileField, {
      props: {
        data: { ...data, hint: 'HTML ou .zip' },
        state: { error: 'Escolha o arquivo a publicar' },
      },
    });

    expect(screen.getByLabelText('Arquivo')).toHaveAttribute('aria-invalid', 'true');
    expect(screen.getByRole('alert')).toHaveTextContent('Escolha o arquivo a publicar');
    expect(screen.queryByText('HTML ou .zip')).not.toBeInTheDocument();
  });

  // triste
  it('does not report anything while disabled', async () => {
    const onChange = vi.fn();
    render(FileField, { props: { data, state: { isDisabled: true }, actions: { onChange } } });

    await userEvent.upload(screen.getByLabelText('Arquivo'), new File(['x'], 'x.html'));

    expect(onChange).not.toHaveBeenCalled();
  });
});

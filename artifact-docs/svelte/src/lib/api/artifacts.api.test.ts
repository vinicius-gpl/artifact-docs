import { describe, expect, it } from 'vitest';

import { toPublishForm } from './artifacts.api';

const file = new File(['<h1>Guia</h1>'], 'guia.html', { type: 'text/html' });

describe('toPublishForm', () => {
  // feliz
  it('sends the file and every filled field in the multipart the Go reads', () => {
    const form = toPublishForm({
      file,
      title: ' Guia rápido ',
      label: 'guia-rapido',
      kind: 'document',
      description: 'Para começar',
      tags: 'hub, marca',
      status: 'active',
      isPrivate: true,
    });

    expect((form.get('file') as File).name).toBe('guia.html');
    expect(form.get('status')).toBe('active');
    expect(form.get('private')).toBe('true');
    expect(form.get('title')).toBe('Guia rápido');
    expect(form.get('kind')).toBe('document');
    expect(form.get('tags')).toBe('hub, marca');
  });

  // triste
  it('leaves empty fields out, so a new version keeps what is published', () => {
    const form = toPublishForm({ file, title: '', label: 'guia', description: '  ', tags: '' });

    expect(form.has('kind')).toBe(false);
    expect(form.has('title')).toBe(false);
    expect(form.has('description')).toBe(false);
    expect(form.has('tags')).toBe(false);
    expect(form.has('status')).toBe(false);
    expect(form.has('private')).toBe(false);
  });
});

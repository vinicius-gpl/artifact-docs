import { render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';

import ShortcutList from './acerola-shortcut-list.svelte';

describe('AcerolaShortcutList', () => {
  // feliz
  it('lists each shortcut as a link with its detail', () => {
    render(ShortcutList, {
      props: {
        data: {
          title: 'Fixados',
          emptyLabel: 'Nada fixado.',
          items: [{ id: '1', label: 'Briefing', href: '/artifacts/briefing', meta: 'Documento' }],
        },
      },
    });

    expect(screen.getByRole('heading', { name: 'Fixados' })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: /Briefing/ })).toHaveAttribute(
      'href',
      '/artifacts/briefing',
    );
    expect(screen.getByText('Documento')).toBeInTheDocument();
  });

  // triste
  it('says how the block fills up when it is empty', () => {
    render(ShortcutList, {
      props: {
        data: { title: 'Recentes', emptyLabel: 'Abra um item para ele aparecer aqui.', items: [] },
      },
    });

    expect(screen.getByText('Abra um item para ele aparecer aqui.')).toBeInTheDocument();
    expect(screen.queryByRole('link')).toBeNull();
  });
});

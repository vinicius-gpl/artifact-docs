import { fireEvent, render, screen } from '@testing-library/svelte';
import { describe, expect, it, vi } from 'vitest';

import { type ArtifactView } from '$lib/hooks/use-artifact-detail/use-artifact-detail.svelte';
import ArtifactViewer from './acerola-artifact-viewer.svelte';

const ARTIFACT: ArtifactView = {
  title: 'Briefing do hub',
  description: 'O que é o hub.',
  url: 'data:text/html,<h1>Briefing</h1>',
  badges: [{ label: 'Documento', tone: 'brand' }],
  isPrivate: true,
  facts: [{ label: 'Formato', value: 'Markdown' }],
  versions: [
    { number: 2, version: 'v2', detail: '25/10/2026 · 12 KB', isCurrent: true },
    { number: 1, version: 'v1', detail: '18/10/2026 · 11 KB', isCurrent: false },
  ],
};

const idle = { isLoading: false, error: null };

describe('AcerolaArtifactViewer', () => {
  // feliz
  it('shows the page sandboxed, without same-origin access', () => {
    render(ArtifactViewer, {
      props: { data: { artifact: ARTIFACT }, state: idle, actions: { onRetry: vi.fn() } },
    });

    const frame = screen.getByTitle('Briefing do hub');
    expect(frame).toHaveAttribute('sandbox', 'allow-scripts allow-popups allow-forms');
    expect(screen.getByRole('heading', { level: 1, name: /Briefing do hub/ })).toBeInTheDocument();
    expect(screen.getByText('Atual')).toBeInTheDocument();
    expect(screen.getByRole('img', { name: 'Privado' })).toBeInTheDocument();
  });

  // triste
  it('shows the reason and retries when the artifact does not open', async () => {
    const onRetry = vi.fn();
    render(ArtifactViewer, {
      props: {
        data: { artifact: null },
        state: { ...idle, error: 'Artefato não encontrado.' },
        actions: { onRetry },
      },
    });

    expect(screen.getByText('Artefato não encontrado.')).toBeInTheDocument();
    await fireEvent.click(screen.getByRole('button', { name: /Tentar de novo/ }));
    expect(onRetry).toHaveBeenCalled();
  });

  it('keeps the way back while loading', () => {
    render(ArtifactViewer, {
      props: {
        data: { artifact: null },
        state: { ...idle, isLoading: true },
        actions: { onRetry: vi.fn() },
      },
    });

    expect(screen.getByRole('link', { name: /Voltar para Artefatos/ })).toHaveAttribute(
      'href',
      '/artifacts',
    );
  });

  // feliz
  it('offers the actions and restores an older version by its number', async () => {
    const actions = {
      onRetry: vi.fn(),
      onPublishVersion: vi.fn(),
      onTogglePrivate: vi.fn(),
      onRestore: vi.fn(),
      onDelete: vi.fn(),
    };
    render(ArtifactViewer, { props: { data: { artifact: ARTIFACT }, state: idle, actions } });

    await fireEvent.click(screen.getByRole('button', { name: 'Restaurar v1' }));
    await fireEvent.click(screen.getByRole('button', { name: 'Tornar público' }));
    await fireEvent.click(screen.getByRole('button', { name: 'Publicar nova versão' }));
    await fireEvent.click(screen.getByRole('button', { name: 'Apagar' }));

    expect(actions.onRestore).toHaveBeenCalledWith(1);
    expect(actions.onTogglePrivate).toHaveBeenCalled();
    expect(actions.onPublishVersion).toHaveBeenCalled();
    expect(actions.onDelete).toHaveBeenCalled();
    expect(screen.queryByRole('button', { name: 'Restaurar v2' })).not.toBeInTheDocument();
  });

  // triste
  it('locks the actions while one runs and shows why the last one failed', () => {
    render(ArtifactViewer, {
      props: {
        data: { artifact: ARTIFACT },
        state: {
          ...idle,
          pendingAction: { kind: 'private' },
          actionError: 'Não consegui salvar a alteração.',
        },
        actions: { onRetry: vi.fn(), onDelete: vi.fn(), onRestore: vi.fn() },
      },
    });

    expect(screen.getByRole('button', { name: 'Apagar' })).toBeDisabled();
    expect(screen.getByRole('button', { name: 'Restaurar v1' })).toBeDisabled();
    expect(screen.getByText('Não consegui salvar a alteração.')).toBeInTheDocument();
  });
});

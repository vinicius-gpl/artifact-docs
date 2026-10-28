import { fireEvent, render, screen } from '@testing-library/svelte';
import { describe, expect, it, vi } from 'vitest';

import { type ArtifactView } from '$lib/hooks/use-artifact-detail/use-artifact-detail.svelte';
import type * as Motion from '$lib/motion/motion';
import ArtifactViewer from './acerola-artifact-viewer.svelte';

/* O voo da tela cheia é do GSAP; aqui só importa o que a prévia faz ANTES e DEPOIS dele. Cada
   teste pode segurar o voo (`landing`) para olhar a prévia no meio do caminho. */
const flight = vi.hoisted(() => ({ landing: null as (() => void) | null }));
vi.mock('$lib/motion/motion', async (importOriginal) => ({
  ...(await importOriginal<typeof Motion>()),
  flipFrom: () =>
    flight.landing === null
      ? Promise.resolve()
      : new Promise<void>((resolve) => {
          flight.landing = resolve;
        }),
}));

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

    expect(screen.getByRole('link', { name: 'Voltar' })).toHaveAttribute('href', '/artifacts');
  });

  // feliz
  it('goes back to the tab of the artifact kind', () => {
    render(ArtifactViewer, {
      props: {
        data: {
          artifact: ARTIFACT,
          back: { href: '/plans', label: 'Voltar para Planejamentos', navKey: 'plans' },
        },
        state: idle,
        actions: { onRetry: vi.fn() },
      },
    });

    expect(screen.getByRole('link', { name: /Voltar para Planejamentos/ })).toHaveAttribute(
      'href',
      '/plans',
    );
  });

  // feliz
  it('offers the actions and restores an older version by its number', async () => {
    const actions = {
      onRetry: vi.fn(),
      onPublishVersion: vi.fn(),
      onEdit: vi.fn(),
      onTogglePrivate: vi.fn(),
      onRestore: vi.fn(),
      onDelete: vi.fn(),
    };
    render(ArtifactViewer, { props: { data: { artifact: ARTIFACT }, state: idle, actions } });

    await fireEvent.click(screen.getByRole('button', { name: 'Restaurar v1' }));
    await fireEvent.click(screen.getByRole('button', { name: 'Tornar público' }));
    await fireEvent.click(screen.getByRole('button', { name: 'Publicar nova versão' }));
    await fireEvent.click(screen.getByRole('button', { name: 'Apagar' }));
    await fireEvent.click(screen.getByRole('button', { name: 'Editar' }));

    expect(actions.onRestore).toHaveBeenCalledWith(1);
    expect(actions.onTogglePrivate).toHaveBeenCalled();
    expect(actions.onPublishVersion).toHaveBeenCalled();
    expect(actions.onDelete).toHaveBeenCalled();
    expect(actions.onEdit).toHaveBeenCalled();
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

  // feliz
  it('opens the published page in an in-app fullscreen frame and leaves it', async () => {
    render(ArtifactViewer, {
      props: { data: { artifact: ARTIFACT }, state: idle, actions: { onRetry: vi.fn() } },
    });

    await fireEvent.click(screen.getByRole('button', { name: 'Tela cheia' }));
    const frame = screen.getByRole('region', { name: 'Briefing do hub em tela cheia' });
    expect(frame).toHaveAttribute('data-expanded', 'true');

    await fireEvent.click(screen.getByRole('button', { name: 'Sair da tela cheia' }));
    expect(screen.getByRole('region', { name: 'Prévia da página publicada' })).toHaveAttribute(
      'data-expanded',
      'false',
    );
  });

  // triste
  it('leaves the fullscreen frame with Escape, and Escape does nothing outside it', async () => {
    render(ArtifactViewer, {
      props: { data: { artifact: ARTIFACT }, state: idle, actions: { onRetry: vi.fn() } },
    });

    await fireEvent.keyDown(window, { key: 'Escape' });
    expect(screen.getByRole('button', { name: 'Tela cheia' })).toBeInTheDocument();

    await fireEvent.click(screen.getByRole('button', { name: 'Tela cheia' }));
    await fireEvent.keyDown(window, { key: 'Escape' });

    expect(screen.getByRole('button', { name: 'Tela cheia' })).toBeInTheDocument();
  });

  // feliz
  it('keeps the preview above the page while it shrinks back, and lets go when it lands', async () => {
    render(ArtifactViewer, {
      props: { data: { artifact: ARTIFACT }, state: idle, actions: { onRetry: vi.fn() } },
    });
    await fireEvent.click(screen.getByRole('button', { name: 'Tela cheia' }));

    flight.landing = () => {};
    await fireEvent.click(screen.getByRole('button', { name: 'Sair da tela cheia' }));
    const preview = screen.getByRole('region', { name: 'Prévia da página publicada' });
    expect(preview).toHaveAttribute('data-settling', 'true');
    expect(preview.className).toContain('z-50');

    flight.landing();
    flight.landing = null;
    await vi.waitFor(() => expect(preview).toHaveAttribute('data-settling', 'false'));
    expect(preview.className).not.toContain('z-50');
  });

  // triste
  it('never holds the preview above the page when entering the fullscreen', async () => {
    render(ArtifactViewer, {
      props: { data: { artifact: ARTIFACT }, state: idle, actions: { onRetry: vi.fn() } },
    });

    await fireEvent.click(screen.getByRole('button', { name: 'Tela cheia' }));

    expect(screen.getByRole('region', { name: 'Briefing do hub em tela cheia' })).toHaveAttribute(
      'data-settling',
      'false',
    );
  });
});

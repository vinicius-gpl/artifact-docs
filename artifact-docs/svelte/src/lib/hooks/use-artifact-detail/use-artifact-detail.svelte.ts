import { goto } from '$app/navigation';
import { createMutation, createQuery, useQueryClient } from '@tanstack/svelte-query';

import { artifactsApi } from '$lib/api/artifacts.api';
import { readError } from '$lib/api/http-client';
import { type PreviewBadge } from '$lib/components/acerola-preview-card/acerola-preview-card.svelte';
import {
  ARTIFACT_FORMAT_LABELS,
  ARTIFACT_KIND_LABELS,
  ARTIFACT_STATUS_LABELS,
  type ArtifactDetail,
  type ArtifactKind,
} from '$lib/contracts/artifact.schema';
import { STATUS_TONES } from '$lib/hooks/use-artifact-gallery/use-artifact-gallery.svelte';
import { mirrorStore } from '$lib/hooks/use-mirror-store/use-mirror-store.svelte';
import { formatDateTime } from '$lib/utils/format-date';
import { formatBytes } from '$lib/utils/format-machine';

export type ArtifactFact = { label: string; value: string };

export type ArtifactVersionRow = {
  /** O número, para restaurar; `version` é o rótulo ("v3"). */
  number: number;
  version: string;
  detail: string;
  isCurrent: boolean;
};

export type ArtifactView = {
  title: string;
  description: string | null;
  url: string;
  badges: PreviewBadge[];
  isPrivate: boolean;
  facts: ArtifactFact[];
  versions: ArtifactVersionRow[];
};

export type BackLink = { href: string; label: string };

/** A aba de cada tipo: é para onde o "Voltar" leva e onde a tela cai depois de apagar. */
export const KIND_TABS: Record<ArtifactKind, BackLink> = {
  document: { href: '/artifacts', label: 'Voltar para Documentos' },
  plan: { href: '/plans', label: 'Voltar para Planejamentos' },
  idea: { href: '/ideas', label: 'Voltar para Ideias' },
};

/** Enquanto o artefato não chega, o "Voltar" não sabe a aba: volta para a lista de sempre. */
const DEFAULT_BACK: BackLink = { href: '/artifacts', label: 'Voltar' };

/** A ação que está em andamento no detalhe — uma por vez. */
export type ArtifactPendingAction =
  { kind: 'private' } | { kind: 'restore'; version: number } | null;

export type ArtifactDetailModel = {
  data: { artifact: ArtifactView | null; back: BackLink };
  state: {
    isLoading: boolean;
    error: string | null;
    pendingAction: ArtifactPendingAction;
    actionError: string | null;
    isDeleteOpen: boolean;
    isDeleting: boolean;
    deleteError: string | null;
  };
  actions: {
    onRetry: () => void;
    onTogglePrivate: () => void;
    onRestore: (version: number) => void;
    onDeleteRequest: () => void;
    onDeleteCancel: () => void;
    onDeleteConfirm: () => void;
  };
};

export function toArtifactView(artifact: ArtifactDetail): ArtifactView {
  const badges: PreviewBadge[] = [{ label: ARTIFACT_KIND_LABELS[artifact.kind], tone: 'brand' }];
  if (artifact.status) {
    badges.push({
      label: ARTIFACT_STATUS_LABELS[artifact.status],
      tone: STATUS_TONES[artifact.status],
    });
  }

  return {
    title: artifact.title,
    description: artifact.description,
    url: artifact.url,
    badges,
    isPrivate: artifact.isPrivate,
    facts: [
      { label: 'Label', value: artifact.label },
      { label: 'Formato', value: ARTIFACT_FORMAT_LABELS[artifact.format] },
      { label: 'Tamanho', value: formatBytes(artifact.sizeBytes) },
      { label: 'Acesso', value: artifact.isPrivate ? 'Privado' : 'Qualquer um com o link' },
      { label: 'Editado em', value: formatDateTime(artifact.updatedAt) },
      { label: 'Criado em', value: formatDateTime(artifact.createdAt) },
    ],
    versions: artifact.versions.map((version) => ({
      number: version.version,
      version: `v${version.version}`,
      detail: `${formatDateTime(version.createdAt)} · ${formatBytes(version.sizeBytes)}`,
      isCurrent: version.version === artifact.currentVersion,
    })),
  };
}

/**
 * Um artefato aberto: a página publicada, os dados dele e as versões — e o que se faz com
 * ele sem subir arquivo: tornar privado ou público, voltar para uma versão anterior, apagar.
 *
 * Toda ação invalida TODAS as consultas de artefato: o mesmo artefato aparece na galeria, na
 * do tipo dele e nos Recentes, e todas precisam mostrar o cadeado e a versão iguais.
 */
export function useArtifactDetailModel(input: { label: string }): ArtifactDetailModel {
  const queryClient = useQueryClient();
  const detail = mirrorStore(
    createQuery({
      queryKey: ['artifacts', 'detail', input.label],
      queryFn: () => artifactsApi.get(input.label),
    }),
  );

  let isDeleteOpen = $state(false);
  const backOf = (): BackLink => {
    const kind = detail.current.data?.kind;

    return kind ? KIND_TABS[kind] : DEFAULT_BACK;
  };
  const refresh = () => queryClient.invalidateQueries({ queryKey: ['artifacts'] });

  const privacy = mirrorStore(
    createMutation({
      mutationFn: (isPrivate: boolean) => artifactsApi.update(input.label, { isPrivate }),
      onSuccess: refresh,
    }),
  );
  const restore = mirrorStore(
    createMutation({
      mutationFn: (version: number) => artifactsApi.restore(input.label, version),
      onSuccess: refresh,
    }),
  );
  const removal = mirrorStore(
    createMutation({
      mutationFn: () => artifactsApi.remove(input.label),
      onSuccess: async () => {
        const back = backOf();
        /* Tira o detalhe do cache antes de sair: voltar pelo histórico não pode mostrar o
           artefato apagado como se existisse. */
        queryClient.removeQueries({ queryKey: ['artifacts', 'detail', input.label] });
        await refresh();
        isDeleteOpen = false;
        await goto(back.href);
      },
    }),
  );

  function pendingAction(): ArtifactPendingAction {
    if (privacy.current.isPending) return { kind: 'private' };
    if (restore.current.isPending)
      return { kind: 'restore', version: restore.current.variables ?? 0 };

    return null;
  }

  return {
    get data() {
      const artifact = detail.current.data;

      return { artifact: artifact ? toArtifactView(artifact) : null, back: backOf() };
    },
    get state() {
      return {
        isLoading: detail.current.isPending,
        error: readError(detail.current.error),
        pendingAction: pendingAction(),
        actionError: readError(privacy.current.error) ?? readError(restore.current.error),
        isDeleteOpen,
        isDeleting: removal.current.isPending,
        deleteError: readError(removal.current.error),
      };
    },
    actions: {
      onRetry: () => {
        void detail.current.refetch();
      },
      onTogglePrivate: () => {
        const artifact = detail.current.data;
        if (!artifact || pendingAction()) return;
        restore.current.reset();
        privacy.current.mutate(!artifact.isPrivate);
      },
      onRestore: (version) => {
        if (pendingAction()) return;
        privacy.current.reset();
        restore.current.mutate(version);
      },
      onDeleteRequest: () => {
        removal.current.reset();
        isDeleteOpen = true;
      },
      onDeleteCancel: () => {
        if (removal.current.isPending) return;
        isDeleteOpen = false;
      },
      onDeleteConfirm: () => {
        if (removal.current.isPending) return;
        removal.current.mutate();
      },
    },
  };
}

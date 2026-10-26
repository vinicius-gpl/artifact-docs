import { createMutation, createQuery, useQueryClient } from '@tanstack/svelte-query';

import { artifactsApi } from '$lib/api/artifacts.api';
import { readError } from '$lib/api/http-client';
import {
  type PreviewBadge,
  type PreviewItem,
} from '$lib/components/acerola-preview-card/acerola-preview-card.svelte';
import { type StatusBadgeTone } from '$lib/components/acerola-status-badge/acerola-status-badge.svelte';
import {
  ARTIFACT_FORMAT_LABELS,
  ARTIFACT_KIND_LABELS,
  ARTIFACT_STATUS_LABELS,
  artifactKindSchema,
  artifactStatusSchema,
  type Artifact,
  type ArtifactKind,
  type ArtifactStatus,
} from '$lib/contracts/artifact.schema';
import { mirrorStore } from '$lib/hooks/use-mirror-store/use-mirror-store.svelte';
import { formatTimeAgo } from '$lib/utils/format-machine';

export type GalleryFilters = {
  search: string;
  kind: ArtifactKind | '';
  status: ArtifactStatus | '';
};

export type FilterOption = { value: string; label: string; tone?: StatusBadgeTone };

export type ShortcutLink = { id: string; label: string; href: string; meta: string };

export type ArtifactGalleryModel = {
  data: { items: PreviewItem[]; pinned: ShortcutLink[]; recent: ShortcutLink[]; total: number };
  ui: { kindOptions: FilterOption[]; statusOptions: FilterOption[] };
  state: {
    isLoading: boolean;
    error: string | null;
    filters: GalleryFilters;
    hasFilters: boolean;
    pendingPinId: string | null;
    pinError: string | null;
  };
  actions: {
    onSearch: (value: string) => void;
    onKindChange: (value: string) => void;
    onStatusChange: (value: string) => void;
    onClearFilters: () => void;
    onRetry: () => void;
    onTogglePin: (id: string) => void;
  };
};

/** O tom de cada situação — decidido aqui, uma vez, para a cor nunca variar entre telas. */
export const STATUS_TONES: Record<ArtifactStatus, StatusBadgeTone> = {
  draft: 'neutral',
  active: 'info',
  done: 'success',
  archived: 'warning',
};

const EMPTY_FILTERS: GalleryFilters = { search: '', kind: '', status: '' };

/** Quantos atalhos cabem em "Recentes" sem virar uma segunda galeria. */
const RECENT_LIMIT = 5;

/**
 * Traduz o artefato para o cartão genérico. `withKind` liga o selo do tipo: na galeria de
 * tudo ele diz se é documento, plano ou ideia; na de planejamentos seria repetir o título da tela.
 */
export function toPreviewItem(artifact: Artifact, withKind: boolean, now: number): PreviewItem {
  const badges: PreviewBadge[] = [];
  if (artifact.status) {
    badges.push({
      label: ARTIFACT_STATUS_LABELS[artifact.status],
      tone: STATUS_TONES[artifact.status],
    });
  }
  if (withKind) badges.push({ label: ARTIFACT_KIND_LABELS[artifact.kind], tone: 'brand' });
  badges.push({ label: ARTIFACT_FORMAT_LABELS[artifact.format], tone: 'neutral' });

  return {
    id: artifact.id,
    title: artifact.title,
    description: artifact.description,
    href: `/artifacts/${artifact.label}`,
    previewUrl: artifact.url,
    thumbUrl: artifact.thumbUrl,
    meta: `Editado ${formatTimeAgo(artifact.updatedAt, now)}`,
    badges,
    isLocked: artifact.isPrivate,
    isPinned: artifact.isPinned,
  };
}

function matchesSearch(artifact: Artifact, search: string): boolean {
  const term = search.trim().toLocaleLowerCase('pt-BR');
  if (!term) return true;

  return [artifact.title, artifact.description ?? '', artifact.label, ...artifact.tags]
    .join(' ')
    .toLocaleLowerCase('pt-BR')
    .includes(term);
}

/**
 * Os filtros correm NO NAVEGADOR, sobre a lista já carregada: a resposta a cada tecla é
 * imediata e não custa uma ida ao Neon por letra. Quando a lista crescer a ponto de pesar, a
 * busca passa para o `?q=` que a API já aceita (`tsvector`).
 */
export function filterArtifacts(items: Artifact[], filters: GalleryFilters): Artifact[] {
  return items
    .filter((artifact) => !filters.kind || artifact.kind === filters.kind)
    .filter((artifact) => !filters.status || artifact.status === filters.status)
    .filter((artifact) => matchesSearch(artifact, filters.search));
}

function toShortcut(artifact: Artifact, meta: string): ShortcutLink {
  return { id: artifact.id, label: artifact.title, href: `/artifacts/${artifact.label}`, meta };
}

export function pinnedOf(items: Artifact[]): ShortcutLink[] {
  return items
    .filter((artifact) => artifact.isPinned)
    .map((artifact) => toShortcut(artifact, ARTIFACT_KIND_LABELS[artifact.kind]));
}

/** "Recentes" é o último ACESSO real, não a última edição (briefing, módulo 2). */
export function recentOf(items: Artifact[], now: number): ShortcutLink[] {
  return items
    .filter((artifact) => artifact.lastViewedAt)
    .sort((a, b) => (b.lastViewedAt ?? '').localeCompare(a.lastViewedAt ?? ''))
    .slice(0, RECENT_LIMIT)
    .map((artifact) => toShortcut(artifact, `Aberto ${formatTimeAgo(artifact.lastViewedAt, now)}`));
}

function kindOptions(): FilterOption[] {
  return artifactKindSchema.options.map((kind) => ({
    value: kind,
    label: ARTIFACT_KIND_LABELS[kind],
  }));
}

function statusOptions(): FilterOption[] {
  return artifactStatusSchema.options.map((status) => ({
    value: status,
    label: ARTIFACT_STATUS_LABELS[status],
    tone: STATUS_TONES[status],
  }));
}

/**
 * A galeria de artefatos — a mesma para as três telas: Artefatos (tudo, com filtro por tipo),
 * Planejamentos e Ideias (um `kind` fixo, com filtro por situação).
 *
 * `kind` fixo vai para a CONSULTA (a API só devolve aquele tipo); o filtro de tipo da galeria
 * de tudo fica no navegador, junto com a busca.
 */
export function useArtifactGalleryModel(input: { kind?: ArtifactKind } = {}): ArtifactGalleryModel {
  const queryClient = useQueryClient();
  const queryKey = ['artifacts', { kind: input.kind ?? null }];
  const list = mirrorStore(
    createQuery({ queryKey, queryFn: () => artifactsApi.list({ kind: input.kind }) }),
  );

  let filters = $state<GalleryFilters>({ ...EMPTY_FILTERS });

  const pin = mirrorStore(
    createMutation({
      mutationFn: (artifact: Artifact) =>
        artifactsApi.update(artifact.label, { isPinned: !artifact.isPinned }),
      /* Invalida TODAS as listas de artefatos: o mesmo item aparece na galeria de tudo e na
         do tipo dele, e as duas precisam mostrar o alfinete igual. */
      onSuccess: () => queryClient.invalidateQueries({ queryKey: ['artifacts'] }),
    }),
  );

  const artifacts = $derived(list.current.data?.items ?? []);
  const visible = $derived(filterArtifacts(artifacts, filters));

  return {
    get data() {
      const now = Date.now();

      return {
        items: visible.map((artifact) => toPreviewItem(artifact, !input.kind, now)),
        pinned: pinnedOf(artifacts),
        recent: recentOf(artifacts, now),
        total: artifacts.length,
      };
    },
    ui: { kindOptions: kindOptions(), statusOptions: statusOptions() },
    get state() {
      const pending = pin.current.isPending ? pin.current.variables : undefined;

      return {
        isLoading: list.current.isPending,
        error: readError(list.current.error),
        filters,
        hasFilters: Boolean(filters.search || filters.kind || filters.status),
        pendingPinId: pending?.id ?? null,
        pinError: readError(pin.current.error),
      };
    },
    actions: {
      onSearch: (value) => {
        filters.search = value;
      },
      onKindChange: (value) => {
        const kind = artifactKindSchema.safeParse(value);
        filters.kind = kind.success ? kind.data : '';
      },
      onStatusChange: (value) => {
        const status = artifactStatusSchema.safeParse(value);
        filters.status = status.success ? status.data : '';
      },
      onClearFilters: () => {
        filters = { ...EMPTY_FILTERS };
      },
      onRetry: () => {
        void list.current.refetch();
      },
      onTogglePin: (id) => {
        const artifact = artifacts.find((item) => item.id === id);
        if (artifact) pin.current.mutate(artifact);
      },
    },
  };
}

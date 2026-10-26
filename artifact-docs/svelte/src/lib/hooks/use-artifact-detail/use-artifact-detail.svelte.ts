import { createQuery } from '@tanstack/svelte-query';

import { artifactsApi } from '$lib/api/artifacts.api';
import { readError } from '$lib/api/http-client';
import { type PreviewBadge } from '$lib/components/acerola-preview-card/acerola-preview-card.svelte';
import {
  ARTIFACT_FORMAT_LABELS,
  ARTIFACT_KIND_LABELS,
  ARTIFACT_STATUS_LABELS,
  type ArtifactDetail,
} from '$lib/contracts/artifact.schema';
import { STATUS_TONES } from '$lib/hooks/use-artifact-gallery/use-artifact-gallery.svelte';
import { mirrorStore } from '$lib/hooks/use-mirror-store/use-mirror-store.svelte';
import { formatDateTime } from '$lib/utils/format-date';
import { formatBytes } from '$lib/utils/format-machine';

export type ArtifactFact = { label: string; value: string };

export type ArtifactVersionRow = { version: string; detail: string; isCurrent: boolean };

export type ArtifactView = {
  title: string;
  description: string | null;
  url: string;
  badges: PreviewBadge[];
  isPrivate: boolean;
  facts: ArtifactFact[];
  versions: ArtifactVersionRow[];
};

export type ArtifactDetailModel = {
  data: { artifact: ArtifactView | null };
  state: { isLoading: boolean; error: string | null };
  actions: { onRetry: () => void };
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
      { label: 'Acesso', value: artifact.isPrivate ? 'Exige login' : 'Quem tiver o link' },
      { label: 'Editado em', value: formatDateTime(artifact.updatedAt) },
      { label: 'Criado em', value: formatDateTime(artifact.createdAt) },
    ],
    versions: artifact.versions.map((version) => ({
      version: `v${version.version}`,
      detail: `${formatDateTime(version.createdAt)} · ${formatBytes(version.sizeBytes)}`,
      isCurrent: version.version === artifact.currentVersion,
    })),
  };
}

/** Um artefato aberto: a página publicada, os dados dele e as versões. */
export function useArtifactDetailModel(input: { label: string }): ArtifactDetailModel {
  const detail = mirrorStore(
    createQuery({
      queryKey: ['artifacts', 'detail', input.label],
      queryFn: () => artifactsApi.get(input.label),
    }),
  );

  return {
    get data() {
      const artifact = detail.current.data;

      return { artifact: artifact ? toArtifactView(artifact) : null };
    },
    get state() {
      return { isLoading: detail.current.isPending, error: readError(detail.current.error) };
    },
    actions: {
      onRetry: () => {
        void detail.current.refetch();
      },
    },
  };
}

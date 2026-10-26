import {
  type Artifact,
  type ArtifactDetail,
  type ArtifactList,
  type ArtifactListQuery,
  type ArtifactPatch,
} from '$lib/contracts/artifact.schema';

import { apiRequest } from './http-client';

/**
 * Uma função por endpoint de `/api/artifacts`. Documento, planejamento e ideia passam todos
 * por aqui: o que muda é o `kind` da consulta.
 */
export const artifactsApi = {
  list: (query: ArtifactListQuery = {}) => apiRequest<ArtifactList>('/artifacts', { query }),
  get: (label: string) => apiRequest<ArtifactDetail>(`/artifacts/${encodeURIComponent(label)}`),
  update: (label: string, patch: ArtifactPatch) =>
    apiRequest<Artifact>(`/artifacts/${encodeURIComponent(label)}`, {
      method: 'PATCH',
      body: patch,
    }),
};

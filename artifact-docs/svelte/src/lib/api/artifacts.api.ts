import {
  type Artifact,
  type ArtifactDetail,
  type ArtifactList,
  type ArtifactListQuery,
  type ArtifactPatch,
  type ArtifactVersionPublishInput,
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
  /** Publica (cria ou sobe uma versão nova). Campo vazio não vai: no Go, "não veio" mantém. */
  publish: (input: ArtifactVersionPublishInput) =>
    apiRequest<Artifact>(`/artifacts/${encodeURIComponent(input.label)}`, {
      method: 'PUT',
      body: toPublishForm(input),
    }),
  restore: (label: string, version: number) =>
    apiRequest<Artifact>(`/artifacts/${encodeURIComponent(label)}/restore`, {
      method: 'POST',
      body: { version },
    }),
  remove: (label: string) =>
    apiRequest<void>(`/artifacts/${encodeURIComponent(label)}`, { method: 'DELETE' }),
};

/** O multipart que o `PUT /api/artifacts/<label>` lê (`modules/artifacts/publish.go`). */
export function toPublishForm(input: ArtifactVersionPublishInput): FormData {
  const form = new FormData();
  form.append('file', input.file, input.file.name);
  if (input.kind) form.append('kind', input.kind);
  for (const field of ['title', 'description', 'tags'] as const) {
    const value = input[field].trim();
    if (value) form.append(field, value);
  }

  return form;
}

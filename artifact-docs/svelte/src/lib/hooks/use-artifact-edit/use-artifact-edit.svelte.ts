import { createMutation, useQueryClient } from '@tanstack/svelte-query';
import { SvelteSet } from 'svelte/reactivity';

import { artifactsApi } from '$lib/api/artifacts.api';
import { readError } from '$lib/api/http-client';
import {
  artifactPatchSchema,
  type ArtifactDetail,
  type ArtifactKind,
  type ArtifactPatch,
  type ArtifactStatus,
} from '$lib/contracts/artifact.schema';
import { mirrorStore } from '$lib/hooks/use-mirror-store/use-mirror-store.svelte';

/** O formulário de editar, como a pessoa digita. */
export type ArtifactEditValues = {
  title: string;
  description: string;
  /** Separadas por vírgula, como a pessoa digita. */
  tags: string;
  /** `''` em documento, que não tem situação. */
  status: string;
  visibility: ArtifactEditVisibility;
};

export type ArtifactEditVisibility = 'private' | 'public';

export type ArtifactEditField = keyof ArtifactEditValues;

export type ArtifactEditModel = {
  data: { values: ArtifactEditValues };
  ui: { hasStatus: boolean };
  state: {
    isOpen: boolean;
    isSubmitting: boolean;
    error: string | null;
    fieldErrors: Partial<Record<ArtifactEditField, string>>;
  };
  actions: {
    onOpen: () => void;
    onClose: () => void;
    onChange: (field: ArtifactEditField, value: string) => void;
    onBlur: (field: ArtifactEditField) => void;
    onSubmit: () => void;
  };
};

const EMPTY: ArtifactEditValues = {
  title: '',
  description: '',
  tags: '',
  status: '',
  visibility: 'private',
};

/** Planejamento e ideia têm situação; documento só existe ou não existe. */
function hasStatus(kind: ArtifactKind | undefined): boolean {
  return kind === 'plan' || kind === 'idea';
}

/** O formulário começa com o que está publicado — editar é mexer no que já existe. */
export function editValuesOf(artifact: ArtifactDetail): ArtifactEditValues {
  return {
    title: artifact.title,
    description: artifact.description ?? '',
    tags: artifact.tags.join(', '),
    status: hasStatus(artifact.kind) ? (artifact.status ?? 'draft') : '',
    visibility: artifact.isPrivate ? 'private' : 'public',
  };
}

/**
 * O que vai no PATCH. Manda tudo de uma vez: o formulário mostra o artefato inteiro, então o
 * que está nele é o que a pessoa quer. Descrição vazia vira `null` (limpa); tags vazias, `[]`.
 * Documento não manda situação: no Go, `status` mudaria um artefato que não tem situação.
 */
export function toPatch(values: ArtifactEditValues, kind: ArtifactKind | undefined): ArtifactPatch {
  const tags = values.tags
    .split(',')
    .map((tag) => tag.trim())
    .filter((tag, index, all) => tag !== '' && all.indexOf(tag) === index);
  const patch: ArtifactPatch = {
    title: values.title,
    description: values.description.trim() || null,
    tags,
    isPrivate: values.visibility !== 'public',
  };
  if (!hasStatus(kind)) return patch;

  /* Sem conversão aqui: um valor fora da lista chega ao schema e vira erro do campo. */
  return { ...patch, status: (values.status || 'draft') as ArtifactStatus };
}

/** O primeiro erro de cada campo, no formato que os campos mostram embaixo deles. */
export function editErrorsOf(
  values: ArtifactEditValues,
  kind: ArtifactKind | undefined,
): Partial<Record<ArtifactEditField, string>> {
  const result = artifactPatchSchema.safeParse(toPatch(values, kind));
  if (result.success) return {};

  const errors: Partial<Record<ArtifactEditField, string>> = {};
  for (const issue of result.error.issues) {
    const field = issue.path[0] as ArtifactEditField;
    errors[field] ??= issue.message;
  }

  return errors;
}

/**
 * EDITAR um artefato publicado — o diálogo do detalhe. Abre com o que está no ar (lido do
 * cache do detalhe, que a tela já carregou) e salva pelo PATCH.
 *
 * Salvo, o detalhe troca na hora (o PATCH devolve o artefato) e TODAS as consultas de artefato
 * são invalidadas: o mesmo card aparece na galeria, na aba do tipo e nos Recentes.
 */
export function useArtifactEditModel(input: { label: string }): ArtifactEditModel {
  const queryClient = useQueryClient();
  const detailKey = ['artifacts', 'detail', input.label];

  let values = $state<ArtifactEditValues>({ ...EMPTY });
  let kind = $state<ArtifactKind | undefined>(undefined);
  let isOpen = $state(false);
  let isSubmitted = $state(false);
  const touched = new SvelteSet<ArtifactEditField>();

  const save = mirrorStore(
    createMutation({
      mutationFn: (patch: ArtifactPatch) => artifactsApi.update(input.label, patch),
      onSuccess: async (artifact) => {
        queryClient.setQueryData<ArtifactDetail>(detailKey, (current) =>
          current ? { ...current, ...artifact } : current,
        );
        isOpen = false;
        await queryClient.invalidateQueries({ queryKey: ['artifacts'] });
      },
    }),
  );

  function visibleErrors(): Partial<Record<ArtifactEditField, string>> {
    const errors = editErrorsOf(values, kind);
    if (isSubmitted) return errors;

    return Object.fromEntries(
      Object.entries(errors).filter(([field]) => touched.has(field as ArtifactEditField)),
    );
  }

  return {
    get data() {
      return { values };
    },
    get ui() {
      return { hasStatus: hasStatus(kind) };
    },
    get state() {
      return {
        isOpen,
        isSubmitting: save.current.isPending,
        error: readError(save.current.error),
        fieldErrors: visibleErrors(),
      };
    },
    actions: {
      onOpen: () => {
        const artifact = queryClient.getQueryData<ArtifactDetail>(detailKey);
        if (!artifact) return;

        values = editValuesOf(artifact);
        kind = artifact.kind;
        isSubmitted = false;
        touched.clear();
        save.current.reset();
        isOpen = true;
      },
      onClose: () => {
        if (save.current.isPending) return;
        isOpen = false;
      },
      onChange: (field, value) => {
        /* O acesso só tem dois valores; qualquer outro texto vira o padrão, privado. */
        if (field === 'visibility') {
          values.visibility = value === 'public' ? 'public' : 'private';
          return;
        }
        values[field] = value;
      },
      onBlur: (field) => {
        touched.add(field);
      },
      onSubmit: () => {
        isSubmitted = true;
        const parsed = artifactPatchSchema.safeParse(toPatch(values, kind));
        if (!parsed.success || save.current.isPending) return;

        save.current.mutate(parsed.data);
      },
    },
  };
}

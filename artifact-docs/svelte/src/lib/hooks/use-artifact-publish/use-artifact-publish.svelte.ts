import { goto } from '$app/navigation';
import { createMutation, useQueryClient } from '@tanstack/svelte-query';
import { SvelteSet } from 'svelte/reactivity';

import { artifactsApi } from '$lib/api/artifacts.api';
import { readError } from '$lib/api/http-client';
import {
  type AcerolaPublishDialogProps,
  type PublishDialogMode,
  type PublishFormField,
  type PublishFormValues,
} from '$lib/components/acerola-publish-dialog/acerola-publish-dialog.svelte';
import {
  artifactPublishSchema,
  artifactVersionPublishSchema,
  PUBLISH_ACCEPT,
  type ArtifactKind,
  type ArtifactVersionPublishInput,
} from '$lib/contracts/artifact.schema';
import { mirrorStore } from '$lib/hooks/use-mirror-store/use-mirror-store.svelte';
import { slugify } from '$lib/utils/slug.util';

type PublishCopy = AcerolaPublishDialogProps['data']['copy'];

export type ArtifactPublishModel = {
  data: { values: PublishFormValues; accept: string; copy: PublishCopy };
  ui: { mode: PublishDialogMode; hasStatus: boolean };
  state: {
    isOpen: boolean;
    isSubmitting: boolean;
    error: string | null;
    fieldErrors: Partial<Record<PublishFormField, string>>;
  };
  actions: {
    onOpen: () => void;
    onClose: () => void;
    onChange: (field: Exclude<PublishFormField, 'file'>, value: string) => void;
    onFileChange: (file: File | null) => void;
    onBlur: (field: PublishFormField) => void;
    onSubmit: () => void;
  };
};

/**
 * O que cada aba publica — o tipo vem da aba, e o diálogo diz o nome dele. A versão nova não
 * precisa: o tipo do artefato não muda numa versão.
 */
export const PUBLISH_COPY: Record<ArtifactKind, PublishCopy> = {
  document: {
    title: 'Publicar documento',
    description: 'O arquivo vira uma página servida pelo hub, com preview em Documentos.',
    submitLabel: 'Publicar documento',
    titlePlaceholder: 'Guia de instalação',
  },
  plan: {
    title: 'Publicar planejamento',
    description: 'O plano vira uma página servida pelo hub, com preview em Planejamentos.',
    submitLabel: 'Publicar planejamento',
    titlePlaceholder: 'Abertura do CNPJ',
  },
  idea: {
    title: 'Publicar ideia',
    description: 'A ideia vira uma página servida pelo hub, com preview em Ideias.',
    submitLabel: 'Publicar ideia',
    titlePlaceholder: 'Gerador de paletas',
  },
};

const VERSION_COPY: PublishCopy = {
  title: 'Publicar nova versão',
  description: 'A versão nova entra no ar no lugar da atual. As anteriores continuam guardadas.',
  submitLabel: 'Publicar versão',
  titlePlaceholder: '',
};

/** Planejamento e ideia têm situação; documento só existe ou não existe. */
function hasStatus(kind: ArtifactKind | undefined): boolean {
  return kind === 'plan' || kind === 'idea';
}

/** O artefato novo nasce privado e, quando tem situação, como rascunho. */
function emptyValues(label: string | undefined, kind: ArtifactKind | undefined): PublishFormValues {
  return {
    file: null,
    title: '',
    label: label ?? '',
    description: '',
    tags: '',
    status: hasStatus(kind) ? 'draft' : '',
    visibility: 'private',
  };
}

/**
 * O que vai para a API. Artefato novo leva o tipo da aba, a situação e o acesso; a versão nova
 * NÃO leva nenhum dos três — mandado, o tipo trocaria o do artefato no Go, e situação e acesso
 * mudam no detalhe.
 */
function toInput(values: PublishFormValues, kind: ArtifactKind | undefined) {
  const { status, visibility, ...fields } = values;
  if (!kind) return fields;

  return {
    ...fields,
    kind,
    status: hasStatus(kind) ? status : undefined,
    isPrivate: visibility !== 'public',
  };
}

/** O primeiro erro de cada campo, no formato que os campos mostram embaixo deles. */
export function publishErrorsOf(
  values: PublishFormValues,
  kind: ArtifactKind | undefined,
): Partial<Record<PublishFormField, string>> {
  const schema = kind ? artifactPublishSchema : artifactVersionPublishSchema;
  const result = schema.safeParse(toInput(values, kind));
  if (result.success) return {};

  const errors: Partial<Record<PublishFormField, string>> = {};
  for (const issue of result.error.issues) {
    const field = issue.path[0] as PublishFormField;
    errors[field] ??= issue.message;
  }

  return errors;
}

/** Só os erros dos campos que a pessoa já deixou — ou todos, depois de tentar enviar. */
function visibleErrors(
  errors: Partial<Record<PublishFormField, string>>,
  touched: ReadonlySet<PublishFormField>,
  isSubmitted: boolean,
): Partial<Record<PublishFormField, string>> {
  if (isSubmitted) return errors;

  return Object.fromEntries(
    Object.entries(errors).filter(([field]) => touched.has(field as PublishFormField)),
  );
}

/**
 * PUBLICAR — o diálogo de cada aba (Documentos, Planejamentos, Ideias) e o da versão nova, no
 * detalhe.
 *
 * Com `kind`, é artefato NOVO daquele tipo: a label nasce do título enquanto a pessoa não
 * mexer nela, e publicado ele abre. Com `label`, é VERSÃO NOVA daquele artefato: a label fica
 * travada, só o arquivo é obrigatório, e o detalhe se atualiza sozinho (todas as consultas de
 * artefato são invalidadas).
 */
export function useArtifactPublishModel(
  input: { kind: ArtifactKind; label?: never } | { label: string; kind?: never },
): ArtifactPublishModel {
  const queryClient = useQueryClient();
  const kind = input.kind;
  const mode: PublishDialogMode = kind ? 'new' : 'version';
  const copy = kind ? PUBLISH_COPY[kind] : VERSION_COPY;

  let values = $state<PublishFormValues>(emptyValues(input.label, kind));
  let isOpen = $state(false);
  let isSubmitted = $state(false);
  let isLabelEdited = $state(false);
  const touched = new SvelteSet<PublishFormField>();

  const publish = mirrorStore(
    createMutation({
      mutationFn: (parsed: ArtifactVersionPublishInput) => artifactsApi.publish(parsed),
      onSuccess: async (artifact) => {
        await queryClient.invalidateQueries({ queryKey: ['artifacts'] });
        isOpen = false;
        if (mode === 'new') await goto(`/artifacts/${artifact.label}`);
      },
    }),
  );

  function reset(): void {
    values = emptyValues(input.label, kind);
    isSubmitted = false;
    isLabelEdited = false;
    touched.clear();
    publish.current.reset();
  }

  return {
    get data() {
      return { values, accept: PUBLISH_ACCEPT, copy };
    },
    ui: { mode, hasStatus: mode === 'new' && hasStatus(kind) },
    get state() {
      return {
        isOpen,
        isSubmitting: publish.current.isPending,
        error: readError(publish.current.error),
        fieldErrors: visibleErrors(publishErrorsOf(values, kind), touched, isSubmitted),
      };
    },
    actions: {
      onOpen: () => {
        reset();
        isOpen = true;
      },
      onClose: () => {
        if (publish.current.isPending) return;
        isOpen = false;
      },
      onChange: (field, value) => {
        /* O acesso só tem dois valores; qualquer outro texto vira o padrão, privado. */
        if (field === 'visibility') {
          values.visibility = value === 'public' ? 'public' : 'private';
          return;
        }
        values[field] = value;
        if (field === 'label') isLabelEdited = true;
        if (field === 'title' && mode === 'new' && !isLabelEdited) values.label = slugify(value);
      },
      onFileChange: (file) => {
        values.file = file;
        touched.add('file');
      },
      onBlur: (field) => {
        touched.add(field);
      },
      onSubmit: () => {
        isSubmitted = true;
        const schema = kind ? artifactPublishSchema : artifactVersionPublishSchema;
        const parsed = schema.safeParse(toInput(values, kind));
        if (!parsed.success || publish.current.isPending) return;

        publish.current.mutate(parsed.data);
      },
    },
  };
}

import { goto } from '$app/navigation';
import { createMutation, useQueryClient } from '@tanstack/svelte-query';
import { SvelteSet } from 'svelte/reactivity';

import { artifactsApi } from '$lib/api/artifacts.api';
import { readError } from '$lib/api/http-client';
import { type OptionPickerOption } from '$lib/components/acerola-option-picker/acerola-option-picker.svelte';
import {
  ARTIFACT_KIND_LABELS,
  artifactKindSchema,
  artifactPublishSchema,
  artifactVersionPublishSchema,
  PUBLISH_ACCEPT,
  type ArtifactKind,
  type ArtifactPublishInput,
  type ArtifactVersionPublishInput,
} from '$lib/contracts/artifact.schema';
import { mirrorStore } from '$lib/hooks/use-mirror-store/use-mirror-store.svelte';
import { slugify } from '$lib/utils/slug.util';

export type PublishField = keyof ArtifactPublishInput;

/** O formulário como a pessoa preenche: o arquivo pode ainda não ter sido escolhido. */
export type PublishValues = Omit<ArtifactPublishInput, 'file'> & { file: File | null };

export type PublishMode = 'new' | 'version';

export type ArtifactPublishModel = {
  data: { values: PublishValues; kindOptions: OptionPickerOption[]; accept: string };
  ui: { mode: PublishMode };
  state: {
    isOpen: boolean;
    isSubmitting: boolean;
    error: string | null;
    fieldErrors: Partial<Record<PublishField, string>>;
  };
  actions: {
    onOpen: () => void;
    onClose: () => void;
    onChange: (field: Exclude<PublishField, 'file'>, value: string) => void;
    onFileChange: (file: File | null) => void;
    onBlur: (field: PublishField) => void;
    onSubmit: () => void;
  };
};

function emptyValues(input: { kind?: ArtifactKind; label?: string }): PublishValues {
  return {
    file: null,
    title: '',
    label: input.label ?? '',
    kind: input.kind ?? 'document',
    description: '',
    tags: '',
  };
}

function kindOptions(): OptionPickerOption[] {
  return artifactKindSchema.options.map((kind) => ({
    value: kind,
    label: ARTIFACT_KIND_LABELS[kind],
  }));
}

/** O primeiro erro de cada campo, no formato que os campos mostram embaixo deles. */
export function publishErrorsOf(
  values: PublishValues,
  mode: PublishMode,
): Partial<Record<PublishField, string>> {
  const schema = mode === 'new' ? artifactPublishSchema : artifactVersionPublishSchema;
  const result = schema.safeParse(values);
  if (result.success) return {};

  const errors: Partial<Record<PublishField, string>> = {};
  for (const issue of result.error.issues) {
    const field = issue.path[0] as PublishField;
    errors[field] ??= issue.message;
  }

  return errors;
}

/** Só os erros dos campos que a pessoa já deixou — ou todos, depois de tentar enviar. */
function visibleErrors(
  errors: Partial<Record<PublishField, string>>,
  touched: ReadonlySet<PublishField>,
  isSubmitted: boolean,
): Partial<Record<PublishField, string>> {
  if (isSubmitted) return errors;

  return Object.fromEntries(
    Object.entries(errors).filter(([field]) => touched.has(field as PublishField)),
  );
}

/**
 * PUBLICAR UM ARTEFATO — o diálogo da galeria (artefato novo) e do detalhe (versão nova).
 *
 * Com `label`, é versão nova daquele artefato: a label fica travada e só o arquivo é
 * obrigatório (título, descrição e tags vazios mantêm o publicado). Sem `label`, é artefato
 * novo: a label nasce do título enquanto a pessoa não mexer nela.
 *
 * Publicou o novo, a tela abre o artefato; publicou a versão, o detalhe se atualiza sozinho
 * (todas as consultas de artefato são invalidadas).
 */
export function useArtifactPublishModel(
  input: { kind?: ArtifactKind; label?: string } = {},
): ArtifactPublishModel {
  const queryClient = useQueryClient();
  const mode: PublishMode = input.label ? 'version' : 'new';

  let values = $state<PublishValues>(emptyValues(input));
  let isOpen = $state(false);
  let isSubmitted = $state(false);
  let isLabelEdited = $state(false);
  const touched = new SvelteSet<PublishField>();

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
    values = emptyValues(input);
    isSubmitted = false;
    isLabelEdited = false;
    touched.clear();
    publish.current.reset();
  }

  return {
    get data() {
      return { values, kindOptions: kindOptions(), accept: PUBLISH_ACCEPT };
    },
    ui: { mode },
    get state() {
      return {
        isOpen,
        isSubmitting: publish.current.isPending,
        error: readError(publish.current.error),
        fieldErrors: visibleErrors(publishErrorsOf(values, mode), touched, isSubmitted),
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
        if (field === 'kind') {
          const kind = artifactKindSchema.safeParse(value);
          if (kind.success) values.kind = kind.data;
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
        const schema = mode === 'new' ? artifactPublishSchema : artifactVersionPublishSchema;
        const parsed = schema.safeParse(values);
        if (!parsed.success || publish.current.isPending) return;

        /* Na versão nova o tipo não vai: mandado, ele trocaria o tipo do artefato. */
        publish.current.mutate(mode === 'new' ? parsed.data : { ...parsed.data, kind: undefined });
      },
    },
  };
}

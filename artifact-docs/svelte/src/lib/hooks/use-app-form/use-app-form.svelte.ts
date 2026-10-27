import { createMutation, createQuery, useQueryClient } from '@tanstack/svelte-query';
import { SvelteSet } from 'svelte/reactivity';

import { appsApi } from '$lib/api/apps.api';
import { readError } from '$lib/api/http-client';
import { appLinkInputSchema, type AppHub, type AppLinkInput } from '$lib/contracts/app-link.schema';
import { mirrorStore } from '$lib/hooks/use-mirror-store/use-mirror-store.svelte';

export type AppFormField = keyof AppLinkInput;

export type AppFormMode = 'new' | 'edit';

export type AppFormModel = {
  data: { values: AppLinkInput; typeSuggestions: string[] };
  ui: { mode: AppFormMode };
  state: {
    isOpen: boolean;
    isSubmitting: boolean;
    error: string | null;
    fieldErrors: Partial<Record<AppFormField, string>>;
    isDeleteOpen: boolean;
    isDeleting: boolean;
    deleteError: string | null;
  };
  actions: {
    onCreate: () => void;
    onEdit: (id: string) => void;
    onClose: () => void;
    onChange: (field: AppFormField, value: string) => void;
    onBlur: (field: AppFormField) => void;
    onSubmit: () => void;
    onDeleteRequest: () => void;
    onDeleteCancel: () => void;
    onDeleteConfirm: () => void;
  };
};

const EMPTY: AppLinkInput = { name: '', url: '', typeName: '' };

/** O primeiro erro de cada campo, no formato que os campos mostram embaixo deles. */
export function appFormErrorsOf(values: AppLinkInput): Partial<Record<AppFormField, string>> {
  const result = appLinkInputSchema.safeParse(values);
  if (result.success) return {};

  const errors: Partial<Record<AppFormField, string>> = {};
  for (const issue of result.error.issues) {
    const field = issue.path[0] as AppFormField;
    errors[field] ??= issue.message;
  }

  return errors;
}

/** Os valores do formulário para editar um app que já existe — o tipo vai pelo nome. */
export function valuesOf(hub: AppHub, id: string): AppLinkInput | null {
  const app = hub.apps.find((item) => item.id === id);
  if (!app) return null;
  const type = hub.types.find((item) => item.id === app.typeId);

  return { name: app.name, url: app.url, typeName: type?.name ?? '' };
}

/** Os tipos que já existem, na ordem do hub — os atalhos embaixo do campo de tipo. */
export function typeSuggestionsOf(hub: AppHub | undefined): string[] {
  return [...(hub?.types ?? [])].sort((a, b) => a.position - b.position).map((type) => type.name);
}

/**
 * CADASTRAR, EDITAR E APAGAR APP — o diálogo da tela de Apps.
 *
 * O tipo é texto livre com os tipos que já existem como atalho (lei de Tesler): digitar um
 * tipo novo cria o tipo, e o que fica sem app some. Não existe tela para gerenciar tipos.
 *
 * Apagar mora dentro da edição e passa por confirmação (ui-standards §10).
 */
export function useAppFormModel(): AppFormModel {
  const queryClient = useQueryClient();
  const hub = mirrorStore(createQuery({ queryKey: ['apps'], queryFn: () => appsApi.hub() }));

  let values = $state<AppLinkInput>({ ...EMPTY });
  let editingId = $state<string | null>(null);
  let isOpen = $state(false);
  let isDeleteOpen = $state(false);
  let isSubmitted = $state(false);
  const touched = new SvelteSet<AppFormField>();

  const refresh = () => queryClient.invalidateQueries({ queryKey: ['apps'] });

  const save = mirrorStore(
    createMutation({
      mutationFn: (input: AppLinkInput) =>
        editingId ? appsApi.update(editingId, input) : appsApi.create(input),
      onSuccess: async () => {
        await refresh();
        isOpen = false;
      },
    }),
  );
  const removal = mirrorStore(
    createMutation({
      mutationFn: (id: string) => appsApi.remove(id),
      onSuccess: async () => {
        await refresh();
        isDeleteOpen = false;
        isOpen = false;
      },
    }),
  );

  function open(id: string | null, initial: AppLinkInput): void {
    values = { ...initial };
    editingId = id;
    isSubmitted = false;
    touched.clear();
    save.current.reset();
    isOpen = true;
  }

  function visibleErrors(): Partial<Record<AppFormField, string>> {
    const errors = appFormErrorsOf(values);
    if (isSubmitted) return errors;

    return Object.fromEntries(
      Object.entries(errors).filter(([field]) => touched.has(field as AppFormField)),
    );
  }

  return {
    get data() {
      return { values, typeSuggestions: typeSuggestionsOf(hub.current.data) };
    },
    get ui() {
      return { mode: editingId ? 'edit' : 'new' } as const;
    },
    get state() {
      return {
        isOpen,
        isSubmitting: save.current.isPending,
        error: readError(save.current.error),
        fieldErrors: visibleErrors(),
        isDeleteOpen,
        isDeleting: removal.current.isPending,
        deleteError: readError(removal.current.error),
      };
    },
    actions: {
      onCreate: () => open(null, EMPTY),
      onEdit: (id) => {
        const initial = hub.current.data ? valuesOf(hub.current.data, id) : null;
        if (initial) open(id, initial);
      },
      onClose: () => {
        if (save.current.isPending) return;
        isOpen = false;
      },
      onChange: (field, value) => {
        values[field] = value;
      },
      onBlur: (field) => {
        touched.add(field);
      },
      onSubmit: () => {
        isSubmitted = true;
        const parsed = appLinkInputSchema.safeParse(values);
        if (!parsed.success || save.current.isPending) return;

        save.current.mutate(parsed.data);
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
        if (!editingId || removal.current.isPending) return;
        removal.current.mutate(editingId);
      },
    },
  };
}

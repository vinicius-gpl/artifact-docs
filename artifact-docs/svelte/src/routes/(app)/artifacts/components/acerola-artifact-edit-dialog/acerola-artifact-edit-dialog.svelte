<script lang="ts" module>
  import type {
    ArtifactEditField,
    ArtifactEditValues,
  } from '$lib/hooks/use-artifact-edit/use-artifact-edit.svelte';

  /**
   * EDITAR um artefato já publicado: o que aparece no card e no detalhe, sem subir arquivo —
   * título, descrição, tags, situação e acesso. O arquivo muda pelo "Publicar nova versão".
   *
   * A label não aparece: ela é o endereço, e trocar quebraria todo link já mandado. A situação
   * só existe em planejamento e ideia (`ui.hasStatus`).
   */
  export type AcerolaArtifactEditDialogProps = {
    data: { values: ArtifactEditValues };
    ui: { hasStatus: boolean };
    state: {
      isOpen: boolean;
      isSubmitting: boolean;
      error: string | null;
      fieldErrors: Partial<Record<ArtifactEditField, string>>;
    };
    actions: {
      onClose: () => void;
      onChange: (field: ArtifactEditField, value: string) => void;
      onBlur: (field: ArtifactEditField) => void;
      onSubmit: () => void;
    };
  };
</script>

<script lang="ts">
  import ActionButton from '$lib/components/acerola-action-button/acerola-action-button.svelte';
  import ErrorState from '$lib/components/acerola-error-state/acerola-error-state.svelte';
  import FilterField from '$lib/components/acerola-filter-field/acerola-filter-field.svelte';
  import OptionPicker from '$lib/components/acerola-option-picker/acerola-option-picker.svelte';
  import ResponsiveDialogContent from '$lib/components/acerola-responsive-dialog-content/acerola-responsive-dialog-content.svelte';
  import SubmitButton from '$lib/components/acerola-submit-button/acerola-submit-button.svelte';
  import TextAreaField from '$lib/components/acerola-text-area-field/acerola-text-area-field.svelte';
  import TextField from '$lib/components/acerola-text-field/acerola-text-field.svelte';
  import { Dialog, DialogFooter, DialogHeader, DialogTitle } from '$lib/components/acerola-dialog/acerola-dialog';
  import { ARTIFACT_STATUS_LABELS, type ArtifactStatus } from '$lib/contracts/artifact.schema';

  /* `dialogState`, e não `state`: com a prop chamada `state`, a runa `$state` ficaria escondida. */
  let { data, ui, state: dialogState, actions }: AcerolaArtifactEditDialogProps = $props();

  const errors = $derived(dialogState.fieldErrors);

  const STATUS_OPTIONS = (Object.keys(ARTIFACT_STATUS_LABELS) as ArtifactStatus[]).map(
    (status) => ({ value: status, label: ARTIFACT_STATUS_LABELS[status] }),
  );
  const VISIBILITY_OPTIONS = [
    { value: 'private', label: 'Privado' },
    { value: 'public', label: 'Com o link' },
  ];
</script>

<Dialog
  open={dialogState.isOpen}
  onOpenChange={(isOpen: boolean) => (isOpen ? undefined : actions.onClose())}
>
  <ResponsiveDialogContent class="sm:max-h-[90dvh] sm:max-w-xl">
    <DialogHeader>
      <DialogTitle class="text-lg font-semibold tracking-tight">Editar</DialogTitle>
    </DialogHeader>

    <form
      class="flex flex-col gap-4 px-1 pb-1"
      novalidate
      onsubmit={(event) => {
        event.preventDefault();
        actions.onSubmit();
      }}
    >
      <TextField
        data={{
          label: 'Título',
          name: 'title',
          value: data.values.title,
          autoComplete: 'off',
          isRequired: true,
        }}
        state={{ error: errors.title, isDisabled: dialogState.isSubmitting }}
        actions={{
          onChange: (value) => actions.onChange('title', value),
          onBlur: () => actions.onBlur('title'),
        }}
      />

      <div class="grid gap-4 sm:grid-cols-2">
        {#if ui.hasStatus}
          <FilterField data={{ label: 'Situação' }}>
            <OptionPicker
              data={{ value: data.values.status, options: STATUS_OPTIONS }}
              ui={{ ariaLabel: 'Situação', fullWidth: true }}
              state={{ isDisabled: dialogState.isSubmitting }}
              actions={{ onChange: (value) => actions.onChange('status', value) }}
            />
          </FilterField>
        {/if}
        <FilterField data={{ label: 'Acesso' }}>
          <OptionPicker
            data={{ value: data.values.visibility, options: VISIBILITY_OPTIONS }}
            ui={{ ariaLabel: 'Acesso', fullWidth: true }}
            state={{ isDisabled: dialogState.isSubmitting }}
            actions={{ onChange: (value) => actions.onChange('visibility', value) }}
          />
        </FilterField>
      </div>

      <TextAreaField
        data={{
          label: 'Descrição',
          name: 'description',
          value: data.values.description,
          maxLength: 1000,
        }}
        ui={{ rows: 2 }}
        state={{ error: errors.description, isDisabled: dialogState.isSubmitting }}
        actions={{
          onChange: (value) => actions.onChange('description', value),
          onBlur: () => actions.onBlur('description'),
        }}
      />

      <TextField
        data={{
          label: 'Tags',
          name: 'tags',
          value: data.values.tags,
          placeholder: 'hub, marca',
          autoComplete: 'off',
        }}
        state={{ error: errors.tags, isDisabled: dialogState.isSubmitting }}
        actions={{
          onChange: (value) => actions.onChange('tags', value),
          onBlur: () => actions.onBlur('tags'),
        }}
      />

      {#if dialogState.error}
        <ErrorState data={{ message: dialogState.error }} ui={{ variant: 'inline' }} />
      {/if}

      <DialogFooter>
        <ActionButton
          data={{ label: 'Cancelar' }}
          ui={{ variant: 'secondary' }}
          state={{ isDisabled: dialogState.isSubmitting }}
          actions={{ onClick: actions.onClose }}
        />
        <SubmitButton
          data={{ label: 'Salvar', loadingLabel: 'Salvando…' }}
          state={{ isLoading: dialogState.isSubmitting }}
        />
      </DialogFooter>
    </form>
  </ResponsiveDialogContent>
</Dialog>

<script lang="ts" module>
  import { type OptionPickerOption } from '$lib/components/acerola-option-picker/acerola-option-picker.svelte';
  import {
    type PublishField,
    type PublishMode,
    type PublishValues,
  } from '$lib/hooks/use-artifact-publish/use-artifact-publish.svelte';

  /**
   * Publicar um artefato: o arquivo e os dados dele, num diálogo (cadastro curto, ui-standards
   * §8). O mesmo diálogo publica o artefato NOVO (galeria) e a VERSÃO NOVA (detalhe) — na
   * versão, a label fica travada e o título vazio mantém o publicado.
   *
   * Seis campos, em três blocos que a pessoa reconhece (Miller): o arquivo; título e label lado
   * a lado (a label nasce do título); tipo, descrição e tags.
   */
  export type AcerolaArtifactPublishDialogProps = {
    data: { values: PublishValues; kindOptions: OptionPickerOption[]; accept: string };
    ui: { mode: PublishMode };
    state: {
      isOpen: boolean;
      isSubmitting: boolean;
      error: string | null;
      fieldErrors: Partial<Record<PublishField, string>>;
    };
    actions: {
      onClose: () => void;
      onChange: (field: Exclude<PublishField, 'file'>, value: string) => void;
      onFileChange: (file: File | null) => void;
      onBlur: (field: PublishField) => void;
      onSubmit: () => void;
    };
  };

  const COPY: Record<PublishMode, { title: string; description: string; submit: string }> = {
    new: {
      title: 'Publicar artefato',
      description: 'O arquivo vira uma página servida pelo hub, com preview na galeria.',
      submit: 'Publicar',
    },
    version: {
      title: 'Publicar nova versão',
      description: 'A versão nova entra no ar no lugar da atual. As anteriores continuam guardadas.',
      submit: 'Publicar versão',
    },
  };
</script>

<script lang="ts">
  import ActionButton from '$lib/components/acerola-action-button/acerola-action-button.svelte';
  import {
    Dialog,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
  } from '$lib/components/acerola-dialog/acerola-dialog';
  import ErrorState from '$lib/components/acerola-error-state/acerola-error-state.svelte';
  import FileField from '$lib/components/acerola-file-field/acerola-file-field.svelte';
  import FilterField from '$lib/components/acerola-filter-field/acerola-filter-field.svelte';
  import OptionPicker from '$lib/components/acerola-option-picker/acerola-option-picker.svelte';
  import ResponsiveDialogContent from '$lib/components/acerola-responsive-dialog-content/acerola-responsive-dialog-content.svelte';
  import SubmitButton from '$lib/components/acerola-submit-button/acerola-submit-button.svelte';
  import TextAreaField from '$lib/components/acerola-text-area-field/acerola-text-area-field.svelte';
  import TextField from '$lib/components/acerola-text-field/acerola-text-field.svelte';

  let { data, ui, state, actions }: AcerolaArtifactPublishDialogProps = $props();

  const copy = $derived(COPY[ui.mode]);
  const isNew = $derived(ui.mode === 'new');
  const errors = $derived(state.fieldErrors);
</script>

<Dialog open={state.isOpen} onOpenChange={(isOpen: boolean) => (isOpen ? undefined : actions.onClose())}>
  <ResponsiveDialogContent class="sm:max-w-xl">
    <DialogHeader>
      <DialogTitle class="text-lg font-semibold tracking-tight">{copy.title}</DialogTitle>
      <DialogDescription class="text-muted-foreground text-xs">{copy.description}</DialogDescription>
    </DialogHeader>

    <form
      class="flex flex-col gap-4"
      novalidate
      onsubmit={(event) => {
        event.preventDefault();
        actions.onSubmit();
      }}
    >
      <FileField
        data={{
          label: 'Arquivo',
          name: 'file',
          accept: data.accept,
          fileName: data.values.file?.name ?? null,
          hint: 'HTML, site em .zip (com index.html), Markdown ou PDF — até 100 MB.',
          isRequired: true,
        }}
        state={{ error: errors.file, isDisabled: state.isSubmitting }}
        actions={{ onChange: actions.onFileChange, onBlur: () => actions.onBlur('file') }}
      />

      <div class="grid gap-4 sm:grid-cols-2">
        <TextField
          data={{
            label: 'Título',
            name: 'title',
            value: data.values.title,
            placeholder: isNew ? 'Guia de instalação' : 'Manter o atual',
            autoComplete: 'off',
            isRequired: isNew,
          }}
          state={{ error: errors.title, isDisabled: state.isSubmitting }}
          actions={{ onChange: (value) => actions.onChange('title', value), onBlur: () => actions.onBlur('title') }}
        />
        <TextField
          data={{
            label: 'Label (endereço)',
            name: 'label',
            value: data.values.label,
            placeholder: 'guia-de-instalacao',
            autoComplete: 'off',
            isRequired: isNew,
          }}
          state={{ error: errors.label, isDisabled: state.isSubmitting || !isNew }}
          actions={{ onChange: (value) => actions.onChange('label', value), onBlur: () => actions.onBlur('label') }}
        />
      </div>

      {#if isNew}
        <FilterField data={{ label: 'Tipo' }}>
          <OptionPicker
            data={{ value: data.values.kind, options: data.kindOptions }}
            ui={{ ariaLabel: 'Tipo', fullWidth: true }}
            state={{ isDisabled: state.isSubmitting }}
            actions={{ onChange: (value) => actions.onChange('kind', value) }}
          />
        </FilterField>
      {/if}

      <TextAreaField
        data={{
          label: 'Descrição',
          name: 'description',
          value: data.values.description,
          placeholder: isNew ? 'Uma linha sobre o que é' : 'Manter a atual',
          maxLength: 1000,
        }}
        ui={{ rows: 2 }}
        state={{ error: errors.description, isDisabled: state.isSubmitting }}
        actions={{ onChange: (value) => actions.onChange('description', value), onBlur: () => actions.onBlur('description') }}
      />

      <TextField
        data={{
          label: 'Tags',
          name: 'tags',
          value: data.values.tags,
          placeholder: isNew ? 'hub, marca' : 'Manter as atuais',
          autoComplete: 'off',
        }}
        state={{ error: errors.tags, isDisabled: state.isSubmitting }}
        actions={{ onChange: (value) => actions.onChange('tags', value), onBlur: () => actions.onBlur('tags') }}
      />

      {#if state.error}
        <ErrorState data={{ message: state.error }} ui={{ variant: 'inline' }} />
      {/if}

      <DialogFooter>
        <ActionButton
          data={{ label: 'Cancelar' }}
          ui={{ variant: 'secondary' }}
          state={{ isDisabled: state.isSubmitting }}
          actions={{ onClick: actions.onClose }}
        />
        <SubmitButton data={{ label: copy.submit, loadingLabel: 'Publicando…' }} state={{ isLoading: state.isSubmitting }} />
      </DialogFooter>
    </form>
  </ResponsiveDialogContent>
</Dialog>

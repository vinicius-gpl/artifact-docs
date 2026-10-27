<script lang="ts" module>
  /**
   * Publicar um arquivo como página: o arquivo e os dados dele, num diálogo (cadastro curto,
   * ui-standards §8). Serve às três abas de publicação — Documentos, Planejamentos e Ideias —,
   * e cada uma diz o que se publica pelo `data.copy`; o tipo vem da aba, e não de um campo.
   *
   * Na VERSÃO NOVA (`ui.mode = 'version'`) a label fica travada e o título vazio mantém o
   * publicado.
   *
   * Cinco campos, em três blocos que a pessoa reconhece (Miller): o arquivo; título e label lado
   * a lado (a label nasce do título); descrição e tags.
   *
   * O conteúdo rola por dentro e tem altura máxima também no desktop: numa tela baixa, o
   * diálogo cortado escondia o botão de publicar. O `px-1` do formulário é o espaço da borda
   * de foco: sem ele, a rolagem corta a borda dos campos que encostam na lateral. Sem margem
   * negativa junto: o formulário ficaria mais largo que a caixa, e ela rolaria para o lado.
   */
  export type PublishFormValues = {
    file: File | null;
    title: string;
    label: string;
    description: string;
    tags: string;
  };

  export type PublishFormField = keyof PublishFormValues;

  export type PublishDialogMode = 'new' | 'version';

  export type AcerolaPublishDialogProps = {
    data: {
      values: PublishFormValues;
      accept: string;
      copy: { title: string; description: string; submitLabel: string; titlePlaceholder: string };
    };
    ui: { mode: PublishDialogMode };
    state: {
      isOpen: boolean;
      isSubmitting: boolean;
      error: string | null;
      fieldErrors: Partial<Record<PublishFormField, string>>;
    };
    actions: {
      onClose: () => void;
      onChange: (field: Exclude<PublishFormField, 'file'>, value: string) => void;
      onFileChange: (file: File | null) => void;
      onBlur: (field: PublishFormField) => void;
      onSubmit: () => void;
    };
  };
</script>

<script lang="ts">
  import ActionButton from '$lib/components/acerola-action-button/acerola-action-button.svelte';
  import ErrorState from '$lib/components/acerola-error-state/acerola-error-state.svelte';
  import FileField from '$lib/components/acerola-file-field/acerola-file-field.svelte';
  import ResponsiveDialogContent from '$lib/components/acerola-responsive-dialog-content/acerola-responsive-dialog-content.svelte';
  import SubmitButton from '$lib/components/acerola-submit-button/acerola-submit-button.svelte';
  import TextAreaField from '$lib/components/acerola-text-area-field/acerola-text-area-field.svelte';
  import TextField from '$lib/components/acerola-text-field/acerola-text-field.svelte';
  import {
    Dialog,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
  } from '$lib/components/ui/dialog';

  let { data, ui, state, actions }: AcerolaPublishDialogProps = $props();

  const isNew = $derived(ui.mode === 'new');
  const errors = $derived(state.fieldErrors);
</script>

<Dialog
  open={state.isOpen}
  onOpenChange={(isOpen: boolean) => (isOpen ? undefined : actions.onClose())}
>
  <ResponsiveDialogContent class="sm:max-h-[90dvh] sm:max-w-xl">
    <DialogHeader>
      <DialogTitle class="text-lg font-semibold tracking-tight">{data.copy.title}</DialogTitle>
      <DialogDescription class="text-muted-foreground text-xs">
        {data.copy.description}
      </DialogDescription>
    </DialogHeader>

    <form
      class="flex flex-col gap-4 px-1 pb-1"
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
            placeholder: isNew ? data.copy.titlePlaceholder : 'Manter o atual',
            autoComplete: 'off',
            isRequired: isNew,
          }}
          state={{ error: errors.title, isDisabled: state.isSubmitting }}
          actions={{
            onChange: (value) => actions.onChange('title', value),
            onBlur: () => actions.onBlur('title'),
          }}
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
          actions={{
            onChange: (value) => actions.onChange('label', value),
            onBlur: () => actions.onBlur('label'),
          }}
        />
      </div>

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
          placeholder: isNew ? 'hub, marca' : 'Manter as atuais',
          autoComplete: 'off',
        }}
        state={{ error: errors.tags, isDisabled: state.isSubmitting }}
        actions={{
          onChange: (value) => actions.onChange('tags', value),
          onBlur: () => actions.onBlur('tags'),
        }}
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
        <SubmitButton
          data={{ label: data.copy.submitLabel, loadingLabel: 'Publicando…' }}
          state={{ isLoading: state.isSubmitting }}
        />
      </DialogFooter>
    </form>
  </ResponsiveDialogContent>
</Dialog>

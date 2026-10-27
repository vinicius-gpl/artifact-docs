<script lang="ts" module>
  import { type AppLinkInput } from '$lib/contracts/app-link.schema';
  import { type AppFormField, type AppFormMode } from '$lib/hooks/use-app-form/use-app-form.svelte';

  /**
   * Cadastrar ou editar um app do hub: nome, endereço e tipo — três campos, num diálogo
   * (cadastro curto, ui-standards §8).
   *
   * O tipo é texto livre, com os tipos que já existem logo embaixo como atalho (Tesler): um
   * clique preenche, e digitar um nome novo cria o tipo. Na edição, "Apagar app" fica no
   * rodapé, longe do botão de salvar, e só pede a confirmação — quem apaga é a tela.
   */
  export type AcerolaAppDialogProps = {
    data: { values: AppLinkInput; typeSuggestions: string[] };
    ui: { mode: AppFormMode };
    state: {
      isOpen: boolean;
      isSubmitting: boolean;
      error: string | null;
      fieldErrors: Partial<Record<AppFormField, string>>;
    };
    actions: {
      onClose: () => void;
      onChange: (field: AppFormField, value: string) => void;
      onBlur: (field: AppFormField) => void;
      onSubmit: () => void;
      onDelete?: () => void;
    };
  };
</script>

<script lang="ts">
  import Trash2 from '@lucide/svelte/icons/trash-2';

  import ActionButton from '$lib/components/acerola-action-button/acerola-action-button.svelte';
  import {
    Dialog,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
  } from '$lib/components/acerola-dialog/acerola-dialog';
  import ErrorState from '$lib/components/acerola-error-state/acerola-error-state.svelte';
  import ResponsiveDialogContent from '$lib/components/acerola-responsive-dialog-content/acerola-responsive-dialog-content.svelte';
  import SubmitButton from '$lib/components/acerola-submit-button/acerola-submit-button.svelte';
  import TextField from '$lib/components/acerola-text-field/acerola-text-field.svelte';

  let { data, ui, state, actions }: AcerolaAppDialogProps = $props();

  const isNew = $derived(ui.mode === 'new');
  const errors = $derived(state.fieldErrors);
</script>

<Dialog
  open={state.isOpen}
  onOpenChange={(isOpen: boolean) => (isOpen ? undefined : actions.onClose())}
>
  <ResponsiveDialogContent class="sm:max-h-[90dvh]">
    <DialogHeader>
      <DialogTitle class="text-lg font-semibold tracking-tight">
        {isNew ? 'Cadastrar app' : 'Editar app'}
      </DialogTitle>
      <DialogDescription class="text-muted-foreground text-xs">
        O atalho aparece no hub, no grupo do tipo, e o hub confere de tempos em tempos se ele
        está no ar.
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
      <TextField
        data={{
          label: 'Nome',
          name: 'name',
          value: data.values.name,
          placeholder: 'Jellyfin',
          autoComplete: 'off',
          isRequired: true,
        }}
        state={{ error: errors.name, isDisabled: state.isSubmitting, isAutoFocused: isNew }}
        actions={{
          onChange: (value) => actions.onChange('name', value),
          onBlur: () => actions.onBlur('name'),
        }}
      />

      <TextField
        data={{
          label: 'Endereço',
          name: 'url',
          value: data.values.url,
          placeholder: 'https://jellyfin.exemplo.com',
          autoComplete: 'url',
          isRequired: true,
        }}
        ui={{ type: 'url' }}
        state={{ error: errors.url, isDisabled: state.isSubmitting }}
        actions={{
          onChange: (value) => actions.onChange('url', value),
          onBlur: () => actions.onBlur('url'),
        }}
      />

      <div class="flex flex-col gap-2">
        <TextField
          data={{
            label: 'Tipo',
            name: 'typeName',
            value: data.values.typeName,
            placeholder: 'Self-hosted',
            autoComplete: 'off',
            isRequired: true,
          }}
          state={{ error: errors.typeName, isDisabled: state.isSubmitting }}
          actions={{
            onChange: (value) => actions.onChange('typeName', value),
            onBlur: () => actions.onBlur('typeName'),
          }}
        />
        {#if data.typeSuggestions.length > 0}
          <div class="flex flex-wrap gap-1.5" role="group" aria-label="Tipos já cadastrados">
            {#each data.typeSuggestions as suggestion (suggestion)}
              <ActionButton
                data={{ label: suggestion }}
                ui={{ variant: 'secondary', size: 'sm' }}
                state={{
                  isPressed: suggestion === data.values.typeName,
                  isDisabled: state.isSubmitting,
                }}
                actions={{ onClick: () => actions.onChange('typeName', suggestion) }}
              />
            {/each}
          </div>
        {/if}
      </div>

      {#if state.error}
        <ErrorState data={{ message: state.error }} ui={{ variant: 'inline' }} />
      {/if}

      <DialogFooter class="sm:items-center">
        {#if !isNew && actions.onDelete}
          <ActionButton
            data={{ label: 'Apagar app' }}
            ui={{ variant: 'ghost', icon: Trash2, className: 'sm:mr-auto' }}
            state={{ isDisabled: state.isSubmitting }}
            actions={{ onClick: actions.onDelete }}
          />
        {/if}
        <ActionButton
          data={{ label: 'Cancelar' }}
          ui={{ variant: 'secondary' }}
          state={{ isDisabled: state.isSubmitting }}
          actions={{ onClick: actions.onClose }}
        />
        <SubmitButton
          data={{ label: isNew ? 'Cadastrar app' : 'Salvar', loadingLabel: 'Salvando…' }}
          state={{ isLoading: state.isSubmitting }}
        />
      </DialogFooter>
    </form>
  </ResponsiveDialogContent>
</Dialog>

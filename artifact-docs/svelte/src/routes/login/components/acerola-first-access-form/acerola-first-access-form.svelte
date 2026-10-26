<script lang="ts" module>
  import type { FirstAccessField } from '$lib/hooks/use-first-access/use-first-access.svelte';

  /**
   * PRIMEIRO ACESSO: cria a conta do dono do hub — nome, e-mail, senha e a confirmação.
   *
   * Só aparece enquanto não existe conta (a tela de login decide). A confirmação existe porque a
   * pessoa não vê o que digita; o erro de "senhas diferentes" fica no campo da confirmação.
   */
  export type AcerolaFirstAccessFormProps = {
    data: {
      values: { name: string; email: string; password: string; passwordConfirmation: string };
    };
    state: {
      isSubmitting: boolean;
      error: string | null;
      fieldErrors: Partial<Record<FirstAccessField, string>>;
    };
    actions: {
      onChange: (field: FirstAccessField, value: string) => void;
      onSubmit: () => void;
    };
  };
</script>

<script lang="ts">
  import ErrorState from '$lib/components/acerola-error-state/acerola-error-state.svelte';
  import SubmitButton from '$lib/components/acerola-submit-button/acerola-submit-button.svelte';
  import TextField from '$lib/components/acerola-text-field/acerola-text-field.svelte';

  let { data, state: formState, actions }: AcerolaFirstAccessFormProps = $props();

  function submit(event: SubmitEvent) {
    event.preventDefault();
    actions.onSubmit();
  }
</script>

<form class="flex flex-col gap-4" novalidate onsubmit={submit}>
  <TextField
    data={{ label: 'Nome', name: 'name', value: data.values.name, autoComplete: 'name' }}
    state={{ error: formState.fieldErrors.name, isDisabled: formState.isSubmitting }}
    actions={{ onChange: (value) => actions.onChange('name', value) }}
  />
  <TextField
    data={{ label: 'E-mail', name: 'email', value: data.values.email, autoComplete: 'email' }}
    ui={{ type: 'email' }}
    state={{ error: formState.fieldErrors.email, isDisabled: formState.isSubmitting }}
    actions={{ onChange: (value) => actions.onChange('email', value) }}
  />
  <TextField
    data={{
      label: 'Senha',
      name: 'password',
      value: data.values.password,
      autoComplete: 'new-password',
    }}
    ui={{ type: 'password' }}
    state={{ error: formState.fieldErrors.password, isDisabled: formState.isSubmitting }}
    actions={{ onChange: (value) => actions.onChange('password', value) }}
  />
  <TextField
    data={{
      label: 'Repita a senha',
      name: 'passwordConfirmation',
      value: data.values.passwordConfirmation,
      autoComplete: 'new-password',
    }}
    ui={{ type: 'password' }}
    state={{
      error: formState.fieldErrors.passwordConfirmation,
      isDisabled: formState.isSubmitting,
    }}
    actions={{ onChange: (value) => actions.onChange('passwordConfirmation', value) }}
  />

  {#if formState.error}
    <ErrorState data={{ message: formState.error }} ui={{ variant: 'inline' }} />
  {/if}

  <SubmitButton
    data={{ label: 'Criar conta e entrar', loadingLabel: 'Criando a conta…' }}
    ui={{ className: 'w-full' }}
    state={{ isLoading: formState.isSubmitting }}
  />
</form>

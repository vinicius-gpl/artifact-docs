<script lang="ts" module>
  import type { LoginField } from '$lib/hooks/use-login/use-login.svelte';

  /**
   * ENTRAR: e-mail e senha. O erro de cada campo fica colado nele; o erro do envio (credencial
   * errada, conta sem acesso) fica acima do botão, com `role="alert"`.
   */
  export type AcerolaSignInFormProps = {
    data: { values: { email: string; password: string } };
    state: {
      isSubmitting: boolean;
      error: string | null;
      fieldErrors: Partial<Record<LoginField, string>>;
    };
    actions: { onChange: (field: LoginField, value: string) => void; onSubmit: () => void };
  };
</script>

<script lang="ts">
  import ErrorState from '$lib/components/acerola-error-state/acerola-error-state.svelte';
  import SubmitButton from '$lib/components/acerola-submit-button/acerola-submit-button.svelte';
  import TextField from '$lib/components/acerola-text-field/acerola-text-field.svelte';

  let { data, state: formState, actions }: AcerolaSignInFormProps = $props();

  function submit(event: SubmitEvent) {
    event.preventDefault();
    actions.onSubmit();
  }
</script>

<form class="flex flex-col gap-4" novalidate onsubmit={submit}>
  <TextField
    data={{ label: 'E-mail', name: 'email', value: data.values.email, autoComplete: 'username' }}
    ui={{ type: 'email' }}
    state={{ error: formState.fieldErrors.email, isDisabled: formState.isSubmitting }}
    actions={{ onChange: (value) => actions.onChange('email', value) }}
  />
  <TextField
    data={{
      label: 'Senha',
      name: 'password',
      value: data.values.password,
      autoComplete: 'current-password',
    }}
    ui={{ type: 'password' }}
    state={{ error: formState.fieldErrors.password, isDisabled: formState.isSubmitting }}
    actions={{ onChange: (value) => actions.onChange('password', value) }}
  />

  {#if formState.error}
    <ErrorState data={{ message: formState.error }} ui={{ variant: 'inline' }} />
  {/if}

  <SubmitButton
    data={{ label: 'Entrar', loadingLabel: 'Entrando…' }}
    ui={{ className: 'w-full' }}
    state={{ isLoading: formState.isSubmitting }}
  />
</form>

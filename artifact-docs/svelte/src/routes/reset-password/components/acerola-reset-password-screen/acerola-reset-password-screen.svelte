<script lang="ts" module>
  import type { NewPasswordField } from '$lib/hooks/use-reset-password/use-reset-password.svelte';

  /**
   * A SENHA NOVA, aberta pelo link do e-mail do primeiro acesso.
   *
   * `isLinkValid` cobre o caso que mais aparece: link vencido ou aberto pela metade. Sem ele, a
   * pessoa preencheria os dois campos para só então descobrir que precisa começar de novo.
   */
  export type AcerolaResetPasswordScreenProps = {
    data: { values: { password: string; passwordConfirmation: string } };
    state: {
      isSubmitting: boolean;
      isLinkValid: boolean;
      error: string | null;
      fieldErrors: Partial<Record<NewPasswordField, string>>;
    };
    actions: { onChange: (field: NewPasswordField, value: string) => void; onSubmit: () => void };
  };
</script>

<script lang="ts">
  import { BrandMark } from '$lib/components/acerola-brand-mark/acerola-brand-mark';
  import ErrorState from '$lib/components/acerola-error-state/acerola-error-state.svelte';
  import SubmitButton from '$lib/components/acerola-submit-button/acerola-submit-button.svelte';
  import TextField from '$lib/components/acerola-text-field/acerola-text-field.svelte';
  import { revealChildren } from '$lib/motion/motion';

  let { data, state: screenState, actions }: AcerolaResetPasswordScreenProps = $props();

  const LINK_CLASS =
    'text-primary-strong inline-flex h-10 items-center self-center text-sm underline-offset-4 hover:underline';

  function submit(event: SubmitEvent) {
    event.preventDefault();
    actions.onSubmit();
  }
</script>

<div class="flex min-h-dvh items-center justify-center px-4 py-12">
  <section
    class="glass rounded-surface flex w-full max-w-sm flex-col gap-6 p-8"
    aria-label="Criar a senha"
    use:revealChildren
  >
    <BrandMark ui={{ size: 'sm' }} />

    {#if !screenState.isLinkValid}
      <h1 class="text-3xl">Link inválido</h1>
      <p class="text-muted-foreground text-sm">O link venceu ou já foi usado.</p>
      <a href="/first-access" class={LINK_CLASS}>Pedir outro link</a>
    {:else}
      <h1 class="text-3xl">Crie sua senha</h1>
      <form class="flex flex-col gap-4" novalidate onsubmit={submit}>
        <TextField
          data={{
            label: 'Senha',
            name: 'password',
            value: data.values.password,
            autoComplete: 'new-password',
          }}
          ui={{ type: 'password' }}
          state={{
            error: screenState.fieldErrors.password,
            isDisabled: screenState.isSubmitting,
            isAutoFocused: true,
          }}
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
            error: screenState.fieldErrors.passwordConfirmation,
            isDisabled: screenState.isSubmitting,
          }}
          actions={{ onChange: (value) => actions.onChange('passwordConfirmation', value) }}
        />

        {#if screenState.error}
          <ErrorState data={{ message: screenState.error }} ui={{ variant: 'inline' }} />
        {/if}

        <SubmitButton
          data={{ label: 'Salvar senha', loadingLabel: 'Salvando…' }}
          ui={{ className: 'w-full' }}
          state={{ isLoading: screenState.isSubmitting }}
        />
      </form>
    {/if}
  </section>
</div>

<script lang="ts" module>
  /**
   * PRIMEIRO ACESSO: o e-mail, para o Neon Auth mandar o link de criar a senha. Serve também
   * para quem esqueceu a senha — é o mesmo caminho.
   *
   * `isSent` troca o formulário pelo aviso de "olhe o e-mail" sem outra rota. O aviso não diz se
   * a conta existe: contar a diferença entregaria quais e-mails têm conta.
   */
  export type AcerolaFirstAccessScreenProps = {
    data: { email: string };
    state: { isSubmitting: boolean; isSent: boolean; error: string | null; fieldError?: string };
    actions: { onChange: (value: string) => void; onSubmit: () => void };
  };
</script>

<script lang="ts">
  import { BrandMark } from '$lib/components/acerola-brand-mark/acerola-brand-mark';
  import ErrorState from '$lib/components/acerola-error-state/acerola-error-state.svelte';
  import SubmitButton from '$lib/components/acerola-submit-button/acerola-submit-button.svelte';
  import TextField from '$lib/components/acerola-text-field/acerola-text-field.svelte';
  import { revealChildren } from '$lib/motion/motion';

  let { data, state: screenState, actions }: AcerolaFirstAccessScreenProps = $props();

  function submit(event: SubmitEvent) {
    event.preventDefault();
    actions.onSubmit();
  }
</script>

<div class="flex min-h-dvh items-center justify-center px-4 py-12">
  <section
    class="glass rounded-surface flex w-full max-w-sm flex-col gap-6 p-8"
    aria-label="Primeiro acesso"
    use:revealChildren
  >
    <BrandMark ui={{ size: 'sm' }} />

    {#if screenState.isSent}
      <h1 class="text-3xl">Confira seu e-mail</h1>
      <p class="text-muted-foreground text-sm">
        Se este e-mail tiver conta, chegou um link para criar a senha.
      </p>
    {:else}
      <h1 class="text-3xl">Primeiro acesso</h1>
      <form class="flex flex-col gap-4" novalidate onsubmit={submit}>
        <TextField
          data={{ label: 'E-mail', name: 'email', value: data.email, autoComplete: 'email' }}
          ui={{ type: 'email' }}
          state={{
            error: screenState.fieldError,
            isDisabled: screenState.isSubmitting,
            isAutoFocused: true,
          }}
          actions={{ onChange: (value) => actions.onChange(value) }}
        />

        {#if screenState.error}
          <ErrorState data={{ message: screenState.error }} ui={{ variant: 'inline' }} />
        {/if}

        <SubmitButton
          data={{ label: 'Enviar link', loadingLabel: 'Enviando…' }}
          ui={{ className: 'w-full' }}
          state={{ isLoading: screenState.isSubmitting }}
        />
      </form>
    {/if}

    <a
      href="/login"
      class="text-primary-strong inline-flex h-10 items-center self-center text-sm underline-offset-4 hover:underline"
    >
      Voltar para entrar
    </a>
  </section>
</div>

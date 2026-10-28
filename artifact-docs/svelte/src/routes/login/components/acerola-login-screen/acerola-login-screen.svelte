<script lang="ts" module>
  import type { AcerolaSignInFormProps } from '../acerola-sign-in-form/acerola-sign-in-form.svelte';

  /**
   * A TELA DE LOGIN — só o cartão do formulário, no centro. O hub é de uma pessoa só: não há
   * o que vender aqui.
   *
   * O primeiro acesso não é cadastro: a conta nasce no painel da Neon, sem senha, e o link
   * leva para a tela que manda o e-mail de criar a senha (a mesma de "esqueci a senha").
   *
   * No modo mock, um selo avisa que é modo local — para ninguém achar que entrou com uma conta
   * de verdade.
   */
  export type AcerolaLoginScreenProps = {
    data: { isMockMode: boolean; signIn: AcerolaSignInFormProps['data'] };
    state: { signIn: AcerolaSignInFormProps['state'] };
    actions: { signIn: AcerolaSignInFormProps['actions'] };
  };
</script>

<script lang="ts">
  import SignInForm from '../acerola-sign-in-form/acerola-sign-in-form.svelte';
  import { BrandMark } from '$lib/components/acerola-brand-mark/acerola-brand-mark';
  import StatusBadge from '$lib/components/acerola-status-badge/acerola-status-badge.svelte';
  import { revealChildren } from '$lib/motion/motion';

  let { data, state: screenState, actions }: AcerolaLoginScreenProps = $props();
</script>

<div class="flex min-h-dvh items-center justify-center px-4 py-12">
  <section
    class="glass rounded-surface flex w-full max-w-sm flex-col gap-6 p-8"
    aria-label="Entrar no artifact-docs"
    use:revealChildren
  >
    <BrandMark ui={{ size: 'sm' }} />

    <h1 class="text-3xl">Entrar</h1>
    <SignInForm data={data.signIn} state={screenState.signIn} actions={actions.signIn} />

    <a
      href="/first-access"
      class="text-primary-strong inline-flex h-10 items-center self-center text-sm underline-offset-4 hover:underline"
    >
      Primeiro acesso ou esqueci a senha
    </a>

    {#if data.isMockMode}
      <div>
        <StatusBadge data={{ label: 'Modo local' }} ui={{ tone: 'warning', size: 'sm' }} />
      </div>
    {/if}
  </section>
</div>

<script lang="ts" module>
  import type { AcerolaFirstAccessFormProps } from '../acerola-first-access-form/acerola-first-access-form.svelte';
  import type { AcerolaSignInFormProps } from '../acerola-sign-in-form/acerola-sign-in-form.svelte';

  /**
   * A TELA DE LOGIN — só o cartão do formulário, no centro. O hub é de uma pessoa só: não há
   * o que vender aqui.
   *
   * Dois formulários, um de cada vez: ENTRAR (e-mail e senha) e PRIMEIRO ACESSO (criar a conta
   * do dono). O botão "Primeiro acesso" só existe enquanto não há conta (`needsSetup`, que vem
   * de `GET /api/auth/setup`): depois que ela existe, oferecer cadastro seria convidar
   * estranhos. Qual formulário aparece é estado visual, e por isso mora aqui.
   *
   * No modo mock, um selo avisa que é modo local — para ninguém achar que entrou com uma conta
   * de verdade.
   */
  export type AcerolaLoginScreenProps = {
    data: {
      isMockMode: boolean;
      needsSetup: boolean;
      signIn: AcerolaSignInFormProps['data'];
      firstAccess: AcerolaFirstAccessFormProps['data'];
    };
    state: {
      signIn: AcerolaSignInFormProps['state'];
      firstAccess: AcerolaFirstAccessFormProps['state'];
    };
    actions: {
      signIn: AcerolaSignInFormProps['actions'];
      firstAccess: AcerolaFirstAccessFormProps['actions'];
    };
  };
</script>

<script lang="ts">
  import FirstAccessForm from '../acerola-first-access-form/acerola-first-access-form.svelte';
  import SignInForm from '../acerola-sign-in-form/acerola-sign-in-form.svelte';
  import ActionButton from '$lib/components/acerola-action-button/acerola-action-button.svelte';
  import { BrandMark } from '$lib/components/acerola-brand-mark/acerola-brand-mark';
  import StatusBadge from '$lib/components/acerola-status-badge/acerola-status-badge.svelte';
  import { revealChildren } from '$lib/motion/motion';

  let { data, state: screenState, actions }: AcerolaLoginScreenProps = $props();

  let isFirstAccess = $state(false);
</script>

<div class="flex min-h-dvh items-center justify-center px-4 py-12">
  <section
    class="glass rounded-surface flex w-full max-w-sm flex-col gap-6 p-8"
    aria-label="Entrar no artifact-docs"
    use:revealChildren
  >
    <BrandMark ui={{ size: 'sm' }} />

    {#if isFirstAccess && data.needsSetup}
      <h1 class="text-3xl">Primeiro acesso</h1>
      <FirstAccessForm
        data={data.firstAccess}
        state={screenState.firstAccess}
        actions={actions.firstAccess}
      />
      <ActionButton
        data={{ label: 'Já tenho conta · Entrar' }}
        ui={{ variant: 'ghost', size: 'lg', className: 'w-full' }}
        actions={{ onClick: () => (isFirstAccess = false) }}
      />
    {:else}
      <h1 class="text-3xl">Entrar</h1>
      <SignInForm data={data.signIn} state={screenState.signIn} actions={actions.signIn} />
      {#if data.needsSetup}
        <ActionButton
          data={{ label: 'Primeiro acesso · Criar a conta' }}
          ui={{ variant: 'secondary', size: 'lg', className: 'w-full' }}
          actions={{ onClick: () => (isFirstAccess = true) }}
        />
      {/if}
    {/if}

    {#if data.isMockMode}
      <div>
        <StatusBadge data={{ label: 'Modo local' }} ui={{ tone: 'warning', size: 'sm' }} />
      </div>
    {/if}
  </section>
</div>

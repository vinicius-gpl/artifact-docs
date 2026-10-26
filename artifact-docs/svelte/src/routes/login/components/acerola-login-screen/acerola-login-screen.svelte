<script lang="ts" module>
  import type { AcerolaFirstAccessFormProps } from '../acerola-first-access-form/acerola-first-access-form.svelte';
  import type { AcerolaSignInFormProps } from '../acerola-sign-in-form/acerola-sign-in-form.svelte';

  /**
   * A TELA DE LOGIN — dividida ao meio: a marca à esquerda (some no celular) e o formulário à
   * direita.
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

  export const HIGHLIGHTS = [
    'Atalhos para os seus apps, com quem está no ar',
    'Artefatos HTML, Markdown e PDF com preview',
    'Planejamentos e ideias, versionados',
  ] as const;
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

<div class="grid min-h-dvh lg:grid-cols-2">
  <section
    class="border-border/70 relative hidden flex-col justify-between border-r p-12 lg:flex"
    aria-label="Sobre o artifact-docs"
  >
    <div class="flex items-center gap-2">
      <BrandMark ui={{ size: 'sm' }} />
      <span class="text-lg font-medium">artifact-docs</span>
    </div>

    <div class="flex max-w-md flex-col gap-6" use:revealChildren>
      <p class="text-5xl leading-tight font-light">
        Seu hub,
        <span class="text-primary-strong">num lugar só.</span>
      </p>
      <p class="text-muted-foreground text-lg">
        Apps, documentos, planejamentos e ideias — publicados, versionados e à mão.
      </p>
      <ul class="text-muted-foreground flex flex-col gap-2 font-mono text-sm">
        {#each HIGHLIGHTS as highlight (highlight)}
          <li><span class="text-primary-strong" aria-hidden="true">+</span> {highlight}</li>
        {/each}
      </ul>
    </div>

    <p class="text-muted-foreground font-mono text-xs">Uso pessoal · acesso só por convite</p>
  </section>

  <section class="flex items-center justify-center px-4 py-12 sm:px-6">
    <div class="glass rounded-surface flex w-full max-w-sm flex-col gap-6 p-8" use:revealChildren>
      <div class="flex items-center gap-2 lg:hidden">
        <BrandMark ui={{ size: 'sm' }} />
        <span class="text-lg font-medium">artifact-docs</span>
      </div>

      {#if isFirstAccess && data.needsSetup}
        <div class="flex flex-col gap-1">
          <h1 class="text-3xl">Primeiro acesso</h1>
          <p class="text-muted-foreground text-sm">
            Crie a conta de quem é dono do hub. Ela é a única que entra.
          </p>
        </div>
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
        <div class="flex flex-col gap-1">
          <h1 class="text-3xl">Entrar</h1>
          <p class="text-muted-foreground text-sm">Use o e-mail e a senha da sua conta.</p>
        </div>
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
        <div class="flex flex-wrap items-center gap-2">
          <StatusBadge data={{ label: 'Modo local' }} ui={{ tone: 'warning', size: 'sm' }} />
          <span class="text-muted-foreground text-xs">
            Usuário de teste já preenchido. Em produção, a conta é a do Neon Auth.
          </span>
        </div>
      {/if}
    </div>
  </section>
</div>

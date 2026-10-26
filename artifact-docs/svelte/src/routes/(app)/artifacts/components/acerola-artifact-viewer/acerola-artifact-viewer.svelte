<script lang="ts" module>
  import type { ArtifactView } from '$lib/hooks/use-artifact-detail/use-artifact-detail.svelte';

  /**
   * Um artefato aberto: a página publicada ocupando o espaço, e ao lado os dados e as versões.
   *
   * A página roda num `<iframe>` com `sandbox="allow-scripts allow-popups allow-forms"` e SEM
   * `allow-same-origin` — o mesmo CSP que o Go manda em `/a/<token>/` (briefing, "Segurança no
   * mesmo domínio"). O HTML enviado não lê cookie nem chama a API como a pessoa logada.
   *
   * A ordem dos estados é a de sempre: carregando → erro (com o motivo; um 404 diz que o
   * artefato não existe) → conteúdo.
   */
  export type AcerolaArtifactViewerProps = {
    data: { artifact: ArtifactView | null };
    state: { isLoading: boolean; error: string | null };
    actions: { onRetry: () => void };
  };
</script>

<script lang="ts">
  import ArrowLeft from '@lucide/svelte/icons/arrow-left';
  import ExternalLink from '@lucide/svelte/icons/external-link';
  import Lock from '@lucide/svelte/icons/lock';

  import { actionButton } from '$lib/components/acerola-action-button/acerola-action-button.svelte';
  import ErrorState from '$lib/components/acerola-error-state/acerola-error-state.svelte';
  import { Skeleton } from '$lib/components/acerola-skeleton/acerola-skeleton';
  import StatusBadge from '$lib/components/acerola-status-badge/acerola-status-badge.svelte';
  import { cn } from '$lib/utils/cn';

  let { data, state, actions }: AcerolaArtifactViewerProps = $props();
</script>

<div class="flex flex-col gap-5">
  <a
    href="/artifacts"
    class={cn(actionButton({ variant: 'ghost', size: 'lg' }), 'inline-flex w-fit items-center gap-2 text-sm')}
  >
    <ArrowLeft class="size-4" aria-hidden="true" />
    Voltar para Artefatos
  </a>

  {#if state.isLoading}
    <div class="flex flex-col gap-3" aria-busy="true" aria-label="Carregando">
      <Skeleton class="rounded-control h-8 w-80" />
      <Skeleton class="rounded-surface h-[70vh]" />
    </div>
  {:else if state.error || !data.artifact}
    <ErrorState
      data={{ title: 'O artefato não abriu', message: state.error ?? 'Artefato não encontrado.' }}
      actions={{ onRetry: actions.onRetry }}
    />
  {:else}
    {@const artifact = data.artifact}
    <header class="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
      <div class="min-w-0">
        <h1 class="flex items-center gap-2 text-2xl">
          {artifact.title}
          {#if artifact.isPrivate}
            <Lock class="text-muted-foreground size-5" aria-label="Privado" role="img" />
          {/if}
        </h1>
        {#if artifact.description}
          <p class="text-muted-foreground mt-1 text-sm">{artifact.description}</p>
        {/if}
        <div class="mt-2 flex flex-wrap gap-1.5">
          {#each artifact.badges as badge (badge.label)}
            <StatusBadge data={{ label: badge.label }} ui={{ tone: badge.tone, size: 'sm' }} />
          {/each}
        </div>
      </div>
      <a
        href={artifact.url}
        target="_blank"
        rel="noopener noreferrer"
        class={cn(actionButton({ variant: 'secondary', size: 'lg' }), 'inline-flex items-center gap-2 text-sm')}
      >
        <ExternalLink class="size-4" aria-hidden="true" />
        Abrir em nova aba
      </a>
    </header>

    <div class="grid gap-5 lg:grid-cols-[minmax(0,1fr)_18rem]">
      <iframe
        src={artifact.url}
        title={artifact.title}
        sandbox="allow-scripts allow-popups allow-forms"
        class="rounded-surface bg-background h-[70vh] w-full border"
      ></iframe>

      <aside class="flex flex-col gap-3" aria-label="Sobre o artefato">
        <section class="glass rounded-surface p-4">
          <h2 class="text-muted-foreground mb-2 text-xs font-medium tracking-wide uppercase">Dados</h2>
          <dl class="grid grid-cols-[auto_1fr] gap-x-3 gap-y-2 text-sm">
            {#each artifact.facts as fact (fact.label)}
              <dt class="text-muted-foreground">{fact.label}</dt>
              <dd class="min-w-0 break-words">{fact.value}</dd>
            {/each}
          </dl>
        </section>
        <section class="glass rounded-surface p-4">
          <h2 class="text-muted-foreground mb-2 text-xs font-medium tracking-wide uppercase">Versões</h2>
          <ol class="flex flex-col gap-2">
            {#each artifact.versions as version (version.version)}
              <li class="flex items-baseline justify-between gap-2 text-sm">
                <span class="font-mono">{version.version}</span>
                <span class="text-muted-foreground text-xs">{version.detail}</span>
                {#if version.isCurrent}
                  <StatusBadge data={{ label: 'Atual' }} ui={{ tone: 'success', size: 'sm' }} />
                {/if}
              </li>
            {/each}
          </ol>
        </section>
      </aside>
    </div>
  {/if}
</div>

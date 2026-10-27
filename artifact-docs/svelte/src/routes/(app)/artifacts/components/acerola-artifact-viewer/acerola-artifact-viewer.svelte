<script lang="ts" module>
  import type {
    ArtifactPendingAction,
    ArtifactView,
    BackLink,
  } from '$lib/hooks/use-artifact-detail/use-artifact-detail.svelte';

  /**
   * Um artefato aberto: a página publicada ocupando o espaço, e ao lado os dados e as versões.
   *
   * A página roda num `<iframe>` com `sandbox="allow-scripts allow-popups allow-forms"` e SEM
   * `allow-same-origin` — o mesmo CSP que o Go manda em `/a/<token>/` (briefing, "Segurança no
   * mesmo domínio"). O HTML enviado não lê cookie nem chama a API como a pessoa logada.
   *
   * A ordem dos estados é a de sempre: carregando → erro (com o motivo; um 404 diz que o
   * artefato não existe) → conteúdo.
   *
   * As ações em dois grupos, cada um perto do que afeta (Conectividade Uniforme): as do
   * ARTEFATO no cabeçalho — privado, apagar e, por último, publicar versão nova, a principal
   * (Von Restorff) —; as da PÁGINA numa barra colada em cima da prévia — abrir em nova aba e
   * tela cheia. Restaurar mora em cada versão antiga da lista.
   *
   * A tela cheia é uma moldura DENTRO do app, e não o F11 do navegador: a prévia cobre a
   * janela com uma barra em cima, onde fica o "Sair da tela cheia" (e o Esc também sai). É o
   * MESMO `<iframe>` mudando de lugar — a página não recarrega nem perde a rolagem.
   */
  export type AcerolaArtifactViewerProps = {
    /** `back` é a aba do tipo do artefato (Documentos, Planejamentos ou Ideias). */
    data: { artifact: ArtifactView | null; back?: BackLink };
    state: {
      isLoading: boolean;
      error: string | null;
      pendingAction?: ArtifactPendingAction;
      actionError?: string | null;
    };
    actions: {
      onRetry: () => void;
      onPublishVersion?: () => void;
      onTogglePrivate?: () => void;
      onRestore?: (version: number) => void;
      onDelete?: () => void;
    };
  };
</script>

<script lang="ts">
  import ArrowLeft from '@lucide/svelte/icons/arrow-left';
  import ExternalLink from '@lucide/svelte/icons/external-link';
  import Lock from '@lucide/svelte/icons/lock';
  import LockOpen from '@lucide/svelte/icons/lock-open';
  import Maximize2 from '@lucide/svelte/icons/maximize-2';
  import Minimize2 from '@lucide/svelte/icons/minimize-2';
  import Trash2 from '@lucide/svelte/icons/trash-2';
  import Upload from '@lucide/svelte/icons/upload';

  import ActionButton, {
    actionButton,
  } from '$lib/components/acerola-action-button/acerola-action-button.svelte';
  import ErrorState from '$lib/components/acerola-error-state/acerola-error-state.svelte';
  import { Skeleton } from '$lib/components/acerola-skeleton/acerola-skeleton';
  import StatusBadge from '$lib/components/acerola-status-badge/acerola-status-badge.svelte';
  import { cn } from '$lib/utils/cn';

  /* `viewState`, e não `state`: com a prop chamada `state`, a runa `$state` ficaria escondida. */
  let { data, state: viewState, actions }: AcerolaArtifactViewerProps = $props();

  const isBusy = $derived(Boolean(viewState.pendingAction));

  /* Estado puramente visual: a prévia está na moldura de tela cheia. */
  let isExpanded = $state(false);

  function onWindowKeydown(event: KeyboardEvent): void {
    if (isExpanded && event.key === 'Escape') isExpanded = false;
  }
</script>

<svelte:window onkeydown={onWindowKeydown} />

<div class="flex flex-col gap-5">
  <a
    href={data.back?.href ?? '/artifacts'}
    class={cn(actionButton({ variant: 'ghost', size: 'lg' }), 'inline-flex w-fit items-center gap-2 text-sm')}
  >
    <ArrowLeft class="size-4" aria-hidden="true" />
    {data.back?.label ?? 'Voltar'}
  </a>

  {#if viewState.isLoading}
    <div class="flex flex-col gap-3" aria-busy="true" aria-label="Carregando">
      <Skeleton class="rounded-control h-8 w-80" />
      <Skeleton class="rounded-surface h-[70vh]" />
    </div>
  {:else if viewState.error || !data.artifact}
    <ErrorState
      data={{ title: 'O artefato não abriu', message: viewState.error ?? 'Artefato não encontrado.' }}
      actions={{ onRetry: actions.onRetry }}
    />
  {:else}
    {@const artifact = data.artifact}
    {@const current = artifact.versions.find((version) => version.isCurrent)}
    <header class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
      <div class="min-w-0">
        <h1 class="flex items-center gap-2 text-2xl">
          {artifact.title}
          {#if artifact.isPrivate}
            <Lock class="text-muted-foreground size-5" aria-label="Privado" role="img" />
          {/if}
        </h1>
        {#if artifact.description}
          <p class="text-muted-foreground mt-1 max-w-prose text-sm">{artifact.description}</p>
        {/if}
        <div class="mt-2 flex flex-wrap gap-1.5">
          {#each artifact.badges as badge (badge.label)}
            <StatusBadge data={{ label: badge.label }} ui={{ tone: badge.tone, size: 'sm' }} />
          {/each}
        </div>
      </div>

      <div class="flex shrink-0 flex-wrap items-center gap-2" role="group" aria-label="Ações do artefato">
        {#if actions.onTogglePrivate}
          <ActionButton
            data={{ label: artifact.isPrivate ? 'Tornar público' : 'Tornar privado', loadingLabel: 'Salvando…' }}
            ui={{ variant: 'secondary', size: 'lg', icon: artifact.isPrivate ? LockOpen : Lock }}
            state={{ isLoading: viewState.pendingAction?.kind === 'private', isDisabled: isBusy }}
            actions={{ onClick: actions.onTogglePrivate }}
          />
        {/if}
        {#if actions.onDelete}
          <ActionButton
            data={{ label: 'Apagar' }}
            ui={{ variant: 'ghost', size: 'lg', icon: Trash2 }}
            state={{ isDisabled: isBusy }}
            actions={{ onClick: actions.onDelete }}
          />
        {/if}
        {#if actions.onPublishVersion}
          <ActionButton
            data={{ label: 'Publicar nova versão' }}
            ui={{ variant: 'primary', size: 'lg', icon: Upload }}
            state={{ isDisabled: isBusy }}
            actions={{ onClick: actions.onPublishVersion }}
          />
        {/if}
      </div>
    </header>

    {#if viewState.actionError}
      <ErrorState data={{ message: viewState.actionError }} ui={{ variant: 'inline' }} />
    {/if}

    <div class="grid gap-5 lg:grid-cols-[minmax(0,1fr)_18rem]">
      <!-- A prévia: a barra da página e o <iframe>. Na tela cheia, o MESMO bloco passa a cobrir
           a janela (fixed) — nada é remontado, a página não recarrega. -->
      <section
        class={cn(
          'flex min-w-0 flex-col',
          isExpanded && 'bg-background fixed inset-0 z-50 gap-2 p-3 sm:p-4',
        )}
        aria-label={isExpanded ? `${artifact.title} em tela cheia` : 'Prévia da página publicada'}
        data-expanded={isExpanded}
      >
        <div class="glass rounded-surface flex min-h-12 items-center gap-2 border px-3">
          <span class="text-muted-foreground min-w-0 flex-1 truncate text-xs">
            {#if isExpanded}
              <span class="text-foreground text-sm font-medium">{artifact.title}</span>
            {:else}
              Página publicada
            {/if}
            {#if current}<span class="font-mono"> · {current.version}</span>{/if}
          </span>
          <a
            href={artifact.url}
            target="_blank"
            rel="noopener noreferrer"
            aria-label="Abrir em nova aba"
            title="Abrir em nova aba"
            class={actionButton({ variant: 'ghost', size: 'lg', isIconOnly: true })}
          >
            <ExternalLink class="size-4" aria-hidden="true" />
          </a>
          <ActionButton
            data={{ label: isExpanded ? 'Sair da tela cheia' : 'Tela cheia' }}
            ui={{
              variant: isExpanded ? 'secondary' : 'ghost',
              size: 'lg',
              icon: isExpanded ? Minimize2 : Maximize2,
              isIconOnly: !isExpanded,
            }}
            state={{ isPressed: isExpanded }}
            actions={{ onClick: () => (isExpanded = !isExpanded) }}
          />
        </div>
        <iframe
          src={artifact.url}
          title={artifact.title}
          sandbox="allow-scripts allow-popups allow-forms"
          class={cn('rounded-surface bg-background w-full border', isExpanded ? 'min-h-0 flex-1' : 'mt-2 h-[70vh]')}
        ></iframe>
      </section>

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
                {:else if actions.onRestore}
                  {@const restoring =
                    viewState.pendingAction?.kind === 'restore' && viewState.pendingAction.version === version.number}
                  <ActionButton
                    data={{ label: `Restaurar ${version.version}`, loadingLabel: 'Restaurando…' }}
                    ui={{ variant: 'ghost', size: 'sm' }}
                    state={{ isLoading: restoring, isDisabled: isBusy }}
                    actions={{ onClick: () => actions.onRestore?.(version.number) }}
                  />
                {/if}
              </li>
            {/each}
          </ol>
        </section>
      </aside>
    </div>
  {/if}
</div>

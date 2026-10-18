<script lang="ts">
  import AppWindow from '@lucide/svelte/icons/app-window';

  import EmptyState from '$lib/components/acerola-empty-state/acerola-empty-state.svelte';
  import PageHeader from '$lib/components/acerola-page-header/acerola-page-header.svelte';

  /* Esqueleto: a tela existe no menu e na rota; o conteúdo entra na feature dela. */
</script>

<svelte:head>
  <title>Apps</title>
</svelte:head>

<div class="mx-auto flex w-full max-w-6xl flex-col gap-6">
  <PageHeader data={{ title: 'Apps', description: 'Atalhos para os apps self-hosted e da VPS, agrupados por tipo.' }} />
  <EmptyState data={{ title: 'Nada por aqui ainda', description: 'Os apps cadastrados aparecem aqui.' }} ui={{ icon: AppWindow }} />
</div>

<script lang="ts">
  import Lightbulb from '@lucide/svelte/icons/lightbulb';

  import EmptyState from '$lib/components/acerola-empty-state/acerola-empty-state.svelte';
  import PageHeader from '$lib/components/acerola-page-header/acerola-page-header.svelte';

  /* Esqueleto: a tela existe no menu e na rota; o conteúdo entra na feature dela. */
</script>

<svelte:head>
  <title>Ideias</title>
</svelte:head>

<div class="mx-auto flex w-full max-w-6xl flex-col gap-6">
  <PageHeader data={{ title: 'Ideias', description: 'O banco de ideias.' }} />
  <EmptyState data={{ title: 'Nada por aqui ainda', description: 'As ideias guardadas aparecem aqui.' }} ui={{ icon: Lightbulb }} />
</div>

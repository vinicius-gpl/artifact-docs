<script lang="ts">
  import ListTodo from '@lucide/svelte/icons/list-todo';

  import EmptyState from '$lib/components/acerola-empty-state/acerola-empty-state.svelte';
  import PageHeader from '$lib/components/acerola-page-header/acerola-page-header.svelte';

  /* Esqueleto: a tela existe no menu e na rota; o conteúdo entra na feature dela. */
</script>

<svelte:head>
  <title>Planejamentos</title>
</svelte:head>

<div class="mx-auto flex w-full max-w-6xl flex-col gap-6">
  <PageHeader data={{ title: 'Planejamentos', description: 'Os planos em andamento.' }} />
  <EmptyState data={{ title: 'Nada por aqui ainda', description: 'Os planejamentos publicados aparecem aqui.' }} ui={{ icon: ListTodo }} />
</div>

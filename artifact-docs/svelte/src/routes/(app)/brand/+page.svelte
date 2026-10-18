<script lang="ts">
  import Palette from '@lucide/svelte/icons/palette';

  import EmptyState from '$lib/components/acerola-empty-state/acerola-empty-state.svelte';
  import PageHeader from '$lib/components/acerola-page-header/acerola-page-header.svelte';

  /* Esqueleto: a tela existe no menu e na rota; o conteúdo entra na feature dela. */
</script>

<svelte:head>
  <title>Marca</title>
</svelte:head>

<div class="mx-auto flex w-full max-w-6xl flex-col gap-6">
  <PageHeader data={{ title: 'Marca', description: 'Itens da marca e gerador de padrão de cores.' }} />
  <EmptyState data={{ title: 'Nada por aqui ainda', description: 'Os itens da marca aparecem aqui.' }} ui={{ icon: Palette }} />
</div>

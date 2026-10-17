<script lang="ts">
  import LayoutGrid from '@lucide/svelte/icons/layout-grid';

  import EmptyState from '$lib/components/acerola-empty-state/acerola-empty-state.svelte';
  import PageHeader from '$lib/components/acerola-page-header/acerola-page-header.svelte';

  /* Esqueleto: a tela existe no menu e na rota; o conteúdo entra na feature dela. */
</script>

<svelte:head>
  <title>Artefatos</title>
</svelte:head>

<div class="mx-auto flex w-full max-w-6xl flex-col gap-6">
  <PageHeader data={{ title: 'Artefatos', description: 'Documentos, páginas e PDFs publicados.' }} />
  <EmptyState data={{ title: 'Nada por aqui ainda', description: 'Os artefatos publicados aparecem aqui, com preview.' }} ui={{ icon: LayoutGrid }} />
</div>

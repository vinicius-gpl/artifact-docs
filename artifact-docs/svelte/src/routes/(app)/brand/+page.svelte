<script lang="ts">
  import PaletteList from './components/acerola-palette-list/acerola-palette-list.svelte';
  import PageHeader from '$lib/components/acerola-page-header/acerola-page-header.svelte';
  import { useBrandPalettesModel } from '$lib/hooks/use-brand-palettes/use-brand-palettes.svelte';

  const brand = useBrandPalettesModel();
</script>

<svelte:head>
  <title>Marca</title>
</svelte:head>

<div class="mx-auto flex w-full max-w-6xl flex-col gap-5">
  <PageHeader
    data={{
      title: 'Marca',
      description: 'As paletas geradas e o CSS da marca, que os templates e as skills usam.',
    }}
  />

  <PaletteList data={brand.data} state={brand.state} actions={brand.actions} />
</div>

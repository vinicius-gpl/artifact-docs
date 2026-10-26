<script lang="ts">
  import { QueryClient, setQueryClientContext } from '@tanstack/svelte-query';

  import { useBrandPalettesModel, type BrandPalettesModel } from './use-brand-palettes.svelte';

  /** O hook usa o svelte-query: monta dentro de um componente, com um `QueryClient`. */
  let { onReady }: { onReady: (model: BrandPalettesModel) => void } = $props();

  setQueryClientContext(new QueryClient({ defaultOptions: { queries: { retry: false } } }));

  onReady(useBrandPalettesModel());
</script>

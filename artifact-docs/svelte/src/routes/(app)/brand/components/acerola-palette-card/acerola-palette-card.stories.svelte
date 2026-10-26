<script module lang="ts">
  import { defineMeta } from '@storybook/addon-svelte-csf';

  import { createMockDatabase } from '$lib/api/mock-fixtures';
  import {
    toPaletteView,
    type PaletteView,
  } from '$lib/hooks/use-brand-palettes/use-brand-palettes.svelte';
  import PaletteCard from './acerola-palette-card.svelte';

  /* As cores vêm do fixture do modo mock: hex de exemplo não fica escrito no componente. */
  const palette: PaletteView = toPaletteView(createMockDatabase(0).palettes[0]!);

  const { Story } = defineMeta({
    title: 'Features/Brand/AcerolaPaletteCard',
    component: PaletteCard,
  });
</script>

<Story name="Default" args={{ data: { palette } }} />

<Story name="Draft" args={{ data: { palette: { ...palette, isBrand: false, name: 'Grupo Azuos · rascunho' } } }} />

<!-- Caso limite: uma cor só. -->
<Story name="SingleColor" args={{ data: { palette: { ...palette, colors: palette.colors.slice(0, 1) } } }} />

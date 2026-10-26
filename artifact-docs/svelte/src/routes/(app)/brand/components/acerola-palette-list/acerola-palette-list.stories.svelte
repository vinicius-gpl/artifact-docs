<script module lang="ts">
  import { defineMeta } from '@storybook/addon-svelte-csf';
  import { fn } from 'storybook/test';

  import { createMockDatabase } from '$lib/api/mock-fixtures';
  import {
    toPaletteView,
    type PaletteView,
  } from '$lib/hooks/use-brand-palettes/use-brand-palettes.svelte';
  import PaletteList from './acerola-palette-list.svelte';

  /* As cores vêm do fixture do modo mock: hex de exemplo não fica escrito no componente. */
  const palette: PaletteView = toPaletteView(createMockDatabase(0).palettes[0]!);

  const idle = { isLoading: false, error: null };

  const { Story } = defineMeta({
    title: 'Features/Brand/AcerolaPaletteList',
    component: PaletteList,
    args: { actions: { onRetry: fn() } },
  });
</script>

<Story name="Default" args={{ data: { palettes: [palette] }, state: idle }} />

<Story name="Loading" args={{ data: { palettes: [] }, state: { ...idle, isLoading: true } }} />

<Story name="LoadFailed" args={{ data: { palettes: [] }, state: { ...idle, error: 'O R2 não respondeu.' } }} />

<!-- Caso limite: nenhuma paleta. -->
<Story name="Empty" args={{ data: { palettes: [] }, state: idle }} />

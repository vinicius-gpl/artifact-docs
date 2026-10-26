import { createQuery } from '@tanstack/svelte-query';

import { readError } from '$lib/api/http-client';
import { palettesApi } from '$lib/api/palettes.api';
import { type Palette, type PaletteColor } from '$lib/contracts/palette.schema';
import { mirrorStore } from '$lib/hooks/use-mirror-store/use-mirror-store.svelte';
import { formatDate } from '$lib/utils/format-date';

export type PaletteView = {
  id: string;
  name: string;
  isBrand: boolean;
  colors: PaletteColor[];
  downloadUrl: string;
  /** O nome do arquivo baixado — a pessoa vê, então pode ter o nome da paleta. */
  downloadName: string;
  createdLabel: string;
};

export type BrandPalettesModel = {
  data: { palettes: PaletteView[] };
  state: { isLoading: boolean; error: string | null };
  actions: { onRetry: () => void };
};

/** "Artifact-docs · Carbon" vira `artifact-docs-carbon.css`. */
export function fileNameOf(name: string): string {
  const slug = name
    .normalize('NFD')
    .replace(/[̀-ͯ]/g, '')
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '');

  return `${slug || 'paleta'}.css`;
}

export function toPaletteView(palette: Palette): PaletteView {
  return {
    id: palette.id,
    name: palette.name,
    isBrand: palette.isBrand,
    colors: palette.colors,
    downloadUrl: palette.outputUrl,
    downloadName: fileNameOf(palette.name),
    createdLabel: `Gerada em ${formatDate(palette.createdAt)}`,
  };
}

/** A oficial primeiro: é a que os templates e as skills usam, e a que a pessoa procura. */
export function sortPalettes(palettes: Palette[]): Palette[] {
  return [...palettes].sort((a, b) => Number(b.isBrand) - Number(a.isBrand));
}

/** A casa da marca: as paletas geradas, com a oficial em destaque e o `.css` para baixar. */
export function useBrandPalettesModel(): BrandPalettesModel {
  const list = mirrorStore(
    createQuery({ queryKey: ['palettes'], queryFn: () => palettesApi.list() }),
  );

  return {
    get data() {
      return { palettes: sortPalettes(list.current.data?.items ?? []).map(toPaletteView) };
    },
    get state() {
      return { isLoading: list.current.isPending, error: readError(list.current.error) };
    },
    actions: {
      onRetry: () => {
        void list.current.refetch();
      },
    },
  };
}

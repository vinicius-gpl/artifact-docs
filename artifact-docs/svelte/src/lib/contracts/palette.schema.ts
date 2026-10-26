import { z } from 'zod';

/**
 * Paleta — a FORMA do que a API devolve em `/api/palettes` (briefing, módulo 5).
 *
 * A pessoa sobe um `.css` com as cores base; o backend gera as escalas e guarda o `.css`
 * resultante. A paleta marcada com `isBrand` é a oficial: os templates e as skills puxam dela.
 *
 * As cores chegam como DADO (`value`), não como classe: são justamente o que a tela precisa
 * mostrar como amostra. O lado Go mora em `artifact-docs/src-go/` — mudou aqui, muda lá.
 */
export const paletteColorSchema = z.object({
  /** O nome do token, como no CSS gerado (`primary`, `accent-hero`). */
  token: z.string().min(1),
  /** A cor no tema claro, em hexadecimal. */
  light: z.string().regex(/^#[0-9a-f]{6}$/i),
  /** A cor no tema escuro, em hexadecimal. */
  dark: z.string().regex(/^#[0-9a-f]{6}$/i),
});

export type PaletteColor = z.infer<typeof paletteColorSchema>;

export const paletteSchema = z.object({
  id: z.string().min(1),
  name: z.string().min(1),
  isBrand: z.boolean(),
  colors: z.array(paletteColorSchema),
  /** O `.css` gerado, para baixar. */
  outputUrl: z.string().min(1),
  createdAt: z.string().datetime(),
});

export type Palette = z.infer<typeof paletteSchema>;

export const paletteListSchema = z.object({ items: z.array(paletteSchema) });

export type PaletteList = z.infer<typeof paletteListSchema>;

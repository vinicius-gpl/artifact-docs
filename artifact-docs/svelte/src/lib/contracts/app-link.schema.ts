import { z } from 'zod';

/**
 * Hub de apps — a FORMA do que a API devolve em `/api/apps`.
 *
 * Chama `app-link` e não `app` porque "app" já é o nome do próprio sistema em todo canto do
 * SvelteKit (`$app/state`, `routes/(app)`); um `AppSchema` aqui seria ambíguo na primeira
 * leitura.
 *
 * Os tipos são criados pela pessoa (briefing, módulo 4) e os apps aparecem agrupados por eles,
 * na ordem de `position`. O lado Go mora em `artifact-docs/src-go/` — mudou aqui, muda lá.
 */
export const appLinkStatusSchema = z.enum(['online', 'offline', 'unknown']);

export type AppLinkStatus = z.infer<typeof appLinkStatusSchema>;

export const APP_LINK_STATUS_LABELS: Record<AppLinkStatus, string> = {
  online: 'No ar',
  offline: 'Fora do ar',
  unknown: 'Sem verificação',
};

export const appTypeSchema = z.object({
  id: z.string().min(1),
  name: z.string().min(1),
  position: z.number().int().nonnegative(),
});

export type AppType = z.infer<typeof appTypeSchema>;

export const appLinkSchema = z.object({
  id: z.string().min(1),
  typeId: z.string().min(1),
  name: z.string().min(1),
  url: z.string().url(),
  /** A logo enviada; sem ela, a tela usa o favicon do próprio endereço. */
  logoUrl: z.string().nullable(),
  status: appLinkStatusSchema,
  /** Quando a goroutine pingou pela última vez. `null` = nunca. */
  checkedAt: z.string().datetime().nullable(),
});

export type AppLink = z.infer<typeof appLinkSchema>;

export const appHubSchema = z.object({
  types: z.array(appTypeSchema),
  apps: z.array(appLinkSchema),
});

export type AppHub = z.infer<typeof appHubSchema>;

/**
 * O CADASTRO DE APP — `POST /api/apps` e `PATCH /api/apps/<id>`. O tipo vai pelo NOME: o que
 * não existe nasce na hora, e o que fica sem apps some. O Go confere o mesmo
 * (`modules/apps/validation.go`); mudou aqui, muda lá.
 */
export const appLinkInputSchema = z.object({
  name: z
    .string()
    .trim()
    .min(1, 'Informe o nome do app')
    .max(80, 'O nome tem no máximo 80 caracteres'),
  url: z
    .string()
    .trim()
    .max(2048, 'O endereço é longo demais')
    .regex(/^https?:\/\/[^\s/]+/i, 'Informe o endereço completo, começando com https://'),
  typeName: z
    .string()
    .trim()
    .min(1, 'Informe o tipo do app')
    .max(40, 'O tipo tem no máximo 40 caracteres'),
});

export type AppLinkInput = z.infer<typeof appLinkInputSchema>;

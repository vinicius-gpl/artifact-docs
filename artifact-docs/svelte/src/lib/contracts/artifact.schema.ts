import { z } from 'zod';

/**
 * Artefato — a FORMA do que a API devolve em `/api/artifacts`.
 *
 * Documento, planejamento e ideia são o mesmo artefato com um `kind` diferente (briefing,
 * módulo 3): compartilham galeria, busca, preview e versões. É por isso que existe um contrato
 * só, e não três.
 *
 * O lado Go mora em `artifact-docs/src-go/` e precisa devolver exatamente isto — mudou aqui,
 * muda lá. Enquanto o Go não tem a rota, quem responde é o mock (`lib/api/mock-server.ts`),
 * e ele também passa por este schema.
 */
export const artifactKindSchema = z.enum(['document', 'plan', 'idea']);

export type ArtifactKind = z.infer<typeof artifactKindSchema>;

export const ARTIFACT_KIND_LABELS: Record<ArtifactKind, string> = {
  document: 'Documento',
  plan: 'Planejamento',
  idea: 'Ideia',
};

/** Como o arquivo chegou. Markdown vira HTML no upload; o formato fica para o selo do card. */
export const artifactFormatSchema = z.enum(['html', 'md', 'pdf']);

export type ArtifactFormat = z.infer<typeof artifactFormatSchema>;

export const ARTIFACT_FORMAT_LABELS: Record<ArtifactFormat, string> = {
  html: 'HTML',
  md: 'Markdown',
  pdf: 'PDF',
};

/**
 * A situação de um planejamento ou ideia. Documento não tem situação (`null`): ele só existe
 * ou não existe.
 */
export const artifactStatusSchema = z.enum(['draft', 'active', 'done', 'archived']);

export type ArtifactStatus = z.infer<typeof artifactStatusSchema>;

export const ARTIFACT_STATUS_LABELS: Record<ArtifactStatus, string> = {
  draft: 'Rascunho',
  active: 'Em andamento',
  done: 'Concluído',
  archived: 'Arquivado',
};

/** A label é o endereço do artefato: só minúscula, número e hífen (briefing, "Segurança"). */
export const artifactLabelSchema = z
  .string()
  .min(1, 'Informe a label')
  .max(80, 'A label tem no máximo 80 caracteres')
  .regex(/^[a-z0-9]+(?:-[a-z0-9]+)*$/, 'Use só letras minúsculas, números e hífen');

export const artifactSchema = z.object({
  id: z.string().min(1),
  label: artifactLabelSchema,
  title: z.string().min(1),
  description: z.string().nullable(),
  kind: artifactKindSchema,
  format: artifactFormatSchema,
  status: artifactStatusSchema.nullable(),
  isPrivate: z.boolean(),
  isPinned: z.boolean(),
  /**
   * O endereço servido (`/a/<token>/`). É ele que o preview e o visualizador abrem num
   * `<iframe>` com sandbox — a tela nunca monta esse link sozinha.
   */
  url: z.string().min(1),
  /** A miniatura gerada por quem publicou; sem ela, o card usa o próprio `url` reduzido. */
  thumbUrl: z.string().nullable(),
  currentVersion: z.number().int().positive(),
  sizeBytes: z.number().int().nonnegative(),
  tags: z.array(z.string()),
  createdAt: z.string().datetime(),
  updatedAt: z.string().datetime(),
  /** Último acesso real — é o que ordena "Recentes". */
  lastViewedAt: z.string().datetime().nullable(),
});

export type Artifact = z.infer<typeof artifactSchema>;

export const artifactListSchema = z.object({ items: z.array(artifactSchema) });

export type ArtifactList = z.infer<typeof artifactListSchema>;

export const artifactVersionSchema = z.object({
  version: z.number().int().positive(),
  sizeBytes: z.number().int().nonnegative(),
  hasThumb: z.boolean(),
  createdAt: z.string().datetime(),
});

export type ArtifactVersion = z.infer<typeof artifactVersionSchema>;

export const artifactDetailSchema = artifactSchema.extend({
  versions: z.array(artifactVersionSchema),
});

export type ArtifactDetail = z.infer<typeof artifactDetailSchema>;

/** O que a lista aceita de filtro, na query string de `GET /api/artifacts`. */
export type ArtifactListQuery = {
  kind?: ArtifactKind;
  q?: string;
};

/**
 * O que dá para mudar sem subir versão nova: `PATCH /api/artifacts/<label>`. Campo que não vai
 * fica como está; `description: null` e `tags: []` limpam. As regras são as da publicação — o
 * Go confere o mesmo (`modules/artifacts/handlers.go`, `toPatch`).
 */
export const artifactPatchSchema = z
  .object({
    isPinned: z.boolean(),
    isPrivate: z.boolean(),
    status: artifactStatusSchema.nullable(),
    title: z
      .string()
      .trim()
      .min(1, 'Informe o título')
      .max(200, 'O título tem no máximo 200 caracteres'),
    description: z
      .string()
      .trim()
      .max(1000, 'A descrição tem no máximo 1000 caracteres')
      .nullable(),
    tags: z
      .array(z.string().trim().min(1).max(40, 'Cada tag tem no máximo 40 caracteres'))
      .max(20, 'Use no máximo 20 tags'),
  })
  .partial();

export type ArtifactPatch = z.infer<typeof artifactPatchSchema>;

/** O que se publica (briefing, módulo 1): HTML solto, um site em .zip, Markdown ou PDF. */
export const PUBLISH_ACCEPT = '.html,.htm,.zip,.md,.markdown,.pdf';

/** O maior envio que a API aceita — o `maxUploadBytes` do Go. */
export const MAX_PUBLISH_BYTES = 100 * 1024 * 1024;

const publishFileSchema = z
  .instanceof(File, { message: 'Escolha o arquivo a publicar' })
  .refine((file) => /\.(html?|zip|md|markdown|pdf)$/i.test(file.name), {
    message: 'Envie um arquivo .html, .zip, .md ou .pdf',
  })
  .refine((file) => file.size > 0, { message: 'O arquivo está vazio' })
  .refine((file) => file.size <= MAX_PUBLISH_BYTES, { message: 'O arquivo passa de 100 MB' });

/**
 * O formulário de PUBLICAR UM ARTEFATO NOVO — `PUT /api/artifacts/<label>` em multipart. O Go
 * confere o mesmo (`modules/artifacts/publish.go`); mudou aqui, muda lá.
 */
export const artifactPublishSchema = z.object({
  file: publishFileSchema,
  title: z
    .string()
    .trim()
    .min(1, 'Informe o título')
    .max(200, 'O título tem no máximo 200 caracteres'),
  label: artifactLabelSchema,
  kind: artifactKindSchema,
  description: z.string().trim().max(1000, 'A descrição tem no máximo 1000 caracteres'),
  /** Separadas por vírgula, como a pessoa digita. */
  tags: z.string().trim(),
  /**
   * Situação de planejamento e ideia; documento não tem. Sem ela, o Go põe "rascunho" no
   * planejamento e na ideia novos.
   */
  status: artifactStatusSchema.optional(),
  /** Sem ele, o artefato novo nasce PRIVADO (no Go): publicar aberto é escolha, não descuido. */
  isPrivate: z.boolean().optional(),
});

export type ArtifactPublishInput = z.infer<typeof artifactPublishSchema>;

/**
 * UMA VERSÃO NOVA de um artefato que já existe: só o arquivo é obrigatório. Título,
 * descrição e tags vazios mantêm o que já está publicado — e o tipo nem vai: no Go, `kind`
 * preenchido troca o tipo do artefato.
 */
export const artifactVersionPublishSchema = artifactPublishSchema.extend({
  title: z.string().trim().max(200, 'O título tem no máximo 200 caracteres'),
  kind: artifactKindSchema.optional(),
});

export type ArtifactVersionPublishInput = z.infer<typeof artifactVersionPublishSchema>;

/** O corpo do rollback: `POST /api/artifacts/<label>/restore`. */
export const artifactRestoreSchema = z.object({ version: z.number().int().positive() });

export type ArtifactRestore = z.infer<typeof artifactRestoreSchema>;

import {
  appHubSchema,
  appLinkInputSchema,
  appLinkSchema,
  type AppLink,
  type AppLinkInput,
} from '$lib/contracts/app-link.schema';
import {
  artifactDetailSchema,
  artifactKindSchema,
  artifactLabelSchema,
  artifactListSchema,
  artifactPatchSchema,
  artifactRestoreSchema,
  artifactSchema,
  type Artifact,
  type ArtifactFormat,
} from '$lib/contracts/artifact.schema';

import { ApiError, type RequestOptions } from './http-client';
import {
  hasMockSession,
  mockMe,
  mockSetup,
  mockSignIn,
  mockSetPassword,
  setMockSession,
} from './mock-auth';
import { createMockDatabase, htmlUrl, mockPage, type MockDatabase } from './mock-fixtures';

/**
 * O BACKEND DE MENTIRA — responde às mesmas rotas `/api/...` que o Go vai responder.
 *
 * Liga com `VITE_API_MOCKS=true` (o `mise run dev` liga por padrão). Quem desvia para cá é o
 * `apiRequest`, e só ele: os `*.api.ts`, os hooks e as telas não sabem que existe mock. Por
 * isso trocar pelo Go não muda uma linha de tela — a rota passa a existir no Go e o modo mock
 * é desligado (`VITE_API_MOCKS=false mise run dev`).
 *
 * A tabela `ROUTES` abaixo é também a LISTA DO QUE O GO PRECISA IMPLEMENTAR: método, caminho
 * e o schema da resposta. Toda resposta passa pelo schema de `lib/contracts` antes de sair —
 * se o mock fugir do contrato, quebra aqui, e não na tela.
 *
 * Os dados vivem em memória: recarregar a página volta ao começo. É de propósito — mock que
 * persiste vira um segundo banco, com dado que ninguém lembra de onde veio.
 */

type MockContext = {
  db: MockDatabase;
  params: Record<string, string>;
  query: URLSearchParams;
  body: unknown;
};

type MockRoute = {
  method: NonNullable<RequestOptions['method']>;
  /** Responde sem sessão — só o próprio login. Todo o resto devolve 401, como o Go. */
  isPublic?: boolean;
  /** Caminho com `:param`, igual ao padrão do `http.ServeMux` do Go (`{label}`). */
  path: string;
  handle: (context: MockContext) => unknown;
};

/** Um atraso curto, para o "carregando" aparecer como vai aparecer com a rede de verdade. */
const MOCK_LATENCY_MS = 250;

let db: MockDatabase = createMockDatabase();

/** Volta o banco ao começo — os testes chamam antes de cada caso. */
export function resetMockDatabase(now?: number): void {
  db = createMockDatabase(now);
}

function findArtifact(label: string): Artifact {
  const artifact = db.artifacts.find((item) => item.label === label);
  if (!artifact) throw new ApiError(404, 'Artefato não encontrado. Ele pode ter sido apagado.');

  return artifact;
}

/** A busca do Go é `tsvector` em português; aqui basta título, descrição e tags. */
function matchesSearch(artifact: Artifact, search: string): boolean {
  const haystack = [artifact.title, artifact.description ?? '', ...artifact.tags]
    .join(' ')
    .toLocaleLowerCase('pt-BR');

  return haystack.includes(search.toLocaleLowerCase('pt-BR'));
}

function listArtifacts({ query }: MockContext) {
  const kind = artifactKindSchema.safeParse(query.get('kind'));
  const search = query.get('q')?.trim() ?? '';
  const items = db.artifacts
    .filter((artifact) => !kind.success || artifact.kind === kind.data)
    .filter((artifact) => !search || matchesSearch(artifact, search))
    .sort((a, b) => b.updatedAt.localeCompare(a.updatedAt));

  return artifactListSchema.parse({ items });
}

function getArtifact({ params }: MockContext) {
  const artifact = findArtifact(params.label ?? '');
  /* Abrir conta como acesso: é o que ordena "Recentes". */
  artifact.lastViewedAt = new Date().toISOString();

  return artifactDetailSchema.parse({ ...artifact, versions: db.versions[artifact.id] ?? [] });
}

function patchArtifact({ params, body }: MockContext) {
  const artifact = findArtifact(params.label ?? '');
  const patch = artifactPatchSchema.safeParse(body);
  if (!patch.success) throw new ApiError(400, 'Alteração inválida.');

  Object.assign(artifact, patch.data);

  return artifactSchema.parse(artifact);
}

/** O formato pelo nome do arquivo — a mesma regra do `buildBundle` do Go. */
function formatOf(fileName: string): ArtifactFormat | null {
  if (/\.html?$/i.test(fileName) || /\.zip$/i.test(fileName)) return 'html';
  if (/\.(md|markdown)$/i.test(fileName)) return 'md';
  if (/\.pdf$/i.test(fileName)) return 'pdf';

  return null;
}

function formText(form: FormData, field: string): string | null {
  const value = form.get(field);

  return typeof value === 'string' && value.trim() ? value.trim() : null;
}

/**
 * O que o mock serve no lugar de `/a/<token>/`: o próprio HTML enviado (`htmlUrl`). Markdown,
 * .zip e PDF viram uma página dizendo o que foi publicado — o mock não converte nem descompacta.
 */
async function servedUrlOf(file: File, title: string, format: ArtifactFormat): Promise<string> {
  if (format === 'html' && /\.html?$/i.test(file.name)) {
    return htmlUrl(await file.text());
  }

  return mockPage(title, [`Arquivo publicado: ${file.name}.`, 'No Go, ele é servido de verdade.']);
}

function newArtifact(label: string, form: FormData, format: ArtifactFormat): Artifact {
  const title = formText(form, 'title');
  const kind = artifactKindSchema.safeParse(form.get('kind'));
  if (!title || !kind.success) {
    throw new ApiError(400, 'Artefato novo precisa de título e tipo (title e kind).');
  }

  const now = new Date().toISOString();
  const artifact: Artifact = {
    id: `artifact-${crypto.randomUUID()}`,
    label,
    title,
    description: null,
    kind: kind.data,
    format,
    status: null,
    isPrivate: false,
    isPinned: false,
    url: '',
    thumbUrl: null,
    currentVersion: 0,
    sizeBytes: 0,
    tags: [],
    createdAt: now,
    updatedAt: now,
    lastViewedAt: null,
  };
  db.artifacts.push(artifact);
  db.versions[artifact.id] = [];

  return artifact;
}

type PublishBody = { label: string; form: FormData; file: File; format: ArtifactFormat };

/** Confere o PUT na mesma ordem e com as mesmas mensagens do Go. */
function readPublishBody(params: Record<string, string>, body: unknown): PublishBody {
  const label = artifactLabelSchema.safeParse(params.label);
  if (!label.success) {
    throw new ApiError(400, 'Use só letras minúsculas, números e hífen na label (até 80).');
  }

  const file = body instanceof FormData ? body.get('file') : null;
  if (!(body instanceof FormData) || !(file instanceof File)) {
    throw new ApiError(400, 'Envie o arquivo no campo file.');
  }

  const format = formatOf(file.name);
  if (!format) throw new ApiError(400, 'Envie um arquivo .html, .zip, .md ou .pdf.');

  return { label: label.data, form: body, file, format };
}

/** Campo que não veio mantém o valor — o "nil é não veio" do Go. */
function publishedFields(artifact: Artifact, form: FormData): Partial<Artifact> {
  const tags = formText(form, 'tags') ?? '';

  return {
    title: formText(form, 'title') ?? artifact.title,
    description: form.has('description') ? formText(form, 'description') : artifact.description,
    tags: form.has('tags')
      ? tags
          .split(',')
          .map((tag) => tag.trim())
          .filter(Boolean)
      : artifact.tags,
  };
}

/** O PUT de publicação: cria o artefato ou sobe a versão seguinte, como o Go. */
async function publishArtifact({ params, body }: MockContext) {
  const { label, form, file, format } = readPublishBody(params, body);
  const artifact =
    db.artifacts.find((item) => item.label === label) ?? newArtifact(label, form, format);
  const versions = db.versions[artifact.id] ?? [];
  const version = Math.max(0, ...versions.map((item) => item.version)) + 1;
  const now = new Date().toISOString();

  Object.assign(artifact, publishedFields(artifact, form), {
    format,
    currentVersion: version,
    sizeBytes: file.size,
    updatedAt: now,
  });
  artifact.url = await servedUrlOf(file, artifact.title, format);
  versions.unshift({ version, sizeBytes: file.size, hasThumb: false, createdAt: now });
  db.versions[artifact.id] = versions;

  return artifactSchema.parse(artifact);
}

function restoreArtifact({ params, body }: MockContext) {
  const artifact = findArtifact(params.label ?? '');
  const restore = artifactRestoreSchema.safeParse(body);
  if (!restore.success) throw new ApiError(400, 'Informe a versão a restaurar.');
  const exists = (db.versions[artifact.id] ?? []).some(
    (item) => item.version === restore.data.version,
  );
  if (!exists) throw new ApiError(404, 'Versão não encontrada.');

  artifact.currentVersion = restore.data.version;
  artifact.updatedAt = new Date().toISOString();

  return artifactSchema.parse(artifact);
}

function deleteArtifact({ params }: MockContext) {
  const artifact = findArtifact(params.label ?? '');
  db.artifacts = db.artifacts.filter((item) => item.id !== artifact.id);
  delete db.versions[artifact.id];

  return undefined;
}

function readAppInput(body: unknown): AppLinkInput {
  const input = appLinkInputSchema.safeParse(body);
  if (!input.success)
    throw new ApiError(400, input.error.issues[0]?.message ?? 'Cadastro inválido.');

  return input.data;
}

/** O tipo pelo nome, sem diferença de maiúscula — o `upsert_app_type.sql` do Go. */
function typeIdFor(name: string): string {
  const existing = db.appHub.types.find((type) => type.name.toLowerCase() === name.toLowerCase());
  if (existing) return existing.id;

  const position = Math.max(-1, ...db.appHub.types.map((type) => type.position)) + 1;
  const created = { id: `type-${crypto.randomUUID()}`, name, position };
  db.appHub.types.push(created);

  return created.id;
}

/** O tipo que ficou sem nenhum app some, como no Go. */
function dropEmptyTypes(): void {
  db.appHub.types = db.appHub.types.filter((type) =>
    db.appHub.apps.some((app) => app.typeId === type.id),
  );
}

function findApp(id: string): AppLink {
  const app = db.appHub.apps.find((item) => item.id === id);
  if (!app) throw new ApiError(404, 'App não encontrado. Ele pode ter sido apagado.');

  return app;
}

function createApp({ body }: MockContext) {
  const input = readAppInput(body);
  const app: AppLink = {
    id: `app-${crypto.randomUUID()}`,
    typeId: typeIdFor(input.typeName),
    name: input.name,
    url: input.url,
    logoUrl: null,
    status: 'unknown',
    checkedAt: null,
  };
  db.appHub.apps.push(app);

  return appLinkSchema.parse(app);
}

function updateApp({ params, body }: MockContext) {
  const app = findApp(params.id ?? '');
  const input = readAppInput(body);
  const isNewAddress = app.url !== input.url;

  Object.assign(app, {
    typeId: typeIdFor(input.typeName),
    name: input.name,
    url: input.url,
    status: isNewAddress ? 'unknown' : app.status,
    checkedAt: isNewAddress ? null : app.checkedAt,
  });
  dropEmptyTypes();

  return appLinkSchema.parse(app);
}

function deleteApp({ params }: MockContext) {
  const app = findApp(params.id ?? '');
  db.appHub.apps = db.appHub.apps.filter((item) => item.id !== app.id);
  dropEmptyTypes();

  return undefined;
}

const ROUTES: MockRoute[] = [
  { method: 'GET', path: '/auth/me', handle: mockMe },
  { method: 'GET', path: '/auth/setup', isPublic: true, handle: mockSetup },
  {
    method: 'POST',
    path: '/auth/mock/set-password',
    isPublic: true,
    handle: ({ body }) => mockSetPassword(body),
  },
  {
    method: 'POST',
    path: '/auth/mock/sign-in',
    isPublic: true,
    handle: ({ body }) => mockSignIn(body),
  },
  {
    method: 'POST',
    path: '/auth/mock/sign-out',
    isPublic: true,
    handle: () => setMockSession(null),
  },
  { method: 'GET', path: '/artifacts', handle: listArtifacts },
  { method: 'GET', path: '/artifacts/:label', handle: getArtifact },
  { method: 'PATCH', path: '/artifacts/:label', handle: patchArtifact },
  { method: 'PUT', path: '/artifacts/:label', handle: publishArtifact },
  { method: 'POST', path: '/artifacts/:label/restore', handle: restoreArtifact },
  { method: 'DELETE', path: '/artifacts/:label', handle: deleteArtifact },
  { method: 'GET', path: '/apps', handle: () => appHubSchema.parse(db.appHub) },
  { method: 'POST', path: '/apps', handle: createApp },
  { method: 'PATCH', path: '/apps/:id', handle: updateApp },
  { method: 'DELETE', path: '/apps/:id', handle: deleteApp },
];

/** Casa `/artifacts/abertura-cnpj` com `/artifacts/:label` e devolve `{ label }`. */
export function matchPath(pattern: string, path: string): Record<string, string> | null {
  const expected = pattern.split('/');
  const actual = path.split('/');
  if (expected.length !== actual.length) return null;

  const params: Record<string, string> = {};
  for (const [index, part] of expected.entries()) {
    const value = actual[index] ?? '';
    if (part.startsWith(':')) params[part.slice(1)] = decodeURIComponent(value);
    else if (part !== value) return null;
  }

  return params;
}

function toQuery(query: RequestOptions['query']): URLSearchParams {
  const params = new URLSearchParams();
  for (const [key, value] of Object.entries(query ?? {})) {
    if (value === undefined || value === null || value === '') continue;
    params.set(key, String(value));
  }

  return params;
}

/**
 * Responde como o Go responderia: o corpo já no formato do contrato, ou um `ApiError` com
 * status e mensagem em português. Rota que não existe é 404 — igual ao Go hoje.
 */
export async function handleMockRequest<TResponse>(
  path: string,
  options: RequestOptions = {},
  latencyMs: number = MOCK_LATENCY_MS,
): Promise<TResponse> {
  if (latencyMs > 0) await new Promise((resolve) => setTimeout(resolve, latencyMs));

  const method = options.method ?? 'GET';
  for (const route of ROUTES) {
    if (route.method !== method) continue;
    const params = matchPath(route.path, path);
    if (!params) continue;
    if (!route.isPublic && !hasMockSession()) throw new ApiError(401, 'Entre para continuar.');

    /* Cópia: a tela não pode mexer no "banco" segurando a referência da resposta. */
    const result = await route.handle({
      db,
      params,
      query: toQuery(options.query),
      body: options.body,
    });
    return structuredClone(result) as TResponse;
  }

  throw new ApiError(404, `Rota de mock não encontrada: ${method} ${path}`);
}

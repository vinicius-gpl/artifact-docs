import { appHubSchema } from '$lib/contracts/app-link.schema';
import {
  artifactDetailSchema,
  artifactKindSchema,
  artifactListSchema,
  artifactPatchSchema,
  artifactSchema,
  type Artifact,
} from '$lib/contracts/artifact.schema';
import { paletteListSchema } from '$lib/contracts/palette.schema';
import { sessionUserSchema } from '$lib/contracts/user.schema';

import { ApiError, type RequestOptions } from './http-client';
import { createMockDatabase, MOCK_USER, type MockDatabase } from './mock-fixtures';

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

const ROUTES: MockRoute[] = [
  { method: 'GET', path: '/auth/me', handle: () => sessionUserSchema.parse(MOCK_USER) },
  { method: 'GET', path: '/artifacts', handle: listArtifacts },
  { method: 'GET', path: '/artifacts/:label', handle: getArtifact },
  { method: 'PATCH', path: '/artifacts/:label', handle: patchArtifact },
  { method: 'GET', path: '/apps', handle: () => appHubSchema.parse(db.appHub) },
  {
    method: 'GET',
    path: '/palettes',
    handle: () => paletteListSchema.parse({ items: db.palettes }),
  },
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

    /* Cópia: a tela não pode mexer no "banco" segurando a referência da resposta. */
    const result = route.handle({ db, params, query: toQuery(options.query), body: options.body });
    return structuredClone(result) as TResponse;
  }

  throw new ApiError(404, `Rota de mock não encontrada: ${method} ${path}`);
}

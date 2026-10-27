import { type AppHub } from '$lib/contracts/app-link.schema';
import { type Artifact, type ArtifactVersion } from '$lib/contracts/artifact.schema';
import { type SessionUser } from '$lib/contracts/user.schema';

import { MOCK_CREDENTIALS } from './auth.api';

/**
 * OS DADOS DE MENTIRA do modo mock (`VITE_API_MOCKS=true`) — inventados, nunca dado real.
 *
 * Cada coleção tem o MESMO formato que o Go vai devolver (os tipos vêm de `lib/contracts`, e o
 * `mock-server` ainda passa tudo pelo schema antes de responder). Quando a rota existir no Go,
 * nada aqui precisa mudar na tela: é só desligar o modo mock.
 *
 * Os artefatos não têm arquivo servido de verdade: o `url` é um `data:` com uma página HTML
 * mínima, para o preview e o visualizador funcionarem igual funcionariam com `/a/<token>/`.
 */

/**
 * A senha do usuário de teste. É valor de teste, só do modo mock: em produção quem confere a
 * senha é o Neon Auth, e esta rota nem existe no Go.
 */
export const MOCK_PASSWORD = MOCK_CREDENTIALS.password;

/** O usuário de mentira do modo mock — é com ele que o login de localhost entra. */
export const MOCK_USER: SessionUser = {
  id: 'user-mock',
  email: MOCK_CREDENTIALS.email,
  name: 'Usuário de teste',
  image: null,
};

/** Uma página mínima, no estilo da marca, para servir de artefato publicado. */
export function mockPage(title: string, lines: string[]): string {
  const body = lines.map((line) => `<p>${line}</p>`).join('');
  const html = `<!doctype html><html lang="pt-BR"><head><meta charset="utf-8"><title>${title}</title>
<style>body{margin:0;padding:48px;font-family:'IBM Plex Sans',system-ui,sans-serif;background:#f4f4f3;color:#141416}
h1{font-weight:400;font-size:40px;margin:0 0 16px;border-left:8px solid #ffa58a;padding-left:16px}
p{font-size:18px;line-height:1.6;color:#626267;max-width:60ch}</style></head>
<body><h1>${title}</h1>${body}</body></html>`;

  return `data:text/html;charset=utf-8,${encodeURIComponent(html)}`;
}

type ArtifactSeed = Pick<Artifact, 'label' | 'title' | 'kind' | 'format'> &
  Partial<Omit<Artifact, 'label' | 'title' | 'kind' | 'format'>> & {
    /** Há quantos dias foi editado — as datas saem relativas a "agora". */
    editedDaysAgo: number;
    /** Há quantos dias foi aberto; sem ele, nunca foi aberto. */
    viewedDaysAgo?: number;
    lines?: string[];
  };

const DAY_MS = 24 * 60 * 60 * 1000;

function daysAgo(days: number, now: number): string {
  return new Date(now - days * DAY_MS).toISOString();
}

function artifactFrom(seed: ArtifactSeed, index: number, now: number): Artifact {
  const { editedDaysAgo, viewedDaysAgo, lines, ...fields } = seed;

  return {
    id: `artifact-${index + 1}`,
    description: null,
    status: null,
    isPrivate: false,
    isPinned: false,
    url: mockPage(seed.title, lines ?? [fields.description ?? 'Conteúdo de exemplo.']),
    thumbUrl: null,
    currentVersion: 1,
    sizeBytes: 12_000 + index * 3_517,
    tags: [],
    createdAt: daysAgo(editedDaysAgo + 30, now),
    updatedAt: daysAgo(editedDaysAgo, now),
    lastViewedAt: viewedDaysAgo === undefined ? null : daysAgo(viewedDaysAgo, now),
    ...fields,
  };
}

const ARTIFACT_SEEDS: ArtifactSeed[] = [
  {
    label: 'briefing-artifact-docs',
    title: 'Briefing do artifact-docs',
    kind: 'document',
    format: 'md',
    description: 'O que é o hub, os módulos e a stack.',
    isPinned: true,
    currentVersion: 4,
    tags: ['hub', 'arquitetura'],
    editedDaysAgo: 0,
    viewedDaysAgo: 0,
    lines: ['Hub pessoal num lugar só: documentação, artefatos, planejamentos e ideias.'],
  },
  {
    label: 'acerola-reader-docs',
    title: 'Documentação do acerola-reader',
    kind: 'document',
    format: 'html',
    description: 'Guia de uso e referência da API do leitor.',
    currentVersion: 7,
    tags: ['acerola'],
    editedDaysAgo: 2,
  },
  {
    label: 'manual-da-marca',
    title: 'Manual da marca Grupo Azuos',
    kind: 'document',
    format: 'pdf',
    description: 'Logo, cores, tipografia e aplicações.',
    isPinned: true,
    isPrivate: true,
    tags: ['marca'],
    editedDaysAgo: 12,
    viewedDaysAgo: 4,
  },
  {
    label: 'relatorio-custos-vps',
    title: 'Relatório de custos da VPS',
    kind: 'document',
    format: 'html',
    description: 'Consumo de RAM, banda e o que cabe em 1 GB.',
    isPrivate: true,
    currentVersion: 2,
    tags: ['infra'],
    editedDaysAgo: 5,
  },
  {
    label: 'abertura-cnpj',
    title: 'Abertura do CNPJ',
    kind: 'plan',
    format: 'md',
    status: 'active',
    description: 'Natureza jurídica, contador, prazos e custos.',
    isPinned: true,
    isPrivate: true,
    currentVersion: 3,
    editedDaysAgo: 1,
    viewedDaysAgo: 1,
  },
  {
    label: 'erp-pessoal',
    title: 'ERP pessoal',
    kind: 'plan',
    format: 'md',
    status: 'draft',
    description: 'Finanças, contratos e notas num lugar só.',
    editedDaysAgo: 9,
  },
  {
    label: 'custo-iam',
    title: 'Custo de IAM',
    kind: 'plan',
    format: 'html',
    status: 'done',
    description: 'Comparativo entre Neon Auth, Clerk e login próprio.',
    currentVersion: 5,
    editedDaysAgo: 21,
  },
  {
    label: 'tunel-wireguard',
    title: 'Túnel WireGuard',
    kind: 'plan',
    format: 'md',
    status: 'active',
    description: 'Acesso aos apps de casa a partir da VPS.',
    isPrivate: true,
    editedDaysAgo: 3,
  },
  {
    label: 'gerador-de-paletas',
    title: 'Gerador de paletas em OKLCH',
    kind: 'idea',
    format: 'md',
    status: 'active',
    description: 'Subir o CSS base e receber as escalas 50–950.',
    tags: ['marca'],
    editedDaysAgo: 4,
  },
  {
    label: 'docs-por-commit',
    title: 'Documentação republicada a cada commit',
    kind: 'idea',
    format: 'md',
    status: 'draft',
    description: 'A skill roda no CI e atualiza o artefato sozinha.',
    editedDaysAgo: 15,
  },
  {
    label: 'galeria-publica',
    title: 'Galeria pública de projetos',
    kind: 'idea',
    format: 'html',
    status: 'archived',
    description: 'Vitrine com os artefatos marcados como públicos.',
    editedDaysAgo: 40,
  },
];

const APP_HUB_SEED: AppHub = {
  types: [
    { id: 'type-self-hosted', name: 'Self-hosted', position: 0 },
    { id: 'type-vps', name: 'Na VPS', position: 1 },
    { id: 'type-tools', name: 'Ferramentas', position: 2 },
  ],
  apps: [
    {
      id: 'app-1',
      typeId: 'type-self-hosted',
      name: 'Jellyfin',
      url: 'https://jellyfin.exemplo.com',
      logoUrl: null,
      status: 'online',
      checkedAt: null,
    },
    {
      id: 'app-2',
      typeId: 'type-self-hosted',
      name: 'Home Assistant',
      url: 'https://casa.exemplo.com',
      logoUrl: null,
      status: 'online',
      checkedAt: null,
    },
    {
      id: 'app-3',
      typeId: 'type-self-hosted',
      name: 'Nextcloud',
      url: 'https://nuvem.exemplo.com',
      logoUrl: null,
      status: 'offline',
      checkedAt: null,
    },
    {
      id: 'app-4',
      typeId: 'type-vps',
      name: 'artifact-docs',
      url: 'https://hub.exemplo.com',
      logoUrl: null,
      status: 'online',
      checkedAt: null,
    },
    {
      id: 'app-5',
      typeId: 'type-vps',
      name: 'Uptime Kuma',
      url: 'https://status.exemplo.com',
      logoUrl: null,
      status: 'online',
      checkedAt: null,
    },
    {
      id: 'app-6',
      typeId: 'type-tools',
      name: 'Neon',
      url: 'https://console.neon.tech',
      logoUrl: null,
      status: 'unknown',
      checkedAt: null,
    },
    {
      id: 'app-7',
      typeId: 'type-tools',
      name: 'Cloudflare R2',
      url: 'https://dash.cloudflare.com',
      logoUrl: null,
      status: 'unknown',
      checkedAt: null,
    },
  ],
};

export type MockDatabase = {
  artifacts: Artifact[];
  versions: Record<string, ArtifactVersion[]>;
  appHub: AppHub;
};

/** Um banco novo, com as datas relativas a `now` — o teste passa um `now` fixo. */
export function createMockDatabase(now: number = Date.now()): MockDatabase {
  const artifacts = ARTIFACT_SEEDS.map((seed, index) => artifactFrom(seed, index, now));
  const versions = Object.fromEntries(
    artifacts.map((artifact) => [artifact.id, versionsOf(artifact)]),
  );

  return { artifacts, versions, appHub: structuredClone(APP_HUB_SEED) };
}

/** Uma versão por número até a atual, uma semana entre cada uma, a mais nova primeiro. */
function versionsOf(artifact: Artifact): ArtifactVersion[] {
  const updated = new Date(artifact.updatedAt).getTime();

  return Array.from({ length: artifact.currentVersion }, (_, index) => {
    const version = artifact.currentVersion - index;

    return {
      version,
      sizeBytes: artifact.sizeBytes - index * 512,
      hasThumb: false,
      createdAt: new Date(updated - index * 7 * DAY_MS).toISOString(),
    };
  });
}

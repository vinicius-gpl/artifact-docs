# Mocks da API — a tela antes do Go

A SPA já funciona inteira sem o backend: com `VITE_API_MOCKS=true`, toda chamada a `/api` é
respondida por `artifact-docs/svelte/src/lib/api/mock-server.ts`, com dados de mentira em memória
(`mock-fixtures.ts`). O `mise run dev` liga o modo mock por padrão.

```bash
mise run dev                         # tela com mocks (padrão)
VITE_API_MOCKS=false mise run dev    # tela falando com o Go de verdade
```

## Como trocar pelo Go

O desvio para o mock está num lugar só: `apiRequest`, em `lib/api/http-client.ts`. Hooks, telas
e os `*.api.ts` não sabem que o mock existe. Para trocar uma rota:

1. Implemente a rota no módulo Go da feature (`artifact-docs/src-go/modules/<feature>/`) e ligue-a
   em `src-go/app/server/server.go`, devolvendo **exatamente** o schema
   Zod de `lib/contracts/` — os dois lados precisam combinar.
2. Rode com `VITE_API_MOCKS=false` e confira a tela.
3. Quando todas as rotas existirem, apague `mock-server.ts`, `mock-fixtures.ts` e o desvio no
   `http-client.ts`.

## O que o Go precisa implementar

A tabela `ROUTES` do `mock-server.ts` é a lista viva. Hoje:

| Método | Rota | Resposta (schema) | No Go | Observação |
|---|---|---|---|---|
| `GET` | `/api/auth/config` | `{ neonAuthUrl }` | ✅ | pública; a tela só usa quando a URL não veio no build |
| `GET` | `/api/auth/setup` | `setupStatusSchema` | ✅ | pública; ver [`auth.md`](./auth.md) |
| `GET` | `/api/auth/me` | `sessionUserSchema` | ✅ | 401 sem token, 403 fora da lista permitida |
| `GET` | `/api/artifacts?kind=&q=` | `artifactListSchema` | ✅ | mais recente primeiro; `q` é a busca `tsvector` |
| `GET` | `/api/artifacts/{label}` | `artifactDetailSchema` | ✅ | grava `last_viewed_at` (é o que ordena "Recentes") |
| `PATCH` | `/api/artifacts/{label}` | `artifactSchema` | ✅ | corpo `artifactPatchSchema` (fixar, privado, situação) |
| `PUT` | `/api/artifacts/{label}` | `artifactSchema` | ✅ | multipart (`artifactPublishSchema`): cria ou sobe versão; 201 quando é novo |
| `POST` | `/api/artifacts/{label}/restore` | `artifactSchema` | ✅ | corpo `artifactRestoreSchema`; o rollback só troca o ponteiro |
| `DELETE` | `/api/artifacts/{label}` | 204 | ✅ | apaga o registro e os arquivos no R2 |
| `GET` | `/api/apps` | `appHubSchema` | ✅ | tipos e apps juntos; a tela agrupa |

Com tudo ✅, o próximo passo é conferir a tela com `VITE_API_MOCKS=false` contra um banco de
verdade e, dando certo, apagar o mock (passo 3 acima). Sem `DATABASE_URL` ou
`NEON_AUTH_JWKS_URL`, o Go sobe assim mesmo e as rotas que dependem deles respondem 503.

As rotas `/api/auth/mock/*` só existem no mock (login, primeiro acesso e sair de mentira); em
produção isso acontece no Neon Auth.

Erro segue o formato que o `http-client` já lê: status HTTP e corpo
`{ "message": "texto em português", "details": [] }`.

## Limites do mock

- Os dados voltam ao começo a cada recarga da página — de propósito, para não virar um segundo banco.
- O `url` de cada artefato é um `data:` com uma página mínima (ou, ao publicar um `.html` pelo
  mock, o próprio HTML enviado). Markdown, `.zip` e PDF publicados no mock viram uma página
  dizendo o que foi enviado: quem converte e descompacta é o Go. O preview e o visualizador funcionam;
  "Abrir em nova aba" não, porque o navegador bloqueia abrir `data:` direto. Com o Go, o link é
  `/a/<token>/` e abre normalmente.

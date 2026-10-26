# CLAUDE.md

Instruções para o Claude Code neste repositório. Leia inteiro antes da primeira mudança.

## O projeto

Hub pessoal: central de documentos e artefatos HTML/MD/PDF com galeria e preview, planejamentos,
banco de ideias, hub de apps e casa da marca. O porquê e o como estão em
[`docs/briefing.md`](./docs/briefing.md).

- **Stack:** SPA SvelteKit (`artifact-docs/svelte/`, `adapter-static`) embutida num binário Go (`artifact-docs/main.go` +
  `artifact-docs/src-go/`). Neon (Postgres) para dados e login, R2 (API S3) para arquivos, Caddy na frente.
- **RAM é o limite:** a VPS tem 1 GB. Nada de Node em produção, biblioteca padrão do Go
  primeiro, nada de processo extra no servidor.
- **A infra do front foi copiada do acerola-ticket** — libs, versões, componentes, tokens e
  regras. Melhoria lá pode ser trazida para cá, e vice-versa.
- Quem pede é o dono do projeto: decida os detalhes técnicos pelo CONTRIBUTING e pelas skills
  e explique em uma frase. Só pergunte o que não dá para decidir.

## Git

Siga a skill **`git-flow`**:

1. Toda mudança nasce numa `feature/<nome>` (ou `bugfix/<nome>`) criada a partir da `develop`.
2. O trabalho é commitado nessa branch (skill **`git-commit`**, formato `[tipo](local): Mensagem`).
3. No fim: trazer a `develop` para a branch, enviar a branch e abrir o PR para a `develop` com
   título `[merge](local): …`.
4. **Teste não roda na máquina — roda no CI do PR** (`.github/workflows/tests.yml`). Acompanhe
   com `gh pr checks <n> --watch`; vermelho → corrija na branch e envie de novo.
5. Só com CI verde e OK explícito: merge na `develop`.
6. A `main` só recebe release/hotfix, em Modo Admin (`git config project.admin true`). A trava
   (`scripts/git/branch-guard.sh`) recusa o resto. **Nunca** ative o Modo Admin por conta própria.

## As regras de código

O padrão está em **`CONTRIBUTING.md`**, e ele manda. Os pontos que mais quebram:

1. **Idioma:** o usuário vê → português. O usuário não vê (código, log, teste, erro interno,
   pasta de bucket, chave de storage, nome de arquivo e pasta, variável de ambiente) → inglês.
   Comentário em português, explicando o **porquê**.
2. **Early return.** Nunca `if/else` alinhado. Complexidade máxima 10 (ESLint e `gocyclo`).
3. **MVVM:** rota (`+page.svelte`) só compõe · hook (`lib/hooks/use-*/use-*.svelte.ts`) tem
   estado e dados, sem marcação · componente de UI é função pura de props.
4. **Props em quatro grupos:** `data`, `ui`, `state`, `actions`.
5. **`lib/components/ui/` e `lib/hooks/ui/` não se editam** (são do CLI do shadcn). Envolva num
   `acerola-*`. Existe no shadcn-svelte? Instale, não recrie. **Nunca reescreva um componente
   existente.** Todo componente próprio tem prefixo `acerola-*`. Todo campo de formulário tem
   `h-10`. Detalhes: skill **`design-system`**.
6. **Todo componente tem `.stories.svelte`.**
7. **Todo código com lógica tem teste do caminho feliz e do triste** (`// feliz`, `// triste`),
   no Vitest e no `go test`.
8. **Contrato:** o schema Zod em `artifact-docs/svelte/src/lib/contracts/` valida o formulário; o lado Go mora
   em `artifact-docs/src-go/`. Os dois precisam combinar — mudou um, muda o outro.
9. **Export nomeado. Sem barril (`index.ts`).**

## Use as skills

| Pedido | Skill | Atalho pt-BR |
|---|---|---|
| **Criar, mover ou renomear qualquer arquivo**; "isso está no lugar certo?" | `design-system` | `/sistema-de-design` |
| Componente visual novo ou alterado | `ui-component` (e `ui-standards` como referência) | `/componente-ui` |
| Desenhar ou revisar tela, texto de tela | `ui-standards` | `/ui-padrao` |
| Cor, fonte, forma, "segue a marca?" (padrão Carbon, acabamento Vidro) | `brand-design` | `/design-da-marca` |
| Quantidade de opções, tamanho de alvo, agrupamento, fluxo | `ux-laws` | `/leis-de-ux` |
| "Salva", "commita" | `git-commit` | `/commitar` |
| **Qualquer mudança de código** (começar, terminar, juntar, enviar) | `git-flow` | `/git-fluxo` |
| Conflito de merge | `resolve-conflict` | `/resolver-conflito` |

## Onde fica cada coisa

```
artifact-docs/                                  o app (como acerola/dashboard no acerola-ticket)
├── main.go                                     sobe o processo e embute a SPA
├── mise.toml                                   tarefas do app (dev, build, Go, web)
├── src-go/<pacote>/                            Go (config, server, web…)
├── src-go/web/dist/                            cópia da SPA compilada (gerada)
└── svelte/src/
    ├── routes/(app)/<feature>/                 tela (só composição): artifacts, plans, ideas, apps, brand
    ├── routes/(app)/<feature>/components/acerola-<nome>/  componente da feature
    ├── lib/hooks/use-<nome>/                   estado e dados das telas
    ├── lib/components/acerola-<nome>/          componente genérico
    ├── lib/components/ui/                      ⛔ CLI do shadcn (lib/hooks/ui/ idem)
    ├── lib/contracts/                          schemas Zod
    └── lib/navigation/navigation.ts            menu lateral — tela nova entra com uma linha
docs/                                           briefing e decisões
.claude/ .github/ .githooks/ scripts/git/       infra do repositório (raiz)
mise.toml                                       ferramentas, setup, design e atalhos para o app
```

A casca (`acerola-app-shell`) é uma só, montada em `routes/(app)/+layout.svelte`, que também
guarda a sessão. A margem do conteúdo é dela: a tela define só a largura máxima
(`mx-auto w-full max-w-6xl`).

## Comandos

Tudo passa pelo **mise**. Da raiz, `dev`, `build` e `check` são atalhos para as tarefas de
`artifact-docs/mise.toml`; dentro de `artifact-docs/` estão todas as do app.

```bash
mise run setup          # uma vez: hooks do git, Git-Flow e dependências do front
mise run dev            # API Go :3010 + SPA :5010 (em segundo plano: não termina)
mise run check          # tudo o que o CI roda — para quem quiser à mão; o Claude não roda (ver Git)
mise run design:check   # regras de design — só reprova violação nova
mise run brand:check    # regras da marca e das leis de UX (kit de design)
mise run build          # binário único em artifact-docs/bin/artifact-docs
mise tasks              # lista tudo
```

Para conferir a tela, use o navegador em http://localhost:5010. O `dev` sobe a tela com o
**modo mock** ligado: as rotas que o Go ainda não tem são respondidas por
`svelte/src/lib/api/mock-server.ts` (ver [`docs/api-mocks.md`](./docs/api-mocks.md)). Rota nova
nasce no mock e no contrato Zod; quando o Go a implementa, o mock dela sai.

## O que SEMPRE pede confirmação antes

- `git merge --no-ff` na develop — só depois do OK de que está funcionando.
- `git push`, abrir PR, qualquer coisa que saia da máquina.
- `git reset --hard`, `git checkout -- .`, `git clean`, apagar branch — perde trabalho.
- Instalar dependência nova, npm ou Go (diga qual e por quê em uma frase).

## O que NUNCA fazer

- Commit direto na `main` ou na `develop`; `--no-verify`; ativar o Modo Admin sem pedido.
- `npm install --force` ou `--legacy-peer-deps`.
- Editar `lib/components/ui/`, `lib/hooks/ui/` ou `artifact-docs/src-go/web/dist/`.
- Reescrever componente existente além do que foi pedido, ou recriar à mão componente que o
  shadcn-svelte já tem.
- Colocar segredo em variável `VITE_` ou em arquivo versionado.
- Rodar suíte de teste na máquina (`mise run check`, `web:test`, `go test`…) — quem verifica é
  o CI do PR.
- Dizer que terminou sem o CI do PR verde.

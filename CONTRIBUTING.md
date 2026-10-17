# Contribuindo

Quando a regra for verificável por máquina, ela é hook ou gate de CI.

> **Trabalhando com o Claude?** O `CLAUDE.md` manda o Claude seguir este documento, e as skills em `.claude/skills/` fazem o passo a passo de cada tarefa comum. Este arquivo é a referência para quando alguém — pessoa ou Claude — tiver dúvida sobre o porquê.

## Preparando a máquina

Depois de clonar, uma vez:

```sh
mise install       # Node, Go e golangci-lint nas versões do mise.toml
mise run setup     # hooks do git, template de commit, Git-Flow e dependências do front
```

---

## 1. Idioma

A régua é uma só: **o usuário vê, é português. O usuário não vê, é inglês.** Comentário é a única exceção — ele é português mesmo estando no código.

| O quê                                                                                  | Idioma                                                        |
| -------------------------------------------------------------------------------------- | ------------------------------------------------------------- |
| Nome de variável, função, classe, tipo, arquivo, pasta, branch, tabela, coluna, rota   | **Inglês**                                                    |
| Pasta/prefixo de bucket (R2/S3), chave de storage, nome de fila, variável de ambiente  | **Inglês**                                                    |
| Nome de arquivo de documentação (`docs/architecture.md`)                               | **Inglês** (conteúdo em pt-BR)                                |
| Nome de arquivo baixado pelo usuário (`apps.csv`)                                  | **Português (pt-BR)**                                         |
| Nome de skill (`.claude/skills/<name>/`); atalho pt-BR em `.claude/commands/`          | **Inglês** (atalho em pt-BR: `/socorro` → `troubleshoot`)     |
| `console.log`, `Logger`, mensagem de log, nome de métrica                              | **Inglês**                                                    |
| `throw new Error(...)` interno, mensagem de exceção, texto de `assert`                 | **Inglês**                                                    |
| `describe` / `it` de teste, nome de story do Storybook                                 | **Inglês**                                                    |
| Comentário no código                                                                   | **Português (pt-BR)**                                         |
| Texto na tela: rótulo, título, botão, placeholder, `aria-label`, `title`, estado vazio | **Português (pt-BR)**                                         |
| Mensagem de erro mostrada ao usuário                                                   | **Português (pt-BR)**                                         |
| Mensagem de commit e descrição de PR                                                   | **Português (pt-BR)**                                         |
| Documentação (`.md`)                                                                   | **Português (pt-BR)**, com identificadores técnicos em inglês |

O teste que resolve qualquer dúvida: **esse texto pode aparecer na tela de quem usa o sistema?** Se pode, é português. Se ele só existe para quem programa — log, stack trace, nome de teste, chave de objeto —, é inglês.

Comentário explica **por quê**, nunca **o quê**.

---

## 2. Early return — nunca `if/else` alinhado

Sem exceção. Condição que não se aplica sai da função na hora.

Guarda primeiro, caminho feliz por último e sem indentação.

No template Svelte, mostre cedo os estados de carregamento, erro e vazio com `{#if}`, e deixe o conteúdo principal por último, sem aninhamento.

**Lint:** `no-else-return`, `complexity: ['error', 10]`, `max-depth: ['error', 2]`.

---

## 3. MVVM — hook nunca encosta em template

A camada de apresentação não sabe de onde o dado vem. O hook não sabe como o dado é desenhado. Os dois se encontram na rota.

```text
routes/(app)/artifacts/+page.svelte  → compõe: chama o hook, passa para o componente
lib/hooks/use-…/use-….svelte.ts   → estado, efeitos, queries, mutations, handlers. ZERO template.
lib/components/…/….svelte         → função pura de props → template. ZERO hook de dado.
```

Permitido dentro de um componente de UI:

* `$props()`;
* `$derived()` de cálculo puro de apresentação;
* `$state()` de estado puramente visual, como mostrar/esconder senha.

Proibido:

* `createQuery`, `createMutation`, `createInfiniteQuery`, `createQueries`;
* `goto` (navegação direta — vem por `actions`, montada no hook);
* `$effect` que busca dado.

**Lint:** `no-restricted-syntax` recusa `createQuery`/`createMutation`/`goto` em `lib/components/**` (fora de `lib/components/ui/`) e cobra export nomeado em `lib/hooks/**`.

---

## 4. Props semanticamente separadas

Toda prop entra em um de quatro grupos. Nada de prop solta na raiz.

| Grupo     | Contém                                                            | Nunca contém         |
| --------- | ----------------------------------------------------------------- | -------------------- |
| `data`    | Valores de domínio a exibir                                       | Callback, classe CSS |
| `ui`      | Só aparência: `variant`, `size`, `tone`, `align`, `className`     | Dado, callback       |
| `state`   | Estado de interação: `isLoading`, `isDisabled`, `isOpen`, `error` | Dado de domínio      |
| `actions` | Só callbacks, sempre `on*`                                        | Qualquer outra coisa |

A exceção é `children` (`Snippet`), que fica na raiz — é como o Svelte compõe.

Booleano de estado usa prefixo `is`/`has`. Callback usa prefixo `on`. O handler que o view-model expõe usa o mesmo nome do callback que consome.

---

## 5. Componentes: baixado, genérico, de feature

```text
lib/components/ui/                                ⛔ CLI do shadcn-svelte. NÃO EDITE.
lib/hooks/ui/                                     ⛔ hook que o CLI gera ("hooks": "$lib/hooks/ui" no components.json).
lib/components/acerola-<nome>/                    ✅ genérico: envolve o ui, sem domínio, usado por 2+ features.
routes/(app)/<context>/<feature>/components/acerola-<nome>/ ✅ da feature: conhece o domínio, só a própria rota importa.
```

Regras:

1. **Prefixo `acerola-*` em todo componente próprio**, genérico ou de feature.
2. **Existe no shadcn-svelte? Instale** (`npx shadcn-svelte@latest add <nome>`, dentro de
   `svelte/`) e envolva num `acerola-*` com `tv()`. Recriar à mão é proibido; copiar o markup do
   baixado para editar também.
3. **Rodar o `add` sobrescreve `ui/`** — e é para sobrescrever mesmo. Por isso nada nosso mora lá.
4. **Nunca reescrever componente existente.** Mudança limitada ao que foi pedido.
5. **Campo de formulário tem `h-10`**, na variante base do `acerola-*`. Nem mais, nem menos. A
   tela não compensa altura.
6. **Só `lib/components/acerola-*` importa `lib/components/ui/`.** Componente de feature importa
   os `acerola-*` genéricos. Feature não importa componente de outra feature.
7. **A rota só compõe**: hook + componente + classes de cor, espaço e layout. UI/UX de componente
   na rota é violação.

**Lint:** `no-restricted-imports` bloqueia `lib/components/ui/**` fora de `lib/components/**`.

Regra completa, árvore de decisão e checagens: skill `design-system`.

---

## 6. Storybook para 100% dos componentes

Componente `acerola-*` (em `lib/components/` ou em `routes/**/components/`) **sem** `.stories.svelte` não entra.

Cada story cobre, no mínimo:

1. **Default** — o uso mais comum.
2. **Todas as variantes** de `ui`.
3. **Estados** — loading, disabled, erro, vazio.
4. **Caso limite** — texto longo que quebra o layout, lista com um item só.

Componente de rota e view-model não têm story — não são componentes de UI reaproveitáveis.

---

## 7. Teste do caminho feliz e do caminho triste, sempre

Todo PR com lógica traz os dois.

Nos arquivos de teste, os dois ficam marcados com `// feliz` e `// triste`.

O que precisa de teste:

* regra de negócio;
* schema;
* view-model;
* mapper;
* service;
* policy de acesso;
* qualquer correção de bug;
* autenticação;
* autorização;
* sessão;
* escalada de privilégio.

O que não precisa:

* arquivo gerado;
* re-export;
* story.

**Toda escalada de privilégio tem teste.** Perfil somente leitura tentando escrever, identidade pela metade, autoria enviada no corpo: cada uma é uma linha trocada que passa na revisão sem ninguém ver.

**Níveis:** unitário (Vitest), componente (Vitest + Testing Library), E2E da API (Vitest + Supertest contra o banco de teste na Neon, `TEST_DATABASE_URL`), E2E da web (Playwright), mutação (Stryker).

---

## 8. Backend (Go)

- Biblioteca padrão primeiro. Dependência nova só quando a padrão não resolve — a VPS tem 1 GB,
  e cada pacote é binário e memória a mais.
- Código em `src-go/<pacote>/`, nome de pacote curto e em inglês. `main.go` na raiz só liga as
  peças (o `go:embed` da SPA é o motivo de ele ficar lá).
- Configuração só por variável de ambiente (`src-go/config`). Valor inválido derruba a partida;
  nunca cai num padrão em silêncio.
- Erro da API tem um formato só, `{"message": "..."}` em pt-BR — o mesmo que a SPA lê em
  `lib/api/api-error.type.ts`. Endereço de API que não existe é 404 em JSON, nunca o `index.html`.
- Teste ao lado (`_test.go`), com `// feliz` e `// triste` acima de cada `Test…`.
- **Lint:** `gofmt`, `go vet` e `golangci-lint` (`.golangci.yml`: `gocyclo` 10, o mesmo limite do ESLint).

---

## 9. Nomes de arquivo

| Tipo           | Padrão                                      | Exemplo                                                |
| -------------- | ------------------------------------------- | ------------------------------------------------------- |
| Componente genérico | `lib/components/acerola-<nome>/acerola-<nome>.svelte` | `lib/components/acerola-text-field/acerola-text-field.svelte` |
| Componente de feature | `routes/(app)/<feature>/components/acerola-<nome>/acerola-<nome>.svelte` | `routes/(app)/artifacts/components/acerola-artifact-card/acerola-artifact-card.svelte` |
| Story          | `acerola-<nome>.stories.svelte`             | `acerola-artifact-card.stories.svelte`                 |
| Hook (view-model) | `lib/hooks/use-<nome>/use-<nome>.svelte.ts` — pasta própria, prefixo `use-`, nunca arquivo solto | `lib/hooks/use-app-shell/use-app-shell.svelte.ts` |
| Teste          | `<arquivo>.test.ts` · Go: `<arquivo>_test.go` | `format-date.test.ts` · `spa_test.go`                 |
| E2E            | `<fluxo>.e2e.ts`                            | `login.e2e.ts`                                         |
| Contrato Zod   | `lib/contracts/<nome>.schema.ts`            | `user.schema.ts`                                       |
| Pacote Go      | `src-go/<pacote>/<arquivo>.go`              | `src-go/server/spa.go`                                 |
| Migration      | `src-go/migrations/<NNNN>_<nome>.sql`       | `0001_artifacts.sql`                                   |
| Util puro      | `<nome>.util.ts`                            | `download-file.util.ts`                                |

Export nomeado sempre. `export default` só onde o framework exige.

Arquivo de barril (`index.ts` que re-exporta) é proibido.

---

## 10. Commits

Formato obrigatório:

```text
[tipo](local): Mensagem
```

Exemplos:

```text
[feat](backend): Upload de artefato salva a versão nova no R2
[fix](web): Card da galeria mostra a data de edição certa
[refactor](backend): Extrai a assinatura do link para um pacote próprio
[test](backend): Link revogado devolve 404
[chore](infra): Caddy passa a fazer proxy só para o binário Go
```

Tipos:

| Tipo       | Quando                             |
| ---------- | ---------------------------------- |
| `feat`     | Funcionalidade nova                |
| `fix`      | Correção de defeito                |
| `refactor` | Muda a forma, não o comportamento  |
| `test`     | Só teste                           |
| `docs`     | Só documentação                    |
| `style`    | Só formatação                      |
| `perf`     | Ganho de desempenho                |
| `build`    | Build, dependências, empacotamento |
| `ci`       | Pipeline                           |
| `chore`    | Manutenção                         |
| `merge`    | Fechamento de branch               |

Local:

| Local     | Alcance                                                               |
| --------- | --------------------------------------------------------------------- |
| `backend` | `src-go/`, `main.go`, `go.mod`, `go.sum`                              |
| `web`     | `svelte/`                                                             |
| `db`      | `src-go/migrations/`                                                  |
| `infra`   | `deploy/` (Caddy, systemd, VPS)                                       |
| `tooling` | `mise.toml`, `.golangci.yml`, `.vscode/`                              |
| `git`     | `.githooks/`, `scripts/git/`, `.gitmessage`, `.gitattributes`, `.gitignore` |
| `ci`      | `.github/`                                                            |
| `docs`    | `docs/` e `.md` na raiz                                               |
| `deps`    | Atualização de dependência                                            |
| `claude`  | `.claude/` e `CLAUDE.md`                                              |

Mensagem em português, primeira letra maiúscula, sem ponto final, cabeçalho até 100 caracteres, descrevendo **o que a mudança faz**.

`[fix](backend): Alterações no handler` é ruim.

`[fix](backend): Corrige rejeição de sessão expirada` é melhor.

Uma mudança que toca dois locais são dois commits. Se não der para separar, use o local mais relevante — nunca `[feat](backend,web)`.

A regra mora em um lugar só, `.githooks/commit-msg`: o git roda esse arquivo em cada commit, e o CI (`Checks/Commits`) roda o mesmo arquivo em cada commit que chega ao GitHub.

`--no-verify` é proibido.

---

## 11. Branches — Gitflow

| Branch             | Quem mexe                          | Sai de    | Volta para         |
| ------------------ | ---------------------------------- | --------- | ------------------ |
| `main`             | Só quem administra (Modo Admin)    | —         | —                  |
| `develop`          | Só por merge `--no-ff`             | `main`    | —                  |
| `feature/<nome>`   | Todo trabalho                      | `develop` | `develop`          |
| `bugfix/<nome>`    | Todo trabalho                      | `develop` | `develop`          |
| `release/<versão>` | Só quem administra                 | `develop` | `main` e `develop` |
| `hotfix/<versão>`  | Só quem administra                 | `main`    | `main` e `develop` |

Nome de branch em inglês, kebab-case: `feature/artifact-upload`, `bugfix/empty-label-error`.

A trava (`scripts/git/branch-guard.sh`, nos hooks `pre-commit`, `pre-merge-commit` e
`pre-push`) recusa commit direto na `main` e na `develop`, e merge ou envio para a `main`. Quem
administra libera a própria máquina com `git config project.admin true` — fica no `.git/config`
daquele clone, nunca no repositório.

### O ciclo de toda mudança

1. Toda mudança nasce numa `feature/` ou `bugfix/`, criada a partir da `develop` atualizada.
2. O trabalho é commitado na branch.
3. No fim, a `develop` é trazida para a branch, e conflito se resolve na branch.
4. `mise run check` verde e sistema rodando (`mise run dev`).
5. Conferido e com OK, `git merge --no-ff` da branch na `develop` (ou PR com título `[merge](local): …`).
6. Envio da `develop` e branch apagada.

O passo a passo é a skill `git-flow`.

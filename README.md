# artifact-docs

Hub pessoal num lugar só: documentação de projetos, artefatos HTML/MD/PDF com galeria e preview,
planejamentos, banco de ideias, hub de apps e casa da marca. O que é e como funciona está no
[briefing](./docs/briefing.md).

## Rodando

Precisa só do [mise](https://mise.jdx.dev):

```bash
mise install       # Node, Go e golangci-lint nas versões do projeto
mise run setup     # hooks do git, Git-Flow e dependências do front
mise run dev       # API Go em :3010 e a tela em http://localhost:5010
```

`mise tasks` lista todos os comandos. `mise run check` é o que precisa passar antes de juntar
qualquer coisa na `develop`.

## Estrutura

| Pasta | O quê |
|---|---|
| `artifact-docs/` | O app: `svelte/` (SPA SvelteKit, `adapter-static`, infra do acerola-ticket), `src-go/` (backend Go) e `main.go`, que sobe o processo e embute a SPA. |
| `docs/` | Briefing e decisões. |
| `.claude/` | Skills e hooks do Claude Code. |

Em produção é **um binário só** (`mise run build` → `artifact-docs/bin/artifact-docs`), com a tela embutida:
sem Node no servidor.

## Contribuindo

As regras de código, commit e branch estão no [CONTRIBUTING](./CONTRIBUTING.md). Toda mudança
nasce numa `feature/` a partir da `develop`.

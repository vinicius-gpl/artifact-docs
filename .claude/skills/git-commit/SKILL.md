---
name: git-commit
description: Salva o trabalho no histórico do git no formato obrigatório [tipo](local) Mensagem — separa por local, escreve a mensagem em português no que a mudança faz, respeita o hook commit-msg. Use quando a pessoa pede "salva", "commita", "guarda o que fizemos", ou ao fim de cada etapa de uma feature.
---

# Commit

## 1. Onde estou?

```bash
git status
git branch --show-current
```

- Em `main` ou `develop`? Não commite aí — o git recusa (a não ser em Modo Admin, `git config
  project.admin true`). Crie a branch antes, pela skill `git-flow` (Fase 1): as mudanças não
  commitadas vão junto para a branch nova.
- Commit na branch de feature é só um **ponto de salvamento**: não precisa do OK da pessoa. O
  OK é para juntar na develop, e isso é a Fase 4 do `git-flow`.
- Nada para commitar? Diga isso e pare.
- O hook não roda? Confira `git config --get core.hooksPath` — tem que ser `.githooks`. Se não
  for, rode `sh scripts/git/setup.sh`.

## 2. Separar por local

Uma mudança que toca dois locais são **dois commits**. Os locais:

| Local | Arquivos |
|---|---|
| `db` | `artifact-docs/src-go/migrations/` |
| `backend` | o resto de `artifact-docs/src-go/`, `artifact-docs/main.go`, `artifact-docs/go.mod`, `artifact-docs/go.sum` |
| `web` | `artifact-docs/svelte/` |
| `infra` | `deploy/` (Caddy, systemd, VPS) |
| `tooling` | `mise.toml`, `artifact-docs/mise.toml`, `artifact-docs/.golangci.yml`, `.vscode/` |
| `git` | `.githooks/`, `scripts/git/`, `.gitmessage`, `.gitattributes`, `.gitignore` |
| `ci` | `.github/` |
| `docs` | `docs/` e `.md` da raiz |
| `deps` | só `artifact-docs/go.mod`/`artifact-docs/go.sum`/`package.json`/lockfile por atualização de dependência |
| `claude` | `.claude/`, `CLAUDE.md` |

Ordem quando é uma feature: `db` → `backend` → `web`.
Arquivo que não cabe em nenhum: junte com o local mais relevante da mesma mudança.

Adicione **por caminho**, nunca `git add -A` às cegas:

```bash
git add artifact-docs/src-go/migrations
git diff --cached --stat     # confira o que vai
```

**Nunca** adicione: `.env`, `node_modules`, `dist`, `bin/`, `coverage`, `artifact-docs/src-go/web/dist/*` (só o `.gitkeep`). Se aparecerem no
status, o `.gitignore` está errado — avise.

## 3. A mensagem

```
[tipo](local): Mensagem
```

- **tipo:** `feat` · `fix` · `refactor` · `test` · `docs` · `style` · `perf` · `build` · `ci` · `chore` · `merge`
- **Mensagem em português**, primeira letra maiúscula, **sem ponto final**, cabeçalho até 100 caracteres.
- Diga o que a mudança **faz** para quem usa ou mantém, não em que arquivo mexeu.

| ❌ | ✅ |
|---|---|
| `[feat](web): Alterações na tela` | `[feat](web): Galeria filtra artefatos por tipo` |
| `[fix](backend): fix bug` | `[fix](backend): Link revogado não abre mais o artefato` |
| `[feat](backend,web): Apps` | dois commits |

Se precisar de corpo (o porquê de uma decisão), deixe uma linha em branco e escreva em
português.

```bash
git commit -m "[feat](db): Tabela de artefatos com versão atual e token do link"
```

Termine a mensagem com as linhas de atribuição que o sistema indicar, se houver.

## 4. O hook recusou

- **"✋ Commit recusado: …"** (formato): leia o motivo, corrija a mensagem e tente de novo.
- **"✋ A branch main é protegida"** / **"Na branch develop só entram…"** → você está na branch
  errada. Siga o `git-flow` (Fase 1) — não procure contorno.
- **Nunca use `--no-verify`.** O CI (`Checks/Commits`) confere os mesmos commits depois do
  push — pular o hook só adia a recusa.

## 5. Depois

Mostre em uma linha o que foi salvo (`git log --oneline -3`). **Não faça `git push`** sem a
pessoa pedir — push publica. Terminar e juntar na develop é o `git-flow`, Fase 3.

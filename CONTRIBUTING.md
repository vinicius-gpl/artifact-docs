# Contribuindo

Quando a regra for verificável por máquina, ela é hook ou gate de CI.

> **Trabalhando com o Claude?** A skill `git-commit` (atalho `/commitar`) faz o passo a passo. Este arquivo é a referência para quando alguém — pessoa ou Claude — tiver dúvida sobre o porquê.

## Preparando a máquina

Depois de clonar, uma vez:

```sh
sh scripts/git/setup.sh
```

Isso aponta o git para os hooks de `.githooks/` e liga o template `.gitmessage`. Não instala nada.

---

## Commits

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
| `backend` | `server/` (Go)                                                        |
| `web`     | `web/` (Svelte)                                                       |
| `db`      | `server/migrations/` e seeds                                          |
| `infra`   | `deploy/` (Caddy, systemd, VPS)                                       |
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

## Branches

Por enquanto, tudo vai direto na `main`. O Gitflow entra depois.

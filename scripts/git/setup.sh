#!/bin/sh
# Prepara o git desta máquina para o padrão do projeto. Rode uma vez depois de clonar:
#
#   sh scripts/git/setup.sh
#
# 1. O git passa a rodar os hooks de `.githooks/`: formato do commit (`commit-msg`), trava de
#    branch (`pre-commit`, `pre-merge-commit`, `pre-push`) e checagem de design (`pre-push`).
# 2. O template de mensagem (`.gitmessage`) aparece no editor.
# 3. O Git-Flow do Tower lê as chaves `gitflow.*`: gravá-las faz o Tower já abrir com o
#    Git-Flow inicializado, com os mesmos nomes de branch que o Claude e a trava usam. Só grava
#    o que ainda não existe — quem já configurou diferente não é sobrescrito.
#
# Não depende de npm nem de husky: é só configuração local do git.
set -e

cd "$(git rev-parse --show-toplevel)"

git config --local core.hooksPath .githooks
git config --local commit.template .gitmessage
chmod +x .githooks/* scripts/git/*.sh

set_default() {
  git config --local --get "$1" >/dev/null || git config --local "$1" "$2"
}

set_default gitflow.branch.master main
set_default gitflow.branch.develop develop
for prefix in feature bugfix release hotfix support; do
  set_default "gitflow.prefix.$prefix" "$prefix/"
done
set_default gitflow.prefix.versiontag v

echo "Git configurado: hooks em .githooks/, template .gitmessage e Git-Flow (main/develop)."

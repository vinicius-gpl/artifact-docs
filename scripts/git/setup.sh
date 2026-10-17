#!/bin/sh
# Prepara o git desta máquina para o padrão do projeto. Rode uma vez depois de clonar:
#
#   sh scripts/git/setup.sh
#
# 1. O git passa a rodar os hooks de `.githooks/` (o `commit-msg` confere o formato).
# 2. O template de mensagem (`.gitmessage`) aparece no editor.
#
# Não depende de npm nem de husky: é só configuração local do git.
set -e

cd "$(git rev-parse --show-toplevel)"

git config --local core.hooksPath .githooks
git config --local commit.template .gitmessage
chmod +x .githooks/*

echo "Git configurado: hooks em .githooks/ e template .gitmessage."

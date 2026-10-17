#!/usr/bin/env sh
#
# A TRAVA DE BRANCH — CONTRIBUTING, "Branches".
#
# Roda nos hooks do git (pre-commit, pre-merge-commit, pre-push), então vale para o terminal,
# para o Claude e para o Tower. Ela recusa:
#
#   main     → commit, merge e envio. A main só recebe release e hotfix, e isso é do admin.
#   develop  → commit direto. Ela só recebe feature pronta, por merge.
#
# Quem administra libera a própria máquina UMA vez, e só ela:
#
#   git config project.admin true
#
# A configuração fica no `.git/config` daquele clone e não é versionada.

action="$1"

if [ "$(git config --get project.admin 2>/dev/null)" = "true" ]; then
  exit 0
fi

branch=$(git symbolic-ref --quiet --short HEAD 2>/dev/null)
git_dir=$(git rev-parse --git-dir 2>/dev/null)

refuse() {
  printf '\n' >&2
  printf '  ✋ %s\n' "$1" >&2
  printf '\n' >&2
  printf '  %s\n' "$2" >&2
  printf '\n' >&2
  printf '  Nada foi perdido: suas alterações continuam aqui.\n' >&2
  printf '\n' >&2
  exit 1
}

case "$action" in
  commit)
    if [ "$branch" = "main" ]; then
      refuse \
        'A branch "main" é protegida. Ninguém salva alterações direto nela.' \
        'Crie uma branch de feature a partir da develop (skill git-flow, Fase 1). As alterações vão junto.'
    fi

    # Concluir um merge em andamento é permitido: é o merge da feature chegando.
    if [ "$branch" = "develop" ] && [ ! -f "$git_dir/MERGE_HEAD" ]; then
      refuse \
        'Na branch "develop" só entram funcionalidades prontas, por merge.' \
        'Crie uma branch de feature a partir da develop (skill git-flow, Fase 1). As alterações vão junto.'
    fi
    ;;

  merge)
    if [ "$branch" = "main" ]; then
      refuse \
        'Juntar alterações na "main" é tarefa de quem administra o projeto.' \
        'As funcionalidades terminam na "develop". Release e hotfix são com o admin (git config project.admin true).'
    fi
    ;;

  push)
    # O git entrega uma linha por referência enviada: <local> <sha> <remota> <sha>.
    while read -r _local_ref _local_sha remote_ref _remote_sha; do
      if [ "$remote_ref" = "refs/heads/main" ]; then
        refuse \
          'Enviar para a "main" é tarefa de quem administra o projeto.' \
          'Envie a "develop" ou a sua branch de feature.'
      fi
    done
    ;;
esac

exit 0

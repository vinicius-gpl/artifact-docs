#!/bin/sh
# Etapas do deploy chamadas pela Stack do Komodo (.komodo/komodo.toml). Igual em todos os apps.
#
#   sh komodo.sh pre-deploy  <app> <servidor>            confere que o Build deixou a imagem aqui
#   sh komodo.sh deploy      <app> <servidor> <comando>  envolve o comando `docker stack` do Komodo
#   sh komodo.sh post-deploy <app> <servidor>            publica a rota no Caddy do servidor
#
# A imagem vem do Build do Komodo neste mesmo servidor, sem registry:
# <app>:<commit curto>-<servidor>. A stack é <servidor>_<app> e o Swarm chama os serviços de
# <servidor>_<app>_<serviço>, que é para onde o Caddy faz proxy pela rede edge.
set -eu

step="${1:?Etapa: pre-deploy, deploy ou post-deploy}"
APP_NAME="${2:?Nome do app}"
DEPLOY_SERVER="${3:?Nome do servidor}"
shift 3

APP_REVISION="$(git rev-parse --short HEAD)"
APP_IMAGE="$APP_NAME:$APP_REVISION-$DEPLOY_SERVER"
STACK_NAME="${DEPLOY_SERVER}_${APP_NAME}"
export APP_NAME DEPLOY_SERVER APP_REVISION APP_IMAGE STACK_NAME

case "$step" in
  pre-deploy)
    if ! docker image inspect "$APP_IMAGE" >/dev/null 2>&1; then
      echo "Imagem $APP_IMAGE não existe neste servidor: rode o Build antes." >&2
      exit 1
    fi
    ;;

  deploy)
    # O Komodo guarda a saída do `docker stack config`: as variáveis ficam fora dela.
    case "$*" in
      "docker stack config "*) exec "$@" --skip-interpolation ;;
    esac
    exec "$@"
    ;;

  post-deploy)
    sh route.sh
    ;;

  *)
    echo "Etapa inválida: $step" >&2
    exit 1
    ;;
esac

#!/bin/sh
# Deploy steps called by the Komodo Stack (.komodo/komodo.toml). Same file in every app.
#
#   sh komodo.sh pre-deploy  <app> <server>            checks that the Build left the image here
#   sh komodo.sh deploy      <app> <server> <command>  wraps Komodo's `docker stack` command
#   sh komodo.sh post-deploy <app> <server>            publishes the route on the server's Caddy
#
# The image comes from the Komodo Build on this same server, with no registry:
# <app>:<short commit>-<server>. The stack is <server>_<app> and Swarm names the services
# <server>_<app>_<service>, which is what Caddy proxies to over the edge network.
set -eu

step="${1:?Step: pre-deploy, deploy or post-deploy}"
APP_NAME="${2:?App name}"
DEPLOY_SERVER="${3:?Server name}"
shift 3

APP_REVISION="$(git rev-parse --short HEAD)"
APP_IMAGE="$APP_NAME:$APP_REVISION-$DEPLOY_SERVER"
STACK_NAME="${DEPLOY_SERVER}_${APP_NAME}"
export APP_NAME DEPLOY_SERVER APP_REVISION APP_IMAGE STACK_NAME

case "$step" in
  pre-deploy)
    if ! docker image inspect "$APP_IMAGE" >/dev/null 2>&1; then
      echo "Image $APP_IMAGE not found on this server: run the Build first." >&2
      exit 1
    fi
    ;;

  deploy)
    # Komodo stores the output of `docker stack config`: keep variables out of it.
    case "$*" in
      "docker stack config "*) exec "$@" --skip-interpolation ;;
    esac
    exec "$@"
    ;;

  post-deploy)
    sh route.sh
    ;;

  *)
    echo "Unknown step: $step" >&2
    exit 1
    ;;
esac

#!/bin/sh
# Publishes the app on the server's Caddy (owned by config): writes conf.d/<stack>.caddy from
# <app>.caddy, validates and reloads. Called by komodo.sh post-deploy, with APP_NAME and
# STACK_NAME exported. Same file in every repository.
set -eu

: "${APP_NAME:?}"
: "${STACK_NAME:?}"

conf=/etc/caddy/conf.d
dir="$(cd "$(dirname "$0")" && pwd)"

docker exec -i caddy sh -c "cat > $conf/$STACK_NAME.caddy" <"$dir/$APP_NAME.caddy"

if ! docker exec caddy caddy validate --config /etc/caddy/Caddyfile >/dev/null 2>&1; then
  docker exec caddy rm -f "$conf/$STACK_NAME.caddy"
  echo "Invalid $APP_NAME.caddy: route removed, Caddy kept the previous config." >&2
  exit 1
fi

docker exec caddy caddy reload --config /etc/caddy/Caddyfile

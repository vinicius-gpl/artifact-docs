#!/bin/sh
# Publica o app no Caddy do servidor (que é do config): grava conf.d/<stack>.caddy a partir do
# site.caddy, valida e recarrega. Chamado pelo komodo.sh post-deploy, com STACK_NAME exportado.
set -eu

: "${STACK_NAME:?}"

# Rotas do deploy antigo pelo Ansible, trocadas pelo <stack>.caddy (mesmo domínio: o Caddy recusa os dois).
LEGACY_ROUTES="artifact-docs.caddy"

conf=/etc/caddy/conf.d
dir="$(cd "$(dirname "$0")" && pwd)"

for legacy in $LEGACY_ROUTES; do
  docker exec caddy rm -f "$conf/$legacy"
done

docker exec -i caddy sh -c "cat > $conf/$STACK_NAME.caddy" <"$dir/site.caddy"

if ! docker exec caddy caddy validate --config /etc/caddy/Caddyfile >/dev/null 2>&1; then
  docker exec caddy rm -f "$conf/$STACK_NAME.caddy"
  echo "site.caddy inválido: a rota saiu e o Caddy ficou com a config anterior." >&2
  exit 1
fi

docker exec caddy caddy reload --config /etc/caddy/Caddyfile

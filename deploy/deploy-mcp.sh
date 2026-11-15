#!/bin/sh
# Aprovisiona la demo y actualiza el MCP con estado de rollback propio.
set -eu
if [ $# -ne 1 ] || [ -z "$1" ]; then
  printf 'Uso: deploy-mcp.sh <image-tag>\n' >&2
  exit 2
fi
root=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
cd "$root"
umask 077
# Los secretos se generan en el VPS, se conservan y nunca viajan en argumentos SSH.
if [ ! -f .env.mcp ]; then
  admin_password=$(openssl rand -hex 24)
  reader_password=$(openssl rand -hex 24)
  {
    printf 'MCP_ADMIN_PASSWORD=%s\n' "$admin_password"
    printf 'MCP_READER_PASSWORD=%s\n' "$reader_password"
    printf 'MCP_PUBLIC_URL=https://seguimiento0-medicion.aedestec.com\n'
  } > .env.mcp.tmp
  mv .env.mcp.tmp .env.mcp
fi
export IMAGE_TAG=$1
demo_compose() {
  docker compose --env-file .env.mcp -f docker-compose.mcp.yml "$@"
}
cleanup() {
  demo_compose stop bootstrap >/dev/null
}
trap cleanup EXIT
demo_compose up -d --pull always --wait bootstrap
demo_compose exec -T postgres psql -U postgres -d seguimiento_medicion \
  -v ON_ERROR_STOP=1 < deploy/mcp-reader.sql
project_count=$(demo_compose exec -T postgres psql -U postgres \
  -d seguimiento_medicion -At -v ON_ERROR_STOP=1 -c 'SELECT count(*) FROM projects')
if [ "$project_count" = 0 ]; then
  demo_compose exec -T postgres psql -U postgres -d seguimiento_medicion \
    -v ON_ERROR_STOP=1 < backend/cmd/mcp/testdata/demo.sql
fi
cleanup
trap - EXIT
export COMPOSE_FILE="$root/docker-compose.mcp.yml"
export COMPOSE_ENV_FILE="$root/.env.mcp"
export STATE_FILE="$root/.deploy-mcp-sha"
export HEALTH_URL=http://127.0.0.1:3004/healthz
sh "$root/deploy/deploy.sh" "$@"

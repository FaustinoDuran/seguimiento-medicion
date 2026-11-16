#!/bin/sh
# Prepara datos compartidos de UI/MCP y actualiza el servidor de lectura.
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
  reader_password=$(openssl rand -hex 24)
  {
    printf 'MCP_READER_PASSWORD=%s\n' "$reader_password"
    printf 'MCP_PUBLIC_URL=https://seguimiento-medicion.aedestec.com\n'
  } > .env.mcp.tmp
  mv .env.mcp.tmp .env.mcp
fi
# Corrige el dominio original sin regenerar credenciales ni alterar otros orígenes.
if grep -qx 'MCP_PUBLIC_URL=https://seguimiento0-medicion.aedestec.com' .env.mcp; then
  sed 's|^MCP_PUBLIC_URL=https://seguimiento0-medicion.aedestec.com$|MCP_PUBLIC_URL=https://seguimiento-medicion.aedestec.com|' \
    .env.mcp > .env.mcp.tmp
  mv .env.mcp.tmp .env.mcp
fi
export IMAGE_TAG=$1
database_name=$(printf 'SELECT current_database();\n' | sh deploy/app-psql.sh "$1" -At)
case "$database_name" in
  ''|*[!a-zA-Z0-9_]*) echo 'Nombre de base no válido para la URL MCP' >&2; exit 1 ;;
esac
export MCP_DATABASE_NAME="$database_name"
reader_password=$(sed -n 's/^MCP_READER_PASSWORD=//p' .env.mcp)
case "$reader_password" in
  ''|*[!a-fA-F0-9]*) echo 'MCP_READER_PASSWORD debe ser hexadecimal sin comillas' >&2; exit 1 ;;
esac
sh scripts/seed-project.sh "$1"
# Contraseña por stdin: no aparece en argumentos de psql ni de Docker.
{
  printf '\\set reader_password %s\n' "$reader_password"
  cat deploy/mcp-reader.sql
} | sh deploy/app-psql.sh "$1"
export COMPOSE_FILE="$root/docker-compose.mcp.yml"
export COMPOSE_ENV_FILE="$root/.env.mcp"
export STATE_FILE="$root/.deploy-mcp-sha"
export HEALTH_URL=http://127.0.0.1:3008/healthz
sh "$root/deploy/deploy.sh" "$@"

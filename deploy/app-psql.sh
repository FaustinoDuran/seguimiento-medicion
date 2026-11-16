#!/bin/sh
# Ejecuta SQL de stdin usando la identidad PostgreSQL del servicio de la app.
set -eu
tag=${1:?Uso: app-psql.sh <image-tag> [psql-options]}
shift
root=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
cd "$root"
IMAGE_TAG=$tag docker compose --env-file .env -f docker-compose.prod.yml \
  exec -T postgres sh -c \
  'exec psql -X -U "$POSTGRES_USER" -d "$POSTGRES_DB" -v ON_ERROR_STOP=1 "$@"' sh "$@"

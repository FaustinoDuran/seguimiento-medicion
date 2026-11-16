#!/bin/sh
# Carga aditiva. Local: DATABASE_URL=... sh scripts/seed-project.sh
# VPS: sh scripts/seed-project.sh <sha-publicado>
set -eu
root=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
if [ "$#" -eq 1 ]; then
  sh "$root/deploy/app-psql.sh" "$1" < "$root/scripts/seed-project.sql"
elif [ "$#" -eq 0 ] && [ -n "${DATABASE_URL:-}" ]; then
  PGDATABASE=$DATABASE_URL psql -X -v ON_ERROR_STOP=1 < "$root/scripts/seed-project.sql"
else
  echo 'Uso: seed-project.sh <sha-publicado>, o DATABASE_URL=... seed-project.sh' >&2
  exit 2
fi

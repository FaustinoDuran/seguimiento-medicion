#!/bin/sh
# Reemplaza los contenedores de producción por el SHA indicado.
# Si /health no responde, vuelve al SHA guardado en .deploy-sha.
#
# Uso, desde el checkout del VPS (/opt/seguimiento-medicion):
#   API_IMAGE=... WEB_IMAGE=... ./deploy/deploy.sh <sha>

set -eu

if [ $# -ne 1 ] || [ -z "$1" ]; then
  printf 'Uso: deploy.sh <image-tag>\n' >&2
  exit 2
fi

tag=$1
root=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
cd "$root"

compose_file=${COMPOSE_FILE:-docker-compose.prod.yml}
state_file=${STATE_FILE:-$root/.deploy-sha}
health_url=${HEALTH_URL:-http://127.0.0.1:3002/health}
env_file=${COMPOSE_ENV_FILE:-.env}

previous=
if [ -f "$state_file" ]; then
  previous=$(cat "$state_file")
fi

deploy_tag() {
  IMAGE_TAG=$1 docker compose --env-file "$env_file" -f "$compose_file" pull || return 1
  IMAGE_TAG=$1 docker compose --env-file "$env_file" -f "$compose_file" up -d --no-build --wait || return 1
}

failed=0
deploy_tag "$tag" || failed=1

healthy=0
if [ "$failed" -eq 0 ]; then
  if curl -fsS --retry 10 --retry-connrefused --retry-delay 2 "$health_url" >/dev/null; then
    healthy=1
  fi
fi

if [ "$healthy" -eq 1 ]; then
  printf '%s\n' "$tag" > "$state_file"
  exit 0
fi

printf 'El chequeo de /health falló para %s\n' "$tag" >&2
if [ -n "$previous" ] && [ "$previous" != "$tag" ]; then
  printf 'Vuelta al SHA %s\n' "$previous" >&2
  deploy_tag "$previous" || exit 1
  printf '%s\n' "$previous" > "$state_file"
fi
exit 1

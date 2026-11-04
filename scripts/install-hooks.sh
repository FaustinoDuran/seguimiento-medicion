#!/bin/sh
# Instala el hook pre-push versionado de TASK-014 (#61) en este clon.
#
# Uso, desde cualquier directorio del repo:
#   ./scripts/install-hooks.sh
#
# Es idempotente y conserva un hook pre-push previo como pre-push.local.

set -eu

fail() {
  printf '[install-hooks] %s\n' "$1" >&2
  exit 1
}

repository_root=$(git rev-parse --show-toplevel 2>/dev/null) \
  || fail "Ejecutá este script dentro del repositorio."
cd "$repository_root"

source_hook="$repository_root/.githooks/pre-push"
[ -f "$source_hook" ] || fail "No existe $source_hook."
chmod +x "$source_hook"

hooks_dir=$(git rev-parse --git-path hooks)
mkdir -p "$hooks_dir"

target="$hooks_dir/pre-push"
backup="$hooks_dir/pre-push.local"

if [ -L "$target" ] && [ "$(readlink "$target")" = "$source_hook" ]; then
  printf '[install-hooks] El hook pre-push ya estaba instalado en %s.\n' "$target"
  exit 0
fi

if [ -f "$target" ] && cmp -s "$target" "$source_hook"; then
  printf '[install-hooks] El hook pre-push ya estaba instalado en %s.\n' "$target"
  exit 0
fi

if [ -e "$target" ] || [ -L "$target" ]; then
  if [ -e "$backup" ] || [ -L "$backup" ]; then
    fail "Existe un pre-push distinto y también $backup. Resolvé el conflicto a mano para no perder ningún hook."
  fi
  mv "$target" "$backup"
  chmod +x "$backup" 2>/dev/null || true
  printf '[install-hooks] Se conservó el hook previo en %s; se ejecutará después de la suite.\n' "$backup"
fi

if ln -s "$source_hook" "$target" 2>/dev/null; then
  :
else
  # Sin soporte de symlinks: se copia y hay que reinstalar si el hook cambia.
  cp "$source_hook" "$target"
  printf '[install-hooks] Symlinks no disponibles: se copió el hook. Reinstalá después de cada cambio en .githooks/.\n'
fi

chmod +x "$target"

printf '[install-hooks] Hook instalado en %s.\n' "$target"
printf '[install-hooks] Cada push ejecutará desde backend/: go test -count=1 -race ./...\n'
printf '[install-hooks] Requiere DATABASE_URL, BACKLOG_TEST_DATABASE_URL y SPRINTS_TEST_DATABASE_URL. Ver README.\n'

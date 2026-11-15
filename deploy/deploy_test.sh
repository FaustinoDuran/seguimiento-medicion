#!/bin/sh
# Prueba deploy.sh con docker y curl falsos. No toca el daemon ni la red.

set -eu

root=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

mkdir -p "$tmp/bin"
cat > "$tmp/bin/docker" << 'EOF'
#!/bin/sh
printf '%s %s\n' "${IMAGE_TAG:-}" "$*" >> "${DEPLOY_LOG:?}"
if [ "${IMAGE_TAG:-}" = "fail" ]; then
  exit 1
fi
exit 0
EOF
cat > "$tmp/bin/curl" << 'EOF'
#!/bin/sh
exit "${CURL_EXIT:-0}"
EOF
chmod +x "$tmp/bin/docker" "$tmp/bin/curl"

export PATH="$tmp/bin:$PATH"
export COMPOSE_FILE="docker-compose.prod.yml"
export HEALTH_URL="http://127.0.0.1:9/health"
export DEPLOY_LOG="$tmp/log"
export STATE_FILE="$tmp/state"

fail() {
  printf 'deploy_test: %s\n' "$1" >&2
  exit 1
}

assert_eq() {
  if [ "$1" != "$2" ]; then
    printf 'deploy_test: esperaba [%s], obtuvo [%s]\n' "$2" "$1" >&2
    exit 1
  fi
}

pull_line() {
  printf '%s compose -f %s pull' "$1" "$COMPOSE_FILE"
}

up_line() {
  printf '%s compose -f %s up -d --no-build --wait' "$1" "$COMPOSE_FILE"
}

: > "$DEPLOY_LOG"
rm -f "$STATE_FILE"
export CURL_EXIT=0
"$root/deploy/deploy.sh" abc123 || fail "un deploy sano tiene que salir 0"
assert_eq "$(cat "$STATE_FILE")" "abc123"
grep -qx "$(pull_line abc123)" "$DEPLOY_LOG" || fail "falta el pull del tag nuevo"
grep -qx "$(up_line abc123)" "$DEPLOY_LOG" || fail "falta el up del tag nuevo"

: > "$DEPLOY_LOG"
printf 'oldsha\n' > "$STATE_FILE"
export CURL_EXIT=1
set +e
"$root/deploy/deploy.sh" newsha
status=$?
set -e
[ "$status" -eq 1 ] || fail "un /health fallido tiene que salir 1, salió $status"
assert_eq "$(cat "$STATE_FILE")" "oldsha"
grep -qx "$(pull_line newsha)" "$DEPLOY_LOG" || fail "falta el pull del tag fallido"
grep -qx "$(up_line oldsha)" "$DEPLOY_LOG" || fail "falta el up de rollback"

: > "$DEPLOY_LOG"
rm -f "$STATE_FILE"
set +e
"$root/deploy/deploy.sh" newsha
status=$?
set -e
[ "$status" -eq 1 ] || fail "el primer deploy fallido tiene que salir 1"
[ ! -f "$STATE_FILE" ] || fail "sin SHA previo no se escribe estado"
if grep -q '^oldsha ' "$DEPLOY_LOG"; then
  fail "hizo rollback sin SHA previo"
fi

: > "$DEPLOY_LOG"
printf 'oldsha\n' > "$STATE_FILE"
export CURL_EXIT=0
set +e
"$root/deploy/deploy.sh" fail
status=$?
set -e
[ "$status" -eq 1 ] || fail "un compose fallido tiene que salir 1, salió $status"
assert_eq "$(cat "$STATE_FILE")" "oldsha"
grep -qx "$(pull_line fail)" "$DEPLOY_LOG" || fail "falta el pull que falla"
grep -qx "$(pull_line oldsha)" "$DEPLOY_LOG" || fail "falta el pull de rollback"

: > "$DEPLOY_LOG"
printf 'abc123\n' > "$STATE_FILE"
export CURL_EXIT=1
set +e
"$root/deploy/deploy.sh" abc123
status=$?
set -e
[ "$status" -eq 1 ] || fail "repetir el mismo SHA fallido tiene que salir 1"
assert_eq "$(cat "$STATE_FILE")" "abc123"
pulls=$(grep -c ' pull$' "$DEPLOY_LOG" || true)
[ "$pulls" -eq 1 ] || fail "no tiene que volver a desplegar el mismo SHA, pulls=$pulls"

printf 'deploy_test: ok\n'

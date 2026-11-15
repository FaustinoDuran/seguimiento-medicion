#!/bin/sh
# Verifica primer despliegue, repetición y fallo de aprovisionamiento sin Docker.
set -eu
root=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
mkdir -p "$tmp/bin" "$tmp/repo/deploy" "$tmp/repo/backend/cmd/mcp/testdata"
cp "$root/deploy/deploy-mcp.sh" "$root/deploy/deploy.sh" \
  "$root/deploy/mcp-reader.sql" "$tmp/repo/deploy/"
cp "$root/backend/cmd/mcp/testdata/demo.sql" "$tmp/repo/backend/cmd/mcp/testdata/"
cat > "$tmp/bin/docker" <<'EOF'
#!/bin/sh
printf '%s\n' "$*" >> "$MOCK_LOG"
case "$*" in
  *'up -d --pull always --wait bootstrap'*) exit "${MOCK_BOOTSTRAP_EXIT:-0}" ;;
  *'SELECT count(*) FROM projects'*) printf '%s\n' "${MOCK_COUNT:-0}" ;;
  *'exec -T postgres psql'*)
    input=$(cat)
    case "$input" in *'INSERT INTO projects'*) echo seeded >> "$MOCK_LOG" ;; esac ;;
esac
EOF
cat > "$tmp/bin/curl" <<'EOF'
#!/bin/sh
exit 0
EOF
chmod +x "$tmp/bin/docker" "$tmp/bin/curl"
export PATH="$tmp/bin:$PATH" MOCK_LOG="$tmp/log"
sh "$tmp/repo/deploy/deploy-mcp.sh" first
test -s "$tmp/repo/.env.mcp"
grep -qx 'MCP_PUBLIC_URL=https://seguimiento0-medicion.aedestec.com' "$tmp/repo/.env.mcp"
test "$(cat "$tmp/repo/.deploy-mcp-sha")" = first
test "$(grep -c '^seeded$' "$MOCK_LOG")" = 1
before=$(cksum < "$tmp/repo/.env.mcp")
export MOCK_COUNT=2
sh "$tmp/repo/deploy/deploy-mcp.sh" second
test "$before" = "$(cksum < "$tmp/repo/.env.mcp")"
test "$(grep -c '^seeded$' "$MOCK_LOG")" = 1
test "$(cat "$tmp/repo/.deploy-mcp-sha")" = second
export MOCK_BOOTSTRAP_EXIT=1
if sh "$tmp/repo/deploy/deploy-mcp.sh" broken; then
  echo 'El bootstrap fallido debe rechazar el despliegue' >&2
  exit 1
fi
test "$(cat "$tmp/repo/.deploy-mcp-sha")" = second
tail -1 "$MOCK_LOG" | grep -q 'stop bootstrap$'
printf 'deploy-mcp_test: ok\n'

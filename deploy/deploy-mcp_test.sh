#!/bin/sh
# Verifica primer despliegue, repetición y fallo de aprovisionamiento sin Docker.
set -eu
root=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
mkdir -p "$tmp/bin" "$tmp/repo/deploy" "$tmp/repo/scripts"
cp "$root/deploy/deploy-mcp.sh" "$root/deploy/deploy.sh" \
  "$root/deploy/mcp-reader.sql" "$root/deploy/app-psql.sh" "$tmp/repo/deploy/"
cp "$root/scripts/seed-project.sh" "$root/scripts/seed-project.sql" "$tmp/repo/scripts/"
cat > "$tmp/bin/docker" <<'EOF'
#!/bin/sh
printf '%s\n' "$*" >> "$MOCK_LOG"
case "$*" in
  *'exec -T postgres'*)
    input=$(cat)
    case "$input" in
      *'SELECT current_database();'*) echo seguimiento_medicion ;;
      *'INSERT INTO projects'*) echo seeded >> "$MOCK_LOG"; exit "${MOCK_SEED_EXIT:-0}" ;;
      *'ALTER ROLE mcp_app_reader'*)
        case "$input" in *'\set reader_password '*) echo role-ready >> "$MOCK_LOG" ;; *) exit 1 ;; esac ;;
    esac ;;
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
grep -qx 'MCP_PUBLIC_URL=https://seguimiento-medicion.aedestec.com' "$tmp/repo/.env.mcp"
test "$(cat "$tmp/repo/.deploy-mcp-sha")" = first
test "$(grep -c '^seeded$' "$MOCK_LOG")" = 1
before=$(cksum < "$tmp/repo/.env.mcp")
sh "$tmp/repo/deploy/deploy-mcp.sh" second
test "$before" = "$(cksum < "$tmp/repo/.env.mcp")"
test "$(grep -c '^seeded$' "$MOCK_LOG")" = 2
test "$(grep -c '^role-ready$' "$MOCK_LOG")" = 2
test "$(cat "$tmp/repo/.deploy-mcp-sha")" = second
credentials_before=$(sed '/^MCP_PUBLIC_URL=/d' "$tmp/repo/.env.mcp" | cksum)
sed 's|seguimiento-medicion.aedestec.com|seguimiento0-medicion.aedestec.com|' \
  "$tmp/repo/.env.mcp" > "$tmp/old-env"
cp "$tmp/old-env" "$tmp/repo/.env.mcp"
sh "$tmp/repo/deploy/deploy-mcp.sh" corrected
grep -qx 'MCP_PUBLIC_URL=https://seguimiento-medicion.aedestec.com' "$tmp/repo/.env.mcp"
test "$credentials_before" = "$(sed '/^MCP_PUBLIC_URL=/d' "$tmp/repo/.env.mcp" | cksum)"
export MOCK_SEED_EXIT=1
if sh "$tmp/repo/deploy/deploy-mcp.sh" broken; then
  echo 'El seed fallido debe rechazar el despliegue' >&2
  exit 1
fi
test "$(cat "$tmp/repo/.deploy-mcp-sha")" = corrected
test "$(grep -c '^role-ready$' "$MOCK_LOG")" = 3
if grep -q 'up .*bootstrap' "$MOCK_LOG"; then
  echo 'No debe iniciar una base o API de demo independiente' >&2; exit 1
fi
printf 'deploy-mcp_test: ok\n'

#!/usr/bin/env bash
# Build demo container, run kong, generate README screenshots.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
OUT="$ROOT/docs/screenshots"
CACHE="$OUT/.cache"

mkdir -p "$CACHE" "$OUT"

echo "→ building demo image..."
docker build -f "$ROOT/docker/demo/Dockerfile" -t kong-demo "$ROOT" -q

# Pin time inside container for stable NEXT columns
FIXED_DATE="${KONG_DEMO_DATE:-2026-09-01T12:00:00}"

capture() {
  local name="$1"
  shift
  docker run --rm -e TZ=America/Sao_Paulo "$@" kong-demo "$@" > "$CACHE/${name}.txt" 2>&1 || true
}

# Override CMD per run
docker run --rm \
  -e TZ=America/Sao_Paulo \
  --entrypoint /bin/sh \
  kong-demo -c "date -s '$FIXED_DATE' >/dev/null 2>&1; kong list" \
  > "$CACHE/kong-list.txt"

docker run --rm \
  -e TZ=America/Sao_Paulo \
  --entrypoint /bin/sh \
  kong-demo -c "date -s '$FIXED_DATE' >/dev/null 2>&1; kong list --json" \
  > "$CACHE/kong-list-json.txt"

docker run --rm \
  -e TZ=America/Sao_Paulo \
  --entrypoint /bin/sh \
  kong-demo -c "date -s '$FIXED_DATE' >/dev/null 2>&1; kong version" \
  > "$CACHE/kong-version.txt"

cat > "$CACHE/kong-install.txt" <<EOF
→ downloading kong demo (linux/amd64)
✓ kong demo installed → /usr/local/bin/kong

kong demo (linux/amd64)
EOF

echo "→ rendering PNGs..."
python3 "$ROOT/scripts/terminal_png.py" \
  --title "kong list" \
  --prompt "root@prod-worker-03:~#" \
  --input "$CACHE/kong-list.txt" \
  --output "$OUT/kong-list.png"

python3 "$ROOT/scripts/terminal_png.py" \
  --title "kong list --json" \
  --prompt "root@prod-worker-03:~#" \
  --input "$CACHE/kong-list-json.txt" \
  --output "$OUT/kong-list-json.png" \
  --width 1100 \
  --font-size 13

python3 "$ROOT/scripts/terminal_png.py" \
  --title "install" \
  --prompt "$" \
  --input "$CACHE/kong-install.txt" \
  --command "curl -fsSL https://raw.githubusercontent.com/carvalhosauro/kong/main/scripts/install.sh | bash" \
  --output "$OUT/kong-install.png" \
  --height 140

echo "✓ screenshots in docs/screenshots/"
ls -la "$OUT"/*.png

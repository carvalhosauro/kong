#!/usr/bin/env bash
# kong installer — one-liner: curl -fsSL https://raw.githubusercontent.com/carvalhosauro/kong/main/scripts/install.sh | bash
set -euo pipefail

KONG_VERSION="${KONG_VERSION:-latest}"
KONG_GITHUB_REPO="${KONG_GITHUB_REPO:-carvalhosauro/kong}"
KONG_INSTALL_DIR="${KONG_INSTALL_DIR:-}"
USE_SUDO="${USE_SUDO:-auto}"

usage() {
  cat <<'EOF'
kong installer

Usage:
  curl -fsSL https://raw.githubusercontent.com/carvalhosauro/kong/main/scripts/install.sh | bash

Environment:
  KONG_VERSION=v0.1.0     Pin version (default: latest release)
  KONG_INSTALL_DIR=PATH   Install directory (default: /usr/local/bin or ~/.local/bin)
  KONG_GITHUB_REPO=org/repo
  USE_SUDO=0|1|auto       Use sudo for system paths (default: auto)

Examples:
  curl -fsSL ... | bash
  KONG_INSTALL_DIR=$HOME/.local/bin curl -fsSL ... | bash
  KONG_VERSION=v0.1.0 curl -fsSL ... | bash
EOF
}

for arg in "$@"; do
  case "$arg" in
    -h|--help) usage; exit 0 ;;
  esac
done

err() { echo "kong install: $*" >&2; }

need_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    err "missing required command: $1"
    exit 1
  fi
}

need_cmd curl
need_cmd tar
need_cmd uname

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "$OS" in
  linux) ;;
  *)
    err "unsupported OS: $OS (kong targets Linux servers with cron)"
    exit 1
    ;;
esac

ARCH="$(uname -m)"
case "$ARCH" in
  x86_64|amd64) ARCH=amd64 ;;
  aarch64|arm64) ARCH=arm64 ;;
  *)
    err "unsupported architecture: $ARCH"
    exit 1
    ;;
esac

if [[ "$KONG_VERSION" == "latest" ]]; then
  need_cmd grep
  KONG_VERSION="$(
    curl -fsSL "https://api.github.com/repos/${KONG_GITHUB_REPO}/releases/latest" \
      | grep -o '"tag_name": *"[^"]*"' \
      | head -1 \
      | sed -E 's/.*"v?([^"]+)".*/\1/'
  )"
  if [[ -z "$KONG_VERSION" ]]; then
    err "could not resolve latest release from GitHub"
    err "set KONG_VERSION manually or build from source: go install github.com/carvalhosauro/kong/cmd/kong@latest"
    exit 1
  fi
fi

KONG_VERSION="${KONG_VERSION#v}"

if [[ -z "$KONG_INSTALL_DIR" ]]; then
  if [[ -w /usr/local/bin ]] 2>/dev/null; then
    KONG_INSTALL_DIR="/usr/local/bin"
  elif [[ -d "$HOME/.local/bin" ]] || mkdir -p "$HOME/.local/bin" 2>/dev/null; then
    KONG_INSTALL_DIR="$HOME/.local/bin"
  else
    KONG_INSTALL_DIR="/usr/local/bin"
  fi
fi

mkdir -p "$KONG_INSTALL_DIR"

install_bin() {
  local src="$1" dest="$2"
  if [[ -w "$(dirname "$dest")" ]]; then
    install -m 0755 "$src" "$dest"
  elif [[ "$USE_SUDO" != "0" ]] && command -v sudo >/dev/null 2>&1; then
    sudo install -m 0755 "$src" "$dest"
  else
    err "cannot write to $dest (try KONG_INSTALL_DIR=\$HOME/.local/bin or USE_SUDO=1)"
    exit 1
  fi
}

ARCHIVE="kong_${KONG_VERSION}_linux_${ARCH}.tar.gz"
URL="https://github.com/${KONG_GITHUB_REPO}/releases/download/v${KONG_VERSION}/${ARCHIVE}"
CHECKSUMS_URL="https://github.com/${KONG_GITHUB_REPO}/releases/download/v${KONG_VERSION}/checksums.txt"

tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT

echo "→ downloading kong v${KONG_VERSION} (${OS}/${ARCH})"
curl -fsSL "$URL" -o "$tmpdir/${ARCHIVE}"

if curl -fsSL "$CHECKSUMS_URL" -o "$tmpdir/checksums.txt" 2>/dev/null; then
  expected="$(grep " ${ARCHIVE}$" "$tmpdir/checksums.txt" | awk '{print $1}')"
  if [[ -n "$expected" ]]; then
    if command -v sha256sum >/dev/null 2>&1; then
      actual="$(sha256sum "$tmpdir/${ARCHIVE}" | awk '{print $1}')"
    elif command -v shasum >/dev/null 2>&1; then
      actual="$(shasum -a 256 "$tmpdir/${ARCHIVE}" | awk '{print $1}')"
    fi
    if [[ -n "${actual:-}" && "$actual" != "$expected" ]]; then
      err "checksum mismatch for ${ARCHIVE}"
      exit 1
    fi
  fi
fi

tar -xzf "$tmpdir/${ARCHIVE}" -C "$tmpdir"
install_bin "$tmpdir/kong" "${KONG_INSTALL_DIR}/kong"

echo ""
echo "✓ kong v${KONG_VERSION} installed → ${KONG_INSTALL_DIR}/kong"
echo ""
"${KONG_INSTALL_DIR}/kong" version
echo ""
echo "Try it:"
echo "  kong list"
echo "  kong list --json | jq '.jobs[] | {user, schedule, command}'"

if [[ ":$PATH:" != *":${KONG_INSTALL_DIR}:"* ]]; then
  echo ""
  echo "Add to PATH:"
  echo "  export PATH=\"${KONG_INSTALL_DIR}:\$PATH\""
fi

#!/usr/bin/env sh
# stdix installer
# Usage: curl -fsSL https://raw.githubusercontent.com/stdix/stdix/main/install.sh | sh
# Override install directory: INSTALL_DIR=/usr/local/bin sh install.sh

set -e

REPO="stdix/stdix"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"

# ── Detect OS and arch ────────────────────────────────────────────────────────

OS="$(uname -s)"
ARCH="$(uname -m)"

case "$OS" in
  Linux)  GOOS="linux" ;;
  Darwin) GOOS="darwin" ;;
  *)
    echo "Unsupported OS: $OS" >&2
    exit 1
    ;;
esac

case "$ARCH" in
  x86_64)          GOARCH="amd64" ;;
  aarch64 | arm64) GOARCH="arm64" ;;
  *)
    echo "Unsupported architecture: $ARCH" >&2
    exit 1
    ;;
esac

BINARY="stdix-${GOOS}-${GOARCH}"

# ── Resolve latest release tag ────────────────────────────────────────────────

echo "Fetching latest stdix release..."
TAG="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
  | grep '"tag_name"' \
  | head -1 \
  | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')"

if [ -z "$TAG" ]; then
  echo "Failed to determine latest release tag." >&2
  exit 1
fi

echo "Installing stdix ${TAG} (${GOOS}/${GOARCH})..."

BASE_URL="https://github.com/${REPO}/releases/download/${TAG}"

# ── Download binary + checksums ───────────────────────────────────────────────

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

curl -fsSL -o "${TMP}/${BINARY}"        "${BASE_URL}/${BINARY}"
curl -fsSL -o "${TMP}/checksums.sha256" "${BASE_URL}/checksums.sha256"

# ── Verify checksum ───────────────────────────────────────────────────────────

cd "$TMP"

if command -v sha256sum >/dev/null 2>&1; then
  grep "${BINARY}" checksums.sha256 | sha256sum --check --status
elif command -v shasum >/dev/null 2>&1; then
  grep "${BINARY}" checksums.sha256 | shasum -a 256 --check --status
else
  echo "Warning: no sha256sum or shasum found — skipping checksum verification." >&2
fi

cd - >/dev/null

# ── Install ───────────────────────────────────────────────────────────────────

mkdir -p "$INSTALL_DIR"
install -m 755 "${TMP}/${BINARY}" "${INSTALL_DIR}/stdix"

echo ""
echo "stdix ${TAG} installed to ${INSTALL_DIR}/stdix"

# Hint if INSTALL_DIR is not on PATH
case ":${PATH}:" in
  *":${INSTALL_DIR}:"*) ;;
  *)
    echo ""
    echo "Add ${INSTALL_DIR} to your PATH:"
    echo "  export PATH=\"\$PATH:${INSTALL_DIR}\""
    ;;
esac

echo ""
echo "Get started:"
echo "  cd your-project"
echo "  stdix init --registry-url https://github.com/codref/stdix-registry/releases/latest/download/registry.db"
echo "  stdix sync"
echo "  stdix deploy"

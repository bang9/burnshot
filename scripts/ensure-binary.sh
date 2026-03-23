#!/usr/bin/env bash
set -euo pipefail

BINARY_NAME="burnshot"
INSTALL_DIR="$HOME/.local/bin"
REPO="bang9/burnshot"
VERSION_CHECK="cli"
BUILD_TARGET="./cmd/burnshot"

# Extract expected version from plugin.json
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PLUGIN_JSON="$SCRIPT_DIR/../.claude-plugin/plugin.json"

get_expected_version() {
  if command -v jq &>/dev/null; then
    jq -r '.version' "$PLUGIN_JSON"
  else
    grep '"version"' "$PLUGIN_JSON" | head -1 | sed 's/.*"version"[[:space:]]*:[[:space:]]*"\(.*\)".*/\1/'
  fi
}

get_current_version() {
  if [ "$VERSION_CHECK" = "cli" ] && command -v "$BINARY_NAME" &>/dev/null; then
    "$BINARY_NAME" --version 2>/dev/null || echo ""
  else
    echo ""
  fi
}

detect_platform() {
  local os arch
  os="$(uname -s | tr '[:upper:]' '[:lower:]')"
  arch="$(uname -m)"
  case "$arch" in
    x86_64|amd64) arch="amd64" ;;
    arm64|aarch64) arch="arm64" ;;
  esac
  echo "${os}-${arch}"
}

EXPECTED=$(get_expected_version)
CURRENT=$(get_current_version)

if [ "$CURRENT" = "$EXPECTED" ]; then
  exit 0
fi

PLATFORM=$(detect_platform)
URL="https://github.com/$REPO/releases/download/$EXPECTED/$BINARY_NAME-$PLATFORM"

mkdir -p "$INSTALL_DIR"
if curl -fsSL "$URL" -o "$INSTALL_DIR/$BINARY_NAME"; then
  chmod +x "$INSTALL_DIR/$BINARY_NAME"
else
  # Source fallback
  if command -v go &>/dev/null; then
    cd "$SCRIPT_DIR/.."
    go build -ldflags "-s -w -X main.version=$EXPECTED" -o "$INSTALL_DIR/$BINARY_NAME" "$BUILD_TARGET"
  fi
fi

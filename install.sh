#!/usr/bin/env bash
set -euo pipefail

REPO="bang9/burnshot"
BINARY_NAME="burnshot"
INSTALL_DIR="$HOME/.local/bin"

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

get_latest_version() {
  curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | sed 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\(.*\)".*/\1/'
}

main() {
  local version platform url
  version=$(get_latest_version)
  platform=$(detect_platform)
  url="https://github.com/$REPO/releases/download/$version/$BINARY_NAME-$platform"

  echo "Installing $BINARY_NAME $version ($platform)..."
  mkdir -p "$INSTALL_DIR"
  curl -fsSL "$url" -o "$INSTALL_DIR/$BINARY_NAME"
  chmod +x "$INSTALL_DIR/$BINARY_NAME"

  # Ensure PATH includes install dir
  if ! echo "$PATH" | grep -q "$INSTALL_DIR"; then
    echo "Add $INSTALL_DIR to your PATH:"
    echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
  fi

  echo "Done! Run: $BINARY_NAME"
}

main

#!/bin/sh
# Install heic-converter from GitHub releases.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/kkito0726/heic-converter/main/install.sh | sh
#
# Environment variables:
#   VERSION  release tag to install (default: latest)
#   BIN_DIR  install destination (default: /usr/local/bin if writable, else ~/.local/bin)
set -eu

REPO="kkito0726/heic-converter"
NAME="heic-converter"
VERSION="${VERSION:-latest}"

fail() {
  echo "Error: $1" >&2
  exit 1
}

command -v curl >/dev/null 2>&1 || fail "curl is required"

os=$(uname -s)
case "$os" in
  Darwin) os=darwin ;;
  Linux) os=linux ;;
  MINGW* | MSYS* | CYGWIN*) os=windows ;;
  *) fail "unsupported OS: $os (supported: macOS, Linux, Windows)" ;;
esac

arch=$(uname -m)
case "$arch" in
  x86_64 | amd64) arch=amd64 ;;
  aarch64 | arm64) arch=arm64 ;;
  *) fail "unsupported architecture: $arch (supported: amd64, arm64)" ;;
esac

if [ "$VERSION" = "latest" ]; then
  VERSION=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" |
    sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p')
  [ -n "$VERSION" ] || fail "could not resolve the latest release tag"
fi

ext=tar.gz
bin="$NAME"
if [ "$os" = "windows" ]; then
  ext=zip
  bin="$NAME.exe"
fi
asset="${NAME}_${os}_${arch}.${ext}"
url="https://github.com/$REPO/releases/download/$VERSION/$asset"

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

echo "Downloading $NAME $VERSION ($os/$arch)..."
curl -fsSL -o "$tmp/$asset" "$url" || fail "download failed: $url"

if [ "$ext" = "zip" ]; then
  command -v unzip >/dev/null 2>&1 || fail "unzip is required"
  unzip -q "$tmp/$asset" -d "$tmp"
else
  tar -xzf "$tmp/$asset" -C "$tmp"
fi
[ -f "$tmp/$bin" ] || fail "binary $bin not found in archive"

if [ -z "${BIN_DIR:-}" ]; then
  if [ "$os" = "windows" ]; then
    BIN_DIR="$HOME/bin"
  elif [ -w /usr/local/bin ]; then
    BIN_DIR=/usr/local/bin
  else
    BIN_DIR="$HOME/.local/bin"
  fi
fi
mkdir -p "$BIN_DIR"
cp "$tmp/$bin" "$BIN_DIR/$bin"
chmod +x "$BIN_DIR/$bin"

echo "Installed $NAME $VERSION to $BIN_DIR/$bin"
case ":$PATH:" in
  *":$BIN_DIR:"*) ;;
  *) echo "NOTE: $BIN_DIR is not in your PATH. Add it, e.g.: export PATH=\"$BIN_DIR:\$PATH\"" ;;
esac

#!/bin/sh
# Install hivetrack-mcp binary from GitHub releases.
# Usage: curl -fsSL https://raw.githubusercontent.com/The127/hivetrack/main/install-mcp.sh | sh
set -eu

REPO="The127/hivetrack"
BINARY="hivetrack-mcp"

detect_os() {
  case "$(uname -s)" in
    Linux*)  echo "linux" ;;
    Darwin*) echo "darwin" ;;
    MINGW*|MSYS*|CYGWIN*) echo "windows" ;;
    *) printf "Unsupported OS: %s\n" "$(uname -s)" >&2; exit 1 ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64)  echo "amd64" ;;
    aarch64|arm64)  echo "arm64" ;;
    *) printf "Unsupported architecture: %s\n" "$(uname -m)" >&2; exit 1 ;;
  esac
}

OS="$(detect_os)"
ARCH="$(detect_arch)"
EXT=""
if [ "$OS" = "windows" ]; then
  EXT=".exe"
fi

printf "Detected: %s/%s\n" "$OS" "$ARCH"

# Determine install directory
if [ "$(id -u)" = "0" ]; then
  INSTALL_DIR="/usr/local/bin"
else
  INSTALL_DIR="${HOME}/.local/bin"
  mkdir -p "$INSTALL_DIR"
fi

# Fetch latest release tag
if command -v curl >/dev/null 2>&1; then
  LATEST=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')
elif command -v wget >/dev/null 2>&1; then
  LATEST=$(wget -qO- "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')
else
  printf "Error: curl or wget required\n" >&2
  exit 1
fi

if [ -z "$LATEST" ]; then
  printf "Error: could not determine latest release\n" >&2
  exit 1
fi

printf "Latest release: %s\n" "$LATEST"

BASE_URL="https://github.com/${REPO}/releases/download/${LATEST}"
ASSET="${BINARY}-${OS}-${ARCH}${EXT}"
CHECKSUMS="checksums.txt"

# Download binary and checksums
TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

printf "Downloading %s...\n" "$ASSET"
if command -v curl >/dev/null 2>&1; then
  curl -fsSL -o "${TMPDIR}/${ASSET}" "${BASE_URL}/${ASSET}"
  curl -fsSL -o "${TMPDIR}/${CHECKSUMS}" "${BASE_URL}/${CHECKSUMS}"
else
  wget -q -O "${TMPDIR}/${ASSET}" "${BASE_URL}/${ASSET}"
  wget -q -O "${TMPDIR}/${CHECKSUMS}" "${BASE_URL}/${CHECKSUMS}"
fi

# Verify checksum
printf "Verifying checksum...\n"
EXPECTED=$(grep "${ASSET}" "${TMPDIR}/${CHECKSUMS}" | awk '{print $1}')
if [ -z "$EXPECTED" ]; then
  printf "Error: checksum not found for %s\n" "$ASSET" >&2
  exit 1
fi

if command -v sha256sum >/dev/null 2>&1; then
  ACTUAL=$(sha256sum "${TMPDIR}/${ASSET}" | awk '{print $1}')
elif command -v shasum >/dev/null 2>&1; then
  ACTUAL=$(shasum -a 256 "${TMPDIR}/${ASSET}" | awk '{print $1}')
else
  printf "Warning: no sha256sum or shasum found, skipping checksum verification\n" >&2
  ACTUAL="$EXPECTED"
fi

if [ "$EXPECTED" != "$ACTUAL" ]; then
  printf "Error: checksum mismatch\n  expected: %s\n  actual:   %s\n" "$EXPECTED" "$ACTUAL" >&2
  exit 1
fi

printf "Checksum OK\n"

# Install
chmod +x "${TMPDIR}/${ASSET}"
mv "${TMPDIR}/${ASSET}" "${INSTALL_DIR}/${BINARY}${EXT}"
printf "Installed to %s/%s%s\n" "$INSTALL_DIR" "$BINARY" "$EXT"

# Post-install instructions
printf "\n"
printf "Setup:\n"
printf "  export HIVETRACK_URL=https://your-hivetrack-instance.example.com\n"
printf "\n"
printf "MCP client config (e.g. claude_desktop_config.json):\n"
printf '  {\n'
printf '    "mcpServers": {\n'
printf '      "hivetrack": {\n'
printf '        "command": "%s/%s%s",\n' "$INSTALL_DIR" "$BINARY" "$EXT"
printf '        "env": { "HIVETRACK_URL": "https://your-hivetrack-instance.example.com" }\n'
printf '      }\n'
printf '    }\n'
printf '  }\n'

# Check PATH
case ":${PATH}:" in
  *":${INSTALL_DIR}:"*) ;;
  *) printf "\nNote: %s is not in your PATH. Add it:\n  export PATH=\"%s:\$PATH\"\n" "$INSTALL_DIR" "$INSTALL_DIR" ;;
esac

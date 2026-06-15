#!/bin/sh
set -e

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# Normalize architecture names
case "$ARCH" in
    x86_64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    i386|i686) ARCH="386" ;;
esac

VERSION="${VERSION:-latest}"
REPO="vibelog/vibelog"

if [ "$VERSION" = "latest" ]; then
    URL="https://github.com/${REPO}/releases/latest/download/vibelog_${OS}_${ARCH}.tar.gz"
else
    URL="https://github.com/${REPO}/releases/download/${VERSION}/vibelog_${OS}_${ARCH}.tar.gz"
fi

echo "Installing vibelog for ${OS}/${ARCH}..."
echo "URL: ${URL}"

# Create temp directory
TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

# Download and extract
curl -fsSL "$URL" -o "${TMP_DIR}/vibelog.tar.gz"
tar -xzf "${TMP_DIR}/vibelog.tar.gz" -C "$TMP_DIR"

# Determine install location
if [ -w /usr/local/bin ]; then
    INSTALL_DIR="/usr/local/bin"
else
    INSTALL_DIR="${HOME}/.local/bin"
    mkdir -p "$INSTALL_DIR"
fi

# Install binary
cp "${TMP_DIR}/vibelog" "${INSTALL_DIR}/vibelog"
chmod +x "${INSTALL_DIR}/vibelog"

echo "✓ vibelog installed to ${INSTALL_DIR}/vibelog"

# Check if install dir is in PATH
case ":$PATH:" in
    *":${INSTALL_DIR}:"*) ;;
    *)
        echo ""
        echo "⚠️  ${INSTALL_DIR} is not in your PATH."
        echo "   Add this to your shell profile:"
        echo "   export PATH="${INSTALL_DIR}:$PATH""
        ;;
esac

echo ""
echo "Get started:"
echo "  vibelog init       # Install hooks in current repo"
echo "  vibelog tui        # Browse sessions"
echo "  vibelog help       # Show all commands"

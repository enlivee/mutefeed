#!/usr/bin/env bash
set -e

REPO="stillunbuilt/mutefeed"
BINARY="mutefeed"
INSTALL_DIR="/usr/local/bin"

if [ ! -w "$INSTALL_DIR" ]; then
    INSTALL_DIR="$HOME/.local/bin"
    mkdir -p "$INSTALL_DIR"
fi

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
    x86_64|amd64)
        ARCH="amd64"
        ;;
    arm64|aarch64)
        ARCH="arm64"
        ;;
    *)
        echo "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

echo "=> Installing MuteFeed for $OS-$ARCH into $INSTALL_DIR..."

# If Go is installed and local source exists, install directly
if [ -f "go.mod" ] && command -v go >/dev/null 2>&1; then
    echo "=> Building from source..."
    go build -o "$INSTALL_DIR/$BINARY" .
    echo "=> Successfully installed: $INSTALL_DIR/$BINARY"
    exit 0
fi

LATEST_TAG=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
if [ -z "$LATEST_TAG" ]; then
    LATEST_TAG="v1.0.0"
fi

TARBALL="mutefeed_${OS}_${ARCH}.tar.gz"
URL="https://github.com/$REPO/releases/download/$LATEST_TAG/$TARBALL"

TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

echo "=> Downloading $URL..."
if curl -sSL --fail "$URL" -o "$TMP_DIR/$TARBALL"; then
    tar -xzf "$TMP_DIR/$TARBALL" -C "$TMP_DIR"
    mv "$TMP_DIR/$BINARY" "$INSTALL_DIR/$BINARY"
    chmod +x "$INSTALL_DIR/$BINARY"
    echo "=> Successfully installed MuteFeed $LATEST_TAG to $INSTALL_DIR/$BINARY"
else
    echo "=> Release download not found, attempting go install..."
    if command -v go >/dev/null 2>&1; then
        go install "github.com/$REPO@latest"
        echo "=> Installed via go install!"
    else
        echo "=> Error: Pre-built binary not found. Please install Go or check release URL."
        exit 1
    fi
fi

echo ""
echo "Quick start:"
echo "  mutefeed                    # Start web dashboard at http://localhost:8899"
echo "  mutefeed check <url>        # Test a link for noise and slop"
echo "  mutefeed test <feed-url>    # Inspect a full RSS feed in terminal"

#!/usr/bin/env bash
set -e

REPO="enlivee/mutefeed"
BINARY="mutefeed"
INSTALL_DIR="/usr/local/bin"

# Determine target directory
if [ ! -d "$INSTALL_DIR" ] || [ ! -w "$INSTALL_DIR" ]; then
    INSTALL_DIR="$HOME/.local/bin"
    mkdir -p "$INSTALL_DIR"
fi

# Detect Operating System
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "$OS" in
    darwin)
        OS="darwin"
        ;;
    linux)
        OS="linux"
        ;;
    *)
        echo "Error: Unsupported operating system: $OS (MuteFeed supports macOS and Linux)"
        exit 1
        ;;
esac

# Detect Architecture
ARCH="$(uname -m)"
case "$ARCH" in
    x86_64|amd64)
        ARCH="amd64"
        ;;
    arm64|aarch64)
        ARCH="arm64"
        ;;
    *)
        echo "Error: Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

echo "=> Installing MuteFeed ($OS-$ARCH) into $INSTALL_DIR..."

# If local repo source exists, build directly
if [ -f "go.mod" ] && grep -q "mutefeed" go.mod 2>/dev/null && command -v go >/dev/null 2>&1; then
    echo "=> Building from local source..."
    go build -ldflags="-s -w" -o "$INSTALL_DIR/$BINARY" .
    chmod +x "$INSTALL_DIR/$BINARY"
    echo "=> Successfully installed: $INSTALL_DIR/$BINARY"
    exit 0
fi

# Fetch latest release tag from GitHub
LATEST_TAG=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" 2>/dev/null | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
if [ -z "$LATEST_TAG" ]; then
    LATEST_TAG="v1.0.0"
fi

TARBALL="mutefeed_${OS}_${ARCH}.tar.gz"
URL="https://github.com/$REPO/releases/download/$LATEST_TAG/$TARBALL"

TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

echo "=> Downloading MuteFeed $LATEST_TAG from $URL..."
if curl -sSL --fail "$URL" -o "$TMP_DIR/$TARBALL"; then
    tar -xzf "$TMP_DIR/$TARBALL" -C "$TMP_DIR"
    
    # Locate binary in extracted archive
    if [ -f "$TMP_DIR/$BINARY" ]; then
        mv "$TMP_DIR/$BINARY" "$INSTALL_DIR/$BINARY"
    else
        FOUND_BIN=$(find "$TMP_DIR" -type f -name "$BINARY" | head -n 1)
        if [ -n "$FOUND_BIN" ]; then
            mv "$FOUND_BIN" "$INSTALL_DIR/$BINARY"
        else
            echo "Error: Could not locate $BINARY binary inside release archive."
            exit 1
        fi
    fi
    chmod +x "$INSTALL_DIR/$BINARY"
    echo "=> Successfully installed MuteFeed $LATEST_TAG to $INSTALL_DIR/$BINARY"
else
    echo "=> Pre-built binary release download not available, attempting go install..."
    if command -v go >/dev/null 2>&1; then
        go install "github.com/$REPO@latest"
        echo "=> Installed via go install!"
    else
        echo "=> Error: Could not download pre-built binary and Go compiler is not installed."
        echo "   Please download manually from https://github.com/$REPO/releases or install Go."
        exit 1
    fi
fi

# Check if target directory is in PATH
case ":$PATH:" in
    *":$INSTALL_DIR:"*) ;;
    *)
        echo ""
        echo "Note: $INSTALL_DIR is not currently in your PATH."
        echo "To use 'mutefeed' from any terminal, add it with:"
        echo "  export PATH=\"$INSTALL_DIR:\$PATH\""
        ;;
esac

echo ""
echo "MuteFeed installed successfully!"
echo ""
echo "Quick start:"
echo "  mutefeed                    # Start web dashboard at http://localhost:8899"
echo "  mutefeed check <url>        # Test a link for noise and slop"
echo "  mutefeed test <feed-url>    # Inspect a full RSS feed in terminal"
echo "  mutefeed --help             # View all commands and flags"

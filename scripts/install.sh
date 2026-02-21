#!/bin/bash
set -e

# 1. Setup Atlas Bin Directory
ATLAS_DIR="$HOME/.atlas/bin"
mkdir -p "$ATLAS_DIR"

# 2. Check for Go
if ! command -v go &> /dev/null; then
    echo "❌ Error: Go is not installed. Please install Go (1.25+) before running this script."
    exit 1
fi

# 3. Check for gobake
if ! command -v gobake &> /dev/null; then
    echo "🛰️  gobake not found in PATH, attempting to install..."
    go install github.com/fezcode/gobake/cmd/gobake@latest
    GOBIN=$(go env GOBIN)
    if [ -z "$GOBIN" ]; then
        GOBIN=$(go env GOPATH)/bin
    fi
    export PATH="$GOBIN:$PATH"
    
    if ! command -v gobake &> /dev/null; then
        echo "❌ Error: gobake could not be installed/found. Please install it manually."
        exit 1
    fi
fi

# 4. Clone and Build atlas.hub
TEMP_DIR=$(mktemp -d)
echo "🛰️  Bootstrapping atlas.hub..."
git clone https://github.com/fezcode/atlas.hub.git "$TEMP_DIR"
cd "$TEMP_DIR"

echo "🛰️  Building..."
gobake build

# 5. Detect System Info
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
if [ "$ARCH" = "x86_64" ]; then ARCH="amd64"; fi
if [ "$ARCH" = "aarch64" ]; then ARCH="arm64"; fi

BINARY="build/atlas.hub-$OS-$ARCH"

# 6. Relocate
if [ -f "$BINARY" ]; then
    mv "$BINARY" "$ATLAS_DIR/atlas.hub"
    chmod +x "$ATLAS_DIR/atlas.hub"
    echo "✅ atlas.hub installed to $ATLAS_DIR/atlas.hub"
else
    echo "❌ Error: Build failed, binary not found."
    exit 1
fi

# 7. Cleanup
cd "$HOME"
rm -rf "$TEMP_DIR"

echo "🚀 Starting Atlas Hub..."
"$ATLAS_DIR/atlas.hub"

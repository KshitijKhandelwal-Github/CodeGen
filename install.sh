cat > install.sh << 'EOF'
#!/bin/bash

# CodeGen Pro Installation Script
# Usage: curl -fsSL https://raw.githubusercontent.com/KshitijKhandelwal-Github/CodeGen/main/install.sh | bash

set -e

echo "════════════════════════════════════════════════════════"
echo "                 Installing CodeGen Pro"
echo "════════════════════════════════════════════════════════"
echo ""

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64)
        ARCH="amd64"
        ;;
    arm64|aarch64)
        ARCH="arm64"
        ;;
    *)
        echo "❌ Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

case $OS in
    darwin)
        OS="darwin"
        ;;
    linux)
        OS="linux"
        ;;
    *)
        echo "❌ Unsupported OS: $OS"
        exit 1
        ;;
esac

# Get latest release
VERSION=$(curl -s https://api.github.com/repos/KshitijKhandelwal-Github/CodeGen/releases/latest | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
if [ -z "$VERSION" ]; then
    VERSION="v0.1.0"
fi

echo "→ Detected: $OS/$ARCH"
echo "→ Latest version: $VERSION"
echo ""

# Download URL
BINARY_NAME="codegen-${OS}-${ARCH}"
if [ "$OS" = "darwin" ]; then
    BINARY_NAME="codegen-macos-${ARCH}"
fi

DOWNLOAD_URL="https://github.com/KshitijKhandelwal-Github/CodeGen/releases/download/${VERSION}/${BINARY_NAME}"

echo "→ Downloading from: $DOWNLOAD_URL"
echo ""

# Download binary
TMP_DIR=$(mktemp -d)
cd $TMP_DIR

if ! curl -fsSL "$DOWNLOAD_URL" -o codegen; then
    echo "❌ Download failed. Using build from source as fallback..."
    
    # Fallback: clone and build
    if command -v go >/dev/null 2>&1; then
        git clone https://github.com/YOUR_USERNAME/codegen-pro.git
        cd codegen-pro
        make build
        BINARY_PATH="bin/codegen"
    else
        echo "❌ Go is not installed. Cannot build from source."
        echo "   Install Go from: https://golang.org/dl/"
        exit 1
    fi
else
    BINARY_PATH="codegen"
    chmod +x codegen
fi

# Determine install location
if [ -w "/usr/local/bin" ]; then
    INSTALL_DIR="/usr/local/bin"
elif [ -w "$HOME/.local/bin" ]; then
    INSTALL_DIR="$HOME/.local/bin"
    mkdir -p "$INSTALL_DIR"
else
    INSTALL_DIR="$HOME/bin"
    mkdir -p "$INSTALL_DIR"
fi

echo "→ Installing to: $INSTALL_DIR"
echo ""

# Install binary
if [ -w "$INSTALL_DIR" ]; then
    cp $BINARY_PATH "$INSTALL_DIR/codegen"
else
    sudo cp $BINARY_PATH "$INSTALL_DIR/codegen"
fi

# Cleanup
cd /
rm -rf $TMP_DIR

# Verify installation
if command -v codegen >/dev/null 2>&1; then
    echo "✓ Installation successful!"
    echo ""
    codegen --version
    echo ""
    echo "════════════════════════════════════════════════════════"
    echo "                CodeGen is ready to use!"
    echo "════════════════════════════════════════════════════════"
    echo ""
    echo "Quick start:"
    echo "  1. cd /path/to/your/project"
    echo "  2. codegen init"
    echo "  3. codegen generate --readme --docker"
    echo ""
    echo "Documentation: https://github.com/KshitijKhandelwal-Github/CodeGen"
else
    echo "⚠️  Installation complete but 'codegen' not in PATH"
    echo ""
    echo "Add to your PATH by adding this to ~/.zshrc or ~/.bashrc:"
    echo "  export PATH=\"\$PATH:$INSTALL_DIR\""
    echo ""
    echo "Then run: source ~/.zshrc"
fi
EOF

chmod +x install.sh
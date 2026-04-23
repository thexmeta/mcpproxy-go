#!/bin/bash
# MCPProxy Debian x64 Release Build Script
# Creates a Debian package for x64 architecture
#
# Usage: ./scripts/build-debian.sh [version]
# Example: ./scripts/build-debian.sh v0.21.3

set -e

# Configuration
VERSION=${1:-$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.21.3")}
VERSION_NO_V=${VERSION#v}  # Remove leading 'v' for nfpm
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)
ARCH="amd64"
OUTPUT_DIR="releases/debian"
PACKAGE_NAME="mcpproxy"

echo "========================================"
echo "MCPProxy Debian x64 Build"
echo "========================================"
echo "Version: $VERSION"
echo "Commit: $COMMIT"
echo "Date: $BUILD_DATE"
echo "Architecture: $ARCH"
echo "========================================"

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Set LDFLAGS for Go build
LDFLAGS="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${BUILD_DATE} -X github.com/smart-mcp-proxy/mcpproxy-go/internal/httpapi.buildVersion=${VERSION}"

# Build frontend (required for embedded web UI)
echo ""
echo "Step 1: Building frontend..."
cd frontend
npm install --silent
npm run build
cd ..

# Copy frontend to web directory for embedding
rm -rf web/frontend
mkdir -p web/frontend
cp -r frontend/dist web/frontend/

# Build Go binary for Linux x64
echo ""
echo "Step 2: Building Go binary (linux/amd64)..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "$LDFLAGS" -o "$OUTPUT_DIR/mcpproxy" ./cmd/mcpproxy

# Verify binary was built
if [ ! -f "$OUTPUT_DIR/mcpproxy" ]; then
    echo "❌ Failed to build mcpproxy binary"
    exit 1
fi

echo "✓ Binary built successfully: $OUTPUT_DIR/mcpproxy"

# Build Debian package using nfpm
echo ""
echo "Step 3: Building Debian package..."

# Check if nfpm is installed
if ! command -v nfpm &> /dev/null; then
    echo "❌ nfpm not found. Installing..."
    if command -v brew &> /dev/null; then
        brew install nfpm
    elif command -v apt-get &> /dev/null; then
        # Try to install via apt (may not have latest version)
        sudo apt-get update
        sudo apt-get install -y nfpm
    else
        echo "Please install nfpm: https://nfpm.dev/generate/"
        exit 1
    fi
fi

# Set environment variables for nfpm
export NFPM_VERSION="$VERSION_NO_V"
export NFPM_ARCH="$ARCH"

# Build the .deb package
nfpm package --config packaging/linux/nfpm.yaml --packager deb --target "$OUTPUT_DIR/${PACKAGE_NAME}_${VERSION_NO_V}_linux_${ARCH}.deb"

# Verify package was built
if [ ! -f "$OUTPUT_DIR/${PACKAGE_NAME}_${VERSION_NO_V}_linux_${ARCH}.deb" ]; then
    echo "❌ Failed to build Debian package"
    exit 1
fi

echo "✓ Debian package built successfully"

# Show package info
echo ""
echo "Step 4: Package information..."
dpkg-deb --info "$OUTPUT_DIR/${PACKAGE_NAME}_${VERSION_NO_V}_linux_${ARCH}.deb"

# Summary
echo ""
echo "========================================"
echo "Build Summary"
echo "========================================"
echo "Package: ${PACKAGE_NAME}_${VERSION_NO_V}_linux_${ARCH}.deb"
echo "Location: $(pwd)/$OUTPUT_DIR/"
echo ""
echo "To install:"
echo "  sudo dpkg -i $OUTPUT_DIR/${PACKAGE_NAME}_${VERSION_NO_V}_linux_${ARCH}.deb"
echo ""
echo "To install system-wide (copy to /usr/local/bin):"
echo "  sudo cp $OUTPUT_DIR/mcpproxy /usr/local/bin/"
echo ""
echo "========================================"

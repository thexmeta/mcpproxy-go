#!/bin/bash
# MCPProxy Cross-Platform Release Build Script
# Usage: ./scripts/build-release.sh v0.21.3

set -e

VERSION=${1:-$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.21.3")}
COMMIT=$(git rev-parse --short HEAD)
BUILD_DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)

echo "========================================"
echo "MCPProxy Release Build"
echo "Version: $VERSION"
echo "Commit: $COMMIT"
echo "Date: $BUILD_DATE"
echo "========================================"

# Setup LDFLAGS
LDFLAGS="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${BUILD_DATE} -X github.com/smart-mcp-proxy/mcpproxy-go/internal/httpapi.buildVersion=${VERSION}"

# Create releases directory
RELEASES_DIR="releases/${VERSION}"
mkdir -p "$RELEASES_DIR"

# Build matrix
declare -a PLATFORMS=(
    "linux:amd64"
    "linux:arm64"
    "darwin:amd64"
    "darwin:arm64"
    "windows:amd64"
    "windows:arm64"
)

for PLATFORM in "${PLATFORMS[@]}"; do
    OS=${PLATFORM%%:*}
    ARCH=${PLATFORM##*:}
    
    echo ""
    echo "Building for ${OS}/${ARCH}..."
    
    # Set output names
    if [ "$OS" = "windows" ]; then
        CORE_BIN="mcpproxy.exe"
        TRAY_BIN="mcpproxy-tray.exe"
        ARCHIVE_EXT="zip"
    else
        CORE_BIN="mcpproxy"
        TRAY_BIN="mcpproxy-tray"
        ARCHIVE_EXT="tar.gz"
    fi
    
    # Build core binary
    echo "  Building core binary..."
    CGO_ENABLED=0 GOOS=$OS GOARCH=$ARCH go build -ldflags "$LDFLAGS" -o "$RELEASES_DIR/$CORE_BIN" ./cmd/mcpproxy
    
    # Build tray binary for macOS and Windows
    if [ "$OS" = "darwin" ] || [ "$OS" = "windows" ]; then
        echo "  Building tray binary..."
        if [ "$OS" = "darwin" ]; then
            # macOS requires CGO for tray
            CGO_ENABLED=1 GOOS=$OS GOARCH=$ARCH go build -ldflags "$LDFLAGS" -o "$RELEASES_DIR/$TRAY_BIN" ./cmd/mcpproxy-tray
        else
            # Windows also requires CGO
            CGO_ENABLED=1 GOOS=$OS GOARCH=$ARCH go build -ldflags "$LDFLAGS" -o "$RELEASES_DIR/$TRAY_BIN" ./cmd/mcpproxy-tray
        fi
    fi
    
    # Create archive
    ARCHIVE_BASE="mcpproxy-${VERSION#v}-${OS}-${ARCH}"
    echo "  Creating archive: ${ARCHIVE_BASE}.${ARCHIVE_EXT}"
    
    if [ "$OS" = "windows" ]; then
        # ZIP for Windows
        (cd "$RELEASES_DIR" && zip -q "../${ARCHIVE_BASE}.zip" $CORE_BIN $TRAY_BIN)
    else
        # TAR.GZ for Unix-like systems
        (cd "$RELEASES_DIR" && tar -czf "../${ARCHIVE_BASE}.tar.gz" $CORE_BIN $TRAY_BIN 2>/dev/null || tar -czf "../${ARCHIVE_BASE}.tar.gz" $CORE_BIN)
    fi
    
    # Clean up individual binaries
    rm -f "$RELEASES_DIR/$CORE_BIN" "$RELEASES_DIR/$TRAY_BIN" 2>/dev/null || true
    
    echo "  ✓ Completed ${OS}/${ARCH}"
done

echo ""
echo "========================================"
echo "Build Summary"
echo "========================================"
ls -lh "$RELEASES_DIR"/*.zip "$RELEASES_DIR"/*.tar.gz 2>/dev/null || true

echo ""
echo "Release artifacts created in: $RELEASES_DIR"
echo ""
echo "To create GitHub release:"
echo "  git push origin $VERSION"
echo "  gh release create $VERSION --notes-file releases/RELEASE_NOTES_${VERSION#v}.md $RELEASES_DIR/*"

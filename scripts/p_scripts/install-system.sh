#!/bin/bash
# MCPProxy System-wide Installation Script
# Installs mcpproxy to /usr/local/bin for system-wide access
#
# Usage: sudo ./scripts/install-system.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
BINARY="$PROJECT_DIR/releases/debian/mcpproxy"
INSTALL_DIR="/usr/local/bin"
SERVICE_FILE="$PROJECT_DIR/packaging/linux/mcpproxy.service"

echo "========================================"
echo "MCPProxy System Installation"
echo "========================================"

# Check if binary exists
if [ ! -f "$BINARY" ]; then
    echo "❌ Binary not found: $BINARY"
    echo "Please run './scripts/build-debian.sh' first"
    exit 1
fi

# Install binary
echo ""
echo "Step 1: Installing binary to $INSTALL_DIR..."
cp "$BINARY" "$INSTALL_DIR/mcpproxy"
chmod 755 "$INSTALL_DIR/mcpproxy"
echo "✓ Binary installed: $INSTALL_DIR/mcpproxy"

# Install systemd service
if [ -f "$SERVICE_FILE" ]; then
    echo ""
    echo "Step 2: Installing systemd service..."
    cp "$SERVICE_FILE" /etc/systemd/system/mcpproxy.service
    chmod 644 /etc/systemd/system/mcpproxy.service
    echo "✓ Service file installed"
    
    # Reload systemd
    if command -v systemctl >/dev/null 2>&1; then
        echo ""
        echo "Step 3: Enabling systemd service..."
        systemctl daemon-reload
        systemctl enable mcpproxy.service
        echo "✓ Service enabled"
    fi
fi

# Create config directory
echo ""
echo "Step 4: Setting up configuration..."
mkdir -p /etc/mcpproxy
if [ ! -f /etc/mcpproxy/mcp_config.json ]; then
    cp "$PROJECT_DIR/packaging/linux/mcp_config.json.example" /etc/mcpproxy/mcp_config.json
    echo "✓ Created default configuration: /etc/mcpproxy/mcp_config.json"
else
    echo "✓ Configuration already exists: /etc/mcpproxy/mcp_config.json"
fi

# Create data directory
mkdir -p /var/lib/mcpproxy
chown -R $(whoami):$(whoami) /var/lib/mcpproxy
echo "✓ Data directory created: /var/lib/mcpproxy"

# Verify installation
echo ""
echo "========================================"
echo "Installation Complete!"
echo "========================================"
echo ""
echo "Binary location: $(which mcpproxy)"
echo "Version: $(mcpproxy --version 2>&1 || echo 'unknown')"
echo ""
echo "To start the service:"
echo "  sudo systemctl start mcpproxy"
echo ""
echo "To enable on boot:"
echo "  sudo systemctl enable mcpproxy"
echo ""
echo "To check status:"
echo "  sudo systemctl status mcpproxy"
echo ""
echo "Configuration file: /etc/mcpproxy/mcp_config.json"
echo "Data directory: /var/lib/mcpproxy"
echo "========================================"

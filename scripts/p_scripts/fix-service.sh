#!/bin/bash
# Fix mcpproxy service file
set -e

SERVICE_FILE="/etc/systemd/system/mcpproxy.service"

echo "Fixing mcpproxy service file..."

# Backup original
cp "$SERVICE_FILE" "${SERVICE_FILE}.bak"

# Fix the path and user
sed -i 's|ExecStart=/usr/bin/mcpproxy|ExecStart=/usr/local/bin/mcpproxy|g' "$SERVICE_FILE"
sed -i 's|^User=mcpproxy|User=root|g' "$SERVICE_FILE"
sed -i 's|^Group=mcpproxy|Group=root|g' "$SERVICE_FILE"

echo "Service file updated:"
cat "$SERVICE_FILE" | grep -E "(User|Group|ExecStart)"

echo ""
echo "Now run:"
echo "  sudo systemctl daemon-reload"
echo "  sudo systemctl start mcpproxy"
echo "  sudo systemctl status mcpproxy"

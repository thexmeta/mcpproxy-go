# Debian x64 Build and Installation Guide

## Overview

This guide covers building and installing MCPProxy on Debian-based systems (x64 architecture).

## Prerequisites

- Go 1.24+ installed
- Node.js and npm installed (for frontend build)
- `nfpm` for package creation (auto-installed by build script)
- sudo privileges for system installation

## Build Artifacts

The build process creates two types of artifacts in `releases/debian/`:

1. **Standalone binary**: `mcpproxy` -可以直接运行的二进制文件
2. **Debian package**: `mcpproxy_0.21.3_linux_amd64.deb` - 用于系统安装的 .deb 包

## Quick Start

### Option 1: Using the Build Script (Recommended)

```bash
# Build the Debian package
./scripts/build-debian.sh v0.21.3

# Install system-wide (requires sudo)
sudo ./scripts/install-system.sh
```

### Option 2: Manual Build and Installation

#### Step 1: Build the binary

```bash
cd /mnt/Meta/Projects/Go/mcpproxy-go

# Build frontend (if needed)
cd frontend && npm install && npm run build
cd ..

# Build Go binary for Linux x64
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
  -ldflags "-s -w -X main.version=v0.21.3 -X main.commit=$(git rev-parse --short HEAD)" \
  -o releases/debian/mcpproxy \
  ./cmd/mcpproxy
```

#### Step 2: Install system-wide

```bash
# Copy binary to system path
sudo cp releases/debian/mcpproxy /usr/local/bin/
sudo chmod 755 /usr/local/bin/mcpproxy

# Install systemd service
sudo cp packaging/linux/mcpproxy.service /etc/systemd/system/
sudo chmod 644 /etc/systemd/system/mcpproxy.service
sudo systemctl daemon-reload
sudo systemctl enable mcpproxy

# Create configuration directory
sudo mkdir -p /etc/mcpproxy
sudo cp packaging/linux/mcp_config.json.example /etc/mcpproxy/mcp_config.json

# Create data directory
sudo mkdir -p /var/lib/mcpproxy
```

#### Step 3: Verify installation

```bash
# Check binary is accessible
which mcpproxy
mcpproxy --version

# Check service status
sudo systemctl status mcpproxy
```

## Installation Methods

### Method A: Using .deb Package

```bash
# Install using dpkg
sudo dpkg -i releases/debian/mcpproxy_0.21.3_linux_amd64.deb

# Or install using apt (handles dependencies automatically)
sudo apt install ./releases/debian/mcpproxy_0.21.3_linux_amd64.deb
```

### Method B: Using Install Script

```bash
sudo ./scripts/install-system.sh
```

### Method C: Manual Installation

See "Option 2: Manual Build and Installation" above.

## Post-Installation

### Start the service

```bash
sudo systemctl start mcpproxy
```

### Enable on boot

```bash
sudo systemctl enable mcpproxy
```

### Check status

```bash
sudo systemctl status mcpproxy
```

### View logs

```bash
sudo journalctl -u mcpproxy -f
```

### Configuration

The default configuration file is located at:

- `/etc/mcpproxy/mcp_config.json`

Edit this file to configure your MCP servers and proxy settings.

### Access the Web UI

Once running, access the web interface at:

```
http://localhost:8080/ui/
```

API documentation is available at:

```
http://localhost:8080/swagger/
```

## Uninstallation

### If installed via .deb package:

```bash
sudo apt remove mcpproxy
# or
sudo dpkg -r mcpproxy
```

### If installed manually:

```bash
sudo systemctl stop mcpproxy
sudo systemctl disable mcpproxy
sudo rm /usr/local/bin/mcpproxy
sudo rm /etc/systemd/system/mcpproxy.service
sudo rm -rf /etc/mcpproxy
sudo rm -rf /var/lib/mcpproxy
sudo systemctl daemon-reload
```

## Troubleshooting

### Binary won't start

Check if the binary is executable and in your PATH:

```bash
which mcpproxy
ls -lh /usr/local/bin/mcpproxy
```

### Service won't start

Check systemd service status:

```bash
sudo systemctl status mcpproxy
sudo journalctl -u mcpproxy -n 50
```

### Port already in use

If port 8080 is already in use, modify the configuration or use:

```bash
mcpproxy serve --listen 0.0.0.0:8081
```

## Build Script Details

### build-debian.sh

Creates both standalone binary and .deb package:

- Builds frontend assets
- Compiles Go binary for Linux x64
- Packages into .deb format using nfpm
- Output: `releases/debian/mcpproxy_0.21.3_linux_amd64.deb`

### install-system.sh

System-wide installation script:

- Copies binary to `/usr/local/bin`
- Installs systemd service
- Creates configuration directories
- Enables and starts service

## Development Mode

For development with hot-reload:

```bash
# Start frontend dev server
cd frontend && npm run dev

# In another terminal, start backend
go run ./cmd/mcpproxy serve --dev
```

## Testing

Run tests before deployment:

```bash
# Unit tests
go test ./...

# E2E tests
make test-e2e
```

## Support

For issues or questions:

- GitHub Issues: https://github.com/smart-mcp-proxy/mcpproxy-go/issues
- Documentation: https://mcpproxy.app/docs

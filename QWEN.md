# MCPProxy-Go Project Context

## Project Overview

**MCPProxy** is a Go-based desktop application that acts as a smart proxy for AI agents using the Model Context Protocol (MCP). It provides:

- **Intelligent tool discovery** – BM25 search across hundreds of MCP servers
- **Massive token savings** – ~99% reduction by loading one `retrieve_tools` function instead of hundreds of schemas
- **Built-in security quarantine** – Protection against Tool Poisoning Attacks
- **Cross-platform support** – macOS (Intel & Apple Silicon), Windows (x64 & ARM64), Linux (x64 & ARM64)
- **System tray UI** – Native menu bar application with real-time status

**Key Technologies:**
- Go 1.24 (toolchain go1.24.10)
- Vue 3.5 + TypeScript 5.9 + Vite 5 (frontend)
- Cobra CLI framework
- BBolt database (`~/.mcpproxy/config.db`)
- Chi router (HTTP API)
- Zap (structured logging)
- Bleve (BM25 search indexing)
- mark3labs/mcp-go (MCP protocol)

## Architecture

### Core Components

| Component | Purpose |
|-----------|---------|
| `cmd/mcpproxy/` | Core HTTP server + MCP proxy (headless daemon) |
| `cmd/mcpproxy-tray/` | System tray GUI application (CGO-based) |
| `internal/runtime/` | Lifecycle management, event bus, background services |
| `internal/httpapi/` | REST API endpoints (`/api/v1/*`) |
| `internal/upstream/` | 3-layer client: core → managed → CLI |
| `internal/management/` | Centralized server management service |
| `internal/storage/` | BBolt database layer |
| `internal/index/` | Bleve search index for tool discovery |
| `internal/oauth/` | OAuth 2.1 with PKCE (RFC 8252, RFC 8707) |
| `internal/security/` | Sensitive data detection (secrets, credentials, PII) |
| `frontend/` | Vue 3 + TypeScript web UI |

### Architecture Highlights

- **Core + Tray Split**: Tray app manages core server lifecycle via Unix sockets (macOS/Linux) or named pipes (Windows)
- **Event-Driven**: Real-time sync via SSE (`/events`) and event bus
- **Unified Management**: CLI, REST API, and MCP protocol share centralized `internal/management/service.go`
- **Docker Isolation**: Automatic containerization of stdio servers for security
- **Docker Recovery**: Automatic reconnection when Docker engine recovers from outage

## Building and Running

### Prerequisites

- Go 1.24+ (toolchain go1.24.10)
- Node.js 18.18+ (for frontend)
- npm (for frontend dependencies)

### Development Setup

```bash
# Install development dependencies
make dev-setup

# This installs:
# - swag (OpenAPI generator v2.0.0-rc4)
# - Frontend npm dependencies
# - Playwright browsers for E2E tests
```

### Build Commands

```bash
# Complete build (swagger + frontend + backend)
make build

# Individual components
make swagger         # Generate OpenAPI 3.1 spec
make frontend-build  # Build Vue frontend for production
make backend-dev     # Build backend with dev flag (loads frontend from disk)

# Manual Go build
go build -o mcpproxy ./cmd/mcpproxy
GOOS=darwin CGO_ENABLED=1 go build -o mcpproxy-tray ./cmd/mcpproxy-tray
```

### Running

```bash
# Core server (headless)
./mcpproxy serve
./mcpproxy serve --listen :8080
./mcpproxy serve --log-level=debug

# Tray application (auto-starts core)
./mcpproxy-tray

# Development mode (frontend hot reload)
./mcpproxy-dev serve
```

### Configuration

**Default Locations:**
- **Config**: `~/.mcpproxy/mcp_config.json`
- **Data**: `~/.mcpproxy/config.db` (BBolt)
- **Index**: `~/.mcpproxy/index.bleve/`
- **Logs**: `~/.mcpproxy/logs/`

**Minimal Config:**
```jsonc
{
  "listen": "127.0.0.1:8080",
  "data_dir": "~/.mcpproxy",
  "enable_tray": true,
  "mcpServers": [
    {
      "name": "local-python",
      "command": "python",
      "args": ["-m", "my_server"],
      "protocol": "stdio",
      "enabled": true
    }
  ]
}
```

**Environment Variables:**
- `MCPPROXY_LISTEN` – Override network binding
- `MCPPROXY_API_KEY` – API key for REST authentication
- `MCPPROXY_DEBUG` – Enable debug mode
- `HEADLESS` – Run without browser launching

## Testing

### Test Commands

```bash
# Unit tests
go test ./internal/... -v
go test -race ./internal/... -v

# With coverage
make test-coverage

# E2E tests
./scripts/test-api-e2e.sh           # Quick API E2E
./scripts/run-oauth-e2e.sh          # OAuth E2E with Playwright
./scripts/run-all-tests.sh          # Full test suite

# Frontend tests
cd frontend && npm run test
cd frontend && npm run coverage
```

### Linting

```bash
# Go linter (golangci-lint v1.59.1+)
./scripts/run-linter.sh
make lint

# Frontend linting
cd frontend && npm run lint
```

### OpenAPI Verification

```bash
# Regenerate and verify artifacts are committed
make swagger-verify
```

## CLI Commands

### Server Management

```bash
# List all servers with status
mcpproxy upstream list

# Enable/disable servers
mcpproxy upstream enable <name>
mcpproxy upstream disable <name>
mcpproxy upstream enable --all

# Restart servers
mcpproxy upstream restart <name>
mcpproxy upstream restart --all

# View logs
mcpproxy upstream logs <name> --tail=100 --follow

# Health diagnostics
mcpproxy doctor
```

### Tool Management (Per-Server)

```bash
# List tools for a specific server
mcpproxy tools list <server-name>

# Enable/disable individual tools
mcpproxy tools disable <server-name> <tool-name>
mcpproxy tools enable <server-name> <tool-name>
mcpproxy tools toggle <server-name> <tool-name>

# Rename tools (for better clarity or AI context)
mcpproxy tools rename <server-name> <old-tool-name> <new-name>
mcpproxy tools rename <server-name> <tool-name> --description="New description"

# Bulk operations
mcpproxy tools disable-all <server-name> --except=tool1,tool2
mcpproxy tools enable-all <server-name>

# View tool usage statistics
mcpproxy tools stats <server-name>
```

**Web UI:** Access tool management at `http://localhost:3303/ui/` → Select server → Tools tab

### Activity Log

```bash
mcpproxy activity list
mcpproxy activity list --type tool_call --status error
mcpproxy activity list --request-id <id>  # For error correlation
mcpproxy activity watch              # Real-time stream
mcpproxy activity show <id>
mcpproxy activity summary            # 24h statistics
mcpproxy activity export --output audit.jsonl
```

### Secrets Management

```bash
mcpproxy secrets set github_token
mcpproxy secrets list
mcpproxy secrets get github_token
mcpproxy secrets delete github_token
```

### OAuth Diagnostics

```bash
mcpproxy auth status --server=<name>
mcpproxy auth status --all
```

### Output Formatting

All CLI commands support multiple output formats:

```bash
mcpproxy upstream list -o json    # JSON for scripting
mcpproxy upstream list -o yaml    # YAML
mcpproxy upstream list --json     # Shorthand
```

**Environment:** `MCPPROXY_OUTPUT=json` sets default format

## HTTP API

**Base Path:** `/api/v1`

| Endpoint | Description |
|----------|-------------|
| `GET /api/v1/status` | Server status and statistics |
| `GET /api/v1/servers` | List all upstream servers |
| `POST /api/v1/servers/{name}/enable` | Enable/disable server |
| `POST /api/v1/servers/{name}/quarantine` | Quarantine server |
| `GET /api/v1/tools` | Search tools across servers |
| `GET /api/v1/activity` | List activity records |
| `GET /events` | SSE stream for live updates |

**Authentication:**
- Header: `X-API-Key: your-api-key`
- Query: `?apikey=your-api-key`

**Example:**
```bash
curl -H "X-API-Key: your-api-key" http://127.0.0.1:8080/api/v1/servers
curl "http://127.0.0.1:8080/events?apikey=your-api-key"
```

**Security Notes:**
- MCP endpoints (`/mcp`, `/mcp/`) are **unprotected** for client compatibility
- REST API requires authentication (auto-generated API key if not provided)
- API key persisted to `~/.mcpproxy/mcp_config.json`

## Key Features

### OAuth 2.1 Support

- Zero-config OAuth for most servers (auto-detection via 401 responses)
- RFC 8707 Resource Auto-Detection from server metadata
- RFC 8252 compliant dynamic port allocation
- PKCE security enabled by default
- Automatic token refresh and storage
- **Static Credentials Support**: Configure `client_id` and `client_secret` via keyring for reliable authentication
- **OAuth Bug Fix (v0.21.4)**: Fixed token persistence issue where `GetOAuthHandler()` was called from error instead of configured client

**GitHub Copilot MCP Configuration Example:**
```jsonc
{
  "name": "Github",
  "url": "https://api.githubcopilot.com/mcp/",
  "protocol": "streamable-http",
  "env": {
    "GITHUB_TOKEN": "${keyring:github_token}"
  },
  "oauth": {
    "client_id": "${keyring:github_client_id}",
    "client_secret": "${keyring:github_client_secret}"
  },
  "enabled": true
}
```

**Setup Commands:**
```bash
mcpproxy secrets set github_token
mcpproxy secrets set github_client_id
mcpproxy secrets set github_client_secret
mcpproxy auth login --server=Github
```

### Tool Management

Per-server tool customization for better AI agent control:

- **Enable/Disable Tools**: Fine-tune which tools are available per server
- **Rename Tools**: Improve clarity and AI context with custom names
- **Bulk Operations**: Enable/disable multiple tools at once
- **Usage Statistics**: Track which tools are used most
- **Web UI Integration**: Visual tool management interface

**Use Cases:**
- Disable dangerous tools (e.g., file deletion, code execution)
- Rename ambiguous tools for better AI understanding
- Reduce token usage by hiding unused tools
- Create server-specific tool presets

### Docker Security Isolation

- Automatic runtime detection (uvx→Python, npx→Node.js, etc.)
- Container-per-server isolation
- Environment variable passing
- Automatic cleanup on shutdown
- Recovery when Docker engine restarts

### Sensitive Data Detection

Automatic scanning for:
- Cloud credentials (AWS, GCP, Azure)
- Private keys (RSA, EC, OpenSSH, PGP)
- API tokens (GitHub, Stripe, OpenAI, Anthropic, etc.)
- Database credentials
- Credit cards (Luhn validated)
- High-entropy strings

### Code Execution

JavaScript (ES5.1+) tool for orchestrating multiple MCP tools:

```bash
mcpproxy code exec --code="({ result: input.value * 2 })" --input='{"value": 21}'
mcpproxy code exec --code="call_tool('github', 'get_user', {username: input.user})"
```

## Development Guidelines

### File Organization

- Use `internal/` subdirectories for encapsulated packages
- Follow Go conventions (gofmt, goimports)
- Unit tests in `*_test.go` alongside source files
- E2E tests in `internal/server/e2e_test.go`

### Coding Style

- Run `gofmt` / `goimports` on all Go sources
- Use descriptive package names (`runtime`, `upstream`, `storage`)
- TypeScript/Vue: Prettier defaults, PascalCase components
- Configuration/DTO structs in `internal/contracts`, `internal/httpapi`

### Testing Practices

- Write failing tests before implementing features
- Name Go tests `TestFeatureScenario`
- Use snapshot/fixture data under `tests/` with explicit prefixes
- Target specific suites: `go test ./internal/server -run TestMCP -v`

### Commit Guidelines

- Concise, imperative messages (e.g., `Fix upstream disable locking`)
- No AI co-author tags
- PR descriptions: summarize impact, verification commands, link issues
- Attach UI screenshots or log excerpts for behavior changes

## Debugging

```bash
# Quick diagnostics
mcpproxy doctor

# Server status
mcpproxy upstream list

# Server logs
mcpproxy upstream logs <name> --tail=100 --follow

# Main log (macOS)
tail -f ~/Library/Logs/mcpproxy/main.log

# Main log (Linux)
tail -f ~/.mcpproxy/logs/main.log

# Debug mode
mcpproxy serve --log-level=debug --tray=false
```

### Windows-Specific Debugging

```powershell
# Check if both processes are running
tasklist | findstr mcpproxy

# Kill all MCPProxy processes
taskkill /F /IM mcpproxy.exe /IM mcpproxy-tray.exe

# Check named pipe exists
Get-ChildItem \\.\pipe\ | Where-Object Name -like "*mcpproxy*"

# Check listening ports
netstat -ano | findstr :8080

# Run tray from project directory (important!)
cd E:\Projects\Go\mcpproxy-go
.\mcpproxy-tray.exe
```

**Windows Troubleshooting:**
- **Tray won't launch core:** Ensure both `mcpproxy.exe` and `mcpproxy-tray.exe` are in the same directory
- **Pipe not found errors:** Normal during first 2 seconds of startup (see `docs/lessons-learned.md`)
- **Port conflicts:** Check with `netstat -ano | findstr :8080`, change port in config if needed

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Success |
| `1` | General error |
| `2` | Port conflict |
| `3` | Database locked |
| `4` | Config error |
| `5` | Permission error |

## Important Notes

- **Database Locking**: Kill all existing instances before running (`pkill mcpproxy`)
- **Port Conflicts**: Default port 8080 may be in use; check with `lsof -i :8080`
- **CGO for Tray**: macOS tray binary requires `CGO_ENABLED=1`
- **Frontend Embedding**: Production builds embed frontend from `web/frontend/`
- **OpenAPI Artifacts**: Commit `oas/swagger.yaml` and `oas/docs.go` after regeneration

## Documentation

- **Architecture**: `docs/architecture.md`
- **Setup Guide**: `docs/setup.md`
- **Configuration**: `docs/configuration.md`
- **CLI Commands**: `docs/cli-management-commands.md`
- **Docker Isolation**: `docs/docker-isolation.md`
- **OAuth**: `docs/oauth-resource-autodetect.md`
- **Code Execution**: `docs/code_execution/`
- **Security**: `docs/features/security-quarantine.md`

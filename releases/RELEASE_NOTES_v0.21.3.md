# MCPProxy v0.21.3 Release Notes

**Release Date:** March 15, 2026

**Tag:** `v0.21.3`

## Overview

MCPProxy v0.21.3 is a maintenance release focusing on Windows compatibility improvements, MCP gateway integration enhancements, and bug fixes for server configuration handling.

## What's New

### MCP Gateway Skill
- Added comprehensive MCP Proxy Gateway connection skill documentation
- New skill file for AI agents to connect through MCP server endpoint (`/mcp`)
- Includes usage patterns, tool discovery workflows, and security features

### Windows Platform Improvements
- Fixed unresolved secret references in data directory expansion on Windows
- Fixed Windows backslash escaping in configuration tests
- Improved tray application stability on Windows

## Bug Fixes

### Core
- **fix:** CopyServerConfig missing SkipQuarantine and Shared fields - Server configuration now properly preserves quarantine bypass settings and shared server flags during copy operations

### Windows-Specific
- **fix:** Handle unresolved secret refs in data_dir on Windows - Proper handling of environment variable expansion in data directory paths
- **fix:** Escape Windows backslashes in TestLoadConfig_DataDirExpandFailure - Test stability improvements for Windows path handling

### MCP Integration
- **fix:** Adapt retrieve_tools instructions for code execution routing mode - Updated tool discovery instructions to properly route code execution requests

## Files in This Release

### Windows (amd64)
- `mcpproxy-v0.21.3-windows-amd64.zip` - Contains:
  - `mcpproxy.exe` (43.5 MB) - Core MCP proxy server
  - `mcpproxy-tray.exe` (31.0 MB) - System tray application

## Installation

### Windows
1. Download `mcpproxy-v0.21.3-windows-amd64.zip`
2. Extract to a directory (e.g., `C:\Program Files\MCPProxy\`)
3. Run `mcpproxy-tray.exe` for system tray integration
4. Or run `mcpproxy.exe serve` for headless operation

### Quick Start
```powershell
# Start the core server
.\mcpproxy.exe serve

# Or start with tray UI
.\mcpproxy-tray.exe
```

## Configuration

**Default Locations:**
- Config: `~/.mcpproxy/mcp_config.json`
- Data: `~/.mcpproxy/config.db`
- Logs: `~/.mcpproxy/logs/`

**Web UI:** http://localhost:8080/ui/
**API Docs:** http://localhost:8080/swagger/

## MCP Gateway Connection

Connect your AI agent to the MCP Proxy Gateway:

```typescript
import { Client } from "@modelcontextprotocol/sdk/client/index.js";
import { StreamableHTTPClientTransport } from "@modelcontextprotocol/sdk/client/streamableHttp.js";

const client = new Client({ name: "my-agent", version: "1.0.0" });
const transport = new StreamableHTTPClientTransport(
  new URL("http://127.0.0.1:8080/mcp")
);
await client.connect(transport);
```

See `skills/SKILL.md` for detailed usage patterns.

## Verification

Verify your installation:
```bash
# Check version
.\mcpproxy.exe --version

# Run diagnostics
.\mcpproxy.exe doctor

# Check server status
.\mcpproxy.exe upstream list
```

## Upgrade Notes

This is a compatible upgrade from v0.21.2. No migration required.

## Known Issues

- None reported in this release

## Support

- **Documentation:** https://docs.mcpproxy.app/
- **Issues:** https://github.com/smart-mcp-proxy/mcpproxy-go/issues
- **Discussions:** https://github.com/smart-mcp-proxy/mcpproxy-go/discussions

## Checksums

```
SHA256 (mcpproxy.exe) = [Run: certutil -hashfile mcpproxy.exe SHA256]
SHA256 (mcpproxy-tray.exe) = [Run: certutil -hashfile mcpproxy-tray.exe SHA256]
```

---

**Full Changelog:** https://github.com/smart-mcp-proxy/mcpproxy-go/compare/v0.21.2...v0.21.3

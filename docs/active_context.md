# Active Context - MCPProxy-Go

**Last Updated:** 2026-04-04
**Current Focus:** UI Navigation Fixed - Upstream Server Issues Pending Investigation

## Current State

### ✅ Completed This Session (2026-04-04)

1. **HTTP 404 Error Detection (v0.23.12)** ✅
   - Added 404 to isAuthError() and isOAuthError() checks
   - 404 errors now properly caught during connection

2. **Restart Features (v0.23.2 → v0.23.11)** ✅
   - Per-server restart button on server cards
   - Soft restart (MCP servers) in Configuration page
   - Hard restart (full process) in Configuration page
   - Tray support for exit code 100

3. **Telemetry Disabled (v0.23.9)** ✅
   - Default: DISABLED
   - No data collection or sending

4. **UI Navigation Fix (v0.23.15)** ✅
   - "Review Tools" button navigates to first server with pending tools
   - Server Detail page reads ?tab= query parameter
   - Dashboard links include ?tab=tools

### ❌ Unresolved Issues

1. **404 HTTP Errors from Upstream Servers**
   - NOT an mcpproxy bug
   - Server URLs are wrong or have migrated
   - User needs to verify URLs in config

2. **Timeout Errors for stdio Servers**
   - NOT an mcpproxy bug
   - MCP server processes not responding
   - User needs to run server commands manually

3. **Tools Not Showing**
   - UI navigation fixed
   - If server 404s, no tools discovered (upstream issue)
   - If quarantined, must be approved first

## Active State

### Running Processes
- **MCPProxy Tray:** Stopped
- **Core Server:** Stopped

### Database State
- **Path:** `C:\Users\eserk\.mcpproxy\config.db`
- **Config:** `C:\Users\eserk\.mcpproxy\mcp_config.json`

### Build Status
- **Latest:** v0.23.15
- **Deployed:** `D:\Development\CodeMode\mcpproxy-go\`

## Open Tasks

### High Priority (User Action Required)
- [ ] Verify upstream server URLs (404 errors)
- [ ] Test stdio server commands manually (timeout errors)
- [ ] Check if server endpoints have migrated

### Medium Priority (UI/UX)
- [ ] Add better error messages for 404 upstream servers
- [ ] Add troubleshooting guide for common server issues
- [ ] Add "Test Connection" button for servers

### Low Priority
- [ ] Add bulk tool approval from Servers list page
- [ ] Add server health dashboard
- [ ] Add notification when tools need approval

## Scratchpad

### Debug Commands
```powershell
# Check server config
Get-Content "C:\Users\eserk\.mcpproxy\mcp_config.json" | ConvertFrom-Json | Select-Object -ExpandProperty mcpServers | Select-Object name, url, command, args

# Test HTTP server
curl -v https://your-server-url/mcp

# Check logs for errors
Get-Content "D:\Development\bin\logs\*.log" -Tail 100 | Select-String "404|timeout|failed"
```

### Tool Approval Flow
```
1. Server connects → Tools discovered
2. checkToolApprovals() → Tools marked as "pending"
3. filterBlockedTools() → Pending tools blocked from index
4. User goes to Server Detail → Tools tab
5. User sees quarantined tools with "Approve" buttons
6. User approves → Tools added to index → Available for use
```

### Known Issues
- 404 errors are from upstream servers, not mcpproxy
- Timeout errors are from stdio servers not responding
- UI navigation is now fixed
- Quarantine is enabled by default (secure)

## Related Documentation
- Session Summary: `docs/session_summary.md`
- Architecture: `docs/architecture.md`
- CLI Commands: `docs/cli-management-commands.md`

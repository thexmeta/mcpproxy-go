# Session Summary - 2026-04-04

## Session Overview
**Date:** 2026-04-04
**Status:** Completed
**Releases:** v0.23.12 → v0.23.15 (4 releases)

---

## Key Achievements

### 1. Fix HTTP 404 Error Detection ✅
**Release:** v0.23.12
- Added "404", "Not Found" to `isAuthError()` checks
- Added "404" to `isOAuthError()` checks
- Files: `internal/upstream/core/connection_http.go`, `connection_oauth.go`

### 2. Restart Button (Per-Server) ✅
**Release:** v0.23.2
- Added restart button to each server card in Servers page
- Located between Logout and Details buttons
- Uses existing `/api/v1/servers/{id}/restart` endpoint

### 3. Proxy-Wide Restart Buttons ✅
**Releases:** v0.23.8 → v0.23.11
- Soft Restart: Restarts all MCP servers (management.RestartAll)
- Hard Restart: Full process restart with Windows process detachment
- UI: Two buttons in Configuration page (yellow Soft, red Hard)
- Tray support for exit code 100

### 4. Disable Telemetry by Default ✅
**Release:** v0.23.9
- `IsTelemetryEnabled()` returns false by default
- No heartbeat data, anonymous ID, or feedback sent
- Original code preserved in `.disabled` files

### 5. Review Tools Navigation Fix ✅
**Release:** v0.23.15
- "Review Tools" button now navigates to first server with pending tools
- Server Detail page reads `?tab=` query parameter
- Dashboard pending tools links include `?tab=tools`
- Direct path from Dashboard → Server Detail → Tools tab

---

## Unresolved Issues (Next Session)

### 1. 404 Errors from Upstream Servers
**Status:** Not fixed (upstream server issue)
- HTTP 404 errors from Avalonia, Serena, and other MCP servers
- Root cause: Wrong server URLs or migrated endpoints
- NOT an mcpproxy code issue
- Action: User needs to verify/update server URLs in config

### 2. Timeout Errors for stdio Servers
**Status:** Not fixed (upstream server issue)
- "context deadline exceeded" for Avalonia and Serena
- Root cause: MCP server process not responding to initialize()
- Possible causes: Not installed, crashing, missing dependencies
- Action: User needs to run server commands manually to debug

### 3. Tools Not Showing in Server Detail
**Status:** Partially fixed (UI navigation fixed)
- UI now correctly navigates to Tools tab
- Tools tab shows quarantined tools with approve buttons
- If server returns 404, no tools are discovered (upstream issue)
- If tools are quarantined, they must be approved first

---

## Releases Built & Deployed

| Version | Changes | Status |
|---------|---------|--------|
| v0.23.12 | 404 error detection | ✅ Deployed |
| v0.23.13 | Quarantine disabled (reverted) | ✅ Deployed |
| v0.23.14 | Quarantine reverted to enabled | ✅ Deployed |
| v0.23.15 | Review Tools navigation | ✅ Deployed |

---

## Architecture Changes

### Tool Quarantine Behavior
- Default: ENABLED (tools blocked until approved)
- `IsQuarantineEnabled()` returns true when nil
- Users must approve tools before they appear in tool list
- Can be disabled per-server with `skip_quarantine: true`

### Restart Architecture
```
Soft Restart: management.RestartAll() → MCP servers restart
Hard Restart (Tray): os.Exit(100) → Tray detects → Launches new core
Hard Restart (Standalone): exec.Command(exe, args...) → Detached process
```

### Navigation Flow
```
Dashboard → "Review Tools" → /servers/{firstServer}?tab=tools
Dashboard → Server name → /servers/{serverName}?tab=tools
Server Detail → Reads ?tab= query param → Auto-selects tab
```

---

## Files Modified This Session

**Backend:**
- `internal/config/config.go` - Quarantine default
- `internal/upstream/core/connection_http.go` - 404 detection
- `internal/upstream/core/connection_oauth.go` - 404 detection, OAuth handler
- `internal/server/server.go` - RequestRestart(), RequestHardRestart()
- `internal/httpapi/server.go` - /restart, /restart/hard endpoints

**Frontend:**
- `frontend/src/views/Dashboard.vue` - Review Tools button
- `frontend/src/views/ServerDetail.vue` - Tab query parameter
- `frontend/src/views/Settings.vue` - Soft/Hard restart buttons
- `frontend/src/services/api.ts` - restartProxyHard()

**Tray:**
- `cmd/mcpproxy-tray/main.go` - MCPPROXY_TRAY_PARENT env var
- `cmd/mcpproxy-tray/internal/monitor/process.go` - Exit code 100
- `cmd/mcpproxy-tray/internal/state/states.go` - EventCoreRestart
- `cmd/mcpproxy-tray/internal/state/machine.go` - State transition

---

## Deployment Target
📁 `D:\Development\CodeMode\mcpproxy-go\`

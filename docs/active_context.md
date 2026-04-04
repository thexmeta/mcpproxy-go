# Active Context - MCPProxy-Go

**Last Updated:** 2026-04-04 (Session 2 end)
**Current Focus:** Config file ↔ in-memory state sync COMPLETE. Tool display fix COMPLETE. Tool toggle race condition tracked as mcpproxy-go-807.

## What's Done This Session

### Architecture Fix (8 files)
- ✅ Config file is now authoritative source of truth for ALL server fields
- ✅ `SaveConfiguration()`: reads config snapshot → writes to storage + file
- ✅ `EnableServer()`/`QuarantineServer()`: update snapshot first, then save both
- ✅ `GetConfig()`: reads from `ConfigSnapshot()` not stale `r.cfg`
- ✅ `SkipQuarantine`/`Shared` fields added to `UpstreamRecord`
- ✅ `getDisabledToolsFromConfig()`: reads from config file directly

### UI Tool Display Fixes
- ✅ `/tools` endpoint always filters disabled tools (was returning all)
- ✅ `/tools/all` endpoint includes `enabled` field (was missing)
- ✅ `Tool.Enabled` field added to contracts struct + converter
- ✅ Config file had 13 disabled tools for Avalonia, server only has 19 tools total — 9 disabled tools no longer exist on the server (removed in server update)

### Test Fixes
- ✅ Fixed pre-broken `diagnostics_test.go` and `service_tool_preference_test.go` mocks
- ✅ Fixed `TestService_GetToolPreferences` expectation

### Integration Tests
- ✅ All 3 config sync tests pass
- ✅ All 18 servers show matching disabled_tools between `/api/v1/config` and `/api/v1/servers`

## Active State

### Running
- **MCPProxy:** Running at `127.0.0.1:3303` (manual start, not service)
- **Deployed binary:** `D:\Development\CodeMode\mcpproxy-go\mcpproxy.exe`
- **Config:** `C:\Users\eserk\.mcpproxy\mcp_config.json`
- **Database:** `C:\Users\eserk\.mcpproxy\config.db`

### Verified Working
- Config file → in-memory → storage sync: all fields match across all 3 stores
- `/api/v1/config` ↔ `/api/v1/servers` ↔ `/tools/all`: consistent disabled tools
- `/tools`: returns only enabled tools (disabled filtered out)
- All tests passing (except 3 pre-existing config test failures)

## Open Tasks

### High Priority
- [ ] **mcpproxy-go-807** — Fix tool toggle double-click race condition (P2 bug). Race between optimistic UI update and fetchServers() refresh. Fix: disable button during PATCH, skip fetchServers() after success, add loading spinner.

### Medium Priority
- [ ] **mcpproxy-go-37f** — Testing phase for Tool Management epic
- [ ] Clean up 9 stale disabled_tools in Avalonia config (tools no longer exist on server)

### Low Priority
- [ ] Add bulk tool approval from Servers list page
- [ ] Add server health dashboard
- [ ] Add notification when tools need approval

## Scratchpad

### Build/Deploy Commands
```powershell
# Quick rebuild and deploy
Stop-Service -Name 'MCP-Proxy' -Force; Start-Sleep -Seconds 3
go build -o mcpproxy.exe ./cmd/mcpproxy
copy /Y mcpproxy.exe "D:\Development\CodeMode\mcpproxy-go\mcpproxy.exe"
Start-Service -Name 'MCP-Proxy'; Start-Sleep -Seconds 10

# Or manual start
"D:\Development\CodeMode\mcpproxy-go\mcpproxy.exe" serve
```

### Known Issues
- Config file has 9 stale disabled_tools for Avalonia (tools removed from server)
- 3 pre-existing config test failures (unrelated to our changes)
- Management test pre-broken mocks fixed

## Related Documentation
- Session Summary: `docs/session_summary.md`
- Architecture: `docs/architecture.md`
- CLI Commands: `docs/cli-management-commands.md`

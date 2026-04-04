# Active Context - MCPProxy-Go

**Last Updated:** 2026-04-04
**Current Focus:** All major bugs fixed — pending push to remote (credential issue), minor UI polish remaining

## What's Done This Session

### Backend Fixes (12 commits)
1. ✅ `GET /tools/all` route — fixes 404 on server card tools page
2. ✅ `PATCH /config` route alias — fixes 404 on Exclude Disabled Tools toggle
3. ✅ `GET/PUT/DELETE /tools/preferences` — fixes 404 on tool enable/disable
4. ✅ `EnableServer()` sync — config file and API state now match immediately
5. ✅ `UpstreamRecord` fields — disabled_tools/ExcludeDisabledTools persisted to storage
6. ✅ `handlePatchServer` — uses management service (works for all server sources)
7. ✅ `UpdateServerDisabledTools` — now saves to storage too
8. ✅ Server card toggle — fetchServers() after success instead of relying on SSE

### UI Improvements
1. ✅ "Approval Required" stat card (between Total Tools and Quarantined)
2. ✅ Disabled/Excluded tools as full cards (same layout as enabled)
3. ✅ Enabled tools section with green badge
4. ✅ Enable/Disable buttons instead of toggle switches
5. ✅ Telemetry banner disabled

### DevOps
1. ✅ Deploy script: stop service → build → deploy → start service
2. ✅ No more mcpproxy-new.exe staging binary

### Epic Closure
- ✅ Tool Management Epic (mcpproxy-go-81y) — all tasks closed
- ✅ mcpproxy-go-cd7, mcpproxy-go-351, mcpproxy-go-d6o closed

## Active State

### Running
- **MCPProxy Service:** Running on `127.0.0.1:3303` (PID varies)
- **Deployed binary:** `D:\Development\CodeMode\mcpproxy-go\mcpproxy.exe` (61.5 MB)
- **Config:** `C:\Users\eserk\.mcpproxy\mcp_config.json`
- **Database:** `C:\Users\eserk\.mcpproxy\config.db`

### Verified Working
- API and config file sync: disabled_tools match in both
- Enable/disable toggle: API → config file → storage all in sync
- Server cards: tools display correctly with disabled_tools

## Open Tasks

### High Priority
- [ ] **Push to remote** — Blocked by GitHub credential mismatch (`thexmeta` vs `smart-mcp-proxy` org). 12 commits pending (~150 lines changed).
- [ ] **UI: Tool toggle double-click bug** — Sometimes clicking "Disable" says "enabled" and requires second click. Race condition between optimistic update and fetchServers refresh.

### Medium Priority
- [ ] **mcpproxy-go-37f** — Testing phase for Tool Management epic
- [ ] **Better error messages** for 404 upstream servers (user-side config issue, not mcpproxy bug)

### Low Priority
- [ ] Add bulk tool approval from Servers list page
- [ ] Add server health dashboard
- [ ] Add notification when tools need approval

## Scratchpad

### Build/Deploy Commands
```powershell
# Quick rebuild and deploy (service must be stopped first)
Stop-Service -Name 'MCP-Proxy' -Force; Start-Sleep -Seconds 3
go build -o mcpproxy.exe ./cmd/mcpproxy
copy /Y mcpproxy.exe "D:\Development\CodeMode\mcpproxy-go\mcpproxy.exe"
Start-Service -Name 'MCP-Proxy'; Start-Sleep -Seconds 10

# Or use deploy script (needs release zip)
powershell -ExecutionPolicy Bypass -File scripts\deploy.ps1 -Version v0.23.15
```

### Known Issues
- Push blocked: `remote: Permission to smart-mcp-proxy/mcpproxy-go.git denied to thexmeta`
- Upstream server 404s are NOT mcpproxy bugs — wrong URLs in user's config
- stdio server timeouts are NOT mcpproxy bugs — MCP processes not responding

## Related Documentation
- Session Summary: `docs/session_summary.md`
- Architecture: `docs/architecture.md`
- CLI Commands: `docs/cli-management-commands.md`

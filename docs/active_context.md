# Active Context - MCPProxy-Go

**Last Updated:** March 3, 2026  
**Session Status:** ✅ Complete - Windows tray fix implemented and tested

---

## Current Focus

**Completed:** Fixed Windows tray application failing to launch core server
- ✅ Named pipe race condition resolved (2-second initial delay)
- ✅ Shell wrapper split for platform-specific launch (cmd.exe on Windows, bash/zsh on macOS)
- ✅ Frontend built and embedded in binaries

**Next Session Priorities:**
1. Add Windows integration test for tray→core launch
2. Document Windows-specific troubleshooting
3. Add startup timing metrics

---

## Active State

### Build Artifacts
| Binary | Location | Status |
|--------|----------|--------|
| `mcpproxy.exe` | `E:\Projects\Go\mcpproxy-go\` | ✅ Built with fix |
| `mcpproxy-tray.exe` | `E:\Projects\Go\mcpproxy-go\` | ✅ Built with fix |

### Frontend
- **Location:** `web/frontend/dist/` (embedded in binaries)
- **Source:** `frontend/dist/` (Vue 3.5 + TypeScript + DaisyUI)
- **Status:** ✅ Built and embedded

### Configuration
- **Default Config Path:** `~/.mcpproxy/mcp_config.json` (Windows: `%USERPROFILE%\.mcpproxy\mcp_config.json`)
- **Default Data Dir:** `~/.mcpproxy/`
- **Default Listen Port:** `:8080` (or configured port, e.g., `:3303`)
- **Named Pipe (Windows):** `\\.\pipe\mcpproxy-<username>`

### Key Code Locations
| Component | Path |
|-----------|------|
| Tray Core Launcher | `cmd/mcpproxy-tray/main.go:1207-1280` |
| Windows Shell Wrapper | `cmd/mcpproxy-tray/main_windows.go:23-30` |
| macOS Shell Wrapper | `cmd/mcpproxy-tray/main_darwin.go:13-90` |
| Health Monitor | `cmd/mcpproxy-tray/internal/monitor/health.go` |
| Named Pipe Listener | `internal/server/listener_windows.go` |

---

## Scratchpad (Cleared)

**Cleared Items:**
- ~~Windows pipe race condition analysis~~ → Fixed
- ~~Shell wrapper debugging~~ → Fixed with platform split
- ~~Frontend placeholder~~ → Built proper frontend

---

## Testing Notes

### Verified Working
- ✅ Tray launches core on Windows
- ✅ Both processes run simultaneously
- ✅ Named pipe communication works after initial delay
- ✅ Web UI accessible at `http://127.0.0.1:<port>/ui/`
- ✅ All unit tests pass

### Known Configuration
- Core listening on port `3303` (from user's config)
- Web UI: `http://127.0.0.1:3303/ui/`

---

## Resume Command

```bash
cd E:\Projects\Go\mcpproxy-go && taskkill /F /IM mcpproxy*.exe 2>$null; .\mcpproxy-tray.exe
```

**For next session development:**
```bash
cd E:\Projects\Go\mcpproxy-go
go test ./cmd/mcpproxy-tray/... -v  # Run tray tests
go build -o mcpproxy-tray.exe ./cmd/mcpproxy-tray  # Rebuild tray
```

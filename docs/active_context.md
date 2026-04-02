# Active Context - MCPProxy-Go

**Last Updated:** 2026-04-03
**Current Focus:** Hard Restart Feature - COMPLETE ✅

## Current State

### ✅ Completed This Session (2026-04-03)

1. **Disable All Telemetry by Default (v0.23.9)** ✅
   - Config.IsTelemetryEnabled() returns false by default
   - No heartbeat data sent
   - No anonymous ID generated
   - No feedback submission
   - Original code preserved in .disabled files

2. **Add Hard Restart Feature (v0.23.10)** ✅
   - RequestHardRestart() in Server layer
   - POST /api/v1/restart/hard endpoint
   - Two buttons in Settings UI (Soft/Hard restart)
   - Tray support for exit code 100
   - Proper Windows process detachment

3. **Updated Tray.exe (v0.23.11)** ✅
   - Both mcpproxy.exe and mcpproxy-tray.exe rebuilt
   - Tray handles exit code 100 correctly
   - State machine: EventCoreRestart → StateLaunchingCore

### 🎯 Next Session Tasks

**Priority:** Medium

#### Testing
1. **Hard restart production test** - Verify full process restart works in real environment
2. **Tray restart verification** - Confirm tray properly restarts core after exit code 100
3. **Standalone mode test** - Test hard restart with custom config path

#### Enhancements
1. **Telemetry toggle UI** - Add enable/disable option in Settings page
2. **Restart confirmation logging** - Log all restart attempts and results
3. **Restart history/audit log** - Track when restarts occurred

#### Documentation
1. Add user guide for soft vs hard restart
2. Document telemetry privacy policy
3. Update API documentation with /restart/hard endpoint

## Active State

### Running Processes
- **MCPProxy Tray:** Stopped (ready for restart)
- **Core Server:** Stopped (ready for restart)
- **Build Status:** v0.23.11 deployed

### Database State
- **Path:** `C:\Users\eserk\.mcpproxy\config.db`
- **Status:** Active
- **Config:** `C:\Users\eserk\.mcpproxy\mcp_config.json`

### Build Artifacts
- **Location:** `releases/v0.23.11/` and `releases/mcpproxy-0.23.11-windows-amd64.zip`
- **Deployed:** `D:\Development\CodeMode\mcpproxy-go\`

### API Endpoints
| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/api/v1/restart` | POST | Soft restart (MCP servers) |
| `/api/v1/restart/hard` | POST | Hard restart (full process) |

## Open Tasks

### High Priority (Testing)
- [ ] Hard restart in production environment
- [ ] Tray restart after exit code 100
- [ ] Standalone hard restart with custom config

### Medium Priority (Enhancements)
- [ ] Telemetry enable/disable UI toggle
- [ ] Restart confirmation logging
- [ ] Restart history/audit log

### Low Priority (Documentation)
- [ ] User guide for restart types
- [ ] Telemetry privacy documentation
- [ ] API documentation update

## Scratchpad

### Resume Commands

```bash
# Start tray for testing
cd D:\Development\CodeMode\mcpproxy-go
.\mcpproxy-tray.exe

# Start standalone for testing
.\mcpproxy.exe serve --config "C:\Users\eserk\.mcpproxy\mcp_config.json" --log-level=debug --log-to-file --log-dir "D:\Development\bin\logs"

# Test soft restart
curl -X POST -H "X-API-Key: YOUR_KEY" http://127.0.0.1:8080/api/v1/restart

# Test hard restart
curl -X POST -H "X-API-Key: YOUR_KEY" http://127.0.0.1:8080/api/v1/restart/hard

# Check logs for restart messages
Get-Content "D:\Development\bin\logs\*.log" -Tail 100 | Select-String "RESTART"
```

### Code Snippets

```go
// HARD RESTART - Full process restart
func (s *Server) RequestHardRestart() error {
    runningUnderTray := os.Getenv("MCPPROXY_TRAY_PARENT") == "1"
    
    if runningUnderTray {
        s.StopServer()
        os.Exit(100)  // Signal tray to restart
    }
    
    // Standalone: spawn new process
    cmd := exec.Command(exe, args...)
    cmd.SysProcAttr = &syscall.SysProcAttr{
        CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
    }
    cmd.Start()
    cmd.Process.Release()
    time.Sleep(2 * time.Second)
    s.StopServer()
    os.Exit(0)
}
```

### Known Issues

None currently. All builds passing, features verified working.

## Related Documentation

- Session Summary: `session_summary_2026-04-03.md`
- Telemetry: `internal/telemetry/telemetry_disabled.go`
- Hard Restart: `internal/server/server.go` (RequestHardRestart)
- HTTP API: `internal/httpapi/server.go` (/restart/hard endpoint)
- Frontend: `frontend/src/views/Settings.vue` (restart buttons)
- Tray: `cmd/mcpproxy-tray/internal/monitor/process.go` (exit code 100)
- State Machine: `cmd/mcpproxy-tray/internal/state/machine.go`

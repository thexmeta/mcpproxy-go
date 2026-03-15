# Session Summary - Windows Tray Core Launch Fix

**Date:** March 3, 2026  
**Session Focus:** Fix Windows tray application failing to launch core server

---

## Key Achievements

### 1. Fixed Named Pipe Race Condition (Windows)

**Problem:** Tray health checks started immediately after launching core, but named pipe (`\\.\pipe\mcpproxy-<username>`) didn't exist yet, causing:
- `The system cannot find the file specified` errors
- 1-minute timeout waiting for core readiness
- Retry loops with excessive logging

**Solution:** Added 2-second initial delay in `WaitForReady()` on Windows to allow core server time to create the named pipe listener.

**Files Modified:**
- `cmd/mcpproxy-tray/internal/monitor/health.go`
  - Added `isWindows()` helper function
  - Added initial delay in `WaitForReady()` for Windows
  - Enhanced error handling for named pipe errors
  - Suppressed "pipe not found" warnings during startup

### 2. Fixed Windows Core Launch Shell Wrapper

**Problem:** `wrapCoreLaunchWithShell()` used Unix shell syntax (`/bin/bash -l -c "exec ..."`) which doesn't exist on Windows, causing tray to fail launching core.

**Solution:** Split platform-specific shell wrapper into separate files:
- **macOS:** Uses user's shell (`/bin/zsh`, `/bin/bash`) with `-l -c` flags
- **Windows:** Uses `cmd.exe /c` with proper Windows command-line quoting

**Files Modified:**
- `cmd/mcpproxy-tray/main_darwin.go` (NEW) - macOS-specific shell wrapper
- `cmd/mcpproxy-tray/main_windows.go` - Added Windows-specific shell wrapper
- `cmd/mcpproxy-tray/main.go` - Removed platform-specific functions

### 3. Built Frontend for Web UI

**Problem:** Web UI showed empty page because frontend wasn't built (only placeholder existed).

**Solution:** Built Vue.js frontend and embedded in binaries.

**Commands Executed:**
```bash
cd frontend && npm install && npm run build
# Copied dist/ to web/frontend/dist/
go build -o mcpproxy.exe ./cmd/mcpproxy
go build -o mcpproxy-tray.exe ./cmd/mcpproxy-tray
```

---

## Critical Fix Details

### Why the Named Pipe Delay Was Needed

Windows named pipes require the **server** (core) to call `ListenPipe()` before the **client** (tray) can connect. The sequence:

1. Tray launches core process
2. Core starts goroutine → calls `Start()` → creates HTTP server → creates named pipe listener
3. **Race:** Tray health checks started at step 1, pipe created at step 3
4. **Result:** `ERROR_FILE_NOT_FOUND` for 60 seconds until timeout

The 2-second delay gives the core enough time to complete startup and create the pipe before health checks begin.

### Why Shell Wrapper Split Was Critical

The original `wrapCoreLaunchWithShell()` in `main.go` was called on **all platforms** but only worked on Unix:

```go
// Unix-only code that was running on Windows too:
shellPath, _ := selectUserShell()  // Returns error on Windows (no /bin/bash)
return shellPath, []string{"-l", "-c", command}, nil  // Windows cmd.exe doesn't understand -l -c
```

This caused the core launch to fail silently, leaving the tray running without the core server.

---

## Test Results

### Before Fix
```
tasklist | findstr mcpproxy
mcpproxy-tray.exe    12916    # Only tray running, core failed to start
```

### After Fix
```
tasklist | findstr mcpproxy
mcpproxy-tray.exe    32776    # Tray running
mcpproxy.exe         30240    # Core successfully launched!
```

### Build Verification
- ✅ `go build ./cmd/mcpproxy/...` - Success
- ✅ `go build ./cmd/mcpproxy-tray/...` - Success
- ✅ `go test ./cmd/mcpproxy-tray/internal/api/...` - All 27 tests pass
- ✅ `go test ./internal/socket/...` - All tests pass

---

## Open Tasks

- [ ] Add integration test for Windows tray core launch
- [ ] Consider making initial delay configurable via environment variable
- [ ] Document Windows-specific troubleshooting in docs/
- [ ] Add startup timing metrics to observe pipe creation delay

---

## Files Changed Summary

| File | Change Type | Description |
|------|-------------|-------------|
| `cmd/mcpproxy-tray/internal/monitor/health.go` | Modified | Added Windows pipe delay + error handling |
| `cmd/mcpproxy-tray/internal/api/client.go` | Modified | Suppress pipe errors during startup |
| `cmd/mcpproxy-tray/main_darwin.go` | Created | macOS shell wrapper functions |
| `cmd/mcpproxy-tray/main_windows.go` | Modified | Windows shell wrapper functions |
| `cmd/mcpproxy-tray/main.go` | Modified | Removed platform-specific functions |
| `web/frontend/dist/*` | Built | Vue.js frontend build artifacts |

---

## Next Session Tasks

Run with `bd` (Claude Code task tool):

1. **Add Windows integration test** for tray launching core
2. **Document Windows troubleshooting** in `docs/troubleshooting-windows.md`
3. **Add startup metrics** to measure pipe creation time
4. **Review log output** for any remaining Windows-specific issues

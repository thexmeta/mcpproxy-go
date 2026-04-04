# Session Summary - 2026-04-04 (Session 2)

## Key Achievements

### Critical Architecture Fix: Config File ↔ In-Memory State Sync
**Problem:** Config file (`mcp_config.json`) and in-memory/BBolt state diverged on every server change. The system used a circular dependency where `SaveConfiguration()` read from storage as source of truth, but `LoadConfiguredServers()` used config file as source of truth. Fields like `SkipQuarantine` and `Shared` were silently dropped during storage writes.

**Fix (8 files changed):**
1. Added `SkipQuarantine` and `Shared` fields to `UpstreamRecord` (storage/models.go)
2. Rewrote `SaveConfiguration()` — config snapshot is now source of truth, writes to both storage AND config file
3. Rewrote `EnableServer()`/`QuarantineServer()` — update config snapshot first, then save both stores
4. Rewrote `syncServerToStorage()` — copies full server config from snapshot, not just 2 fields
5. Fixed `UpdateServerDisabledTools()` — now updates configSvc snapshot too
6. Fixed `GetConfig()` — reads from `ConfigSnapshot()` instead of stale `r.cfg`
7. Added `IsToolDisabled`/`DisableTool`/`EnableTool` helper methods to `ServerConfig`

### UI Tool Display Fix: `/tools` Endpoint Was Returning Disabled Tools
**Problem:** `GET /servers/{name}/tools` returned ALL tools (enabled + disabled) because it only filtered when `exclude_disabled_tools` was set. Disabled tools showed as "enabled" in the UI.

**Fix:** `GetServerTools()` now always filters out disabled tools — this endpoint is "enabled tools only" by definition.

### UI Tool Display Fix: `enabled` Field Missing from Tool Response
**Problem:** `/servers/{name}/tools/all` returned tools without `enabled` field, so the UI couldn't distinguish enabled vs disabled tools.

**Fix:**
- Added `Enabled` field to `contracts.Tool` struct
- Updated `ConvertGenericToolsToTyped` to extract `enabled` from raw maps
- `GetAllServerTools()` reads `DisabledTools` from config file (authoritative) and marks tools accordingly

### UI Tool Display Fix: Config File Had 13 Disabled Tools But Only 4 Existed
**Root cause identified:** The Avalonia MCP server binary no longer has 9 of the 13 tools listed in `disabled_tools` (e.g., `perform_health_check`, `create_avalonia_project`). The config file has stale entries from a previous server version. The UI correctly shows only the 4 disabled tools that actually exist on the server.

### New Beads Issue
- **mcpproxy-go-807** (P2 bug): "Fix tool toggle double-click race condition" — tracked for next session

### Test Fixes
- Fixed pre-broken `diagnostics_test.go` — added `StorageManager()` to `mockRuntimeOperations`
- Fixed pre-broken `service_tool_preference_test.go` — added `StorageManager()` to `mockRuntime`
- Fixed `TestService_GetToolPreferences` — corrected expectation (returns empty map without storage)

## Files Changed This Session

| File | Change |
|------|--------|
| `internal/storage/models.go` | Added `SkipQuarantine`, `Shared` to `UpstreamRecord` |
| `internal/storage/manager.go` | Updated 4 conversion methods for new fields |
| `internal/runtime/lifecycle.go` | `SaveConfiguration`: config snapshot → storage + file; `EnableServer`/`QuarantineServer`: snapshot first, then save |
| `internal/runtime/runtime.go` | `GetConfig()`: reads from `ConfigSnapshot`; `getDisabledToolsFromConfig()`: reads from config file; `GetServerTools()`: always filters disabled; `UpdateServerDisabledTools()`: updates configSvc |
| `internal/management/service.go` | `syncServerToStorage`: copies full server config, not just 2 fields |
| `internal/config/config.go` | Added `IsToolDisabled`, `DisableTool`, `EnableTool` helpers |
| `internal/contracts/types.go` | Added `Enabled` field to `Tool` struct |
| `internal/contracts/converters.go` | Extract `enabled` field in `ConvertGenericToolsToTyped` |
| `internal/management/service_test.go` | Added `StorageManager()` to mock |
| `internal/management/service_tool_preference_test.go` | Added `StorageManager()` to mock; fixed broken test |

## Architecture Decisions
- **Config file is authoritative source of truth** for all server fields at all times
- `SaveConfiguration()`: reads config snapshot → writes to storage + file → updates in-memory
- `GetServerTools()` returns only enabled tools (disabled tools filtered out always)
- `GetAllServerTools()` returns all tools with `enabled` field from config file
- `getDisabledToolsFromConfig()` reads directly from config file, not stale in-memory

## Test Results
- `go build ./...` — Clean
- `go test ./internal/storage/...` — All PASS
- `go test ./internal/runtime/...` — All PASS
- `go test ./internal/management/...` — All PASS (pre-broken tests fixed)
- `go test ./internal/config/...` — 3 pre-existing failures (unrelated)

## Integration Test Results
- Test 1: Toggle server enabled via API → config file + API match ✅
- Test 2: Edit config file → restart → API picks up change ✅
- Test 3: PATCH disabled_tools → config file + API match ✅
- `/api/v1/config` disabled_tools count matches `/api/v1/servers` for all 18 servers ✅
- `/tools` endpoint filters disabled tools ✅
- `/tools/all` endpoint includes `enabled` field ✅

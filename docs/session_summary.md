# Session Summary - 2026-04-04

## Key Achievements

### Bug Fixes
1. **404 on server card tools page** — Added `GET /tools/all` route and `handleGetAllServerTools` handler. Frontend calls `/tools/all` but route was never registered.
2. **404 on Exclude Disabled Tools toggle** — Added `PATCH /config` route alias. Frontend called `/config` but only `/` existed.
3. **404 on tool enable/disable toggle** — Added `GET/PUT/DELETE /tools/preferences` routes and handlers (`handleGetToolPreferences`, `handleUpdateToolPreference`, `handleDeleteToolPreference`).
4. **Enable/disable toggle not reflecting in config file** — Made `LoadConfiguredServers()` run synchronously in `EnableServer()` instead of async goroutine.
5. **Config file disabled_tools not shown in UI** — Added `DisabledTools` and `ExcludeDisabledTools` to `UpstreamRecord` struct and all storage read/write paths.
6. **Server card toggle switch not updating** — Store's `enableServer`/`disableServer` now explicitly calls `fetchServers(true)` after success instead of relying on SSE.

### UI Improvements
1. **Approval Required stat** — New stat card between "Total Tools" and "Quarantined" showing total pending+changed tools count.
2. **Disabled tools as cards** — Disabled and excluded tools now render as full cards (same as enabled tools) with Enable/Include buttons.
3. **Enabled tools section** — Separate section with green header and badge, appears before disabled tools.
4. **Tool toggle buttons** — Replaced switch UI with explicit "Disable" (red) / "Enable" (green) buttons.
5. **Telemetry banner removed** — Disabled TelemetryBanner.vue.

### Deploy Script
1. **Service restart added** — Deploy script now stops service before copying binary, starts after.
2. **No staging binary** — Removed `mcpproxy-new.exe` staging pattern. Stop service → build to `mcpproxy.exe` → deploy → start.

### Epic Closure
- **Tool Management Epic (mcpproxy-go-81y)** — All 6 subtasks closed (7v7, 3en, cd7, 351, d6o, 37f remaining for testing).

## Files Changed (by category)

### Backend — Route/Handler fixes
- `internal/httpapi/server.go` — Added `/tools/all` route, `/config` alias, tool preferences handlers, `GetAllServerTools` interface method, `handlePatchServer` rewrite to use management service.
- `internal/server/server.go` — Added `GetAllServerTools`, updated `UpdateServer` to apply `ExcludeDisabledTools` and `DisabledTools`.

### Backend — Storage persistence
- `internal/storage/models.go` — Added `DisabledTools`, `ExcludeDisabledTools` to `UpstreamRecord`.
- `internal/storage/manager.go` — Updated all read/write paths to include these fields.

### Backend — Config sync
- `internal/runtime/lifecycle.go` — Made `EnableServer()` synchronous for `LoadConfiguredServers`.
- `internal/runtime/runtime.go` — `UpdateServerDisabledTools` now saves to storage. `GetAllServerTools` returns `disabled_tools` from config.
- `internal/management/service.go` — Added `syncServerToStorage`, `StorageManager()` to interface, `disabled_tools` extraction in GetAllServers, patch handling.
- `internal/contracts/types.go` — Added `DisabledTools` to `Server` struct.

### Frontend
- `frontend/src/types/api.ts` — Added `disabled_tools` to Server interface.
- `frontend/src/services/api.ts` — Added `setDisabledTools` method.
- `frontend/src/stores/servers.ts` — Added `totalQuarantinedTools` computed, `fetchServers(true)` after enable/disable.
- `frontend/src/views/Servers.vue` — Added "Approval Required" stat card.
- `frontend/src/views/ServerDetail.vue` — Reorganized tools into Enabled/Disabled/Excluded sections with card rendering, added `toggleDisabledTool`, `filteredDisabledTools`, `enabledToolsList` computed.
- `frontend/src/components/TelemetryBanner.vue` — Disabled (empty template).

### DevOps
- `scripts/deploy.ps1` — Service stop before copy, start after, proper error handling.

## Architecture Decisions
- **Config file is source of truth**: At startup, config file → storage → API. Live changes write to both config file and storage synchronously.
- **Synchronous enable/disable**: `EnableServer()` must complete `LoadConfiguredServers` before returning so the API reflects the new state immediately.

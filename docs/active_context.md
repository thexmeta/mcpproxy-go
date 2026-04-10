# Active Context - MCPProxy-Go

**Last Updated:** 2026-04-10 (Session end)
**Current Focus:** All requested fixes completed and deployed. Ready for next tasks.

## What's Done This Session

### 1. Telemetry Banner Removal — DONE
- Removed `<TelemetryBanner />` from `Dashboard.vue` (usage + import)
- Rebuilt frontend and re-embedded in Go binary

### 2. Repositories & Configuration Sections Fixed — DONE
- Changed sidebar styling from muted (`text-base-content/70`) to active (`font-medium`)
- **Root cause**: Stale `web/frontend/dist/` was embedded in Go binary. Fresh `frontend/dist/` was never copied to `web/frontend/dist/` before `go build`.
- Fixed both `build.ps1` and `deploy.ps1` to include the copy step

### 3. TypeScript Build Errors Fixed — DONE
- Fixed missing `})` in `ServerDetail.vue` `excludedToolsList` computed (TS1005)
- Fixed duplicate `tabParam` variable in `ServerDetail.vue` `onMounted()` (TS2451)

### 4. Merge Conflict Resolved — DONE
- `oas/docs.go` — accepted `--theirs` (auto-generated swagger file)

### 5. Build & Deploy Scripts Updated — DONE
- `scripts/build.ps1` — now does: npm install → npm run build → copy to web/frontend/dist → go build
- `scripts/deploy.ps1` — complete rewrite: frontend build → copy → go build → stop service → deploy → restart

## Active State

### Running
- **MCPProxy:** Service running on `127.0.0.1:3303`
- **Deployed binary:** `D:\Development\CodeMode\mcpproxy-go\mcpproxy.exe`
- **Config:** `C:\Users\eserk\.mcpproxy\mcp_config.json`
- **Database:** `C:\Users\eserk\.mcpproxy\config.db`

### Git State
- **Branch:** main, 70+ commits ahead of origin/main
- **Push:** Still blocked by 403 credential issue (needs credential fix for `smart-mcp-proxy/mcpproxy-go`)

### OAS Conflict
- `oas/docs.go` resolved with `--theirs`, staged but not yet committed

## Open Tasks

### Medium Priority
- [ ] Fix GitHub push credentials (403 error on `smart-mcp-proxy/mcpproxy-go`)
- [ ] Clean up 3 stale disabled_tools in Avalonia config: `force_garbage_collection`, `generate_localization_system`, `create_avalonia_project`

### Low Priority
- [ ] Update `baseline-browser-mapping` npm package (out of date warning)
- [ ] Update `caniuse-lite` browserslist database (7 months old warning)
- [ ] Consider adding `make frontend-build` equivalent as a PowerShell function for cross-platform consistency

## Scratchpad

### Build/Deploy Command (Now Scripted)
```powershell
# Full build + deploy pipeline (recommended)
cd E:\Projects\Go\mcpproxy-go
powershell -ExecutionPolicy Bypass -File scripts\deploy.ps1 -Version v0.23.15
```

### Manual Build Steps (If Needed)
```powershell
# Step 1: Build frontend
cd E:\Projects\Go\mcpproxy-go\frontend && npm run build

# Step 2: CRITICAL - Copy to embed directory
Remove-Item -Recurse -Force ..\web\frontend; New-Item -ItemType Directory -Path ..\web\frontend -Force; Copy-Item -Recurse dist ..\web\frontend\dist

# Step 3: Build Go binary
cd E:\Projects\Go\mcpproxy-go && go build -ldflags "-s -w" -o mcpproxy.exe ./cmd/mcpproxy

# Step 4: Deploy
Stop-Service -Name 'MCP-Proxy' -Force; Copy-Item mcpproxy.exe D:\Development\CodeMode\mcpproxy-go\mcpproxy.exe -Force; Start-Service -Name 'MCP-Proxy'
```

### Key File Locations
- **Go embed source:** `web/frontend/dist/` (NOT `frontend/dist/`)
- **Embed directive:** `web/web.go` line 13: `//go:embed frontend/dist`
- **Makefile frontend-build:** Lines 70-79 (handles the copy correctly)

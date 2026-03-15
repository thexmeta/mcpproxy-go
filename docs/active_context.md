# Active Context - MCPProxy-Go Tool Management Feature

**Last Updated:** 2026-03-15  
**Current Focus:** Tool Management Feature Implementation

## Current State

### ✅ Completed This Session
1. **OAuth Bug Fix** - Fixed token persistence in `connection_oauth.go`
2. **Documentation** - Updated QWEN.md with Tool Management CLI commands
3. **Config Resolution** - Fixed BOM issue, configured GitHub Copilot MCP with OAuth

### 🎯 Next Task: Implement Tool Management

**Priority:** High  
**Estimated Effort:** 4-6 hours

**Implementation Plan:**

#### Phase 1: Backend API (Go)
1. Add tool preferences to database schema (`internal/storage/bbolt.go`)
   - New bucket: `tool_preferences`
   - Key: `{server_name}:{tool_name}`
   - Value: `{enabled: bool, custom_name: string, custom_description: string}`

2. Add management service methods (`internal/management/service.go`)
   - `GetToolPreferences(serverName string) (map[string]*ToolPreference, error)`
   - `UpdateToolPreference(serverName, toolName string, pref *ToolPreference) error`
   - `BulkUpdateToolPreferences(serverName string, updates map[string]*ToolPreference) error`

3. Add HTTP API endpoints (`internal/httpapi/server.go`)
   - `GET /api/v1/servers/{name}/tools` - List tools with preferences
   - `POST /api/v1/servers/{name}/tools/{tool}/enable` - Enable tool
   - `POST /api/v1/servers/{name}/tools/{tool}/disable` - Disable tool
   - `POST /api/v1/servers/{name}/tools/{tool}/rename` - Rename tool
   - `POST /api/v1/servers/{name}/tools/bulk` - Bulk operations

4. Add CLI commands (`internal/upstream/cli/tools.go` - new file)
   - `mcpproxy tools list <server>`
   - `mcpproxy tools enable <server> <tool>`
   - `mcpproxy tools disable <server> <tool>`
   - `mcpproxy tools rename <server> <tool> <new-name>`
   - `mcpproxy tools stats <server>`

#### Phase 2: Web UI (Vue 3)
1. Create Tools view component (`frontend/src/views/Tools.vue`)
2. Add tool management API client (`frontend/src/services/api.ts`)
3. Add route to router (`frontend/src/router/index.ts`)
4. Integrate with server details page

#### Phase 3: Testing
1. Unit tests for management service
2. E2E tests for API endpoints
3. Manual testing with GitHub Copilot MCP server

## Active State

### Running Processes
- **MCPProxy Tray:** Running (auto-restarts core)
- **Core Server:** Running on `127.0.0.1:3303`
- **Web UI:** Accessible at `http://127.0.0.1:3303`

### Database State
- **Path:** `C:\Users\eserk\.mcpproxy\config.db`
- **Status:** Active, no schema changes needed yet
- **OAuth Tokens:** GitHub Copilot MCP configured

### Config State
- **Path:** `C:\Users\eserk\.mcpproxy\mcp_config.json`
- **Format:** UTF-8 (no BOM)
- **GitHub Server:** Configured with OAuth static credentials

## Open Tasks

### High Priority
- [ ] Design database schema for tool preferences
- [ ] Implement backend API endpoints
- [ ] Add CLI commands for tool management

### Medium Priority
- [ ] Create Web UI for tool management
- [ ] Add tool usage statistics tracking
- [ ] Implement bulk operations

### Low Priority
- [ ] Add tool presets (save/load configurations)
- [ ] Add tool search/filter in Web UI
- [ ] Add tool call history and analytics

## Scratchpad

### API Design Notes
```typescript
// Tool Preference Structure
interface ToolPreference {
  enabled: boolean;
  customName?: string;
  customDescription?: string;
  originalName: string;
  lastUsed?: Date;
  callCount?: number;
}

// API Request/Response
interface ListToolsResponse {
  server: string;
  tools: Array<{
    name: string;
    description: string;
    preference?: ToolPreference;
    enabled: boolean;
    displayName: string;
  }>;
}

interface RenameToolRequest {
  newName: string;
  newDescription?: string;
}
```

### Implementation Considerations
1. **Tool Filtering:** Should happen at the MCP protocol layer to prevent AI from seeing disabled tools
2. **Name Persistence:** Store in database, not config (config is for server-level, DB is for tool-level)
3. **Backwards Compatibility:** Tools without preferences should use MCP server defaults
4. **Performance:** Cache tool preferences in memory, reload on changes

### Test Scenarios
- Disable a tool → AI shouldn't see it in tool list
- Rename a tool → AI sees new name, MCP server receives original name
- Bulk disable → Only specified tools affected
- Server restart → Preferences persist

## Next Session Resume Command

```bash
cd E:\Projects\Go\mcpproxy-go
# Start implementing tool management backend
# 1. Review this active_context.md
# 2. Create database schema for tool_preferences
# 3. Implement management service methods
# 4. Add HTTP API endpoints
# 5. Add CLI commands
```

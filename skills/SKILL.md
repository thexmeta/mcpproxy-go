# MCP Gateway Connection Skill

Connect AI agents to the MCP Proxy Gateway through its MCP server endpoint for intelligent tool discovery and unified access to hundreds of MCP servers.

## Invocation

Invoke this skill when users ask to:
- "Connect to MCP Proxy"
- "Use mcpproxy gateway"
- "Access MCP servers through proxy"
- "Configure MCP gateway connection"
- "Set up MCP tool discovery"

## Capabilities

### Connection Setup
- Configure MCP client connections (Claude Desktop, programmatic)
- Setup authentication (API key management)
- Handle multiple transport protocols (HTTP, SSE, Unix socket, named pipe)

### Tool Discovery
- Use `retrieve_tools` for BM25-powered tool search across all upstream servers
- Parse tool annotations (readOnlyHint, destructiveHint, etc.)
- Follow `call_with` recommendations for tool variant selection

### Tool Execution
- Execute read-only tools via `call_tool_read`
- Execute state-modifying tools via `call_tool_write`
- Execute destructive tools via `call_tool_destructive`
- Orchestrate multi-tool workflows via `code_execution`
- Handle paginated responses via `read_cache`

### Server Management
- List upstream servers
- Add/remove/update servers
- Manage security quarantine
- View server logs

## Usage Examples

### Basic Connection
```typescript
import { Client } from "@modelcontextprotocol/sdk/client/index.js";
import { StreamableHTTPClientTransport } from "@modelcontextprotocol/sdk/client/streamableHttp.js";

const client = new Client({ name: "my-agent", version: "1.0.0" });
const transport = new StreamableHTTPClientTransport(
  new URL("http://127.0.0.1:8080/mcp")
);
await client.connect(transport);
```

### Tool Discovery Pattern
```javascript
// ALWAYS start with retrieve_tools
const tools = await mcp.callTool({
  name: "retrieve_tools",
  arguments: { query: "create GitHub repository", limit: 10 }
});

// Use exact tool name from results
const tool = tools.tools.find(t => t.name.includes("create_repo"));

// Call with appropriate variant
await mcp.callTool({
  name: "call_tool_write", // Based on call_with recommendation
  arguments: {
    name: tool.name,
    args_json: JSON.stringify({ repo: "my-repo" }),
    intent_data_sensitivity: "public",
    intent_reason: "User wants to create a new repository"
  }
});
```

### Multi-Step Orchestration
```javascript
await mcp.callTool({
  name: "code_execution",
  arguments: {
    code: `
      const user = call_tool('github', 'get_user', {username: input.user});
      const repos = call_tool('github', 'list_repos', {user: input.user});
      return { user: user.result, repo_count: repos.result.length };
    `,
    input: { user: "octocat" },
    options: { timeout_ms: 60000, max_tool_calls: 5 }
  }
});
```

## Configuration

### MCP Endpoint
- **Primary:** `http://127.0.0.1:8080/mcp`
- **Unix Socket:** `unix:///tmp/mcpproxy.sock` (macOS/Linux)
- **Named Pipe:** `npipe:////./pipe/mcpproxy` (Windows)

### Authentication
- **Config File:** `~/.mcpproxy/mcp_config.json`
- **Header:** `X-API-Key: your-api-key`
- **Query:** `?apikey=your-api-key`

### Available Tools
1. `retrieve_tools` - Tool discovery (ALWAYS CALL FIRST)
2. `call_tool_read` - Read-only operations
3. `call_tool_write` - State-modifying operations
4. `call_tool_destructive` - Irreversible operations
5. `read_cache` - Paginated data retrieval
6. `code_execution` - JavaScript orchestration
7. `upstream_servers` - Server management
8. `quarantine_security` - Security management

## Security Features

### Quarantine Protection
- New servers auto-quarantined to prevent Tool Poisoning Attacks
- Review with: `quarantine_security list`
- Unquarantine with: `quarantine_security unquarantine --name <server>`

### Intent Declaration
- `intent_data_sensitivity`: public, internal, private, unknown
- `intent_reason`: Explain why the tool is being called

## Troubleshooting

### Connection Failed
```bash
# Check if gateway is running
curl http://127.0.0.1:8080/api/v1/status

# Check port usage
netstat -ano | findstr :8080  # Windows
lsof -i :8080  # macOS/Linux

# Check logs
tail -f ~/.mcpproxy/logs/main.log
```

### Tool Not Found
```bash
# Always use retrieve_tools first
mcp.callTool({ name: "retrieve_tools", arguments: { query: "..." } })

# Check server status
mcpproxy upstream list

# Check quarantine
mcpproxy quarantine_security list
```

## Related Skills
- `mcp-builder` - For creating new MCP servers
- `context7-efficient` - For library documentation lookup
- `browser-automation` - For web-based MCP interactions

## Resources
- **Web UI:** `http://127.0.0.1:8080/ui/`
- **Swagger Docs:** `http://127.0.0.1:8080/swagger/`
- **Activity Log:** `/api/v1/activity`
- **Full Documentation:** See `skills/mcp-gateway-connection.md`

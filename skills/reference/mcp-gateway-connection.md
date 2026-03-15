# MCP Proxy Gateway Connection Skill

## Overview

This skill enables AI agents to connect to and use the **MCP Proxy Gateway** through its MCP server endpoint. The gateway provides intelligent tool discovery, security quarantine, and unified access to hundreds of MCP servers through a single connection point.

## Connection Configuration

### MCP Server Endpoint

**Primary Endpoint (Streamable HTTP):**
```
http://127.0.0.1:8080/mcp
```

**Alternative Endpoints:**
- Legacy SSE: `http://127.0.0.1:8080/mcp` (with SSE transport headers)
- Unix Socket (macOS/Linux): `unix:///tmp/mcpproxy.sock`
- Named Pipe (Windows): `npipe:////./pipe/mcpproxy`

### Authentication

The gateway requires API key authentication for REST API access. The MCP endpoint itself is unprotected for client compatibility.

**API Key Location:** `~/.mcpproxy/mcp_config.json`

**Authentication Methods:**
```javascript
// Header-based (recommended for REST API)
headers: {
  "X-API-Key": "your-api-key"
}

// Query parameter (for URLs)
?apikey=your-api-key
```

## Connection Setup Examples

### Claude Desktop Configuration

Add to your `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "mcpproxy": {
      "url": "http://127.0.0.1:8080/mcp",
      "transport": {
        "type": "http"
      }
    }
  }
}
```

### Programmatic Connection (TypeScript/JavaScript)

```typescript
import { Client } from "@modelcontextprotocol/sdk/client/index.js";
import { StreamableHTTPClientTransport } from "@modelcontextprotocol/sdk/client/streamableHttp.js";

const client = new Client({
  name: "my-agent",
  version: "1.0.0",
});

const transport = new StreamableHTTPClientTransport(
  new URL("http://127.0.0.1:8080/mcp")
);

await client.connect(transport);

// List available tools
const tools = await client.listTools();
console.log(`Connected! ${tools.tools.length} tools available`);
```

### Programmatic Connection (Python)

```python
from mcp import ClientSession, StdioServerParameters
from mcp.client.stdio import stdio_client
from mcp.client.streamable_http import streamablehttp_client

async with streamablehttp_client("http://127.0.0.1:8080/mcp") as (read, write):
    async with ClientSession(read, write) as session:
        await session.initialize()
        
        # List available tools
        tools = await session.list_tools()
        print(f"Connected! {len(tools.tools)} tools available")
```

## Core Tools Available

The MCP Proxy Gateway provides these built-in tools:

### 1. `retrieve_tools` - Primary Tool Discovery

**Purpose:** Discover relevant tools across all upstream MCP servers using BM25 search.

**When to use:** ALWAYS call this FIRST before attempting to use any specific tool.

**Parameters:**
```json
{
  "query": "natural language description of what you want to accomplish",
  "limit": 10,
  "include_stats": false,
  "debug": false
}
```

**Example Usage:**
```javascript
// Find tools for GitHub operations
const result = await mcp.callTool({
  name: "retrieve_tools",
  arguments: {
    query: "create GitHub repository",
    limit: 10
  }
});

// Result includes:
// - tool name in format "server:tool"
// - annotations (readOnlyHint, destructiveHint, etc.)
// - call_with recommendation (call_tool_read/write/destructive)
```

### 2. `call_tool_read` - Read-Only Operations

**Purpose:** Execute read-only tools (search, query, list, get, fetch, etc.)

**When to use:** For tools with names containing: `search`, `query`, `list`, `get`, `fetch`, `find`, `check`, `view`, `read`, `show`, `describe`, `lookup`, `retrieve`, `browse`, `explore`, `discover`, `scan`, `inspect`, `analyze`, `examine`, `validate`, `verify`

**Parameters:**
```json
{
  "name": "server:tool",
  "args_json": "{\"param1\": \"value1\"}",
  "intent_data_sensitivity": "public|internal|private|unknown",
  "intent_reason": "Why this tool is being called"
}
```

**Example:**
```javascript
await mcp.callTool({
  name: "call_tool_read",
  arguments: {
    name: "github:get_user",
    args_json: "{\"username\": \"octocat\"}",
    intent_data_sensitivity: "public",
    intent_reason: "User requested to check their GitHub profile"
  }
});
```

### 3. `call_tool_write` - State-Modifying Operations

**Purpose:** Execute tools that modify state (create, update, modify, add, set, send, etc.)

**When to use:** For tools with names containing: `create`, `update`, `modify`, `add`, `set`, `send`, `edit`, `change`, `write`, `post`, `put`, `patch`, `insert`, `upload`, `submit`, `assign`, `configure`, `enable`, `register`, `subscribe`, `publish`, `move`, `copy`, `rename`, `merge`

**Parameters:**
```json
{
  "name": "server:tool",
  "args_json": "{\"param1\": \"value1\"}",
  "intent_data_sensitivity": "public|internal|private|unknown",
  "intent_reason": "Why this modification is needed"
}
```

**Example:**
```javascript
await mcp.callTool({
  name: "call_tool_write",
  arguments: {
    name: "github:create_issue",
    args_json: "{\"repo\": \"owner/repo\", \"title\": \"Bug fix\"}",
    intent_data_sensitivity: "public",
    intent_reason: "User requested to create a new issue for tracking"
  }
});
```

### 4. `call_tool_destructive` - Irreversible Operations

**Purpose:** Execute destructive tools (delete, remove, drop, revoke, disable, destroy, etc.)

**When to use:** For tools with names containing: `delete`, `remove`, `drop`, `revoke`, `disable`, `destroy`, `purge`, `reset`, `clear`, `unsubscribe`, `cancel`, `terminate`, `close`, `archive`, `ban`, `block`, `disconnect`, `kill`, `wipe`, `truncate`, `force`, `hard`

**Parameters:**
```json
{
  "name": "server:tool",
  "args_json": "{\"param1\": \"value1\"}",
  "intent_data_sensitivity": "public|internal|private|unknown",
  "intent_reason": "Justification for deletion"
}
```

**Example:**
```javascript
await mcp.callTool({
  name: "call_tool_destructive",
  arguments: {
    name: "github:delete_repo",
    args_json: "{\"repo\": \"owner/old-repo\"}",
    intent_data_sensitivity: "private",
    intent_reason: "User confirmed deletion of obsolete repository"
  }
});
```

### 5. `read_cache` - Paginated Data Retrieval

**Purpose:** Retrieve paginated data when responses are truncated.

**When to use:** When a tool response indicates truncation with a cache key.

**Parameters:**
```json
{
  "key": "cache-key-from-truncation-message",
  "offset": 0,
  "limit": 50
}
```

### 6. `code_execution` - JavaScript Orchestration

**Purpose:** Execute JavaScript code to orchestrate multiple MCP tools in a single request.

**When to use:** For multi-step workflows requiring conditional logic, loops, or data transformations.

**Parameters:**
```json
{
  "code": "JavaScript ES5.1+ code",
  "input": {},
  "options": {
    "timeout_ms": 120000,
    "max_tool_calls": 10,
    "allowed_servers": ["github", "notion"]
  }
}
```

**Example:**
```javascript
await mcp.callTool({
  name: "code_execution",
  arguments: {
    code: `
      const user = call_tool('github', 'get_user', {username: input.username});
      if (!user.ok) throw new Error(user.error.message);
      
      const repos = call_tool('github', 'list_repos', {user: input.username});
      
      return {
        user: user.result,
        repo_count: repos.result.length,
        timestamp: Date.now()
      };
    `,
    input: { username: "octocat" },
    options: {
      timeout_ms: 60000,
      max_tool_calls: 5
    }
  }
});
```

### 7. `upstream_servers` - Server Management

**Purpose:** Manage upstream MCP servers (add, remove, update, list).

**Operations:**
- `list` - List all configured servers
- `add` - Add a new server
- `remove` - Remove a server
- `update` - Update server configuration
- `patch` - Partial update with smart merge
- `tail_log` - View server logs

**Example - List Servers:**
```javascript
await mcp.callTool({
  name: "upstream_servers",
  arguments: {
    operation: "list"
  }
});
```

**Example - Add Server:**
```javascript
await mcp.callTool({
  name: "upstream_servers",
  arguments: {
    operation: "add",
    name: "weather-server",
    url: "https://api.weather.com/mcp",
    protocol: "http",
    headers_json: "{\"Authorization\": \"Bearer token\"}"
  }
});
```

### 8. `quarantine_security` - Security Management

**Purpose:** Review and manage quarantined servers (security feature).

**Operations:**
- `list` - List quarantined servers
- `unquarantine` - Remove server from quarantine
- `quarantine` - Add server to quarantine

## Usage Patterns

### Pattern 1: Tool Discovery Workflow

```javascript
// Step 1: ALWAYS start with retrieve_tools
const tools = await mcp.callTool({
  name: "retrieve_tools",
  arguments: {
    query: "search for files in the repository",
    limit: 10
  }
});

// Step 2: Parse results to find the right tool
const searchTool = tools.tools.find(t => 
  t.name.includes("search") && t.name.includes("file")
);

// Step 3: Use the appropriate call_tool variant
const result = await mcp.callTool({
  name: "call_tool_read", // Based on call_with recommendation
  arguments: {
    name: searchTool.name,
    args_json: JSON.stringify({ query: "README" }),
    intent_data_sensitivity: "public",
    intent_reason: "User wants to find README files"
  }
});
```

### Pattern 2: Multi-Step Orchestration

```javascript
// Use code_execution for complex workflows
await mcp.callTool({
  name: "code_execution",
  arguments: {
    code: `
      // Step 1: Get user info
      const user = call_tool('github', 'get_user', {username: input.user});
      if (!user.ok) return {error: user.error.message};
      
      // Step 2: List their repos
      const repos = call_tool('github', 'list_repos', {user: input.user});
      
      // Step 3: Filter and transform
      const publicRepos = repos.result.filter(r => !r.private);
      
      return {
        user: user.result.login,
        total_repos: repos.result.length,
        public_repos: publicRepos.length,
        top_repos: publicRepos.slice(0, 5).map(r => r.name)
      };
    `,
    input: { user: "octocat" },
    options: {
      timeout_ms: 60000,
      max_tool_calls: 10
    }
  }
});
```

### Pattern 3: Error Handling with Intent

```javascript
try {
  const result = await mcp.callTool({
    name: "call_tool_write",
    arguments: {
      name: "github:create_issue",
      args_json: JSON.stringify({
        repo: "owner/repo",
        title: "Bug Report"
      }),
      intent_data_sensitivity: "public",
      intent_reason: "Creating issue for bug tracking"
    }
  });
  
  if (result.isError) {
    // Handle error with context
    console.error("Failed to create issue:", result.content[0].text);
  }
} catch (error) {
  console.error("Tool call failed:", error.message);
}
```

## Security Features

### Quarantine Protection

Newly added servers are automatically quarantined to prevent Tool Poisoning Attacks (TPA).

**Check Quarantine Status:**
```javascript
await mcp.callTool({
  name: "quarantine_security",
  arguments: {
    operation: "list"
  }
});
```

**Unquarantine a Server:**
```javascript
await mcp.callTool({
  name: "quarantine_security",
  arguments: {
    operation: "unquarantine",
    name: "server-name"
  }
});
```

### Intent Declaration

All tool calls support intent declaration for audit and security purposes:

- `intent_data_sensitivity`: Classify data being accessed (public, internal, private, unknown)
- `intent_reason`: Explain why the tool is being called

## REST API Endpoints

For programmatic management, the gateway also provides REST API endpoints:

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/status` | GET | Server status and statistics |
| `/api/v1/servers` | GET | List all upstream servers |
| `/api/v1/servers/{name}/enable` | POST | Enable server |
| `/api/v1/servers/{name}/disable` | POST | Disable server |
| `/api/v1/servers/{name}/quarantine` | POST | Quarantine server |
| `/api/v1/tools` | GET | Search tools |
| `/api/v1/activity` | GET | List activity records |
| `/events` | GET | SSE stream for live updates |

**Example - REST API Call:**
```bash
curl -H "X-API-Key: your-api-key" \
  http://127.0.0.1:8080/api/v1/servers
```

## Troubleshooting

### Connection Issues

**Problem:** Cannot connect to MCP endpoint

**Solutions:**
1. Verify the gateway is running: `curl http://127.0.0.1:8080/api/v1/status`
2. Check if port 8080 is in use: `netstat -ano | findstr :8080` (Windows) or `lsof -i :8080` (macOS/Linux)
3. Verify firewall settings allow localhost connections
4. Check gateway logs: `~/.mcpproxy/logs/main.log`

### Tool Not Found

**Problem:** Tool call fails with "tool not found"

**Solutions:**
1. Always call `retrieve_tools` first to discover available tools
2. Use exact tool name from `retrieve_tools` results (format: `server:tool`)
3. Verify the upstream server is enabled and connected
4. Check if server is quarantined using `quarantine_security`

### Authentication Errors

**Problem:** REST API returns 401 Unauthorized

**Solutions:**
1. Get API key from config: `~/.mcpproxy/mcp_config.json`
2. Include API key in header: `X-API-Key: your-api-key`
3. Or use query parameter: `?apikey=your-api-key`

### Server Quarantine

**Problem:** Newly added server is quarantined

**Solutions:**
1. This is expected behavior for security
2. Review server using `quarantine_security list`
3. Unquarantine if trusted: `quarantine_security unquarantine --name <server>`

## Best Practices

1. **Always start with `retrieve_tools`** - Never guess tool names
2. **Use the correct tool variant** - Follow `call_with` recommendations
3. **Declare intent** - Provide `intent_data_sensitivity` and `intent_reason`
4. **Handle pagination** - Use `read_cache` when responses are truncated
5. **Orchestrate with code_execution** - For complex multi-step workflows
6. **Monitor activity** - Use `/api/v1/activity` for audit trails
7. **Review quarantine** - Regularly check quarantined servers

## Additional Resources

- **Configuration:** `~/.mcpproxy/mcp_config.json`
- **Logs:** `~/.mcpproxy/logs/main.log`
- **Database:** `~/.mcpproxy/config.db`
- **Documentation:** `/swagger/` endpoint (OpenAPI spec)
- **Web UI:** `http://127.0.0.1:8080/ui/`

## Support

For issues or questions:
1. Check logs: `mcpproxy upstream logs <server-name> --tail=100`
2. Run diagnostics: `mcpproxy doctor`
3. View activity: `mcpproxy activity list --type tool_call --status error`
4. Check server status: `mcpproxy upstream list`

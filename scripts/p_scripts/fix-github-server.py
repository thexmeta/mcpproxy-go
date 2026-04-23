import json

config_path = r"C:\Users\eserk\.mcpproxy\mcp_config.json"

with open(config_path, 'r', encoding='utf-8-sig') as f:
    config = json.load(f)

# Find and replace the Github server with stdio-based official GitHub MCP
for i, server in enumerate(config.get('mcpServers', [])):
    if server.get('name') == 'Github':
        # Replace with official GitHub MCP server using stdio
        config['mcpServers'][i] = {
            "name": "Github",
            "protocol": "stdio",
            "command": "npx",
            "args": ["-y", "@modelcontextprotocol/server-github"],
            "env": {
                "GITHUB_TOKEN": "${keyring:github_token}"
            },
            "enabled": True,
            "quarantined": False
        }
        print("Replaced Github server with official GitHub MCP (stdio mode)")
        break

# Save the updated config
with open(config_path, 'w', encoding='utf-8-sig') as f:
    json.dump(config, f, indent=2, ensure_ascii=False)

print("Config file updated successfully")
print("The official GitHub MCP server uses GITHUB_TOKEN directly - no OAuth needed!")

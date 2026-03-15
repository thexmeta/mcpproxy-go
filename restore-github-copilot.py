import json

config_path = r"C:\Users\eserk\.mcpproxy\mcp_config.json"

with open(config_path, 'r', encoding='utf-8-sig') as f:
    config = json.load(f)

# Find and replace the Github server back to Copilot MCP
for i, server in enumerate(config.get('mcpServers', [])):
    if server.get('name') == 'Github':
        # Restore original GitHub Copilot MCP server config
        config['mcpServers'][i] = {
            "name": "Github",
            "url": "https://api.githubcopilot.com/mcp/",
            "protocol": "streamable-http",
            "env": {
                "GITHUB_TOKEN": "${keyring:github_token}"
            },
            "oauth": {
                "client_id": "Ov23liuYmIajHaMzTrB7",
                "client_secret": "${keyring:github_oauth_client_secret}"
            },
            "enabled": True,
            "quarantined": False
        }
        print("Restored Github server to GitHub Copilot MCP (streamable-http with OAuth)")
        break

# Save the updated config
with open(config_path, 'w', encoding='utf-8') as f:
    json.dump(config, f, indent=2)

print("Config file restored successfully")

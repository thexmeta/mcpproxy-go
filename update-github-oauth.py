import json

config_path = r"C:\Users\eserk\.mcpproxy\mcp_config.json"

with open(config_path, 'r', encoding='utf-8-sig') as f:
    config = json.load(f)

# Find and update the Github server config
for server in config.get('mcpServers', []):
    if server.get('name') == 'Github':
        # Ensure oauth object exists
        if 'oauth' not in server:
            server['oauth'] = {}
        # Add client_secret reference
        server['oauth']['client_secret'] = '${keyring:github_oauth_client_secret}'
        print(f"Updated Github server:")
        print(f"  client_id: {server['oauth'].get('client_id')}")
        print(f"  client_secret: {server['oauth'].get('client_secret')}")
        break

# Save the updated config
with open(config_path, 'w', encoding='utf-8') as f:
    json.dump(config, f, indent=2, ensure_ascii=False)

print("Config file updated successfully")

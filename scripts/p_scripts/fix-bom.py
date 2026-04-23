import json

config_path = r"C:\Users\eserk\.mcpproxy\mcp_config.json"

with open(config_path, 'r', encoding='utf-8-sig') as f:
    config = json.load(f)

# Save without BOM
with open(config_path, 'w', encoding='utf-8') as f:
    json.dump(config, f, indent=2)

print("Config file saved without BOM")

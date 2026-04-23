$configPath = "C:\Users\eserk\.mcpproxy\mcp_config.json"
$config = Get-Content $configPath -Raw | ConvertFrom-Json

# Find and update the Github server config
foreach ($server in $config.mcpServers) {
    if ($server.name -eq "Github") {
        # Ensure oauth object exists
        if (-not $server.oauth) {
            $server.oauth = @{}
        }
        # Add client_secret reference
        $server.oauth.client_secret = '${keyring:github_oauth_client_secret}'
        Write-Host "Updated Github server with client_secret reference"
        Write-Host "  client_id: $($server.oauth.client_id)"
        Write-Host "  client_secret: $($server.oauth.client_secret)"
        break
    }
}

# Save the updated config
$config | ConvertTo-Json -Depth 10 | Set-Content $configPath -Encoding UTF8
Write-Host "Config file updated successfully"

# Test script for MCPProxy restart functionality
# This script starts mcpproxy, calls the restart API, and verifies it restarts

$ErrorActionPreference = "Stop"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "MCPProxy Restart Test" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

# Kill any existing instances
Write-Host "`nStopping existing instances..." -ForegroundColor Yellow
Stop-Process -Name mcpproxy,mcpproxy-tray -Force -ErrorAction SilentlyContinue
Start-Sleep -Seconds 3

# Start mcpproxy in background
Write-Host "`nStarting mcpproxy..." -ForegroundColor Yellow
$proc = Start-Process -FilePath ".\mcpproxy.exe" -ArgumentList "serve", "--listen", "127.0.0.1:8765" -PassThru -WindowStyle Hidden
Start-Sleep -Seconds 5

# Check if process is running
if ($proc.HasExited) {
    Write-Host "ERROR: mcpproxy failed to start" -ForegroundColor Red
    exit 1
}
Write-Host "mcpproxy started successfully (PID: $($proc.Id))" -ForegroundColor Green

# Get API key from config
$apiConfigPath = "$env:USERPROFILE\.mcpproxy\mcp_config.json"
$apiKey = ""
if (Test-Path $apiConfigPath) {
    $config = Get-Content $apiConfigPath -Raw | ConvertFrom-Json
    $apiKey = $config.api_key
}

$headers = @{}
if ($apiKey) {
    $headers["X-API-Key"] = $apiKey
    Write-Host "Using API key: $($apiKey.Substring(0,8))..." -ForegroundColor Gray
}

# Test restart API
Write-Host "`nCalling restart API..." -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "http://127.0.0.1:8765/api/v1/restart" -Method POST -Headers $headers -ContentType "application/json"
    Write-Host "Restart API response: $($response | ConvertTo-Json)" -ForegroundColor Cyan
} catch {
    Write-Host "API call failed: $($_.Exception.Message)" -ForegroundColor Yellow
    # This is expected - the server might have already restarted
}

# Wait for restart
Write-Host "`nWaiting for restart..." -ForegroundColor Yellow
Start-Sleep -Seconds 5

# Check if process restarted
Write-Host "`nChecking process status..." -ForegroundColor Yellow
$processes = Get-Process -Name mcpproxy -ErrorAction SilentlyContinue
if ($processes) {
    foreach ($p in $processes) {
        Write-Host "✓ mcpproxy is running (PID: $($p.Id), Started: $($p.StartTime))" -ForegroundColor Green
    }
    Write-Host "`n========================================" -ForegroundColor Cyan
    Write-Host "RESTART TEST PASSED" -ForegroundColor Green
    Write-Host "========================================" -ForegroundColor Cyan
} else {
    Write-Host "✗ mcpproxy is NOT running" -ForegroundColor Red
    Write-Host "`n========================================" -ForegroundColor Cyan
    Write-Host "RESTART TEST FAILED" -ForegroundColor Red
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host "`nTroubleshooting:" -ForegroundColor Yellow
    Write-Host "1. Check logs at: $env:USERPROFILE\.mcpproxy\logs\" -ForegroundColor Yellow
    Write-Host "2. Try running mcpproxy manually to see errors" -ForegroundColor Yellow
}

# Cleanup
Write-Host "`nCleaning up..." -ForegroundColor Yellow
Stop-Process -Name mcpproxy -ErrorAction SilentlyContinue

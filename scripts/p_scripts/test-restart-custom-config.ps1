# Test restart with custom config path
$ErrorActionPreference = "Stop"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "MCPProxy Restart Test (Custom Config)" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

$ConfigPath = "C:\Users\eserk\.mcpproxy\mcp_config.json"
$LogDir = "D:\Development\bin\logs"
$ListenAddr = "127.0.0.1:8766"

# Kill any existing instances
Write-Host "`nStopping existing instances..." -ForegroundColor Yellow
Stop-Process -Name mcpproxy,mcpproxy-tray -Force -ErrorAction SilentlyContinue
Start-Sleep -Seconds 3

# Start mcpproxy with custom config
Write-Host "`nStarting mcpproxy with custom config..." -ForegroundColor Yellow
Write-Host "  Config: $ConfigPath" -ForegroundColor Gray
Write-Host "  Log Dir: $LogDir" -ForegroundColor Gray
Write-Host "  Listen: $ListenAddr" -ForegroundColor Gray

$proc = Start-Process -FilePath ".\mcpproxy.exe" `
    -ArgumentList "serve", "--config", $ConfigPath, "--listen", $ListenAddr, "--log-level=debug" `
    -PassThru -WindowStyle Hidden

Start-Sleep -Seconds 5

# Check if process is running
if ($proc.HasExited) {
    Write-Host "ERROR: mcpproxy failed to start" -ForegroundColor Red
    exit 1
}
Write-Host "mcpproxy started successfully (PID: $($proc.Id))" -ForegroundColor Green

# Get API key from config
$apiConfigPath = $ConfigPath
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
    $response = Invoke-RestMethod -Uri "http://$ListenAddr/api/v1/restart" -Method POST -Headers $headers -ContentType "application/json"
    Write-Host "Restart API response: $($response | ConvertTo-Json)" -ForegroundColor Cyan
} catch {
    Write-Host "API call failed: $($_.Exception.Message)" -ForegroundColor Yellow
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
    
    # Check if it's a NEW process (different PID or newer start time)
    $newProcess = $processes | Where-Object { $_.Id -ne $proc.Id }
    if ($newProcess) {
        Write-Host "`n========================================" -ForegroundColor Cyan
        Write-Host "RESTART TEST PASSED ✅" -ForegroundColor Green
        Write-Host "========================================" -ForegroundColor Cyan
        Write-Host "Original PID: $($proc.Id)" -ForegroundColor Yellow
        Write-Host "New PID: $($newProcess.Id)" -ForegroundColor Green
    } else {
        Write-Host "`n========================================" -ForegroundColor Cyan
        Write-Host "RESTART TEST INCONCLUSIVE" -ForegroundColor Yellow
        Write-Host "========================================" -ForegroundColor Cyan
        Write-Host "Same process still running (PID: $($proc.Id))" -ForegroundColor Yellow
        Write-Host "Check logs to see if restart occurred" -ForegroundColor Yellow
    }
} else {
    Write-Host "✗ mcpproxy is NOT running" -ForegroundColor Red
    Write-Host "`n========================================" -ForegroundColor Cyan
    Write-Host "RESTART TEST FAILED ❌" -ForegroundColor Red
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host "`nTroubleshooting:" -ForegroundColor Yellow
    Write-Host "1. Check logs at: $LogDir" -ForegroundColor Yellow
    Write-Host "2. Look for REQUESTRESTART entries" -ForegroundColor Yellow
}

# Cleanup
Write-Host "`nCleaning up..." -ForegroundColor Yellow
Stop-Process -Name mcpproxy -ErrorAction SilentlyContinue

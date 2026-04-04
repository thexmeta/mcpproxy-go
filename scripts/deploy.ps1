# MCPProxy Deploy Script
# Copies built binaries to target folder and restarts the service

param(
    [string]$Version = "v0.23.15",
    [string]$TargetPath = "D:\Development\CodeMode\mcpproxy-go"
)

$ErrorActionPreference = "Stop"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "MCPProxy Deploy" -ForegroundColor Cyan
Write-Host "Version: $Version" -ForegroundColor Cyan
Write-Host "Target: $TargetPath" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

# Extract the release
$VersionNoV = $Version -replace '^v', ''
$ZipPath = "releases\mcpproxy-${VersionNoV}-windows-amd64.zip"
$ExtractPath = "releases\${Version}-deploy-temp"

if (!(Test-Path $ZipPath)) {
    Write-Host "Error: Release zip not found at $ZipPath" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "Extracting release..." -ForegroundColor Yellow
if (Test-Path $ExtractPath) {
    Remove-Item -Recurse -Force $ExtractPath
}
Expand-Archive -Path $ZipPath -DestinationPath $ExtractPath -Force

# Verify target directory exists
if (!(Test-Path $TargetPath)) {
    Write-Host "Creating target directory: $TargetPath" -ForegroundColor Yellow
    New-Item -ItemType Directory -Path $TargetPath -Force | Out-Null
}

# Stop MCP-Proxy service before copying
Write-Host ""
Write-Host "Stopping MCP-Proxy service..." -ForegroundColor Yellow
$ServiceStopped = $false
try {
    $Service = Get-Service -Name "MCP-Proxy" -ErrorAction SilentlyContinue
    if ($Service -and $Service.Status -eq "Running") {
        Stop-Service -Name "MCP-Proxy" -Force -WarningAction SilentlyContinue
        Start-Sleep -Seconds 3
        $CheckService = Get-Service -Name "MCP-Proxy" -ErrorAction SilentlyContinue
        if ($CheckService.Status -eq "Stopped") {
            Write-Host "  MCP-Proxy service stopped" -ForegroundColor Green
            $ServiceStopped = $true
        } else {
            Write-Host "  Warning: Service not fully stopped, waiting..." -ForegroundColor Yellow
            Start-Sleep -Seconds 2
            $ServiceStopped = $true
        }
    } else {
        Write-Host "  MCP-Proxy service not running or not found, proceeding with copy" -ForegroundColor Yellow
        $ServiceStopped = $true
    }
}
catch {
    Write-Host "  Warning: Failed to stop service: $_" -ForegroundColor Yellow
    Write-Host "  Attempting copy anyway..." -ForegroundColor Yellow
    $ServiceStopped = $true
}

# Copy binaries
Write-Host ""
Write-Host "Copying binaries to target..." -ForegroundColor Yellow

$FilesToCopy = @(
    "mcpproxy.exe",
    "mcpproxy-tray.exe"
)

$CopySuccess = 0
$CopyFailed = 0

foreach ($File in $FilesToCopy) {
    $Source = "$ExtractPath\$File"
    $Dest = "$TargetPath\$File"

    if (Test-Path $Source) {
        try {
            Copy-Item -Path $Source -Destination $Dest -Force
            Write-Host "  Copied: $File" -ForegroundColor Green
            $CopySuccess++
        }
        catch {
            Write-Host "  Error: Failed to copy $File - $_" -ForegroundColor Red
            $CopyFailed++
        }
    } else {
        Write-Host "  Warning: $File not found in release" -ForegroundColor Yellow
    }
}

if ($CopyFailed -gt 0) {
    Write-Host ""
    Write-Host "Warning: $CopyFailed file(s) failed to copy" -ForegroundColor Red
}

# Restart MCP-Proxy service
if ($ServiceStopped) {
    Write-Host ""
    Write-Host "Starting MCP-Proxy service..." -ForegroundColor Yellow
    try {
        Start-Service -Name "MCP-Proxy" -WarningAction SilentlyContinue
        Start-Sleep -Seconds 3
        $ServiceStatus = Get-Service -Name "MCP-Proxy" -ErrorAction SilentlyContinue
        if ($ServiceStatus) {
            Write-Host "  Service status: $($ServiceStatus.Status)" -ForegroundColor Green
        } else {
            Write-Host "  MCP-Proxy service not found" -ForegroundColor Yellow
        }
    }
    catch {
        Write-Host "  Warning: Failed to start service: $_" -ForegroundColor Yellow
    }
}

# Cleanup
Write-Host ""
Write-Host "Cleaning up..." -ForegroundColor Yellow
Remove-Item -Recurse -Force $ExtractPath

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Deploy Complete!" -ForegroundColor Green
Write-Host "  Copied: $CopySuccess | Failed: $CopyFailed" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan

Write-Host ""
Write-Host "Deployed files:" -ForegroundColor Yellow
Get-ChildItem $TargetPath -Filter "*.exe" | ForEach-Object {
    Write-Host "  $($_.Name) ($([math]::Round($_.Length/1MB, 2)) MB)"
}

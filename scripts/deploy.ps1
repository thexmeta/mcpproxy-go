# MCPProxy Deploy Script
# Copies built binaries to target folder

param(
    [string]$Version = "v0.21.5",
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

# Copy binaries
Write-Host ""
Write-Host "Copying binaries to target..." -ForegroundColor Yellow

$FilesToCopy = @(
    "mcpproxy.exe",
    "mcpproxy-tray.exe"
)

foreach ($File in $FilesToCopy) {
    $Source = "$ExtractPath\$File"
    $Dest = "$TargetPath\$File"
    
    if (Test-Path $Source) {
        Copy-Item -Path $Source -Destination $Dest -Force
        Write-Host "  Copied: $File" -ForegroundColor Green
    } else {
        Write-Host "  Warning: $File not found in release" -ForegroundColor Orange
    }
}

# Cleanup
Write-Host ""
Write-Host "Cleaning up..." -ForegroundColor Yellow
Remove-Item -Recurse -Force $ExtractPath

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Deploy Complete!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan

Write-Host ""
Write-Host "Deployed files:" -ForegroundColor Yellow
Get-ChildItem $TargetPath -Filter "*.exe" | ForEach-Object {
    Write-Host "  $($_.Name) ($([math]::Round($_.Length/1MB, 2)) MB)"
}

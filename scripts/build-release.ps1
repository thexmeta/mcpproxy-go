# MCPProxy Windows Release Build Script
# Usage: .\scripts\build-release.ps1 -Version "v0.21.3"

param(
    [string]$Version = "v0.21.3",
    [switch]$SkipFrontend
)

$ErrorActionPreference = "Stop"

$Commit = git rev-parse --short HEAD
$BuildDate = Get-Date -Format "yyyy-MM-ddTHH:mm:ssZ"
$LdFlags = "-s -w -X main.version=${Version} -X main.commit=${Commit} -X main.date=${BuildDate} -X github.com/smart-mcp-proxy/mcpproxy-go/internal/httpapi.buildVersion=${Version}"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "MCPProxy Windows Release Build" -ForegroundColor Cyan
Write-Host "Version: $Version" -ForegroundColor Cyan
Write-Host "Commit: $Commit" -ForegroundColor Cyan
Write-Host "Date: $BuildDate" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

# Create releases directory
$ReleasesDir = "releases\$Version"
if (!(Test-Path $ReleasesDir)) {
    New-Item -ItemType Directory -Path $ReleasesDir | Out-Null
}

# Build frontend if not skipped
if (!$SkipFrontend) {
    Write-Host ""
    Write-Host "Building frontend..." -ForegroundColor Yellow
    Set-Location frontend
    npm install
    npm run build
    Set-Location ..
    
    # Copy to embed location
    if (Test-Path "web\frontend") {
        Remove-Item -Recurse -Force "web\frontend"
    }
    New-Item -ItemType Directory -Path "web\frontend\dist" | Out-Null
    Copy-Item -Path "frontend\dist\*" -Destination "web\frontend\dist" -Recurse
}

# Generate OpenAPI spec
Write-Host ""
Write-Host "Generating OpenAPI specification..." -ForegroundColor Yellow
$swagPath = (Get-Command swag -ErrorAction SilentlyContinue).Source
if ($swagPath) {
    & $swagPath init -g cmd/mcpproxy/main.go --output oas --outputTypes go,yaml --v3.1 --exclude specs
} else {
    Write-Host "Warning: swag not found, skipping OpenAPI generation" -ForegroundColor Orange
}

# Build architectures
$Architectures = @("amd64", "arm64")

foreach ($Arch in $Architectures) {
    Write-Host ""
    Write-Host "Building for Windows/${Arch}..." -ForegroundColor Yellow
    
    $Env:GOOS = "windows"
    $Env:GOARCH = $Arch
    $Env:CGO_ENABLED = "1"
    
    # Build core binary
    Write-Host "  Building mcpproxy.exe..." -ForegroundColor Gray
    go build -ldflags $LdFlags -o "$ReleasesDir\mcpproxy.exe" ./cmd/mcpproxy
    
    # Build tray binary
    Write-Host "  Building mcpproxy-tray.exe..." -ForegroundColor Gray
    go build -ldflags $LdFlags -o "$ReleasesDir\mcpproxy-tray.exe" ./cmd/mcpproxy-tray
    
    # Create ZIP archive
    $ArchiveName = "mcpproxy-${Version.TrimStart('v')}-windows-${Arch}"
    Write-Host "  Creating ${ArchiveName}.zip..." -ForegroundColor Gray
    Compress-Archive -Path "$ReleasesDir\*.exe" -DestinationPath "releases\${ArchiveName}.zip" -Force
    
    # Clean up individual binaries in releases dir
    Remove-Item "$ReleasesDir\*.exe" -Force
    
    Write-Host "  ✓ Completed Windows/${Arch}" -ForegroundColor Green
}

# Clear environment
Remove-Item Env:\GOOS -ErrorAction SilentlyContinue
Remove-Item Env:\GOARCH -ErrorAction SilentlyContinue
Remove-Item Env:\CGO_ENABLED -ErrorAction SilentlyContinue

# Summary
Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Build Summary" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

Get-ChildItem "releases\$Version\*.zip" | Format-Table Name, Length -AutoSize
Get-ChildItem "releases\*.zip" | Where-Object { $_.Name -like "*${Version}*" } | Format-Table Name, Length -AutoSize

Write-Host ""
Write-Host "Release artifacts created in: releases\$Version" -ForegroundColor Green
Write-Host ""
Write-Host "To verify binaries:" -ForegroundColor Yellow
Write-Host "  .\releases\$Version\mcpproxy.exe --version"
Write-Host ""
Write-Host "To create GitHub release:" -ForegroundColor Yellow
Write-Host "  git push origin $Version"
Write-Host "  gh release create $Version --notes-file releases\RELEASE_NOTES_${Version.TrimStart('v')}.md releases\*.zip"

# MCPProxy Deploy Script
# Builds frontend, copies to web/embed dir, builds Go binary, copies to target, restarts service

param(
    [string]$Version = "v0.23.15",
    [string]$TargetPath = "D:\Development\CodeMode\mcpproxy-go"
)

$ErrorActionPreference = "Stop"

$RootDir = Split-Path -Parent $PSScriptRoot
$FrontendDir = Join-Path $RootDir "frontend"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "MCPProxy Deploy" -ForegroundColor Cyan
Write-Host "Version: $Version" -ForegroundColor Cyan
Write-Host "Target: $TargetPath" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

# Step 1: Build frontend
Write-Host ""
Write-Host "Step 1/5: Building frontend..." -ForegroundColor Cyan
Push-Location $FrontendDir
try {
    npm install
    if ($LASTEXITCODE -ne 0) {
        throw "Failed to install frontend dependencies"
    }
    npm run build
    if ($LASTEXITCODE -ne 0) {
        throw "Failed to build frontend"
    }
}
finally {
    Pop-Location
}
Write-Host "  Frontend built successfully" -ForegroundColor Green

# Step 2: Copy to web/frontend/dist (for Go embed)
Write-Host ""
Write-Host "Step 2/5: Copying frontend to web/frontend/dist..." -ForegroundColor Cyan
$WebFrontendDir = Join-Path $RootDir "web\frontend"
if (Test-Path $WebFrontendDir) {
    Remove-Item -Recurse -Force $WebFrontendDir
}
New-Item -ItemType Directory -Path $WebFrontendDir -Force | Out-Null
Copy-Item -Recurse (Join-Path $FrontendDir "dist") $WebFrontendDir
Write-Host "  Copied to web/frontend/dist" -ForegroundColor Green

# Step 3: Build Go binary
Write-Host ""
Write-Host "Step 3/5: Building Go binary..." -ForegroundColor Cyan
Push-Location $RootDir
try {
    $LDFLAGS = "-X main.version=$Version -X main.commit=$(git rev-parse --short HEAD) -X main.date=$(Get-Date -Format 'yyyy-MM-ddTHH:mm:ssZ') -s -w"
    go build -ldflags $LDFLAGS -o mcpproxy.exe ./cmd/mcpproxy
    if ($LASTEXITCODE -ne 0) {
        throw "Failed to build Go binary"
    }
    $BinaryPath = Join-Path $RootDir "mcpproxy.exe"
    Write-Host "  Go binary built: $BinaryPath" -ForegroundColor Green
}
finally {
    Pop-Location
}

# Verify target directory exists
if (!(Test-Path $TargetPath)) {
    Write-Host "Creating target directory: $TargetPath" -ForegroundColor Yellow
    New-Item -ItemType Directory -Path $TargetPath -Force | Out-Null
}

# Stop MCP-Proxy service before copying
Write-Host ""
Write-Host "Step 4/5: Stopping MCP-Proxy service..." -ForegroundColor Yellow
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

# Copy freshly built binary
Write-Host ""
Write-Host "Step 5/5: Copying binaries to target..." -ForegroundColor Yellow

$RootBinary = Join-Path $RootDir "mcpproxy.exe"
$DestBinary = Join-Path $TargetPath "mcpproxy.exe"

$CopySuccess = 0
$CopyFailed = 0

if (Test-Path $RootBinary) {
    try {
        Copy-Item -Path $RootBinary -Destination $DestBinary -Force
        Write-Host "  Copied: mcpproxy.exe" -ForegroundColor Green
        $CopySuccess++
    }
    catch {
        Write-Host "  Error: Failed to copy mcpproxy.exe - $_" -ForegroundColor Red
        $CopyFailed++
    }
} else {
    Write-Host "  Error: mcpproxy.exe not found at $RootBinary" -ForegroundColor Red
    $CopyFailed++
}

# Also copy tray if it exists
$RootTray = Join-Path $RootDir "mcpproxy-tray.exe"
$DestTray = Join-Path $TargetPath "mcpproxy-tray.exe"
if (Test-Path $RootTray) {
    try {
        Copy-Item -Path $RootTray -Destination $DestTray -Force
        Write-Host "  Copied: mcpproxy-tray.exe" -ForegroundColor Green
        $CopySuccess++
    }
    catch {
        Write-Host "  Error: Failed to copy mcpproxy-tray.exe - $_" -ForegroundColor Red
        $CopyFailed++
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
        Start-Sleep -Seconds 5
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

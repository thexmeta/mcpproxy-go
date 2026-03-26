@echo off
REM ============================================================================
REM MCPProxy Windows x64 Release Build Script
REM ============================================================================
REM Usage: scripts\build-release.bat [version]
REM Example: scripts\build-release.bat v0.21.4
REM
REM This script builds the MCPProxy release for Windows x64 (amd64) including:
REM - Frontend web assets (embedded in backend)
REM - mcpproxy.exe (core daemon)
REM - mcpproxy-tray.exe (system tray UI)
REM - ZIP archive for distribution
REM ============================================================================

setlocal EnableDelayedExpansion

REM Configuration
set "VERSION=%~1"
if "%VERSION%"=="" set "VERSION=v0.21.4"

REM Get commit hash
for /f "delims=" %%i in ('git rev-parse --short HEAD') do set "COMMIT=%%i"

REM Get current date in ISO format
for /f "delims=" %%i in ('powershell -NoProfile -Command "Get-Date -Format 'yyyy-MM-ddTHH:mm:ssZ'"') do set "BUILD_DATE=%%i"

REM Build flags
set "LDFLAGS=-s -w -X main.version=%VERSION% -X main.commit=%COMMIT% -X main.date=%BUILD_DATE% -X github.com/smart-mcp-proxy/mcpproxy-go/internal/httpapi.buildVersion=%VERSION%"

REM Architecture
set "GOOS=windows"
set "GOARCH=amd64"
set "CGO_ENABLED=1"

REM Directories
set "SCRIPT_DIR=%~dp0"
set "PROJECT_DIR=%SCRIPT_DIR%.."
set "RELEASES_DIR=%PROJECT_DIR%\releases\%VERSION%"

REM ============================================================================
REM Header
REM ============================================================================
echo ========================================
echo MCPProxy Windows x64 Release Build
echo ========================================
echo Version: %VERSION%
echo Commit: %COMMIT%
echo Date: %BUILD_DATE%
echo Output: %RELEASES_DIR%
echo ========================================
echo.

REM ============================================================================
REM Change to project directory
REM ============================================================================
cd /d "%PROJECT_DIR%"

REM ============================================================================
REM Create releases directory
REM ============================================================================
echo Creating releases directory...
if not exist "%RELEASES_DIR%" (
    mkdir "%RELEASES_DIR%"
    echo Created: %RELEASES_DIR%
)

REM ============================================================================
REM Build Frontend
REM ============================================================================
echo.
echo Building frontend...
cd /d "%PROJECT_DIR%\frontend"

call npm install
if errorlevel 1 (
    echo ERROR: npm install failed
    cd /d "%PROJECT_DIR%"
    exit /b 1
)

call npm run build
if errorlevel 1 (
    echo ERROR: npm build failed
    cd /d "%PROJECT_DIR%"
    exit /b 1
)

cd /d "%PROJECT_DIR%"

REM Copy frontend build to embed location
echo Copying frontend assets to embed location...
if exist "%PROJECT_DIR%\web\frontend" (
    rmdir /s /q "%PROJECT_DIR%\web\frontend"
)
mkdir "%PROJECT_DIR%\web\frontend\dist"
xcopy /E /I /Y "%PROJECT_DIR%\frontend\dist\*" "%PROJECT_DIR%\web\frontend\dist\"

echo Frontend build complete.

REM ============================================================================
REM Generate OpenAPI Specification
REM ============================================================================
echo.
echo Generating OpenAPI specification...
where swag >nul 2>&1
if %errorlevel% equ 0 (
    swag init -g cmd/mcpproxy/main.go --output oas --outputTypes go,yaml --v3.1 --exclude specs
    echo OpenAPI spec generated.
) else (
    echo WARNING: swag not found in PATH, skipping OpenAPI generation
    echo Install with: go install github.com/swaggo/swag/cmd/swag@latest
)

REM ============================================================================
REM Build Backend Binaries
REM ============================================================================
echo.
echo Building for Windows %GOARCH%...
echo GOOS=%GOOS%
echo GOARCH=%GOARCH%
echo CGO_ENABLED=%CGO_ENABLED%
echo.

set "GOOS=%GOOS%"
set "GOARCH=%GOARCH%"
set "CGO_ENABLED=%CGO_ENABLED%"

REM Build core binary
echo Building mcpproxy.exe...
go build -ldflags "%LDFLAGS%" -o "%RELEASES_DIR%\mcpproxy.exe" ./cmd/mcpproxy
if errorlevel 1 (
    echo ERROR: Failed to build mcpproxy.exe
    set "GOOS="
    set "GOARCH="
    set "CGO_ENABLED="
    exit /b 1
)
echo   Created: %RELEASES_DIR%\mcpproxy.exe

REM Build tray binary
echo Building mcpproxy-tray.exe...
go build -ldflags "%LDFLAGS%" -o "%RELEASES_DIR%\mcpproxy-tray.exe" ./cmd/mcpproxy-tray
if errorlevel 1 (
    echo ERROR: Failed to build mcpproxy-tray.exe
    set "GOOS="
    set "GOARCH="
    set "CGO_ENABLED="
    exit /b 1
)
echo   Created: %RELEASES_DIR%\mcpproxy-tray.exe

REM Clear environment variables
set "GOOS="
set "GOARCH="
set "CGO_ENABLED="

REM ============================================================================
REM Create ZIP Archive
REM ============================================================================
echo.
echo Creating ZIP archive...

REM Remove 'v' prefix from version for archive name
set "VERSION_NO_V=%VERSION:v=%"
set "ARCHIVE_NAME=mcpproxy-%VERSION_NO_V%-windows-amd64"

REM Create ZIP using PowerShell (more reliable than built-in zip)
powershell -NoProfile -Command "Compress-Archive -Path '%RELEASES_DIR%\*.exe' -DestinationPath '%PROJECT_DIR%\releases\%ARCHIVE_NAME%.zip' -Force"

if exist "%PROJECT_DIR%\releases\%ARCHIVE_NAME%.zip" (
    echo   Created: releases\%ARCHIVE_NAME%.zip
) else (
    echo ERROR: Failed to create ZIP archive
)

REM Clean up individual binaries in releases dir
del /q "%RELEASES_DIR%\*.exe"

REM ============================================================================
REM Summary
REM ============================================================================
echo.
echo ========================================
echo Build Summary
echo ========================================
echo.
echo Release artifacts:
dir /b "%PROJECT_DIR%\releases\*.zip" | findstr /c:"%VERSION_NO_V%"
echo.
echo Location: %RELEASES_DIR%
echo.
echo To verify the build:
echo   %RELEASES_DIR%\mcpproxy.exe --version
echo   %RELEASES_DIR%\mcpproxy-tray.exe --version
echo.
echo To create GitHub release:
echo   git tag %VERSION%
echo   git push origin %VERSION%
echo.
echo ========================================
echo Build completed successfully!
echo ========================================

endlocal
exit /b 0

# Lessons Learned - MCPProxy-Go

**Last Updated:** March 15, 2026

---

## OAuth Token Persistence Bug Fix (v0.21.4)

### Problem: OAuth Tokens Not Being Saved to Database

**Symptom:** OAuth authorization completes successfully in browser ("Authorization Successful"), but server shows "Authentication required" and "no valid token available" errors.

**Root Cause:** The OAuth handler was being obtained from the error object using `client.GetOAuthHandler(authErr)` instead of from the configured client using `c.GetOAuthHandler()`. The error-based handler doesn't have access to the `TokenStore` that was configured in the OAuth config, so tokens couldn't be persisted.

**Why This Happened:**
1. The code pattern `client.GetOAuthHandler(authErr)` extracts the handler from the OAuth authorization error
2. This handler was created during the initial failed connection attempt, before the OAuth config with TokenStore was fully set up
3. When `ProcessAuthorizationResponse` tried to save the token, the TokenStore was nil or not properly configured

**Fix Applied:**
```go
// OLD (broken):
oauthHandler := client.GetOAuthHandler(authErr)
if oauthHandler == nil {
    return fmt.Errorf("failed to get OAuth handler from error")
}

// NEW (fixed):
oauthHandler := c.GetOAuthHandler()  // Get from configured client
if oauthHandler == nil {
    // Fallback to error-based handler for backward compatibility
    oauthHandler = client.GetOAuthHandler(authErr)
    if oauthHandler == nil {
        return fmt.Errorf("failed to get OAuth handler: client not configured with OAuth")
    }
    c.logger.Warn("⚠️ Using OAuth handler from error (fallback mode)")
} else {
    c.logger.Info("✅ OAuth handler obtained from configured client (token persistence enabled)")
}
```

**Locations Fixed:** `internal/upstream/core/connection_oauth.go`
- Line ~964: `handleOAuthAuthorization()` method
- Line ~1357: `handleOAuthAuthorizationWithResult()` method  
- Line ~1878: `getAuthorizationURLQuick()` method

**Testing:**
- GitHub Copilot MCP server OAuth flow now persists tokens correctly
- Token survives daemon restarts
- Subsequent connections use persisted token without re-authentication

**Lesson:** Always get handlers from the configured client instance, not from error objects. Error-based extraction should only be a fallback for backward compatibility.

---

## Secret Resolution Patterns (v0.21.3)

### Fallback Resolver Pattern for Environment Variables

**Pattern:** When resolving `${env:NAME}` references, first check process environment variables, then fall back to keyring lookup.

**Why:** Users set secrets via UI (`/api/v1/secrets`) which stores in keyring, but configurations reference `${env:NAME}`. The fallback allows both patterns to work without breaking existing behavior.

**Implementation:**
```go
// In EnvProvider.Resolve()
func (p *EnvProvider) Resolve(ctx context.Context, ref Ref) (string, error) {
    // 1. Try actual environment variables first
    value := os.Getenv(ref.Name)
    if value != "" {
        return value, nil
    }
    
    // 2. Fall back to keyring if fallback resolver configured
    if p.fallbackResolver != nil {
        keyringRef := Ref{Type: "keyring", Name: ref.Name}
        keyringProvider := p.fallbackResolver.providers["keyring"]
        if keyringProvider != nil && keyringProvider.IsAvailable() {
            keyringValue, err := keyringProvider.Resolve(ctx, keyringRef)
            if err == nil && keyringValue != "" {
                return keyringValue, nil
            }
        }
    }
    
    return "", fmt.Errorf("environment variable %s not found", ref.Name)
}
```

**Benefits:**
- Actual env vars take precedence (for deployment automation)
- Keyring-stored values work for UI-managed secrets
- No breaking changes to existing configs
- Enables flexible secret management workflows

---

## OAuth Configuration Validation (v0.21.3)

### Fail Fast with Clear Error Messages

**Pattern:** Validate OAuth credentials (`client_id`, `client_secret`) at the point of use, not just at configuration load time.

**Why:** Silent failures during OAuth flow confuse users. Explicit validation with clear error messages enables administrators to fix configuration before users encounter broken login.

**Implementation:**
```go
// In OAuthHandler.HandleLogin()
if h.config.OAuth.ClientID == "" {
    h.logger.Errorw("OAuth login attempted but client_id not configured",
        "provider", h.config.OAuth.Provider)
    http.Error(w, "OAuth client_id not configured - server administrator must set teams.oauth.client_id", 
        http.StatusInternalServerError)
    return
}

if h.config.OAuth.ClientSecret == "" {
    h.logger.Errorw("OAuth login attempted but client_secret not configured",
        "provider", h.config.OAuth.Provider)
    http.Error(w, "OAuth client_secret not configured", http.StatusInternalServerError)
    return
}
```

**Frontend Error Handling:**
```typescript
// Check OAuth config before redirecting
async checkOAuthConfig(): Promise<string | null> {
  const response = await fetch(`${API_BASE}/auth/login`, {
    method: 'GET',
    redirect: 'manual',
  })
  
  if (response.status === 500) {
    const text = await response.text()
    if (text.includes('client_id')) {
      return text // Show to user
    }
  }
  return null
}
```

**Benefits:**
- Immediate feedback on misconfiguration
- Actionable error messages for administrators
- Prevents broken login flows
- Reduces support tickets

---

## Windows Development Patterns

### 1. Always Add Initial Delay for Named Pipe Creation

**Pattern:** When launching a process that creates a named pipe, add a 2-second delay before attempting to connect.

**Why:** Windows named pipes require the server to call `ListenPipe()` before clients can connect. Process startup + pipe creation takes time.

**Implementation:**
```go
// In health monitor or connection logic
if isWindows() {
    initialDelay := 2 * time.Second
    time.Sleep(initialDelay)  // Or use select with context for cancellation
}
```

**Applies To:**
- Tray → Core communication
- Any Windows service using named pipes
- IPC between processes on Windows

---

### 2. Split Platform-Specific Shell Launchers

**Pattern:** Never put shell execution logic in shared code. Always use `//go:build` tags for platform-specific files.

**Why:** Unix shells (`/bin/bash`, `/bin/zsh`) don't exist on Windows. Windows `cmd.exe` doesn't understand Unix flags (`-l`, `-c`).

**Correct Structure:**
```
cmd/app/
  main.go              # Shared code only
  main_darwin.go       //go:build darwin - macOS-specific
  main_windows.go      //go:build windows - Windows-specific
```

**Example:**
```go
// main_darwin.go
func wrapCoreLaunchWithShell(binary string, args []string) (string, []string, error) {
    return "/bin/zsh", []string{"-l", "-c", command}, nil
}

// main_windows.go
func wrapCoreLaunchWithShell(binary string, args []string) (string, []string, error) {
    return "cmd.exe", []string{"/c", command}, nil
}
```

---

### 3. Suppress Expected Startup Errors on Windows

**Pattern:** During initial startup, suppress "pipe not found" errors in logs.

**Why:** These errors are expected and normal during the brief window while the pipe is being created. Logging them creates noise and confusion.

**Implementation:**
```go
if isWindows() && strings.Contains(errMsg, "The system cannot find the file specified") {
    logger.Debug("Pipe not found - core starting (expected)")
    // Don't log as warning/error
} else {
    logger.Warn("Actual error", err)
}
```

---

### 4. Use tasklist/taskkill for Process Management on Windows

**Pattern:** Windows doesn't support Unix signals. Use `tasklist` and `taskkill` commands.

**Why:** Windows uses different process management APIs. Signals like SIGTERM don't exist.

**Examples:**
```go
// Check if process exists
exec.Command("tasklist", "/FI", "PID eq 1234", "/FO", "CSV", "/NH")

// Kill process tree
exec.Command("taskkill", "/PID", "1234", "/T", "/F")  // /T = tree, /F = force
```

---

### 5. Always Quote Windows Command-Line Arguments

**Pattern:** Use proper quoting for Windows command-line arguments, especially with spaces.

**Why:** Windows command parsing is different from Unix shells. Spaces and special characters need escaping.

**Implementation:**
```go
func windowsQuote(arg string) string {
    if strings.ContainsAny(arg, " \t\n\v\"") {
        escaped := strings.ReplaceAll(arg, `"`, `\"`)
        return `"` + escaped + `"`
    }
    return arg
}
```

---

## General Development Patterns

### 6. Build Frontend Before Embedding

**Pattern:** Always build frontend assets before running `go build` if using `//go:embed`.

**Why:** Go embed requires files to exist at compile time. Missing files cause build failures.

**Workflow:**
```bash
cd frontend && npm install && npm run build
cp -r frontend/dist web/frontend/
go build -o app ./cmd/app
```

---

### 7. Test Both Binaries Together

**Pattern:** When developing tray + core architecture, always test both binaries from the same directory.

**Why:** The tray looks for the core binary in the same directory first. Different directories cause "binary not found" errors.

**Test Command:**
```bash
cd E:\Projects\Go\mcpproxy-go
taskkill /F /IM mcpproxy*.exe 2>$null
.\mcpproxy-tray.exe
tasklist | findstr mcpproxy  # Verify both running
```

---

## Debugging Tips

### Windows Named Pipe Debugging

```powershell
# Check if pipe exists (PowerShell)
Get-ChildItem \\.\pipe\ | Where-Object Name -like "*mcpproxy*"

# Check process listening on pipe (advanced)
# Use Process Explorer or similar tools
```

### Core Launch Debugging

```bash
# Run tray from project directory
cd E:\Projects\Go\mcpproxy-go
.\mcpproxy-tray.exe

# Check both processes
tasklist | findstr mcpproxy

# Check listening ports
netstat -ano | findstr :3303  # Or configured port
```

---

## Architecture Decisions

### Why Separate Tray and Core?

1. **Auto-start**: Tray can auto-launch core on login
2. **Port conflict resolution**: Tray can detect and handle port conflicts
3. **Independent operation**: Tray provides UI even if core crashes
4. **Real-time sync**: SSE + socket communication for live updates

### Why Named Pipes Over TCP for Tray-Core?

1. **Security**: OS-level authentication (same user only)
2. **No API key needed**: Socket connections bypass API key validation
3. **Performance**: Lower latency than TCP loopback
4. **Firewall friendly**: No network exposure

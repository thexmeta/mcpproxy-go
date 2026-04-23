# GitHub OAuth Troubleshooting Guide

## Issue: Authenticated from GitHub but Nothing Happens on Proxy Side

When you authenticate with GitHub but the proxy doesn't redirect back or show any response, follow these diagnostic steps:

---

## 1. Check Callback URL Configuration

The **most common issue** is a callback URL mismatch between GitHub OAuth App settings and the proxy.

### GitHub OAuth App Settings

Go to: https://github.com/settings/developers → Your OAuth App → Settings

**Authorization callback URL** must exactly match:

```
# For local development (HTTP):
http://localhost:8080/api/v1/auth/callback

# For production (HTTPS):
https://your-domain.com/api/v1/auth/callback

# For LAN access:
http://192.168.1.100:8080/api/v1/auth/callback
```

### Common Mistakes

❌ **Wrong:**
- `http://localhost:8080/ui/callback` (wrong path)
- `http://localhost:8080/auth/callback` (missing `/api/v1`)
- `http://localhost:8080/api/v1/auth` (missing `/callback`)
- `https://localhost:8080/api/v1/auth/callback` (wrong scheme for HTTP)

✅ **Correct:**
- `http://localhost:8080/api/v1/auth/callback`

---

## 2. Check Proxy Logs

Run the proxy with debug logging to see what's happening:

```bash
# Windows PowerShell
.\mcpproxy.exe serve --log-level=debug

# Or check the log file
Get-Content ~\.mcpproxy\logs\main.log -Tail 100 -Wait
```

### Look for These Log Messages

**On Login Click:**
```
INFO initiating OAuth login
  provider: github
  client_id: Iv1.abc123...
```

**On GitHub Callback:**
```
INFO OAuth callback received
  provider: github
  has_code: true
  has_state: true
  remote_addr: 127.0.0.1:xxxxx
```

**If you DON'T see "OAuth callback received":**
- GitHub is not redirecting back
- Callback URL mismatch (see step 1)
- GitHub OAuth app not configured correctly

**If you see callback but then errors:**
```
ERROR OAuth code exchange failed
  error: ...
  callback_url: http://localhost:8080/api/v1/auth/callback
```
- The `callback_url` in logs must match GitHub settings exactly

---

## 3. Verify OAuth Configuration

Check your MCPProxy config file (`~/.mcpproxy/mcp_config.json`):

```json
{
  "teams": {
    "enabled": true,
    "admin_emails": ["your-email@example.com"],
    "oauth": {
      "provider": "github",
      "client-id": "Iv1.abc123...",
      "client-secret": "your-client-secret-here",
      "allowed-domains": ["example.com"]
    }
  }
}
```

### Required Fields

| Field | Description | Example |
|-------|-------------|---------|
| `provider` | Must be `"github"` | `"github"` |
| `client-id` | GitHub OAuth Client ID | `"Iv1.abc123..."` |
| `client-secret` | GitHub OAuth Client Secret | `"shh_..."` |
| `allowed-domains` | Email domains allowed | `["example.com"]` |

### Get GitHub OAuth Credentials

1. Go to https://github.com/settings/developers
2. Click "New OAuth App" or select existing app
3. Copy **Client ID** and generate/copy **Client Secret**

---

## 4. Check Allowed Domains

If your GitHub email domain isn't in `allowed-domains`, login will fail:

```
WARN OAuth login domain not allowed
  email: user@gmail.com
  allowed_domains: ["example.com"]
```

**Fix:** Add your email domain to the config:

```json
{
  "teams": {
    "oauth": {
      "allowed-domains": ["example.com", "gmail.com"]
    }
  }
}
```

Or use `*` to allow all domains (not recommended for production):

```json
{
  "teams": {
    "oauth": {
      "allowed-domains": ["*"]
    }
  }
}
```

---

## 5. Test the OAuth Flow Manually

### Step 1: Initiate Login

Open browser to:
```
http://localhost:8080/ui/
```

Click "Sign in with your organization"

### Step 2: Check Browser Redirect

You should be redirected to GitHub:
```
https://github.com/login/oauth/authorize?
  client_id=Iv1.abc123...
  &redirect_uri=http://localhost:8080/api/v1/auth/callback
  &response_type=code
  &scope=user:email%20read:user
  &state=...
  &code_challenge=...
```

**If you DON'T see GitHub login page:**
- Check proxy is running: `.\mcpproxy.exe --version`
- Check proxy logs for errors
- Verify OAuth config has `client-id`

### Step 3: Authorize on GitHub

Click "Authorize" on GitHub

**If GitHub shows error:**
- "Redirect URI mismatch": Fix callback URL (step 1)
- "Invalid client_id": Check client-id in config
- "Application not found": OAuth app doesn't exist

### Step 4: Check Callback

After authorizing, GitHub redirects back to:
```
http://localhost:8080/api/v1/auth/callback?code=...&state=...
```

**If you see browser error:**
- 400 Bad Request: State validation failed (state expired or mismatch)
- 500 Internal Server Error: Check proxy logs for details

---

## 6. Common Error Messages

### "invalid or expired state parameter"

**Cause:** CSRF state mismatch or timeout

**Fix:**
- State expires after 10 minutes
- Try login again (don't wait too long)
- Check system time is synchronized

### "authentication failed"

**Cause:** Token exchange failed

**Check logs for:**
```
ERROR OAuth code exchange failed
  error: ...
```

**Common causes:**
- Wrong client secret
- Callback URL mismatch
- GitHub API rate limiting

### "email address is required"

**Cause:** GitHub account has no verified email

**Fix:**
- Add verified email to GitHub account
- Or check GitHub privacy settings

### "email domain not allowed"

**Cause:** Email domain not in `allowed-domains`

**Fix:** Add domain to config (see step 4)

---

## 7. Quick Diagnostic Commands

```bash
# Check if proxy is running
.\mcpproxy.exe --version

# Check proxy status
curl http://localhost:8080/api/v1/status

# Check OAuth config (requires API key)
curl -H "X-API-Key: your-api-key" \
  http://localhost:8080/api/v1/config | \
  jq '.config.teams.oauth'

# Tail logs
Get-Content ~\.mcpproxy\logs\main.log -Tail 50 -Wait
```

---

## 8. Still Not Working?

### Enable Verbose Logging

Edit config (`~/.mcpproxy/mcp_config.json`):

```json
{
  "logging": {
    "level": "debug",
    "enable_file": true,
    "enable_console": true
  }
}
```

Restart proxy and check logs:
```bash
.\mcpproxy.exe serve --log-level=debug
```

### Test with Different Browser

Some browsers cache OAuth redirects. Try:
- Incognito/Private mode
- Different browser (Chrome, Firefox, Edge)

### Check Network Tab

Open browser DevTools → Network tab:
1. Click login
2. Look for redirect to GitHub
3. After GitHub auth, look for callback to `/api/v1/auth/callback`
4. Check response status code

### Contact Support

If still stuck, provide:
- Proxy logs (last 100 lines)
- GitHub OAuth App settings (screenshot, redact secrets)
- Config file (redact secrets)
- Browser network tab screenshots

---

## 9. Working Configuration Example

### GitHub OAuth App Settings

```
Application name: MCPProxy Local Dev
Homepage URL: http://localhost:8080
Authorization callback URL: http://localhost:8080/api/v1/auth/callback
```

### MCPProxy Config

```json
{
  "listen": "127.0.0.1:8080",
  "teams": {
    "enabled": true,
    "admin_emails": ["dev@example.com"],
    "oauth": {
      "provider": "github",
      "client-id": "Iv1.abc123def456",
      "client-secret": "shh_secret_here",
      "allowed-domains": ["example.com"]
    }
  }
}
```

### Expected Flow

1. User clicks login on `http://localhost:8080/ui/`
2. Redirect to GitHub login
3. User authorizes
4. GitHub redirects to `http://localhost:8080/api/v1/auth/callback?code=...`
5. Proxy exchanges code for tokens
6. Proxy creates session and sets cookie
7. Redirect to `http://localhost:8080/ui/` (logged in)

---

**Last Updated:** March 15, 2026  
**Version:** v0.21.3

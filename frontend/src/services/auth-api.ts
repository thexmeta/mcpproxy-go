// Auth API client for server edition
const API_BASE = '/api/v1'

export interface UserProfile {
  id: string
  email: string
  display_name: string
  role: 'admin' | 'user'
  provider: string
  created_at: string
  last_login_at: string
}

export interface BearerTokenResponse {
  token: string
  expires_at: string
}

export const authApi = {
  // Get current user profile (returns null if not authenticated)
  async getMe(): Promise<UserProfile | null> {
    try {
      const response = await fetch(`${API_BASE}/auth/me`, { credentials: 'include' })
      if (response.status === 401) return null
      if (!response.ok) throw new Error(`HTTP ${response.status}`)
      return await response.json()
    } catch {
      return null
    }
  },

  // Generate bearer token for MCP clients
  async generateToken(): Promise<BearerTokenResponse> {
    const response = await fetch(`${API_BASE}/auth/token`, {
      method: 'POST',
      credentials: 'include',
    })
    if (!response.ok) throw new Error(`HTTP ${response.status}`)
    return await response.json()
  },

  // Log out
  async logout(): Promise<void> {
    await fetch(`${API_BASE}/auth/logout`, {
      method: 'POST',
      credentials: 'include',
    })
  },

  // Get login URL
  getLoginUrl(redirectUri?: string): string {
    const params = new URLSearchParams()
    if (redirectUri) params.set('redirect_uri', redirectUri)
    return `${API_BASE}/auth/login${params.toString() ? '?' + params.toString() : ''}`
  },

  // Check if OAuth is properly configured (returns error message if not)
  async checkOAuthConfig(): Promise<string | null> {
    try {
      // Make a test request to the login endpoint to check configuration
      const response = await fetch(`${API_BASE}/auth/login`, {
        method: 'GET',
        redirect: 'manual', // Don't follow redirect, just check response
      })
      
      // If we get 500, OAuth might not be configured
      if (response.status === 500) {
        const text = await response.text()
        if (text.includes('client_id') || text.includes('client_secret') || text.includes('OAuth not configured')) {
          return text || 'OAuth configuration error'
        }
      }
      return null // No error
    } catch (e) {
      // Network error or other issue
      return null
    }
  },
}

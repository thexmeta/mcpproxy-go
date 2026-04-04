// Package contracts defines typed data transfer objects for API communication
package contracts

import (
	"time"
)

// APIResponse is the standard wrapper for all API responses
type APIResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty" swaggertype:"object"`
	Error     string      `json:"error,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
}

// Server represents an upstream MCP server configuration and status
type Server struct {
	ID                   string               `json:"id"`
	Name                 string               `json:"name"`
	URL                  string               `json:"url,omitempty"`
	Protocol             string               `json:"protocol"`
	Command              string               `json:"command,omitempty"`
	Args                 []string             `json:"args,omitempty"`
	WorkingDir           string               `json:"working_dir,omitempty"`
	Env                  map[string]string    `json:"env,omitempty"`
	Headers              map[string]string    `json:"headers,omitempty"`
	OAuth                *OAuthConfig         `json:"oauth,omitempty"`
	Enabled              bool                 `json:"enabled"`
	Quarantined          bool                 `json:"quarantined"`
	Connected            bool                 `json:"connected"`
	Connecting           bool                 `json:"connecting"`
	Status               string               `json:"status"`
	LastError            string               `json:"last_error,omitempty"`
	ConnectedAt          *time.Time           `json:"connected_at,omitempty"`
	LastReconnectAt      *time.Time           `json:"last_reconnect_at,omitempty"`
	ReconnectCount       int                  `json:"reconnect_count"`
	ToolCount            int                  `json:"tool_count"`
	Created              time.Time            `json:"created"`
	Updated              time.Time            `json:"updated"`
	Isolation            *IsolationConfig     `json:"isolation,omitempty"`
	IsolationDefaults    *IsolationDefaults   `json:"isolation_defaults,omitempty"`
	Authenticated        bool                 `json:"authenticated"`                  // OAuth authentication status
	OAuthStatus          string               `json:"oauth_status,omitempty"`         // OAuth status: "authenticated", "expired", "error", "none"
	TokenExpiresAt       *time.Time           `json:"token_expires_at,omitempty"`     // When the OAuth token expires (ISO 8601)
	ToolListTokenSize    int                  `json:"tool_list_token_size,omitempty"` // Token size for this server's tools
	ShouldRetry          bool                 `json:"should_retry,omitempty"`
	RetryCount           int                  `json:"retry_count,omitempty"`
	LastRetryTime        *time.Time           `json:"last_retry_time,omitempty"`
	UserLoggedOut        bool                 `json:"user_logged_out,omitempty"`  // True if user explicitly logged out (prevents auto-reconnection)
	Health               *HealthStatus        `json:"health,omitempty"`           // Unified health status calculated by the backend
	Quarantine           *QuarantineStats     `json:"quarantine,omitempty"`       // Tool quarantine metrics for this server
	ReconnectOnUse       bool                 `json:"reconnect_on_use,omitempty"` // Attempt reconnection when a tool call targets this disconnected server
	SecurityScan         *SecurityScanSummary `json:"security_scan,omitempty"`    // Latest security scan results summary
	Diagnostic           *Diagnostic          `json:"diagnostic,omitempty"`       // Spec 044 diagnostic details
	ErrorCode            string               `json:"error_code,omitempty"`       // Spec 044 stable error code
	ExcludeDisabledTools bool                 `json:"exclude_disabled_tools,omitempty"` // If true, exclude disabled tools from all tool listings
	DisabledTools        []string             `json:"disabled_tools,omitempty"`         // List of explicitly disabled tool names
}

// Diagnostic is the REST-API representation of a classified server failure.
// It is additive: clients that pre-date spec 044 simply ignore it.
type Diagnostic struct {
	Code        string              `json:"code"`
	Severity    string              `json:"severity"`
	Cause       string              `json:"cause,omitempty"`
	DetectedAt  *time.Time          `json:"detected_at,omitempty"`
	UserMessage string              `json:"user_message,omitempty"`
	FixSteps    []DiagnosticFixStep `json:"fix_steps,omitempty"`
	DocsURL     string              `json:"docs_url,omitempty"`
}

// DiagnosticFixStep mirrors internal/diagnostics.FixStep for the REST API.
type DiagnosticFixStep struct {
	Type        string `json:"type"`
	Label       string `json:"label"`
	Command     string `json:"command,omitempty"`
	URL         string `json:"url,omitempty"`
	FixerKey    string `json:"fixer_key,omitempty"`
	Destructive bool   `json:"destructive,omitempty"`
}

// SecurityScanSummary provides a compact scan status for the server list view.
type SecurityScanSummary struct {
	LastScanAt    *time.Time     `json:"last_scan_at,omitempty"`
	RiskScore     int            `json:"risk_score"` // 0-100
	Status        string         `json:"status"`     // "clean", "warnings", "dangerous", "failed", "not_scanned", "scanning"
	FindingCounts *FindingCounts `json:"finding_counts,omitempty"`
}

// FindingCounts groups findings by user-facing threat category.
type FindingCounts struct {
	Dangerous int `json:"dangerous"` // Tool poisoning, active prompt injection
	Warning   int `json:"warning"`   // Rug pull, supply chain CVEs with exploits
	Info      int `json:"info"`      // Low-severity CVEs, informational
	Total     int `json:"total"`
}

// QuarantineStats represents tool quarantine metrics for a server.
type QuarantineStats struct {
	PendingCount int `json:"pending_count"` // Number of newly discovered tools awaiting approval
	ChangedCount int `json:"changed_count"` // Number of tools whose description/schema changed since approval
}

// OAuthConfig represents OAuth configuration for a server
type OAuthConfig struct {
	AuthURL        string            `json:"auth_url"`
	TokenURL       string            `json:"token_url"`
	ClientID       string            `json:"client_id"`
	Scopes         []string          `json:"scopes,omitempty"`
	ExtraParams    map[string]string `json:"extra_params,omitempty"`
	RedirectPort   int               `json:"redirect_port,omitempty"`
	PKCEEnabled    bool              `json:"pkce_enabled,omitempty"`
	TokenExpiresAt *time.Time        `json:"token_expires_at,omitempty"` // When the OAuth token expires
	TokenValid     bool              `json:"token_valid,omitempty"`      // Whether token is currently valid
}

// IsolationConfig represents Docker isolation configuration as it is
// exposed over the REST API. Mirrors the per-server overrides in
// config.IsolationConfig so the web UI and native tray can both edit
// the isolation policy for a specific server.
type IsolationConfig struct {
	Enabled     *bool    `json:"enabled,omitempty"`
	Image       string   `json:"image,omitempty"`
	NetworkMode string   `json:"network_mode,omitempty"`
	ExtraArgs   []string `json:"extra_args,omitempty"`
	WorkingDir  string   `json:"working_dir,omitempty"`
	MemoryLimit string   `json:"memory_limit,omitempty"`
	CPULimit    string   `json:"cpu_limit,omitempty"`
	Timeout     string   `json:"timeout,omitempty"`
}

// IsolationDefaults represents the baseline Docker isolation settings
// that would apply to a server when it has no overrides.
type IsolationDefaults struct {
	Enabled           bool              `json:"enabled"`
	DefaultImages     map[string]string `json:"default_images,omitempty"`
	Registry          string            `json:"registry,omitempty"`
	NetworkMode       string            `json:"network_mode,omitempty"`
	MemoryLimit       string            `json:"memory_limit,omitempty"`
	CPULimit          string            `json:"cpu_limit,omitempty"`
	Timeout           string            `json:"timeout,omitempty"`
	EnableCacheVolume bool              `json:"enable_cache_volume"`
}

// HealthStatus provides a unified view of server connectivity and readiness.
type HealthStatus struct {
	Healthy     bool       `json:"healthy"`
	Status      string     `json:"status"` // "connected", "connecting", "failed", "disabled", "not_scanned"
	Message     string     `json:"message,omitempty"`
	LastCheckAt *time.Time `json:"last_check_at,omitempty"`
}

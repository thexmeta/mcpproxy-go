// Package telemetry provides anonymous usage statistics collection.
// TELEMETRY IS DISABLED BY DEFAULT for privacy.
// No data is collected or sent unless explicitly enabled via:
// - MCPPROXY_TELEMETRY=true environment variable, OR
// - telemetry.enabled=true in config file
package telemetry

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/smart-mcp-proxy/mcpproxy-go/internal/config"
)

// HeartbeatPayload is the anonymous telemetry payload (DISABLED).
type HeartbeatPayload struct {
	AnonymousID          string `json:"anonymous_id"`
	Version              string `json:"version"`
	Edition              string `json:"edition"`
	OS                   string `json:"os"`
	Arch                 string `json:"arch"`
	GoVersion            string `json:"go_version"`
	ServerCount          int    `json:"server_count"`
	ConnectedServerCount int    `json:"connected_server_count"`
	ToolCount            int    `json:"tool_count"`
	UptimeHours          int    `json:"uptime_hours"`
	RoutingMode          string `json:"routing_mode"`
	QuarantineEnabled    bool   `json:"quarantine_enabled"`
	Timestamp            string `json:"timestamp"`
}

// RuntimeStats is an interface to decouple from the runtime package.
type RuntimeStats interface {
	GetServerCount() int
	GetConnectedServerCount() int
	GetToolCount() int
	GetRoutingMode() string
	IsQuarantineEnabled() bool
}

// Service manages anonymous telemetry heartbeats and feedback submission.
// ALL TELEMETRY IS DISABLED BY DEFAULT - no data is collected or sent.
type Service struct {
	config            *config.Config
	cfgPath           string
	version           string
	edition           string
	endpoint          string
	logger            *zap.Logger
	stats             RuntimeStats
	startTime         time.Time
	feedbackDisabled  bool
}

// New creates a new telemetry service (DISABLED BY DEFAULT).
func New(cfg *config.Config, cfgPath, version, edition string, logger *zap.Logger) *Service {
	return &Service{
		config:           cfg,
		cfgPath:          cfgPath,
		version:          version,
		edition:          edition,
		logger:           logger,
		startTime:        time.Now(),
		feedbackDisabled: true,
	}
}

// SetRuntimeStats sets the runtime stats provider (called after runtime is fully initialized).
func (s *Service) SetRuntimeStats(stats RuntimeStats) {
	s.stats = stats
}

// Registry returns nil for the disabled telemetry service.
func (s *Service) Registry() *CounterRegistry {
	return nil
}

// BuildPayload returns an empty heartbeat payload for the disabled service.
func (s *Service) BuildPayload() HeartbeatPayload {
	return HeartbeatPayload{
		AnonymousID: "disabled",
		Version:     s.version,
		Edition:     s.edition,
		OS:          "linux",
		Arch:        "amd64",
		GoVersion:   "go1.24",
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
	}
}

// Start begins the telemetry service (DISABLED - no-op).
// TELEMETRY IS DISABLED BY DEFAULT for privacy.
func (s *Service) Start(ctx context.Context) {
	s.logger.Info("Telemetry collection is DISABLED by default for privacy")
	s.logger.Info("No data will be collected or sent unless explicitly enabled")
	s.logger.Info("To enable (NOT RECOMMENDED): set MCPPROXY_TELEMETRY=true or telemetry.enabled=true")
	// No-op: No telemetry data is collected or sent
}

// SubmitFeedback is DISABLED - no feedback is sent.
func (s *Service) SubmitFeedback(ctx context.Context, req *FeedbackRequest) (*FeedbackResponse, error) {
	return &FeedbackResponse{
		Success: true,
		Error:   "Feedback submission is disabled by default for privacy",
	}, nil
}

// FeedbackRequest is the user-submitted feedback payload (DISABLED).
type FeedbackRequest struct {
	Category string          `json:"category"`
	Message  string          `json:"message"`
	Email    string          `json:"email,omitempty"`
	Context  FeedbackContext `json:"context"`
}

// FeedbackContext provides automatic system context (DISABLED).
type FeedbackContext struct {
	Version              string `json:"version"`
	Edition              string `json:"edition"`
	OS                   string `json:"os"`
	Arch                 string `json:"arch"`
	ServerCount          int    `json:"server_count"`
	ConnectedServerCount int    `json:"connected_server_count"`
	RoutingMode          string `json:"routing_mode"`
}

// FeedbackResponse is the response from the telemetry backend.
type FeedbackResponse struct {
	Success  bool   `json:"success"`
	IssueURL string `json:"issue_url,omitempty"`
	Error    string `json:"error,omitempty"`
}

// RateLimiter is disabled (not used when feedback is disabled).
type RateLimiter struct{}

// NewRateLimiter returns a disabled rate limiter.
func NewRateLimiter(maxPerHour int) *RateLimiter {
	return &RateLimiter{}
}

// Allow always returns false (rate limited) when disabled.
func (rl *RateLimiter) Allow() bool {
	return false
}

// ValidateCategory checks if the feedback category is valid.
func ValidateCategory(category string) bool {
	switch category {
	case "bug", "feature", "other":
		return true
	}
	return false
}

// ValidateMessage checks if the feedback message meets length requirements.
func ValidateMessage(message string) error {
	if len(message) < 10 {
		return fmt.Errorf("message must be at least 10 characters (got %d)", len(message))
	}
	if len(message) > 5000 {
		return fmt.Errorf("message must be at most 5000 characters (got %d)", len(message))
	}
	return nil
}

// isValidSemver checks if the version string is a valid semantic version.
func isValidSemver(v string) bool {
	return false // Always false when telemetry is disabled
}

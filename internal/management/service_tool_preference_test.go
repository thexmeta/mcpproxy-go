package management

import (
	"testing"

	"github.com/smart-mcp-proxy/mcpproxy-go/internal/config"
	"go.uber.org/zap"
)

// Test helper functions for slice manipulation
func TestAddToSlice(t *testing.T) {
	slice := []string{}

	// Add first item
	slice = addToSlice(slice, "tool1")
	if len(slice) != 1 {
		t.Errorf("Expected 1 item, got %d", len(slice))
	}
	if slice[0] != "tool1" {
		t.Errorf("Expected tool1, got %s", slice[0])
	}

	// Add second item
	slice = addToSlice(slice, "tool2")
	if len(slice) != 2 {
		t.Errorf("Expected 2 items, got %d", len(slice))
	}

	// Duplicate add should be idempotent
	slice = addToSlice(slice, "tool1")
	if len(slice) != 2 {
		t.Errorf("Expected idempotent add, got %d items", len(slice))
	}
}

func TestRemoveFromSlice(t *testing.T) {
	slice := []string{"tool1", "tool2", "tool3"}

	// Remove first item
	slice = removeFromSlice(slice, "tool1")
	if len(slice) != 2 {
		t.Errorf("Expected 2 items, got %d", len(slice))
	}
	if slice[0] != "tool2" || slice[1] != "tool3" {
		t.Errorf("Unexpected order: %v", slice)
	}

	// Remove last item
	slice = removeFromSlice(slice, "tool3")
	if len(slice) != 1 {
		t.Errorf("Expected 1 item, got %d", len(slice))
	}

	// Remove middle item
	slice = []string{"tool1", "tool2", "tool3"}
	slice = removeFromSlice(slice, "tool2")
	if len(slice) != 2 {
		t.Errorf("Expected 2 items, got %d", len(slice))
	}
	if slice[0] != "tool1" || slice[1] != "tool3" {
		t.Errorf("Unexpected order: %v", slice)
	}

	// Remove non-existent should be safe
	slice = removeFromSlice(slice, "nonexistent")
	if len(slice) != 2 {
		t.Errorf("Expected no change, got %d items", len(slice))
	}
}

func TestService_GetToolPreferences(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	cfg := &config.Config{
		Servers: []*config.ServerConfig{
			{
				Name:          "server1",
				DisabledTools: []string{"tool2", "tool3"},
			},
		},
	}

	svc := &ServiceImpl{
		config: cfg,
		logger: logger.Sugar(),
	}

	prefs, err := svc.GetToolPreferences(nil, "server1")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(prefs) != 2 {
		t.Errorf("Expected 2 preferences, got %d", len(prefs))
	}

	// tool1 should not be in the list (it's enabled)
	if _, ok := prefs["tool1"]; ok {
		t.Errorf("tool1 should not be in disabled preferences")
	}

	// tool2 and tool3 should be disabled
	if _, ok := prefs["tool2"]; !ok {
		t.Errorf("tool2 should be in disabled preferences")
	}
	// Enabled=false means the tool is disabled
	if prefs["tool2"].Enabled != false {
		t.Errorf("tool2 should be marked as disabled (Enabled=false)")
	}
}

func TestService_GetToolPreferences_NoDisabled(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	cfg := &config.Config{
		Servers: []*config.ServerConfig{
			{
				Name:          "server1",
				DisabledTools: []string{},
			},
		},
	}

	svc := &ServiceImpl{
		config: cfg,
		logger: logger.Sugar(),
	}

	prefs, err := svc.GetToolPreferences(nil, "server1")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(prefs) != 0 {
		t.Errorf("Expected 0 preferences, got %d", len(prefs))
	}
}

func TestService_GetToolPreferences_ServerNotFound(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	cfg := &config.Config{
		Servers: []*config.ServerConfig{},
	}

	svc := &ServiceImpl{
		config: cfg,
		logger: logger.Sugar(),
	}

	prefs, err := svc.GetToolPreferences(nil, "nonexistent")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(prefs) != 0 {
		t.Errorf("Expected 0 preferences for nonexistent server, got %d", len(prefs))
	}
}

func TestService_BulkUpdateToolPreferences(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	_ = logger // Suppress unused warning

	cfg := &config.Config{
		Servers: []*config.ServerConfig{
			{
				Name:          "server1",
				DisabledTools: []string{},
			},
		},
	}

	// This test just verifies the config structure exists
	if len(cfg.Servers) != 1 {
		t.Errorf("Expected 1 server, got %d", len(cfg.Servers))
	}

	// Verify ServerConfig helper methods work
	server := cfg.Servers[0]

	// Test IsToolDisabled
	if server.IsToolDisabled("tool1") {
		t.Errorf("tool1 should not be disabled initially")
	}

	// Test DisableTool
	server.DisableTool("tool1")
	if !server.IsToolDisabled("tool1") {
		t.Errorf("tool1 should be disabled after DisableTool")
	}

	// Test EnableTool
	server.EnableTool("tool1")
	if server.IsToolDisabled("tool1") {
		t.Errorf("tool1 should not be disabled after EnableTool")
	}

	// Disable multiple tools
	server.DisableTool("tool1")
	server.DisableTool("tool2")
	server.DisableTool("tool3")

	if len(server.DisabledTools) != 3 {
		t.Errorf("Expected 3 disabled tools, got %d", len(server.DisabledTools))
	}

	// Duplicate disable should be idempotent
	server.DisableTool("tool1")
	if len(server.DisabledTools) != 3 {
		t.Errorf("Duplicate disable should be idempotent, got %d", len(server.DisabledTools))
	}
}

func TestServerConfig_DisabledToolsHelpers(t *testing.T) {
	server := &config.ServerConfig{
		Name:          "test-server",
		DisabledTools: []string{"tool1"},
	}

	// Test IsToolDisabled
	if !server.IsToolDisabled("tool1") {
		t.Errorf("tool1 should be disabled")
	}
	if server.IsToolDisabled("tool2") {
		t.Errorf("tool2 should not be disabled")
	}

	// Test DisableTool
	server.DisableTool("tool2")
	if len(server.DisabledTools) != 2 {
		t.Errorf("Expected 2 disabled tools, got %d", len(server.DisabledTools))
	}

	// Test EnableTool
	server.EnableTool("tool1")
	if len(server.DisabledTools) != 1 {
		t.Errorf("Expected 1 disabled tool after enable, got %d", len(server.DisabledTools))
	}
	if server.IsToolDisabled("tool1") {
		t.Errorf("tool1 should not be disabled after EnableTool")
	}
}

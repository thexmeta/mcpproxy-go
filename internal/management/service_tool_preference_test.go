package management

import (
	"fmt"
	"strings"
	"testing"

	"github.com/smart-mcp-proxy/mcpproxy-go/internal/config"
	"github.com/smart-mcp-proxy/mcpproxy-go/internal/contracts"
	"github.com/smart-mcp-proxy/mcpproxy-go/internal/storage"
	"github.com/smart-mcp-proxy/mcpproxy-go/internal/upstream/core"
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

// Tests for custom tool names and descriptions

func TestService_UpdateToolPreference_WithCustomFields(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	cfg := &config.Config{
		Servers: []*config.ServerConfig{
			{
				Name:          "server1",
				DisabledTools: []string{},
			},
		},
	}

	// Create mock storage
	mockStorage := &mockStorage{
		preferences: make(map[string]*storage.ToolPreferenceRecord),
	}

	svc := &ServiceImpl{
		config:  cfg,
		storage: mockStorage,
		logger:  logger.Sugar(),
		runtime: &mockRuntime{
			config: cfg,
		},
	}

	// Update tool preference with custom name and description
	pref := &contracts.ToolPreference{
		Enabled:           true,
		CustomName:        "custom_tool_name",
		CustomDescription: "This is a custom description",
	}

	err := svc.UpdateToolPreference(nil, "server1", "original_tool", pref)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify storage was updated
	record, err := mockStorage.GetToolPreference("server1", "original_tool")
	if err != nil {
		t.Fatalf("Failed to get preference from storage: %v", err)
	}

	if record.CustomName != "custom_tool_name" {
		t.Errorf("Expected custom name 'custom_tool_name', got '%s'", record.CustomName)
	}

	if record.CustomDescription != "This is a custom description" {
		t.Errorf("Expected custom description 'This is a custom description', got '%s'", record.CustomDescription)
	}

	if !record.Enabled {
		t.Errorf("Expected tool to be enabled")
	}
}

func TestService_GetToolPreferences_WithCustomFields(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	cfg := &config.Config{
		Servers: []*config.ServerConfig{
			{
				Name:          "server1",
				DisabledTools: []string{},
			},
		},
	}

	// Create mock storage with custom fields
	mockStorage := &mockStorage{
		preferences: map[string]*storage.ToolPreferenceRecord{
			"server1:tool1": {
				ServerName:        "server1",
				ToolName:          "tool1",
				Enabled:           true,
				CustomName:        "my_custom_tool",
				CustomDescription: "My custom tool description",
			},
			"server1:tool2": {
				ServerName: "server1",
				ToolName:   "tool2",
				Enabled:    false,
				// No custom fields
			},
		},
	}

	svc := &ServiceImpl{
		config:  cfg,
		storage: mockStorage,
		logger:  logger.Sugar(),
	}

	prefs, err := svc.GetToolPreferences(nil, "server1")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(prefs) != 2 {
		t.Errorf("Expected 2 preferences, got %d", len(prefs))
	}

	// tool1 should have custom fields
	tool1, ok := prefs["tool1"]
	if !ok {
		t.Errorf("tool1 not found in preferences")
	} else {
		if tool1.CustomName != "my_custom_tool" {
			t.Errorf("Expected custom name 'my_custom_tool', got '%s'", tool1.CustomName)
		}
		if tool1.CustomDescription != "My custom tool description" {
			t.Errorf("Expected custom description 'My custom tool description', got '%s'", tool1.CustomDescription)
		}
		if !tool1.Enabled {
			t.Errorf("Expected tool1 to be enabled")
		}
	}

	// tool2 should not have custom fields
	tool2, ok := prefs["tool2"]
	if !ok {
		t.Errorf("tool2 not found in preferences")
	} else {
		if tool2.CustomName != "" {
			t.Errorf("Expected empty custom name for tool2, got '%s'", tool2.CustomName)
		}
		if tool2.CustomDescription != "" {
			t.Errorf("Expected empty custom description for tool2, got '%s'", tool2.CustomDescription)
		}
		if tool2.Enabled {
			t.Errorf("Expected tool2 to be disabled")
		}
	}
}

func TestService_UpdateToolPreference_OnlyCustomName(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	cfg := &config.Config{
		Servers: []*config.ServerConfig{
			{
				Name:          "server1",
				DisabledTools: []string{},
			},
		},
	}

	mockStorage := &mockStorage{
		preferences: make(map[string]*storage.ToolPreferenceRecord),
	}

	svc := &ServiceImpl{
		config:  cfg,
		storage: mockStorage,
		logger:  logger.Sugar(),
		runtime: &mockRuntime{
			config: cfg,
		},
	}

	// Update only custom name
	pref := &contracts.ToolPreference{
		Enabled:    true,
		CustomName: "renamed_tool",
	}

	err := svc.UpdateToolPreference(nil, "server1", "original_tool", pref)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	record, err := mockStorage.GetToolPreference("server1", "original_tool")
	if err != nil {
		t.Fatalf("Failed to get preference from storage: %v", err)
	}

	if record.CustomName != "renamed_tool" {
		t.Errorf("Expected custom name 'renamed_tool', got '%s'", record.CustomName)
	}

	if record.CustomDescription != "" {
		t.Errorf("Expected empty custom description, got '%s'", record.CustomDescription)
	}
}

func TestService_UpdateToolPreference_OnlyCustomDescription(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	cfg := &config.Config{
		Servers: []*config.ServerConfig{
			{
				Name:          "server1",
				DisabledTools: []string{},
			},
		},
	}

	mockStorage := &mockStorage{
		preferences: make(map[string]*storage.ToolPreferenceRecord),
	}

	svc := &ServiceImpl{
		config:  cfg,
		storage: mockStorage,
		logger:  logger.Sugar(),
		runtime: &mockRuntime{
			config: cfg,
		},
	}

	// Update only custom description
	pref := &contracts.ToolPreference{
		Enabled:           true,
		CustomDescription: "Updated description for this tool",
	}

	err := svc.UpdateToolPreference(nil, "server1", "original_tool", pref)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	record, err := mockStorage.GetToolPreference("server1", "original_tool")
	if err != nil {
		t.Fatalf("Failed to get preference from storage: %v", err)
	}

	if record.CustomDescription != "Updated description for this tool" {
		t.Errorf("Expected custom description 'Updated description for this tool', got '%s'", record.CustomDescription)
	}

	if record.CustomName != "" {
		t.Errorf("Expected empty custom name, got '%s'", record.CustomName)
	}
}

// Mock storage for testing
type mockStorage struct {
	preferences map[string]*storage.ToolPreferenceRecord
}

func (m *mockStorage) GetToolPreference(serverName, toolName string) (*storage.ToolPreferenceRecord, error) {
	key := serverName + ":" + toolName
	if record, ok := m.preferences[key]; ok {
		return record, nil
	}
	return nil, fmt.Errorf("preference not found")
}

func (m *mockStorage) ListToolPreferences(serverName string) ([]*storage.ToolPreferenceRecord, error) {
	var records []*storage.ToolPreferenceRecord
	for key, record := range m.preferences {
		if strings.HasPrefix(key, serverName+":") {
			records = append(records, record)
		}
	}
	return records, nil
}

func (m *mockStorage) SaveToolPreference(record *storage.ToolPreferenceRecord) error {
	key := record.ServerName + ":" + record.ToolName
	m.preferences[key] = record
	return nil
}

func (m *mockStorage) DeleteToolPreference(serverName, toolName string) error {
	key := serverName + ":" + toolName
	delete(m.preferences, key)
	return nil
}

// Mock runtime for testing
type mockRuntime struct {
	config *config.Config
}

func (m *mockRuntime) UpdateServerDisabledTools(serverName string, disabledTools []string) error {
	for _, server := range m.config.Servers {
		if server.Name == serverName {
			server.DisabledTools = disabledTools
			break
		}
	}
	return nil
}

func (m *mockRuntime) SaveConfiguration() error {
	return nil
}

// Stub implementations for other RuntimeOperations methods (not used in these tests)
func (m *mockRuntime) EnableServer(serverName string, enabled bool) error {
	return nil
}

func (m *mockRuntime) RestartServer(serverName string) error {
	return nil
}

func (m *mockRuntime) GetAllServers() ([]map[string]interface{}, error) {
	return nil, nil
}

func (m *mockRuntime) BulkEnableServers(serverNames []string, enabled bool) (map[string]error, error) {
	return nil, nil
}

func (m *mockRuntime) GetServerTools(serverName string) ([]map[string]interface{}, error) {
	return nil, nil
}

func (m *mockRuntime) GetAllServerTools(serverName string) ([]map[string]interface{}, error) {
	return nil, nil
}

func (m *mockRuntime) TriggerOAuthLogin(serverName string) error {
	return nil
}

func (m *mockRuntime) TriggerOAuthLoginQuick(serverName string) (*core.OAuthStartResult, error) {
	return nil, nil
}

func (m *mockRuntime) TriggerOAuthLogout(serverName string) error {
	return nil
}

func (m *mockRuntime) RefreshOAuthToken(serverName string) error {
	return nil
}

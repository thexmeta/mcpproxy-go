package storage

import (
	"os"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestToolPreference_SaveAndGet(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Create logger
	logger, _ := zap.NewDevelopment()
	sugar := logger.Sugar()

	// Create database
	db, err := NewBoltDB(tmpDir, sugar)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Create test record
	record := &ToolPreferenceRecord{
		ServerName:        "test-server",
		ToolName:          "test-tool",
		Enabled:           true,
		CustomName:        "Custom Tool Name",
		CustomDescription: "Custom description",
		Created:           time.Now(),
	}

	// Save
	err = db.SaveToolPreference(record)
	if err != nil {
		t.Fatalf("Failed to save tool preference: %v", err)
	}

	// Get
	retrieved, err := db.GetToolPreference("test-server", "test-tool")
	if err != nil {
		t.Fatalf("Failed to get tool preference: %v", err)
	}

	// Verify
	if retrieved.ServerName != "test-server" {
		t.Errorf("Expected server name 'test-server', got '%s'", retrieved.ServerName)
	}
	if retrieved.ToolName != "test-tool" {
		t.Errorf("Expected tool name 'test-tool', got '%s'", retrieved.ToolName)
	}
	if !retrieved.Enabled {
		t.Errorf("Expected enabled=true, got false")
	}
	if retrieved.CustomName != "Custom Tool Name" {
		t.Errorf("Expected custom name 'Custom Tool Name', got '%s'", retrieved.CustomName)
	}
	if retrieved.CustomDescription != "Custom description" {
		t.Errorf("Expected custom description 'Custom description', got '%s'", retrieved.CustomDescription)
	}
}

func TestToolPreference_GetNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	logger, _ := zap.NewDevelopment()
	db, _ := NewBoltDB(tmpDir, logger.Sugar())
	defer db.Close()

	_, err := db.GetToolPreference("nonexistent", "tool")
	if err == nil {
		t.Fatal("Expected error for non-existent tool preference")
	}
}

func TestToolPreference_List(t *testing.T) {
	tmpDir := t.TempDir()
	logger, _ := zap.NewDevelopment()
	db, _ := NewBoltDB(tmpDir, logger.Sugar())
	defer db.Close()

	// Create test records
	records := []*ToolPreferenceRecord{
		{ServerName: "server1", ToolName: "tool1", Enabled: true, CustomName: "Tool 1"},
		{ServerName: "server1", ToolName: "tool2", Enabled: false, CustomName: "Tool 2"},
		{ServerName: "server2", ToolName: "tool1", Enabled: true, CustomName: "Tool 1"},
	}

	for _, r := range records {
		if err := db.SaveToolPreference(r); err != nil {
			t.Fatalf("Failed to save: %v", err)
		}
	}

	// List all
	all, err := db.ListToolPreferences("")
	if err != nil {
		t.Fatalf("Failed to list all: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("Expected 3 records, got %d", len(all))
	}

	// List by server
	server1, err := db.ListToolPreferences("server1")
	if err != nil {
		t.Fatalf("Failed to list server1: %v", err)
	}
	if len(server1) != 2 {
		t.Errorf("Expected 2 records for server1, got %d", len(server1))
	}

	server2, err := db.ListToolPreferences("server2")
	if err != nil {
		t.Fatalf("Failed to list server2: %v", err)
	}
	if len(server2) != 1 {
		t.Errorf("Expected 1 record for server2, got %d", len(server2))
	}
}

func TestToolPreference_Delete(t *testing.T) {
	tmpDir := t.TempDir()
	logger, _ := zap.NewDevelopment()
	db, _ := NewBoltDB(tmpDir, logger.Sugar())
	defer db.Close()

	// Create and save
	record := &ToolPreferenceRecord{
		ServerName: "test-server",
		ToolName:   "test-tool",
		Enabled:    true,
	}
	if err := db.SaveToolPreference(record); err != nil {
		t.Fatalf("Failed to save: %v", err)
	}

	// Delete
	err := db.DeleteToolPreference("test-server", "test-tool")
	if err != nil {
		t.Fatalf("Failed to delete: %v", err)
	}

	// Verify deleted
	_, err = db.GetToolPreference("test-server", "test-tool")
	if err == nil {
		t.Fatal("Expected error after deletion")
	}
}

func TestToolPreference_DeleteServer(t *testing.T) {
	tmpDir := t.TempDir()
	logger, _ := zap.NewDevelopment()
	db, _ := NewBoltDB(tmpDir, logger.Sugar())
	defer db.Close()

	// Create test records
	records := []*ToolPreferenceRecord{
		{ServerName: "server1", ToolName: "tool1", Enabled: true},
		{ServerName: "server1", ToolName: "tool2", Enabled: false},
		{ServerName: "server2", ToolName: "tool1", Enabled: true},
	}

	for _, r := range records {
		if err := db.SaveToolPreference(r); err != nil {
			t.Fatalf("Failed to save: %v", err)
		}
	}

	// Delete all for server1
	err := db.DeleteServerToolPreferences("server1")
	if err != nil {
		t.Fatalf("Failed to delete server preferences: %v", err)
	}

	// Verify server1 deleted
	server1, err := db.ListToolPreferences("server1")
	if err != nil {
		t.Fatalf("Failed to list server1: %v", err)
	}
	if len(server1) != 0 {
		t.Errorf("Expected 0 records for server1, got %d", len(server1))
	}

	// Verify server2 still exists
	server2, err := db.ListToolPreferences("server2")
	if err != nil {
		t.Fatalf("Failed to list server2: %v", err)
	}
	if len(server2) != 1 {
		t.Errorf("Expected 1 record for server2, got %d", len(server2))
	}
}

func TestToolPreference_Update(t *testing.T) {
	tmpDir := t.TempDir()
	logger, _ := zap.NewDevelopment()
	db, _ := NewBoltDB(tmpDir, logger.Sugar())
	defer db.Close()

	// Create and save
	record := &ToolPreferenceRecord{
		ServerName:      "test-server",
		ToolName:        "test-tool",
		Enabled:         true,
		CustomName:      "Original Name",
		CustomDescription: "Original description",
	}
	if err := db.SaveToolPreference(record); err != nil {
		t.Fatalf("Failed to save: %v", err)
	}

	// Update
	record.Enabled = false
	record.CustomName = "Updated Name"
	record.CustomDescription = "Updated description"
	if err := db.SaveToolPreference(record); err != nil {
		t.Fatalf("Failed to update: %v", err)
	}

	// Verify update
	retrieved, err := db.GetToolPreference("test-server", "test-tool")
	if err != nil {
		t.Fatalf("Failed to get: %v", err)
	}
	if retrieved.Enabled {
		t.Errorf("Expected enabled=false after update")
	}
	if retrieved.CustomName != "Updated Name" {
		t.Errorf("Expected 'Updated Name', got '%s'", retrieved.CustomName)
	}
	if retrieved.CustomDescription != "Updated description" {
		t.Errorf("Expected 'Updated description', got '%s'", retrieved.CustomDescription)
	}
}

func TestToolPreference_Key(t *testing.T) {
	record := &ToolPreferenceRecord{
		ServerName: "my-server",
		ToolName:   "my-tool",
	}

	expected := "my-server:my-tool"
	if record.Key() != expected {
		t.Errorf("Expected key '%s', got '%s'", expected, record.Key())
	}
}

func TestToolPreference_MarshalUnmarshal(t *testing.T) {
	original := &ToolPreferenceRecord{
		ServerName:        "test-server",
		ToolName:          "test-tool",
		Enabled:           true,
		CustomName:        "Custom",
		CustomDescription: "Description",
		Created:           time.Now(),
		Updated:           time.Now(),
	}

	// Marshal
	data, err := original.MarshalBinary()
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Unmarshal
	var unmarshaled ToolPreferenceRecord
	err = unmarshaled.UnmarshalBinary(data)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Verify
	if unmarshaled.ServerName != original.ServerName {
		t.Errorf("ServerName mismatch")
	}
	if unmarshaled.ToolName != original.ToolName {
		t.Errorf("ToolName mismatch")
	}
	if unmarshaled.Enabled != original.Enabled {
		t.Errorf("Enabled mismatch")
	}
	if unmarshaled.CustomName != original.CustomName {
		t.Errorf("CustomName mismatch")
	}
	if unmarshaled.CustomDescription != original.CustomDescription {
		t.Errorf("CustomDescription mismatch")
	}
}

// TestMain sets up test database cleanup
func TestMain(m *testing.M) {
	// Run tests
	exitCode := m.Run()

	// Cleanup temp files
	os.RemoveAll("test.db")

	os.Exit(exitCode)
}

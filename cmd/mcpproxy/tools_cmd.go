package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/smart-mcp-proxy/mcpproxy-go/internal/cli/output"
	"github.com/smart-mcp-proxy/mcpproxy-go/internal/cliclient"
	"github.com/smart-mcp-proxy/mcpproxy-go/internal/config"
	"github.com/smart-mcp-proxy/mcpproxy-go/internal/logs"
	"github.com/smart-mcp-proxy/mcpproxy-go/internal/secret"
	"github.com/smart-mcp-proxy/mcpproxy-go/internal/socket"
	"github.com/smart-mcp-proxy/mcpproxy-go/internal/storage"
	"github.com/smart-mcp-proxy/mcpproxy-go/internal/transport"
	"github.com/smart-mcp-proxy/mcpproxy-go/internal/upstream/managed"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	toolsCmd = &cobra.Command{
		Use:   "tools",
		Short: "Tools management commands",
		Long:  "Commands for managing and debugging MCP tools from upstream servers",
	}

	toolsListCmd = &cobra.Command{
		Use:   "list",
		Short: "List tools from an upstream server",
		Long: `List all available tools from a specific upstream server.
This command is primarily used for debugging upstream server connections and tool discovery.

Examples:
  mcpproxy tools list --server=github-server --log-level=trace
  mcpproxy tools list --server=weather-api --log-level=debug
  mcpproxy tools list --server=local-script --log-level=info
  mcpproxy tools list --server=jetbrains-sse --trace-transport  # Enable HTTP/SSE frame tracing`,
		RunE: runToolsList,
	}

	toolsPreferencesCmd = &cobra.Command{
		Use:   "preferences",
		Short: "Manage tool preferences",
		Long:  "Commands for managing tool preferences (enable/disable, custom names)",
	}

	toolsPrefsListCmd = &cobra.Command{
		Use:   "list",
		Short: "List tool preferences for a server",
		Long:  "List all tool preferences (enable/disable state, custom names) for a specific server",
		RunE:  runToolsPrefsList,
	}

	toolsPrefsEnableCmd = &cobra.Command{
		Use:   "enable <server> <tool>",
		Short: "Enable a tool",
		Long:  "Enable a specific tool for a server",
		RunE:  runToolsPrefsEnable,
	}

	toolsPrefsDisableCmd = &cobra.Command{
		Use:   "disable <server> <tool>",
		Short: "Disable a tool",
		Long:  "Disable a specific tool for a server (hides it from AI agents)",
		RunE:  runToolsPrefsDisable,
	}

	toolsPrefsToggleCmd = &cobra.Command{
		Use:   "toggle <server> <tool>",
		Short: "Toggle tool enable/disable state",
		Long:  "Toggle the enable/disable state of a specific tool",
		RunE:  runToolsPrefsToggle,
	}

	// Command flags
	serverName     string
	toolsLogLevel  string
	configPath     string
	timeout        time.Duration
	traceTransport bool // Enable HTTP/SSE frame-by-frame tracing
)

// GetToolsCommand returns the tools command for adding to the root command
func GetToolsCommand() *cobra.Command {
	return toolsCmd
}

func init() {
	// toolsCmd will be added to rootCmd in main.go
	toolsCmd.AddCommand(toolsListCmd)
	toolsCmd.AddCommand(toolsPreferencesCmd)

	// Add preference subcommands
	toolsPreferencesCmd.AddCommand(toolsPrefsListCmd)
	toolsPreferencesCmd.AddCommand(toolsPrefsEnableCmd)
	toolsPreferencesCmd.AddCommand(toolsPrefsDisableCmd)
	toolsPreferencesCmd.AddCommand(toolsPrefsToggleCmd)

	// Define flags for tools list command
	toolsListCmd.Flags().StringVarP(&serverName, "server", "s", "", "Name of the upstream server to query (required)")
	toolsListCmd.Flags().StringVarP(&toolsLogLevel, "log-level", "l", "info", "Log level (trace, debug, info, warn, error)")
	toolsListCmd.Flags().StringVarP(&configPath, "config", "c", "", "Path to MCP configuration file (default: ~/.mcpproxy/mcp_config.json)")
	toolsListCmd.Flags().DurationVarP(&timeout, "timeout", "t", 30*time.Second, "Connection timeout")
	// Note: -o/--output flag is inherited from root command via globalOutputFormat
	toolsListCmd.Flags().BoolVar(&traceTransport, "trace-transport", false, "Enable detailed HTTP/SSE frame-by-frame tracing (useful for debugging SSE connection issues)")

	// Mark required flags
	err := toolsListCmd.MarkFlagRequired("server")
	if err != nil {
		panic(fmt.Sprintf("Failed to mark server flag as required: %v", err))
	}

	// Add examples and usage help
	toolsListCmd.Example = `  # List tools with trace logging to see all JSON-RPC frames
  mcpproxy tools list --server=github-server --log-level=trace

  # List tools with debug output
  mcpproxy tools list --server=weather-api --log-level=debug

  # Use custom config file
  mcpproxy tools list --server=local-script --config=/path/to/config.json

  # Set custom timeout
  mcpproxy tools list --server=slow-server --timeout=60s`

	// Preference commands examples
	toolsPrefsListCmd.Example = `  mcpproxy tools preferences list --server=github-server
  mcpproxy tools preferences list -s weather-api`

	toolsPrefsEnableCmd.Example = `  mcpproxy tools preferences enable github-server create_issue
  mcpproxy tools preferences enable my-server dangerous_tool`

	toolsPrefsDisableCmd.Example = `  mcpproxy tools preferences disable github-server delete_file
  mcpproxy tools preferences disable my-server risky_operation`

	toolsPrefsToggleCmd.Example = `  mcpproxy tools preferences toggle github-server some_tool`
}

func runToolsList(_ *cobra.Command, _ []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Enable transport tracing if requested
	if traceTransport {
		transport.GlobalTraceEnabled = true
		fmt.Println("🔍 HTTP/SSE TRANSPORT TRACING ENABLED")
		fmt.Println("   All HTTP requests/responses and SSE frames will be logged")
		fmt.Println()
	}

	// Load configuration
	globalConfig, err := loadToolsConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create logger
	logger, err := logs.SetupCommandLogger(false, toolsLogLevel, false, "")
	if err != nil {
		return fmt.Errorf("failed to setup logger: %w", err)
	}
	defer func() { _ = logger.Sync() }()

	// Check if daemon is running and use client mode
	if shouldUseToolsDaemon(globalConfig.DataDir) {
		logger.Info("Detected running daemon, using client mode via socket",
			zap.String("server", serverName))
		return runToolsListClientMode(ctx, globalConfig.DataDir, serverName, logger)
	}

	// No daemon detected, use standalone mode
	logger.Info("No daemon detected, using standalone mode",
		zap.String("server", serverName))
	return runToolsListStandalone(ctx, serverName, globalConfig, logger)
}

// loadToolsConfig loads the MCP configuration file for tools command
func loadToolsConfig() (*config.Config, error) {
	var configFilePath string

	if configPath != "" {
		configFilePath = configPath
	} else {
		// Use default path
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %w", err)
		}
		configFilePath = filepath.Join(homeDir, ".mcpproxy", "mcp_config.json")
	}

	// Check if config file exists
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("configuration file not found at %s. Please run 'mcpproxy' daemon first to create the config", configFilePath)
	}

	// Load configuration using file-based loading
	globalConfig, err := config.LoadFromFile(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config from %s: %w", configFilePath, err)
	}

	return globalConfig, nil
}

// getAvailableServerNames returns a list of available server names
func getAvailableServerNames(globalConfig *config.Config) []string {
	var names []string
	for _, server := range globalConfig.Servers {
		names = append(names, server.Name)
	}
	return names
}

// outputToolsFromMetadata formats and displays tools from ToolMetadata (standalone mode) using unified formatters.
func outputToolsFromMetadata(tools []*config.ToolMetadata, serverName string) error {
	// Convert to map format for unified output
	toolMaps := make([]map[string]interface{}, len(tools))
	for i, tool := range tools {
		toolMaps[i] = map[string]interface{}{
			"name":        tool.Name,
			"description": tool.Description,
			"server":      serverName,
			"full_name":   fmt.Sprintf("%s:%s", serverName, tool.Name),
		}
		// Include schema in debug/trace mode
		if (toolsLogLevel == "debug" || toolsLogLevel == "trace") && tool.ParamsJSON != "" {
			toolMaps[i]["schema"] = tool.ParamsJSON
		}
	}

	outputFormat := ResolveOutputFormat()
	formatter, err := GetOutputFormatter()
	if err != nil {
		return output.NewStructuredError(output.ErrCodeInvalidOutputFormat, err.Error()).
			WithGuidance("Use -o table, -o json, or -o yaml")
	}

	// For JSON/YAML, format directly
	if outputFormat == "json" || outputFormat == "yaml" {
		result, fmtErr := formatter.Format(toolMaps)
		if fmtErr != nil {
			return fmt.Errorf("failed to format output: %w", fmtErr)
		}
		fmt.Println(result)
		return nil
	}

	// Table format: show name and description
	headers := []string{"NAME", "DESCRIPTION"}
	var rows [][]string
	for _, tool := range tools {
		rows = append(rows, []string{tool.Name, tool.Description})
	}

	result, fmtErr := formatter.FormatTable(headers, rows)
	if fmtErr != nil {
		return fmt.Errorf("failed to format table: %w", fmtErr)
	}
	fmt.Print(result)
	return nil
}

// shouldUseToolsDaemon checks if daemon is running by detecting socket file.
func shouldUseToolsDaemon(dataDir string) bool {
	socketPath := socket.DetectSocketPath(dataDir)
	return socket.IsSocketAvailable(socketPath)
}

// runToolsListClientMode executes tools list via daemon HTTP API over socket.
func runToolsListClientMode(ctx context.Context, dataDir, serverName string, logger *zap.Logger) error {
	socketPath := socket.DetectSocketPath(dataDir)
	client := cliclient.NewClient(socketPath, logger.Sugar())

	// Ping daemon to verify connectivity
	pingCtx, pingCancel := context.WithTimeout(ctx, 2*time.Second)
	defer pingCancel()
	if err := client.Ping(pingCtx); err != nil {
		logger.Warn("Failed to ping daemon, falling back to standalone mode",
			zap.Error(err),
			zap.String("socket_path", socketPath))
		// Fall back to standalone mode
		cfg, err := loadToolsConfig()
		if err != nil {
			return fmt.Errorf("failed to load config for standalone mode: %w", err)
		}
		return runToolsListStandalone(ctx, serverName, cfg, logger)
	}

	fmt.Fprintf(os.Stderr, "ℹ️  Using daemon mode (via socket) - fast execution\n\n")

	// Fetch tools from daemon
	tools, err := client.GetServerTools(ctx, serverName)
	if err != nil {
		// T027: Use cliError to include request_id in error output
		return cliError("failed to get server tools from daemon", err)
	}

	// Output results
	return outputTools(tools, logger)
}

// outputTools formats and displays tools based on output format using unified formatters.
func outputTools(tools []map[string]interface{}, _ *zap.Logger) error {
	outputFormat := ResolveOutputFormat()
	formatter, err := GetOutputFormatter()
	if err != nil {
		return output.NewStructuredError(output.ErrCodeInvalidOutputFormat, err.Error()).
			WithGuidance("Use -o table, -o json, or -o yaml")
	}

	// For JSON/YAML, format directly
	if outputFormat == "json" || outputFormat == "yaml" {
		result, fmtErr := formatter.Format(tools)
		if fmtErr != nil {
			return fmt.Errorf("failed to format output: %w", fmtErr)
		}
		fmt.Println(result)
		return nil
	}

	// Table format: show name and description
	headers := []string{"NAME", "DESCRIPTION"}
	var rows [][]string
	for _, tool := range tools {
		name, _ := tool["name"].(string)
		desc, _ := tool["description"].(string)
		rows = append(rows, []string{name, desc})
	}

	result, fmtErr := formatter.FormatTable(headers, rows)
	if fmtErr != nil {
		return fmt.Errorf("failed to format table: %w", fmtErr)
	}
	fmt.Print(result)
	return nil
}

// runToolsListStandalone executes tools list in standalone mode (original behavior).
func runToolsListStandalone(ctx context.Context, serverName string, globalConfig *config.Config, logger *zap.Logger) error {
	// Find server config
	var serverConfig *config.ServerConfig
	for _, server := range globalConfig.Servers {
		if server.Name == serverName {
			serverConfig = server
			break
		}
	}
	if serverConfig == nil {
		return fmt.Errorf("server '%s' not found in configuration. Available servers: %v",
			serverName, getAvailableServerNames(globalConfig))
	}

	fmt.Printf("🚀 MCP Tools List - Server: %s\n", serverName)
	fmt.Printf("📝 Log Level: %s\n", toolsLogLevel)
	fmt.Printf("⏱️  Timeout: %v\n", timeout)
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	// Create storage (optional, for OAuth persistence)
	var db *storage.BoltDB
	if globalConfig.DataDir != "" {
		boltDB, err := storage.NewBoltDB(globalConfig.DataDir, logger.Sugar())
		if err != nil {
			logger.Warn("Failed to create storage, OAuth will use in-memory")
		} else {
			db = boltDB
			defer db.Close()
		}
	}

	// Create secret resolver
	secretResolver := secret.NewResolver()

	// Create log config for managed client
	logConfig := &config.LogConfig{
		Level:         toolsLogLevel,
		EnableConsole: true,
		EnableFile:    false,
		JSONFormat:    false,
	}

	// Create managed client (same as serve mode!)
	managedClient, err := managed.NewClient(serverName, serverConfig, logger, logConfig, globalConfig, db, secretResolver)
	if err != nil {
		return fmt.Errorf("failed to create managed client: %w", err)
	}

	// Connect to server
	fmt.Printf("🔗 Connecting to server '%s'...\n", serverName)
	if err := managedClient.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect to server '%s': %w", serverName, err)
	}

	// Ensure cleanup on exit
	defer func() {
		fmt.Printf("🔌 Disconnecting from server...\n")
		if disconnectErr := managedClient.Disconnect(); disconnectErr != nil {
			fmt.Fprintf(os.Stderr, "⚠️  Warning: Failed to disconnect cleanly: %v\n", disconnectErr)
		}
	}()

	// List tools
	tools, err := managedClient.ListTools(ctx)
	if err != nil {
		return fmt.Errorf("failed to list tools: %w", err)
	}

	// Output results using unified formatter
	if len(tools) == 0 {
		outputFormat := ResolveOutputFormat()
		if outputFormat == "table" {
			fmt.Printf("⚠️  No tools found on server '%s'\n", serverName)
			fmt.Printf("💡 This could indicate:\n")
			fmt.Printf("   • Server doesn't support tools\n")
			fmt.Printf("   • Server is not properly configured\n")
			fmt.Printf("   • Connection issues during tool discovery\n")
			return nil
		}
		// For JSON/YAML, output empty array
	}

	return outputToolsFromMetadata(tools, serverName)
}

// runToolsPrefsList lists tool preferences for a server
func runToolsPrefsList(_ *cobra.Command, _ []string) error {
	if serverName == "" {
		return fmt.Errorf("--server flag is required")
	}

	// Get socket path for daemon communication
	dataDir := getDefaultDataDir()
	socketPath := socket.DetectSocketPath(dataDir)

	if !socket.IsSocketAvailable(socketPath) {
		return fmt.Errorf("daemon not running at %s. Start with 'mcpproxy serve' first", socketPath)
	}

	client := cliclient.NewClient(socketPath, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	prefs, err := client.GetToolPreferences(ctx, serverName)
	if err != nil {
		return fmt.Errorf("failed to get tool preferences: %w", err)
	}

	outputFormat := ResolveOutputFormat()
	formatter, err := GetOutputFormatter()
	if err != nil {
		return err
	}

	if outputFormat == "table" {
		headers := []string{"TOOL", "ENABLED", "CUSTOM NAME", "CUSTOM DESCRIPTION"}
		var rows [][]string
		for toolName, pref := range prefs {
			enabled := "✓"
			if !pref.Enabled {
				enabled = "✗"
			}
			customName := pref.CustomName
			if customName == "" {
				customName = "-"
			}
			customDesc := pref.CustomDescription
			if customDesc == "" {
				customDesc = "-"
			}
			rows = append(rows, []string{toolName, enabled, customName, customDesc})
		}

		result, err := formatter.FormatTable(headers, rows)
		if err != nil {
			return fmt.Errorf("failed to format table: %w", err)
		}
		fmt.Print(result)
	} else {
		// JSON/YAML format
		result, err := formatter.Format(prefs)
		if err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}
		fmt.Println(result)
	}

	return nil
}

// runToolsPrefsEnable enables a tool
func runToolsPrefsEnable(_ *cobra.Command, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: mcpproxy tools preferences enable <server> <tool>")
	}
	server := args[0]
	tool := args[1]

	return updateToolPreference(server, tool, true)
}

// runToolsPrefsDisable disables a tool
func runToolsPrefsDisable(_ *cobra.Command, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: mcpproxy tools preferences disable <server> <tool>")
	}
	server := args[0]
	tool := args[1]

	return updateToolPreference(server, tool, false)
}

// runToolsPrefsToggle toggles tool enable/disable state
func runToolsPrefsToggle(_ *cobra.Command, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: mcpproxy tools preferences toggle <server> <tool>")
	}
	server := args[0]
	tool := args[1]

	// Get current preferences
	dataDir := getDefaultDataDir()
	socketPath := socket.DetectSocketPath(dataDir)

	if !socket.IsSocketAvailable(socketPath) {
		return fmt.Errorf("daemon not running. Start with 'mcpproxy serve' first")
	}

	client := cliclient.NewClient(socketPath, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	prefs, err := client.GetToolPreferences(ctx, server)
	if err != nil {
		return fmt.Errorf("failed to get current preferences: %w", err)
	}

	// Toggle the state
	currentEnabled := true
	if pref, ok := prefs[tool]; ok {
		currentEnabled = pref.Enabled
	}

	return updateToolPreference(server, tool, !currentEnabled)
}

// updateToolPreference updates a tool preference via HTTP API
func updateToolPreference(server, tool string, enabled bool) error {
	dataDir := getDefaultDataDir()
	socketPath := socket.DetectSocketPath(dataDir)

	if !socket.IsSocketAvailable(socketPath) {
		return fmt.Errorf("daemon not running. Start with 'mcpproxy serve' first")
	}

	// Build API URL
	apiURL := buildSocketURL(socketPath, fmt.Sprintf("/api/v1/servers/%s/tools/preferences/%s", server, tool))

	// Build request body
	body := map[string]interface{}{
		"enabled": enabled,
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make PUT request
	req, err := http.NewRequest("PUT", apiURL, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error (%s): %s", resp.Status, string(body))
	}

	action := "enabled"
	if !enabled {
		action = "disabled"
	}
	fmt.Printf("✅ Tool '%s' %s on server '%s'\n", tool, action, server)
	return nil
}

// getDefaultDataDir returns the default data directory
func getDefaultDataDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, ".mcpproxy")
}

// buildSocketURL builds a URL for socket communication
func buildSocketURL(socketPath, path string) string {
	// For Windows named pipes
	if len(socketPath) > 8 && socketPath[:8] == "npipe://" {
		return socketPath + path
	}
	// For Unix sockets
	if len(socketPath) > 7 && socketPath[:7] == "unix://" {
		return socketPath + path
	}
	return socketPath + path
}

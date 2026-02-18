package cli

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/davidl71/exarp-go/internal/config"
	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/factory"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/internal/logging"
	"github.com/davidl71/exarp-go/internal/prompts"
	"github.com/davidl71/exarp-go/internal/resources"
	"github.com/davidl71/exarp-go/internal/tools"
	mcpcli "github.com/davidl71/mcp-go-core/pkg/mcp/cli"
	mcplog "github.com/davidl71/mcp-go-core/pkg/mcp/logging"
)

// EnsureConfigAndDatabase loads centralized config (if available), sets global config,
// and initializes the database. Uses centralized config when available, falls back to legacy config.
// Returns error only if project root cannot be found; database init failures are logged but not returned.
// Call this at startup before any config getters or database operations.
func EnsureConfigAndDatabase(projectRoot string) {
	// Try to use centralized config first
	fullCfg, err := config.LoadConfig(projectRoot)
	if err == nil {
		config.SetGlobalConfig(fullCfg)
		logging.ConfigureFromConfig(fullCfg.Logging)
		// Convert centralized config DatabaseConfig to DatabaseConfigFields
		dbCfg := database.DatabaseConfigFields{
			SQLitePath:          fullCfg.Database.SQLitePath,
			JSONFallbackPath:    fullCfg.Database.JSONFallbackPath,
			BackupPath:          fullCfg.Database.BackupPath,
			MaxConnections:      fullCfg.Database.MaxConnections,
			ConnectionTimeout:   int64(fullCfg.Database.ConnectionTimeout.Seconds()),
			QueryTimeout:        int64(fullCfg.Database.QueryTimeout.Seconds()),
			RetryAttempts:       fullCfg.Database.RetryAttempts,
			RetryInitialDelay:   int64(fullCfg.Database.RetryInitialDelay.Seconds()),
			RetryMaxDelay:       int64(fullCfg.Database.RetryMaxDelay.Seconds()),
			RetryMultiplier:     fullCfg.Database.RetryMultiplier,
			AutoVacuum:          fullCfg.Database.AutoVacuum,
			WALMode:             fullCfg.Database.WALMode,
			CheckpointInterval:  fullCfg.Database.CheckpointInterval,
			BackupRetentionDays: fullCfg.Database.BackupRetentionDays,
		}

		if err := database.InitWithCentralizedConfig(projectRoot, dbCfg); err != nil {
			logWarn(context.Background(), "Database initialization with centralized config failed", "error", err, "operation", "EnsureConfigAndDatabase", "fallback", "legacy config")
			// Fall through to legacy init
		} else {
			logDebug(context.Background(), "Database initialized with centralized config", "path", fullCfg.Database.SQLitePath, "operation", "EnsureConfigAndDatabase")
			return
		}
	}

	// Fall back to legacy config (SetGlobalConfig already set defaults via GetGlobalConfig)
	if err := database.Init(projectRoot); err != nil {
		logWarn(context.Background(), "Database initialization failed", "error", err, "operation", "EnsureConfigAndDatabase", "fallback", "JSON")
		return
	}

	logDebug(context.Background(), "Database initialized", "path", projectRoot+"/.todo2/todo2.db", "operation", "EnsureConfigAndDatabase")
}

// initializeDatabase initializes the database if project root is found
// Note: Database remains open for the duration of CLI execution
// It will be closed when the process exits or explicitly closed
// Uses centralized config if available, falls back to legacy config.
func initializeDatabase() {
	projectRoot, err := tools.FindProjectRoot()
	if err != nil {
		logWarn(context.Background(), "Could not find project root", "error", err, "operation", "initializeDatabase", "fallback", "JSON")
		return
	}

	EnsureConfigAndDatabase(projectRoot)
}

// setupServer creates and configures the MCP server.
func setupServer() (framework.MCPServer, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	server, err := factory.NewServerFromConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create server: %w", err)
	}

	if err := tools.RegisterAllTools(server); err != nil {
		return nil, fmt.Errorf("failed to register tools: %w", err)
	}

	if err := prompts.RegisterAllPrompts(server); err != nil {
		return nil, fmt.Errorf("failed to register prompts: %w", err)
	}

	if err := resources.RegisterAllResources(server); err != nil {
		return nil, fmt.Errorf("failed to register resources: %w", err)
	}

	return server, nil
}

// Run starts the CLI interface.
// Uses mcp-go-core ParseArgs for structured CLI dispatch; subcommands (config, task, tui, tui3270) keep os.Args-based remainder.
func Run() error {
	parsed := mcpcli.ParseArgs(os.Args[1:])

	// Subcommand dispatch (config, task, tui, tui3270)
	switch parsed.Command {
	case "config":
		return handleConfigCommand(parsed)
	case "task":
		initializeDatabase()

		server, err := setupServer()
		if err != nil {
			return err
		}

		return handleTaskCommand(server, parsed)
	case "tui":
		initializeDatabase()

		server, err := setupServer()
		if err != nil {
			return err
		}

		status := parsed.Subcommand
		if status == "" && len(parsed.Positional) > 0 {
			status = parsed.Positional[0]
		}

		return RunTUI(server, status)
	case "lock":
		initializeDatabase()
		return handleLockCommand(parsed)
	case "session":
		initializeDatabase()

		server, err := setupServer()
		if err != nil {
			return err
		}

		return handleSessionCommand(server, parsed)
	case "cursor":
		initializeDatabase()
		return handleCursorCommand(parsed)
	case "tui3270":
		tuiParsed := mcpcli.ParseArgs(os.Args[2:])
		daemon := tuiParsed.GetBoolFlag("daemon", false) || tuiParsed.GetBoolFlag("d", false)

		pidFile := tuiParsed.GetFlag("pid-file", "")
		if pidFile == "" {
			pidFile = tuiParsed.GetFlag("pidfile", "")
		}

		status := ""
		port := 3270

		for _, p := range append([]string{tuiParsed.Subcommand}, tuiParsed.Positional...) {
			if p == "" {
				continue
			}

			if parsedPort, err := strconv.Atoi(p); err == nil && parsedPort > 0 {
				port = parsedPort
			} else if status == "" {
				status = p
			}
		}

		initializeDatabase()

		server, err := setupServer()
		if err != nil {
			return err
		}

		return RunTUI3270(server, status, port, daemon, pidFile)
	}

	// Flag-based modes (-tool, -list, -test, -i, -completion); use flag package (ParseArgs doesn't handle -flag value skip)
	var (
		toolName    = flag.String("tool", "", "Tool name to execute")
		argsJSON    = flag.String("args", "{}", "Tool arguments as JSON")
		listTools   = flag.Bool("list", false, "List all available tools")
		testTool    = flag.String("test", "", "Test a tool with example arguments")
		interactive = flag.Bool("i", false, "Interactive mode")
		completion  = flag.String("completion", "", "Generate shell completion script (bash|zsh|fish)")
	)

	_ = flag.CommandLine.Parse(os.Args[1:])

	initializeDatabase()

	defer func() {
		if err := database.Close(); err != nil {
			logWarn(context.Background(), "Error closing database", "error", err, "operation", "closeDatabase")
		}
	}()

	server, err := setupServer()
	if err != nil {
		return err
	}

	switch {
	case *completion != "":
		return generateCompletion(server, *completion)
	case *listTools:
		return listAllTools(server)
	case *testTool != "":
		return testToolExecution(server, *testTool)
	case *toolName != "":
		return executeTool(server, *toolName, *argsJSON)
	case *interactive:
		return runInteractive(server)
	default:
		showUsage()
		return nil
	}
}

// IsTTY checks if stdin is a terminal.
// Uses mcp-go-core CLI utilities for shared implementation.
func IsTTY() bool {
	return mcpcli.IsTTY()
}

// IsTTYStdout checks if stdout is a terminal.
// Use for colored output detection: enable colors only when stdout is a TTY.
// Uses mcp-go-core IsTTYFile(os.Stdout).
func IsTTYStdout() bool {
	return mcpcli.IsTTYFile(os.Stdout)
}

// ColorEnabled returns true when colored CLI output is appropriate (stdout is a TTY).
// Callers may use this to enable ANSI color codes for list, tool output, etc.
func ColorEnabled() bool {
	return IsTTYStdout()
}

// ExecutionMode is the detected run mode (cli vs mcp). Re-exported from mcp-go-core.
type ExecutionMode = mcpcli.ExecutionMode

const (
	ModeCLI ExecutionMode = mcpcli.ModeCLI
	ModeMCP ExecutionMode = mcpcli.ModeMCP
)

// DetectMode detects execution mode from TTY: ModeCLI if stdin is a terminal, else ModeMCP.
// Uses mcp-go-core DetectMode. Prefer HasCLIFlags for explicit CLI args (e.g. -tool, task).
func DetectMode() ExecutionMode {
	return mcpcli.DetectMode()
}

func colorizeToolName(name string) string {
	if !ColorEnabled() {
		return name
	}

	const green, reset = "\033[32m", "\033[0m"

	return green + name + reset
}

// listAllTools lists all available tools.
func listAllTools(server framework.MCPServer) error {
	toolList := server.ListTools()
	if len(toolList) == 0 {
		_, _ = fmt.Println("No tools registered")
		return nil
	}

	_, _ = fmt.Printf("Available tools (%d total):\n\n", len(toolList))
	for _, tool := range toolList {
		_, _ = fmt.Printf("  %s\n", colorizeToolName(tool.Name))

		if tool.Description != "" {
			// Truncate long descriptions
			desc := tool.Description
			if len(desc) > 80 {
				desc = desc[:77] + "..."
			}

			_, _ = fmt.Printf("    %s\n", desc)
		}

		_, _ = fmt.Println()
	}

	return nil
}

// executeTool executes a tool with the given arguments.
func executeTool(server framework.MCPServer, toolName, argsJSON string) error {
	ctx := context.Background()

	// Add operation context for logging
	ctx = mcplog.WithOperation(ctx, "executeTool")
	ctx = mcplog.WithRequestID(ctx, generateRequestID())

	// Parse JSON arguments
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		logError(ctx, "Failed to parse JSON arguments", "error", err, "tool", toolName)
		return fmt.Errorf("failed to parse JSON arguments: %w", err)
	}

	// Convert to json.RawMessage for tool handler
	argsBytes, err := json.Marshal(args)
	if err != nil {
		logError(ctx, "Failed to marshal arguments", "error", err, "tool", toolName)
		return fmt.Errorf("failed to marshal arguments: %w", err)
	}

	// Execute tool via server
	_, _ = fmt.Printf("Executing tool: %s\n", toolName)
	_, _ = fmt.Printf("Arguments: %s\n\n", argsJSON)

	// Use context-aware logger to include request_id and operation
	logDebug(ctx, "Executing tool", "tool", toolName)

	// Track performance
	perf := StartPerformanceLogging(ctx, "tool_execution", DefaultSlowThreshold)
	defer perf.Finish()

	result, err := server.CallTool(ctx, toolName, argsBytes)
	if err != nil {
		perf.FinishWithError(err)
		logError(ctx, "Tool execution failed", "error", err, "tool", toolName)

		return fmt.Errorf("tool execution failed: %w", err)
	}

	// Display results
	if len(result) == 0 {
		_, _ = fmt.Println("Tool executed successfully (no output)")
		return nil
	}

	_, _ = fmt.Println("Result:")

	for i, content := range result {
		if len(result) > 1 {
			_, _ = fmt.Printf("\n[%d] ", i+1)
		}

		_, _ = fmt.Println(content.Text)
	}

	return nil
}

// runSessionHandoffList calls the session tool with handoff list and prints the result.
func runSessionHandoffList(server framework.MCPServer) error {
	ctx := context.Background()
	args := map[string]interface{}{"action": "handoff", "sub_action": "list"}

	argsBytes, err := json.Marshal(args)
	if err != nil {
		return err
	}

	result, err := server.CallTool(ctx, "session", argsBytes)
	if err != nil {
		return err
	}

	if len(result) == 0 {
		_, _ = fmt.Println("No handoff notes.")
		return nil
	}

	for _, content := range result {
		_, _ = fmt.Println(content.Text)
	}

	return nil
}

// handleSessionCommand handles the "session" subcommand (e.g. session handoffs).
func handleSessionCommand(server framework.MCPServer, parsed *mcpcli.Args) error {
	sub := strings.ToLower(strings.TrimSpace(parsed.Subcommand))
	if sub == "" && len(parsed.Positional) > 0 {
		sub = strings.ToLower(strings.TrimSpace(parsed.Positional[0]))
	}

	switch sub {
	case "handoffs", "list":
		return runSessionHandoffList(server)
	default:
		_, _ = fmt.Println("Session subcommands: handoffs")
		_, _ = fmt.Println("  exarp-go session handoffs   View session handoff notes")

		return nil
	}
}

// generateRequestID generates a simple request ID for logging.
func generateRequestID() string {
	return fmt.Sprintf("req-%d", time.Now().UnixNano())
}

// findToolByName finds a tool by name in the server's tool list.
func findToolByName(server framework.MCPServer, toolName string) (*framework.ToolInfo, error) {
	toolList := server.ListTools()
	for _, t := range toolList {
		if t.Name == toolName {
			return &t, nil
		}
	}

	return nil, fmt.Errorf("tool %q not found", toolName)
}

// displayToolInfo displays information about a tool.
func displayToolInfo(toolInfo *framework.ToolInfo) {
	_, _ = fmt.Printf("\nTool: %s\n", toolInfo.Name)
	if toolInfo.Description != "" {
		_, _ = fmt.Printf("Description: %s\n", toolInfo.Description)
	}
}

// displayExampleArgs displays example arguments for a tool.
func displayExampleArgs(args map[string]interface{}) error {
	exampleJSON, err := json.MarshalIndent(args, "  ", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal example arguments: %w", err)
	}

	_, _ = fmt.Println("\nExample Arguments:")
	_, _ = fmt.Printf("  %s\n", string(exampleJSON))

	return nil
}

// displayTestResults displays the results of a tool test.
func displayTestResults(result []framework.TextContent) {
	_, _ = fmt.Println("✅ Test passed!")

	if len(result) == 0 {
		return
	}

	_, _ = fmt.Println("\nOutput:")

	for _, content := range result {
		output := content.Text
		if len(output) > 500 {
			output = output[:497] + "..."
		}

		_, _ = fmt.Printf("  %s\n", output)
	}
}

// testToolExecution tests a tool with example arguments (for feature testing).
func testToolExecution(server framework.MCPServer, toolName string) error {
	_, _ = fmt.Printf("Feature Testing: %s\n", toolName)
	_, _ = fmt.Println("=" + strings.Repeat("=", len(toolName)+18))

	toolInfo, err := findToolByName(server, toolName)
	if err != nil {
		return err
	}

	displayToolInfo(toolInfo)

	exampleArgs := generateExampleArgs(toolInfo.Schema)
	if displayErr := displayExampleArgs(exampleArgs); displayErr != nil {
		return displayErr
	}

	_, _ = fmt.Println("\nExecuting with example arguments...")
	_, _ = fmt.Println()

	argsBytes, err := json.Marshal(exampleArgs)
	if err != nil {
		return fmt.Errorf("failed to marshal arguments: %w", err)
	}

	result, err := server.CallTool(context.Background(), toolName, argsBytes)
	if err != nil {
		_, _ = fmt.Printf("❌ Test failed: %v\n", err)
		return fmt.Errorf("tool execution failed: %w", err)
	}

	displayTestResults(result)

	return nil
}

// getExampleValueForType generates an example value for a given property type.
func getExampleValueForType(propType string, propMap map[string]interface{}) interface{} {
	switch propType {
	case "string":
		if enum, ok := propMap["enum"].([]interface{}); ok && len(enum) > 0 {
			return enum[0]
		}

		return "example"
	case "boolean":
		return false
	case "number", "integer":
		return 0
	case "array":
		return []interface{}{}
	case "object":
		return map[string]interface{}{}
	default:
		return nil
	}
}

// generateExampleArgs generates example arguments from a tool schema.
func generateExampleArgs(schema framework.ToolSchema) map[string]interface{} {
	args := make(map[string]interface{})
	if schema.Properties == nil {
		return args
	}

	for name, prop := range schema.Properties {
		propMap, isMap := prop.(map[string]interface{})
		if !isMap {
			continue
		}

		propType, hasType := propMap["type"].(string)
		if !hasType {
			continue
		}

		args[name] = getExampleValueForType(propType, propMap)
	}

	return args
}

// handleInteractiveCommand processes a single command in interactive mode.
func handleInteractiveCommand(server framework.MCPServer, cmd string) bool {
	switch cmd {
	case "exit", "quit":
		return false
	case "help":
		showHelp()
	case "list":
		if err := listAllTools(server); err != nil {
			_, _ = fmt.Printf("Error listing tools: %v\n", err)
		}
	case "handoffs":
		if err := runSessionHandoffList(server); err != nil {
			_, _ = fmt.Printf("Error listing handoffs: %v\n", err)
		}
	default:
		_, _ = fmt.Printf("Unknown command: %s\n", cmd)
		_, _ = fmt.Println("Type 'help' for available commands")
	}

	return true
}

// readInteractiveInput reads a line of input from the user.
func readInteractiveInput() (string, error) {
	_, _ = fmt.Print("exarp-go> ")

	var input string

	_, err := fmt.Scanln(&input)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(input), nil
}

// parseCommand extracts the command from user input.
func parseCommand(input string) string {
	if input == "" {
		return ""
	}

	parts := strings.Fields(input)
	if len(parts) == 0 {
		return ""
	}

	return parts[0]
}

// runInteractive starts an interactive CLI session.
func runInteractive(server framework.MCPServer) error {
	_, _ = fmt.Println("Interactive CLI Mode")
	_, _ = fmt.Println("===================")
	_, _ = fmt.Println("Type 'help' for commands, 'exit' to quit")
	_, _ = fmt.Println()

	for {
		input, err := readInteractiveInput()
		if err != nil {
			return nil
		}

		cmd := parseCommand(input)
		if cmd == "" {
			continue
		}

		if !handleInteractiveCommand(server, cmd) {
			return nil
		}
	}
}

// showUsage displays usage information.
func showUsage() {
	_, _ = fmt.Println("exarp-go - Command Line Interface")
	_, _ = fmt.Println()
	_, _ = fmt.Println("Usage:")
	_, _ = fmt.Println("  exarp-go [flags]")
	_, _ = fmt.Println("  exarp-go task <command> [options]")
	_, _ = fmt.Println("  exarp-go tui [status]")
	_, _ = fmt.Println("  exarp-go tui3270 [status] [port]")
	_, _ = fmt.Println("  exarp-go session handoffs")
	_, _ = fmt.Println()
	_, _ = fmt.Println("Flags:")
	_, _ = fmt.Println("  -tool <name>        Execute a tool")
	_, _ = fmt.Println("  -args <json>        Tool arguments as JSON (default: {})")
	_, _ = fmt.Println("  -list               List all available tools")
	_, _ = fmt.Println("  -test <name>        Test a tool with example arguments")
	_, _ = fmt.Println("  -i                  Interactive mode")
	_, _ = fmt.Println("  -completion <shell> Generate shell completion script (bash|zsh|fish)")
	_, _ = fmt.Println()
	_, _ = fmt.Println("Task Commands:")
	_, _ = fmt.Println("  task list [--status <status>] [--priority <priority>] [--tag <tag>]")
	_, _ = fmt.Println("  task status <task-id>")
	_, _ = fmt.Println("  task update <task-id> [--new-status <status>] [--new-priority <priority>]")
	_, _ = fmt.Println("  task create <name> [--description <text>] [--priority <priority>]")
	_, _ = fmt.Println("  task show <task-id>")
	_, _ = fmt.Println("  task help")
	_, _ = fmt.Println()
	_, _ = fmt.Println("TUI Commands:")
	_, _ = fmt.Println("  tui                 Launch terminal user interface for task management")
	_, _ = fmt.Println("  tui <status>        Launch TUI filtered by status (e.g., 'Todo', 'In Progress')")
	_, _ = fmt.Println("  tui3270 [status] [port] [--daemon] [--pid-file FILE]  Launch 3270 TUI server")
	_, _ = fmt.Println("    --daemon, -d          Run in background mode (writes PID file, redirects logs)")
	_, _ = fmt.Println("    --pid-file FILE       Write PID to FILE (default: .exarp-go-tui3270.pid)")
	_, _ = fmt.Println()
	_, _ = fmt.Println("Session:")
	_, _ = fmt.Println("  session handoffs       View session handoff notes (from previous sessions)")
	_, _ = fmt.Println()
	_, _ = fmt.Println("Cursor integration:")
	_, _ = fmt.Println("  cursor run [task-id]       Run Cursor agent with task context (e.g. cursor run T-123)")
	_, _ = fmt.Println("  cursor run -p \"prompt\"     Run Cursor agent with custom prompt")
	_, _ = fmt.Println("  cursor run --no-interactive  Non-interactive (agent -p); use with -p or task-id")
	_, _ = fmt.Println("  cursor run --mode plan|ask|agent  Cursor agent mode (default: plan)")
	_, _ = fmt.Println("  Run from project root so exarp-go MCP tools are available.")
	_, _ = fmt.Println("  Install Cursor CLI: curl https://cursor.com/install -fsS | bash")
	_, _ = fmt.Println("  Set CURSOR_API_KEY for non-interactive/CI use. See docs/CURSOR_API_AND_CLI_INTEGRATION.md")
	_, _ = fmt.Println()
	_, _ = fmt.Println("Examples:")
	_, _ = fmt.Println("  exarp-go -list")
	_, _ = fmt.Println("  exarp-go -tool lint -args '{\"action\":\"run\",\"path\":\".\"}'")
	_, _ = fmt.Println("  exarp-go -test lint")
	_, _ = fmt.Println("  exarp-go -i")
	_, _ = fmt.Println("  exarp-go tui")
	_, _ = fmt.Println("  exarp-go tui \"In Progress\"")
	_, _ = fmt.Println("  exarp-go task list --status \"In Progress\"")
	_, _ = fmt.Println("  exarp-go task update T-1 --new-status \"Done\"")
	_, _ = fmt.Println("  exarp-go task update T-1 --new-priority high")
	_, _ = fmt.Println("  exarp-go session handoffs")
	_, _ = fmt.Println("  exarp-go cursor run T-123")
	_, _ = fmt.Println("  exarp-go cursor run -p \"Implement feature X\" --no-interactive")
	_, _ = fmt.Println("  exarp-go -completion bash > /usr/local/etc/bash_completion.d/exarp-go")
	_, _ = fmt.Println()
}

// showHelp displays help information.
func showHelp() {
	_, _ = fmt.Println("Available commands:")
	_, _ = fmt.Println("  help              Show this help message")
	_, _ = fmt.Println("  list              List all available tools")
	_, _ = fmt.Println("  handoffs          View session handoff notes")
	_, _ = fmt.Println("  exit, quit        Exit interactive mode")
	_, _ = fmt.Println()
}

// generateCompletion generates shell completion scripts.
func generateCompletion(server framework.MCPServer, shell string) error {
	toolList := server.ListTools()
	toolNames := make([]string, 0, len(toolList))

	for _, tool := range toolList {
		toolNames = append(toolNames, tool.Name)
	}

	switch shell {
	case "bash":
		return generateBashCompletion(toolNames)
	case "zsh":
		return generateZshCompletion(toolNames)
	case "fish":
		return generateFishCompletion(toolNames)
	default:
		return fmt.Errorf("unsupported shell: %s (supported: bash, zsh, fish)", shell)
	}
}

// generateBashCompletion generates bash completion script.
func generateBashCompletion(toolNames []string) error {
	_, _ = fmt.Println("# exarp-go bash completion")
	_, _ = fmt.Println("_exarp_go() {")
	_, _ = fmt.Println("    local cur prev opts")
	_, _ = fmt.Println("    COMPREPLY=()")
	_, _ = fmt.Println("    cur=\"${COMP_WORDS[COMP_CWORD]}\"")
	_, _ = fmt.Println("    prev=\"${COMP_WORDS[COMP_CWORD-1]}\"")
	_, _ = fmt.Println("    opts=\"-tool -args -list -test -i -completion\"")
	_, _ = fmt.Println()
	_, _ = fmt.Println("    case \"${prev}\" in")
	_, _ = fmt.Println("        -tool|--tool)")
	_, _ = fmt.Printf("            COMPREPLY=($(compgen -W \"%s\" -- \"${cur}\"))\n", strings.Join(toolNames, " "))
	_, _ = fmt.Println("            return 0")
	_, _ = fmt.Println("            ;;")
	_, _ = fmt.Println("        -test|--test)")
	_, _ = fmt.Printf("            COMPREPLY=($(compgen -W \"%s\" -- \"${cur}\"))\n", strings.Join(toolNames, " "))
	_, _ = fmt.Println("            return 0")
	_, _ = fmt.Println("            ;;")
	_, _ = fmt.Println("        -completion|--completion)")
	_, _ = fmt.Println("            COMPREPLY=($(compgen -W \"bash zsh fish\" -- \"${cur}\"))")
	_, _ = fmt.Println("            return 0")
	_, _ = fmt.Println("            ;;")
	_, _ = fmt.Println("        *)")
	_, _ = fmt.Println("            if [[ ${cur} == -* ]] ; then")
	_, _ = fmt.Println("                COMPREPLY=($(compgen -W \"${opts}\" -- \"${cur}\"))")
	_, _ = fmt.Println("            fi")
	_, _ = fmt.Println("            ;;")
	_, _ = fmt.Println("    esac")
	_, _ = fmt.Println("}")
	_, _ = fmt.Println("complete -F _exarp_go exarp-go")

	return nil
}

// generateZshCompletion generates zsh completion script.
func generateZshCompletion(toolNames []string) error {
	_, _ = fmt.Println("#compdef exarp-go")
	_, _ = fmt.Println()
	_, _ = fmt.Println("_exarp_go() {")
	_, _ = fmt.Println("    local context state line")
	_, _ = fmt.Println("    local -a tools")
	_, _ = fmt.Printf("    tools=(%s)\n", strings.Join(toolNames, " "))
	_, _ = fmt.Println()
	_, _ = fmt.Println("    _arguments \\")
	_, _ = fmt.Println("        '(-tool --tool)-tool[Tool name to execute]:tool:->tools' \\")
	_, _ = fmt.Println("        '(-tool --tool)--tool[Tool name to execute]:tool:->tools' \\")
	_, _ = fmt.Println("        '(-args --args)-args[Tool arguments as JSON]:args:' \\")
	_, _ = fmt.Println("        '(-args --args)--args[Tool arguments as JSON]:args:' \\")
	_, _ = fmt.Println("        '(-list --list)-list[List all available tools]' \\")
	_, _ = fmt.Println("        '(-list --list)--list[List all available tools]' \\")
	_, _ = fmt.Println("        '(-test --test)-test[Test a tool with example arguments]:test:->tools' \\")
	_, _ = fmt.Println("        '(-test --test)--test[Test a tool with example arguments]:test:->tools' \\")
	_, _ = fmt.Println("        '(-i --interactive)-i[Interactive mode]' \\")
	_, _ = fmt.Println("        '(-i --interactive)--interactive[Interactive mode]' \\")
	_, _ = fmt.Println("        '(-completion --completion)-completion[Generate shell completion script]:completion:(bash zsh fish)' \\")
	_, _ = fmt.Println("        '(-completion --completion)--completion[Generate shell completion script]:completion:(bash zsh fish)'")
	_, _ = fmt.Println()
	_, _ = fmt.Println("    case $state in")
	_, _ = fmt.Println("        tools)")
	_, _ = fmt.Println("            _describe 'tools' tools")
	_, _ = fmt.Println("            ;;")
	_, _ = fmt.Println("    esac")
	_, _ = fmt.Println("}")
	_, _ = fmt.Println()
	_, _ = fmt.Println("_exarp_go \"$@\"")

	return nil
}

// generateFishCompletion generates fish completion script.
func generateFishCompletion(toolNames []string) error {
	_, _ = fmt.Println("# exarp-go fish completion")
	_, _ = fmt.Println()
	_, _ = fmt.Printf("complete -c exarp-go -n '__fish_use_subcommand' -s t -l tool -d 'Tool name to execute' -xa '%s'\n", strings.Join(toolNames, " "))
	_, _ = fmt.Println("complete -c exarp-go -n '__fish_use_subcommand' -l args -d 'Tool arguments as JSON'")
	_, _ = fmt.Println("complete -c exarp-go -n '__fish_use_subcommand' -s l -l list -d 'List all available tools'")
	_, _ = fmt.Printf("complete -c exarp-go -n '__fish_use_subcommand' -l test -d 'Test a tool with example arguments' -xa '%s'\n", strings.Join(toolNames, " "))
	_, _ = fmt.Println("complete -c exarp-go -n '__fish_use_subcommand' -s i -l interactive -d 'Interactive mode'")
	_, _ = fmt.Println("complete -c exarp-go -n '__fish_use_subcommand' -l completion -d 'Generate shell completion script' -xa 'bash zsh fish'")

	return nil
}

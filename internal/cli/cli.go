package cli

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/davidl71/exarp-go/internal/config"
	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/factory"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/internal/prompts"
	"github.com/davidl71/exarp-go/internal/resources"
	"github.com/davidl71/exarp-go/internal/tools"
	"golang.org/x/term"
)

// Run starts the CLI interface
func Run() error {
	// Parse command line arguments
	var (
		toolName    = flag.String("tool", "", "Tool name to execute")
		argsJSON    = flag.String("args", "{}", "Tool arguments as JSON")
		listTools   = flag.Bool("list", false, "List all available tools")
		testTool    = flag.String("test", "", "Test a tool with example arguments")
		interactive = flag.Bool("i", false, "Interactive mode")
		completion  = flag.String("completion", "", "Generate shell completion script (bash|zsh|fish)")
	)
	flag.Parse()

	// Initialize database (before server creation)
	projectRoot, err := tools.FindProjectRoot()
	if err != nil {
		log.Printf("Warning: Could not find project root: %v (database unavailable, will use JSON fallback)", err)
	} else {
		if err := database.Init(projectRoot); err != nil {
			log.Printf("Warning: Database initialization failed: %v (fallback to JSON)", err)
		} else {
			defer database.Close()
			log.Printf("Database initialized: %s/.todo2/todo2.db", projectRoot)
		}
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create server (we'll use it to access tools)
	server, err := factory.NewServerFromConfig(cfg)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	// Register all components
	if err := tools.RegisterAllTools(server); err != nil {
		return fmt.Errorf("failed to register tools: %w", err)
	}

	if err := prompts.RegisterAllPrompts(server); err != nil {
		return fmt.Errorf("failed to register prompts: %w", err)
	}

	if err := resources.RegisterAllResources(server); err != nil {
		return fmt.Errorf("failed to register resources: %w", err)
	}

	// Handle different CLI modes
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

// IsTTY checks if stdin is a terminal
func IsTTY() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

// listAllTools lists all available tools
func listAllTools(server framework.MCPServer) error {
	toolList := server.ListTools()
	if len(toolList) == 0 {
		_, _ = fmt.Println("No tools registered")
		return nil
	}

	_, _ = fmt.Printf("Available tools (%d total):\n\n", len(toolList))
	for _, tool := range toolList {
		_, _ = fmt.Printf("  %s\n", tool.Name)
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

// executeTool executes a tool with the given arguments
func executeTool(server framework.MCPServer, toolName, argsJSON string) error {
	ctx := context.Background()

	// Parse JSON arguments
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return fmt.Errorf("failed to parse JSON arguments: %w", err)
	}

	// Convert to json.RawMessage for tool handler
	argsBytes, err := json.Marshal(args)
	if err != nil {
		return fmt.Errorf("failed to marshal arguments: %w", err)
	}

	// Execute tool via server
	_, _ = fmt.Printf("Executing tool: %s\n", toolName)
	_, _ = fmt.Printf("Arguments: %s\n\n", argsJSON)

	result, err := server.CallTool(ctx, toolName, argsBytes)
	if err != nil {
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

// testToolExecution tests a tool with example arguments (for feature testing)
func testToolExecution(server framework.MCPServer, toolName string) error {
	_, _ = fmt.Printf("Feature Testing: %s\n", toolName)
	_, _ = fmt.Println("=" + strings.Repeat("=", len(toolName)+18))

	// Get tool info
	toolList := server.ListTools()
	var toolInfo *framework.ToolInfo
	for _, t := range toolList {
		if t.Name == toolName {
			toolInfo = &t
			break
		}
	}

	if toolInfo == nil {
		return fmt.Errorf("tool %q not found", toolName)
	}

	_, _ = fmt.Printf("\nTool: %s\n", toolInfo.Name)
	if toolInfo.Description != "" {
		_, _ = fmt.Printf("Description: %s\n", toolInfo.Description)
	}

	// Generate example arguments based on schema
	exampleArgs := generateExampleArgs(toolInfo.Schema)
	exampleJSON, err := json.MarshalIndent(exampleArgs, "  ", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal example arguments: %w", err)
	}

	_, _ = fmt.Println("\nExample Arguments:")
	_, _ = fmt.Printf("  %s\n", string(exampleJSON))

	_, _ = fmt.Println("\nExecuting with example arguments...")
	_, _ = fmt.Println()

	// Execute with example arguments
	argsBytes, err := json.Marshal(exampleArgs)
	if err != nil {
		return fmt.Errorf("failed to marshal arguments: %w", err)
	}
	result, err := server.CallTool(context.Background(), toolName, argsBytes)
	if err != nil {
		_, _ = fmt.Printf("❌ Test failed: %v\n", err)
		return err
	}

	_, _ = fmt.Println("✅ Test passed!")
	if len(result) > 0 {
		_, _ = fmt.Println("\nOutput:")
		for _, content := range result {
			// Truncate very long outputs
			output := content.Text
			if len(output) > 500 {
				output = output[:497] + "..."
			}
			_, _ = fmt.Printf("  %s\n", output)
		}
	}

	return nil
}

// generateExampleArgs generates example arguments from a tool schema
func generateExampleArgs(schema framework.ToolSchema) map[string]interface{} {
	args := make(map[string]interface{})
	if schema.Properties == nil {
		return args
	}

	for name, prop := range schema.Properties {
		propMap, ok := prop.(map[string]interface{})
		if !ok {
			continue
		}

		propType, ok := propMap["type"].(string)
		if !ok {
			continue
		}
		switch propType {
		case "string":
			if enum, ok := propMap["enum"].([]interface{}); ok && len(enum) > 0 {
				args[name] = enum[0]
			} else {
				args[name] = "example"
			}
		case "boolean":
			args[name] = false
		case "number", "integer":
			args[name] = 0
		case "array":
			args[name] = []interface{}{}
		case "object":
			args[name] = map[string]interface{}{}
		default:
			args[name] = nil
		}
	}

	return args
}

// runInteractive starts an interactive CLI session
func runInteractive(server framework.MCPServer) error {
	fmt.Println("Interactive CLI Mode")
	fmt.Println("===================")
	fmt.Println("Type 'help' for commands, 'exit' to quit")
	fmt.Println()

	// Simple interactive loop
	for {
		_, _ = fmt.Print("exarp-go> ")
		var input string
		_, err := fmt.Scanln(&input)
		if err != nil {
			// EOF or other error - exit gracefully
			return nil
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		parts := strings.Fields(input)
		if len(parts) == 0 {
			continue
		}

		cmd := parts[0]
		switch cmd {
		case "exit", "quit":
			return nil
		case "help":
			showHelp()
		case "list":
			if err := listAllTools(server); err != nil {
				_, _ = fmt.Printf("Error listing tools: %v\n", err)
			}
		default:
			_, _ = fmt.Printf("Unknown command: %s\n", cmd)
			_, _ = fmt.Println("Type 'help' for available commands")
		}
	}
}

// showUsage displays usage information
func showUsage() {
	_, _ = fmt.Println("exarp-go - Command Line Interface")
	_, _ = fmt.Println()
	_, _ = fmt.Println("Usage:")
	_, _ = fmt.Println("  exarp-go [flags]")
	_, _ = fmt.Println()
	_, _ = fmt.Println("Flags:")
	_, _ = fmt.Println("  -tool <name>        Execute a tool")
	_, _ = fmt.Println("  -args <json>        Tool arguments as JSON (default: {})")
	_, _ = fmt.Println("  -list               List all available tools")
	_, _ = fmt.Println("  -test <name>        Test a tool with example arguments")
	_, _ = fmt.Println("  -i                  Interactive mode")
	_, _ = fmt.Println("  -completion <shell> Generate shell completion script (bash|zsh|fish)")
	_, _ = fmt.Println()
	_, _ = fmt.Println("Examples:")
	_, _ = fmt.Println("  exarp-go -list")
	_, _ = fmt.Println("  exarp-go -tool lint -args '{\"action\":\"run\",\"path\":\".\"}'")
	_, _ = fmt.Println("  exarp-go -test lint")
	_, _ = fmt.Println("  exarp-go -i")
	_, _ = fmt.Println("  exarp-go -completion bash > /usr/local/etc/bash_completion.d/exarp-go")
	_, _ = fmt.Println()
}

// showHelp displays help information
func showHelp() {
	_, _ = fmt.Println("Available commands:")
	_, _ = fmt.Println("  help              Show this help message")
	_, _ = fmt.Println("  list              List all available tools")
	_, _ = fmt.Println("  exit, quit        Exit interactive mode")
	_, _ = fmt.Println()
}

// generateCompletion generates shell completion scripts
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

// generateBashCompletion generates bash completion script
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

// generateZshCompletion generates zsh completion script
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

// generateFishCompletion generates fish completion script
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

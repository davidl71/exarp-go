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
	tools := server.ListTools()
	if len(tools) == 0 {
		fmt.Println("No tools registered")
		return nil
	}

	fmt.Printf("Available tools (%d total):\n\n", len(tools))
	for _, tool := range tools {
		fmt.Printf("  %s\n", tool.Name)
		if tool.Description != "" {
			// Truncate long descriptions
			desc := tool.Description
			if len(desc) > 80 {
				desc = desc[:77] + "..."
			}
			fmt.Printf("    %s\n", desc)
		}
		fmt.Println()
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
	fmt.Printf("Executing tool: %s\n", toolName)
	fmt.Printf("Arguments: %s\n\n", argsJSON)

	result, err := server.CallTool(ctx, toolName, argsBytes)
	if err != nil {
		return fmt.Errorf("tool execution failed: %w", err)
	}

	// Display results
	if len(result) == 0 {
		fmt.Println("Tool executed successfully (no output)")
		return nil
	}

	fmt.Println("Result:")
	for i, content := range result {
		if len(result) > 1 {
			fmt.Printf("\n[%d] ", i+1)
		}
		fmt.Println(content.Text)
	}

	return nil
}

// testToolExecution tests a tool with example arguments (for feature testing)
func testToolExecution(server framework.MCPServer, toolName string) error {
	fmt.Printf("Feature Testing: %s\n", toolName)
	fmt.Println("=" + strings.Repeat("=", len(toolName)+18))

	// Get tool info
	tools := server.ListTools()
	var toolInfo *framework.ToolInfo
	for _, t := range tools {
		if t.Name == toolName {
			toolInfo = &t
			break
		}
	}

	if toolInfo == nil {
		return fmt.Errorf("tool %q not found", toolName)
	}

	fmt.Printf("\nTool: %s\n", toolInfo.Name)
	if toolInfo.Description != "" {
		fmt.Printf("Description: %s\n", toolInfo.Description)
	}

	// Generate example arguments based on schema
	exampleArgs := generateExampleArgs(toolInfo.Schema)
	exampleJSON, _ := json.MarshalIndent(exampleArgs, "  ", "  ")

	fmt.Println("\nExample Arguments:")
	fmt.Printf("  %s\n", string(exampleJSON))

	fmt.Println("\nExecuting with example arguments...")
	fmt.Println()

	// Execute with example arguments
	argsBytes, _ := json.Marshal(exampleArgs)
	result, err := server.CallTool(context.Background(), toolName, argsBytes)
	if err != nil {
		fmt.Printf("❌ Test failed: %v\n", err)
		return err
	}

	fmt.Println("✅ Test passed!")
	if len(result) > 0 {
		fmt.Println("\nOutput:")
		for _, content := range result {
			// Truncate very long outputs
			output := content.Text
			if len(output) > 500 {
				output = output[:497] + "..."
			}
			fmt.Printf("  %s\n", output)
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

		propType, _ := propMap["type"].(string)
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
		fmt.Print("exarp-go> ")
		var input string
		fmt.Scanln(&input)

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
			listAllTools(server)
		default:
			fmt.Printf("Unknown command: %s\n", cmd)
			fmt.Println("Type 'help' for available commands")
		}
	}
}

// showUsage displays usage information
func showUsage() {
	fmt.Println("exarp-go - Command Line Interface")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  exarp-go [flags]")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -tool <name>        Execute a tool")
	fmt.Println("  -args <json>        Tool arguments as JSON (default: {})")
	fmt.Println("  -list               List all available tools")
	fmt.Println("  -test <name>        Test a tool with example arguments")
	fmt.Println("  -i                  Interactive mode")
	fmt.Println("  -completion <shell> Generate shell completion script (bash|zsh|fish)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  exarp-go -list")
	fmt.Println("  exarp-go -tool lint -args '{\"action\":\"run\",\"path\":\".\"}'")
	fmt.Println("  exarp-go -test lint")
	fmt.Println("  exarp-go -i")
	fmt.Println("  exarp-go -completion bash > /usr/local/etc/bash_completion.d/exarp-go")
	fmt.Println()
}

// showHelp displays help information
func showHelp() {
	fmt.Println("Available commands:")
	fmt.Println("  help              Show this help message")
	fmt.Println("  list              List all available tools")
	fmt.Println("  exit, quit        Exit interactive mode")
	fmt.Println()
}

// generateCompletion generates shell completion scripts
func generateCompletion(server framework.MCPServer, shell string) error {
	tools := server.ListTools()
	toolNames := make([]string, 0, len(tools))
	for _, tool := range tools {
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
	fmt.Println("# exarp-go bash completion")
	fmt.Println("_exarp_go() {")
	fmt.Println("    local cur prev opts")
	fmt.Println("    COMPREPLY=()")
	fmt.Println("    cur=\"${COMP_WORDS[COMP_CWORD]}\"")
	fmt.Println("    prev=\"${COMP_WORDS[COMP_CWORD-1]}\"")
	fmt.Println("    opts=\"-tool -args -list -test -i -completion\"")
	fmt.Println()
	fmt.Println("    case \"${prev}\" in")
	fmt.Println("        -tool|--tool)")
	fmt.Printf("            COMPREPLY=($(compgen -W \"%s\" -- \"${cur}\"))\n", strings.Join(toolNames, " "))
	fmt.Println("            return 0")
	fmt.Println("            ;;")
	fmt.Println("        -test|--test)")
	fmt.Printf("            COMPREPLY=($(compgen -W \"%s\" -- \"${cur}\"))\n", strings.Join(toolNames, " "))
	fmt.Println("            return 0")
	fmt.Println("            ;;")
	fmt.Println("        -completion|--completion)")
	fmt.Println("            COMPREPLY=($(compgen -W \"bash zsh fish\" -- \"${cur}\"))")
	fmt.Println("            return 0")
	fmt.Println("            ;;")
	fmt.Println("        *)")
	fmt.Println("            if [[ ${cur} == -* ]] ; then")
	fmt.Println("                COMPREPLY=($(compgen -W \"${opts}\" -- \"${cur}\"))")
	fmt.Println("            fi")
	fmt.Println("            ;;")
	fmt.Println("    esac")
	fmt.Println("}")
	fmt.Println("complete -F _exarp_go exarp-go")
	return nil
}

// generateZshCompletion generates zsh completion script
func generateZshCompletion(toolNames []string) error {
	fmt.Println("#compdef exarp-go")
	fmt.Println()
	fmt.Println("_exarp_go() {")
	fmt.Println("    local context state line")
	fmt.Println("    local -a tools")
	fmt.Printf("    tools=(%s)\n", strings.Join(toolNames, " "))
	fmt.Println()
	fmt.Println("    _arguments \\")
	fmt.Println("        '(-tool --tool)-tool[Tool name to execute]:tool:->tools' \\")
	fmt.Println("        '(-tool --tool)--tool[Tool name to execute]:tool:->tools' \\")
	fmt.Println("        '(-args --args)-args[Tool arguments as JSON]:args:' \\")
	fmt.Println("        '(-args --args)--args[Tool arguments as JSON]:args:' \\")
	fmt.Println("        '(-list --list)-list[List all available tools]' \\")
	fmt.Println("        '(-list --list)--list[List all available tools]' \\")
	fmt.Println("        '(-test --test)-test[Test a tool with example arguments]:test:->tools' \\")
	fmt.Println("        '(-test --test)--test[Test a tool with example arguments]:test:->tools' \\")
	fmt.Println("        '(-i --interactive)-i[Interactive mode]' \\")
	fmt.Println("        '(-i --interactive)--interactive[Interactive mode]' \\")
	fmt.Println("        '(-completion --completion)-completion[Generate shell completion script]:completion:(bash zsh fish)' \\")
	fmt.Println("        '(-completion --completion)--completion[Generate shell completion script]:completion:(bash zsh fish)'")
	fmt.Println()
	fmt.Println("    case $state in")
	fmt.Println("        tools)")
	fmt.Println("            _describe 'tools' tools")
	fmt.Println("            ;;")
	fmt.Println("    esac")
	fmt.Println("}")
	fmt.Println()
	fmt.Println("_exarp_go \"$@\"")
	return nil
}

// generateFishCompletion generates fish completion script
func generateFishCompletion(toolNames []string) error {
	fmt.Println("# exarp-go fish completion")
	fmt.Println()
	fmt.Printf("complete -c exarp-go -n '__fish_use_subcommand' -s t -l tool -d 'Tool name to execute' -xa '%s'\n", strings.Join(toolNames, " "))
	fmt.Println("complete -c exarp-go -n '__fish_use_subcommand' -l args -d 'Tool arguments as JSON'")
	fmt.Println("complete -c exarp-go -n '__fish_use_subcommand' -s l -l list -d 'List all available tools'")
	fmt.Printf("complete -c exarp-go -n '__fish_use_subcommand' -l test -d 'Test a tool with example arguments' -xa '%s'\n", strings.Join(toolNames, " "))
	fmt.Println("complete -c exarp-go -n '__fish_use_subcommand' -s i -l interactive -d 'Interactive mode'")
	fmt.Println("complete -c exarp-go -n '__fish_use_subcommand' -l completion -d 'Generate shell completion script' -xa 'bash zsh fish'")
	return nil
}

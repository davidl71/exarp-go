// Package cli - cursor subcommand. Tag hints for Todo2: #cli #feature
//
// Implements "exarp-go cursor run [task-id]" and "exarp-go cursor run -p <prompt>"
// to invoke the Cursor CLI "agent" with task context or a custom prompt.
// See docs/CURSOR_API_AND_CLI_INTEGRATION.md §3.1.

package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/davidl71/exarp-go/internal/framework"
	mcpcli "github.com/davidl71/mcp-go-core/pkg/mcp/cli"
)

// handleCursorCommand dispatches cursor subcommands (currently: run).
func handleCursorCommand(server framework.MCPServer, parsed *mcpcli.Args) error {
	sub := parsed.Subcommand
	if sub == "" && len(parsed.Positional) > 0 {
		sub = parsed.Positional[0]
	}

	switch sub {
	case "run":
		return handleCursorRun(server, parsed)
	case "help", "":
		return showCursorUsage()
	default:
		return fmt.Errorf("unknown cursor command: %s (use: run, help)", sub)
	}
}

// handleCursorRun implements "exarp-go cursor run [task-id] [-p prompt] [--mode plan|ask|agent] [--no-interactive]".
func handleCursorRun(server framework.MCPServer, parsed *mcpcli.Args) error {
	prompt := parsed.GetFlag("p", "")
	if prompt == "" {
		prompt = parsed.GetFlag("prompt", "")
	}

	mode := parsed.GetFlag("mode", "")
	noInteractive := parsed.GetBoolFlag("no-interactive", false)

	// Collect positional args after "run" subcommand
	positional := positionalAfterSub(parsed)

	// Determine task ID from positional args (T-xxx pattern)
	var taskID string
	for _, p := range positional {
		if strings.HasPrefix(p, "T-") {
			taskID = p
			break
		}
	}

	// Build prompt from task if task ID given and no explicit prompt
	if taskID != "" && prompt == "" {
		taskPrompt, err := buildPromptFromTask(server, taskID)
		if err != nil {
			return fmt.Errorf("failed to load task %s: %w", taskID, err)
		}

		prompt = taskPrompt
	}

	if prompt == "" {
		_, _ = fmt.Fprintln(os.Stderr, "Error: provide a task ID or -p \"prompt\"")
		_, _ = fmt.Fprintln(os.Stderr)

		return showCursorUsage()
	}

	// Find agent binary
	agentPath := AgentBinary()
	if agentPath == "" {
		return fmt.Errorf("cursor CLI agent not found on PATH\n\nInstall: curl https://cursor.com/install -fsS | bash\nThen:    export PATH=\"$HOME/.local/bin:$PATH\"")
	}

	// Build agent args
	agentArgs := buildAgentArgs(prompt, mode, noInteractive)

	// Resolve project root for working directory
	projectRoot, err := findCursorProjectRoot()
	if err != nil {
		return fmt.Errorf("could not determine project root: %w", err)
	}

	// Print what we're doing
	_, _ = fmt.Printf("Running Cursor agent in %s\n", projectRoot)
	_, _ = fmt.Printf("  agent %s\n\n", strings.Join(agentArgs, " "))

	// Execute agent (foreground, inheriting stdin/stdout/stderr)
	return execAgent(agentPath, agentArgs, projectRoot)
}

// buildPromptFromTask loads a task from Todo2 and builds a prompt.
func buildPromptFromTask(server framework.MCPServer, taskID string) (string, error) {
	ctx := context.Background()

	toolArgs := map[string]interface{}{
		"action":        "sync",
		"sub_action":    "list",
		"task_id":       taskID,
		"output_format": "json",
	}

	argsBytes, err := json.Marshal(toolArgs)
	if err != nil {
		return "", fmt.Errorf("marshal args: %w", err)
	}

	result, err := server.CallTool(ctx, "task_workflow", argsBytes)
	if err != nil {
		return "", err
	}

	if len(result) == 0 {
		return "", fmt.Errorf("task %s not found", taskID)
	}

	// Try to parse JSON response for task content
	content := extractTaskContent(result[0].Text, taskID)

	return PromptForTask(taskID, content), nil
}

// extractTaskContent attempts to extract task content from a task_workflow JSON response.
// Falls back to using the raw text if JSON parsing fails.
func extractTaskContent(text, taskID string) string {
	// Try parsing as JSON to get structured task info
	var response map[string]interface{}
	if err := json.Unmarshal([]byte(text), &response); err == nil {
		// Look for tasks array in response
		if tasks, ok := response["tasks"].([]interface{}); ok {
			for _, t := range tasks {
				if task, ok := t.(map[string]interface{}); ok {
					id, _ := task["id"].(string)
					if id == taskID {
						content, _ := task["content"].(string)
						if content != "" {
							return content
						}
					}
				}
			}
		}

		// Try single task in response
		if content, ok := response["content"].(string); ok && content != "" {
			return content
		}
	}

	// Fallback: use task ID as content
	return "Task " + taskID
}

// buildAgentArgs constructs the arguments for the Cursor agent CLI.
func buildAgentArgs(prompt, mode string, noInteractive bool) []string {
	var args []string

	if noInteractive {
		args = append(args, "-p", prompt)
	} else {
		args = append(args, prompt)
	}

	if mode != "" {
		args = append(args, "--mode="+mode)
	}

	return args
}

// findCursorProjectRoot returns the project root, preferring CWD if it looks like a project.
func findCursorProjectRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return cwd, nil
}

// execAgent runs the Cursor agent binary as a foreground process.
func execAgent(agentPath string, args []string, projectRoot string) error {
	cmd := exec.Command(agentPath, args...)
	cmd.Dir = projectRoot
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	return cmd.Run()
}

// positionalAfterSub returns positional args, skipping the first one if it matches the subcommand "run".
func positionalAfterSub(parsed *mcpcli.Args) []string {
	pos := parsed.Positional
	if len(pos) > 0 && pos[0] == "run" {
		pos = pos[1:]
	}

	return pos
}

// showCursorUsage displays usage for the cursor subcommand.
func showCursorUsage() error {
	_, _ = fmt.Println("Cursor CLI Integration")
	_, _ = fmt.Println("======================")
	_, _ = fmt.Println()
	_, _ = fmt.Println("Usage:")
	_, _ = fmt.Println("  exarp-go cursor run <task-id>              Run Cursor agent with task prompt")
	_, _ = fmt.Println("  exarp-go cursor run -p \"prompt\"            Run with custom prompt")
	_, _ = fmt.Println("  exarp-go cursor run <task-id> --mode plan  Run in plan mode")
	_, _ = fmt.Println("  exarp-go cursor run -p \"...\" --no-interactive  Non-interactive mode")
	_, _ = fmt.Println("  exarp-go cursor help                      Show this help")
	_, _ = fmt.Println()
	_, _ = fmt.Println("Options:")
	_, _ = fmt.Println("  -p, --prompt <text>   Custom prompt (overrides task prompt)")
	_, _ = fmt.Println("  --mode <mode>         Agent mode: plan, ask, or agent (default: agent)")
	_, _ = fmt.Println("  --no-interactive      Non-interactive mode (uses agent -p)")
	_, _ = fmt.Println()
	_, _ = fmt.Println("Examples:")
	_, _ = fmt.Println("  exarp-go cursor run T-123")
	_, _ = fmt.Println("  exarp-go cursor run T-123 --mode plan")
	_, _ = fmt.Println("  exarp-go cursor run -p \"Review the backlog and suggest next task\"")
	_, _ = fmt.Println("  exarp-go cursor run -p \"Fix performance issues\" --no-interactive")
	_, _ = fmt.Println()
	_, _ = fmt.Println("Prerequisites:")
	_, _ = fmt.Println("  Install Cursor CLI: curl https://cursor.com/install -fsS | bash")
	_, _ = fmt.Println("  Ensure 'agent' is on PATH: export PATH=\"$HOME/.local/bin:$PATH\"")
	_, _ = fmt.Println("  Optional: set CURSOR_API_KEY for non-interactive/CI use")
	_, _ = fmt.Println()

	return nil
}

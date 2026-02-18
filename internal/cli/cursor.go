// Package cli - Cursor CLI integration. Tag hints for Todo2: #cli #cursor
//
// cursor run invokes the Cursor CLI "agent" from project root with a task or custom prompt.
// See docs/CURSOR_API_AND_CLI_INTEGRATION.md §3.1.

package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/tools"
	mcpcli "github.com/davidl71/mcp-go-core/pkg/mcp/cli"
)

// handleCursorCommand handles the "cursor" top-level command (subcommand: run).
func handleCursorCommand(parsed *mcpcli.Args) error {
	sub := strings.ToLower(strings.TrimSpace(parsed.Subcommand))
	if sub == "" && len(parsed.Positional) > 0 {
		// cursor run T-123 → subcommand might be "run", first positional T-123
		first := parsed.Positional[0]
		if strings.HasPrefix(first, "T-") {
			sub = "run"
		} else {
			sub = strings.ToLower(first)
		}
	}

	if sub == "" {
		sub = "run"
	}

	switch sub {
	case "run":
		return runCursorRun(parsed)
	default:
		return fmt.Errorf("unknown cursor command: %s (use: cursor run [task-id] or cursor run -p \"prompt\")", sub)
	}
}

// runCursorRun runs the Cursor agent with a prompt from task-id or -p.
// Options: --mode plan|ask|agent, -p "prompt", --no-interactive.
func runCursorRun(parsed *mcpcli.Args) error {
	if parsed.HasFlag("help") || parsed.HasFlag("h") {
		printCursorRunHelp()
		return nil
	}

	projectRoot, err := tools.FindProjectRoot()
	if err != nil {
		return fmt.Errorf("project root: %w", err)
	}

	customPrompt := parsed.GetFlag("p", "")
	customPrompt = strings.TrimSpace(customPrompt)

	if customPrompt == "" {
		customPrompt = parsed.GetFlag("prompt", "")
		customPrompt = strings.TrimSpace(customPrompt)
	}

	taskID := parsed.GetFlag("task", "")
	if taskID == "" {
		for _, p := range parsed.Positional {
			if strings.HasPrefix(p, "T-") {
				taskID = p
				break
			}
		}
	}

	var prompt string
	if customPrompt != "" {
		prompt = customPrompt
	} else if taskID != "" {
		task, err := loadTaskForCursor(taskID)
		if err != nil {
			return fmt.Errorf("load task %s: %w", taskID, err)
		}

		prompt = PromptForTask(task.ID, task.Content)

		if task.LongDescription != "" {
			desc := strings.TrimSpace(task.LongDescription)
			if len(desc) > 400 {
				desc = desc[:397] + "..."
			}

			prompt = prompt + "\n\n" + desc
		}
	} else {
		return fmt.Errorf("provide a task-id (e.g. T-123) or -p \"prompt\" (see: exarp-go cursor run --help)")
	}

	if prompt == "" {
		return fmt.Errorf("empty prompt")
	}

	mode := parsed.GetFlag("mode", "plan")
	if mode != "plan" && mode != "ask" && mode != "agent" {
		mode = "plan"
	}

	noInteractive := parsed.GetBoolFlag("no-interactive", false)

	execPath, baseArgs := agentCommand()
	if execPath == "" {
		fmt.Fprintln(os.Stderr, "agent command not on PATH (default: cursor agent; set EXARP_AGENT_CMD or install Cursor CLI)")
		fmt.Fprintln(os.Stderr, "  Install: curl https://cursor.com/install -fsS | bash")
		fmt.Fprintln(os.Stderr, "  See: https://cursor.com/docs/cli/overview")

		return fmt.Errorf("agent command not found")
	}

	args := append([]string{}, baseArgs...)
	args = append(args, "--mode="+mode)

	if noInteractive {
		args = append(args, "-p", prompt)
	} else {
		args = append(args, prompt)
	}

	ctx := context.Background()
	cmd := exec.CommandContext(ctx, execPath, args...)
	cmd.Dir = projectRoot
	cmd.Env = os.Environ()
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if noInteractive {
		err := cmd.Run()
		if err != nil {
			exitErr := &exec.ExitError{}
			if errors.As(err, &exitErr) {
				os.Exit(exitErr.ExitCode())
			}

			return err
		}

		return nil
	}

	// Interactive: on macOS open new terminal so user can interact; otherwise run in foreground
	if runtime.GOOS == "darwin" {
		result := runInNewTerminal(projectRoot, prompt)
		if !result.Launched {
			return fmt.Errorf("%s", result.Message)
		}

		fmt.Fprintln(os.Stderr, result.Message)

		return nil
	}

	return cmd.Run()
}

// printCursorRunHelp prints usage for cursor run.
func printCursorRunHelp() {
	fmt.Println("Usage: exarp-go cursor run [task-id] | exarp-go cursor run -p \"prompt\"")
	fmt.Println()
	fmt.Println("  Run Cursor agent from project root with task context or a custom prompt.")
	fmt.Println("  Agent is detected on PATH (try 'agent' or EXARP_AGENT_CMD, e.g. 'cursor agent').")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -p, --prompt <text>     Custom prompt (alternative to task-id)")
	fmt.Println("  --task <task-id>        Task ID (e.g. T-123); or pass as positional after run")
	fmt.Println("  --mode plan|ask|agent   Cursor agent mode (default: plan)")
	fmt.Println("  --no-interactive        Use agent -p (non-interactive); for scripts/CI")
	fmt.Println("  -h, --help              Show this help")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  exarp-go cursor run T-123")
	fmt.Println("  exarp-go cursor run -p \"Implement the refactor\" --no-interactive")
	fmt.Println("  exarp-go cursor run --task T-123 --mode agent")
	fmt.Println()
	fmt.Println("Install Cursor CLI: curl https://cursor.com/install -fsS | bash")
	fmt.Println("See: docs/CURSOR_API_AND_CLI_INTEGRATION.md")
}

// loadTaskForCursor loads a task by ID for cursor run (requires DB init).
func loadTaskForCursor(taskID string) (*database.Todo2Task, error) {
	ctx := context.Background()

	task, err := database.GetTask(ctx, taskID)
	if err != nil {
		return nil, err
	}

	if task == nil {
		return nil, fmt.Errorf("task %s not found", taskID)
	}

	return task, nil
}

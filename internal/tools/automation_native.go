// automation_native.go — MCP automation tool: core handler, agent runner, step helpers.
package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/proto"
)

// AutomationResponseToMap converts AutomationResponse to a map for response.FormatResult (unmarshals result_json).
func AutomationResponseToMap(resp *proto.AutomationResponse) map[string]interface{} {
	if resp == nil {
		return nil
	}

	out := map[string]interface{}{"action": resp.GetAction()}

	if resp.GetResultJson() != "" {
		var payload map[string]interface{}
		if json.Unmarshal([]byte(resp.GetResultJson()), &payload) == nil {
			for k, v := range payload {
				out[k] = v
			}
		}
	}

	return out
}

// handleAutomationNative handles the automation tool with native Go implementation
// Implements all actions: "daily", "nightly", "sprint", and "discover".
func handleAutomationNative(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	action, _ := params["action"].(string)
	if action == "" {
		action = "daily"
	}

	switch action {
	case "daily":
		return handleAutomationDaily(ctx, params)
	case "discover":
		return handleAutomationDiscover(ctx, params)
	case "nightly":
		return handleAutomationNightly(ctx, params)
	case "sprint":
		return handleAutomationSprint(ctx, params)
	default:
		return nil, fmt.Errorf("unknown automation action: %s (use 'daily', 'nightly', 'sprint', or 'discover')", action)
	}
}

// automationAgentCommand returns the executable path and base args for the Cursor agent command.
// Respects EXARP_AGENT_CMD (e.g. "agent" or "cursor agent"). Used by automation cursor-agent step.
func automationAgentCommand() (string, []string) {
	if raw := os.Getenv("EXARP_AGENT_CMD"); raw != "" {
		parts := strings.Fields(strings.TrimSpace(raw))
		if len(parts) == 0 {
			return "", nil
		}

		if path, err := exec.LookPath(parts[0]); err == nil {
			if len(parts) > 1 {
				return path, parts[1:]
			}

			return path, nil
		}

		return "", nil
	}

	if path, err := exec.LookPath("agent"); err == nil {
		return path, nil
	}

	parts := strings.Fields(strings.TrimSpace("cursor agent"))
	if len(parts) == 0 {
		return "", nil
	}

	if path, err := exec.LookPath(parts[0]); err == nil {
		if len(parts) > 1 {
			return path, parts[1:]
		}

		return path, nil
	}

	return "", nil
}

func cursorAgentTimeout() time.Duration {
	if raw := strings.TrimSpace(os.Getenv("EXARP_AGENT_TIMEOUT_SECONDS")); raw != "" {
		if sec, err := strconv.Atoi(raw); err == nil && sec > 0 {
			return time.Duration(sec) * time.Second
		}
	}

	return 60 * time.Second
}

// runCursorAgentStep runs the Cursor CLI agent with -p prompt in projectRoot and captures stdout/stderr.
// Returns combined output, exit code, and any error. Used when automation use_cursor_agent is true.
func runCursorAgentStep(projectRoot, prompt string) (output string, exitCode int, err error) {
	if projectRoot == "" || prompt == "" {
		return "", -1, fmt.Errorf("project root and prompt required")
	}

	agentPath, baseArgs := automationAgentCommand()
	if agentPath == "" {
		return "", -1, fmt.Errorf("agent not on PATH (set EXARP_AGENT_CMD or install Cursor CLI: https://cursor.com/docs/cli/overview)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), cursorAgentTimeout())
	defer cancel()

	args := append([]string{}, baseArgs...)
	args = append(args, "-p", prompt)
	cmd := exec.CommandContext(ctx, agentPath, args...)
	cmd.Dir = projectRoot
	cmd.Env = os.Environ()
	cmd.Stdin = nil

	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	runErr := cmd.Run()

	if stdout.Len() > 0 || stderr.Len() > 0 {
		var b strings.Builder
		if stdout.Len() > 0 {
			b.Write(stdout.Bytes())
		}

		if stderr.Len() > 0 {
			if b.Len() > 0 {
				b.WriteString("\n--- stderr ---\n")
			}

			b.Write(stderr.Bytes())
		}

		output = strings.TrimSpace(b.String())
	}

	code := 0

	if runErr != nil {
		exitErr := &exec.ExitError{}
		if errors.As(runErr, &exitErr) {
			code = exitErr.ExitCode()
		} else {
			code = -1
		}
	}

	return output, code, runErr
}

// appendCursorAgentStepIfRequested runs the optional Cursor agent step and appends to results.
// Uses params["use_cursor_agent"] and params["cursor_agent_prompt"]. No-op if use_cursor_agent is false or project root not found.
func appendCursorAgentStepIfRequested(ctx context.Context, params map[string]interface{}, results map[string]interface{}, startTime time.Time) {
	if useCursor, _ := params["use_cursor_agent"].(bool); !useCursor {
		return
	}

	projectRoot, rootErr := FindProjectRoot()
	if rootErr != nil || projectRoot == "" {
		return
	}

	prompt, _ := params["cursor_agent_prompt"].(string)
	if prompt == "" {
		prompt = "Review the backlog and suggest which task to do next"
	}

	startCursor := time.Now()
	out, code, runErr := runCursorAgentStep(projectRoot, prompt)
	dur := time.Since(startCursor).Seconds()
	status := "success"
	sum := fmt.Sprintf("exit_code=%d, duration_sec=%.2f", code, dur)

	if runErr != nil {
		status = "failed"
		sum = runErr.Error()
	}

	taskCursor := map[string]interface{}{
		"task_id":   "cursor_agent",
		"task_name": "Cursor Agent Step",
		"status":    status,
		"duration":  dur,
		"error":     nil,
		"summary":   sum,
	}
	if runErr != nil {
		taskCursor["error"] = runErr.Error()
	}

	if out != "" {
		taskCursor["output"] = out
	}

	tasksRun, _ := results["tasks_run"].([]map[string]interface{})
	tasksRun = append(tasksRun, taskCursor)

	results["tasks_run"] = tasksRun
	if status == "success" {
		tasksSucceeded, _ := results["tasks_succeeded"].([]string)
		results["tasks_succeeded"] = append(tasksSucceeded, "cursor_agent")
	} else {
		tasksFailed, _ := results["tasks_failed"].([]string)
		results["tasks_failed"] = append(tasksFailed, "cursor_agent")
	}
}


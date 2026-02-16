// Package cli - child agent execution. Tag hints for Todo2: #cli #tui
//
// RunTaskExecutionFlow runs the Cursor CLI "agent" in project root with a prompt
// so a task, plan, wave, or handoff can be executed in a child agent (new Cursor session).

package cli

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// JobOutputMsg is sent when a non-interactive child agent completes (for background jobs output capture).
type JobOutputMsg struct {
	Pid      int
	Output   string
	ExitCode int
	Err      error
}

// ChildAgentKind is the kind of context to run in the child agent.
type ChildAgentKind string

const (
	ChildAgentTask    ChildAgentKind = "task"
	ChildAgentPlan    ChildAgentKind = "plan"
	ChildAgentWave    ChildAgentKind = "wave"
	ChildAgentHandoff ChildAgentKind = "handoff"
)

// ChildAgentRunResult is the result of starting a child agent (for TUI feedback).
type ChildAgentRunResult struct {
	Kind     ChildAgentKind
	Prompt   string
	Launched bool
	Message  string // "Launched" or error description
	Pid      int    // Process ID of launched agent (0 if not available)
}

// AgentBinary returns the path to the Cursor CLI "agent" binary, or "" if not found.
func AgentBinary() string {
	path, err := exec.LookPath("agent")
	if err != nil {
		return ""
	}

	return path
}

// RunChildAgent starts the Cursor CLI agent in projectRoot with the given prompt.
// It uses -p (non-interactive); the process runs in the background so the TUI can continue.
// Returns a result suitable for TUI display (Launched true + message, or Launched false + error message).
func RunChildAgent(projectRoot, prompt string) (result ChildAgentRunResult) {
	return runChildAgent(projectRoot, prompt, false)
}

// RunChildAgentWithOutputCapture starts the non-interactive agent and captures stdout+stderr.
// Returns the launch result and a channel that receives JobOutputMsg when the process exits.
func RunChildAgentWithOutputCapture(projectRoot, prompt string) (result ChildAgentRunResult, done <-chan JobOutputMsg) {
	return runChildAgentWithCapture(projectRoot, prompt)
}

// RunChildAgentInteractive starts the Cursor CLI agent in projectRoot with the given prompt
// in interactive mode (agent opens with the prompt as initial message; user can continue the conversation).
// On macOS, runs in a new Terminal window for full interaction; otherwise runs in background.
func RunChildAgentInteractive(projectRoot, prompt string) (result ChildAgentRunResult) {
	return runChildAgent(projectRoot, prompt, true)
}

// runChildAgent starts the Cursor CLI agent; interactive=true runs in new terminal (darwin) or "agent prompt", interactive=false runs "agent -p prompt".
func runChildAgent(projectRoot, prompt string, interactive bool) (result ChildAgentRunResult) {
	if prompt == "" {
		result.Launched = false
		result.Message = "no prompt"

		return result
	}

	agentPath := AgentBinary()
	if agentPath == "" {
		result.Launched = false
		result.Message = "agent not on PATH (install Cursor CLI: https://cursor.com/docs/cli/overview)"

		return result
	}

	if projectRoot == "" {
		result.Launched = false
		result.Message = "no project root"

		return result
	}

	if interactive && runtime.GOOS == "darwin" {
		return runInNewTerminal(projectRoot, prompt)
	}

	ctx := context.Background()
	var cmd *exec.Cmd
	if interactive {
		cmd = exec.CommandContext(ctx, agentPath, prompt)
	} else {
		cmd = exec.CommandContext(ctx, agentPath, "-p", prompt)
	}
	cmd.Dir = projectRoot
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Env = os.Environ()

	if err := cmd.Start(); err != nil {
		result.Launched = false
		result.Message = err.Error()

		return result
	}

	if cmd.Process != nil {
		result.Pid = cmd.Process.Pid
	}

	go func() { _ = cmd.Wait() }()

	result.Launched = true
	result.Prompt = prompt

	prefix := "Launched: "
	if interactive {
		prefix = "Launched (interactive): "
	}
	if len(prompt) > 60 {
		result.Message = prefix + prompt[:57] + "..."
	} else {
		result.Message = prefix + prompt
	}

	return result
}

// runInNewTerminal opens a new Terminal/iTerm tab and runs the agent interactively (macOS).
// Uses a temp file for the prompt to avoid shell quoting issues. Prefers iTerm tab when running in iTerm.
func runInNewTerminal(projectRoot, prompt string) (result ChildAgentRunResult) {
	tmp, err := os.CreateTemp("", "exarp-agent-prompt-*.txt")
	if err != nil {
		result.Launched = false
		result.Message = "failed to create temp file: " + err.Error()
		return result
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if _, err := tmp.WriteString(prompt); err != nil {
		result.Launched = false
		result.Message = "failed to write prompt: " + err.Error()
		return result
	}
	if err := tmp.Close(); err != nil {
		result.Launched = false
		result.Message = "failed to close temp file: " + err.Error()
		return result
	}

	if runIniTermTab(projectRoot, tmpPath) {
		result.Launched = true
		result.Prompt = prompt
		result.Pid = 0
		if len(prompt) > 60 {
			result.Message = "Launched (interactive): " + prompt[:57] + "..."
		} else {
			result.Message = "Launched (interactive): " + prompt
		}
		return result
	}

	// Fallback to Terminal.app
	script := `on run argv
  set projectRoot to item 1 of argv
  set promptPath to item 2 of argv
  set promptText to read (POSIX file promptPath) as text
  tell application "Terminal" to do script "cd " & quoted form of projectRoot & " && agent " & quoted form of promptText
  tell application "Terminal" to activate
end run`
	cmd := exec.Command("osascript", "-e", script, "--", projectRoot, tmpPath)
	cmd.Env = os.Environ()
	if err := cmd.Run(); err != nil {
		result.Launched = false
		result.Message = "failed to open Terminal: " + err.Error()
		return result
	}

	result.Launched = true
	result.Prompt = prompt
	result.Pid = 0
	if len(prompt) > 60 {
		result.Message = "Launched (interactive): " + prompt[:57] + "..."
	} else {
		result.Message = "Launched (interactive): " + prompt
	}
	return result
}

// runIniTermTab opens a new iTerm tab and runs the agent. Returns true if iTerm was used.
func runIniTermTab(projectRoot, promptPath string) bool {
	if os.Getenv("ITERM_SESSION_ID") == "" && os.Getenv("TERM_PROGRAM") != "iTerm.app" {
		return false
	}
	script := `on run argv
  set projectRoot to item 1 of argv
  set promptPath to item 2 of argv
  tell application "iTerm" to activate
  tell current window
    set tb to create tab with default profile
  end tell
  tell current session of tb to write text "cd " & quoted form of projectRoot & " && agent \"$(cat " & quoted form of promptPath & ")\""
end run`
	cmd := exec.Command("osascript", "-e", script, "--", projectRoot, promptPath)
	cmd.Env = os.Environ()
	return cmd.Run() == nil
}

// runChildAgentWithCapture starts the non-interactive agent, captures output, and sends JobOutputMsg when done.
func runChildAgentWithCapture(projectRoot, prompt string) (result ChildAgentRunResult, done <-chan JobOutputMsg) {
	ch := make(chan JobOutputMsg, 1)
	if prompt == "" {
		result.Launched = false
		result.Message = "no prompt"
		ch <- JobOutputMsg{Err: fmt.Errorf("no prompt")}
		return result, ch
	}

	agentPath := AgentBinary()
	if agentPath == "" {
		result.Launched = false
		result.Message = "agent not on PATH (install Cursor CLI: https://cursor.com/docs/cli/overview)"
		ch <- JobOutputMsg{Err: fmt.Errorf("agent not on PATH")}
		return result, ch
	}

	if projectRoot == "" {
		result.Launched = false
		result.Message = "no project root"
		ch <- JobOutputMsg{Err: fmt.Errorf("no project root")}
		return result, ch
	}

	ctx := context.Background()
	cmd := exec.CommandContext(ctx, agentPath, "-p", prompt)
	cmd.Dir = projectRoot
	cmd.Stdin = nil

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		result.Launched = false
		result.Message = err.Error()
		ch <- JobOutputMsg{Err: err}
		return result, ch
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		result.Launched = false
		result.Message = err.Error()
		ch <- JobOutputMsg{Err: err}
		return result, ch
	}

	cmd.Env = os.Environ()

	if err := cmd.Start(); err != nil {
		result.Launched = false
		result.Message = err.Error()
		ch <- JobOutputMsg{Err: err}
		return result, ch
	}

	pid := 0
	if cmd.Process != nil {
		pid = cmd.Process.Pid
	}

	result.Launched = true
	result.Prompt = prompt
	if len(prompt) > 60 {
		result.Message = "Launched: " + prompt[:57] + "..."
	} else {
		result.Message = "Launched: " + prompt
	}

	go func() {
		var stdoutBuf, stderrBuf bytes.Buffer
		pipeDone := make(chan struct{}, 2)
		go func() { _, _ = io.Copy(&stdoutBuf, stdoutPipe); pipeDone <- struct{}{} }()
		go func() { _, _ = io.Copy(&stderrBuf, stderrPipe); pipeDone <- struct{}{} }()
		exitErr := cmd.Wait()
		<-pipeDone
		<-pipeDone
		var combined strings.Builder
		if stdoutBuf.Len() > 0 {
			combined.Write(stdoutBuf.Bytes())
		}
		if stderrBuf.Len() > 0 {
			if combined.Len() > 0 {
				combined.WriteString("\n--- stderr ---\n")
			}
			combined.Write(stderrBuf.Bytes())
		}
		exitCode := 0
		if exitErr != nil {
			if exit, ok := exitErr.(*exec.ExitError); ok {
				exitCode = exit.ExitCode()
			}
		}
		ch <- JobOutputMsg{Pid: pid, Output: strings.TrimSpace(combined.String()), ExitCode: exitCode, Err: exitErr}
	}()

	return result, ch
}

// PromptForTask builds a child-agent prompt for a single task (e.g. "Work on T-123: Task name").
func PromptForTask(taskID, content string) string {
	content = strings.TrimSpace(content)
	if content == "" {
		content = "Task " + taskID
	}

	if len(content) > 200 {
		content = content[:197] + "..."
	}

	return fmt.Sprintf("Work on %s: %s", taskID, content)
}

// PromptForPlan builds a child-agent prompt for the main project plan.
func PromptForPlan(projectRoot string) string {
	_ = projectRoot // agent runs in project root
	return "Execute the current project plan in .cursor/plans (open the .plan.md for this repo and work through the next steps)."
}

// PromptForWave builds a child-agent prompt for a dependency wave (wave index and task IDs).
func PromptForWave(waveIndex int, taskIDs []string) string {
	if len(taskIDs) == 0 {
		return fmt.Sprintf("Work on Wave %d (no tasks in wave).", waveIndex)
	}

	if len(taskIDs) == 1 {
		return fmt.Sprintf("Work on Wave %d task: %s", waveIndex, taskIDs[0])
	}

	return fmt.Sprintf("Work on Wave %d tasks: %s", waveIndex, strings.Join(taskIDs, ", "))
}

// PromptForHandoff builds a child-agent prompt from a handoff entry (summary + next steps).
func PromptForHandoff(summary string, nextSteps []interface{}) string {
	prompt := "Resume from handoff."

	if summary != "" {
		if len(summary) > 300 {
			summary = summary[:297] + "..."
		}

		prompt = "Resume from handoff. Summary: " + summary
	}

	if len(nextSteps) > 0 {
		var steps []string

		for _, s := range nextSteps {
			if str, ok := s.(string); ok && str != "" {
				steps = append(steps, str)
			}
		}

		if len(steps) > 0 {
			prompt += " Next steps: " + strings.Join(steps, "; ")
			if len(prompt) > 400 {
				prompt = prompt[:397] + "..."
			}
		}
	}

	return prompt
}

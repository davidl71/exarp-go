// Package cli - child agent execution. Tag hints for Todo2: #cli #tui
//
// RunTaskExecutionFlow runs the Cursor CLI "agent" in project root with a prompt
// so a task, plan, wave, or handoff can be executed in a child agent (new Cursor session).

package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

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
	Kind    ChildAgentKind
	Prompt  string
	Launched bool
	Message string // "Launched" or error description
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
// It starts the process in the background (does not wait) so the TUI can continue.
// Returns a result suitable for TUI display (Launched true + message, or Launched false + error message).
func RunChildAgent(projectRoot, prompt string) (result ChildAgentRunResult) {
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

	// Run agent -p "..." in project root. Use Start() so we don't block.
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, agentPath, "-p", prompt)
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
	go func() { _ = cmd.Wait() }()
	result.Launched = true
	result.Prompt = prompt
	if len(prompt) > 60 {
		result.Message = "Launched: " + prompt[:57] + "..."
	} else {
		result.Message = "Launched: " + prompt
	}
	return result
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

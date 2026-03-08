// agent_runner_cursor_cli.go — AgentRunner implementation for Cursor CLI agent.
package tools

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type CursorCLIRunner struct{}

func NewCursorCLIRunner() *CursorCLIRunner {
	return &CursorCLIRunner{}
}

func (r *CursorCLIRunner) Type() AgentType {
	return AgentTypeCursorCLI
}

func (r *CursorCLIRunner) Available() bool {
	agentPath, _ := automationAgentCommand()
	return agentPath != ""
}

func (r *CursorCLIRunner) Run(ctx context.Context, opts AgentRunnerOptions) AgentResult {
	if opts.ProjectRoot == "" || opts.Prompt == "" {
		return AgentResult{Error: fmt.Errorf("project root and prompt required")}
	}

	agentPath, baseArgs := automationAgentCommand()
	if agentPath == "" {
		return AgentResult{Error: fmt.Errorf("agent not on PATH (set EXARP_AGENT_CMD or install Cursor CLI)")}
	}

	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = cursorAgentTimeout()
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	args := append([]string{}, baseArgs...)
	args = append(args, "-p", opts.Prompt)
	if opts.Mode != "" {
		args = append(args, "--mode="+opts.Mode)
	}

	cmd := exec.CommandContext(ctx, agentPath, args...)
	cmd.Dir = opts.ProjectRoot
	cmd.Env = os.Environ()
	cmd.Stdin = nil

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	runErr := cmd.Run()

	var output string
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

	return AgentResult{
		Output:   output,
		ExitCode: code,
		Error:    runErr,
	}
}

// Package cli - tests for cursor run subcommand.
package cli

import (
	"os"
	"strings"
	"testing"

	mcpcli "github.com/davidl71/mcp-go-core/pkg/mcp/cli"
)

func TestHandleCursorCommand_Help(t *testing.T) {
	// cursor run --help and cursor run -h should return nil (help printed to stdout)
	for _, args := range [][]string{
		{"cursor", "run", "--help"},
		{"cursor", "run", "-h"},
	} {
		parsed := mcpcli.ParseArgs(args)
		if parsed.Command != "cursor" {
			t.Logf("ParseArgs(%v) => command=%q", args, parsed.Command)
		}
		err := handleCursorCommand(parsed)
		if err != nil {
			t.Errorf("handleCursorCommand(%v) err = %v", args, err)
		}
	}
}

func TestHandleCursorCommand_NoTaskOrPrompt(t *testing.T) {
	parsed := mcpcli.ParseArgs([]string{"cursor", "run"})
	err := handleCursorCommand(parsed)
	if err == nil {
		t.Error("expected error when no task-id or -p")
	}
	if !strings.Contains(err.Error(), "task-id") && !strings.Contains(err.Error(), "prompt") {
		t.Errorf("error should mention task-id or prompt: %v", err)
	}
}

func TestPromptForTask(t *testing.T) {
	tests := []struct {
		taskID  string
		content string
		want    string
	}{
		{"T-1", "Fix bug", "Work on T-1: Fix bug"},
		{"T-2", "", "Work on T-2: Task T-2"},
		{"T-3", strings.Repeat("x", 250), "Work on T-3: " + strings.Repeat("x", 197) + "..."},
	}
	for _, tt := range tests {
		got := PromptForTask(tt.taskID, tt.content)
		if got != tt.want {
			t.Errorf("PromptForTask(%q, %q) = %q, want %q", tt.taskID, tt.content, got, tt.want)
		}
	}
}

func TestAgentCommand_EnvOverride(t *testing.T) {
	// When EXARP_AGENT_CMD is set, agentCommand should use it (we can't assert path without agent)
	const fake = "/nonexistent/agent"
	os.Setenv("EXARP_AGENT_CMD", fake)
	defer os.Unsetenv("EXARP_AGENT_CMD")
	path, args := agentCommand()
	if path != "" {
		// If /nonexistent/agent exists (unlikely), path would be set
		t.Logf("EXARP_AGENT_CMD set: path=%q args=%v", path, args)
	}
	// When unset, we try "agent" then "cursor agent" - just ensure we don't panic
	os.Unsetenv("EXARP_AGENT_CMD")
	path, args = agentCommand()
	t.Logf("EXARP_AGENT_CMD unset: path=%q args=%v", path, args)
}

func TestPrintCursorRunHelp_NoPanic(t *testing.T) {
	// Ensure help printer doesn't panic (output goes to stdout)
	printCursorRunHelp()
}

func TestRunCursorRun_CustomPrompt(t *testing.T) {
	// Just ensure we don't panic when -p is provided; actual exec would need agent on PATH
	parsed := mcpcli.ParseArgs([]string{"cursor", "run", "-p", "hello"})
	// Run in a dir that is not a project root to avoid DB; we expect project root error or agent not found
	err := runCursorRun(parsed)
	if err == nil {
		return // e.g. agent found and ran (unlikely in test)
	}
	msg := err.Error()
	if !strings.Contains(msg, "project root") && !strings.Contains(msg, "agent") && !strings.Contains(msg, "empty") {
		t.Logf("runCursorRun(-p hello) err = %v (acceptable)", err)
	}
}

func TestLoadTaskForCursor_NoDB(t *testing.T) {
	// Without initializing DB, GetTask may fail or return nil
	_, err := loadTaskForCursor("T-nonexistent")
	if err != nil {
		// Expected when DB not init or task missing
		if !strings.Contains(err.Error(), "not found") && !strings.Contains(err.Error(), "database") {
			t.Logf("loadTaskForCursor err = %v", err)
		}
	}
}

// Ensure cursor run subcommand is accepted (integration with ParseArgs)
func TestCursorRun_ParseArgs(t *testing.T) {
	cases := []struct {
		args        []string
		wantCommand string
		wantSub     string
	}{
		{[]string{"cursor", "run", "T-123"}, "cursor", "run"},
		{[]string{"cursor", "run", "-p", "hi"}, "cursor", "run"},
		{[]string{"cursor", "run", "--mode=agent"}, "cursor", "run"},
	}
	for _, c := range cases {
		parsed := mcpcli.ParseArgs(c.args)
		if parsed.Command != c.wantCommand {
			t.Errorf("ParseArgs(%v) Command = %q, want %q", c.args, parsed.Command, c.wantCommand)
		}
		if parsed.Subcommand != c.wantSub {
			t.Errorf("ParseArgs(%v) Subcommand = %q, want %q", c.args, parsed.Subcommand, c.wantSub)
		}
	}
}

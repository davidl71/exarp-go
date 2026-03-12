package cli

import (
	"strings"
	"testing"

	"github.com/davidl71/exarp-go/internal/config"
)

func TestTransitionToRejectsInvalidModeChange(t *testing.T) {
	server := setupMockServer(t)
	m := initialModel(server, "", "/test", "test-project", 0, 0)
	m.mode = ModeScorecard

	if ok := m.transitionTo(ModeConfig); ok {
		t.Fatal("transitionTo should reject scorecard -> config")
	}
	if m.mode != ModeScorecard {
		t.Fatalf("mode = %s, want %s", m.mode, ModeScorecard)
	}
}

func TestKeyMatchesUsesConfiguredBindings(t *testing.T) {
	server := setupMockServer(t)
	m := initialModel(server, "", "/test", "test-project", 0, 0)
	m.configData.Tasks.Keybindings["refresh"] = []string{"ctrl+r"}

	if !m.keyMatches("ctrl+r", KeyActionRefresh) {
		t.Fatal("expected custom refresh binding to match")
	}
	if m.keyMatches("R", KeyActionRefresh) {
		t.Fatal("default refresh binding should be overridden by custom binding")
	}
}

func TestViewHelpUsesConfiguredBindings(t *testing.T) {
	server := setupMockServer(t)
	m := initialModel(server, "", "/test", "test-project", 100, 40)
	m.configData = config.GetDefaults()
	m.configData.Tasks.Keybindings["create_task"] = []string{"n"}

	help := m.viewHelp()
	if !strings.Contains(help, "n") {
		t.Fatal("help output should include configured create-task binding")
	}
	if strings.Contains(help, "+  Create new task inline") {
		t.Fatal("help output should not show default create-task binding after override")
	}
}

func TestT3270HelpersAndSharedCommands(t *testing.T) {
	if got := t3270Pad("abcdef", 5); got != "ab..." {
		t.Fatalf("t3270Pad() = %q, want %q", got, "ab...")
	}
	if got := nextStatusFilter("Todo"); got != "In Progress" {
		t.Fatalf("nextStatusFilter() = %q, want %q", got, "In Progress")
	}
	if got := extractMenuOption(" 2. scorecard "); got != "2" {
		t.Fatalf("extractMenuOption() = %q, want 2", got)
	}
	name, args, ok := recommendationToCommand("Run go test for failing Go tests")
	if !ok || name != "make" || len(args) != 1 || args[0] != "test" {
		t.Fatalf("recommendationToCommand() = (%q, %v, %v), want (make, [test], true)", name, args, ok)
	}
}

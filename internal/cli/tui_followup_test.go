package cli

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/davidl71/exarp-go/internal/models"
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
	m.configData = getConfigDefaultsForTUI()
	m.configData.Tasks.Keybindings["create_task"] = []string{"n"}

	help := m.viewHelp()
	if !strings.Contains(help, "n") {
		t.Fatal("help output should include configured create-task binding")
	}
	if strings.Contains(help, "+  Create new task inline") {
		t.Fatal("help output should not show default create-task binding after override")
	}
}

func TestViewUsesAltScreen(t *testing.T) {
	server := setupMockServer(t)
	m := initialModel(server, "", "/test", "test-project", 100, 40)

	v := m.View()
	if !v.AltScreen {
		t.Fatal("expected AltScreen to be enabled")
	}
}

func TestTaskDetailViewportSyncAndScroll(t *testing.T) {
	server := setupMockServer(t)
	m := initialModel(server, "", "/test", "test-project", 80, 16)
	m.loading = false
	m.mode = ModeTasks
	m.tasks = []*models.Todo2Task{{
		ID:              "T-100",
		Content:         "A long task",
		Status:          "Todo",
		LongDescription: strings.Repeat("viewport line content ", 60),
	}}

	detailKey := m.bindingsFor(KeyActionDetail)[0]
	updated, _, handled := m.handleActionKeys(detailKey, tea.KeyPressMsg(tea.Key{}))
	if !handled {
		t.Fatal("expected task detail action to be handled")
	}
	if updated.mode != ModeTaskDetail {
		t.Fatalf("mode = %s, want %s", updated.mode, ModeTaskDetail)
	}
	if updated.taskDetailViewport.Height() <= 0 || updated.taskDetailViewport.Width() <= 0 {
		t.Fatal("expected task detail viewport to be sized")
	}
	if updated.taskDetailViewport.TotalLineCount() == 0 {
		t.Fatal("expected viewport content to be populated")
	}

	scrolled, _, handled := updated.handleDetailOverlayKeys("down")
	if !handled {
		t.Fatal("expected viewport scroll key to be handled")
	}
	if scrolled.taskDetailViewport.YOffset() <= 0 {
		t.Fatal("expected viewport Y offset to increase after scroll")
	}
}

func TestHelpBubbleUsesConfiguredRefreshBinding(t *testing.T) {
	server := setupMockServer(t)
	m := initialModel(server, "", "/test", "test-project", 100, 40)
	m.configData.Tasks.Keybindings["refresh"] = []string{"ctrl+r"}
	m.mode = ModeTasks

	helpBubble := m.viewHelpBubble()
	if !strings.Contains(helpBubble, "ctrl+r") {
		t.Fatal("expected help bubble to include configured refresh binding")
	}
	if !strings.Contains(helpBubble, "refresh") {
		t.Fatal("expected help bubble to include refresh help text")
	}
}

func TestHandleActionKeysUsesConfiguredRefreshBinding(t *testing.T) {
	server := setupMockServer(t)
	m := initialModel(server, "", "/test", "test-project", 100, 40)
	m.loading = false
	m.mode = ModeTasks
	m.configData.Tasks.Keybindings["refresh"] = []string{"ctrl+r"}

	updated, cmd, handled := m.handleActionKeys("ctrl+r", runeKey('r'))
	if !handled {
		t.Fatal("expected configured refresh key to be handled")
	}
	if cmd == nil {
		t.Fatal("expected refresh to return a command")
	}
	if !updated.loading {
		t.Fatal("expected refresh to set loading=true")
	}
}

func TestCommandPaletteRefreshReturnsLoadCommand(t *testing.T) {
	server := setupMockServer(t)
	m := initialModel(server, "", "/test", "test-project", 100, 40)
	m.loading = false

	cmd := m.executeCommand("refresh - Reload tasks")
	if cmd == nil {
		t.Fatal("expected refresh command from command palette")
	}
	if !m.loading {
		t.Fatal("expected command palette refresh to set loading=true")
	}
}

func TestCommandPaletteQuitReturnsQuitCommand(t *testing.T) {
	server := setupMockServer(t)
	m := initialModel(server, "", "/test", "test-project", 100, 40)

	cmd := m.executeCommand("quit")
	if cmd == nil {
		t.Fatal("expected quit command")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatal("expected tea.QuitMsg from quit command")
	}
}

func TestPasteMsgUpdatesSearchState(t *testing.T) {
	server := setupMockServer(t)
	m := initialModel(server, "", "/test", "test-project", 100, 40)
	m.loading = false
	m.mode = ModeTasks
	m.tasks = createTestTasks(3)
	m.searchMode = true

	updated, _ := m.Update(tea.PasteMsg{Content: "Task 2"})
	updatedModel := updated.(model)

	if updatedModel.searchQuery != "Task 2" {
		t.Fatalf("searchQuery = %q, want %q", updatedModel.searchQuery, "Task 2")
	}
	if len(updatedModel.filteredIndices) == 0 {
		t.Fatal("expected filtered indices after paste")
	}
	if updatedModel.cursor != 0 {
		t.Fatalf("cursor = %d, want 0", updatedModel.cursor)
	}
}

func TestKeyPressMsgMovesCursor(t *testing.T) {
	server := setupMockServer(t)
	m := initialModel(server, "", "/test", "test-project", 100, 40)
	m.loading = false
	m.mode = ModeTasks
	m.tasks = createTestTasks(3)

	updated, _ := m.Update(tea.KeyPressMsg(tea.Key{Code: 'j', Text: "j"}))
	updatedModel := updated.(model)

	if updatedModel.cursor != 1 {
		t.Fatalf("cursor = %d, want 1", updatedModel.cursor)
	}
}

func TestSyncTaskComponentsUsesVisibleTasks(t *testing.T) {
	server := setupMockServer(t)
	m := initialModel(server, "", "/test", "test-project", 100, 40)
	m.loading = false
	m.mode = ModeTasks
	m.tasks = createTestTasks(3)
	m.searchQuery = "Task 2"
	m.filteredIndices = m.computeFilteredIndices()
	m.syncTaskComponents()

	if len(m.visibleTasks()) != 1 {
		t.Fatalf("visibleTasks len = %d, want 1", len(m.visibleTasks()))
	}
	if len(m.taskList.Items()) != 1 {
		t.Fatalf("taskList items = %d, want 1", len(m.taskList.Items()))
	}
	if len(m.taskTable.Rows()) != 1 {
		t.Fatalf("taskTable rows = %d, want 1", len(m.taskTable.Rows()))
	}
	if got := m.taskTable.Rows()[0][0]; got != "T-2" {
		t.Fatalf("first visible row id = %q, want %q", got, "T-2")
	}
}

func TestBubbleListDoesNotStealDetailBinding(t *testing.T) {
	server := setupMockServer(t)
	m := initialModel(server, "", "/test", "test-project", 100, 40)
	m.loading = false
	m.mode = ModeTasks
	m.tasks = createTestTasks(3)
	m.useBubbleList = true
	m.syncTaskComponents()
	detailKey := m.bindingsFor(KeyActionDetail)[0]

	updated, _ := m.Update(tea.KeyPressMsg(tea.Key{Text: detailKey}))
	updatedModel := updated.(model)

	if updatedModel.mode != ModeTaskDetail {
		t.Fatalf("mode = %s, want %s", updatedModel.mode, ModeTaskDetail)
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

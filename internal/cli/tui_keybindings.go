package cli

import (
	"strings"

	"charm.land/bubbles/v2/key"
)

const (
	KeyActionQuit              = "quit"
	KeyActionHelp              = "help"
	KeyActionRefresh           = "refresh"
	KeyActionToggleAutoRefresh = "toggle_auto_refresh"
	KeyActionSearch            = "search"
	KeyActionCreateTask        = "create_task"
	KeyActionBulkStatus        = "bulk_status"
	KeyActionStatusDone        = "status_done"
	KeyActionStatusInProgress  = "status_in_progress"
	KeyActionStatusTodo        = "status_todo"
	KeyActionStatusReview      = "status_review"
	KeyActionStatusCancelled   = "status_cancelled"
	KeyActionStatusBlocked     = "status_blocked"
	KeyActionCycleSpinner      = "cycle_spinner"
	KeyActionDetail            = "detail"
	KeyActionBack              = "back"
	KeyActionNextTab           = "next_tab"
	KeyActionPrevTab           = "prev_tab"
)

func (m model) keyMatches(key, action string) bool {
	normalizedKey := normalizeBinding(key)
	for _, binding := range m.bindingsFor(action) {
		if normalizeBinding(binding) == normalizedKey {
			return true
		}
	}

	return false
}

func (m model) bindingsFor(action string) []string {
	defaults := getDefaultTaskKeybindingsForTUI()
	if m.configData == nil || len(m.configData.Tasks.Keybindings) == 0 {
		return append([]string(nil), defaults[action]...)
	}

	if configured, ok := m.configData.Tasks.Keybindings[action]; ok && len(configured) > 0 {
		return append([]string(nil), configured...)
	}

	return append([]string(nil), defaults[action]...)
}

func bindingList(bindings []string) string {
	return strings.Join(bindings, " / ")
}

func normalizeBinding(binding string) string {
	return strings.ToLower(strings.TrimSpace(binding))
}

type tuiHelpKeyMap struct {
	short []key.Binding
	full  [][]key.Binding
}

func (k tuiHelpKeyMap) ShortHelp() []key.Binding {
	return k.short
}

func (k tuiHelpKeyMap) FullHelp() [][]key.Binding {
	return k.full
}

func helpBinding(keys []string, label, desc string) key.Binding {
	if label == "" {
		label = bindingList(keys)
	}

	return key.NewBinding(
		key.WithKeys(keys...),
		key.WithHelp(label, desc),
	)
}

func (m model) helpKeyMap() (string, tuiHelpKeyMap) {
	closeHelp := helpBinding(m.bindingsFor(KeyActionHelp), bindingList(m.bindingsFor(KeyActionHelp)), "toggle help")
	quit := helpBinding(m.bindingsFor(KeyActionQuit), bindingList(m.bindingsFor(KeyActionQuit)), "quit")

	switch m.mode {
	case ModeTasks:
		rows := [][]key.Binding{
			{
				helpBinding([]string{"up", "down", "j", "k"}, "↑↓/j/k", "move"),
				helpBinding(m.bindingsFor(KeyActionDetail), bindingList(m.bindingsFor(KeyActionDetail)), "open detail"),
				helpBinding([]string{"space"}, "Space", "select"),
			},
			{
				helpBinding(m.bindingsFor(KeyActionSearch), bindingList(m.bindingsFor(KeyActionSearch)), "search"),
				helpBinding([]string{"n", "N"}, "n/N", "next/prev match"),
				helpBinding([]string{"o", "O"}, "o/O", "sort"),
			},
			{
				helpBinding(m.bindingsFor(KeyActionStatusTodo), bindingList(m.bindingsFor(KeyActionStatusTodo)), "todo"),
				helpBinding(m.bindingsFor(KeyActionStatusInProgress), bindingList(m.bindingsFor(KeyActionStatusInProgress)), "in progress"),
				helpBinding(m.bindingsFor(KeyActionStatusDone), bindingList(m.bindingsFor(KeyActionStatusDone)), "done"),
				helpBinding(m.bindingsFor(KeyActionStatusReview), bindingList(m.bindingsFor(KeyActionStatusReview)), "review"),
			},
			{
				helpBinding(m.bindingsFor(KeyActionCreateTask), bindingList(m.bindingsFor(KeyActionCreateTask)), "new task"),
				helpBinding(m.bindingsFor(KeyActionRefresh), bindingList(m.bindingsFor(KeyActionRefresh)), "refresh"),
				helpBinding([]string{"A"}, "A", "analyze"),
			},
			{
				helpBinding([]string{"p"}, "p", "scorecard"),
				helpBinding([]string{"H"}, "H", "handoffs"),
				helpBinding([]string{"w"}, "w", "waves"),
				helpBinding([]string{"b"}, "b", "jobs"),
				helpBinding([]string{"c"}, "c", "config"),
			},
			{closeHelp, quit},
		}
		return "TASKS — Keybindings", tuiHelpKeyMap{short: rows[len(rows)-1], full: rows}
	case ModeTaskDetail:
		rows := [][]key.Binding{{
			helpBinding([]string{"up", "down", "j", "k"}, "↑↓/j/k", "scroll"),
			helpBinding([]string{"pgup", "pgdown", "ctrl+u", "ctrl+d"}, "PgUp/PgDn", "page"),
			helpBinding([]string{"g", "G"}, "g/G", "top/bottom"),
			helpBinding(append([]string{"esc"}, m.bindingsFor(KeyActionBack)...), "Esc/"+bindingList(m.bindingsFor(KeyActionBack)), "close"),
		}, {closeHelp, quit}}
		return "TASK DETAIL — Keybindings", tuiHelpKeyMap{short: rows[len(rows)-1], full: rows}
	case ModeScorecard:
		rows := [][]key.Binding{{
			helpBinding([]string{"up", "down", "j", "k"}, "↑↓/j/k", "move"),
			helpBinding(m.bindingsFor(KeyActionRefresh), bindingList(m.bindingsFor(KeyActionRefresh)), "refresh"),
			helpBinding([]string{"e"}, "e", "implement"),
			helpBinding([]string{"E"}, "E", "child agent"),
			helpBinding([]string{"p"}, "p", "back"),
		}, {closeHelp, quit}}
		return "SCORECARD — Keybindings", tuiHelpKeyMap{short: rows[len(rows)-1], full: rows}
	case ModeHandoffs:
		rows := [][]key.Binding{{
			helpBinding([]string{"up", "down", "j", "k"}, "↑↓/j/k", "move"),
			helpBinding([]string{"enter"}, "Enter", "detail"),
			helpBinding([]string{"e"}, "e", "run & close"),
			helpBinding([]string{"i"}, "i", "interactive"),
			helpBinding([]string{"a", "x", "d"}, "a/x/d", "approve/close/delete"),
			helpBinding([]string{"H"}, "H", "back"),
		}, {closeHelp, quit}}
		return "HANDOFFS — Keybindings", tuiHelpKeyMap{short: rows[len(rows)-1], full: rows}
	case ModeWaves:
		rows := [][]key.Binding{{
			helpBinding([]string{"up", "down", "j", "k"}, "↑↓/j/k", "move"),
			helpBinding([]string{"enter"}, "Enter", "expand/collapse"),
			helpBinding(m.bindingsFor(KeyActionRefresh), bindingList(m.bindingsFor(KeyActionRefresh)), "refresh"),
			helpBinding([]string{"A"}, "A", "analysis"),
			helpBinding([]string{"E"}, "E", "execute"),
			helpBinding([]string{"w"}, "w", "back"),
		}, {closeHelp, quit}}
		return "WAVES — Keybindings", tuiHelpKeyMap{short: rows[len(rows)-1], full: rows}
	case ModeConfig, ModeConfigSection:
		rows := [][]key.Binding{{
			helpBinding([]string{"up", "down", "j", "k"}, "↑↓/j/k", "move"),
			helpBinding([]string{"enter"}, "Enter", "open"),
			helpBinding([]string{"s", "u"}, "s/u", "save"),
			helpBinding(m.bindingsFor(KeyActionRefresh), bindingList(m.bindingsFor(KeyActionRefresh)), "reload"),
			helpBinding([]string{"c"}, "c", "back"),
		}, {closeHelp, quit}}
		return "CONFIG — Keybindings", tuiHelpKeyMap{short: rows[len(rows)-1], full: rows}
	case ModeTaskAnalysis:
		rows := [][]key.Binding{{
			helpBinding([]string{"y"}, "y", "write waves"),
			helpBinding([]string{"E"}, "E", "execute"),
			helpBinding([]string{"esc", "p"}, "Esc/p", "back"),
		}, {closeHelp, quit}}
		return "TASK ANALYSIS — Keybindings", tuiHelpKeyMap{short: rows[len(rows)-1], full: rows}
	case ModeJobs:
		rows := [][]key.Binding{{
			helpBinding([]string{"up", "down", "j", "k"}, "↑↓/j/k", "move"),
			helpBinding([]string{"enter"}, "Enter", "detail"),
			helpBinding([]string{"b"}, "b", "back"),
		}, {closeHelp, quit}}
		return "JOBS — Keybindings", tuiHelpKeyMap{short: rows[len(rows)-1], full: rows}
	default:
		rows := [][]key.Binding{{
			helpBinding([]string{"up", "down", "j", "k"}, "↑↓/j/k", "move"),
			helpBinding([]string{"enter"}, "Enter", "select"),
			helpBinding(m.bindingsFor(KeyActionRefresh), bindingList(m.bindingsFor(KeyActionRefresh)), "refresh"),
			quit,
		}, {closeHelp, quit}}
		return "HELP — Keybindings", tuiHelpKeyMap{short: rows[len(rows)-1], full: rows}
	}
}

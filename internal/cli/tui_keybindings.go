package cli

import (
	"strings"
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

package cli

import (
	"time"

	tea "charm.land/bubbletea/v2"
)

func (m model) View() tea.View {
	var v tea.View

	if m.showHelp {
		v = tea.NewView(m.viewHelp())
	} else if m.mode == ModeTaskDetail && m.taskDetailTask != nil {
		v = tea.NewView(m.viewTaskDetail())
	} else if m.mode == ModeConfigSection {
		v = tea.NewView(m.viewConfigSection())
	} else if m.mode == ModeConfig {
		v = tea.NewView(m.viewConfig())
	} else if m.mode == ModeScorecard {
		v = tea.NewView(m.viewScorecard())
	} else if m.mode == ModeHandoffs {
		v = tea.NewView(m.viewHandoffs())
	} else if m.mode == ModeWaves {
		v = tea.NewView(m.viewWaves())
	} else if m.mode == ModeTaskAnalysis {
		v = tea.NewView(m.viewTaskAnalysis())
	} else if m.mode == ModeJobs {
		v = tea.NewView(m.viewJobs())
	} else {
		v = tea.NewView(m.viewTasks())
	}

	// Set window title
	v.WindowTitle = m.windowTitle()

	// Hide cursor during loading
	if m.loading {
		v.Cursor = nil
	}

	// Add contextual help bubble overlay at the bottom
	helpBubble := m.viewContextualHelp()
	if helpBubble != "" {
		// Append help bubble to the bottom
		v.Content += "\n" + helpBubble
	}

	return v
}

// windowTitle returns the terminal window title based on current mode.
func (m model) windowTitle() string {
	title := "exarp-go"
	if m.projectName != "" {
		title = m.projectName + " - " + title
	}
	switch m.mode {
	case ModeTasks:
		if m.status != "" {
			title += " [" + m.status + "]"
		}
	case ModeConfig:
		title += " (Config)"
	case ModeScorecard:
		title += " (Scorecard)"
	case ModeHandoffs:
		title += " (Handoffs)"
	case ModeWaves:
		title += " (Waves)"
	case ModeTaskAnalysis:
		title += " (Analysis)"
	case ModeJobs:
		title += " (Jobs)"
	case ModeTaskDetail:
		title += " (Task Detail)"
	case ModeConfigSection:
		title += " (Config Section)"
	}
	return title
}

// updateContextualHelp updates the contextual help message based on current mode and state.
func (m *model) updateContextualHelp() {
	var help string

	switch m.mode {
	case ModeTasks:
		if m.searchMode {
			help = "Type to filter, Enter to apply, Esc to cancel"
		} else if len(m.selected) > 0 {
			help = "Space to toggle selection, D to bulk update"
		} else if m.cursor > 0 {
			help = "j/k or arrows to navigate, Enter to view, Space to select"
		} else {
			help = "Press ? for help, / to search"
		}
	case ModeConfig:
		help = "Arrow keys to navigate, Enter to edit, s to save, q to quit"
	case ModeScorecard:
		help = "r to refresh, p to return to tasks"
	case ModeHandoffs:
		help = "Enter to view, e to execute, d to delete, a to approve"
	case ModeWaves:
		help = "Enter to expand wave, o to reorder, r to refresh"
	default:
		help = ""
	}

	m.contextualHelp = help
	m.contextualHelpTime = time.Now().Unix()
}

// viewContextualHelp returns a contextual help bubble if help text is set and recent.
func (m model) viewContextualHelp() string {
	if m.contextualHelp == "" {
		return ""
	}

	// Auto-hide after 5 seconds
	if time.Now().Unix()-m.contextualHelpTime > 5 {
		return ""
	}

	// Only show if not in help mode and not loading
	if m.showHelp || m.loading {
		return ""
	}

	return softBorderStyle.Render(m.contextualHelp)
}

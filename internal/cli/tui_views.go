package cli

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
)

func (m model) View() tea.View {
	var v tea.View

	if m.mode == ModeTaskDetail && m.taskDetailTask != nil {
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

	// Overlay contextual help bubble when '?' is toggled on.
	// The bubble shows mode-specific keybindings appended below the current view.
	if m.showHelp {
		v.Content += "\n" + m.viewHelpBubble()
	}

	// Add brief contextual hint overlay at the bottom (auto-hides after 5s)
	if !m.showHelp {
		helpBubble := m.viewContextualHelp()
		if helpBubble != "" {
			v.Content += "\n" + helpBubble
		}
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

// viewHelpBubble renders a compact contextual help bubble showing the most relevant
// keybindings for the current mode/view. It is displayed as an overlay appended to the
// current view when the user toggles '?' and dismissed by '?' or Escape.
func (m model) viewHelpBubble() string {
	var b strings.Builder

	switch m.mode {
	case ModeTasks:
		b.WriteString(headerStyle.Render("TASKS — Keybindings"))
		b.WriteString("\n\n")
		b.WriteString(normalStyle.Render("Navigate"))
		b.WriteString("\n  ")
		b.WriteString(helpStyle.Render("j / k / ↑↓"))
		b.WriteString("  Move cursor    ")
		b.WriteString(helpStyle.Render("g / G"))
		b.WriteString("  First / last\n  ")
		b.WriteString(helpStyle.Render("Enter"))
		b.WriteString("        Select task    ")
		b.WriteString(helpStyle.Render("Space"))
		b.WriteString("  Toggle selection\n\n")
		b.WriteString(normalStyle.Render("Search & sort"))
		b.WriteString("\n  ")
		b.WriteString(helpStyle.Render(bindingList(m.bindingsFor(KeyActionSearch))))
		b.WriteString("  Search    ")
		b.WriteString(helpStyle.Render("n / N"))
		b.WriteString("  Next / prev match\n  ")
		b.WriteString(helpStyle.Render("f"))
		b.WriteString("  Cycle filter    ")
		b.WriteString(helpStyle.Render("o / O"))
		b.WriteString("  Cycle / flip sort\n\n")
		b.WriteString(normalStyle.Render("Task status"))
		b.WriteString("\n  ")
		b.WriteString(helpStyle.Render(bindingList(m.bindingsFor(KeyActionStatusDone))))
		b.WriteString("  Done    ")
		b.WriteString(helpStyle.Render(bindingList(m.bindingsFor(KeyActionStatusInProgress))))
		b.WriteString("  In Progress    ")
		b.WriteString(helpStyle.Render(bindingList(m.bindingsFor(KeyActionStatusTodo))))
		b.WriteString("  Todo\n  ")
		b.WriteString(helpStyle.Render(bindingList(m.bindingsFor(KeyActionStatusReview))))
		b.WriteString("  Review    ")
		b.WriteString(helpStyle.Render(bindingList(m.bindingsFor(KeyActionBulkStatus))))
		b.WriteString("  Bulk update selected\n\n")
		b.WriteString(normalStyle.Render("Actions"))
		b.WriteString("\n  ")
		b.WriteString(helpStyle.Render("s"))
		b.WriteString("  Task detail    ")
		b.WriteString(helpStyle.Render(bindingList(m.bindingsFor(KeyActionCreateTask))))
		b.WriteString("  New task    ")
		b.WriteString(helpStyle.Render(bindingList(m.bindingsFor(KeyActionRefresh))))
		b.WriteString("  Refresh\n  ")
		b.WriteString(helpStyle.Render("E"))
		b.WriteString("  Child agent    ")
		b.WriteString(helpStyle.Render("v"))
		b.WriteString("  Toggle density    ")
		b.WriteString(helpStyle.Render("A"))
		b.WriteString("  Analyze\n\n")
		b.WriteString(normalStyle.Render("Switch view"))
		b.WriteString("\n  ")
		b.WriteString(helpStyle.Render("p"))
		b.WriteString("  Scorecard    ")
		b.WriteString(helpStyle.Render("H"))
		b.WriteString("  Handoffs    ")
		b.WriteString(helpStyle.Render("w"))
		b.WriteString("  Waves\n  ")
		b.WriteString(helpStyle.Render("b"))
		b.WriteString("  Jobs        ")
		b.WriteString(helpStyle.Render("c"))
		b.WriteString("  Config\n")

	case ModeTaskDetail:
		b.WriteString(headerStyle.Render("TASK DETAIL — Keybindings"))
		b.WriteString("\n\n")
		b.WriteString(normalStyle.Render("Scroll"))
		b.WriteString("\n  ")
		b.WriteString(helpStyle.Render("j / k / ↑↓"))
		b.WriteString("  Up/down    ")
		b.WriteString(helpStyle.Render("PgUp / PgDn"))
		b.WriteString("  Page up/down\n  ")
		b.WriteString(helpStyle.Render("g / G"))
		b.WriteString("        Top / bottom\n\n")
		b.WriteString(normalStyle.Render("Close"))
		b.WriteString("\n  ")
		b.WriteString(helpStyle.Render("Esc / Enter / s"))
		b.WriteString("  Return to task list\n")

	case ModeScorecard:
		b.WriteString(headerStyle.Render("SCORECARD — Keybindings"))
		b.WriteString("\n\n")
		b.WriteString(normalStyle.Render("Navigate & act"))
		b.WriteString("\n  ")
		b.WriteString(helpStyle.Render("j / k / ↑↓"))
		b.WriteString("  Move between recommendations\n  ")
		b.WriteString(helpStyle.Render(bindingList(m.bindingsFor(KeyActionRefresh))))
		b.WriteString("  Refresh    ")
		b.WriteString(helpStyle.Render("e"))
		b.WriteString("  Implement selected\n  ")
		b.WriteString(helpStyle.Render("E"))
		b.WriteString("  Run in child agent    ")
		b.WriteString(helpStyle.Render("p"))
		b.WriteString("  Back to tasks\n")

	case ModeHandoffs:
		b.WriteString(headerStyle.Render("HANDOFFS — Keybindings"))
		b.WriteString("\n\n")
		b.WriteString(normalStyle.Render("Navigate & act"))
		b.WriteString("\n  ")
		b.WriteString(helpStyle.Render("j / k / ↑↓"))
		b.WriteString("  Move cursor    ")
		b.WriteString(helpStyle.Render("Enter"))
		b.WriteString("  View detail\n  ")
		b.WriteString(helpStyle.Render("e"))
		b.WriteString("  Execute & close    ")
		b.WriteString(helpStyle.Render("i"))
		b.WriteString("  Interactive agent\n  ")
		b.WriteString(helpStyle.Render("a"))
		b.WriteString("  Approve    ")
		b.WriteString(helpStyle.Render("x"))
		b.WriteString("  Close    ")
		b.WriteString(helpStyle.Render("d"))
		b.WriteString("  Delete\n  ")
		b.WriteString(helpStyle.Render("H"))
		b.WriteString("  Back to tasks\n")

	case ModeWaves:
		b.WriteString(headerStyle.Render("WAVES — Keybindings"))
		b.WriteString("\n\n")
		b.WriteString(normalStyle.Render("Navigate & act"))
		b.WriteString("\n  ")
		b.WriteString(helpStyle.Render("j / k / ↑↓"))
		b.WriteString("  Move cursor    ")
		b.WriteString(helpStyle.Render("Enter"))
		b.WriteString("  Expand/collapse wave\n  ")
		b.WriteString(helpStyle.Render(bindingList(m.bindingsFor(KeyActionRefresh))))
		b.WriteString("  Refresh    ")
		b.WriteString(helpStyle.Render("A"))
		b.WriteString("  Run analysis\n  ")
		b.WriteString(helpStyle.Render("E"))
		b.WriteString("  Execute in child agent    ")
		b.WriteString(helpStyle.Render("w"))
		b.WriteString("  Back to tasks\n")

	case ModeConfig, ModeConfigSection:
		b.WriteString(headerStyle.Render("CONFIG — Keybindings"))
		b.WriteString("\n\n")
		b.WriteString(normalStyle.Render("Navigate & act"))
		b.WriteString("\n  ")
		b.WriteString(helpStyle.Render("j / k / ↑↓"))
		b.WriteString("  Move between sections    ")
		b.WriteString(helpStyle.Render("Enter"))
		b.WriteString("  Open section\n  ")
		b.WriteString(helpStyle.Render("s / u"))
		b.WriteString("  Save config    ")
		b.WriteString(helpStyle.Render(bindingList(m.bindingsFor(KeyActionRefresh))))
		b.WriteString("  Reload\n  ")
		b.WriteString(helpStyle.Render("c"))
		b.WriteString("  Back to tasks\n")

	case ModeTaskAnalysis:
		b.WriteString(headerStyle.Render("TASK ANALYSIS — Keybindings"))
		b.WriteString("\n\n")
		b.WriteString(normalStyle.Render("Actions"))
		b.WriteString("\n  ")
		b.WriteString(helpStyle.Render("y"))
		b.WriteString("  Write waves plan    ")
		b.WriteString(helpStyle.Render("E"))
		b.WriteString("  Execute in child agent\n  ")
		b.WriteString(helpStyle.Render("Esc / p"))
		b.WriteString("  Return to previous view\n")

	case ModeJobs:
		b.WriteString(headerStyle.Render("JOBS — Keybindings"))
		b.WriteString("\n\n")
		b.WriteString(normalStyle.Render("Navigate & act"))
		b.WriteString("\n  ")
		b.WriteString(helpStyle.Render("j / k / ↑↓"))
		b.WriteString("  Move cursor    ")
		b.WriteString(helpStyle.Render("Enter"))
		b.WriteString("  View job detail\n  ")
		b.WriteString(helpStyle.Render("b"))
		b.WriteString("  Back to tasks\n")

	default:
		b.WriteString(headerStyle.Render("HELP — Keybindings"))
		b.WriteString("\n\n")
		b.WriteString(normalStyle.Render("Navigation"))
		b.WriteString("\n  ")
		b.WriteString(helpStyle.Render("j / k / ↑↓"))
		b.WriteString("  Move cursor    ")
		b.WriteString(helpStyle.Render("Enter"))
		b.WriteString("  Select\n\n")
		b.WriteString(normalStyle.Render("Global"))
		b.WriteString("\n  ")
		b.WriteString(helpStyle.Render(bindingList(m.bindingsFor(KeyActionRefresh))))
		b.WriteString("  Refresh    ")
		b.WriteString(helpStyle.Render(bindingList(m.bindingsFor(KeyActionQuit))))
		b.WriteString("  Quit\n")
	}

	// Footer: always-available dismiss hint
	b.WriteString("\n")
	b.WriteString(borderStyle.Render(strings.Repeat("─", 44)))
	b.WriteString("\n  ")
	b.WriteString(helpStyle.Render(bindingList(m.bindingsFor(KeyActionHelp)) + " / Esc"))
	b.WriteString("  Close help    ")
	b.WriteString(helpStyle.Render(bindingList(m.bindingsFor(KeyActionQuit))))
	b.WriteString("  Quit\n")

	return softBorderStyle.Render(b.String())
}

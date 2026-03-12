package cli

import tea "charm.land/bubbletea/v2"

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

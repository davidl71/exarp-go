package cli

func (m model) View() string {
	if m.showHelp {
		return m.viewHelp()
	}

	if m.mode == ModeTaskDetail && m.taskDetailTask != nil {
		return m.viewTaskDetail()
	}

	if m.mode == ModeConfigSection {
		return m.viewConfigSection()
	}

	if m.mode == ModeConfig {
		return m.viewConfig()
	}

	if m.mode == ModeScorecard {
		return m.viewScorecard()
	}

	if m.mode == ModeHandoffs {
		return m.viewHandoffs()
	}

	if m.mode == ModeWaves {
		return m.viewWaves()
	}

	if m.mode == ModeTaskAnalysis {
		return m.viewTaskAnalysis()
	}

	if m.mode == ModeJobs {
		return m.viewJobs()
	}

	return m.viewTasks()
}

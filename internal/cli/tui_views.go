package cli

func (m model) View() string {
	if m.showHelp {
		return m.viewHelp()
	}

	if m.mode == "taskDetail" && m.taskDetailTask != nil {
		return m.viewTaskDetail()
	}

	if m.mode == "configSection" {
		return m.viewConfigSection()
	}

	if m.mode == "config" {
		return m.viewConfig()
	}

	if m.mode == "scorecard" {
		return m.viewScorecard()
	}

	if m.mode == "handoffs" {
		return m.viewHandoffs()
	}

	if m.mode == "waves" {
		return m.viewWaves()
	}

	if m.mode == "taskAnalysis" {
		return m.viewTaskAnalysis()
	}

	if m.mode == "jobs" {
		return m.viewJobs()
	}

	return m.viewTasks()
}

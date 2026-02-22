// tui_update_navigation.go — Navigation key handlers (up/down/j/k) for TUI.
package cli

import (
	tea "github.com/charmbracelet/bubbletea"
)

// handleNavigationKeys handles up/down/j/k/pgup/pgdn/home/end/g/G navigation across all modes.
func (m model) handleNavigationKeys(key string) (model, tea.Cmd, bool) {
	switch key {
	case "up", "k":
		return m.handleUpKey()
	case "down", "j":
		return m.handleDownKey()
	case "pgup", "ctrl+u":
		return m.handlePageUpKey()
	case "pgdown", "ctrl+d":
		return m.handlePageDownKey()
	case "home", "g":
		return m.handleHomeKey()
	case "end", "G":
		return m.handleEndKey()
	}
	return m, nil, false
}

// handlePageUpKey jumps up by one page.
func (m model) handlePageUpKey() (model, tea.Cmd, bool) {
	if m.mode != ModeTasks {
		return m, nil, false
	}
	pageSize := m.taskListPageSize()
	m.cursor -= pageSize
	if m.cursor < 0 {
		m.cursor = 0
	}
	m.ensureCursorVisible()
	return m, nil, true
}

// handlePageDownKey jumps down by one page.
func (m model) handlePageDownKey() (model, tea.Cmd, bool) {
	if m.mode != ModeTasks {
		return m, nil, false
	}
	vis := m.visibleIndices()
	pageSize := m.taskListPageSize()
	m.cursor += pageSize
	if len(vis) > 0 && m.cursor >= len(vis) {
		m.cursor = len(vis) - 1
	}
	m.ensureCursorVisible()
	return m, nil, true
}

// handleHomeKey jumps to the first item.
func (m model) handleHomeKey() (model, tea.Cmd, bool) {
	if m.mode != ModeTasks {
		return m, nil, false
	}
	m.cursor = 0
	m.ensureCursorVisible()
	return m, nil, true
}

// handleEndKey jumps to the last item.
func (m model) handleEndKey() (model, tea.Cmd, bool) {
	if m.mode != ModeTasks {
		return m, nil, false
	}
	vis := m.visibleIndices()
	if len(vis) > 0 {
		m.cursor = len(vis) - 1
	}
	m.ensureCursorVisible()
	return m, nil, true
}

// handleUpKey handles up arrow and 'k' key across all modes.
func (m model) handleUpKey() (model, tea.Cmd, bool) {
	switch m.mode {
	case ModeScorecard:
		if len(m.scorecardRecs) > 0 && m.scorecardRecCursor > 0 {
			m.scorecardRecCursor--
		}
		return m, nil, true

	case ModeConfig:
		if m.configCursor > 0 {
			m.configCursor--
		}
		return m, nil, true

	case ModeTasks:
		vis := m.visibleIndices()
		if len(vis) > 0 && m.cursor > 0 {
			m.cursor--
			m.ensureCursorVisible()
		}
		return m, nil, true

	case ModeHandoffs:
		if m.handoffDetailIndex < 0 && len(m.handoffEntries) > 0 && m.handoffCursor > 0 {
			m.handoffCursor--
		}
		return m, nil, true

	case ModeWaves:
		if m.waveDetailLevel >= 0 && m.waveMoveTaskID == "" {
			// In wave detail view
			ids := m.waves[m.waveDetailLevel]
			if len(ids) > 0 && m.waveTaskCursor > 0 {
				m.waveTaskCursor--
			}
		} else if m.waveDetailLevel < 0 && len(m.waves) > 0 && m.waveCursor > 0 {
			// In wave list view
			m.waveCursor--
		}
		return m, nil, true

	case ModeJobs:
		if m.jobsDetailIndex < 0 && len(m.jobs) > 0 && m.jobsCursor > 0 {
			m.jobsCursor--
		}
		return m, nil, true
	}

	return m, nil, false
}

// handleDownKey handles down arrow and 'j' key across all modes.
func (m model) handleDownKey() (model, tea.Cmd, bool) {
	switch m.mode {
	case ModeScorecard:
		if len(m.scorecardRecs) > 0 && m.scorecardRecCursor < len(m.scorecardRecs)-1 {
			m.scorecardRecCursor++
		}
		return m, nil, true

	case ModeConfig:
		if m.configCursor < len(m.configSections)-1 {
			m.configCursor++
		}
		return m, nil, true

	case ModeTasks:
		vis := m.visibleIndices()
		if len(vis) > 0 && m.cursor < len(vis)-1 {
			m.cursor++
			m.ensureCursorVisible()
		}
		return m, nil, true

	case ModeHandoffs:
		if m.handoffDetailIndex < 0 && len(m.handoffEntries) > 0 && m.handoffCursor < len(m.handoffEntries)-1 {
			m.handoffCursor++
		}
		return m, nil, true

	case ModeWaves:
		if m.waveDetailLevel >= 0 && m.waveMoveTaskID == "" {
			// In wave detail view
			ids := m.waves[m.waveDetailLevel]
			if len(ids) > 0 && m.waveTaskCursor < len(ids)-1 {
				m.waveTaskCursor++
			}
		} else if m.waveDetailLevel < 0 && len(m.waves) > 0 {
			// In wave list view
			levels := sortedWaveLevels(m.waves)
			if m.waveCursor < len(levels)-1 {
				m.waveCursor++
			}
		}
		return m, nil, true

	case ModeJobs:
		if m.jobsDetailIndex < 0 && len(m.jobs) > 0 && m.jobsCursor < len(m.jobs)-1 {
			m.jobsCursor++
		}
		return m, nil, true
	}

	return m, nil, false
}

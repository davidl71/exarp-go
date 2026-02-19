// tui_update_navigation.go — Navigation key handlers (up/down/j/k) for TUI.
package cli

import (
	tea "github.com/charmbracelet/bubbletea"
)

// handleNavigationKeys handles up/down/j/k navigation across all modes.
func (m model) handleNavigationKeys(key string) (model, tea.Cmd, bool) {
	switch key {
	case "up", "k":
		return m.handleUpKey()
	case "down", "j":
		return m.handleDownKey()
	}
	return m, nil, false
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

// tui_update_handlers.go — Extracted keyboard and message handlers for TUI Update() method.
// This file reduces complexity in tui_update.go by extracting mode-specific logic into focused methods.
package cli

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/davidl71/exarp-go/internal/models"
	"github.com/davidl71/exarp-go/internal/tools"
)

// handleGlobalKeys handles keys that work in all modes (quit, help, esc).
func (m model) handleGlobalKeys(key string) (model, tea.Cmd, bool) {
	switch key {
	case "ctrl+c", "q":
		if m.showHelp {
			m.showHelp = false
			return m, nil, true
		}

		if m.mode == ModeConfig && m.configChanged {
			// Ask for confirmation before quitting with unsaved changes
			// For now, just quit (could add confirmation dialog later)
		}

		return m, tea.Quit, true

	case "?", "h":
		m.showHelp = !m.showHelp
		return m, nil, true

	case "esc":
		if m.showHelp {
			m.showHelp = false
			return m, nil, true
		}

		if m.searchMode {
			m.searchMode = false
			m.searchQuery = ""
			m.filteredIndices = nil
			return m, nil, true
		}

		if m.mode == ModeTaskAnalysis {
			m.mode = m.taskAnalysisReturnMode
			if m.taskAnalysisReturnMode == "" {
				m.mode = ModeTasks
			}
			return m, nil, true
		}

		if m.mode == ModeWaves {
			if m.waveDetailLevel >= 0 {
				m.waveDetailLevel = -1
			} else {
				m.mode = ModeTasks
				m.cursor = 0
			}
			return m, nil, true
		}

		if m.mode == ModeJobs {
			if m.jobsDetailIndex >= 0 {
				m.jobsDetailIndex = -1
			} else {
				m.mode = ModeTasks
				m.cursor = 0
			}
			return m, nil, true
		}

		return m, nil, true
	}

	return m, nil, false // not handled
}

// handleSearchKeys handles search mode input.
func (m model) handleSearchKeys(key string, msg tea.KeyMsg) (model, tea.Cmd, bool) {
	if !m.searchMode || m.mode != ModeTasks {
		return m, nil, false
	}

	switch key {
	case "enter":
		m.searchMode = false
		return m, nil, true

	case "backspace":
		if len(m.searchQuery) > 0 {
			m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
		}
		m.filteredIndices = m.computeFilteredIndices()
		m.cursor = 0
		m.viewportOffset = 0
		return m, nil, true

	default:
		if len(msg.String()) == 1 && msg.Type == tea.KeyRunes {
			m.searchQuery += msg.String()
		}
		m.filteredIndices = m.computeFilteredIndices()
		m.cursor = 0
		m.viewportOffset = 0
		return m, nil, true
	}
}

// handleCreateKeys handles inline task creation input.
func (m model) handleCreateKeys(key string, msg tea.KeyMsg) (model, tea.Cmd, bool) {
	if !m.createMode || m.mode != ModeTasks {
		return m, nil, false
	}

	switch key {
	case "esc":
		m.createMode = false
		m.createInput = ""
		return m, nil, true

	case "enter":
		name := strings.TrimSpace(m.createInput)
		m.createMode = false
		m.createInput = ""
		if name == "" {
			return m, nil, true
		}
		m.loading = true
		return m, createTaskCmd(m.server, name), true

	case "backspace":
		if len(m.createInput) > 0 {
			m.createInput = m.createInput[:len(m.createInput)-1]
		}
		return m, nil, true

	default:
		if len(msg.String()) == 1 && msg.Type == tea.KeyRunes {
			m.createInput += msg.String()
		}
		return m, nil, true
	}
}

// handleBulkStatusKeys handles bulk status update prompt.
func (m model) handleBulkStatusKeys(key string) (model, tea.Cmd, bool) {
	if !m.bulkStatusPrompt || m.mode != ModeTasks {
		return m, nil, false
	}

	switch key {
	case "d", "i", "t", "r":
		var newStatus string
		switch key {
		case "d":
			newStatus = models.StatusDone
		case "i":
			newStatus = models.StatusInProgress
		case "t":
			newStatus = models.StatusTodo
		case "r":
			newStatus = models.StatusReview
		}

		// Get selected task IDs
		taskIDs := []string{}
		for idx := range m.selected {
			if idx < len(m.tasks) && m.tasks[idx] != nil {
				taskIDs = append(taskIDs, m.tasks[idx].ID)
			}
		}

		if len(taskIDs) > 0 {
			m.loading = true
			m.bulkStatusPrompt = false
			return m, bulkUpdateStatusCmd(m.server, taskIDs, newStatus), true
		}

		m.bulkStatusPrompt = false
		return m, nil, true

	case "esc":
		m.bulkStatusPrompt = false
		return m, nil, true
	}

	return m, nil, true // consumed in bulk mode
}

// handleDetailOverlayKeys handles Esc/Enter/Space for detail overlays.
func (m model) handleDetailOverlayKeys(key string) (model, tea.Cmd, bool) {
	// Task detail overlay — scroll support + close
	if m.mode == ModeTaskDetail {
		switch key {
		case "esc", "enter", " ", "s":
			m.mode = ModeTasks
			m.taskDetailTask = nil
			m.taskDetailScrollTop = 0
			return m, nil, true
		case "up", "k":
			if m.taskDetailScrollTop > 0 {
				m.taskDetailScrollTop--
			}
			return m, nil, true
		case "down", "j":
			m.taskDetailScrollTop++
			return m, nil, true
		case "pgup", "ctrl+u":
			m.taskDetailScrollTop -= 10
			if m.taskDetailScrollTop < 0 {
				m.taskDetailScrollTop = 0
			}
			return m, nil, true
		case "pgdown", "ctrl+d":
			m.taskDetailScrollTop += 10
			return m, nil, true
		case "home", "g":
			m.taskDetailScrollTop = 0
			return m, nil, true
		case "end", "G":
			m.taskDetailScrollTop = 9999
			return m, nil, true
		}
	}

	// Config section detail overlay
	if m.mode == ModeConfigSection {
		switch key {
		case "esc", "enter", " ":
			m.mode = ModeConfig
			m.configSectionText = ""
			return m, nil, true
		}
	}

	// Handoff detail view
	if m.mode == ModeHandoffs && m.handoffDetailIndex >= 0 {
		switch key {
		case "esc", "enter", " ":
			m.handoffDetailIndex = -1
			return m, nil, true
		}
	}

	// Wave detail view
	if m.mode == ModeWaves && m.waveDetailLevel >= 0 {
		switch key {
		case "esc", "enter", " ":
			if m.waveMoveTaskID != "" {
				m.waveMoveTaskID = ""
				m.waveMoveMsg = ""
				return m, nil, true
			}

			m.waveUpdateMsg = ""
			m.waveDetailLevel = -1
			return m, nil, true
		}
	}

	// Job detail view
	if m.mode == ModeJobs && m.jobsDetailIndex >= 0 {
		switch key {
		case "esc", "enter", " ":
			m.jobsDetailIndex = -1
			return m, nil, true
		}
	}

	return m, nil, false
}

// handleTasksInlineStatus handles single-task status shortcuts (d/i/t/r).
func (m model) handleTasksInlineStatus(key string) (model, tea.Cmd, bool) {
	if m.mode != ModeTasks || m.searchMode {
		return m, nil, false
	}

	switch key {
	case "d", "i", "t", "r":
		vis := m.visibleIndices()
		if len(vis) > 0 && m.cursor < len(vis) {
			realIdx := m.realIndexAt(m.cursor)
			if realIdx < len(m.tasks) && m.tasks[realIdx] != nil {
				task := m.tasks[realIdx]
				var newStatus string
				switch key {
				case "d":
					newStatus = models.StatusDone
				case "i":
					newStatus = models.StatusInProgress
				case "t":
					newStatus = models.StatusTodo
				case "r":
					newStatus = models.StatusReview
				}

				if task.Status != newStatus {
					m.loading = true
					return m, updateTaskStatusCmd(m.server, task.ID, newStatus), true
				}
			}
		}
		return m, nil, true

	case "D":
		// Bulk status update: if tasks are selected, show status selection prompt
		if len(m.selected) > 0 {
			m.bulkStatusPrompt = true
		}
		return m, nil, true
	}

	return m, nil, false
}

// handleViewToggleKeys handles keys that toggle between views (p, H, b, w, c).
func (m model) handleViewToggleKeys(key string) (model, tea.Cmd, bool) {
	switch key {
	case "p":
		// Back from scorecard or task analysis to previous view
		switch m.mode {
		case ModeScorecard:
			m.mode = ModeTasks
			m.cursor = 0
		case ModeTaskAnalysis:
			m.mode = m.taskAnalysisReturnMode
			if m.taskAnalysisReturnMode == "" {
				m.mode = ModeTasks
			}
		case ModeTasks:
			m.mode = ModeScorecard
			m.scorecardLoading = true
			m.scorecardErr = nil
			m.scorecardText = ""
			return m, loadScorecard(m.projectRoot, false), true
		case ModeHandoffs:
			m.mode = ModeTasks
			m.cursor = 0
		}
		return m, nil, true

	case "H":
		// Toggle handoffs view (session handoff notes)
		if m.mode == ModeHandoffs {
			m.mode = ModeTasks
			m.cursor = 0
			return m, nil, true
		}

		m.mode = ModeHandoffs
		m.handoffLoading = true
		m.handoffErr = nil
		m.handoffText = ""
		m.handoffEntries = nil
		m.handoffCursor = 0
		m.handoffSelected = make(map[int]struct{})
		m.handoffDetailIndex = -1
		m.handoffActionMsg = ""
		return m, loadHandoffs(m.server), true

	case "b":
		// Toggle background jobs view
		if m.mode == ModeJobs {
			m.mode = ModeTasks
			m.cursor = 0
			return m, nil, true
		}

		m.mode = ModeJobs
		m.jobsCursor = 0
		m.jobsDetailIndex = -1
		return m, nil, true

	case "w":
		// Toggle waves view (dependency-order waves from backlog)
		if m.mode == ModeWaves {
			m.mode = ModeTasks
			m.cursor = 0
			m.waveDetailLevel = -1
			return m, nil, true
		}

		if m.mode == ModeTasks && len(m.tasks) > 0 {
			m.mode = ModeWaves
			m.waveDetailLevel = -1
			m.waveCursor = 0
			// Compute waves
			taskList := make([]tools.Todo2Task, 0, len(m.tasks))
			for _, t := range m.tasks {
				if t != nil {
					taskList = append(taskList, *t)
				}
			}

			waves, err := tools.ComputeWavesForTUI(m.projectRoot, taskList)
			if err != nil {
				m.waves = nil
			} else {
				m.waves = waves
			}
		}
		return m, nil, true

	case "c":
		// Toggle between tasks and config view
		switch m.mode {
		case ModeTasks:
			m.mode = ModeConfig
			m.configCursor = 0
		case ModeConfig:
			m.mode = ModeTasks
			m.cursor = 0
			m.configSaveMessage = ""
		case ModeHandoffs:
			m.mode = ModeTasks
			m.cursor = 0
		case ModeWaves, ModeJobs:
			m.mode = ModeTasks
			m.cursor = 0
		}
		return m, nil, true
	}

	return m, nil, false
}

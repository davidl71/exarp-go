// tui_update_actions.go — Action key handlers for TUI Update(): enter/space, refresh, handoffs, config, waves, child agent.
// Handles keys that trigger commands or mode-specific actions (not sort/filter, which is in tui_update_filters.go).
package cli

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/davidl71/exarp-go/internal/config"
)

// handleActionKeys handles action keys (enter, space, e, i, r, x, a, d, u, s, R, U, Q, A, y, m, 0-9, E, L) and default.
// Returns (model, cmd, true) so the caller always gets a handled result when delegating to this.
func (m model) handleActionKeys(key string, msg tea.KeyMsg) (model, tea.Cmd, bool) {
	switch key {
	case "enter", " ", "e", "i":
		// In handoffs: "i" = start interactive agent with handoff (do not close)
		if m.mode == ModeHandoffs && msg.String() == "i" && len(m.handoffEntries) > 0 {
			var h map[string]interface{}
			if m.handoffDetailIndex >= 0 && m.handoffDetailIndex < len(m.handoffEntries) {
				h = m.handoffEntries[m.handoffDetailIndex]
			} else if m.handoffCursor < len(m.handoffEntries) {
				h = m.handoffEntries[m.handoffCursor]
			}

			if h != nil {
				m.childAgentMsg = ""
				sum, _ := h["summary"].(string)

				var steps []interface{}
				if s, ok := h["next_steps"].([]interface{}); ok {
					steps = s
				}

				prompt := PromptForHandoff(sum, steps)

				return m, runChildAgentCmdInteractive(m.projectRoot, prompt, ChildAgentHandoff), true
			}

			return m, nil, true
		}

		// In handoffs: "e" = execute current handoff in agent and close it
		if m.mode == ModeHandoffs && msg.String() == "e" && len(m.handoffEntries) > 0 {
			var h map[string]interface{}
			var id string

			if m.handoffDetailIndex >= 0 && m.handoffDetailIndex < len(m.handoffEntries) {
				h = m.handoffEntries[m.handoffDetailIndex]
				id, _ = h["id"].(string)
			} else if m.handoffCursor < len(m.handoffEntries) {
				h = m.handoffEntries[m.handoffCursor]
				id, _ = h["id"].(string)
			}

			if id != "" && h != nil {
				m.childAgentMsg = ""
				sum, _ := h["summary"].(string)

				var steps []interface{}
				if s, ok := h["next_steps"].([]interface{}); ok {
					steps = s
				}

				prompt := PromptForHandoff(sum, steps)

				return m, tea.Batch(
					runChildAgentCmd(m.projectRoot, prompt, ChildAgentHandoff),
					runHandoffAction(m.server, m.projectRoot, []string{id}, "close"),
				), true
			}

			return m, nil, true
		}

		if m.mode == ModeScorecard {
			if len(m.scorecardRecs) > 0 && m.scorecardRecCursor < len(m.scorecardRecs) {
				rec := m.scorecardRecs[m.scorecardRecCursor]
				if _, _, ok := recommendationToCommand(rec); ok {
					return m, runRecommendationCmd(m.projectRoot, rec), true
				}
			}

			return m, nil, true
		}

		if m.mode == ModeConfig {
			return m, showConfigSection(m.configSections[m.configCursor], m.configData), true
		}
		if m.mode == ModeTasks {
			vis := m.visibleIndices()
			if len(vis) > 0 && m.cursor < len(vis) {
				realIdx := m.realIndexAt(m.cursor)
				if _, ok := m.selected[realIdx]; ok {
					delete(m.selected, realIdx)
				} else {
					m.selected[realIdx] = struct{}{}
				}
			}
			return m, nil, true
		}
		if m.mode == ModeHandoffs && m.handoffDetailIndex < 0 && len(m.handoffEntries) > 0 {
			if msg.String() == " " {
				if _, ok := m.handoffSelected[m.handoffCursor]; ok {
					delete(m.handoffSelected, m.handoffCursor)
				} else {
					m.handoffSelected[m.handoffCursor] = struct{}{}
				}
			} else {
				m.handoffDetailIndex = m.handoffCursor
			}
			return m, nil, true
		}
		if m.mode == ModeWaves && m.waveDetailLevel < 0 && len(m.waves) > 0 {
			levels := sortedWaveLevels(m.waves)
			if m.waveCursor < len(levels) {
				m.waveDetailLevel = levels[m.waveCursor]
				m.waveTaskCursor = 0
			}
			return m, nil, true
		}
		if m.mode == ModeJobs && m.jobsDetailIndex < 0 && len(m.jobs) > 0 {
			m.jobsDetailIndex = m.jobsCursor
		}

		return m, nil, true

	case "r":
		switch m.mode {
		case ModeConfig:
			m.configSaveMessage = ""
			cfg, err := config.LoadConfig(m.projectRoot)
			if err == nil {
				m.configData = cfg
				m.configChanged = false
			}
			return m, nil, true
		case ModeScorecard:
			m.scorecardLoading = true
			return m, loadScorecard(m.projectRoot, false), true
		case ModeHandoffs:
			m.handoffLoading = true
			return m, loadHandoffs(m.server), true
		case ModeWaves:
			m.loading = true
			return m, loadTasks(m.server, m.status), true
		case ModeTaskAnalysis:
			if !m.taskAnalysisLoading {
				m.taskAnalysisLoading = true
				action := m.taskAnalysisAction
				if action == "" {
					action = "parallelization"
				}
				return m, runTaskAnalysis(m.server, action), true
			}
			return m, nil, true
		default:
			m.loading = true
			return m, loadTasks(m.server, m.status), true
		}

	case "x":
		if m.mode == ModeHandoffs && len(m.handoffEntries) > 0 {
			ids := handoffIDsFromSelection(m)
			if len(ids) > 0 {
				return m, runHandoffAction(m.server, m.projectRoot, ids, "close"), true
			}
		}
		return m, nil, true

	case "a":
		if m.mode == ModeScorecard {
			return m, nil, true
		}
		if m.mode == ModeHandoffs && len(m.handoffEntries) > 0 {
			ids := handoffIDsFromSelection(m)
			if len(ids) > 0 {
				return m, runHandoffAction(m.server, m.projectRoot, ids, "approve"), true
			}
		}
		if m.mode == ModeTasks {
			m.autoRefresh = !m.autoRefresh
			if m.autoRefresh {
				return m, tick(), true
			}
		}
		return m, nil, true

	case "d":
		if m.mode == ModeHandoffs && len(m.handoffEntries) > 0 {
			ids := handoffIDsFromSelection(m)
			if len(ids) > 0 {
				return m, runHandoffAction(m.server, m.projectRoot, ids, "delete"), true
			}
		}
		return m, nil, true

	case "u":
		if m.mode == ModeConfig {
			return m, saveConfig(m.projectRoot, m.configData), true
		}
		return m, nil, true

	case "s":
		if m.mode == ModeScorecard {
			return m, nil, true
		}
		switch m.mode {
		case ModeTasks:
			vis := m.visibleIndices()
			if len(vis) > 0 && m.cursor < len(vis) {
				m.mode = ModeTaskDetail
				m.taskDetailTask = m.tasks[m.realIndexAt(m.cursor)]
				return m, nil, true
			}
		case ModeTaskDetail:
			m.mode = ModeTasks
			m.taskDetailTask = nil
			return m, nil, true
		default:
			return m, saveConfig(m.projectRoot, m.configData), true
		}
		return m, nil, true

	case "R":
		if m.mode == ModeWaves {
			m.loading = true
			m.err = nil
			return m, runWavesRefreshTools(m.server), true
		}
		return m, nil, true

	case "U":
		if m.mode == ModeWaves {
			m.loading = true
			m.waveUpdateMsg = ""
			return m, runReportUpdateWavesFromPlan(m.server, m.projectRoot), true
		}
		return m, nil, true

	case "Q":
		if m.mode == ModeWaves && m.queueEnabled && len(m.waves) > 0 {
			levels := sortedWaveLevels(m.waves)
			waveIdx := 0
			if m.waveDetailLevel >= 0 {
				for i, l := range levels {
					if l == m.waveDetailLevel {
						waveIdx = i
						break
					}
				}
			} else if m.waveCursor < len(levels) {
				waveIdx = m.waveCursor
			}
			if waveIdx < len(levels) {
				level := levels[waveIdx]
				m.loading = true
				m.queueEnqueueMsg = ""
				return m, runEnqueueWave(m.projectRoot, level), true
			}
		}
		return m, nil, true

	case "A":
		if m.mode == ModeTasks || m.mode == ModeWaves {
			m.taskAnalysisReturnMode = m.mode
			m.mode = ModeTaskAnalysis
			m.taskAnalysisLoading = true
			m.taskAnalysisErr = nil
			m.taskAnalysisText = ""
			m.taskAnalysisAction = "parallelization"
			m.taskAnalysisApproveMsg = ""
			m.taskAnalysisApproveLoading = false
			return m, runTaskAnalysis(m.server, "parallelization"), true
		}
		if m.mode == ModeTaskAnalysis && !m.taskAnalysisLoading {
			m.taskAnalysisLoading = true
			m.taskAnalysisApproveMsg = ""
			return m, runTaskAnalysis(m.server, m.taskAnalysisAction), true
		}
		return m, nil, true

	case "y":
		if m.mode == ModeTaskAnalysis && !m.taskAnalysisLoading && !m.taskAnalysisApproveLoading {
			m.taskAnalysisApproveLoading = true
			m.taskAnalysisApproveMsg = ""
			return m, runReportParallelExecutionPlan(m.server, m.projectRoot), true
		}
		return m, nil, true

	case "m":
		if m.mode == ModeWaves && m.waveDetailLevel >= 0 && m.waveMoveTaskID == "" {
			ids := m.waves[m.waveDetailLevel]
			if len(ids) > 0 && m.waveTaskCursor < len(ids) {
				m.waveMoveTaskID = ids[m.waveTaskCursor]
				m.waveMoveMsg = ""
			}
		}
		return m, nil, true

	case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
		if m.mode == ModeWaves && m.waveMoveTaskID != "" {
			targetLevel := int(msg.String()[0] - '0')
			levels := sortedWaveLevels(m.waves)

			if targetLevel < 0 || targetLevel >= len(levels) {
				m.waveMoveMsg = "Invalid wave number"
				return m, nil, true
			}

			level := levels[targetLevel]
			var newDeps []string
			if level == 0 {
				newDeps = nil
			} else {
				prevLevel := levels[targetLevel-1]
				prevIDs := m.waves[prevLevel]
				if len(prevIDs) == 0 {
					m.waveMoveMsg = "No tasks in previous wave"
					return m, nil, true
				}
				newDeps = []string{prevIDs[0]}
			}

			return m, moveTaskToWaveCmd(m.server, m.waveMoveTaskID, newDeps), true
		}
		return m, nil, true

	case "E":
		m.childAgentMsg = ""
		if m.mode == ModeTasks {
			vis := m.visibleIndices()
			if len(vis) > 0 && m.cursor < len(vis) {
				task := m.tasks[m.realIndexAt(m.cursor)]
				if task != nil {
					prompt := PromptForTask(task.ID, task.Content)
					return m, runChildAgentCmd(m.projectRoot, prompt, ChildAgentTask), true
				}
			}
		} else if m.mode == ModeTaskDetail && m.taskDetailTask != nil {
			prompt := PromptForTask(m.taskDetailTask.ID, m.taskDetailTask.Content)
			return m, runChildAgentCmd(m.projectRoot, prompt, ChildAgentTask), true
		} else if m.mode == ModeHandoffs {
			if m.handoffDetailIndex >= 0 && m.handoffDetailIndex < len(m.handoffEntries) {
				h := m.handoffEntries[m.handoffDetailIndex]
				sum, _ := h["summary"].(string)
				var steps []interface{}
				if s, ok := h["next_steps"].([]interface{}); ok {
					steps = s
				}
				prompt := PromptForHandoff(sum, steps)
				return m, runChildAgentCmd(m.projectRoot, prompt, ChildAgentHandoff), true
			}
			if len(m.handoffEntries) > 0 && m.handoffCursor < len(m.handoffEntries) {
				h := m.handoffEntries[m.handoffCursor]
				sum, _ := h["summary"].(string)
				var steps []interface{}
				if s, ok := h["next_steps"].([]interface{}); ok {
					steps = s
				}
				prompt := PromptForHandoff(sum, steps)
				return m, runChildAgentCmd(m.projectRoot, prompt, ChildAgentHandoff), true
			}
		} else if m.mode == ModeWaves && len(m.waves) > 0 {
			levels := sortedWaveLevels(m.waves)
			waveIdx := 0
			if m.waveDetailLevel >= 0 {
				for i, l := range levels {
					if l == m.waveDetailLevel {
						waveIdx = i
						break
					}
				}
			} else if m.waveCursor < len(levels) {
				waveIdx = m.waveCursor
			}
			if waveIdx < len(levels) {
				level := levels[waveIdx]
				ids := m.waves[level]
				prompt := PromptForWave(level, ids)
				return m, runChildAgentCmd(m.projectRoot, prompt, ChildAgentWave), true
			}
		}
		return m, nil, true

	case "L":
		m.childAgentMsg = ""
		if m.mode == ModeTasks || m.mode == ModeTaskDetail {
			prompt := PromptForPlan(m.projectRoot)
			return m, runChildAgentCmd(m.projectRoot, prompt, ChildAgentPlan), true
		}
		return m, nil, true

	default:
		if m.childAgentMsg != "" {
			m.childAgentMsg = ""
		}
		return m, nil, true
	}
}

// handoffIDsFromSelection returns handoff IDs from detail view, selected entries, or current cursor.
func handoffIDsFromSelection(m model) []string {
	if m.handoffDetailIndex >= 0 && m.handoffDetailIndex < len(m.handoffEntries) {
		if id, _ := m.handoffEntries[m.handoffDetailIndex]["id"].(string); id != "" {
			return []string{id}
		}
	}
	ids := m.handoffSelectedIDs()
	if len(ids) == 0 && m.handoffCursor < len(m.handoffEntries) {
		if id, _ := m.handoffEntries[m.handoffCursor]["id"].(string); id != "" {
			return []string{id}
		}
	}
	return ids
}

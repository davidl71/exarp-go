// tui_update.go — TUI Update() method: all Bubbletea message handling.
package cli

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/davidl71/exarp-go/internal/tools"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Update terminal dimensions (SIGWINCH on Unix triggers this via Bubble Tea).
		// Clamp to minimums so iTerm2 (or other terminals that report 0/stale size)
		// don't break window size or task detail ("s") layout.
		m.width = msg.Width
		m.height = msg.Height

		if m.width < minTermWidth {
			m.width = minTermWidth
		}

		if m.height < minTermHeight {
			m.height = minTermHeight
		}

		return m, nil

	case taskLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}

		m.tasks = msg.tasks
		m.computeHierarchyOrder() // always compute so hierarchy depth/order available

		if m.sortOrder == SortByHierarchy && len(m.hierarchyOrder) > 0 {
			// use hierarchy order (parent then children, with depth)
		} else {
			sortTasksBy(m.tasks, m.sortOrder, m.sortAsc)
		}

		if m.searchQuery != "" {
			m.filteredIndices = m.computeFilteredIndices()
		} else {
			m.filteredIndices = nil
		}

		vis := m.visibleIndices()
		if len(vis) > 0 && m.cursor >= len(vis) {
			m.cursor = len(vis) - 1
		}
		m.ensureCursorVisible()

		m.lastUpdate = time.Now()
		// If in waves view, recompute waves (prefer docs/PARALLEL_EXECUTION_PLAN_RESEARCH.md)
		if m.mode == ModeWaves && len(m.tasks) > 0 {
			taskList := make([]tools.Todo2Task, 0, len(m.tasks))

			for _, t := range m.tasks {
				if t != nil {
					taskList = append(taskList, *t)
				}
			}

			waves, err := tools.ComputeWavesForTUI(m.projectRoot, taskList)
			if err == nil {
				m.waves = waves
				if m.waveDetailLevel >= 0 {
					if ids := m.waves[m.waveDetailLevel]; len(ids) > 0 {
						if m.waveTaskCursor >= len(ids) {
							m.waveTaskCursor = len(ids) - 1
						}
					} else {
						m.waveTaskCursor = 0
					}
				}
			} else {
				m.waves = nil
			}
		}
		// Continue auto-refresh if enabled
		if m.autoRefresh {
			return m, tick()
		}

		return m, nil

	case wavesRefreshDoneMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		// Reload tasks so waves recompute from updated backlog
		m.loading = true

		return m, loadTasks(m.server, m.status)

	case statusUpdateDoneMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.loading = true
		return m, loadTasks(m.server, m.status)

	case taskCreatedMsg:
		m.loading = false
		if msg.err != nil {
			m.childAgentMsg = fmt.Sprintf("Create failed: %v", msg.err)
		} else if msg.taskID != "" {
			m.childAgentMsg = fmt.Sprintf("Created %s", msg.taskID)
		} else {
			m.childAgentMsg = "Task created"
		}
		m.loading = true
		return m, loadTasks(m.server, m.status)

	case bulkStatusUpdateDoneMsg:
		m.loading = false
		if msg.err != nil {
			m.bulkStatusMsg = fmt.Sprintf("Bulk update failed: %v", msg.err)
		} else {
			m.bulkStatusMsg = fmt.Sprintf("Updated %d/%d tasks", msg.updated, msg.total)
			// Clear selection after successful bulk update
			m.selected = make(map[int]struct{})
		}
		m.loading = true
		return m, loadTasks(m.server, m.status)

	case taskAnalysisLoadedMsg:
		m.taskAnalysisLoading = false
		m.taskAnalysisErr = msg.err
		m.taskAnalysisAction = msg.action

		if msg.err == nil {
			m.taskAnalysisText = msg.text
		} else {
			m.taskAnalysisText = ""
		}

		return m, nil

	case taskAnalysisApproveDoneMsg:
		m.taskAnalysisApproveLoading = false
		if msg.err != nil {
			m.taskAnalysisApproveMsg = "Error: " + msg.err.Error()
		} else {
			m.taskAnalysisApproveMsg = msg.message
		}

		return m, nil

	case moveTaskToWaveDoneMsg:
		m.waveMoveTaskID = ""
		if msg.err != nil {
			m.waveMoveMsg = "Error: " + msg.err.Error()
		} else {
			m.waveMoveMsg = "Moved " + msg.taskID + " to wave"
		}

		return m, loadTasks(m.server, m.status)

	case updateWavesFromPlanDoneMsg:
		m.loading = false
		m.waveUpdateMsg = ""

		if msg.err != nil {
			m.waveUpdateMsg = "Error: " + msg.err.Error()
		} else if msg.message != "" {
			m.waveUpdateMsg = msg.message
		}

		return m, loadTasks(m.server, m.status)

	case enqueueWaveDoneMsg:
		m.loading = false
		if msg.err != nil {
			m.queueEnqueueMsg = "Enqueue error: " + msg.err.Error()
		} else {
			m.queueEnqueueMsg = fmt.Sprintf("Enqueued %d tasks from wave %d to Redis", msg.enqueued, msg.waveLevel)
		}
		return m, nil

	case tickMsg:
		// Auto-refresh tasks periodically (only in tasks mode)
		if m.mode == ModeTasks && m.autoRefresh && !m.loading {
			m.loading = true
			return m, loadTasks(m.server, m.status)
		}
		// Continue ticking
		if m.mode == ModeTasks && m.autoRefresh {
			return m, tick()
		}

		return m, nil

	case configSavedMsg:
		m.configChanged = false
		return m, nil

	case configSectionDetailMsg:
		m.mode = ModeConfigSection
		m.configSectionText = msg.text

		return m, nil

	case configSaveResultMsg:
		m.configSaveMessage = msg.message
		m.configSaveSuccess = msg.success

		if msg.success {
			m.configChanged = false
		}

		return m, nil

	case scorecardLoadedMsg:
		m.scorecardLoading = false
		m.scorecardErr = msg.err
		m.scorecardRunOutput = ""

		if msg.err == nil {
			m.scorecardText = msg.text
			m.scorecardRecs = msg.recommendations

			if len(m.scorecardRecs) > 0 {
				m.scorecardRecCursor = 0
			}
		}

		return m, nil

	case handoffLoadedMsg:
		m.handoffLoading = false
		m.handoffErr = msg.err

		if msg.err == nil {
			m.handoffText = msg.text
			m.handoffEntries = msg.entries

			if len(m.handoffEntries) > 0 {
				if m.handoffCursor >= len(m.handoffEntries) {
					m.handoffCursor = len(m.handoffEntries) - 1
				}
			} else {
				m.handoffCursor = 0
			}
			// Keep selection only for indices that still exist
			newSel := make(map[int]struct{})

			for i := range m.handoffSelected {
				if i >= 0 && i < len(m.handoffEntries) {
					newSel[i] = struct{}{}
				}
			}

			m.handoffSelected = newSel
		}

		return m, nil

	case handoffActionDoneMsg:
		if msg.err != nil {
			m.handoffActionMsg = fmt.Sprintf("%s failed: %v", msg.action, msg.err)
		} else {
			verb := msg.action + "d"
			switch msg.action {
			case "delete":
				verb = "deleted"
			case "close":
				verb = "closed"
			case "approve":
				verb = "approved"
			}

			m.handoffActionMsg = fmt.Sprintf("%d handoff(s) %s.", msg.updated, verb)
			m.handoffSelected = make(map[int]struct{})
			m.handoffDetailIndex = -1 // return to list after close/approve/delete
		}

		return m, loadHandoffs(m.server)

	case runRecommendationResultMsg:
		if msg.err != nil {
			m.scorecardRunOutput = "Error: " + msg.err.Error()
		} else {
			m.scorecardRunOutput = strings.TrimSpace(msg.output)
			if m.scorecardRunOutput == "" {
				m.scorecardRunOutput = "(command completed)"
			}
		}
		// Refresh scorecard with full checks so updated state (e.g. coverage after make test) is shown
		m.scorecardLoading = true

		return m, loadScorecard(m.projectRoot, true)

	case childAgentResultMsg:
		if msg.Result.Launched {
			m.childAgentMsg = msg.Result.Message
			m.jobs = append(m.jobs, BackgroundJob{
				Kind:      msg.Result.Kind,
				Prompt:    msg.Result.Prompt,
				StartedAt: time.Now(),
				Pid:       msg.Result.Pid,
			})
		} else {
			m.childAgentMsg = "Child agent: " + msg.Result.Message
		}

		return m, nil

	case jobCompletedMsg:
		for i := range m.jobs {
			if m.jobs[i].Pid == msg.Pid {
				m.jobs[i].Output = msg.Output
				if msg.Err != nil {
					m.jobs[i].Output += "\n(exit: " + msg.Err.Error() + ")"
				}

				break
			}
		}

		return m, nil

	case tea.KeyMsg:
		key := msg.String()

		// Clear transient feedback messages on any keypress so they don't persist
		m.clearTransientMessages()

		// Try global keys first (quit, help, esc)
		if newModel, cmd, handled := m.handleGlobalKeys(key); handled {
			return newModel, cmd
		}

		// When help is open, ignore all other keys
		if m.showHelp {
			return m, nil
		}

		// Try search mode keys
		if newModel, cmd, handled := m.handleSearchKeys(key, msg); handled {
			return newModel, cmd
		}

		// Try inline task creation keys
		if newModel, cmd, handled := m.handleCreateKeys(key, msg); handled {
			return newModel, cmd
		}

		// Try bulk status update keys
		if newModel, cmd, handled := m.handleBulkStatusKeys(key); handled {
			return newModel, cmd
		}

		// Try inline status change keys (d/i/t/r/D)
		if newModel, cmd, handled := m.handleTasksInlineStatus(key); handled {
			return newModel, cmd
		}

		// Try detail overlay close keys (esc/enter/space for overlays)
		if newModel, cmd, handled := m.handleDetailOverlayKeys(key); handled {
			return newModel, cmd
		}

		// Try view toggle keys (p, H, b, w, c)
		if newModel, cmd, handled := m.handleViewToggleKeys(key); handled {
			return newModel, cmd
		}

		// Try navigation keys (up/down/j/k)
		if newModel, cmd, handled := m.handleNavigationKeys(key); handled {
			return newModel, cmd
		}

		// Sort/filter keys (o, O, /, n, N, tab)
		if newModel, cmd, handled := m.handleSortFilterKeys(key); handled {
			return newModel, cmd
		}

		// Action keys (enter, space, e, i, r, x, a, d, u, s, R, U, Q, A, y, m, 0-9, E, L, default)
		newModel, cmd, _ := m.handleActionKeys(key, msg)
		return newModel, cmd
	}

	return m, nil
}

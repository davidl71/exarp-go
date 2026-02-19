// tui_update.go — TUI Update() method: all Bubbletea message handling.
package cli

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/davidl71/exarp-go/internal/config"
	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/models"
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

		if m.sortOrder == "hierarchy" && len(m.hierarchyOrder) > 0 {
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

		m.lastUpdate = time.Now()
		// If in waves view, recompute waves (prefer docs/PARALLEL_EXECUTION_PLAN_RESEARCH.md)
		if m.mode == "waves" && len(m.tasks) > 0 {
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

		return m, loadTasks(m.status)

	case statusUpdateDoneMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.loading = true
		return m, loadTasks(m.status)

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

		return m, loadTasks(m.status)

	case updateWavesFromPlanDoneMsg:
		m.loading = false
		m.waveUpdateMsg = ""

		if msg.err != nil {
			m.waveUpdateMsg = "Error: " + msg.err.Error()
		} else if msg.message != "" {
			m.waveUpdateMsg = msg.message
		}

		return m, loadTasks(m.status)

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
		if m.mode == "tasks" && m.autoRefresh && !m.loading {
			m.loading = true
			return m, loadTasks(m.status)
		}
		// Continue ticking
		if m.mode == "tasks" && m.autoRefresh {
			return m, tick()
		}

		return m, nil

	case configSavedMsg:
		m.configChanged = false
		return m, nil

	case configSectionDetailMsg:
		m.mode = "configSection"
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
		switch msg.String() {
		case "ctrl+c", "q":
			if m.showHelp {
				m.showHelp = false
				return m, nil
			}

			if m.mode == "config" && m.configChanged {
				// Ask for confirmation before quitting with unsaved changes
				// For now, just quit (could add confirmation dialog later)
			}

			return m, tea.Quit

		case "?", "h":
			m.showHelp = !m.showHelp
			return m, nil

		case "esc":
			if m.showHelp {
				m.showHelp = false
				return m, nil
			}

			if m.searchMode {
				m.searchMode = false
				m.searchQuery = ""
				m.filteredIndices = nil

				return m, nil
			}

			if m.mode == "taskAnalysis" {
				m.mode = m.taskAnalysisReturnMode
				if m.taskAnalysisReturnMode == "" {
					m.mode = "tasks"
				}

				return m, nil
			}

			if m.mode == "waves" {
				if m.waveDetailLevel >= 0 {
					m.waveDetailLevel = -1
				} else {
					m.mode = "tasks"
					m.cursor = 0
				}

				return m, nil
			}

			if m.mode == "jobs" {
				if m.jobsDetailIndex >= 0 {
					m.jobsDetailIndex = -1
				} else {
					m.mode = "tasks"
					m.cursor = 0
				}

				return m, nil
			}

			return m, nil
		}

		// When help is open, ignore all other keys (handled above)
		if m.showHelp {
			return m, nil
		}

		// Search mode: accept input, Enter to apply, Esc already handled above
		if m.searchMode && m.mode == "tasks" {
			switch msg.String() {
			case "enter":
				m.searchMode = false
				m.filteredIndices = m.computeFilteredIndices()

				if len(m.visibleIndices()) > 0 && m.cursor >= len(m.visibleIndices()) {
					m.cursor = len(m.visibleIndices()) - 1
				}

				return m, nil
			case "backspace":
				if len(m.searchQuery) > 0 {
					m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
				}

				return m, nil
			default:
				if len(msg.String()) == 1 && msg.Type == tea.KeyRunes {
					m.searchQuery += msg.String()
				}

				return m, nil
			}
		}

		// Inline status change in task list: d=Done, i=In Progress, t=Todo
		if m.mode == "tasks" && !m.searchMode {
			switch msg.String() {
			case "d", "i", "t":
				vis := m.visibleIndices()
				if len(vis) > 0 && m.cursor < len(vis) {
					realIdx := m.realIndexAt(m.cursor)
					if realIdx < len(m.tasks) && m.tasks[realIdx] != nil {
						task := m.tasks[realIdx]
						var newStatus string
						switch msg.String() {
						case "d":
							newStatus = models.StatusDone
						case "i":
							newStatus = models.StatusInProgress
						case "t":
							newStatus = models.StatusTodo
						}
						if task.Status != newStatus {
							m.loading = true
							return m, updateTaskStatusCmd(task.ID, newStatus)
						}
					}
				}
				return m, nil
			}
		}

		// When task detail is open, Esc / Enter / Space close it
		if m.mode == "taskDetail" {
			switch msg.String() {
			case "esc", "enter", " ":
				m.mode = "tasks"
				m.taskDetailTask = nil

				return m, nil
			}
		}

		// When config section detail is open, Esc / Enter / Space close it
		if m.mode == "configSection" {
			switch msg.String() {
			case "esc", "enter", " ":
				m.mode = "config"
				m.configSectionText = ""

				return m, nil
			}
		}

		// When handoff detail is open, Esc / Enter / Space close it
		if m.mode == "handoffs" && m.handoffDetailIndex >= 0 {
			switch msg.String() {
			case "esc", "enter", " ":
				m.handoffDetailIndex = -1
				return m, nil
			}
		}

		// When wave detail is open (viewing tasks for a wave), Esc / Enter / Space close it (or cancel move)
		if m.mode == "waves" && m.waveDetailLevel >= 0 {
			switch msg.String() {
			case "esc", "enter", " ":
				if m.waveMoveTaskID != "" {
					m.waveMoveTaskID = ""
					m.waveMoveMsg = ""

					return m, nil
				}

				m.waveUpdateMsg = ""
				m.waveDetailLevel = -1

				return m, nil
			}
		}

		// When job detail is open, Esc / Enter / Space close it
		if m.mode == "jobs" && m.jobsDetailIndex >= 0 {
			switch msg.String() {
			case "esc", "enter", " ":
				m.jobsDetailIndex = -1
				return m, nil
			}
		}

		switch msg.String() {
		case "p":
			// Back from scorecard or task analysis to previous view
			switch m.mode {
			case "scorecard":
				m.mode = "tasks"
				m.cursor = 0
			case "taskAnalysis":
				m.mode = m.taskAnalysisReturnMode
				if m.taskAnalysisReturnMode == "" {
					m.mode = "tasks"
				}
			case "tasks":
				m.mode = "scorecard"
				m.scorecardLoading = true
				m.scorecardErr = nil
				m.scorecardText = ""

				return m, loadScorecard(m.projectRoot, false)
			case "handoffs":
				m.mode = "tasks"
				m.cursor = 0
			}

			return m, nil

		case "H":
			// Toggle handoffs view (session handoff notes)
			if m.mode == "handoffs" {
				m.mode = "tasks"
				m.cursor = 0

				return m, nil
			}

			m.mode = "handoffs"
			m.handoffLoading = true
			m.handoffErr = nil
			m.handoffText = ""
			m.handoffEntries = nil
			m.handoffCursor = 0
			m.handoffSelected = make(map[int]struct{})
			m.handoffDetailIndex = -1
			m.handoffActionMsg = ""

			return m, loadHandoffs(m.server)

		case "b":
			// Toggle background jobs view
			if m.mode == "jobs" {
				m.mode = "tasks"
				m.cursor = 0

				return m, nil
			}

			m.mode = "jobs"
			m.jobsCursor = 0
			m.jobsDetailIndex = -1

			return m, nil

		case "w":
			// Toggle waves view (dependency-order waves from backlog)
			if m.mode == "waves" {
				m.mode = "tasks"
				m.cursor = 0
				m.waveDetailLevel = -1

				return m, nil
			}

			if m.mode == "tasks" && len(m.tasks) > 0 {
				m.mode = "waves"
				m.waveDetailLevel = -1
				m.waveCursor = 0
				// Compute waves (prefer docs/PARALLEL_EXECUTION_PLAN_RESEARCH.md)
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

			return m, nil

		case "c":
			// Toggle between tasks and config view
			switch m.mode {
			case "tasks":
				m.mode = "config"
				m.configCursor = 0
			case "config":
				m.mode = "tasks"
				m.cursor = 0
				m.configSaveMessage = ""
			case "handoffs":
				m.mode = "tasks"
				m.cursor = 0
			case "waves", "jobs":
				m.mode = "tasks"
				m.cursor = 0
			}

			return m, nil

		case "up", "k":
			if m.mode == "scorecard" {
				if len(m.scorecardRecs) > 0 && m.scorecardRecCursor > 0 {
					m.scorecardRecCursor--
				}

				return m, nil
			}

			if m.mode == "config" {
				if m.configCursor > 0 {
					m.configCursor--
				}
			} else if m.mode == "tasks" {
				vis := m.visibleIndices()
				if len(vis) > 0 && m.cursor > 0 {
					m.cursor--
				}
			} else if m.mode == "handoffs" && m.handoffDetailIndex < 0 && len(m.handoffEntries) > 0 && m.handoffCursor > 0 {
				m.handoffCursor--
			} else if m.mode == "waves" && m.waveDetailLevel >= 0 && m.waveMoveTaskID == "" {
				ids := m.waves[m.waveDetailLevel]
				if len(ids) > 0 && m.waveTaskCursor > 0 {
					m.waveTaskCursor--
				}
			} else if m.mode == "waves" && m.waveDetailLevel < 0 && len(m.waves) > 0 && m.waveCursor > 0 {
				m.waveCursor--
			} else if m.mode == "jobs" && m.jobsDetailIndex < 0 && len(m.jobs) > 0 && m.jobsCursor > 0 {
				m.jobsCursor--
			}

			return m, nil

		case "down", "j":
			if m.mode == "scorecard" {
				if len(m.scorecardRecs) > 0 && m.scorecardRecCursor < len(m.scorecardRecs)-1 {
					m.scorecardRecCursor++
				}

				return m, nil
			}

			if m.mode == "config" {
				if m.configCursor < len(m.configSections)-1 {
					m.configCursor++
				}
			} else if m.mode == "tasks" {
				vis := m.visibleIndices()
				if len(vis) > 0 && m.cursor < len(vis)-1 {
					m.cursor++
				}
			} else if m.mode == "handoffs" && m.handoffDetailIndex < 0 && len(m.handoffEntries) > 0 && m.handoffCursor < len(m.handoffEntries)-1 {
				m.handoffCursor++
			} else if m.mode == "waves" && m.waveDetailLevel >= 0 && m.waveMoveTaskID == "" {
				ids := m.waves[m.waveDetailLevel]
				if len(ids) > 0 && m.waveTaskCursor < len(ids)-1 {
					m.waveTaskCursor++
				}
			} else if m.mode == "waves" && m.waveDetailLevel < 0 && len(m.waves) > 0 {
				levels := sortedWaveLevels(m.waves)
				if m.waveCursor < len(levels)-1 {
					m.waveCursor++
				}
			} else if m.mode == "jobs" && m.jobsDetailIndex < 0 && len(m.jobs) > 0 && m.jobsCursor < len(m.jobs)-1 {
				m.jobsCursor++
			}

			return m, nil

		case "enter", " ", "e", "i":
			// In handoffs: "i" = start interactive agent with handoff (do not close)
			if m.mode == "handoffs" && msg.String() == "i" && len(m.handoffEntries) > 0 {
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

					return m, runChildAgentCmdInteractive(m.projectRoot, prompt, ChildAgentHandoff)
				}

				return m, nil
			}

			// In handoffs: "e" = execute current handoff in agent and close it
			if m.mode == "handoffs" && msg.String() == "e" && len(m.handoffEntries) > 0 {
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
					)
				}

				return m, nil
			}

			if m.mode == "scorecard" {
				if len(m.scorecardRecs) > 0 && m.scorecardRecCursor < len(m.scorecardRecs) {
					rec := m.scorecardRecs[m.scorecardRecCursor]
					if _, _, ok := recommendationToCommand(rec); ok {
						return m, runRecommendationCmd(m.projectRoot, rec)
					}
				}

				return m, nil
			}

			if m.mode == "config" {
				// Open config section editor (for now, just show section details)
				return m, showConfigSection(m.configSections[m.configCursor], m.configData)
			} else if m.mode == "tasks" {
				// Toggle selection (cursor is index into visible list)
				vis := m.visibleIndices()
				if len(vis) > 0 && m.cursor < len(vis) {
					realIdx := m.realIndexAt(m.cursor)
					if _, ok := m.selected[realIdx]; ok {
						delete(m.selected, realIdx)
					} else {
						m.selected[realIdx] = struct{}{}
					}
				}
			} else if m.mode == "handoffs" && m.handoffDetailIndex < 0 && len(m.handoffEntries) > 0 {
				if msg.String() == " " {
					// Space: toggle selection
					if _, ok := m.handoffSelected[m.handoffCursor]; ok {
						delete(m.handoffSelected, m.handoffCursor)
					} else {
						m.handoffSelected[m.handoffCursor] = struct{}{}
					}
				} else {
					// Enter or e: open detail
					m.handoffDetailIndex = m.handoffCursor
				}
			} else if m.mode == "waves" && m.waveDetailLevel < 0 && len(m.waves) > 0 {
				levels := sortedWaveLevels(m.waves)
				if m.waveCursor < len(levels) {
					m.waveDetailLevel = levels[m.waveCursor]
					m.waveTaskCursor = 0
				}
			} else if m.mode == "jobs" && m.jobsDetailIndex < 0 && len(m.jobs) > 0 {
				m.jobsDetailIndex = m.jobsCursor
			}

			return m, nil

		case "r":
			switch m.mode {
			case "config":
				// Reload config
				m.configSaveMessage = ""

				cfg, err := config.LoadConfig(m.projectRoot)
				if err == nil {
					m.configData = cfg
					m.configChanged = false
				}

				return m, nil
			case "scorecard":
				// Refresh scorecard (fast mode for manual refresh; use Run then Enter for full refresh)
				m.scorecardLoading = true
				return m, loadScorecard(m.projectRoot, false)
			case "handoffs":
				// Refresh handoffs
				m.handoffLoading = true
				return m, loadHandoffs(m.server)
			case "waves":
				// Refresh tasks (waves recompute on taskLoadedMsg)
				m.loading = true
				return m, loadTasks(m.status)
			case "taskAnalysis":
				if !m.taskAnalysisLoading {
					m.taskAnalysisLoading = true

					action := m.taskAnalysisAction
					if action == "" {
						action = "parallelization"
					}

					return m, runTaskAnalysis(m.server, action)
				}

				return m, nil
			default:
				// Refresh tasks
				m.loading = true
				return m, loadTasks(m.status)
			}

		case "x":
			// Close/dismiss handoffs: from detail view (current item) or list view (selected or current)
			if m.mode == "handoffs" && len(m.handoffEntries) > 0 {
				var ids []string

				if m.handoffDetailIndex >= 0 && m.handoffDetailIndex < len(m.handoffEntries) {
					if id, _ := m.handoffEntries[m.handoffDetailIndex]["id"].(string); id != "" {
						ids = []string{id}
					}
				}

				if len(ids) == 0 {
					ids = m.handoffSelectedIDs()
					if len(ids) == 0 && m.handoffCursor < len(m.handoffEntries) {
						if id, _ := m.handoffEntries[m.handoffCursor]["id"].(string); id != "" {
							ids = []string{id}
						}
					}
				}

				if len(ids) > 0 {
					return m, runHandoffAction(m.server, m.projectRoot, ids, "close")
				}
			}

			return m, nil

		case "a":
			if m.mode == "scorecard" {
				return m, nil
			}

			if m.mode == "handoffs" && len(m.handoffEntries) > 0 {
				// Approve: from detail view (current item) or list view (selected or current)
				var ids []string

				if m.handoffDetailIndex >= 0 && m.handoffDetailIndex < len(m.handoffEntries) {
					if id, _ := m.handoffEntries[m.handoffDetailIndex]["id"].(string); id != "" {
						ids = []string{id}
					}
				}

				if len(ids) == 0 {
					ids = m.handoffSelectedIDs()
					if len(ids) == 0 && m.handoffCursor < len(m.handoffEntries) {
						if id, _ := m.handoffEntries[m.handoffCursor]["id"].(string); id != "" {
							ids = []string{id}
						}
					}
				}

				if len(ids) > 0 {
					return m, runHandoffAction(m.server, m.projectRoot, ids, "approve")
				}
			}

			if m.mode == "tasks" {
				// Toggle auto-refresh
				m.autoRefresh = !m.autoRefresh
				if m.autoRefresh {
					return m, tick()
				}
			}

			return m, nil

		case "d":
			// Delete handoffs: from detail view (current item) or list view (selected or current)
			if m.mode == "handoffs" && len(m.handoffEntries) > 0 {
				var ids []string

				if m.handoffDetailIndex >= 0 && m.handoffDetailIndex < len(m.handoffEntries) {
					if id, _ := m.handoffEntries[m.handoffDetailIndex]["id"].(string); id != "" {
						ids = []string{id}
					}
				}

				if len(ids) == 0 {
					ids = m.handoffSelectedIDs()
					if len(ids) == 0 && m.handoffCursor < len(m.handoffEntries) {
						if id, _ := m.handoffEntries[m.handoffCursor]["id"].(string); id != "" {
							ids = []string{id}
						}
					}
				}

				if len(ids) > 0 {
					return m, runHandoffAction(m.server, m.projectRoot, ids, "delete")
				}
			}

			return m, nil

		case "o":
			// Cycle sort order (id → status → priority → updated → hierarchy → id)
			if m.mode == "tasks" && len(m.tasks) > 0 {
				switch m.sortOrder {
				case "id":
					m.sortOrder = "status"
				case "status":
					m.sortOrder = "priority"
				case "priority":
					m.sortOrder = "updated"
				case "updated":
					m.sortOrder = "hierarchy"
				default:
					m.sortOrder = "id"
				}

				if m.sortOrder == "hierarchy" {
					m.computeHierarchyOrder()
				} else {
					sortTasksBy(m.tasks, m.sortOrder, m.sortAsc)
				}

				if m.cursor >= len(m.visibleIndices()) {
					m.cursor = len(m.visibleIndices()) - 1
				}
			}

			return m, nil

		case "O":
			// Toggle sort direction (asc ↔ desc)
			if m.mode == "tasks" && len(m.tasks) > 0 {
				m.sortAsc = !m.sortAsc
				if m.sortOrder == "hierarchy" {
					m.computeHierarchyOrder()
				} else {
					sortTasksBy(m.tasks, m.sortOrder, m.sortAsc)
				}
			}

			return m, nil

		case "/":
			// Start search/filter (vim-style)
			if m.mode == "tasks" {
				m.searchMode = true
				// Keep previous searchQuery so user can extend or backspace
			}

			return m, nil

		case "n":
			// Next search match (vim-style)
			if m.mode == "tasks" && m.searchQuery != "" {
				vis := m.visibleIndices()
				if len(vis) > 0 && m.cursor < len(vis)-1 {
					m.cursor++
				}
			}

			return m, nil

		case "N":
			// Previous search match (vim-style)
			if m.mode == "tasks" && m.searchQuery != "" {
				if m.cursor > 0 {
					m.cursor--
				}
			}

			return m, nil

		case "tab", "\t":
			// In tasks mode: collapse/expand tree node under cursor (if it has children)
			if m.mode == "tasks" {
				vis := m.visibleIndices()
				if len(vis) > 0 && m.cursor < len(vis) {
					realIdx := m.realIndexAt(m.cursor)

					task := m.tasks[realIdx]
					if task != nil && m.taskHasChildren(task.ID) {
						if _, ok := m.collapsedTaskIDs[task.ID]; ok {
							delete(m.collapsedTaskIDs, task.ID)
						} else {
							m.collapsedTaskIDs[task.ID] = struct{}{}
						}
					}
				}
			}

			return m, nil

		case "u":
			// In config view: update (save current config to .exarp/config.pb protobuf)
			if m.mode == "config" {
				return m, saveConfig(m.projectRoot, m.configData)
			}

			return m, nil

		case "s":
			if m.mode == "scorecard" {
				return m, nil
			}

			switch m.mode {
			case "tasks":
				// Show task details in-TUI (word-wrapped)
				vis := m.visibleIndices()
				if len(vis) > 0 && m.cursor < len(vis) {
					m.mode = "taskDetail"
					m.taskDetailTask = m.tasks[m.realIndexAt(m.cursor)]

					return m, nil
				}
			case "taskDetail":
				// Close task detail on 's' too (so same key can close)
				m.mode = "tasks"
				m.taskDetailTask = nil

				return m, nil
			default:
				// Save config (writes to .exarp/config.pb protobuf)
				return m, saveConfig(m.projectRoot, m.configData)
			}

		case "R":
			// In waves view: run exarp tools (task_workflow sync, task_analysis parallelization) then refresh waves
			if m.mode == "waves" {
				m.loading = true
				m.err = nil

				return m, runWavesRefreshTools(m.server)
			}

			return m, nil

		case "U":
			// In waves view: update Todo2 task dependencies from docs/PARALLEL_EXECUTION_PLAN_RESEARCH.md
			if m.mode == "waves" {
				m.loading = true
				m.waveUpdateMsg = ""

				return m, runReportUpdateWavesFromPlan(m.server, m.projectRoot)
			}

			return m, nil

		case "Q":
			if m.mode == "waves" && m.queueEnabled && len(m.waves) > 0 {
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
					return m, runEnqueueWave(m.projectRoot, level)
				}
			}

			return m, nil

		case "A":
			// In tasks or waves: run task_analysis and show result in dedicated view
			if m.mode == "tasks" || m.mode == "waves" {
				m.taskAnalysisReturnMode = m.mode
				m.mode = "taskAnalysis"
				m.taskAnalysisLoading = true
				m.taskAnalysisErr = nil
				m.taskAnalysisText = ""
				m.taskAnalysisAction = "parallelization"
				m.taskAnalysisApproveMsg = ""
				m.taskAnalysisApproveLoading = false

				return m, runTaskAnalysis(m.server, "parallelization")
			}

			if m.mode == "taskAnalysis" && !m.taskAnalysisLoading {
				// Rerun same action
				m.taskAnalysisLoading = true
				m.taskAnalysisApproveMsg = ""

				return m, runTaskAnalysis(m.server, m.taskAnalysisAction)
			}

			return m, nil

		case "y":
			// In task analysis view: approve = write waves plan to .cursor/plans/parallel-execution-subagents.plan.md
			if m.mode == "taskAnalysis" && !m.taskAnalysisLoading && !m.taskAnalysisApproveLoading {
				m.taskAnalysisApproveLoading = true
				m.taskAnalysisApproveMsg = ""

				return m, runReportParallelExecutionPlan(m.server, m.projectRoot)
			}

			return m, nil

		case "m":
			// In waves expanded view: start "move task to wave" (then press 0-9 to pick target wave)
			if m.mode == "waves" && m.waveDetailLevel >= 0 && m.waveMoveTaskID == "" {
				ids := m.waves[m.waveDetailLevel]
				if len(ids) > 0 && m.waveTaskCursor < len(ids) {
					m.waveMoveTaskID = ids[m.waveTaskCursor]
					m.waveMoveMsg = ""
				}
			}

			return m, nil

		case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
			if m.mode == "waves" && m.waveMoveTaskID != "" {
				targetLevel := int(msg.String()[0] - '0')
				levels := sortedWaveLevels(m.waves)

				if targetLevel < 0 || targetLevel >= len(levels) {
					m.waveMoveMsg = "Invalid wave number"
					return m, nil
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
						return m, nil
					}

					newDeps = []string{prevIDs[0]}
				}

				taskByID := make(map[string]*database.Todo2Task)

				for _, t := range m.tasks {
					if t != nil {
						taskByID[t.ID] = t
					}
				}

				task := taskByID[m.waveMoveTaskID]
				if task == nil {
					m.waveMoveMsg = "Task not found"
					return m, nil
				}

				return m, moveTaskToWaveCmd(task, newDeps)
			}

			return m, nil

		case "E":
			// Execute current context (task, handoff, wave) in child agent
			m.childAgentMsg = ""
			if m.mode == "tasks" {
				vis := m.visibleIndices()
				if len(vis) > 0 && m.cursor < len(vis) {
					task := m.tasks[m.realIndexAt(m.cursor)]
					if task != nil {
						prompt := PromptForTask(task.ID, task.Content)
						return m, runChildAgentCmd(m.projectRoot, prompt, ChildAgentTask)
					}
				}
			} else if m.mode == "taskDetail" && m.taskDetailTask != nil {
				prompt := PromptForTask(m.taskDetailTask.ID, m.taskDetailTask.Content)
				return m, runChildAgentCmd(m.projectRoot, prompt, ChildAgentTask)
			} else if m.mode == "handoffs" {
				if m.handoffDetailIndex >= 0 && m.handoffDetailIndex < len(m.handoffEntries) {
					h := m.handoffEntries[m.handoffDetailIndex]
					sum, _ := h["summary"].(string)

					var steps []interface{}
					if s, ok := h["next_steps"].([]interface{}); ok {
						steps = s
					}

					prompt := PromptForHandoff(sum, steps)

					return m, runChildAgentCmd(m.projectRoot, prompt, ChildAgentHandoff)
				} else if len(m.handoffEntries) > 0 && m.handoffCursor < len(m.handoffEntries) {
					h := m.handoffEntries[m.handoffCursor]
					sum, _ := h["summary"].(string)

					var steps []interface{}
					if s, ok := h["next_steps"].([]interface{}); ok {
						steps = s
					}

					prompt := PromptForHandoff(sum, steps)

					return m, runChildAgentCmd(m.projectRoot, prompt, ChildAgentHandoff)
				}
			} else if m.mode == "waves" && len(m.waves) > 0 {
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

					return m, runChildAgentCmd(m.projectRoot, prompt, ChildAgentWave)
				}
			}

			return m, nil

		case "L":
			// Launch plan in child agent (tasks or taskDetail)
			m.childAgentMsg = ""
			if m.mode == "tasks" || m.mode == "taskDetail" {
				prompt := PromptForPlan(m.projectRoot)
				return m, runChildAgentCmd(m.projectRoot, prompt, ChildAgentPlan)
			}

			return m, nil

		default:
			// Clear child-agent status on any other key
			if m.childAgentMsg != "" {
				m.childAgentMsg = ""
			}
		}
	}

	return m, nil
}

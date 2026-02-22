// tui3270_helpers.go — Command handling, help screen, and utility transactions for the 3270 TUI.
// Extracted from tui3270.go. Handles ISPF-style command parsing, help overlay, child agent results.
package cli

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/davidl71/exarp-go/internal/database"
	"github.com/racingmars/go3270"
)

// loadTasksForStatus loads tasks by status via MCP adapter.
// When status is empty, returns open tasks only (Todo + In Progress).
func (state *tui3270State) loadTasksForStatus(ctx context.Context, status string) ([]*database.Todo2Task, error) {
	return listTasksViaMCP(ctx, state.server, status)
}

// showChildAgentResultTransaction shows a one-screen result then returns to nextTx.
func (state *tui3270State) showChildAgentResultTransaction(message string, nextTx go3270.Tx) go3270.Tx {
	return func(conn net.Conn, devInfo go3270.DevInfo, data any) (go3270.Tx, any, error) {
		screen := go3270.Screen{
			{Row: 2, Col: 2, Content: "CHILD AGENT", Intense: true, Color: go3270.Blue},
			{Row: 4, Col: 2, Content: message, Color: go3270.Green},
			{Row: 22, Col: 2, Content: "PF3=Back to menu", Color: go3270.Turquoise},
		}
		if len(message) > 76 {
			// Wrap
			screen = append(screen, go3270.Field{Row: 5, Col: 2, Content: message[76:]})
		}

		opts := go3270.ScreenOpts{Codepage: devInfo.Codepage()}

		response, err := go3270.ShowScreenOpts(screen, nil, conn, opts)
		if err != nil {
			return nil, nil, err
		}

		if response.AID == go3270.AIDPF1 {
			return state.helpTransaction, state, nil
		}

		return nextTx, state, nil
	}
}

// helpTransaction shows the help screen (PF1).
func (state *tui3270State) helpTransaction(conn net.Conn, devInfo go3270.DevInfo, data any) (go3270.Tx, any, error) {
	lines := []string{
		"EXARP-GO 3270 - HELP",
		"",
		"Main menu: 1=Tasks 2=Config 3=Scorecard 4=Handoffs 5=Exit 6=Agent 7=Health",
		"",
		"Commands (type in COMMAND ===> field):",
		"  TASKS/T  CONFIG  SC  HANDOFFS/HO  MENU/M  HELP/H",
		"  HEALTH/SDSF  GIT/GITLOG  SPRINT/BOARD",
		"  FIND <text>  RESET  VIEW [id]  EDIT [id]  TOP  BOTTOM",
		"  RUN TASK|PLAN|WAVE|HANDOFF",
		"",
		"Line commands (type in S column next to task row):",
		"  S=Select(view)  E=Edit  D=Mark Done  I=Mark In Progress",
		"",
		"PF keys (all screens):",
		"  PF1=Help  PF3=Back/Exit",
		"",
		"PF keys (task list):",
		"  PF7/8=Scroll  PF9=Cycle status filter  PF2=Edit",
		"  PF4=Mark Done  PF5=Mark In Progress  PF6=Mark Todo",
		"  PF10=Mark Review  Enter=Select (click row)",
		"",
		"PF keys (task detail):",
		"  PF2=Edit  PF4=Done  PF5=WIP  PF6=Todo  PF10=Review",
		"",
		"Press PF3 to return.",
	}

	helpPFRow := t3270PFRow(devInfo)
	helpContentMax := t3270ContentMaxRow(devInfo)

	screen := go3270.Screen{
		{Row: 1, Col: 2, Content: "HELP", Intense: true, Color: go3270.Blue},
		{Row: helpPFRow, Col: 2, Content: "PF3=Back to previous screen", Color: go3270.Turquoise},
	}

	maxLines := helpContentMax - 2
	for i, line := range lines {
		if i >= maxLines {
			break
		}

		if len(line) > 78 {
			line = line[:75] + "..."
		}

		screen = append(screen, go3270.Field{Row: 2 + i, Col: 2, Content: line})
	}

	screenOpts := go3270.ScreenOpts{Codepage: devInfo.Codepage()}

	response, err := go3270.ShowScreenOpts(screen, nil, conn, screenOpts)
	if err != nil {
		return nil, nil, err
	}

	if response.AID == go3270.AIDPF3 {
		return state.mainMenuTransaction, state, nil
	}

	return state.helpTransaction, state, nil
}

// handleCommand processes command line input (ISPF-style).
func (state *tui3270State) handleCommand(cmd string, currentTx go3270.Tx) (go3270.Tx, any, error) {
	cmd = strings.TrimSpace(cmd)
	cmdUpper := strings.ToUpper(cmd)

	// Parse command
	parts := strings.Fields(cmdUpper)
	if len(parts) == 0 {
		return currentTx, state, nil
	}

	command := parts[0]
	args := parts[1:]

	switch command {
	case "1":
		state.command = ""
		return state.taskListTransaction, state, nil
	case "2":
		state.command = ""
		return state.configTransaction, state, nil
	case "3":
		state.command = ""
		return state.scorecardTransaction, state, nil
	case "4":
		state.command = ""
		return state.handoffTransaction, state, nil
	case "5":
		state.command = ""
		return nil, nil, nil // Exit
	case "7":
		state.command = ""
		return state.healthTransaction, state, nil
	case "SC", "SCORECARD":
		state.command = ""
		return state.scorecardTransaction, state, nil
	case "HANDOFFS", "HO":
		state.command = ""
		return state.handoffTransaction, state, nil
	case "MENU", "M", "MAIN":
		state.command = ""
		return state.mainMenuTransaction, state, nil
	case "TASKS", "T":
		state.command = ""
		return state.taskListTransaction, state, nil
	case "CONFIG":
		state.command = ""
		return state.configTransaction, state, nil
	case "HELP", "H":
		state.command = ""
		return state.helpTransaction, state, nil
	case "HEALTH", "SDSF":
		state.pushSession("Tasks", fallback)
		state.command = ""
		return state.healthTransaction, state, nil
	case "GIT", "GITLOG":
		state.pushSession("Tasks", fallback)
		state.command = ""
		return state.gitDashboardTransaction, state, nil
	case "SPRINT", "BOARD":
		state.pushSession("Tasks", fallback)
		state.command = ""
		return state.sprintBoardTransaction, state, nil
	case "SWAP":
		state.command = ""
		s := state.popSession()
		if s != nil {
			return s.tx, state, nil
		}
		return state.mainMenuTransaction, state, nil
	case "FIND", "F":
		// Filter/search tasks
		if len(args) > 0 {
			state.filter = strings.Join(args, " ")
			state.cursor = 0
			state.listOffset = 0
			ctx := context.Background()

			var err error

			state.tasks, err = state.loadTasksForStatus(ctx, state.status)
			if err == nil {
				// Simple text search filter
				filtered := []*database.Todo2Task{}
				searchTerm := strings.ToLower(state.filter)

				for _, task := range state.tasks {
					content := strings.ToLower(task.Content + " " + task.LongDescription)
					if strings.Contains(content, searchTerm) {
						filtered = append(filtered, task)
					}
				}

				state.tasks = filtered
			}
		} else {
			state.filter = ""
			ctx := context.Background()

			var err error

			state.tasks, err = state.loadTasksForStatus(ctx, state.status)
			if err != nil {
				logError(context.Background(), "Error reloading tasks", "error", err, "operation", "reloadTasks")
			}
		}

		state.command = ""

		return state.taskListTransaction, state, nil

	case "RESET", "RES":
		// Reset filter
		state.filter = ""
		state.command = ""
		ctx := context.Background()

		var err error

		state.tasks, err = state.loadTasksForStatus(ctx, state.status)
		if err != nil {
			logError(context.Background(), "Error reloading tasks", "error", err, "operation", "reloadTasks")
		}

		return state.taskListTransaction, state, nil

	case "EDIT", "E":
		// Edit task by ID or line number
		if len(args) > 0 {
			taskID := args[0]
			// Check if it's a line number
			if strings.HasPrefix(taskID, "T-") {
				// Find task by ID
				ctx := context.Background()

				task, err := getTaskViaMCP(ctx, state.server, taskID)
				if err == nil {
					state.selectedTask = task
					state.command = ""

					return state.taskEditorTransaction, state, nil
				}
			} else {
				// Try as line number
				var lineNum int
				if _, err := fmt.Sscanf(taskID, "%d", &lineNum); err == nil {
					if lineNum > 0 && lineNum <= len(state.tasks) {
						state.selectedTask = state.tasks[lineNum-1]
						state.cursor = lineNum - 1
						state.command = ""

						return state.taskEditorTransaction, state, nil
					}
				}
			}
		} else if state.cursor < len(state.tasks) {
			// Edit current task
			state.selectedTask = state.tasks[state.cursor]
			state.command = ""

			return state.taskEditorTransaction, state, nil
		}

		state.command = ""

		return currentTx, state, nil

	case "VIEW", "V":
		// View task details
		if len(args) > 0 {
			taskID := args[0]
			if strings.HasPrefix(taskID, "T-") {
				ctx := context.Background()

				task, err := getTaskViaMCP(ctx, state.server, taskID)
				if err == nil {
					state.selectedTask = task
					state.command = ""

					return state.taskDetailTransaction, state, nil
				}
			}
		} else if state.cursor < len(state.tasks) {
			state.selectedTask = state.tasks[state.cursor]
			state.command = ""

			return state.taskDetailTransaction, state, nil
		}

		state.command = ""

		return currentTx, state, nil

	case "TOP":
		// Go to top of list
		state.cursor = 0
		state.listOffset = 0
		state.command = ""

		return state.taskListTransaction, state, nil

	case "BOTTOM", "BOT":
		// Go to bottom of list; use default maxVisible since devInfo is not in scope
		if len(state.tasks) > 0 {
			state.cursor = len(state.tasks) - 1
			mv := 18
			if state.devInfo != nil {
				mv = t3270MaxVisible(state.devInfo)
			}
			if state.cursor >= mv {
				state.listOffset = state.cursor - mv + 1
			}
		}

		state.command = ""

		return state.taskListTransaction, state, nil

	case "RUN":
		// RUN TASK | PLAN | WAVE | HANDOFF - execute in child agent
		state.command = ""

		sub := ""
		if len(args) > 0 {
			sub = args[0]
		}

		switch strings.ToUpper(sub) {
		case "TASK":
			if state.cursor < len(state.tasks) {
				task := state.tasks[state.cursor]
				prompt := PromptForTask(task.ID, task.Content)
				r := RunChildAgent(state.projectRoot, prompt)

				return state.showChildAgentResultTransaction(r.Message, state.taskListTransaction), state, nil
			}

			return state.showChildAgentResultTransaction("No task selected", state.taskListTransaction), state, nil
		case "PLAN":
			prompt := PromptForPlan(state.projectRoot)
			r := RunChildAgent(state.projectRoot, prompt)

			return state.showChildAgentResultTransaction(r.Message, state.taskListTransaction), state, nil
		case "WAVE":
			if len(state.tasks) == 0 {
				return state.showChildAgentResultTransaction("No tasks", state.taskListTransaction), state, nil
			}

			level, ids, err := firstWaveTaskIDs(state.projectRoot, state.tasks)
			if err != nil {
				return state.showChildAgentResultTransaction("No waves", state.taskListTransaction), state, nil
			}

			prompt := PromptForWave(level, ids)
			r := RunChildAgent(state.projectRoot, prompt)

			return state.showChildAgentResultTransaction(r.Message, state.taskListTransaction), state, nil
		case "HANDOFF":
			ctx := context.Background()

			entries, err := fetchHandoffs(ctx, state.server, 5)
			if err != nil || len(entries) == 0 {
				return state.showChildAgentResultTransaction("No handoffs", state.taskListTransaction), state, nil
			}

			h := entries[0]
			steps := make([]interface{}, len(h.NextSteps))
			for i, s := range h.NextSteps {
				steps[i] = s
			}

			prompt := PromptForHandoff(h.Summary, steps)
			r := RunChildAgent(state.projectRoot, prompt)

			return state.showChildAgentResultTransaction(r.Message, state.taskListTransaction), state, nil
		default:
			return state.showChildAgentResultTransaction("RUN TASK|PLAN|WAVE|HANDOFF", state.taskListTransaction), state, nil
		}

	case "AGENT":
		// Alias for RUN (same args)
		state.command = ""

		parts := append([]string{"RUN"}, args...)

		return state.handleCommand(strings.Join(parts, " "), currentTx)

	default:
		// Unknown command - stay on current screen
		state.command = ""
		return currentTx, state, nil
	}
}

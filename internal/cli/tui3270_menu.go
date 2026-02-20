// tui3270_menu.go — Main menu and child agent menu transactions for the 3270 TUI.
// Extracted from tui3270.go. Handles menu option selection and child agent launches.
package cli

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/racingmars/go3270"
)

// mainMenuTransaction shows the main menu.
func (state *tui3270State) mainMenuTransaction(conn net.Conn, devInfo go3270.DevInfo, data any) (go3270.Tx, any, error) {
	title := "EXARP-GO TASK MANAGEMENT"
	if state.projectName != "" {
		title = fmt.Sprintf("%s - %s", state.projectName, title)
	}

	screen := go3270.Screen{
		{Row: 2, Col: 2, Content: title, Intense: true, Color: go3270.Blue},
		{Row: 4, Col: 2, Content: "Select an option (1-6) or type a command below:", Color: go3270.Green},
		{Row: 6, Col: 4, Content: "1. Task List", Color: go3270.Turquoise},
		{Row: 7, Col: 4, Content: "2. Configuration", Color: go3270.Turquoise},
		{Row: 8, Col: 4, Content: "3. Scorecard", Color: go3270.Turquoise},
		{Row: 9, Col: 4, Content: "4. Session handoffs", Color: go3270.Turquoise},
		{Row: 10, Col: 4, Content: "5. Exit", Color: go3270.Turquoise},
		{Row: 11, Col: 4, Content: "6. Run in child agent (task/plan/wave/handoff)", Color: go3270.Turquoise},
		{Row: 12, Col: 2, Content: "Option: ", Intense: true},
		{Row: 12, Col: 10, Content: "", Write: true, Name: "option", Color: go3270.Green},
		{Row: 22, Col: 2, Content: "Command ===>", Intense: true, Color: go3270.Green},
		{Row: 22, Col: 15, Write: true, Name: "command", Content: "", Color: go3270.Turquoise},
		{Row: 23, Col: 2, Content: "PF1=Help  PF3=Exit  Enter=Select option or run command", Color: go3270.Turquoise},
	}

	opts := go3270.ScreenOpts{
		Codepage: devInfo.Codepage(),
	}

	response, err := go3270.ShowScreenOpts(screen, nil, conn, opts)
	if err != nil {
		return nil, nil, err
	}

	// Check AID (Action ID) for PF keys
	if response.AID == go3270.AIDPF3 {
		return nil, nil, nil // Exit
	}

	if response.AID == go3270.AIDPF1 {
		return state.helpTransaction, state, nil
	}

	// Check command line first
	cmd := strings.TrimSpace(response.Values["command"])
	if cmd != "" {
		return state.handleCommand(cmd, state.mainMenuTransaction)
	}

	// Check field value for option (allow "1"-"6", or value after "Option:")
	optionRaw := strings.TrimSpace(response.Values["option"])

	option := extractMenuOption(optionRaw)
	switch option {
	case "1":
		return state.taskListTransaction, state, nil
	case "2":
		return state.configTransaction, state, nil
	case "3":
		return state.scorecardTransaction, state, nil
	case "4":
		return state.handoffTransaction, state, nil
	case "5", "":
		if response.AID == go3270.AIDEnter && option == "" {
			return nil, nil, nil // Exit
		}

		return state.mainMenuTransaction, state, nil
	case "6":
		return state.childAgentMenuTransaction, state, nil
	default:
		// Invalid option, stay on main menu
		return state.mainMenuTransaction, state, nil
	}
}

// extractMenuOption returns "1".."6" from user input, or "" if none.
func extractMenuOption(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	// Single digit
	if len(s) == 1 && s >= "1" && s <= "6" {
		return s
	}
	// Last character (e.g. "Option: 1" -> "1")
	if len(s) > 0 {
		c := s[len(s)-1:]
		if c >= "1" && c <= "6" {
			return c
		}
	}
	// First digit in string
	for _, r := range s {
		if r >= '1' && r <= '6' {
			return string(r)
		}
	}

	return ""
}

// childAgentMenuTransaction shows Run in child agent submenu (option 6).
func (state *tui3270State) childAgentMenuTransaction(conn net.Conn, devInfo go3270.DevInfo, data any) (go3270.Tx, any, error) {
	screen := go3270.Screen{
		{Row: 2, Col: 2, Content: "RUN IN CHILD AGENT", Intense: true, Color: go3270.Blue},
		{Row: 4, Col: 2, Content: "Launches Cursor CLI 'agent' in project root with a prompt.", Color: go3270.Green},
		{Row: 5, Col: 2, Content: "Select (1-5):"},
		{Row: 7, Col: 4, Content: "1. Task (current or first)", Color: go3270.Turquoise},
		{Row: 8, Col: 4, Content: "2. Plan", Color: go3270.Turquoise},
		{Row: 9, Col: 4, Content: "3. Wave (first wave)", Color: go3270.Turquoise},
		{Row: 10, Col: 4, Content: "4. Handoff (first handoff)", Color: go3270.Turquoise},
		{Row: 11, Col: 4, Content: "5. Back to main menu", Color: go3270.Turquoise},
		{Row: 13, Col: 2, Content: "Option: ", Intense: true},
		{Row: 13, Col: 10, Content: "", Write: true, Name: "option_val", Color: go3270.Green},
		{Row: 22, Col: 2, Content: "PF3=Back", Color: go3270.Turquoise},
	}

	opts := go3270.ScreenOpts{Codepage: devInfo.Codepage()}

	response, err := go3270.ShowScreenOpts(screen, nil, conn, opts)
	if err != nil {
		return nil, nil, err
	}

	if response.AID == go3270.AIDPF3 {
		return state.mainMenuTransaction, state, nil
	}

	if response.AID == go3270.AIDPF1 {
		return state.helpTransaction, state, nil
	}

	opt := strings.TrimSpace(response.Values["option_val"])
	if opt == "" {
		return state.childAgentMenuTransaction, state, nil
	}

	var prompt string

	var kind ChildAgentKind

	switch opt {
	case "5":
		return state.mainMenuTransaction, state, nil
	case "1":
		// Task: use current cursor task or first task
		ctx := context.Background()

		tasks, err := state.loadTasksForStatus(ctx, state.status)
		if err != nil || len(tasks) == 0 {
			msg := "No tasks"
			if err != nil {
				msg = err.Error()
			}

			return state.showChildAgentResultTransaction(msg, state.mainMenuTransaction), state, nil
		}

		idx := state.cursor
		if idx >= len(tasks) {
			idx = 0
		}

		task := tasks[idx]
		prompt = PromptForTask(task.ID, task.Content)
		kind = ChildAgentTask
	case "2":
		prompt = PromptForPlan(state.projectRoot)
		kind = ChildAgentPlan
	case "3":
		ctx := context.Background()

		tasks, err := state.loadTasksForStatus(ctx, state.status)
		if err != nil || len(tasks) == 0 {
			msg := "No tasks for wave"
			if err != nil {
				msg = err.Error()
			}

			return state.showChildAgentResultTransaction(msg, state.mainMenuTransaction), state, nil
		}

		level, ids, err := firstWaveTaskIDs(state.projectRoot, tasks)
		if err != nil {
			return state.showChildAgentResultTransaction("No waves", state.mainMenuTransaction), state, nil
		}

		prompt = PromptForWave(level, ids)
		kind = ChildAgentWave
	case "4":
		ctx := context.Background()

		entries, err := fetchHandoffs(ctx, state.server, 5)
		if err != nil || len(entries) == 0 {
			return state.showChildAgentResultTransaction("No handoffs", state.mainMenuTransaction), state, nil
		}

		h := entries[0]
		steps := make([]interface{}, len(h.NextSteps))
		for i, s := range h.NextSteps {
			steps[i] = s
		}

		prompt = PromptForHandoff(h.Summary, steps)
		kind = ChildAgentHandoff
	default:
		return state.childAgentMenuTransaction, state, nil
	}

	r := RunChildAgent(state.projectRoot, prompt)
	r.Kind = kind
	msg := r.Message

	if !r.Launched {
		msg = "Error: " + msg
	}

	return state.showChildAgentResultTransaction(msg, state.mainMenuTransaction), state, nil
}

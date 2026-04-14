// tui3270_menu.go — Main menu and child agent menu transactions for the 3270 TUI.
// Extracted from tui3270.go. Handles menu option selection and child agent launches.
// Layout follows ISPF-style primary option panels (panel banner, option list, Selection/Command lines).
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
	product := "EXARP-GO TASK MANAGEMENT"
	if state.projectName != "" {
		product = fmt.Sprintf("%s - TASK CONSOLE", state.projectName)
	}

	cols := t3270ScreenCols(devInfo)
	rows, _ := devInfo.AltDimensions()
	if rows < 24 {
		rows = 24
	}
	seprow := rows - 4
	pf1 := rows - 2
	pf2 := rows - 1

	rule0 := t3270ISPFRuleLine(cols, " PRIMARY OPTION MENU ")
	topLine, botLine, rPipe, _ := t3270MenuBoxTopBottom(devInfo)
	innerOpt := rPipe - 4
	if innerOpt < 20 {
		innerOpt = 20
	}
	gc := t3270PanelRuleColor()
	banner := t3270ISPFPanelBannerLine(cols, product, "ZEXRPMNU")

	hIntro := "Select an option from the list below.  Enter a number on the Selection line,"
	hIntro2 := "or a verb command on the Command line, then press Enter to process."
	if len(hIntro) > cols {
		hIntro = hIntro[:cols-1] + "~"
	}
	if len(hIntro2) > cols {
		hIntro2 = hIntro2[:cols-1] + "~"
	}

	screen := go3270.Screen{
		{Row: 0, Col: 0, Content: rule0, Color: gc},
		{Row: 1, Col: 0, Content: banner, Color: t3270ISPFTitleColor(), Intense: true},
		{Row: 2, Col: 2, Content: hIntro, Color: go3270.Green},
		{Row: 3, Col: 2, Content: hIntro2, Color: go3270.Green},
		{Row: 4, Col: 2, Content: topLine, Color: gc},
	}
	opts := []struct {
		n   int
		txt string
	}{
		{1, "Task list, line commands, and search"},
		{2, "Configuration and environment"},
		{3, "Project scorecard and checks"},
		{4, "Session handoffs and notes"},
		{5, "Exit Exarp-go / end session"},
		{6, "Run Cursor child agent (task, plan, wave, handoff)"},
		{7, "System health and activity (SDSF-style)"},
	}
	for i, o := range opts {
		r := 5 + i
		line := t3270ISPFOptionLine(o.n, o.txt, innerOpt)
		screen = append(screen,
			go3270.Field{Row: r, Col: 2, Content: "|", Color: gc},
			go3270.Field{Row: r, Col: 4, Content: line, Color: go3270.Turquoise},
			go3270.Field{Row: r, Col: rPipe, Content: "|", Color: gc},
		)
	}
	boxBotRow := 5 + len(opts)
	selRow := boxBotRow + 1
	cmdRow := boxBotRow + 2
	screen = append(screen,
		go3270.Field{Row: boxBotRow, Col: 2, Content: botLine, Color: gc},
		go3270.Field{Row: selRow, Col: 2, Content: "Selection ===>", Intense: true, Color: go3270.Green},
		go3270.Field{Row: selRow, Col: 18, Content: "", Write: true, Name: "option", Color: go3270.Green},
		go3270.Field{Row: cmdRow, Col: 2, Content: "Command ===>", Intense: true, Color: go3270.Green},
		go3270.Field{Row: cmdRow, Col: 18, Write: true, Name: "command", Content: "", Color: go3270.Turquoise},
		go3270.Field{Row: seprow, Col: 0, Content: t3270ISPFRuleLine(cols, ""), Color: gc},
		go3270.Field{Row: pf1, Col: 2, Content: t3270Pad("Enter=Process  PF01=Help  PF03=End  PF12=Cancel", cols-4), Color: go3270.Turquoise},
		go3270.Field{Row: pf2, Col: 2, Content: t3270Pad("Use the Selection line for 1-7; Command line for TASKS, CONFIG, MENU, HELP, ...", cols-4), Color: go3270.Turquoise},
	)

	opts3270 := go3270.ScreenOpts{
		Codepage: devInfo.Codepage(),
	}

	response, err := go3270.ShowScreenOpts(screen, nil, conn, opts3270)
	if err != nil {
		return nil, nil, err
	}

	if response.AID == go3270.AIDPF3 || response.AID == go3270.AIDPF12 {
		return nil, nil, nil // Exit
	}

	if response.AID == go3270.AIDPF1 {
		return state.helpTransaction, state, nil
	}

	cmd := strings.TrimSpace(response.Values["command"])
	if cmd != "" {
		return state.handleCommand(cmd, state.mainMenuTransaction)
	}

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
	case "5":
		return nil, nil, nil // End session (ISPF End)
	case "":
		return state.mainMenuTransaction, state, nil
	case "6":
		return state.childAgentMenuTransaction, state, nil
	case "7":
		return state.healthTransaction, state, nil
	default:
		return state.mainMenuTransaction, state, nil
	}
}

// extractMenuOption returns "1".."7" from user input, or "" if none.
func extractMenuOption(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if len(s) == 1 && s >= "1" && s <= "7" {
		return s
	}
	if len(s) > 0 {
		c := s[len(s)-1:]
		if c >= "1" && c <= "7" {
			return c
		}
	}
	for _, r := range s {
		if r >= '1' && r <= '7' {
			return string(r)
		}
	}
	return ""
}

// childAgentMenuTransaction shows Run in child agent submenu (option 6).
func (state *tui3270State) childAgentMenuTransaction(conn net.Conn, devInfo go3270.DevInfo, data any) (go3270.Tx, any, error) {
	cols := t3270ScreenCols(devInfo)
	rows, _ := devInfo.AltDimensions()
	if rows < 24 {
		rows = 24
	}
	seprow := rows - 4
	pf1 := rows - 2
	pf2 := rows - 1

	rule0 := t3270ISPFRuleLine(cols, " CHILD AGENT LAUNCH ")
	topLine, botLine, rPipe, _ := t3270MenuBoxTopBottom(devInfo)
	innerOpt := rPipe - 4
	if innerOpt < 20 {
		innerOpt = 20
	}
	gc := t3270PanelRuleColor()
	banner := t3270ISPFPanelBannerLine(cols, "Cursor agent launcher in project root", "ZEXRAGNT")

	screen := go3270.Screen{
		{Row: 0, Col: 0, Content: rule0, Color: gc},
		{Row: 1, Col: 0, Content: banner, Color: t3270ISPFTitleColor(), Intense: true},
		{Row: 2, Col: 2, Content: "Choose how to build the agent prompt.  Selection is required.", Color: go3270.Green},
		{Row: 3, Col: 2, Content: "Option 5 returns to the primary menu without running an agent.", Color: go3270.Green},
		{Row: 4, Col: 2, Content: topLine, Color: gc},
	}
	optLab := []struct {
		n   int
		txt string
	}{
		{1, "Task (cursor task or first in list)"},
		{2, "Plan (workspace planning prompt)"},
		{3, "Wave (first parallel-wave tasks)"},
		{4, "Handoff (most recent handoff summary)"},
		{5, "Cancel / return to primary menu"},
	}
	for i, o := range optLab {
		r := 5 + i
		line := t3270ISPFOptionLine(o.n, o.txt, innerOpt)
		screen = append(screen,
			go3270.Field{Row: r, Col: 2, Content: "|", Color: gc},
			go3270.Field{Row: r, Col: 4, Content: line, Color: go3270.Turquoise},
			go3270.Field{Row: r, Col: rPipe, Content: "|", Color: gc},
		)
	}
	br := 5 + len(optLab)
	selRow := br + 1
	screen = append(screen,
		go3270.Field{Row: br, Col: 2, Content: botLine, Color: gc},
		go3270.Field{Row: selRow, Col: 2, Content: "Selection ===>", Intense: true, Color: go3270.Green},
		go3270.Field{Row: selRow, Col: 18, Content: "", Write: true, Name: "option_val", Color: go3270.Green},
		go3270.Field{Row: seprow, Col: 0, Content: t3270ISPFRuleLine(cols, ""), Color: gc},
		go3270.Field{Row: pf1, Col: 2, Content: t3270Pad("Enter=Run selection  PF01=Help  PF03=Primary menu  PF12=Primary menu", cols-4), Color: go3270.Turquoise},
		go3270.Field{Row: pf2, Col: 2, Content: t3270Pad("Invalid option redisplays this panel.", cols-4), Color: go3270.Turquoise},
	)

	opts3270 := go3270.ScreenOpts{Codepage: devInfo.Codepage()}

	response, err := go3270.ShowScreenOpts(screen, nil, conn, opts3270)
	if err != nil {
		return nil, nil, err
	}

	if response.AID == go3270.AIDPF3 || response.AID == go3270.AIDPF12 {
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

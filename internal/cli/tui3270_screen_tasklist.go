// tui3270_screen_tasklist.go — Task list screen transaction for the 3270 TUI.
package cli

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/racingmars/go3270"
)

// taskListTransaction shows the task list in 3270-style tabular layout (aligned columns, header row, COMMAND/SCROLL line).
func (state *tui3270State) taskListTransaction(conn net.Conn, devInfo go3270.DevInfo, data any) (go3270.Tx, any, error) {
	ctx := context.Background()

	showLoadingOverlay(conn, devInfo, "Loading tasks...")

	var err error

	state.tasks, err = state.loadTasksForStatus(ctx, state.status)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	title := "TASK LIST *"
	if state.status != "" {
		title = fmt.Sprintf("TASK LIST (%s) *", state.status)
	}

	if state.filter != "" {
		title = fmt.Sprintf("TASK LIST [%s] *", state.filter)
	}

	maxVisible := t3270MaxVisible(devInfo)

	totalLines := len(state.tasks)
	if totalLines == 0 {
		totalLines = 1
	}

	currentLine := state.cursor + 1
	if currentLine < 1 {
		currentLine = 1
	}

	if currentLine > totalLines {
		currentLine = totalLines
	}

	scrollIndicator := "CS"
	if state.listOffset > 0 {
		scrollIndicator = "MORE"
	} else if state.listOffset+maxVisible >= len(state.tasks) && len(state.tasks) > maxVisible {
		scrollIndicator = "BOTTOM"
	} else if len(state.tasks) <= maxVisible {
		scrollIndicator = "CS"
	} else {
		scrollIndicator = "TOP"
	}

	// Row 1: title left, LINE x OF y, COMMAND ===> with input, SCROLL ===> xx (mainframe-style)
	screen := go3270.Screen{
		{Row: 1, Col: 2, Content: title, Intense: true, Color: go3270.Blue},
		{Row: 1, Col: 45, Content: fmt.Sprintf("LINE %d OF %d", currentLine, totalLines), Intense: true, Color: go3270.Blue},
		{Row: 1, Col: 55, Content: "COMMAND ===>", Intense: true, Color: go3270.Green},
		{Row: 1, Col: 68, Write: true, Name: "command", Content: state.command, Color: go3270.Turquoise},
		{Row: 1, Col: 72, Content: "SCROLL ===> " + scrollIndicator, Intense: true, Color: go3270.Blue},
	}

	// Row 2: blank
	screen = append(screen, go3270.Field{Row: 2, Col: 2, Content: ""})

	// Row 3: column headers (yellow-style, intense white)
	screen = append(screen, go3270.Field{Row: 3, Col: t3270ColS, Content: t3270Pad("S", t3270WidS), Intense: true, Color: go3270.White})
	screen = append(screen, go3270.Field{Row: 3, Col: t3270ColID, Content: t3270Pad("ID", t3270WidID), Intense: true, Color: go3270.White})
	screen = append(screen, go3270.Field{Row: 3, Col: t3270ColStatus, Content: t3270Pad("STATUS", t3270WidStatus), Intense: true, Color: go3270.White})
	screen = append(screen, go3270.Field{Row: 3, Col: t3270ColPriority, Content: t3270Pad("PRIORITY", t3270WidPriority), Intense: true, Color: go3270.White})
	screen = append(screen, go3270.Field{Row: 3, Col: t3270ColContent, Content: t3270Pad("CONTENT", t3270WidContent), Intense: true, Color: go3270.White})

	startIdx := state.listOffset

	endIdx := startIdx + maxVisible
	if endIdx > len(state.tasks) {
		endIdx = len(state.tasks)
	}

	if endIdx < startIdx {
		endIdx = startIdx
	}

	if len(state.tasks) == 0 {
		msg := "No tasks found."
		if state.filter != "" {
			msg = fmt.Sprintf("No tasks match filter: %s", state.filter)
		} else if state.status != "" {
			msg = fmt.Sprintf("No tasks with status: %s", state.status)
		}
		screen = append(screen, go3270.Field{Row: t3270HeaderRow, Col: t3270ColID, Content: msg, Color: go3270.Yellow})
	}

	for row, i := t3270HeaderRow, startIdx; i < endIdx; row, i = row+1, i+1 {
		task := state.tasks[i]
		isCursor := i == state.cursor

		cursorHL := go3270.DefaultHighlight
		if isCursor {
			cursorHL = go3270.ReverseVideo
		}

		// Writable line-command field per row (S=Select, E=Edit, D=Done, I=In Progress)
		sFieldName := fmt.Sprintf("s_%d", i-startIdx)
		screen = append(screen, go3270.Field{Row: row, Col: t3270ColS, Write: true, Name: sFieldName, Content: "", Highlighting: cursorHL, Color: go3270.Green})
		screen = append(screen, go3270.Field{Row: row, Col: t3270ColID, Content: t3270Pad(task.ID, t3270WidID), Highlighting: cursorHL, Color: go3270.Turquoise})

		st := task.Status
		if st == "" {
			st = "-"
		}

		screen = append(screen, go3270.Field{Row: row, Col: t3270ColStatus, Content: t3270Pad(st, t3270WidStatus), Highlighting: cursorHL, Color: statusColor(task.Status)})

		pri := strings.ToUpper(task.Priority)
		if pri == "" {
			pri = "-"
		}

		screen = append(screen, go3270.Field{Row: row, Col: t3270ColPriority, Content: t3270Pad(pri, t3270WidPriority), Highlighting: cursorHL, Color: priorityColor(task.Priority)})

		content := task.Content
		if content == "" {
			content = task.LongDescription
		}

		if content == "" {
			content = "(no description)"
		}

		screen = append(screen, go3270.Field{Row: row, Col: t3270ColContent, Content: t3270Pad(content, t3270WidContent), Highlighting: cursorHL})
	}

	statusRow := t3270StatusRow(devInfo)
	pfRow := t3270PFRow(devInfo)

	statusLine := fmt.Sprintf("Tasks: %d", len(state.tasks))
	if state.cursor < len(state.tasks) {
		statusLine += fmt.Sprintf("  Cursor: %d/%d", state.cursor+1, len(state.tasks))
	}

	if state.filter != "" {
		statusLine += fmt.Sprintf("  Filter: %s", state.filter)
	}

	screen = append(screen, go3270.Field{Row: statusRow, Col: 2, Content: statusLine, Color: go3270.White})

	screen = append(screen, go3270.Field{
		Row:     pfRow,
		Col:     2,
		Content: "S=Sel E=Edit D=Done I=WIP PF1=Help PF3=Back PF7/8=Scrl PF9=Flt PF11=Swap",
		Color:   go3270.Turquoise,
	})

	opts := go3270.ScreenOpts{
		Codepage: devInfo.Codepage(),
	}

	// Prepare initial values for command line
	initialValues := map[string]string{
		"command": state.command,
	}

	response, err := go3270.ShowScreenOpts(screen, initialValues, conn, opts)
	if err != nil {
		return nil, nil, err
	}

	// Save command line input
	state.command = strings.TrimSpace(response.Values["command"])

	// Handle command line first
	if state.command != "" {
		return state.handleCommand(state.command, state.taskListTransaction)
	}

	// Process ISPF-style line commands (S/E/D/I typed in S column).
	// S and E navigate immediately (first match wins). D and I are batched.
	var firstSelectIdx int = -1
	var firstSelectMode string
	var batchUpdates int

	for idx := 0; idx < endIdx-startIdx; idx++ {
		val := strings.TrimSpace(strings.ToUpper(response.Values[fmt.Sprintf("s_%d", idx)]))
		taskIdx := startIdx + idx
		if val == "" || taskIdx >= len(state.tasks) {
			continue
		}
		task := state.tasks[taskIdx]
		switch val {
		case "S", "E":
			if firstSelectIdx < 0 {
				firstSelectIdx = taskIdx
				firstSelectMode = val
			}
		case "D":
			_ = updateTaskFieldsViaMCP(ctx, state.server, task.ID, "Done", task.Priority, task.LongDescription)
			batchUpdates++
		case "I":
			_ = updateTaskFieldsViaMCP(ctx, state.server, task.ID, "In Progress", task.Priority, task.LongDescription)
			batchUpdates++
		}
	}

	if firstSelectIdx >= 0 {
		state.cursor = firstSelectIdx
		state.selectedTask = state.tasks[firstSelectIdx]
		if firstSelectMode == "E" {
			return state.taskEditorTransaction, state, nil
		}
		return state.taskDetailTransaction, state, nil
	}

	if batchUpdates > 0 {
		return state.taskListTransaction, state, nil
	}

	// Handle attention keys
	if response.AID == go3270.AIDPF1 {
		return state.helpTransaction, state, nil
	}

	if response.AID == go3270.AIDPF3 {
		state.command = ""
		state.filter = ""
		state.listOffset = 0

		return state.mainMenuTransaction, state, nil
	}

	// PF11: swap to previous session
	if response.AID == go3270.AIDPF11 {
		s := state.popSession()
		if s != nil {
			state.pushSession("Tasks", state.taskListTransaction)
			return s.tx, state, nil
		}
	}

	if response.AID == go3270.AIDPF7 {
		// Scroll up
		if state.cursor > 0 {
			state.cursor--
			if state.cursor < state.listOffset {
				state.listOffset = state.cursor
			}
		}

		return state.taskListTransaction, state, nil
	}

	if response.AID == go3270.AIDPF8 {
		// Scroll down
		if state.cursor < len(state.tasks)-1 {
			state.cursor++
			if state.cursor >= state.listOffset+maxVisible {
				state.listOffset = state.cursor - maxVisible + 1
			}
		}

		return state.taskListTransaction, state, nil
	}

	if response.AID == go3270.AIDEnter {
		// Cursor-position row selection: Response.Row is 0-based; Field.Row is 1-based
		taskIdx := response.Row - (t3270HeaderRow - 1) + state.listOffset
		if taskIdx >= 0 && taskIdx < len(state.tasks) {
			state.cursor = taskIdx
			state.selectedTask = state.tasks[taskIdx]
			return state.taskDetailTransaction, state, nil
		}

		if state.cursor < len(state.tasks) {
			state.selectedTask = state.tasks[state.cursor]
			return state.taskDetailTransaction, state, nil
		}
	}

	if response.AID == go3270.AIDPF2 {
		if state.cursor < len(state.tasks) {
			state.selectedTask = state.tasks[state.cursor]
			return state.taskEditorTransaction, state, nil
		}
	}

	// PF9: cycle status filter
	if response.AID == go3270.AIDPF9 {
		state.status = nextStatusFilter(state.status)
		state.cursor = 0
		state.listOffset = 0
		state.filter = ""
		return state.taskListTransaction, state, nil
	}

	// PF4=Done, PF5=In Progress, PF6=Todo, PF10=Review
	if response.AID == go3270.AIDPF4 {
		return state.updateTaskStatus("Done")
	}

	if response.AID == go3270.AIDPF5 {
		return state.updateTaskStatus("In Progress")
	}

	if response.AID == go3270.AIDPF6 {
		return state.updateTaskStatus("Todo")
	}

	if response.AID == go3270.AIDPF10 {
		return state.updateTaskStatus("Review")
	}

	if response.AID == go3270.AIDPF12 {
		state.command = ""
		state.filter = ""
		state.listOffset = 0

		return state.mainMenuTransaction, state, nil
	}

	// Default: stay on task list
	return state.taskListTransaction, state, nil
}

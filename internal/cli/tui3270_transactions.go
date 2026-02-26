// tui3270_transactions.go — Screen transactions for the 3270 TUI (task list, detail, config, scorecard, handoff, editor).
// Extracted from tui3270.go. Each transaction renders a screen and handles user input.
package cli

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/davidl71/exarp-go/internal/config"
	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/tools"
	"github.com/racingmars/go3270"
)

var scorecardScoreRe = regexp.MustCompile(`(\d+)\s*/\s*(\d+)`)
var scorecardPctRe = regexp.MustCompile(`(\d+)%`)

// 3270 task list column layout (fixed-width, aligned like mainframe job list).
const (
	t3270ColS        = 2
	t3270ColID       = 5
	t3270ColStatus   = 24
	t3270ColPriority = 37
	t3270ColContent  = 48
	t3270WidS        = 2
	t3270WidID       = 18
	t3270WidStatus   = 12
	t3270WidPriority = 10
	t3270WidContent  = 32

	t3270HeaderRow    = 4  // first data row in task list
	t3270StatusBarRow = 22 // status bar row
	t3270PFKeyRow     = 23 // PF key help row
)

// t3270MaxVisible returns the number of visible task rows based on terminal dimensions.
// Falls back to 18 (default 24-row terminal minus header/status/PF key rows).
func t3270MaxVisible(devInfo go3270.DevInfo) int {
	rows, _ := devInfo.AltDimensions()
	if rows < 24 {
		rows = 24
	}
	visible := rows - 6 // header(3) + status(1) + pfkeys(1) + margin(1)
	if visible < 10 {
		visible = 10
	}
	return visible
}

// t3270ContentMaxRow returns the last usable content row before the status bar.
func t3270ContentMaxRow(devInfo go3270.DevInfo) int {
	rows, _ := devInfo.AltDimensions()
	if rows < 24 {
		rows = 24
	}
	return rows - 4 // Reserve status bar, PF key row, and margin
}

// t3270StatusRow returns the status bar row based on terminal dimensions.
func t3270StatusRow(devInfo go3270.DevInfo) int {
	rows, _ := devInfo.AltDimensions()
	if rows < 24 {
		rows = 24
	}
	return rows - 2
}

// t3270PFRow returns the PF key help row based on terminal dimensions.
func t3270PFRow(devInfo go3270.DevInfo) int {
	rows, _ := devInfo.AltDimensions()
	if rows < 24 {
		rows = 24
	}
	return rows - 1
}

// showLoadingOverlay displays a "Loading..." message on the status bar without clearing the screen.
func showLoadingOverlay(conn net.Conn, devInfo go3270.DevInfo, message string) {
	row := t3270StatusRow(devInfo)
	loadingScreen := go3270.Screen{
		{Row: row, Col: 2, Content: t3270Pad(message, 40), Color: go3270.Yellow, Intense: true},
	}
	_, _ = go3270.ShowScreenOpts(loadingScreen, nil, conn, go3270.ScreenOpts{
		NoClear:    true,
		NoResponse: true,
		Codepage:   devInfo.Codepage(),
	})
}

// statusColor returns the go3270 color for a task status.
func statusColor(status string) go3270.Color {
	switch status {
	case "Done":
		return go3270.Green
	case "In Progress":
		return go3270.Yellow
	case "Todo":
		return go3270.Turquoise
	case "Review":
		return go3270.Pink
	default:
		return go3270.DefaultColor
	}
}

// priorityColor returns the go3270 color for a task priority.
func priorityColor(priority string) go3270.Color {
	switch strings.ToLower(priority) {
	case "high":
		return go3270.Red
	case "medium":
		return go3270.Yellow
	case "low":
		return go3270.Green
	default:
		return go3270.DefaultColor
	}
}

// scorecardLineColor returns a go3270 color based on score patterns in a scorecard line.
func scorecardLineColor(line string) go3270.Color {
	upper := strings.ToUpper(line)

	// Section headers
	if strings.HasPrefix(line, "===") {
		return go3270.Green
	}

	// Explicit pass/fail indicators
	if strings.Contains(upper, "PASS") || strings.Contains(line, "✓") || strings.Contains(line, "✅") {
		return go3270.Green
	}
	if strings.Contains(upper, "FAIL") || strings.Contains(line, "✗") || strings.Contains(line, "❌") {
		return go3270.Red
	}

	// Score pattern: "N/M" (e.g. "85/100")
	if m := scorecardScoreRe.FindStringSubmatch(line); len(m) == 3 {
		num, _ := strconv.Atoi(m[1])
		den, _ := strconv.Atoi(m[2])
		if den > 0 {
			pct := num * 100 / den
			if pct >= 80 {
				return go3270.Green
			}
			if pct >= 50 {
				return go3270.Yellow
			}
			return go3270.Red
		}
	}

	// Percentage pattern: "85%"
	if m := scorecardPctRe.FindStringSubmatch(line); len(m) == 2 {
		pct, _ := strconv.Atoi(m[1])
		if pct >= 80 {
			return go3270.Green
		}
		if pct >= 50 {
			return go3270.Yellow
		}
		return go3270.Red
	}

	return go3270.DefaultColor
}

// statusFilters is the ordered list of status values cycled by PF9.
var statusFilters = []string{"Todo", "In Progress", "Review", "Done", ""}

// nextStatusFilter returns the next status in the cycle after current.
func nextStatusFilter(current string) string {
	for i, s := range statusFilters {
		if s == current {
			return statusFilters[(i+1)%len(statusFilters)]
		}
	}
	return statusFilters[0]
}

// updateTaskStatus updates the cursor task's status via MCP and returns to the task list.
func (state *tui3270State) updateTaskStatus(newStatus string) (go3270.Tx, any, error) {
	if state.cursor >= len(state.tasks) {
		return state.taskListTransaction, state, nil
	}
	task := state.tasks[state.cursor]
	ctx := context.Background()
	if err := updateTaskFieldsViaMCP(ctx, state.server, task.ID, newStatus, task.Priority, task.LongDescription); err != nil {
		logError(ctx, "Error updating task status", "error", err, "task_id", task.ID)
	}
	return state.taskListTransaction, state, nil
}

// updateTaskStatusForSelected updates the selectedTask's status via MCP and returns to the task list.
func (state *tui3270State) updateTaskStatusForSelected(newStatus string) (go3270.Tx, any, error) {
	if state.selectedTask == nil {
		return state.taskListTransaction, state, nil
	}
	ctx := context.Background()
	if err := updateTaskFieldsViaMCP(ctx, state.server, state.selectedTask.ID, newStatus, state.selectedTask.Priority, state.selectedTask.LongDescription); err != nil {
		logError(ctx, "Error updating task status", "error", err, "task_id", state.selectedTask.ID)
	}
	return state.taskListTransaction, state, nil
}

func t3270Pad(s string, width int) string {
	if len(s) >= width {
		if width <= 3 {
			return s[:width]
		}

		return s[:width-3] + "..."
	}

	return s + strings.Repeat(" ", width-len(s))
}

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

// taskDetailTransaction shows task details.
func (state *tui3270State) taskDetailTransaction(conn net.Conn, devInfo go3270.DevInfo, data any) (go3270.Tx, any, error) {
	if state.selectedTask == nil {
		return state.taskListTransaction, state, nil
	}

	task := state.selectedTask

	screen := go3270.Screen{
		{Row: 1, Col: 2, Content: "TASK DETAILS", Intense: true, Color: go3270.Blue},
		{Row: 3, Col: 2, Content: "Task ID:", Intense: true},
		{Row: 3, Col: 12, Content: task.ID, Color: go3270.Turquoise},
		{Row: 4, Col: 2, Content: "Status:", Intense: true},
		{Row: 4, Col: 12, Content: task.Status, Color: statusColor(task.Status)},
		{Row: 5, Col: 2, Content: "Priority:", Intense: true},
		{Row: 5, Col: 12, Content: task.Priority, Color: priorityColor(task.Priority)},
	}

	row := 7
	if task.Content != "" {
		screen = append(screen, go3270.Field{
			Row:     row,
			Col:     2,
			Content: "Content:",
			Intense: true,
		})
		row++

		for _, line := range splitIntoLines(task.Content, 4, 76) {
			if strings.TrimSpace(line) == "" {
				break
			}

			screen = append(screen, go3270.Field{Row: row, Col: 2, Content: line})
			row++
		}
	}

	contentMax := t3270ContentMaxRow(devInfo)
	if task.LongDescription != "" {
		screen = append(screen, go3270.Field{
			Row:     row,
			Col:     2,
			Content: "Description:",
			Intense: true,
		})
		row++

		maxDescLines := contentMax - row
		if maxDescLines < 1 {
			maxDescLines = 1
		}

		for _, line := range splitIntoLines(task.LongDescription, maxDescLines, 76) {
			if row >= contentMax || strings.TrimSpace(line) == "" {
				break
			}

			screen = append(screen, go3270.Field{Row: row, Col: 2, Content: line})
			row++
		}
	}

	if len(task.Tags) > 0 {
		screen = append(screen, go3270.Field{
			Row:     row,
			Col:     2,
			Content: "Tags:",
			Intense: true,
		})
		screen = append(screen, go3270.Field{
			Row:     row,
			Col:     8,
			Content: strings.Join(task.Tags, ", "),
		})
		row++
	}

	if len(task.Dependencies) > 0 {
		screen = append(screen, go3270.Field{
			Row:     row,
			Col:     2,
			Content: "Dependencies:",
			Intense: true,
		})
		screen = append(screen, go3270.Field{
			Row:     row,
			Col:     15,
			Content: strings.Join(task.Dependencies, ", "),
		})
		row++
	}

	detailPFRow := t3270PFRow(devInfo)
	screen = append(screen, go3270.Field{
		Row:     detailPFRow,
		Col:     2,
		Content: "PF1=Help PF2=Edit PF3=Back PF4=Done PF5=WIP PF6=Todo PF10=Review PF12=Cancel",
		Color:   go3270.Turquoise,
	})

	opts := go3270.ScreenOpts{
		Codepage: devInfo.Codepage(),
	}

	response, err := go3270.ShowScreenOpts(screen, nil, conn, opts)
	if err != nil {
		return nil, nil, err
	}

	if response.AID == go3270.AIDPF1 {
		return state.helpTransaction, state, nil
	}

	if response.AID == go3270.AIDPF3 || response.AID == go3270.AIDPF12 {
		return state.taskListTransaction, state, nil
	}

	if response.AID == go3270.AIDPF2 {
		return state.taskEditorTransaction, state, nil
	}

	// PF4=Done, PF5=In Progress, PF6=Todo, PF10=Review (from detail view)
	if response.AID == go3270.AIDPF4 {
		return state.updateTaskStatusForSelected("Done")
	}
	if response.AID == go3270.AIDPF5 {
		return state.updateTaskStatusForSelected("In Progress")
	}
	if response.AID == go3270.AIDPF6 {
		return state.updateTaskStatusForSelected("Todo")
	}
	if response.AID == go3270.AIDPF10 {
		return state.updateTaskStatusForSelected("Review")
	}

	// Default: stay on detail screen
	return state.taskDetailTransaction, state, nil
}

// configTransaction shows configuration editor.
func (state *tui3270State) configTransaction(conn net.Conn, devInfo go3270.DevInfo, data any) (go3270.Tx, any, error) {
	cfg, err := config.LoadConfig(state.projectRoot)
	if err != nil {
		// Show error screen instead of failing the connection; user can PF3 back
		msg := err.Error()
		if len(msg) > 72 {
			msg = msg[:69] + "..."
		}

		cfgPFRow := t3270PFRow(devInfo)
		errScreen := go3270.Screen{
			{Row: 1, Col: 2, Content: "CONFIGURATION", Intense: true, Color: go3270.Blue},
			{Row: 3, Col: 2, Content: "Config could not be loaded (protobuf required).", Color: go3270.Red},
			{Row: 5, Col: 2, Content: "Run: exarp-go config init", Color: go3270.Green},
			{Row: 6, Col: 2, Content: "  or: exarp-go config convert yaml protobuf", Color: go3270.Green},
			{Row: 8, Col: 2, Content: msg, Color: go3270.Yellow},
			{Row: cfgPFRow, Col: 2, Content: "PF3=Back to menu", Color: go3270.Turquoise},
		}
		screenOpts := go3270.ScreenOpts{Codepage: devInfo.Codepage()}

		response, showErr := go3270.ShowScreenOpts(errScreen, nil, conn, screenOpts)
		if showErr != nil {
			return nil, nil, showErr
		}

		if response.AID == go3270.AIDPF1 {
			return state.helpTransaction, state, nil
		}

		return state.mainMenuTransaction, state, nil
	}

	cfgPFRow2 := t3270PFRow(devInfo)
	screen := go3270.Screen{
		{Row: 1, Col: 2, Content: "CONFIGURATION", Intense: true, Color: go3270.Blue},
		{Row: 3, Col: 2, Content: "Configuration sections:", Color: go3270.Green},
		{Row: 5, Col: 4, Content: "1. Timeouts"},
		{Row: 6, Col: 4, Content: "2. Thresholds"},
		{Row: 7, Col: 4, Content: "3. Tasks"},
		{Row: 8, Col: 4, Content: "4. Database"},
		{Row: 9, Col: 4, Content: "5. Security"},
		{Row: 11, Col: 2, Content: "Section:", Write: true, Name: "section"},
		{Row: cfgPFRow2, Col: 2, Content: "PF3=Back  Enter=Select", Color: go3270.Turquoise},
	}

	// Show some config values (simplified - just show that config is loaded)
	if cfg != nil {
		screen = append(screen, go3270.Field{
			Row:     13,
			Col:     2,
			Content: "Configuration loaded successfully",
		})
	}

	opts := go3270.ScreenOpts{
		Codepage: devInfo.Codepage(),
	}

	response, err := go3270.ShowScreenOpts(screen, nil, conn, opts)
	if err != nil {
		return nil, nil, err
	}

	if response.AID == go3270.AIDPF1 {
		return state.helpTransaction, state, nil
	}

	if response.AID == go3270.AIDPF3 {
		return state.mainMenuTransaction, state, nil
	}

	// For now, just go back to main menu
	// In a full implementation, you'd handle section selection
	return state.mainMenuTransaction, state, nil
}

// runRecommendation runs a recommendation command in projectRoot and returns output and error.
// Uses the shared recommendationToCommand from tui_commands.go.
func runRecommendation(projectRoot, rec string) (output string, err error) {
	name, args, ok := recommendationToCommand(rec)
	if !ok {
		return "", fmt.Errorf("no runnable command for this recommendation")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = projectRoot

	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), err
	}

	return string(out), nil
}

// scorecardTransaction shows project scorecard (Go scorecard when Go project + project overview).
// When scorecardFullModeNext is set (e.g. after running a recommendation), uses full checks so coverage is shown.
func (state *tui3270State) scorecardTransaction(conn net.Conn, devInfo go3270.DevInfo, data any) (go3270.Tx, any, error) {
	ctx := context.Background()

	scErrPFRow := t3270PFRow(devInfo)

	var combined strings.Builder

	var recommendations []string

	// Use full mode when returning from "Run #" so updated coverage/lint is shown
	useFullMode := state.scorecardFullModeNext
	state.scorecardFullModeNext = false

	showLoadingOverlay(conn, devInfo, "Loading scorecard...")

	if tools.IsGoProject() {
		scorecardOpts := &tools.ScorecardOptions{FastMode: !useFullMode}

		scorecard, err := tools.GenerateGoScorecard(ctx, state.projectRoot, scorecardOpts)
		if err != nil {
			errScreen := go3270.Screen{
				{Row: 2, Col: 2, Content: "SCORECARD", Intense: true, Color: go3270.Blue},
				{Row: 4, Col: 2, Content: "Error: " + err.Error(), Color: go3270.Red},
				{Row: scErrPFRow, Col: 2, Content: "PF3=Back", Color: go3270.Turquoise},
			}
			screenOpts := go3270.ScreenOpts{Codepage: devInfo.Codepage()}

			if _, showErr := go3270.ShowScreenOpts(errScreen, nil, conn, screenOpts); showErr != nil {
				return nil, nil, showErr
			}

			return state.mainMenuTransaction, state, nil
		}

		recommendations = scorecard.Recommendations

		combined.WriteString("=== Go Scorecard ===\n\n")
		combined.WriteString(tools.FormatGoScorecard(scorecard))
	}

	overviewText, err := tools.GetOverviewText(ctx, state.projectRoot)
	if err != nil {
		if combined.Len() > 0 {
			combined.WriteString("\n\n=== Project Overview ===\n\n(overview failed: ")
			combined.WriteString(err.Error())
			combined.WriteString(")")
		} else {
			errScreen := go3270.Screen{
				{Row: 2, Col: 2, Content: "SCORECARD", Intense: true, Color: go3270.Blue},
				{Row: 4, Col: 2, Content: "Error: " + err.Error(), Color: go3270.Red},
				{Row: scErrPFRow, Col: 2, Content: "PF3=Back", Color: go3270.Turquoise},
			}
			screenOpts := go3270.ScreenOpts{Codepage: devInfo.Codepage()}

			if _, showErr := go3270.ShowScreenOpts(errScreen, nil, conn, screenOpts); showErr != nil {
				return nil, nil, showErr
			}

			return state.mainMenuTransaction, state, nil
		}
	} else {
		if combined.Len() > 0 {
			combined.WriteString("\n\n")
		}

		combined.WriteString("=== Project Overview ===\n\n")
		combined.WriteString(overviewText)
	}

	state.scorecardRecs = recommendations
	text := combined.String()
	lines := strings.Split(text, "\n")
	scContentMax := t3270ContentMaxRow(devInfo)
	maxRows := scContentMax - 2 // Reserve title row and margin
	if maxRows < 10 {
		maxRows = 10
	}

	if len(lines) > maxRows {
		lines = lines[:maxRows]
	}

	scStatusRow := t3270StatusRow(devInfo)
	scPFRow := t3270PFRow(devInfo)

	screen := go3270.Screen{
		{Row: 1, Col: 2, Content: "PROJECT SCORECARD", Intense: true, Color: go3270.Blue},
		{Row: scPFRow, Col: 2, Content: "PF3=Back to menu", Color: go3270.Turquoise},
	}
	if len(state.scorecardRecs) > 0 {
		screen = append(screen,
			go3270.Field{Row: scStatusRow, Col: 2, Content: "Run # (1-" + strconv.Itoa(len(state.scorecardRecs)) + "):", Intense: true},
			go3270.Field{Row: scStatusRow, Col: 24, Write: true, Name: "run_rec", Content: ""},
		)
		screen[1] = go3270.Field{Row: scPFRow, Col: 2, Content: "PF3=Back  Enter=Run selected #"}
	}

	for i, line := range lines {
		if len(line) > 78 {
			line = line[:75] + "..."
		}

		screen = append(screen, go3270.Field{Row: 2 + i, Col: 2, Content: line, Color: scorecardLineColor(line)})
	}

	screenOpts := go3270.ScreenOpts{Codepage: devInfo.Codepage()}

	response, err := go3270.ShowScreenOpts(screen, nil, conn, screenOpts)
	if err != nil {
		return nil, nil, err
	}

	if response.AID == go3270.AIDPF1 {
		return state.helpTransaction, state, nil
	}

	if response.AID == go3270.AIDPF3 {
		return state.mainMenuTransaction, state, nil
	}

	if response.AID == go3270.AIDPF11 {
		s := state.popSession()
		if s != nil {
			state.pushSession("Scorecard", state.scorecardTransaction)
			return s.tx, state, nil
		}
	}

	// Check if user entered a recommendation number to run
	runRec := strings.TrimSpace(response.Values["run_rec"])
	if runRec != "" && len(state.scorecardRecs) > 0 {
		var n int
		if _, parseErr := fmt.Sscanf(runRec, "%d", &n); parseErr == nil && n >= 1 && n <= len(state.scorecardRecs) {
			rec := state.scorecardRecs[n-1]
			out, runErr := runRecommendation(state.projectRoot, rec)
			resultLines := strings.Split(strings.TrimSpace(out), "\n")

			if runErr != nil {
				resultLines = append([]string{"Error: " + runErr.Error()}, resultLines...)
			}

			resultMaxLines := scContentMax - 4
			if resultMaxLines < 10 {
				resultMaxLines = 10
			}
			if len(resultLines) > resultMaxLines {
				resultLines = resultLines[:resultMaxLines]
			}

			resultScreen := go3270.Screen{
				{Row: 1, Col: 2, Content: "IMPLEMENT RESULT (#" + runRec + ")", Intense: true, Color: go3270.Blue},
				{Row: 2, Col: 2, Content: rec, Color: go3270.Green},
				{Row: scPFRow, Col: 2, Content: "PF3=Back to scorecard", Color: go3270.Turquoise},
			}

			for i, line := range resultLines {
				if len(line) > 78 {
					line = line[:75] + "..."
				}

				resultScreen = append(resultScreen, go3270.Field{Row: 4 + i, Col: 2, Content: line})
			}

			resp, showErr := go3270.ShowScreenOpts(resultScreen, nil, conn, screenOpts)
			if showErr != nil {
				return nil, nil, showErr
			}

			if resp.AID == go3270.AIDPF1 {
				return state.helpTransaction, state, nil
			}
			// PF3 or any key -> back to scorecard; next load uses full checks so coverage updates
			state.scorecardFullModeNext = true

			return state.scorecardTransaction, state, nil
		}
	}

	return state.scorecardTransaction, state, nil
}

// handoffTransaction shows session handoff notes (from session tool, handoff list).
// Uses shared fetchHandoffs from tui_mcp_adapter.go.
func (state *tui3270State) handoffTransaction(conn net.Conn, devInfo go3270.DevInfo, data any) (go3270.Tx, any, error) {
	ctx := context.Background()

	entries, err := fetchHandoffs(ctx, state.server, 0)
	hoErrPFRow := t3270PFRow(devInfo)
	if err != nil {
		errScreen := go3270.Screen{
			{Row: 2, Col: 2, Content: "SESSION HANDOFFS", Intense: true, Color: go3270.Blue},
			{Row: 4, Col: 2, Content: "Error: " + err.Error(), Color: go3270.Red},
			{Row: hoErrPFRow, Col: 2, Content: "PF3=Back to menu", Color: go3270.Turquoise},
		}
		screenOpts := go3270.ScreenOpts{Codepage: devInfo.Codepage()}

		if _, showErr := go3270.ShowScreenOpts(errScreen, nil, conn, screenOpts); showErr != nil {
			return nil, nil, showErr
		}

		return state.mainMenuTransaction, state, nil
	}

	hoContentMax := t3270ContentMaxRow(devInfo)
	hoPFRow := t3270PFRow(devInfo)

	screen := go3270.Screen{
		{Row: 1, Col: 2, Content: "SESSION HANDOFFS", Intense: true, Color: go3270.Blue},
		{Row: hoPFRow, Col: 2, Content: "PF3=Back to menu", Color: go3270.Turquoise},
	}

	if len(entries) == 0 {
		screen = append(screen, go3270.Field{Row: 2, Col: 2, Content: "No handoff notes."})
	} else {
		row := 2
		for i, h := range entries {
			if row >= hoContentMax {
				break
			}
			header := fmt.Sprintf("Handoff %d: %s", i+1, h.Host)
			if h.Timestamp != "" {
				ts := h.Timestamp
				if len(ts) > 19 {
					ts = ts[:19]
				}
				header += " (" + ts + ")"
			}
			if len(header) > 78 {
				header = header[:75] + "..."
			}
			screen = append(screen, go3270.Field{Row: row, Col: 2, Content: header, Intense: true, Color: go3270.Turquoise})
			row++

			if h.Summary != "" {
				for _, line := range splitIntoLines(h.Summary, 3, 76) {
					if row >= hoContentMax || strings.TrimSpace(line) == "" {
						break
					}
					screen = append(screen, go3270.Field{Row: row, Col: 4, Content: line})
					row++
				}
			}
			row++
		}
	}

	screenOpts := go3270.ScreenOpts{Codepage: devInfo.Codepage()}

	response, err := go3270.ShowScreenOpts(screen, nil, conn, screenOpts)
	if err != nil {
		return nil, nil, err
	}

	if response.AID == go3270.AIDPF1 {
		return state.helpTransaction, state, nil
	}

	if response.AID == go3270.AIDPF3 {
		return state.mainMenuTransaction, state, nil
	}

	return state.handoffTransaction, state, nil
}

// validStatus returns a Validator that accepts known task statuses.
func validStatus() go3270.Validator {
	valid := map[string]bool{"Todo": true, "In Progress": true, "Review": true, "Done": true}
	return func(input string) bool {
		return valid[strings.TrimSpace(input)]
	}
}

// validPriority returns a Validator that accepts known task priorities.
func validPriority() go3270.Validator {
	valid := map[string]bool{"low": true, "medium": true, "high": true, "": true}
	return func(input string) bool {
		return valid[strings.ToLower(strings.TrimSpace(input))]
	}
}

// taskEditorTransaction shows ISPF-like task editor.
// Uses HandleScreenAlt for validation loop and field merging.
func (state *tui3270State) taskEditorTransaction(conn net.Conn, devInfo go3270.DevInfo, data any) (go3270.Tx, any, error) {
	if state.selectedTask == nil {
		return state.taskListTransaction, state, nil
	}

	task := state.selectedTask
	descLines := splitIntoLines(task.LongDescription, 10, 70)

	screen := go3270.Screen{
		{Row: 1, Col: 2, Content: "EDIT TASK", Intense: true, Color: go3270.Blue},
		{Row: 1, Col: 60, Content: "", Name: "errmsg", Color: go3270.Red},
		{Row: 3, Col: 2, Content: "Task ID:", Intense: true},
		{Row: 3, Col: 12, Content: task.ID, Color: go3270.Turquoise},
		{Row: 4, Col: 2, Content: "Status:", Intense: true},
		{Row: 4, Col: 12, Write: true, Name: "status", Content: task.Status, Color: go3270.Green},
		{Row: 4, Col: 40, Content: "(Todo/In Progress/Review/Done)", Color: go3270.Turquoise},
		{Row: 5, Col: 2, Content: "Priority:", Intense: true},
		{Row: 5, Col: 12, Write: true, Name: "priority", Content: task.Priority, Color: go3270.Green},
		{Row: 5, Col: 40, Content: "(low/medium/high)", Color: go3270.Turquoise},
		{Row: 7, Col: 2, Content: "Description:", Intense: true},
	}

	edContentMax := t3270ContentMaxRow(devInfo)
	edPFRow := t3270PFRow(devInfo)
	edDescMaxRow := edContentMax - 3 // Leave room for command line and spacing

	row := 8
	for i, line := range descLines {
		if row > edDescMaxRow {
			break
		}

		screen = append(screen, go3270.Field{
			Row:     row,
			Col:     2,
			Write:   true,
			Name:    fmt.Sprintf("desc_line%d", i+1),
			Content: line,
		})
		row++
	}

	screen = append(screen, go3270.Field{Row: edContentMax, Col: 2, Content: "Command ===>", Intense: true, Color: go3270.Green})
	screen = append(screen, go3270.Field{Row: edContentMax, Col: 15, Write: true, Name: "command", Content: state.command, Color: go3270.Turquoise})
	screen = append(screen, go3270.Field{
		Row:     edPFRow,
		Col:     2,
		Content: "PF3=Exit  Enter=Save  CANCEL=Abandon",
		Color:   go3270.Turquoise,
	})

	rules := go3270.Rules{
		"status":   {Validator: validStatus(), ErrorText: "Status must be: Todo, In Progress, Review, or Done"},
		"priority": {Validator: validPriority(), ErrorText: "Priority must be: low, medium, or high"},
	}

	initialValues := map[string]string{
		"status":   task.Status,
		"priority": task.Priority,
		"command":  state.command,
	}
	for i, line := range descLines {
		initialValues[fmt.Sprintf("desc_line%d", i+1)] = line
	}

	// HandleScreenAlt loops until validation passes on Enter/PF15, or PF3/PF1/PF12 exits immediately.
	response, err := go3270.HandleScreenAlt(
		screen, rules, initialValues,
		[]go3270.AID{go3270.AIDEnter, go3270.AIDPF15},
		[]go3270.AID{go3270.AIDPF3, go3270.AIDPF1, go3270.AIDPF12},
		"errmsg", 4, 12,
		conn, devInfo, devInfo.Codepage(),
	)
	if err != nil {
		return nil, nil, err
	}

	if response.AID == go3270.AIDPF1 {
		return state.helpTransaction, state, nil
	}

	if response.AID == go3270.AIDPF3 || response.AID == go3270.AIDPF12 {
		state.command = ""
		return state.taskListTransaction, state, nil
	}

	// Check command line for CANCEL
	state.command = strings.TrimSpace(response.Values["command"])
	cmd := strings.ToUpper(state.command)

	if cmd == "CANCEL" || cmd == "C" {
		state.command = ""
		return state.taskListTransaction, state, nil
	}

	if cmd != "" && cmd != "SAVE" && cmd != "S" {
		return state.handleCommand(state.command, state.taskEditorTransaction)
	}

	// Save: validation already passed via HandleScreenAlt
	ctx := context.Background()
	task.Status = strings.TrimSpace(response.Values["status"])
	task.Priority = strings.TrimSpace(response.Values["priority"])

	var descParts []string
	for i := 1; i <= 10; i++ {
		lineKey := fmt.Sprintf("desc_line%d", i)
		if line, ok := response.Values[lineKey]; ok && strings.TrimSpace(line) != "" {
			descParts = append(descParts, strings.TrimSpace(line))
		}
	}

	task.LongDescription = strings.Join(descParts, "\n")

	if err := updateTaskFieldsViaMCP(ctx, state.server, task.ID, task.Status, task.Priority, task.LongDescription); err != nil {
		logError(context.Background(), "Error updating task", "error", err, "operation", "updateTask", "task_id", task.ID)
	}

	state.command = ""

	return state.taskListTransaction, state, nil
}

// healthTransaction shows an SDSF-style system health dashboard.
func (state *tui3270State) healthTransaction(conn net.Conn, devInfo go3270.DevInfo, data any) (go3270.Tx, any, error) {
	ctx := context.Background()

	showLoadingOverlay(conn, devInfo, "Loading health data...")

	hPFRow := t3270PFRow(devInfo)
	hContentMax := t3270ContentMaxRow(devInfo)

	screen := go3270.Screen{
		{Row: 1, Col: 2, Content: "SYSTEM HEALTH / ACTIVITY (SDSF)", Intense: true, Color: go3270.Blue},
		{Row: hPFRow, Col: 2, Content: "PF1=Help  PF3=Back to menu", Color: go3270.Turquoise},
	}

	row := 3

	// Git status section
	screen = append(screen, go3270.Field{Row: row, Col: 2, Content: "--- Git Status ---", Intense: true, Color: go3270.Green})
	row++

	gitText, gitErr := callToolText(ctx, state.server, "health", map[string]interface{}{"action": "git"})
	if gitErr != nil {
		screen = append(screen, go3270.Field{Row: row, Col: 4, Content: "Error: " + gitErr.Error(), Color: go3270.Red})
		row++
	} else {
		for _, line := range strings.Split(strings.TrimSpace(gitText), "\n") {
			if row >= hContentMax-6 {
				break
			}
			if len(line) > 74 {
				line = line[:71] + "..."
			}
			color := go3270.DefaultColor
			if strings.Contains(line, "dirty") || strings.Contains(line, "modified") {
				color = go3270.Yellow
			} else if strings.Contains(line, "clean") || strings.Contains(line, "up to date") {
				color = go3270.Green
			}
			screen = append(screen, go3270.Field{Row: row, Col: 4, Content: line, Color: color})
			row++
		}
	}

	row++

	// Task counts section
	if row < hContentMax-4 {
		screen = append(screen, go3270.Field{Row: row, Col: 2, Content: "--- Task Counts ---", Intense: true, Color: go3270.Green})
		row++

		for _, st := range []string{"Todo", "In Progress", "Review", "Done"} {
			if row >= hContentMax {
				break
			}
			tasks, err := listTasksViaMCP(ctx, state.server, st)
			count := 0
			if err == nil {
				count = len(tasks)
			}
			line := fmt.Sprintf("%-12s %d", st+":", count)
			screen = append(screen, go3270.Field{Row: row, Col: 4, Content: line, Color: statusColor(st)})
			row++
		}
	}

	row++

	// Server status section
	if row < hContentMax-2 {
		screen = append(screen, go3270.Field{Row: row, Col: 2, Content: "--- Server ---", Intense: true, Color: go3270.Green})
		row++

		serverText, serverErr := callToolText(ctx, state.server, "health", map[string]interface{}{"action": "server"})
		if serverErr != nil {
			screen = append(screen, go3270.Field{Row: row, Col: 4, Content: "Error: " + serverErr.Error(), Color: go3270.Red})
		} else {
			for _, line := range strings.Split(strings.TrimSpace(serverText), "\n") {
				if row >= hContentMax {
					break
				}
				if len(line) > 74 {
					line = line[:71] + "..."
				}
				screen = append(screen, go3270.Field{Row: row, Col: 4, Content: line})
				row++
			}
		}
	}

	screenOpts := go3270.ScreenOpts{Codepage: devInfo.Codepage()}

	response, err := go3270.ShowScreenOpts(screen, nil, conn, screenOpts)
	if err != nil {
		return nil, nil, err
	}

	if response.AID == go3270.AIDPF1 {
		return state.helpTransaction, state, nil
	}

	if response.AID == go3270.AIDPF3 {
		return state.mainMenuTransaction, state, nil
	}

	if response.AID == go3270.AIDPF11 {
		s := state.popSession()
		if s != nil {
			state.pushSession("Health", state.healthTransaction)
			return s.tx, state, nil
		}
	}

	return state.healthTransaction, state, nil
}

// gitDashboardTransaction shows a git status/log dashboard.
func (state *tui3270State) gitDashboardTransaction(conn net.Conn, devInfo go3270.DevInfo, data any) (go3270.Tx, any, error) {
	ctx := context.Background()

	showLoadingOverlay(conn, devInfo, "Loading git data...")

	gPFRow := t3270PFRow(devInfo)
	gContentMax := t3270ContentMaxRow(devInfo)

	screen := go3270.Screen{
		{Row: 1, Col: 2, Content: "GIT DASHBOARD", Intense: true, Color: go3270.Blue},
		{Row: gPFRow, Col: 2, Content: "PF1=Help  PF3=Back  PF7/8=Scroll", Color: go3270.Turquoise},
	}

	row := 3

	// Branches section
	screen = append(screen, go3270.Field{Row: row, Col: 2, Content: "--- Branches ---", Intense: true, Color: go3270.Green})
	row++

	branchText, branchErr := callToolText(ctx, state.server, "git_tools", map[string]interface{}{"action": "branches"})
	if branchErr != nil {
		screen = append(screen, go3270.Field{Row: row, Col: 4, Content: "Error: " + branchErr.Error(), Color: go3270.Red})
		row++
	} else {
		for _, line := range strings.Split(strings.TrimSpace(branchText), "\n") {
			if row >= gContentMax/2 {
				break
			}
			if len(line) > 74 {
				line = line[:71] + "..."
			}
			color := go3270.DefaultColor
			if strings.Contains(line, "*") || strings.Contains(line, "current") {
				color = go3270.Green
			}
			screen = append(screen, go3270.Field{Row: row, Col: 4, Content: line, Color: color})
			row++
		}
	}

	row++

	// Recent commits section
	if row < gContentMax-2 {
		screen = append(screen, go3270.Field{Row: row, Col: 2, Content: "--- Recent Commits (last 10) ---", Intense: true, Color: go3270.Green})
		row++

		commitText, commitErr := callToolText(ctx, state.server, "git_tools", map[string]interface{}{"action": "commits", "count": 10})
		if commitErr != nil {
			screen = append(screen, go3270.Field{Row: row, Col: 4, Content: "Error: " + commitErr.Error(), Color: go3270.Red})
			row++
		} else {
			for _, line := range strings.Split(strings.TrimSpace(commitText), "\n") {
				if row >= gContentMax {
					break
				}
				if len(line) > 74 {
					line = line[:71] + "..."
				}
				color := go3270.DefaultColor
				if strings.HasPrefix(strings.TrimSpace(line), "* ") || strings.HasPrefix(strings.TrimSpace(line), "commit ") {
					color = go3270.Yellow
				}
				screen = append(screen, go3270.Field{Row: row, Col: 4, Content: line, Color: color})
				row++
			}
		}
	}

	screenOpts := go3270.ScreenOpts{Codepage: devInfo.Codepage()}

	response, err := go3270.ShowScreenOpts(screen, nil, conn, screenOpts)
	if err != nil {
		return nil, nil, err
	}

	if response.AID == go3270.AIDPF1 {
		return state.helpTransaction, state, nil
	}

	if response.AID == go3270.AIDPF3 {
		return state.mainMenuTransaction, state, nil
	}

	if response.AID == go3270.AIDPF11 {
		s := state.popSession()
		if s != nil {
			state.pushSession("Git", state.gitDashboardTransaction)
			return s.tx, state, nil
		}
	}

	return state.gitDashboardTransaction, state, nil
}

// sprintBoardTransaction shows a kanban-style sprint board with tasks grouped by status.
func (state *tui3270State) sprintBoardTransaction(conn net.Conn, devInfo go3270.DevInfo, data any) (go3270.Tx, any, error) {
	ctx := context.Background()

	showLoadingOverlay(conn, devInfo, "Loading sprint board...")

	sPFRow := t3270PFRow(devInfo)
	sContentMax := t3270ContentMaxRow(devInfo)

	// Load tasks for each status
	type column struct {
		status string
		tasks  []*database.Todo2Task
	}
	columns := []column{
		{"Todo", nil},
		{"In Progress", nil},
		{"Review", nil},
		{"Done", nil},
	}

	for i := range columns {
		tasks, err := listTasksViaMCP(ctx, state.server, columns[i].status)
		if err == nil {
			columns[i].tasks = tasks
		}
	}

	screen := go3270.Screen{
		{Row: 1, Col: 2, Content: "SPRINT BOARD", Intense: true, Color: go3270.Blue},
		{Row: sPFRow, Col: 2, Content: "PF1=Help  PF3=Back", Color: go3270.Turquoise},
	}

	// Column layout: 4 columns x 19 chars, with 1-char separators, starting at col 2
	// Col positions: 2, 22, 42, 62 (each 18 chars wide + 1 separator)
	colWidth := 18
	colStarts := []int{2, 22, 42, 62}

	// Column headers (row 3)
	for i, col := range columns {
		header := fmt.Sprintf("%-*s", colWidth, fmt.Sprintf("%s (%d)", col.status, len(col.tasks)))
		if len(header) > colWidth {
			header = header[:colWidth]
		}
		screen = append(screen, go3270.Field{
			Row: 3, Col: colStarts[i], Content: header,
			Intense: true, Color: statusColor(col.status),
		})
	}

	// Separator line (row 4)
	for i := range columns {
		screen = append(screen, go3270.Field{
			Row: 4, Col: colStarts[i], Content: strings.Repeat("-", colWidth),
			Color: go3270.Green,
		})
	}

	// Task rows (row 5 onwards)
	maxRows := sContentMax - 5
	if maxRows < 5 {
		maxRows = 5
	}

	for rowIdx := 0; rowIdx < maxRows; rowIdx++ {
		screenRow := 5 + rowIdx
		for colIdx, col := range columns {
			if rowIdx < len(col.tasks) {
				task := col.tasks[rowIdx]
				content := task.Content
				if content == "" {
					content = task.ID
				}
				if len(content) > colWidth {
					content = content[:colWidth-1] + "~"
				}
				screen = append(screen, go3270.Field{
					Row: screenRow, Col: colStarts[colIdx],
					Content: fmt.Sprintf("%-*s", colWidth, content),
				})
			}
		}
	}

	screenOpts := go3270.ScreenOpts{Codepage: devInfo.Codepage()}

	response, err := go3270.ShowScreenOpts(screen, nil, conn, screenOpts)
	if err != nil {
		return nil, nil, err
	}

	if response.AID == go3270.AIDPF1 {
		return state.helpTransaction, state, nil
	}

	if response.AID == go3270.AIDPF3 {
		return state.mainMenuTransaction, state, nil
	}

	if response.AID == go3270.AIDPF11 {
		s := state.popSession()
		if s != nil {
			state.pushSession("Board", state.sprintBoardTransaction)
			return s.tx, state, nil
		}
	}

	return state.sprintBoardTransaction, state, nil
}

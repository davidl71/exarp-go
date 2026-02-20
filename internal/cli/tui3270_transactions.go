// tui3270_transactions.go — Screen transactions for the 3270 TUI (task list, detail, config, scorecard, handoff, editor).
// Extracted from tui3270.go. Each transaction renders a screen and handles user input.
package cli

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/davidl71/exarp-go/internal/config"
	"github.com/davidl71/exarp-go/internal/tools"
	"github.com/racingmars/go3270"
)

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

	maxVisible := 18

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

	for row, i := t3270HeaderRow, startIdx; i < endIdx; row, i = row+1, i+1 {
		task := state.tasks[i]
		isCursor := i == state.cursor

		s := "  "
		cursorHL := go3270.DefaultHighlight
		if isCursor {
			s = "> "
			cursorHL = go3270.ReverseVideo
		}

		screen = append(screen, go3270.Field{Row: row, Col: t3270ColS, Content: t3270Pad(s, t3270WidS), Highlighting: cursorHL, Color: go3270.Green})
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

	// Row 22: status bar (white for metadata)
	statusLine := fmt.Sprintf("Tasks: %d", len(state.tasks))
	if state.cursor < len(state.tasks) {
		statusLine += fmt.Sprintf("  Cursor: %d/%d", state.cursor+1, len(state.tasks))
	}

	if state.filter != "" {
		statusLine += fmt.Sprintf("  Filter: %s", state.filter)
	}

	screen = append(screen, go3270.Field{Row: t3270StatusBarRow, Col: 2, Content: statusLine, Color: go3270.White})

	// Row 23: PF key line (turquoise for instructional text)
	screen = append(screen, go3270.Field{
		Row:     t3270PFKeyRow,
		Col:     2,
		Content: "PF1=Help PF3=Back PF7/8=Scroll PF9=Filter PF4=Done PF5=WIP PF6=Todo PF2=Edit",
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

	if response.AID == go3270.AIDPF7 {
		// Scroll up
		if state.cursor > 0 {
			state.cursor--
			// Adjust offset if cursor moved above visible area
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
			// Adjust offset if cursor moved below visible area
			if state.cursor >= state.listOffset+18 {
				state.listOffset = state.cursor - 17
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

	if task.LongDescription != "" {
		screen = append(screen, go3270.Field{
			Row:     row,
			Col:     2,
			Content: "Description:",
			Intense: true,
		})
		row++

		maxDescLines := 20 - row
		if maxDescLines < 1 {
			maxDescLines = 1
		}

		for _, line := range splitIntoLines(task.LongDescription, maxDescLines, 76) {
			if row >= 20 || strings.TrimSpace(line) == "" {
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

	// PF key line
	screen = append(screen, go3270.Field{
		Row:     t3270PFKeyRow,
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

		errScreen := go3270.Screen{
			{Row: 1, Col: 2, Content: "CONFIGURATION", Intense: true, Color: go3270.Blue},
			{Row: 3, Col: 2, Content: "Config could not be loaded (protobuf required).", Color: go3270.Red},
			{Row: 5, Col: 2, Content: "Run: exarp-go config init", Color: go3270.Green},
			{Row: 6, Col: 2, Content: "  or: exarp-go config convert yaml protobuf", Color: go3270.Green},
			{Row: 8, Col: 2, Content: msg, Color: go3270.Yellow},
			{Row: t3270StatusBarRow, Col: 2, Content: "PF3=Back to menu", Color: go3270.Turquoise},
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

	screen := go3270.Screen{
		{Row: 1, Col: 2, Content: "CONFIGURATION", Intense: true, Color: go3270.Blue},
		{Row: 3, Col: 2, Content: "Configuration sections:", Color: go3270.Green},
		{Row: 5, Col: 4, Content: "1. Timeouts"},
		{Row: 6, Col: 4, Content: "2. Thresholds"},
		{Row: 7, Col: 4, Content: "3. Tasks"},
		{Row: 8, Col: 4, Content: "4. Database"},
		{Row: 9, Col: 4, Content: "5. Security"},
		{Row: 11, Col: 2, Content: "Section:", Write: true, Name: "section"},
		{Row: t3270StatusBarRow, Col: 2, Content: "PF3=Back  Enter=Select", Color: go3270.Turquoise},
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

	var combined strings.Builder

	var recommendations []string

	// Use full mode when returning from "Run #" so updated coverage/lint is shown
	useFullMode := state.scorecardFullModeNext
	state.scorecardFullModeNext = false

	if tools.IsGoProject() {
		scorecardOpts := &tools.ScorecardOptions{FastMode: !useFullMode}

		scorecard, err := tools.GenerateGoScorecard(ctx, state.projectRoot, scorecardOpts)
		if err != nil {
			errScreen := go3270.Screen{
				{Row: 2, Col: 2, Content: "SCORECARD", Intense: true, Color: go3270.Blue},
				{Row: 4, Col: 2, Content: "Error: " + err.Error(), Color: go3270.Red},
				{Row: t3270StatusBarRow, Col: 2, Content: "PF3=Back", Color: go3270.Turquoise},
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
				{Row: t3270StatusBarRow, Col: 2, Content: "PF3=Back", Color: go3270.Turquoise},
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
	maxRows := 16

	if len(lines) > maxRows {
		lines = lines[:maxRows]
	}

	screen := go3270.Screen{
		{Row: 1, Col: 2, Content: "PROJECT SCORECARD", Intense: true, Color: go3270.Blue},
		{Row: t3270StatusBarRow, Col: 2, Content: "PF3=Back to menu", Color: go3270.Turquoise},
	}
	if len(state.scorecardRecs) > 0 {
		screen = append(screen,
			go3270.Field{Row: 20, Col: 2, Content: "Run # (1-" + strconv.Itoa(len(state.scorecardRecs)) + "):", Intense: true},
			go3270.Field{Row: 20, Col: 24, Write: true, Name: "run_rec", Content: ""},
		)
		screen[1] = go3270.Field{Row: 22, Col: 2, Content: "PF3=Back  Enter=Run selected #"}
	}

	for i, line := range lines {
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

	if response.AID == go3270.AIDPF1 {
		return state.helpTransaction, state, nil
	}

	if response.AID == go3270.AIDPF3 {
		return state.mainMenuTransaction, state, nil
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

			if len(resultLines) > 18 {
				resultLines = resultLines[:18]
			}

			resultScreen := go3270.Screen{
				{Row: 1, Col: 2, Content: "IMPLEMENT RESULT (#" + runRec + ")", Intense: true, Color: go3270.Blue},
				{Row: 2, Col: 2, Content: rec, Color: go3270.Green},
				{Row: t3270StatusBarRow, Col: 2, Content: "PF3=Back to scorecard", Color: go3270.Turquoise},
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
	if err != nil {
		errScreen := go3270.Screen{
			{Row: 2, Col: 2, Content: "SESSION HANDOFFS", Intense: true, Color: go3270.Blue},
			{Row: 4, Col: 2, Content: "Error: " + err.Error(), Color: go3270.Red},
			{Row: t3270StatusBarRow, Col: 2, Content: "PF3=Back to menu", Color: go3270.Turquoise},
		}
		screenOpts := go3270.ScreenOpts{Codepage: devInfo.Codepage()}

		if _, showErr := go3270.ShowScreenOpts(errScreen, nil, conn, screenOpts); showErr != nil {
			return nil, nil, showErr
		}

		return state.mainMenuTransaction, state, nil
	}

	screen := go3270.Screen{
		{Row: 1, Col: 2, Content: "SESSION HANDOFFS", Intense: true, Color: go3270.Blue},
		{Row: t3270StatusBarRow, Col: 2, Content: "PF3=Back to menu", Color: go3270.Turquoise},
	}

	if len(entries) == 0 {
		screen = append(screen, go3270.Field{Row: 2, Col: 2, Content: "No handoff notes."})
	} else {
		row := 2
		for i, h := range entries {
			if row >= 20 {
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
					if row >= 20 || strings.TrimSpace(line) == "" {
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

	row := 8
	for i, line := range descLines {
		if row > 17 {
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

	screen = append(screen, go3270.Field{Row: 20, Col: 2, Content: "Command ===>", Intense: true, Color: go3270.Green})
	screen = append(screen, go3270.Field{Row: 20, Col: 15, Write: true, Name: "command", Content: state.command, Color: go3270.Turquoise})
	screen = append(screen, go3270.Field{
		Row:     t3270PFKeyRow,
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

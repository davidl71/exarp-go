// tui3270_transactions.go — Screen transactions for the 3270 TUI (task list, detail, config, scorecard, handoff, editor).
// Extracted from tui3270.go. Each transaction renders a screen and handles user input.
package cli

import (
	"context"
	"encoding/json"
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
)

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
		{Row: 1, Col: 2, Content: title, Intense: true},
		{Row: 1, Col: 45, Content: fmt.Sprintf("LINE %d OF %d", currentLine, totalLines), Intense: true},
		{Row: 1, Col: 55, Content: "COMMAND ===>", Intense: true},
		{Row: 1, Col: 68, Write: true, Name: "command", Content: state.command},
		{Row: 1, Col: 72, Content: "SCROLL ===> " + scrollIndicator, Intense: true},
	}

	// Row 2: blank
	screen = append(screen, go3270.Field{Row: 2, Col: 2, Content: ""})

	// Row 3: column headers (intense, like yellow headers)
	screen = append(screen, go3270.Field{Row: 3, Col: t3270ColS, Content: t3270Pad("S", t3270WidS), Intense: true})
	screen = append(screen, go3270.Field{Row: 3, Col: t3270ColID, Content: t3270Pad("ID", t3270WidID), Intense: true})
	screen = append(screen, go3270.Field{Row: 3, Col: t3270ColStatus, Content: t3270Pad("STATUS", t3270WidStatus), Intense: true})
	screen = append(screen, go3270.Field{Row: 3, Col: t3270ColPriority, Content: t3270Pad("PRIORITY", t3270WidPriority), Intense: true})
	screen = append(screen, go3270.Field{Row: 3, Col: t3270ColContent, Content: t3270Pad("CONTENT", t3270WidContent), Intense: true})

	startIdx := state.listOffset

	endIdx := startIdx + maxVisible
	if endIdx > len(state.tasks) {
		endIdx = len(state.tasks)
	}

	if endIdx < startIdx {
		endIdx = startIdx
	}

	for row, i := 4, startIdx; i < endIdx; row, i = row+1, i+1 {
		task := state.tasks[i]

		s := "  "
		if i == state.cursor {
			s = "> "
		}

		screen = append(screen, go3270.Field{Row: row, Col: t3270ColS, Content: t3270Pad(s, t3270WidS)})
		screen = append(screen, go3270.Field{Row: row, Col: t3270ColID, Content: t3270Pad(task.ID, t3270WidID)})

		st := task.Status
		if st == "" {
			st = "-"
		}

		screen = append(screen, go3270.Field{Row: row, Col: t3270ColStatus, Content: t3270Pad(st, t3270WidStatus)})

		pri := strings.ToUpper(task.Priority)
		if pri == "" {
			pri = "-"
		}

		screen = append(screen, go3270.Field{Row: row, Col: t3270ColPriority, Content: t3270Pad(pri, t3270WidPriority)})

		content := task.Content
		if content == "" {
			content = task.LongDescription
		}

		if content == "" {
			content = "(no description)"
		}

		screen = append(screen, go3270.Field{Row: row, Col: t3270ColContent, Content: t3270Pad(content, t3270WidContent)})
	}

	// Row 22: status left, help text
	statusLine := fmt.Sprintf("Tasks: %d", len(state.tasks))
	if state.cursor < len(state.tasks) {
		statusLine += fmt.Sprintf("  Cursor: %d/%d", state.cursor+1, len(state.tasks))
	}

	if state.filter != "" {
		statusLine += fmt.Sprintf("  Filter: %s", state.filter)
	}

	screen = append(screen, go3270.Field{Row: 22, Col: 2, Content: statusLine})

	// Row 23: PF key line
	screen = append(screen, go3270.Field{
		Row:     23,
		Col:     2,
		Content: "PF1=Help  PF3=Back  PF7=Up  PF8=Down  Enter=Select  PF12=Cancel  PF2=Edit",
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

	if response.AID == go3270.AIDEnter || response.AID == go3270.AIDPF1 {
		if state.cursor < len(state.tasks) {
			state.selectedTask = state.tasks[state.cursor]
			return state.taskDetailTransaction, state, nil
		}
	}

	if response.AID == go3270.AIDPF2 {
		// Edit selected task
		if state.cursor < len(state.tasks) {
			state.selectedTask = state.tasks[state.cursor]
			return state.taskEditorTransaction, state, nil
		}
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
		{Row: 1, Col: 2, Content: "TASK DETAILS", Intense: true},
		{Row: 3, Col: 2, Content: "Task ID:", Intense: true},
		{Row: 3, Col: 12, Content: task.ID},
		{Row: 4, Col: 2, Content: "Status:", Intense: true},
		{Row: 4, Col: 12, Content: task.Status},
		{Row: 5, Col: 2, Content: "Priority:", Intense: true},
		{Row: 5, Col: 12, Content: task.Priority},
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
		// Split content into multiple lines if needed
		content := task.Content
		for len(content) > 0 {
			lineLen := 76
			if len(content) < lineLen {
				lineLen = len(content)
			}

			screen = append(screen, go3270.Field{
				Row:     row,
				Col:     2,
				Content: content[:lineLen],
			})
			row++

			if len(content) > lineLen {
				content = content[lineLen:]
			} else {
				break
			}
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
		// Split description into multiple lines
		desc := task.LongDescription
		for len(desc) > 0 && row < 20 {
			lineLen := 76
			if len(desc) < lineLen {
				lineLen = len(desc)
			}

			screen = append(screen, go3270.Field{
				Row:     row,
				Col:     2,
				Content: desc[:lineLen],
			})
			row++

			if len(desc) > lineLen {
				desc = desc[lineLen:]
			} else {
				break
			}
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

	// Add help line
	screen = append(screen, go3270.Field{
		Row:     23,
		Col:     2,
		Content: "PF3=Back  PF12=Cancel",
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
		// Edit task
		return state.taskEditorTransaction, state, nil
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
			{Row: 1, Col: 2, Content: "CONFIGURATION", Intense: true},
			{Row: 3, Col: 2, Content: "Config could not be loaded (protobuf required)."},
			{Row: 5, Col: 2, Content: "Run: exarp-go config init"},
			{Row: 6, Col: 2, Content: "  or: exarp-go config convert yaml protobuf"},
			{Row: 8, Col: 2, Content: msg},
			{Row: 22, Col: 2, Content: "PF3=Back to menu"},
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
		{Row: 1, Col: 2, Content: "CONFIGURATION", Intense: true},
		{Row: 3, Col: 2, Content: "Configuration sections:"},
		{Row: 5, Col: 4, Content: "1. Timeouts"},
		{Row: 6, Col: 4, Content: "2. Thresholds"},
		{Row: 7, Col: 4, Content: "3. Tasks"},
		{Row: 8, Col: 4, Content: "4. Database"},
		{Row: 9, Col: 4, Content: "5. Security"},
		{Row: 11, Col: 2, Content: "Section:", Write: true, Name: "section"},
		{Row: 22, Col: 2, Content: "PF3=Back  Enter=Select"},
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

// recommendationToCommand3270 maps scorecard recommendation text to (name, args); prefers Makefile targets; ok false if not runnable.
func recommendationToCommand3270(rec string) (name string, args []string, ok bool) {
	rec = strings.TrimSpace(rec)

	switch {
	case strings.Contains(rec, "go mod tidy"):
		return "make", []string{"go-mod-tidy"}, true
	case strings.Contains(rec, "go fmt"):
		return "make", []string{"go-fmt"}, true
	case strings.Contains(rec, "go vet"):
		return "make", []string{"go-vet"}, true
	case strings.Contains(rec, "Fix Go build"):
		return "make", []string{"build"}, true
	case strings.Contains(rec, "golangci-lint issues"):
		return "make", []string{"golangci-lint-fix"}, true
	case strings.Contains(rec, "failing Go tests"), strings.Contains(rec, "go test"):
		return "make", []string{"test"}, true
	case strings.Contains(rec, "govulncheck"):
		return "make", []string{"govulncheck"}, true
	case strings.Contains(rec, "test coverage"):
		return "make", []string{"test-coverage"}, true
	default:
		return "", nil, false
	}
}

// runRecommendation3270 runs a recommendation command in projectRoot and returns output and error.
func runRecommendation3270(projectRoot, rec string) (output string, err error) {
	name, args, ok := recommendationToCommand3270(rec)
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
				{Row: 2, Col: 2, Content: "SCORECARD", Intense: true},
				{Row: 4, Col: 2, Content: "Error: " + err.Error()},
				{Row: 22, Col: 2, Content: "PF3=Back"},
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
				{Row: 2, Col: 2, Content: "SCORECARD", Intense: true},
				{Row: 4, Col: 2, Content: "Error: " + err.Error()},
				{Row: 22, Col: 2, Content: "PF3=Back"},
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
		{Row: 1, Col: 2, Content: "PROJECT SCORECARD", Intense: true},
		{Row: 22, Col: 2, Content: "PF3=Back to menu"},
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
			out, runErr := runRecommendation3270(state.projectRoot, rec)
			resultLines := strings.Split(strings.TrimSpace(out), "\n")

			if runErr != nil {
				resultLines = append([]string{"Error: " + runErr.Error()}, resultLines...)
			}

			if len(resultLines) > 18 {
				resultLines = resultLines[:18]
			}

			resultScreen := go3270.Screen{
				{Row: 1, Col: 2, Content: "IMPLEMENT RESULT (#" + runRec + ")", Intense: true},
				{Row: 2, Col: 2, Content: rec},
				{Row: 22, Col: 2, Content: "PF3=Back to scorecard"},
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
func (state *tui3270State) handoffTransaction(conn net.Conn, devInfo go3270.DevInfo, data any) (go3270.Tx, any, error) {
	ctx := context.Background()
	args := map[string]interface{}{"action": "handoff", "sub_action": "list"}

	argsBytes, err := json.Marshal(args)
	if err != nil {
		errScreen := go3270.Screen{
			{Row: 2, Col: 2, Content: "SESSION HANDOFFS", Intense: true},
			{Row: 4, Col: 2, Content: "Error: " + err.Error()},
			{Row: 22, Col: 2, Content: "PF3=Back to menu"},
		}
		screenOpts := go3270.ScreenOpts{Codepage: devInfo.Codepage()}

		if _, showErr := go3270.ShowScreenOpts(errScreen, nil, conn, screenOpts); showErr != nil {
			return nil, nil, showErr
		}

		return state.mainMenuTransaction, state, nil
	}

	result, err := state.server.CallTool(ctx, "session", argsBytes)
	if err != nil {
		errScreen := go3270.Screen{
			{Row: 2, Col: 2, Content: "SESSION HANDOFFS", Intense: true},
			{Row: 4, Col: 2, Content: "Error: " + err.Error()},
			{Row: 22, Col: 2, Content: "PF3=Back to menu"},
		}
		screenOpts := go3270.ScreenOpts{Codepage: devInfo.Codepage()}

		if _, showErr := go3270.ShowScreenOpts(errScreen, nil, conn, screenOpts); showErr != nil {
			return nil, nil, showErr
		}

		return state.mainMenuTransaction, state, nil
	}

	var text strings.Builder
	for _, c := range result {
		text.WriteString(c.Text)
		text.WriteString("\n")
	}

	content := strings.TrimSpace(text.String())
	lines := strings.Split(content, "\n")
	maxRows := 18

	if len(lines) > maxRows {
		lines = lines[:maxRows]
	}

	screen := go3270.Screen{
		{Row: 1, Col: 2, Content: "SESSION HANDOFFS", Intense: true},
		{Row: 22, Col: 2, Content: "PF3=Back to menu"},
	}

	for i, line := range lines {
		if len(line) > 78 {
			line = line[:75] + "..."
		}

		screen = append(screen, go3270.Field{Row: 2 + i, Col: 2, Content: line})
	}

	if len(lines) == 0 {
		screen = append(screen, go3270.Field{Row: 2, Col: 2, Content: "No handoff notes."})
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

// taskEditorTransaction shows ISPF-like task editor.
func (state *tui3270State) taskEditorTransaction(conn net.Conn, devInfo go3270.DevInfo, data any) (go3270.Tx, any, error) {
	if state.selectedTask == nil {
		return state.taskListTransaction, state, nil
	}

	task := state.selectedTask

	// Split description into lines for editing (max 10 lines)
	descLines := splitIntoLines(task.LongDescription, 10, 70)

	screen := go3270.Screen{
		{Row: 1, Col: 2, Content: "EDIT TASK", Intense: true},
		{Row: 3, Col: 2, Content: "Task ID:", Intense: true},
		{Row: 3, Col: 12, Content: task.ID}, // Read-only
		{Row: 4, Col: 2, Content: "Status:", Intense: true},
		{Row: 4, Col: 12, Write: true, Name: "status", Content: task.Status},
		{Row: 5, Col: 2, Content: "Priority:", Intense: true},
		{Row: 5, Col: 12, Write: true, Name: "priority", Content: task.Priority},
		{Row: 7, Col: 2, Content: "Description:", Intense: true},
	}

	// Add description lines as editable fields
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

	// Add command line
	screen = append(screen, go3270.Field{
		Row:     22,
		Col:     2,
		Content: "Command ===>",
		Intense: true,
	})
	screen = append(screen, go3270.Field{
		Row:     22,
		Col:     15,
		Write:   true,
		Name:    "command",
		Content: state.command,
	})

	// Add help line
	screen = append(screen, go3270.Field{
		Row:     23,
		Col:     2,
		Content: "PF3=Exit  PF15=Save  Enter=Execute Command  CANCEL=Abandon",
	})

	opts := go3270.ScreenOpts{
		Codepage: devInfo.Codepage(),
	}

	// Prepare initial values
	initialValues := map[string]string{
		"status":   task.Status,
		"priority": task.Priority,
		"command":  state.command,
	}
	for i, line := range descLines {
		initialValues[fmt.Sprintf("desc_line%d", i+1)] = line
	}

	response, err := go3270.ShowScreenOpts(screen, initialValues, conn, opts)
	if err != nil {
		return nil, nil, err
	}

	// Save command line
	state.command = strings.TrimSpace(response.Values["command"])

	// Handle commands
	cmd := strings.ToUpper(state.command)
	if cmd == "CANCEL" || cmd == "C" {
		state.command = ""
		return state.taskListTransaction, state, nil
	}

	if cmd == "SAVE" || cmd == "S" || response.AID == go3270.AIDPF15 {
		// Save changes
		ctx := context.Background()
		task.Status = strings.TrimSpace(response.Values["status"])
		task.Priority = strings.TrimSpace(response.Values["priority"])

		// Reconstruct description from lines
		var descParts []string

		for i := 1; i <= 10; i++ {
			lineKey := fmt.Sprintf("desc_line%d", i)
			if line, ok := response.Values[lineKey]; ok && strings.TrimSpace(line) != "" {
				descParts = append(descParts, strings.TrimSpace(line))
			}
		}

		task.LongDescription = strings.Join(descParts, "\n")

		// Update task via MCP
		if err := updateTaskFieldsViaMCP(ctx, state.server, task.ID, task.Status, task.Priority, task.LongDescription); err != nil {
			logError(context.Background(), "Error updating task", "error", err, "operation", "updateTask", "task_id", task.ID)
		} else {
			state.command = ""
			return state.taskListTransaction, state, nil
		}
	}

	// Handle other commands
	if state.command != "" {
		return state.handleCommand(state.command, state.taskEditorTransaction)
	}

	// Handle PF keys
	if response.AID == go3270.AIDPF1 {
		return state.helpTransaction, state, nil
	}

	if response.AID == go3270.AIDPF3 {
		state.command = ""
		return state.taskListTransaction, state, nil
	}

	// Default: stay in editor
	return state.taskEditorTransaction, state, nil
}

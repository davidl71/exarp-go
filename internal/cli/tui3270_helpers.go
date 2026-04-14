// tui3270_helpers.go — Command handling, help screen, utility transactions, and display helpers for the 3270 TUI.
// Extracted from tui3270.go and tui3270_transactions.go.
// Handles ISPF-style command parsing, help overlay, child agent results, color helpers, and layout constants.
package cli

import (
	"context"
	"fmt"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/davidl71/exarp-go/internal/database"
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

	t3270HeaderRow    = 4  // first data row in task list (after ISPF rule row, title line, dash, column headers)
	t3270StatusBarRow = 22 // status bar row
	t3270PFKeyRow     = 23 // PF key help row
)

// t3270ScreenCols returns the 3270 screen width; at least 80 for layout math.
func t3270ScreenCols(devInfo go3270.DevInfo) int {
	_, cols := devInfo.AltDimensions()
	if cols < 80 {
		cols = 80
	}
	return cols
}

// t3270ISPFRuleLine builds a full-width dashed rule with an embedded label (ISPF / 3270BBS style).
func t3270ISPFRuleLine(width int, label string) string {
	if width < 20 {
		width = 80
	}
	label = strings.TrimSpace(label)
	if label == "" {
		return strings.Repeat("-", width)
	}
	maxLabel := width - 6
	if len(label) > maxLabel {
		label = label[:maxLabel]
	}
	pad := width - len(label) - 2
	if pad < 4 {
		if len(label) > width {
			return label[:width]
		}
		return label + strings.Repeat("-", width-len(label))
	}
	left := pad / 2
	right := pad - left
	return strings.Repeat("-", left) + " " + label + " " + strings.Repeat("-", right)
}

// t3270MenuBoxTopBottom returns boxed panel top/bottom lines and inner width between borders.
// The box starts at column leftEdge (0-based); the right '+' ends at column cols-1.
func t3270MenuBoxTopBottom(devInfo go3270.DevInfo) (top, bottom string, rightPipeCol, inner int) {
	cols := t3270ScreenCols(devInfo)
	leftEdge := 2 // 0-based column of '+'
	inner = cols - leftEdge - 2 // inner '-' count between '+' corners
	if inner < 40 {
		inner = 40
	}
	top = "+" + strings.Repeat("-", inner) + "+"
	bottom = top
	rightPipeCol = leftEdge + len(top) - 1
	return top, bottom, rightPipeCol, inner
}

// t3270ISPFTitleColor is used for panel names (ISPF-style: intense white vs. legacy blue menus).
func t3270ISPFTitleColor() go3270.Color {
	if noColor3270 {
		return go3270.DefaultColor
	}
	return go3270.White
}

// t3270PanelRuleColor is used for ISPF-style rules and box borders (classic green 3270).
func t3270PanelRuleColor() go3270.Color {
	if noColor3270 {
		return go3270.DefaultColor
	}
	return go3270.Green
}

// t3270ISPFOptionLine builds "NN  description..." truncated to fit between menu box borders (ISPF option list style).
func t3270ISPFOptionLine(option int, description string, innerCols int) string {
	if innerCols < 12 {
		innerCols = 40
	}
	prefix := fmt.Sprintf("%-2d  ", option)
	maxDesc := innerCols - len(prefix)
	if maxDesc < 1 {
		maxDesc = 1
	}
	d := strings.TrimSpace(description)
	for len(d) > maxDesc {
		if maxDesc <= 3 {
			d = d[:maxDesc]
			break
		}
		d = d[:maxDesc-1] + "~"
	}
	return prefix + d
}

// t3270ISPFPanelBannerLine builds a two-part banner: left product text + right panel ID (ISPF panel header strip).
func t3270ISPFPanelBannerLine(cols int, productLeft, panelID string) string {
	if cols < 40 {
		cols = 80
	}
	panelID = strings.TrimSpace(panelID)
	productLeft = strings.TrimSpace(productLeft)
	right := panelID
	if right != "" && !strings.HasPrefix(strings.ToUpper(right), "PANEL") {
		right = "Panel " + right
	}
	leftMax := cols - len(right) - 3
	if leftMax < 10 {
		leftMax = 10
	}
	if len(productLeft) > leftMax {
		productLeft = productLeft[:leftMax-1] + "~"
	}
	pad := cols - len(productLeft) - len(right)
	if pad < 1 {
		pad = 1
	}
	return productLeft + strings.Repeat(" ", pad) + right
}

// t3270MaxVisible returns the number of visible task rows based on terminal dimensions.
// Falls back to 18 (default 24-row terminal minus header/status/PF key rows).
func t3270MaxVisible(devInfo go3270.DevInfo) int {
	rows, _ := devInfo.AltDimensions()
	if rows < 24 {
		rows = 24
	}
	// Top: ISPF rule + title + dash + column headers (4 rows before data at t3270HeaderRow).
	// Bottom: status + two PF-key rows on task list.
	visible := rows - 7
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
	rows, _ := devInfo.AltDimensions()
	if rows < 24 {
		rows = 24
	}
	// Same row as task list ISPF status line (third line from bottom when PF legend uses 2 rows).
	row := rows - 3
	loadingScreen := go3270.Screen{
		{Row: row, Col: 2, Content: t3270Pad(message, 40), Color: go3270.Yellow, Intense: true},
	}
	_, _ = go3270.ShowScreenOpts(loadingScreen, nil, conn, go3270.ScreenOpts{
		NoClear:    true,
		NoResponse: true,
		Codepage:   devInfo.Codepage(),
	})
}

// noColor3270 is true when the NO_COLOR env var is set (https://no-color.org).
var noColor3270 = os.Getenv("NO_COLOR") != ""

// statusColor returns the go3270 color for a task status.
func statusColor(status string) go3270.Color {
	if noColor3270 {
		return go3270.DefaultColor
	}
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
	if noColor3270 {
		return go3270.DefaultColor
	}
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
	if noColor3270 {
		return go3270.DefaultColor
	}
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

// t3270Pad is an alias for truncatePad (used in 3270 screen building).
// Now uses the shared implementation from tui_helpers.go.
var t3270Pad = truncatePad

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

// loadTasksForStatus loads tasks by status via MCP adapter.
// When status is empty, returns open tasks only (Todo + In Progress).
func (state *tui3270State) loadTasksForStatus(ctx context.Context, status string) ([]*database.Todo2Task, error) {
	return listTasksViaMCP(ctx, state.server, status)
}

// showChildAgentResultTransaction shows a one-screen result then returns to nextTx.
func (state *tui3270State) showChildAgentResultTransaction(message string, nextTx go3270.Tx) go3270.Tx {
	return func(conn net.Conn, devInfo go3270.DevInfo, data any) (go3270.Tx, any, error) {
		cols := t3270ScreenCols(devInfo)
		gc := t3270PanelRuleColor()
		pf := t3270PFRow(devInfo)
		screen := go3270.Screen{
			{Row: 0, Col: 0, Content: t3270ISPFRuleLine(cols, " CHILD AGENT RESULT "), Color: gc},
			{Row: 2, Col: 2, Content: "CHILD AGENT", Intense: true, Color: t3270ISPFTitleColor()},
			{Row: 4, Col: 2, Content: message, Color: go3270.Green},
			{Row: pf, Col: 2, Content: "PF01=Help  PF03=Back to menu", Color: go3270.Turquoise},
		}
		if len(message) > 76 {
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
		"Main menu: 1=Tasks 2=Config 3=Scorecard 4=Handoffs 5=Exit 6=Agent 7=Health",
		"",
		"Commands (type in COMMAND ===> field):",
		"  TASKS/T  CONFIG  SC  HANDOFFS/HO  MENU/M  HELP/H",
		"  HEALTH/SDSF  GIT/GITLOG  SPRINT/BOARD  SWAP",
		"  FIND <text>  RESET  VIEW [id]  EDIT [id]  TOP  BOTTOM",
		"  RUN TASK|PLAN|WAVE|HANDOFF",
		"",
		"Line commands (type in S column next to task row):",
		"  S=Select(view)  E=Edit  D=Mark Done  I=Mark In Progress",
		"",
		"PF keys (all screens):",
		"  PF1=Help  PF3=Back/Exit  PF11=Swap session",
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

	cols := t3270ScreenCols(devInfo)
	screen := go3270.Screen{
		{Row: 0, Col: 0, Content: t3270ISPFRuleLine(cols, " HELP "), Color: t3270PanelRuleColor()},
		{Row: 1, Col: 2, Content: "EXARP-GO 3270 HELP TUTORIAL", Intense: true, Color: t3270ISPFTitleColor()},
		{Row: helpPFRow, Col: 2, Content: "PF3=Back to previous screen", Color: go3270.Turquoise},
	}

	maxLines := helpContentMax - 3
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

	parts := strings.Fields(cmdUpper)
	if len(parts) == 0 {
		return currentTx, state, nil
	}

	command := parts[0]
	args := parts[1:]

	if command == "AGENT" {
		state.command = ""
		rest := strings.TrimSpace(strings.Join(args, " "))
		if rest == "" {
			return state.handleCommand("RUN", currentTx)
		}
		return state.handleCommand("RUN "+rest, currentTx)
	}

	if fn, ok := t3270VerbDispatch[command]; ok {
		return fn(state, currentTx, args)
	}

	state.command = ""
	return currentTx, state, nil
}

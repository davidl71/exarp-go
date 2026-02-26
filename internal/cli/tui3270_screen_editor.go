// tui3270_screen_editor.go — Task editor screen transaction for the 3270 TUI.
package cli

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/racingmars/go3270"
)

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

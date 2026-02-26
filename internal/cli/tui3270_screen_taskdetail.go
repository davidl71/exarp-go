// tui3270_screen_taskdetail.go — Task detail screen transaction for the 3270 TUI.
package cli

import (
	"net"
	"strings"

	"github.com/racingmars/go3270"
)

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

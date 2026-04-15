// tui3270_screen_sprintboard.go — Sprint board (kanban-style) screen transaction for the 3270 TUI.
package cli

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/davidl71/exarp-go/internal/database"
	"github.com/racingmars/go3270"
)

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

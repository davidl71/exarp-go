// tui3270_screen_health.go — System health (SDSF-style) screen transaction for the 3270 TUI.
package cli

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/racingmars/go3270"
)

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

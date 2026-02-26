// tui3270_screen_gitdashboard.go — Git dashboard screen transaction for the 3270 TUI.
package cli

import (
	"context"
	"net"
	"strings"

	"github.com/racingmars/go3270"
)

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

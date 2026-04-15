// tui3270_screen_handoff.go — Session handoff screen transaction for the 3270 TUI.
package cli

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/racingmars/go3270"
)

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

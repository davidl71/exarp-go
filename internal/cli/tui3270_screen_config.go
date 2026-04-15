// tui3270_screen_config.go — Configuration screen transaction for the 3270 TUI.
package cli

import (
	"net"

	"github.com/davidl71/exarp-go/internal/config"
	"github.com/racingmars/go3270"
)

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

// tui3270_screen_scorecard.go — Scorecard screen transaction for the 3270 TUI.
package cli

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/davidl71/exarp-go/internal/tools"
	"github.com/racingmars/go3270"
)

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

	scErrPFRow := t3270PFRow(devInfo)

	var combined strings.Builder

	var recommendations []string

	// Use full mode when returning from "Run #" so updated coverage/lint is shown
	useFullMode := state.scorecardFullModeNext
	state.scorecardFullModeNext = false

	showLoadingOverlay(conn, devInfo, "Loading scorecard...")

	if tools.IsGoProject() {
		scorecardOpts := &tools.ScorecardOptions{FastMode: !useFullMode}

		scorecard, err := tools.GenerateGoScorecard(ctx, state.projectRoot, scorecardOpts)
		if err != nil {
			errScreen := go3270.Screen{
				{Row: 2, Col: 2, Content: "SCORECARD", Intense: true, Color: go3270.Blue},
				{Row: 4, Col: 2, Content: "Error: " + err.Error(), Color: go3270.Red},
				{Row: scErrPFRow, Col: 2, Content: "PF3=Back", Color: go3270.Turquoise},
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
				{Row: scErrPFRow, Col: 2, Content: "PF3=Back", Color: go3270.Turquoise},
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
	scContentMax := t3270ContentMaxRow(devInfo)
	maxRows := scContentMax - 2 // Reserve title row and margin
	if maxRows < 10 {
		maxRows = 10
	}

	if len(lines) > maxRows {
		lines = lines[:maxRows]
	}

	scStatusRow := t3270StatusRow(devInfo)
	scPFRow := t3270PFRow(devInfo)

	screen := go3270.Screen{
		{Row: 1, Col: 2, Content: "PROJECT SCORECARD", Intense: true, Color: go3270.Blue},
		{Row: scPFRow, Col: 2, Content: "PF3=Back to menu", Color: go3270.Turquoise},
	}
	if len(state.scorecardRecs) > 0 {
		screen = append(screen,
			go3270.Field{Row: scStatusRow, Col: 2, Content: "Run # (1-" + strconv.Itoa(len(state.scorecardRecs)) + "):", Intense: true},
			go3270.Field{Row: scStatusRow, Col: 24, Write: true, Name: "run_rec", Content: ""},
		)
		screen[1] = go3270.Field{Row: scPFRow, Col: 2, Content: "PF3=Back  Enter=Run selected #"}
	}

	for i, line := range lines {
		if len(line) > 78 {
			line = line[:75] + "..."
		}

		screen = append(screen, go3270.Field{Row: 2 + i, Col: 2, Content: line, Color: scorecardLineColor(line)})
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
			state.pushSession("Scorecard", state.scorecardTransaction)
			return s.tx, state, nil
		}
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

			resultMaxLines := scContentMax - 4
			if resultMaxLines < 10 {
				resultMaxLines = 10
			}
			if len(resultLines) > resultMaxLines {
				resultLines = resultLines[:resultMaxLines]
			}

			resultScreen := go3270.Screen{
				{Row: 1, Col: 2, Content: "IMPLEMENT RESULT (#" + runRec + ")", Intense: true, Color: go3270.Blue},
				{Row: 2, Col: 2, Content: rec, Color: go3270.Green},
				{Row: scPFRow, Col: 2, Content: "PF3=Back to scorecard", Color: go3270.Turquoise},
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

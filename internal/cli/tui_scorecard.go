// tui_scorecard.go — TUI scorecard tab and help view rendering.
package cli

import (
	"fmt"
	"strings"
)

func (m model) viewScorecard() string {
	// Use effective dimensions for consistent layout in iTerm2 and other terminals.
	availableWidth := m.effectiveWidth() - 2
	if availableWidth < 40 {
		availableWidth = 40
	}

	maxTextLines := m.effectiveHeight() - 14
	if maxTextLines < 4 {
		maxTextLines = 4
	}

	var b strings.Builder

	// Header
	title := "SCORECARD"
	if m.projectName != "" {
		title = fmt.Sprintf("%s - %s", strings.ToUpper(m.projectName), title)
	}

	b.WriteString(headerStyle.Render(title))
	b.WriteString(" ")
	b.WriteString(headerLabelStyle.Render("p=back"))
	b.WriteString(" ")
	b.WriteString(headerLabelStyle.Render("r=refresh"))
	b.WriteString("\n")
	b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
	b.WriteString("\n")

	if m.scorecardLoading {
		b.WriteString("\n  Loading scorecard...\n\n")
		b.WriteString(statusBarStyle.Render("Commands: p back  q quit"))

		return b.String()
	}

	if m.scorecardErr != nil {
		b.WriteString(fmt.Sprintf("\n  Error: %v\n\n", m.scorecardErr))
		b.WriteString(statusBarStyle.Render("Commands: p back  r refresh  q quit"))

		return b.String()
	}

	// Show scorecard text; cap lines so recommendations and status stay visible
	lines := strings.Split(m.scorecardText, "\n")
	textLinesShown := 0

	for _, line := range lines {
		if textLinesShown >= maxTextLines {
			b.WriteString(helpStyle.Render("  ... (run report/scorecard for full output)"))
			b.WriteString("\n")

			break
		}

		if len(line) > availableWidth {
			for len(line) > 0 && textLinesShown < maxTextLines {
				end := availableWidth
				if end > len(line) {
					end = len(line)
				}

				b.WriteString(normalStyle.Render(line[:end]))
				b.WriteString("\n")

				line = line[end:]
				textLinesShown++
			}

			if textLinesShown >= maxTextLines {
				b.WriteString(helpStyle.Render("  ... (run report/scorecard for full output)"))
				b.WriteString("\n")

				break
			}
		} else {
			b.WriteString(normalStyle.Render(line))
			b.WriteString("\n")

			textLinesShown++
		}
	}

	// Interactive recommendations: select and run
	if len(m.scorecardRecs) > 0 {
		b.WriteString("\n")
		b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
		b.WriteString("\n")
		b.WriteString(normalStyle.Render("Run recommendation (↑↓ select, e or Enter run):"))
		b.WriteString("\n")

		for i, rec := range m.scorecardRecs {
			prefix := "   "

			if i == m.scorecardRecCursor {
				_, _, ok := recommendationToCommand(rec)
				if ok {
					prefix = " ▶ " // runnable
				} else {
					prefix = " → " // selected but not runnable
				}
			}

			line := rec
			if len(line) > availableWidth-len(prefix)-2 {
				line = line[:availableWidth-len(prefix)-5] + "..."
			}

			fullLine := prefix + line
			if i == m.scorecardRecCursor {
				b.WriteString(highlightRow(fullLine, availableWidth, true))
			} else {
				b.WriteString(normalStyle.Render(fullLine))
			}

			b.WriteString("\n")
		}
	}

	if m.scorecardRunOutput != "" {
		b.WriteString("\n")
		b.WriteString(helpStyle.Render("Last run:"))
		b.WriteString("\n")

		for _, line := range strings.Split(m.scorecardRunOutput, "\n") {
			if len(line) > availableWidth {
				line = line[:availableWidth-3] + "..."
			}

			b.WriteString(normalStyle.Render("  " + line))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
	b.WriteString("\n")
	b.WriteString(statusBarStyle.Render("Commands: p back  r refresh  e implement selected  q quit"))

	return b.String()
}

func (m model) viewHelp() string {
	availableWidth := m.effectiveWidth() - 2
	if availableWidth < 40 {
		availableWidth = 40
	}

	var b strings.Builder

	b.WriteString(headerStyle.Render("HELP - Key bindings"))
	b.WriteString("\n")
	b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
	b.WriteString("\n\n")
	b.WriteString(normalStyle.Render("Navigation:"))
	b.WriteString("\n  ")
	b.WriteString(helpStyle.Render("↑↓ / j / k"))
	b.WriteString("       Move cursor (tasks, config, detail)\n  ")
	b.WriteString(helpStyle.Render("PgUp/PgDn / C-u/d"))
	b.WriteString(" Page up/down\n  ")
	b.WriteString(helpStyle.Render("g / G"))
	b.WriteString("            Jump to first/last item\n  ")
	b.WriteString(helpStyle.Render("Enter / Space"))
	b.WriteString("    Toggle selection (tasks) or open (config)\n\n")
	b.WriteString(normalStyle.Render("Search, filter & sort (tasks):"))
	b.WriteString("\n  ")
	b.WriteString(helpStyle.Render("/"))
	b.WriteString("  Search/filter (type query, Enter apply, Esc cancel)\n  ")
	b.WriteString(helpStyle.Render("n / N"))
	b.WriteString("  Next/previous match\n  ")
	b.WriteString(helpStyle.Render("f"))
	b.WriteString("  Cycle status filter (Todo → WIP → Review → Done → All)\n  ")
	b.WriteString(helpStyle.Render("o"))
	b.WriteString("  Cycle sort order (id → status → priority → updated → hierarchy)\n  ")
	b.WriteString(helpStyle.Render("O"))
	b.WriteString("  Toggle sort direction (asc/desc)\n\n")
	b.WriteString(normalStyle.Render("Views:"))
	b.WriteString("\n  ")
	b.WriteString(helpStyle.Render("v"))
	b.WriteString("  Toggle compact/spacious list density\n  ")
	b.WriteString(helpStyle.Render("c"))
	b.WriteString("  Switch to Config\n  ")
	b.WriteString(helpStyle.Render("p"))
	b.WriteString("  Switch to Scorecard (project health)\n  ")
	b.WriteString(helpStyle.Render("H"))
	b.WriteString("  Switch to Session handoffs (i interactive agent  e run & close  x close  a approve  d delete)\n  ")
	b.WriteString(helpStyle.Render("w"))
	b.WriteString("  Switch to Waves (Enter expand wave)\n  ")
	b.WriteString(helpStyle.Render("A"))
	b.WriteString("  Task analysis (tasks/waves: run parallelization, show result)\n  ")
	b.WriteString(helpStyle.Render("b"))
	b.WriteString("  Switch to Background jobs (child agent launches)\n\n")
	b.WriteString(normalStyle.Render("Child agent (run Cursor agent in project root):"))
	b.WriteString("\n  ")
	b.WriteString(helpStyle.Render("E"))
	b.WriteString("  Execute current context in child agent (task, handoff, or wave)\n  ")
	b.WriteString(helpStyle.Render("i"))
	b.WriteString("  In handoffs: start interactive agent (don't close handoff)\n  ")
	b.WriteString(helpStyle.Render("L"))
	b.WriteString("  Launch plan in child agent (from tasks view)\n\n")
	b.WriteString(normalStyle.Render("Task status (in tasks view):"))
	b.WriteString("\n  ")
	b.WriteString(helpStyle.Render("d"))
	b.WriteString("  Set task to Done (current cursor)\n  ")
	b.WriteString(helpStyle.Render("i"))
	b.WriteString("  Set task to In Progress (current cursor)\n  ")
	b.WriteString(helpStyle.Render("t"))
	b.WriteString("  Set task to Todo (current cursor)\n  ")
	b.WriteString(helpStyle.Render("r"))
	b.WriteString("  Set task to Review (current cursor)\n  ")
	b.WriteString(helpStyle.Render("D"))
	b.WriteString("  Bulk update selected tasks (prompts for status)\n\n")
	b.WriteString(normalStyle.Render("Task creation:"))
	b.WriteString("\n  ")
	b.WriteString(helpStyle.Render("+"))
	b.WriteString("  Create new task inline (type name, Enter create, Esc cancel)\n\n")
	b.WriteString(normalStyle.Render("Actions:"))
	b.WriteString("\n  ")
	b.WriteString(helpStyle.Render("R"))
	b.WriteString("  Refresh (tasks, config, or scorecard); In waves: run exarp tools then refresh\n  ")
	b.WriteString(helpStyle.Render("y"))
	b.WriteString("  In task analysis: write waves plan (parallel-execution-subagents.plan.md)\n  ")
	b.WriteString(helpStyle.Render("e"))
	b.WriteString("  Implement selected recommendation (scorecard)\n  ")
	b.WriteString(helpStyle.Render("a"))
	b.WriteString("  Toggle auto-refresh (tasks only)\n  ")
	b.WriteString(helpStyle.Render("s"))
	b.WriteString("  Show task details (tasks); Save/Update config (config): u or s → .exarp/config.pb\n\n")
	b.WriteString(normalStyle.Render("Other:"))
	b.WriteString("\n  ")
	b.WriteString(helpStyle.Render("q"))
	b.WriteString("  Quit\n  ")
	b.WriteString(helpStyle.Render("? / h"))
	b.WriteString("  This help\n\n")
	b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
	b.WriteString("\n")
	b.WriteString(statusBarStyle.Render("Press ? or h or Esc to close help"))

	return b.String()
}

// isOldSequentialID checks if a task ID uses the old sequential format (T-1, T-2, etc.)
// vs the new epoch format (T-1768158627000)
// Old format: T- followed by a small number (< 10000, typically 1-999)
// New format: T- followed by epoch milliseconds (13 digits, typically 1.6+ trillion).
func isOldSequentialID(taskID string) bool {
	if !strings.HasPrefix(taskID, "T-") {
		return false
	}

	// Extract the number part
	numStr := strings.TrimPrefix(taskID, "T-")

	// Parse as integer
	var num int64
	if _, err := fmt.Sscanf(numStr, "%d", &num); err != nil {
		return false
	}

	// Old sequential IDs are typically small numbers (< 10000)
	// Epoch milliseconds are 13 digits (1.6+ trillion)
	// Use 1000000 (1 million) as the threshold to be safe
	return num < 1000000
}

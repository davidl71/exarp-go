// tui_handoffs.go — TUI handoffs tab view and detail rendering.
package cli

import (
	"encoding/json"
	"fmt"
	"strings"
)

func (m model) viewHandoffs() string {
	availableWidth := m.effectiveWidth() - 2
	if availableWidth < 40 {
		availableWidth = 40
	}

	maxContentLines := m.effectiveHeight() - 10
	if maxContentLines < 6 {
		maxContentLines = 6
	}

	wrapWidth := availableWidth - 2
	if wrapWidth < 38 {
		wrapWidth = 38
	}

	var b strings.Builder

	title := "SESSION HANDOFFS"
	if m.projectName != "" {
		title = fmt.Sprintf("%s - %s", strings.ToUpper(m.projectName), title)
	}

	b.WriteString(headerStyle.Render(title))
	b.WriteString(" ")
	b.WriteString(headerLabelStyle.Render("H=back"))
	b.WriteString(" ")
	b.WriteString(headerLabelStyle.Render("r=refresh"))
	b.WriteString("\n")
	b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
	b.WriteString("\n")

	if m.handoffLoading {
		b.WriteString("\n  ")
		b.WriteString(helpStyle.Render("Loading handoffs..."))
		b.WriteString("\n\n")
		b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
		b.WriteString("\n")
		b.WriteString(statusBarStyle.Render("Commands: H back  q quit"))

		return b.String()
	}

	if m.handoffErr != nil {
		b.WriteString("\n  ")
		b.WriteString(normalStyle.Render(fmt.Sprintf("Error: %v", m.handoffErr)))
		b.WriteString("\n\n")
		b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
		b.WriteString("\n")
		b.WriteString(statusBarStyle.Render("Commands: H back  r refresh  q quit"))

		return b.String()
	}

	if !m.handoffLoading && m.handoffErr == nil && len(m.handoffEntries) == 0 {
		b.WriteString("\n  ")
		b.WriteString(helpStyle.Render("No handoff notes."))
		b.WriteString("\n\n")
		b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
		b.WriteString("\n")
		b.WriteString(statusBarStyle.Render("Commands: H back  r refresh  q quit"))

		return b.String()
	}

	// Detail view: show full handoff when one is selected
	if m.handoffDetailIndex >= 0 && m.handoffDetailIndex < len(m.handoffEntries) {
		return m.viewHandoffDetail(m.handoffEntries[m.handoffDetailIndex], availableWidth, wrapWidth)
	}

	// List view: show handoffs as list with cursor and selection
	if len(m.handoffEntries) > 0 {
		// Header with count and selection
		b.WriteString(headerLabelStyle.Render("Handoffs:"))
		b.WriteString(headerValueStyle.Render(fmt.Sprintf(" %d", len(m.handoffEntries))))

		if len(m.handoffSelected) > 0 {
			b.WriteString(" ")
			b.WriteString(headerLabelStyle.Render("Selected:"))
			b.WriteString(headerValueStyle.Render(fmt.Sprintf(" %d", len(m.handoffSelected))))
		}

		b.WriteString("\n")
		b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
		b.WriteString("\n")
		// Column header
		colCursor := 3
		colNum := 5
		colHost := 18
		colTime := 22

		colSummary := availableWidth - colCursor - colNum - colHost - colTime - 4
		if colSummary < 15 {
			colSummary = 15
		}

		b.WriteString(helpStyle.Render(fmt.Sprintf("%-*s %-*s %-*s %-*s %s", colCursor, "", colNum, "#", colHost, "HOST", colTime, "TIME", "SUMMARY")))
		b.WriteString("\n")
		b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
		b.WriteString("\n")

		for i, h := range m.handoffEntries {
			cursor := "   "
			if m.handoffCursor == i {
				cursor = " → "
				if _, ok := m.handoffSelected[i]; ok {
					cursor = " ✓ "
				}
			} else if _, ok := m.handoffSelected[i]; ok {
				cursor = " ✓ "
			}

			host, _ := h["host"].(string)
			if host == "" {
				host = "—"
			}

			if len(host) > colHost {
				host = host[:colHost-3] + "..."
			}

			ts, _ := h["timestamp"].(string)
			if ts != "" && len(ts) > colTime {
				ts = ts[:colTime-3] + "..."
			}

			if ts == "" {
				ts = "—"
			}

			sum, _ := h["summary"].(string)
			if len(sum) > colSummary {
				sum = sum[:colSummary-3] + "..."
			}

			if sum == "" {
				sum = "(no summary)"
			}

			line := fmt.Sprintf("%-*s %-*d %-*s %-*s %s", colCursor, cursor, colNum, i+1, colHost, host, colTime, ts, sum)
			if m.handoffCursor == i {
				line = highlightRow(line, availableWidth, true)
			} else {
				line = normalStyle.Render(line)
			}

			b.WriteString(line)
			b.WriteString("\n")
		}

		if m.handoffActionMsg != "" {
			b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
			b.WriteString("\n  ")
			b.WriteString(helpStyle.Render(m.handoffActionMsg))
			b.WriteString("\n")
		}

		if m.childAgentMsg != "" {
			b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
			b.WriteString("\n  ")
			b.WriteString(helpStyle.Render(m.childAgentMsg))
			b.WriteString("\n")
		}

		b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
		b.WriteString("\n")
		b.WriteString(statusBarStyle.Render("Enter detail  Space select  i interactive agent  e run agent & close  x close  a approve  d delete  H back  r refresh  q quit"))

		return b.String()
	}

	// Fallback: try to parse as JSON and render (no list state)
	var payload struct {
		Handoffs []map[string]interface{} `json:"handoffs"`
		Count    int                      `json:"count"`
		Total    int                      `json:"total"`
	}

	if err := json.Unmarshal([]byte(m.handoffText), &payload); err == nil && len(payload.Handoffs) > 0 {
		linesUsed := 0
		for i, h := range payload.Handoffs {
			if linesUsed >= maxContentLines {
				b.WriteString("  ")
				b.WriteString(helpStyle.Render("... (run 'exarp-go session handoffs' for full list)"))
				b.WriteString("\n")

				break
			}
			// Header: Handoff N · host · timestamp, then separator line
			host, _ := h["host"].(string)
			if host == "" {
				host = "unknown"
			}

			ts, _ := h["timestamp"].(string)
			if ts != "" && len(ts) > 25 {
				ts = ts[:25]
			}

			b.WriteString("  ")
			b.WriteString(headerLabelStyle.Render(fmt.Sprintf("Handoff %d", i+1)))
			b.WriteString(headerValueStyle.Render(" · " + host))

			if ts != "" {
				b.WriteString(helpStyle.Render(" · " + ts))
			}

			b.WriteString("\n  ")
			b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
			b.WriteString("\n")

			linesUsed += 2

			// Summary
			if sum, ok := h["summary"].(string); ok && sum != "" {
				b.WriteString("  ")
				b.WriteString(headerValueStyle.Render("Summary:"))
				b.WriteString("\n  ")

				for _, wl := range strings.Split(wordWrap(sum, wrapWidth), "\n") {
					if linesUsed >= maxContentLines {
						b.WriteString(helpStyle.Render("  ..."))
						b.WriteString("\n")

						linesUsed++

						break
					}

					b.WriteString(normalStyle.Render("  "+wl) + "\n")

					linesUsed++
				}
			}

			// Blockers
			if blockers, ok := h["blockers"].([]interface{}); ok && len(blockers) > 0 {
				b.WriteString("  ")
				b.WriteString(headerValueStyle.Render("Blockers:"))
				b.WriteString("\n")

				linesUsed++
				for _, bi := range blockers {
					if linesUsed >= maxContentLines {
						break
					}

					bl := ""

					switch v := bi.(type) {
					case string:
						bl = v
					default:
						bl = fmt.Sprintf("%v", v)
					}

					for _, wl := range strings.Split(wordWrap(bl, wrapWidth), "\n") {
						b.WriteString("  ")
						b.WriteString(normalStyle.Render("  • "+wl) + "\n")

						linesUsed++
					}
				}
			}

			// Next steps
			if steps, ok := h["next_steps"].([]interface{}); ok && len(steps) > 0 {
				b.WriteString("  ")
				b.WriteString(headerValueStyle.Render("Next steps:"))
				b.WriteString("\n")

				linesUsed++
				for _, si := range steps {
					if linesUsed >= maxContentLines {
						break
					}

					st := ""

					switch v := si.(type) {
					case string:
						st = v
					default:
						st = fmt.Sprintf("%v", v)
					}

					for _, wl := range strings.Split(wordWrap(st, wrapWidth), "\n") {
						b.WriteString("  ")
						b.WriteString(normalStyle.Render("  • "+wl) + "\n")

						linesUsed++
					}
				}
			}

			b.WriteString("\n")

			linesUsed++
		}

		b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
		b.WriteString("\n")
		b.WriteString(statusBarStyle.Render("Commands: e run agent & close  E run agent  H back  r refresh  q quit"))

		return b.String()
	}

	// Fallback: plain text (e.g. non-JSON or parse failure) with word wrap
	lines := strings.Split(m.handoffText, "\n")
	textLinesShown := 0

	for _, line := range lines {
		if textLinesShown >= maxContentLines {
			b.WriteString("  ")
			b.WriteString(helpStyle.Render("... (run 'exarp-go session handoffs' for full output)"))
			b.WriteString("\n")

			break
		}

		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			b.WriteString("\n")

			textLinesShown++

			continue
		}

		for _, wl := range strings.Split(wordWrap(trimmed, wrapWidth), "\n") {
			if textLinesShown >= maxContentLines {
				break
			}

			b.WriteString("  ")
			b.WriteString(normalStyle.Render(wl))
			b.WriteString("\n")

			textLinesShown++
		}
	}

	b.WriteString("\n")
	b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
	b.WriteString("\n")
	b.WriteString(statusBarStyle.Render("Commands: H back  r refresh  q quit"))

	return b.String()
}

// viewHandoffDetail renders a single handoff in detail (full summary, blockers, next steps).
func (m model) viewHandoffDetail(h map[string]interface{}, availableWidth, wrapWidth int) string {
	var b strings.Builder

	title := "SESSION HANDOFFS"
	if m.projectName != "" {
		title = fmt.Sprintf("%s - %s", strings.ToUpper(m.projectName), title)
	}

	b.WriteString(headerStyle.Render(title))
	b.WriteString(" ")
	b.WriteString(headerLabelStyle.Render("Esc back"))
	b.WriteString("\n")
	b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
	b.WriteString("\n")

	host, _ := h["host"].(string)
	if host == "" {
		host = "unknown"
	}

	ts, _ := h["timestamp"].(string)
	id, _ := h["id"].(string)

	b.WriteString("  ")
	b.WriteString(headerLabelStyle.Render("Host:"))
	b.WriteString(headerValueStyle.Render(" " + host))
	b.WriteString("  ")
	b.WriteString(headerLabelStyle.Render("Time:"))
	b.WriteString(headerValueStyle.Render(" " + ts))

	if id != "" {
		b.WriteString("  ")
		b.WriteString(helpStyle.Render("ID: " + id))
	}

	b.WriteString("\n  ")
	b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
	b.WriteString("\n")

	if sum, ok := h["summary"].(string); ok && sum != "" {
		b.WriteString("  ")
		b.WriteString(headerValueStyle.Render("Summary:"))
		b.WriteString("\n  ")

		for _, wl := range strings.Split(wordWrap(sum, wrapWidth), "\n") {
			b.WriteString(normalStyle.Render("  "+wl) + "\n")
		}

		b.WriteString("\n")
	}

	if blockers, ok := h["blockers"].([]interface{}); ok && len(blockers) > 0 {
		b.WriteString("  ")
		b.WriteString(headerValueStyle.Render("Blockers:"))
		b.WriteString("\n")

		for _, bi := range blockers {
			bl := ""

			switch v := bi.(type) {
			case string:
				bl = v
			default:
				bl = fmt.Sprintf("%v", v)
			}

			for _, wl := range strings.Split(wordWrap(bl, wrapWidth), "\n") {
				b.WriteString("  ")
				b.WriteString(normalStyle.Render("  • "+wl) + "\n")
			}
		}

		b.WriteString("\n")
	}

	if steps, ok := h["next_steps"].([]interface{}); ok && len(steps) > 0 {
		b.WriteString("  ")
		b.WriteString(headerValueStyle.Render("Next steps:"))
		b.WriteString("\n")

		for _, si := range steps {
			st := ""

			switch v := si.(type) {
			case string:
				st = v
			default:
				st = fmt.Sprintf("%v", v)
			}

			for _, wl := range strings.Split(wordWrap(st, wrapWidth), "\n") {
				b.WriteString("  ")
				b.WriteString(normalStyle.Render("  • "+wl) + "\n")
			}
		}

		b.WriteString("\n")
	}

	if m.childAgentMsg != "" {
		b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
		b.WriteString("\n  ")
		b.WriteString(helpStyle.Render(m.childAgentMsg))
		b.WriteString("\n")
	}

	b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
	b.WriteString("\n")
	b.WriteString(statusBarStyle.Render("i interactive  e run & close  x close  a approve  d delete  Esc back  q quit"))

	return b.String()
}

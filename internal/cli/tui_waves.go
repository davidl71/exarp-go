// tui_waves.go — TUI waves tab view rendering.
package cli

import (
	"fmt"
	"strings"

	"github.com/davidl71/exarp-go/internal/database"
)

func (m model) viewWaves() string {
	availableWidth := m.effectiveWidth() - 2
	if availableWidth < 40 {
		availableWidth = 40
	}

	wrapWidth := availableWidth - 2
	if wrapWidth < 30 {
		wrapWidth = 30
	}

	var b strings.Builder

	title := "WAVES (dependency order)"
	if m.projectName != "" {
		title = fmt.Sprintf("%s - %s", strings.ToUpper(m.projectName), title)
	}

	b.WriteString(headerStyle.Render(title))
	b.WriteString(" ")
	b.WriteString(headerLabelStyle.Render("H/w=back"))
	b.WriteString(" ")
	b.WriteString(headerLabelStyle.Render("Enter=expand"))
	b.WriteString(" ")
	b.WriteString(headerLabelStyle.Render("r=refresh"))
	b.WriteString(" ")
	b.WriteString(headerLabelStyle.Render("R=refresh tools"))
	b.WriteString(" ")
	b.WriteString(headerLabelStyle.Render("U=update from plan"))
	b.WriteString(" ")
	b.WriteString(headerLabelStyle.Render("A=analysis"))
	b.WriteString("\n")
	b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
	b.WriteString("\n")

	if len(m.waves) == 0 {
		b.WriteString("\n  ")
		b.WriteString(helpStyle.Render("No backlog waves (Todo/In Progress with dependencies). Refresh tasks (r) or run tools then refresh (R)."))
		b.WriteString("\n\n")
		b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
		b.WriteString("\n")
		b.WriteString(statusBarStyle.Render("Commands: H or w back  r refresh  R refresh tools  q quit"))

		return b.String()
	}

	levels := sortedWaveLevels(m.waves)
	taskByID := make(map[string]*database.Todo2Task)

	for _, t := range m.tasks {
		if t != nil {
			taskByID[t.ID] = t
		}
	}

	if m.waveDetailLevel >= 0 {
		// Expanded: show tasks for the selected wave only
		ids := m.waves[m.waveDetailLevel]
		levels := sortedWaveLevels(m.waves)

		maxWaveIdx := len(levels) - 1
		if maxWaveIdx < 0 {
			maxWaveIdx = 0
		}

		b.WriteString("  ")
		b.WriteString(headerLabelStyle.Render(fmt.Sprintf("Wave %d", m.waveDetailLevel)))
		b.WriteString(headerValueStyle.Render(fmt.Sprintf(" (%d tasks) — Esc/Enter to collapse", len(ids))))
		b.WriteString("\n  ")
		b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
		b.WriteString("\n")

		for i, id := range ids {
			content := id
			if t := taskByID[id]; t != nil && t.Content != "" {
				content = truncatePad(t.Content, availableWidth-6)
			}

			line := "  • " + helpStyle.Render(id) + "  " + normalStyle.Render(content)
			if i == m.waveTaskCursor {
				line = highlightRow(line, availableWidth, true)
			}

			b.WriteString(line)
			b.WriteString("\n")
		}

		if m.waveMoveTaskID != "" {
			b.WriteString("\n  ")
			b.WriteString(helpStyle.Render(fmt.Sprintf("Move %s to wave (0-%d): press 0-%d  Esc cancel", m.waveMoveTaskID, maxWaveIdx, maxWaveIdx)))
			b.WriteString("\n")
		}

		if m.waveMoveMsg != "" {
			b.WriteString("  ")
			b.WriteString(statusBarStyle.Render(m.waveMoveMsg))
			b.WriteString("\n")
		}

		if m.waveUpdateMsg != "" {
			b.WriteString("  ")
			b.WriteString(statusBarStyle.Render(m.waveUpdateMsg))
			b.WriteString("\n")
		}
		if m.queueEnqueueMsg != "" {
			b.WriteString("  ")
			b.WriteString(statusBarStyle.Render(m.queueEnqueueMsg))
			b.WriteString("\n")
		}

		b.WriteString("\n")
		b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
		b.WriteString("\n")

		statusLine := "↑↓ move  m move task  U update from plan  Esc/Enter collapse  E run wave  R refresh  H/w back  q quit"
		if m.queueEnabled {
			statusLine += "  Q enqueue"
		}
		if m.waveMoveTaskID != "" {
			statusLine = "0-9 pick wave  Esc cancel  " + statusLine
		}

		b.WriteString(statusBarStyle.Render(statusLine))

		return b.String()
	}

	// Collapsed: wave list only
	b.WriteString(helpStyle.Render("  #   WAVE        TASKS"))
	b.WriteString("\n")
	b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
	b.WriteString("\n")

	for i, level := range levels {
		ids := m.waves[level]

		cursor := "   "
		if m.waveCursor == i {
			cursor = " → "
		}

		line := fmt.Sprintf("%s %-3d  Wave %-3d   %d tasks", cursor, i+1, level, len(ids))
		if m.waveCursor == i {
			line = highlightRow(line, availableWidth, true)
		} else {
			line = normalStyle.Render(line)
		}

		b.WriteString("  ")
		b.WriteString(line)
		b.WriteString("\n")
	}

	if m.waveUpdateMsg != "" {
		b.WriteString("  ")
		b.WriteString(statusBarStyle.Render(m.waveUpdateMsg))
		b.WriteString("\n")
	}
	if m.queueEnqueueMsg != "" {
		b.WriteString("  ")
		b.WriteString(statusBarStyle.Render(m.queueEnqueueMsg))
		b.WriteString("\n")
	}

	b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
	b.WriteString("\n")
	statusLine := "Enter expand  E run wave  R refresh tools  U update from plan  A analysis  H/w back  r refresh  q quit"
	if m.queueEnabled {
		statusLine = "[Queue] " + statusLine + "  Q enqueue wave"
	}
	b.WriteString(statusBarStyle.Render(statusLine))

	return b.String()
}

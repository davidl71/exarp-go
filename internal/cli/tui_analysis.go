// tui_analysis.go — TUI task analysis tab view rendering.
package cli

import (
	"fmt"
	"strings"
)

func (m model) viewTaskAnalysis() string {
	availableWidth := m.effectiveWidth() - 2
	if availableWidth < 40 {
		availableWidth = 40
	}

	maxTextLines := m.effectiveHeight() - 10
	if maxTextLines < 4 {
		maxTextLines = 4
	}

	var b strings.Builder

	title := "TASK ANALYSIS"
	if m.projectName != "" {
		title = fmt.Sprintf("%s - %s", strings.ToUpper(m.projectName), title)
	}

	b.WriteString(headerStyle.Render(title))
	b.WriteString(" ")
	b.WriteString(headerLabelStyle.Render("action=" + m.taskAnalysisAction))
	b.WriteString(" ")
	b.WriteString(headerLabelStyle.Render("p=back"))
	b.WriteString(" ")
	b.WriteString(headerLabelStyle.Render("r=rerun"))
	b.WriteString(" ")
	b.WriteString(headerLabelStyle.Render("y=write waves"))
	b.WriteString("\n")
	b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
	b.WriteString("\n")

	if m.taskAnalysisLoading {
		b.WriteString("\n  Running task_analysis (" + m.taskAnalysisAction + ")...\n\n")
		b.WriteString(statusBarStyle.Render("Commands: p back  q quit"))

		return b.String()
	}

	if m.taskAnalysisApproveLoading {
		b.WriteString("\n  Writing waves plan to .cursor/plans/parallel-execution-subagents.plan.md ...\n\n")
		b.WriteString(statusBarStyle.Render("Commands: p back  q quit"))

		return b.String()
	}

	if m.taskAnalysisErr != nil {
		b.WriteString(fmt.Sprintf("\n  Error: %v\n\n", m.taskAnalysisErr))
		b.WriteString(statusBarStyle.Render("Commands: p back  r rerun  y write waves  q quit"))

		return b.String()
	}

	lines := strings.Split(m.taskAnalysisText, "\n")
	shown := 0

	for _, line := range lines {
		if shown >= maxTextLines {
			b.WriteString(helpStyle.Render("  ... (run exarp-go -tool task_analysis for full output)"))
			b.WriteString("\n")

			break
		}

		if len(line) > availableWidth {
			for len(line) > 0 && shown < maxTextLines {
				end := availableWidth
				if end > len(line) {
					end = len(line)
				}

				b.WriteString(normalStyle.Render(line[:end]))
				b.WriteString("\n")

				line = line[end:]
				shown++
			}
		} else {
			b.WriteString(normalStyle.Render(line))
			b.WriteString("\n")

			shown++
		}
	}

	if m.taskAnalysisApproveMsg != "" {
		b.WriteString("\n")
		b.WriteString(helpStyle.Render("  " + m.taskAnalysisApproveMsg))
		b.WriteString("\n")
	}

	b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
	b.WriteString("\n")
	b.WriteString(statusBarStyle.Render("Commands: p back  r rerun  y write waves  q quit"))

	return b.String()
}

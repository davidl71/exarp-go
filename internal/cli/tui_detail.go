package cli

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m model) viewTaskDetail() string {
	task := m.taskDetailTask
	if task == nil {
		return ""
	}
	// Use effective dimensions so task detail ("s") behaves consistently in iTerm2 and other terminals.
	availableWidth := m.effectiveWidth() - 2
	if availableWidth < 40 {
		availableWidth = 40
	}

	maxContentLines := m.effectiveHeight() - 10
	if maxContentLines < 8 {
		maxContentLines = 8
	}

	var body strings.Builder
	if task.ID != "" {
		body.WriteString("Task: ")
		body.WriteString(task.ID)
		body.WriteString("\n\n")
	}

	if task.Status != "" {
		body.WriteString("Status: ")
		body.WriteString(task.Status)
		body.WriteString("\n\n")
	}

	if task.Priority != "" {
		body.WriteString("Priority: ")
		body.WriteString(task.Priority)
		body.WriteString("\n\n")
	}

	if task.Content != "" {
		body.WriteString("Content:\n")
		body.WriteString(wordWrap(task.Content, availableWidth-2))
		body.WriteString("\n\n")
	}

	if task.LongDescription != "" {
		body.WriteString("Description:\n")
		body.WriteString(wordWrap(task.LongDescription, availableWidth-2))
		body.WriteString("\n\n")
	}

	if len(task.Tags) > 0 {
		body.WriteString("Tags: ")
		body.WriteString(strings.Join(task.Tags, ", "))
		body.WriteString("\n\n")
	}

	if len(task.Dependencies) > 0 {
		body.WriteString("Dependencies: ")
		body.WriteString(strings.Join(task.Dependencies, ", "))
		body.WriteString("\n")
	}

	allLines := strings.Split(strings.TrimSuffix(body.String(), "\n"), "\n")

	var b strings.Builder

	// Header bar (full-width, same as task list)
	title := "TASK DETAIL"
	if m.projectName != "" {
		title = fmt.Sprintf("%s - %s", strings.ToUpper(m.projectName), title)
	}

	headerLine := headerStyle.Render(title) + " " + headerLabelStyle.Render("Esc/Enter/Space=close")
	if lipgloss.Width(headerLine) < availableWidth {
		padding := strings.Repeat(" ", availableWidth-lipgloss.Width(headerLine))
		headerLine += headerValueStyle.Render(padding)
	}

	b.WriteString(headerLine)
	b.WriteString("\n")
	b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
	b.WriteString("\n")

	for i, line := range allLines {
		if i >= maxContentLines {
			b.WriteString(helpStyle.Render("  ... (truncated)\n"))
			break
		}
		// Constrain line to available width so it never overflows (matches menu behavior)
		if lipgloss.Width(line) > availableWidth {
			runes := []rune(line)
			maxRunes := availableWidth - 3

			if maxRunes > 0 && len(runes) > maxRunes {
				line = string(runes[:maxRunes]) + "..."
			}
		}

		b.WriteString(normalStyle.Render(line))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
	b.WriteString("\n")

	// Status bar (full-width, same as task list)
	statusLine := statusBarStyle.Render("Press Esc, Enter, or Space to close  s=close")
	if lipgloss.Width(statusLine) < availableWidth {
		padding := strings.Repeat(" ", availableWidth-lipgloss.Width(statusLine))
		statusLine += statusBarStyle.Render(padding)
	}

	b.WriteString(statusLine)

	return b.String()
}

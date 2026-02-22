// tui_detail.go — TUI task detail view rendering with scroll support.
package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// buildDetailLines constructs the lines for the task detail body.
func (m model) buildDetailLines(availableWidth int) []string {
	task := m.taskDetailTask
	if task == nil {
		return nil
	}

	contentWidth := availableWidth - 4
	if contentWidth < 30 {
		contentWidth = 30
	}

	var body strings.Builder

	// Header section: ID + Status + Priority on compact lines
	body.WriteString(fmt.Sprintf("  Task:     %s\n", task.ID))
	body.WriteString(fmt.Sprintf("  Status:   %s\n", task.Status))
	if task.Priority != "" {
		body.WriteString(fmt.Sprintf("  Priority: %s\n", task.Priority))
	}
	if task.ParentID != "" {
		body.WriteString(fmt.Sprintf("  Parent:   %s\n", task.ParentID))
	}
	if task.AssignedTo != "" {
		body.WriteString(fmt.Sprintf("  Assigned: %s\n", task.AssignedTo))
	}

	// Timestamps
	if task.CreatedAt != "" || task.LastModified != "" || task.CompletedAt != "" {
		body.WriteString("\n")
		if task.CreatedAt != "" {
			body.WriteString(fmt.Sprintf("  Created:  %s\n", formatTimestamp(task.CreatedAt)))
		}
		if task.LastModified != "" {
			body.WriteString(fmt.Sprintf("  Updated:  %s\n", formatTimestamp(task.LastModified)))
		}
		if task.CompletedAt != "" {
			body.WriteString(fmt.Sprintf("  Done:     %s\n", formatTimestamp(task.CompletedAt)))
		}
	}

	// Content
	if task.Content != "" {
		body.WriteString("\n")
		body.WriteString("  ── Content ──────────────────────────\n")
		for _, line := range strings.Split(wordWrap(task.Content, contentWidth), "\n") {
			body.WriteString("  " + line + "\n")
		}
	}

	// Description
	if task.LongDescription != "" {
		body.WriteString("\n")
		body.WriteString("  ── Description ──────────────────────\n")
		for _, line := range strings.Split(wordWrap(task.LongDescription, contentWidth), "\n") {
			body.WriteString("  " + line + "\n")
		}
	}

	// Tags
	if len(task.Tags) > 0 {
		body.WriteString("\n")
		body.WriteString(fmt.Sprintf("  Tags: %s\n", strings.Join(task.Tags, ", ")))
	}

	// Dependencies
	if len(task.Dependencies) > 0 {
		body.WriteString("\n")
		body.WriteString("  ── Dependencies ─────────────────────\n")
		for _, dep := range task.Dependencies {
			body.WriteString(fmt.Sprintf("    → %s\n", dep))
		}
	}

	// Metadata highlights
	if task.Metadata != nil {
		interesting := []string{"preferred_backend", "wave_level", "estimated_hours"}
		var metaLines []string
		for _, key := range interesting {
			if v, ok := task.Metadata[key]; ok {
				metaLines = append(metaLines, fmt.Sprintf("    %s: %v", key, v))
			}
		}
		if len(metaLines) > 0 {
			body.WriteString("\n")
			body.WriteString("  ── Metadata ─────────────────────────\n")
			for _, ml := range metaLines {
				body.WriteString(ml + "\n")
			}
		}
	}

	raw := strings.TrimSuffix(body.String(), "\n")
	lines := strings.Split(raw, "\n")

	// Clamp each line to available width
	for i, line := range lines {
		if lipgloss.Width(line) > availableWidth {
			runes := []rune(line)
			maxRunes := availableWidth - 3
			if maxRunes > 0 && len(runes) > maxRunes {
				lines[i] = string(runes[:maxRunes]) + "..."
			}
		}
	}

	return lines
}

func (m model) viewTaskDetail() string {
	task := m.taskDetailTask
	if task == nil {
		return ""
	}

	availableWidth := m.effectiveWidth() - 2
	if availableWidth < 40 {
		availableWidth = 40
	}

	// Reserve lines for header(2) + bottom border(1) + status bar(1) + blank(1) = 5
	maxContentLines := m.effectiveHeight() - 5
	if maxContentLines < 8 {
		maxContentLines = 8
	}

	allLines := m.buildDetailLines(availableWidth)

	var b strings.Builder

	// Header bar
	title := "TASK DETAIL"
	if m.projectName != "" {
		title = fmt.Sprintf("%s - %s", strings.ToUpper(m.projectName), title)
	}

	headerLine := headerStyle.Render(title)
	if lipgloss.Width(headerLine) < availableWidth {
		padding := strings.Repeat(" ", availableWidth-lipgloss.Width(headerLine))
		headerLine += headerValueStyle.Render(padding)
	}
	b.WriteString(headerLine)
	b.WriteString("\n")
	b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
	b.WriteString("\n")

	// Scrollable content window
	scrollTop := m.taskDetailScrollTop
	if scrollTop > len(allLines)-maxContentLines {
		scrollTop = len(allLines) - maxContentLines
	}
	if scrollTop < 0 {
		scrollTop = 0
	}

	end := scrollTop + maxContentLines
	if end > len(allLines) {
		end = len(allLines)
	}

	for i := scrollTop; i < end; i++ {
		b.WriteString(normalStyle.Render(allLines[i]))
		b.WriteString("\n")
	}

	// Pad remaining lines so status bar stays at bottom
	rendered := end - scrollTop
	for rendered < maxContentLines {
		b.WriteString("\n")
		rendered++
	}

	b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
	b.WriteString("\n")

	// Status bar
	statusParts := []string{
		helpStyle.Render("↑↓") + " scroll",
		helpStyle.Render("Esc/Enter") + " close",
	}
	if len(allLines) > maxContentLines {
		scrollPct := 0
		if len(allLines)-maxContentLines > 0 {
			scrollPct = scrollTop * 100 / (len(allLines) - maxContentLines)
		}
		statusParts = append(statusParts, fmt.Sprintf("%d%%", scrollPct))
	}

	statusLine := statusBarStyle.Render(strings.Join(statusParts, "  "))
	if lipgloss.Width(statusLine) < availableWidth {
		padding := strings.Repeat(" ", availableWidth-lipgloss.Width(statusLine))
		statusLine += statusBarStyle.Render(padding)
	}
	b.WriteString(statusLine)

	return b.String()
}

// formatTimestamp formats an RFC3339 timestamp to a human-readable relative+absolute format.
func formatTimestamp(ts string) string {
	t, err := time.Parse(time.RFC3339Nano, ts)
	if err != nil {
		t, err = time.Parse(time.RFC3339, ts)
		if err != nil {
			return ts
		}
	}
	ago := time.Since(t)
	relative := formatDuration(ago)
	return fmt.Sprintf("%s (%s ago)", t.Format("2006-01-02 15:04"), relative)
}

// formatDuration returns a compact human-readable duration.
func formatDuration(d time.Duration) string {
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	default:
		days := int(d.Hours() / 24)
		if days == 1 {
			return "1 day"
		}
		return fmt.Sprintf("%d days", days)
	}
}

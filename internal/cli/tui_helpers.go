// tui_helpers.go — TUI layout/formatting utilities (truncate, wrap, indent, cursor markers, priority styling).
package cli

import (
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/davidl71/exarp-go/internal/models"
)

func truncatePad(s string, width int) string {
	if width <= 0 {
		return s
	}

	if len(s) > width {
		if width <= 3 {
			return s[:width]
		}

		return s[:width-3] + "..."
	}

	return s + strings.Repeat(" ", width-len(s))
}

func (m model) indentForTask(realIdx int) string {
	if m.hierarchyDepthByID == nil || realIdx < 0 || realIdx >= len(m.tasks) {
		return ""
	}

	task := m.tasks[realIdx]
	if task == nil {
		return ""
	}

	d := m.hierarchyDepthByID[task.ID]
	if d <= 0 {
		return ""
	}

	return strings.Repeat("  ", d)
}

func (m model) treeMarkerForTask(realIdx int) string {
	if realIdx < 0 || realIdx >= len(m.tasks) {
		return ""
	}

	task := m.tasks[realIdx]
	if task == nil || !m.taskHasChildren(task.ID) {
		return ""
	}

	if _, ok := m.collapsedTaskIDs[task.ID]; ok {
		return "▶ "
	}

	return "▼ "
}

func wordWrap(s string, width int) string {
	if width <= 0 {
		return s
	}

	var out strings.Builder

	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimRight(line, " \t")
		if line == "" {
			out.WriteString("\n")
			continue
		}

		for len(line) > width {
			cut := width
			if idx := strings.LastIndex(line[:min(len(line), width)], " "); idx > 0 {
				cut = idx
			}

			out.WriteString(line[:cut])
			out.WriteString("\n")

			line = strings.TrimLeft(line[cut:], " ")
		}

		out.WriteString(line)
		out.WriteString("\n")
	}

	return strings.TrimSuffix(out.String(), "\n")
}

func sortedWaveLevels(waves map[int][]string) []int {
	levels := make([]int, 0, len(waves))
	for k := range waves {
		levels = append(levels, k)
	}

	sort.Ints(levels)

	return levels
}

func (m model) effectiveWidth() int {
	if m.width >= minTermWidth {
		return m.width
	}

	return minTermWidth
}

func (m model) effectiveHeight() int {
	if m.height >= minTermHeight {
		return m.height
	}

	return minTermHeight
}

func highlightRow(line string, width int, selected bool) string {
	if !selected {
		return line
	}

	visible := lipgloss.Width(line)
	if visible < width {
		line += strings.Repeat(" ", width-visible)
	}

	return selectedStyle.Render(line)
}

// cursorMarkerNarrow returns ">" for active cursor, "✓" for selected+active, " " otherwise.
func cursorMarkerNarrow(isCursor, isSelected bool) string {
	if !isCursor {
		return " "
	}

	if isSelected {
		return "✓"
	}

	return ">"
}

// cursorMarkerWide returns " → " for active cursor, " ✓ " for selected+active, "   " otherwise.
func cursorMarkerWide(isCursor, isSelected bool) string {
	if !isCursor {
		return "   "
	}

	if isSelected {
		return " ✓ "
	}

	return " → "
}

// formatPriorityFull returns the priority string upper-cased and padded, or "---" if empty.
func formatPriorityFull(priority string, width int) string {
	s := strings.ToUpper(priority)
	if s == "" {
		s = "---"
	}

	return truncatePad(s, width)
}

// formatPriorityShort returns the first letter of priority upper-cased, or "-" if empty.
func formatPriorityShort(priority string, width int) string {
	s := "-"
	if priority != "" {
		s = strings.ToUpper(priority[:1])
	}

	return truncatePad(s, width)
}

// stylePriorityInLine applies color styling to the priority short indicator within a line.
func stylePriorityInLine(line, priority, priorityShort string) string {
	if priority == "" {
		return line
	}

	switch strings.ToLower(priority) {
	case models.PriorityHigh:
		return strings.Replace(line, priorityShort, highPriorityStyle.Render(priorityShort), 1)
	case models.PriorityMedium:
		return strings.Replace(line, priorityShort, mediumPriorityStyle.Render(priorityShort), 1)
	case models.PriorityLow:
		return strings.Replace(line, priorityShort, lowPriorityStyle.Render(priorityShort), 1)
	default:
		return line
	}
}

// formatStatus returns the status string padded, or "---" if empty.
func formatStatus(status string, width int) string {
	s := status
	if s == "" {
		s = "---"
	}

	return truncatePad(s, width)
}

// taskContent returns the task's content (or long description), truncated to maxWidth.
func taskContent(content, longDesc string, maxWidth int) string {
	c := content
	if c == "" {
		c = longDesc
	}

	if len(c) > maxWidth && maxWidth > 3 {
		c = c[:maxWidth-3] + "..."
	}

	if c == "" {
		c = "(no description)"
	}

	return c
}

// renderRow applies indent, tree marker, cursor highlight, and normal styling to a row.
func renderRow(line string, indent, marker string, isCursor bool, width int) string {
	line = indent + marker + line
	if isCursor {
		return highlightRow(line, width, true)
	}

	return normalStyle.Render(line)
}

// tui_helpers.go — TUI layout/formatting utilities (truncate, wrap, indent).
package cli

import (
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
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

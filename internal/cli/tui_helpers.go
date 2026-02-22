// tui_helpers.go — TUI layout/formatting utilities (truncate, wrap, indent, cursor markers, priority styling).
package cli

import (
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/davidl71/exarp-go/internal/models"
)

// bubbleStatusFilters is the ordered list for cycling with 'f' key (mirrors 3270 statusFilters).
var bubbleStatusFilters = []string{"Todo", "In Progress", "Review", "Done", ""}

// nextBubbleStatusFilter returns the next status in the filter cycle.
func nextBubbleStatusFilter(current string) string {
	for i, s := range bubbleStatusFilters {
		if s == current {
			return bubbleStatusFilters[(i+1)%len(bubbleStatusFilters)]
		}
	}
	return bubbleStatusFilters[0]
}

// clearTransientMessages clears all one-shot feedback messages so they
// disappear on the next keypress.
func (m *model) clearTransientMessages() {
	m.childAgentMsg = ""
	m.bulkStatusMsg = ""
	m.handoffActionMsg = ""
	m.waveMoveMsg = ""
	m.waveUpdateMsg = ""
	m.queueEnqueueMsg = ""
	m.configSaveMessage = ""
	m.scorecardRunOutput = ""
	m.taskAnalysisApproveMsg = ""
}

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

// styleStatus applies color to a status string (mirrors 3270 statusColor).
func styleStatus(status string) string {
	switch status {
	case models.StatusDone:
		return statusDoneStyle.Render(status)
	case models.StatusInProgress:
		return statusInProgressStyle.Render(status)
	case models.StatusTodo:
		return statusTodoStyle.Render(status)
	case models.StatusReview:
		return statusReviewStyle.Render(status)
	default:
		return status
	}
}

// styleStatusInLine replaces the first occurrence of statusStr with a colored version.
func styleStatusInLine(line, status, statusStr string) string {
	if status == "" {
		return line
	}
	colored := styleStatus(statusStr)
	return strings.Replace(line, statusStr, colored, 1)
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

// taskListPageSize returns the number of task rows that fit in the visible area.
// Reserves lines for: header(1) + separator(1) + bottom separator(1) + status bar(1) + extra(2) = 6.
func (m model) taskListPageSize() int {
	overhead := 6
	if m.searchMode {
		overhead++
	}
	if m.bulkStatusPrompt {
		overhead++
	}
	if m.bulkStatusMsg != "" {
		overhead += 2
	}
	if m.childAgentMsg != "" {
		overhead += 2
	}
	// Medium/wide layouts have a column header + separator = 2 extra lines
	if m.effectiveWidth()-2 >= 80 {
		overhead += 2
	}
	page := m.effectiveHeight() - overhead
	if m.spaciousMode {
		page = page / 2
	}
	if page < 5 {
		page = 5
	}
	return page
}

// ensureCursorVisible adjusts viewportOffset so the cursor is always visible.
func (m *model) ensureCursorVisible() {
	pageSize := m.taskListPageSize()
	if m.cursor < m.viewportOffset {
		m.viewportOffset = m.cursor
	}
	if m.cursor >= m.viewportOffset+pageSize {
		m.viewportOffset = m.cursor - pageSize + 1
	}
	if m.viewportOffset < 0 {
		m.viewportOffset = 0
	}
}

// statusBarEntry represents a key hint for the status bar.
type statusBarEntry struct {
	key  string
	desc string
}

// renderTaskStatusBar returns a width-responsive status bar for the tasks view.
// On narrow terminals it shows only essential commands; on wide terminals it shows all.
func (m model) renderTaskStatusBar(width int) string {
	densityHint := "compact"
	if m.spaciousMode {
		densityHint = "spacious"
	}
	essential := []statusBarEntry{
		{"↑↓", "nav"}, {"/", "search"}, {"f", "filter"}, {"v", densityHint},
		{"s", "detail"}, {"+", "new"}, {"d/i/t/r", "status"}, {"?", "help"}, {"q", "quit"},
	}
	medium := []statusBarEntry{
		{"Space", "select"}, {"o/O", "sort"}, {"n/N", "find"},
		{"Tab", "collapse"}, {"D", "bulk"}, {"a", "auto"},
	}
	wide := []statusBarEntry{
		{"c", "config"}, {"p", "scorecard"}, {"w", "waves"},
		{"H", "handoffs"}, {"A", "analysis"}, {"b", "jobs"},
		{"E", "agent"}, {"L", "plan"}, {"PgUp/Dn", "page"},
	}

	var entries []statusBarEntry
	entries = append(entries, essential...)
	if width >= 100 {
		entries = append(entries, medium...)
	}
	if width >= 160 {
		entries = append(entries, wide...)
	}

	var sb strings.Builder
	for i, e := range entries {
		if i > 0 {
			sb.WriteString("  ")
		}
		sb.WriteString(helpStyle.Render(e.key))
		sb.WriteString(" ")
		sb.WriteString(e.desc)
	}

	line := statusBarStyle.Render(sb.String())
	if lipgloss.Width(line) < width {
		padding := strings.Repeat(" ", width-lipgloss.Width(line))
		line += statusBarStyle.Render(padding)
	}

	return line
}

// renderRow applies indent, tree marker, cursor highlight, and normal styling to a row.
func renderRow(line string, indent, marker string, isCursor bool, width int) string {
	line = indent + marker + line
	if isCursor {
		return highlightRow(line, width, true)
	}

	return normalStyle.Render(line)
}

// spaciousDescLine returns a truncated first line of the task description for spacious mode.
func spaciousDescLine(desc string, maxWidth int) string {
	if desc == "" {
		return ""
	}
	first, _, _ := strings.Cut(desc, "\n")
	first = strings.TrimSpace(first)
	if first == "" {
		return ""
	}
	if len(first) > maxWidth {
		first = first[:maxWidth-3] + "..."
	}
	return first
}

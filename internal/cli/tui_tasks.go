// tui_tasks.go — TUI tasks tab: narrow/medium/wide list rendering.
package cli

import (
	"fmt"
	"strings"

	"github.com/davidl71/exarp-go/internal/models"
	humanize "github.com/dustin/go-humanize"
)

func (m model) viewTasks() string {
	if m.loading {
		return "\n  Loading tasks...\n\n"
	}

	if m.err != nil {
		return fmt.Sprintf("\n  Error: %v\n\n  Press q to quit.\n\n", m.err)
	}

	if len(m.tasks) == 0 {
		return fmt.Sprintf("\n  No tasks found (status: %s)\n\n  Press q to quit, r to refresh.\n\n", m.status)
	}

	// Calculate available width (account for padding). Use effective dimensions so
	// iTerm2 and other terminals that report 0 or stale size still get correct layout.
	availableWidth := m.effectiveWidth() - 2
	if availableWidth < 40 {
		availableWidth = 40
	}

	// Determine layout based on terminal width
	useWideLayout := availableWidth >= 120
	useMediumLayout := availableWidth >= 80

	var b strings.Builder

	// Header bar (top/htop style)
	headerLine := strings.Builder{}

	// Title
	title := "TASKS"
	if m.projectName != "" {
		title = fmt.Sprintf("%s - %s", strings.ToUpper(m.projectName), title)
	}

	if m.status != "" {
		title += fmt.Sprintf(" [%s]", strings.ToUpper(m.status))
	}

	headerLine.WriteString(headerStyle.Render(title))
	headerLine.WriteString(" ")

	// Task count (visible = filtered or all)
	visCount := len(m.visibleIndices())
	totalCount := len(m.tasks)
	taskCountStr := " " + humanize.Comma(int64(visCount))

	if m.searchQuery != "" && totalCount != visCount {
		taskCountStr = fmt.Sprintf(" %s/%s", humanize.Comma(int64(visCount)), humanize.Comma(int64(totalCount)))
	}

	headerLine.WriteString(headerLabelStyle.Render("Tasks:"))
	headerLine.WriteString(headerValueStyle.Render(taskCountStr))
	headerLine.WriteString(" ")

	// Selected count
	selectedCount := len(m.selected)
	if selectedCount > 0 {
		headerLine.WriteString(headerLabelStyle.Render("Selected:"))
		headerLine.WriteString(headerValueStyle.Render(fmt.Sprintf(" %d", selectedCount)))
		headerLine.WriteString(" ")
	}

	// Auto-refresh status
	if m.autoRefresh {
		headerLine.WriteString(headerLabelStyle.Render("Updated:"))
		headerLine.WriteString(headerValueStyle.Render(" " + humanize.Time(m.lastUpdate)))
	} else {
		headerLine.WriteString(headerLabelStyle.Render("Auto-refresh:"))
		headerLine.WriteString(headerValueStyle.Render(" OFF"))
	}

	// Fill remaining space
	headerText := headerLine.String()
	if len(headerText) < availableWidth {
		padding := strings.Repeat(" ", availableWidth-len(headerText))
		headerText += headerValueStyle.Render(padding)
	}

	b.WriteString(headerText)
	b.WriteString("\n")

	// Separator line
	b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
	b.WriteString("\n")

	// Search mode prompt
	if m.searchMode {
		searchPrompt := "/" + m.searchQuery + "_"
		b.WriteString(helpStyle.Render("Search: " + searchPrompt + " (Enter=apply Esc=cancel)"))
		b.WriteString("\n")
	}

	// Bulk status update prompt
	if m.bulkStatusPrompt {
		selectedCount := len(m.selected)
		prompt := fmt.Sprintf("Bulk update %d task(s) - Select status: ", selectedCount)
		prompt += helpStyle.Render("d")
		prompt += "=Done "
		prompt += helpStyle.Render("i")
		prompt += "=In Progress "
		prompt += helpStyle.Render("t")
		prompt += "=Todo "
		prompt += helpStyle.Render("r")
		prompt += "=Review "
		prompt += helpStyle.Render("Esc")
		prompt += "=Cancel"
		b.WriteString(statusBarStyle.Render(prompt))
		b.WriteString("\n")
	}

	// Bulk status update result message
	if m.bulkStatusMsg != "" {
		msgLine := m.bulkStatusMsg
		if len(msgLine) > availableWidth-2 {
			msgLine = msgLine[:availableWidth-5] + "..."
		}
		b.WriteString("\n")
		b.WriteString(helpStyle.Render("  " + msgLine))
		b.WriteString("\n")
	}

	// Task list - adjust layout based on terminal width
	if useWideLayout {
		// Wide layout: multi-column or side-by-side
		m.renderWideTaskList(&b, availableWidth)
	} else if useMediumLayout {
		// Medium layout: single column with more details
		m.renderMediumTaskList(&b, availableWidth)
	} else {
		// Narrow layout: single column, minimal info
		m.renderNarrowTaskList(&b, availableWidth)
	}

	// Child agent result (one-line feedback)
	if m.childAgentMsg != "" {
		msgLine := m.childAgentMsg
		if len(msgLine) > availableWidth-2 {
			msgLine = msgLine[:availableWidth-5] + "..."
		}

		b.WriteString("\n")
		b.WriteString(helpStyle.Render("  " + msgLine))
		b.WriteString("\n")
	}

	// Status bar at bottom (htop style)
	b.WriteString("\n")
	b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
	b.WriteString("\n")

	// Status bar content
	statusBar := strings.Builder{}
	statusBar.WriteString(statusBarStyle.Render("Commands:"))
	statusBar.WriteString(" ")
	statusBar.WriteString(helpStyle.Render("↑↓/jk"))
	statusBar.WriteString(" nav  ")
	statusBar.WriteString(helpStyle.Render("/"))
	statusBar.WriteString(" search  ")
	statusBar.WriteString(helpStyle.Render("n/N"))
	statusBar.WriteString(" next/prev  ")
	statusBar.WriteString(helpStyle.Render("o/O"))
	statusBar.WriteString(" sort  ")
	statusBar.WriteString(helpStyle.Render("Space"))
	statusBar.WriteString(" select  ")
	statusBar.WriteString(helpStyle.Render("Tab"))
	statusBar.WriteString(" collapse  ")
	statusBar.WriteString(helpStyle.Render("s"))
	statusBar.WriteString(" details  ")
	statusBar.WriteString(helpStyle.Render("d/i/t/r"))
	statusBar.WriteString(" status  ")
	statusBar.WriteString(helpStyle.Render("D"))
	statusBar.WriteString(" bulk  ")
	statusBar.WriteString(helpStyle.Render("a"))
	statusBar.WriteString(" auto  ")
	statusBar.WriteString(helpStyle.Render("c"))
	statusBar.WriteString(" config  ")
	statusBar.WriteString(helpStyle.Render("p"))
	statusBar.WriteString(" scorecard  ")
	statusBar.WriteString(helpStyle.Render("w"))
	statusBar.WriteString(" w waves  ")
	statusBar.WriteString(helpStyle.Render("A"))
	statusBar.WriteString(" analysis  ")
	statusBar.WriteString(helpStyle.Render("b"))
	statusBar.WriteString(" jobs  ")
	statusBar.WriteString(helpStyle.Render("E"))
	statusBar.WriteString(" child agent  ")
	statusBar.WriteString(helpStyle.Render("L"))
	statusBar.WriteString(" plan  ")
	statusBar.WriteString(helpStyle.Render("?/h"))
	statusBar.WriteString(" help  ")
	statusBar.WriteString(helpStyle.Render("q"))
	statusBar.WriteString(" quit")

	// Fill remaining space
	statusText := statusBar.String()
	if len(statusText) < availableWidth {
		padding := strings.Repeat(" ", availableWidth-len(statusText))
		statusText += statusBarStyle.Render(padding)
	}

	b.WriteString(statusText)

	return b.String()
}

// renderNarrowTaskList renders tasks in a narrow terminal (< 80 chars).
func (m model) renderNarrowTaskList(b *strings.Builder, width int) {
	vis := m.visibleIndices()
	for idx, realIdx := range vis {
		task := m.tasks[realIdx]
		indent := m.indentForTask(realIdx)
		marker := m.treeMarkerForTask(realIdx)

		cursor := " "
		if m.cursor == idx {
			cursor = ">"
			if _, ok := m.selected[realIdx]; ok {
				cursor = "✓"
			}
		}

		// Minimal info: indent + tree marker + cursor + ID, status, truncated content
		line := fmt.Sprintf("%s%s%s %s", indent, marker, cursor, task.ID)

		if task.Status != "" {
			line += " " + statusStyle.Render(task.Status)
		}

		// Truncate content to fit
		content := task.Content
		if content == "" {
			content = task.LongDescription
		}

		maxContentWidth := width - len(line) - 10 // Reserve space
		if maxContentWidth > 0 && len(content) > maxContentWidth {
			content = content[:maxContentWidth-3] + "..."
		}

		if content != "" && maxContentWidth > 0 {
			line += " " + content
		}

		if m.cursor == idx {
			line = highlightRow(line, width, true)
		} else {
			line = normalStyle.Render(line)
		}

		b.WriteString(line)
		b.WriteString("\n")
	}
}

// renderMediumTaskList renders tasks in a medium terminal (80-120 chars).
func (m model) renderMediumTaskList(b *strings.Builder, width int) {
	// Column headers aligned with column width constants
	header := fmt.Sprintf("%-*s %-*s %-*s %-*s %-*s %s",
		colCursor, "", colIDMedium, "ID", colStatus, "STATUS", colPriority, "PRIORITY", colPRI, "PRI", "DESCRIPTION")
	b.WriteString(helpStyle.Render(header))
	b.WriteString("\n")
	b.WriteString(borderStyle.Render(strings.Repeat("─", width)))
	b.WriteString("\n")

	minDescWidth := width - (colCursor + 1 + colIDMedium + 1 + colStatus + 1 + colPriority + 1 + colPRI + 1)
	if minDescWidth < 10 {
		minDescWidth = 10
	}

	vis := m.visibleIndices()
	for idx, realIdx := range vis {
		task := m.tasks[realIdx]
		indent := m.indentForTask(realIdx)
		marker := m.treeMarkerForTask(realIdx)

		cursor := "   "
		if m.cursor == idx {
			cursor = " → "
			if _, ok := m.selected[realIdx]; ok {
				cursor = " ✓ "
			}
		}

		taskID := truncatePad(task.ID, colIDMedium)

		statusStr := task.Status
		if statusStr == "" {
			statusStr = "---"
		}

		statusStr = truncatePad(statusStr, colStatus)

		priorityFull := strings.ToUpper(task.Priority)
		if priorityFull == "" {
			priorityFull = "---"
		}

		priorityFull = truncatePad(priorityFull, colPriority)

		priorityShort := "-"
		if task.Priority != "" {
			priorityShort = strings.ToUpper(task.Priority[:1])
		}

		priorityShort = truncatePad(priorityShort, colPRI)

		content := task.Content
		if content == "" {
			content = task.LongDescription
		}

		if len(content) > minDescWidth {
			content = content[:minDescWidth-3] + "..."
		}

		if content == "" {
			content = "(no description)"
		}

		line := fmt.Sprintf("%-*s %-*s %-*s %-*s %-*s %s",
			colCursor, cursor, colIDMedium, taskID, colStatus, statusStr, colPriority, priorityFull, colPRI, priorityShort, content)

		if task.Priority != "" {
			switch strings.ToLower(task.Priority) {
			case models.PriorityHigh:
				line = strings.Replace(line, priorityShort, highPriorityStyle.Render(priorityShort), 1)
			case models.PriorityMedium:
				line = strings.Replace(line, priorityShort, mediumPriorityStyle.Render(priorityShort), 1)
			case models.PriorityLow:
				line = strings.Replace(line, priorityShort, lowPriorityStyle.Render(priorityShort), 1)
			}
		}

		line = indent + marker + line
		if m.cursor == idx {
			line = highlightRow(line, width, true)
		} else {
			line = normalStyle.Render(line)
		}

		b.WriteString(line)
		b.WriteString("\n")
	}
}

// Wide-layout constants: compact column widths to maximize space for Description.
// Single space between columns; fixed total = 52 so description gets (width - 53) or more.
const (
	wideColCursor          = 3  // " → " or " ✓ "
	wideColID              = 18 // T-xxxxxxxxxxxxx
	wideColStatus          = 11 // "In Progress"
	wideColPriority        = 8  // "PRIORITY" / "medium"
	wideColPRI             = 3  // H/M/L
	wideColOLD             = 3  // "OLD" or "   "
	wideColSpaces          = 6  // one space between each of 7 columns
	wideFixedColsTotal     = wideColCursor + wideColID + wideColStatus + wideColPriority + wideColPRI + wideColOLD + wideColSpaces
	wideMinDescWidth       = 50
	wideTagsColMin         = 24
	wideShowTagsThreshold  = 160
	wideFixedPlusDescSpace = wideFixedColsTotal + 1
)

// renderWideTaskList renders tasks in a wide terminal (>= 120 chars). Uses full width:
// description column grows with terminal width; TAGS column appears when width >= 160.
func (m model) renderWideTaskList(b *strings.Builder, width int) {
	// Compute description and optional tags column widths so each row fills width.
	maxDescWidth := width - wideFixedPlusDescSpace
	if maxDescWidth < wideMinDescWidth {
		maxDescWidth = wideMinDescWidth
	}

	tagsWidth := 0
	if width >= wideShowTagsThreshold {
		// Reserve space for TAGS; description gets the rest.
		tagsWidth = wideTagsColMin

		maxDescWidth = width - wideFixedPlusDescSpace - 1 - tagsWidth
		if maxDescWidth < wideMinDescWidth {
			maxDescWidth = wideMinDescWidth

			tagsWidth = width - wideFixedPlusDescSpace - maxDescWidth - 1
			if tagsWidth < 5 {
				tagsWidth = 0
			}
		}
	}

	descHeader := truncatePad("DESCRIPTION", maxDescWidth)

	header := fmt.Sprintf("%-*s %-*s %-*s %-*s %-*s %-*s %s",
		wideColCursor, "", wideColID, "ID", wideColStatus, "STATUS", wideColPriority, "PRIORITY", wideColPRI, "PRI", wideColOLD, "OLD", descHeader)
	if tagsWidth > 0 {
		header += " " + truncatePad("TAGS", tagsWidth)
	}

	b.WriteString(helpStyle.Render(header))
	b.WriteString("\n")
	b.WriteString(borderStyle.Render(strings.Repeat("─", width)))
	b.WriteString("\n")

	vis := m.visibleIndices()
	for idx, realIdx := range vis {
		task := m.tasks[realIdx]
		indent := m.indentForTask(realIdx)
		marker := m.treeMarkerForTask(realIdx)

		cursor := "   "
		if m.cursor == idx {
			cursor = " → "
			if _, ok := m.selected[realIdx]; ok {
				cursor = " ✓ "
			}
		}

		taskID := truncatePad(task.ID, wideColID)

		statusStr := task.Status
		if statusStr == "" {
			statusStr = "---"
		}

		statusStr = truncatePad(statusStr, wideColStatus)

		priorityFull := strings.ToUpper(task.Priority)
		if priorityFull == "" {
			priorityFull = "---"
		}

		priorityFull = truncatePad(priorityFull, wideColPriority)

		priorityShort := "-"
		if task.Priority != "" {
			priorityShort = strings.ToUpper(task.Priority[:1])
		}

		priorityShort = truncatePad(priorityShort, wideColPRI)

		oldStr := "   "
		if isOldSequentialID(task.ID) {
			oldStr = "OLD"
		}

		oldIndicator := truncatePad(oldStr, wideColOLD)

		content := task.Content
		if content == "" {
			content = task.LongDescription
		}

		if len(content) > maxDescWidth {
			content = content[:maxDescWidth-3] + "..."
		}

		if content == "" {
			content = "(no description)"
		}

		content = truncatePad(content, maxDescWidth)

		line := fmt.Sprintf("%-*s %-*s %-*s %-*s %-*s %-*s %s",
			wideColCursor, cursor, wideColID, taskID, wideColStatus, statusStr, wideColPriority, priorityFull, wideColPRI, priorityShort, wideColOLD, oldIndicator, content)

		if tagsWidth > 0 && len(task.Tags) > 0 {
			tagsStr := strings.Join(task.Tags, ",")
			if len(tagsStr) > tagsWidth {
				tagsStr = tagsStr[:tagsWidth-3] + "..."
			}

			line += " " + helpStyle.Render(tagsStr)
		}

		if task.Priority != "" {
			switch strings.ToLower(task.Priority) {
			case models.PriorityHigh:
				line = strings.Replace(line, priorityShort, highPriorityStyle.Render(priorityShort), 1)
			case models.PriorityMedium:
				line = strings.Replace(line, priorityShort, mediumPriorityStyle.Render(priorityShort), 1)
			case models.PriorityLow:
				line = strings.Replace(line, priorityShort, lowPriorityStyle.Render(priorityShort), 1)
			}
		}

		if isOldSequentialID(task.ID) {
			line = strings.Replace(line, "OLD", oldIDStyle.Render("OLD"), 1)
		}

		line = indent + marker + line
		if m.cursor == idx {
			line = highlightRow(line, width, true)
		} else {
			line = normalStyle.Render(line)
		}

		b.WriteString(line)
		b.WriteString("\n")
	}
}

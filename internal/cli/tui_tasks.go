// tui_tasks.go — TUI tasks tab: narrow/medium/wide list rendering + bubble/list and bubble/table integration.
package cli

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/table"
	"charm.land/bubbles/v2/textinput"
	humanize "github.com/dustin/go-humanize"

	"github.com/davidl71/exarp-go/internal/database"
)

// defaultTableColumns defines the columns for bubble/table sprintboard view.
var defaultTableColumns = []table.Column{
	{Title: "ID", Width: 18},
	{Title: "Status", Width: 12},
	{Title: "Priority", Width: 10},
	{Title: "Description", Width: 40},
	{Title: "Tags", Width: 20},
}

// taskItem implements list.Item for bubble/list integration.
type taskItem struct {
	task     *database.Todo2Task
	index    int
	selected bool
}

func (t taskItem) FilterValue() string {
	return t.task.Content + " " + t.task.Status + " " + t.task.Priority + " " + strings.Join(t.task.Tags, " ")
}

func (t taskItem) Title() string {
	return t.task.Content
}

func (t taskItem) Description() string {
	return fmt.Sprintf("[%s] %s", t.task.Status, t.task.Priority)
}

// itemsFromTasks converts database tasks to list items.
func itemsFromTasks(tasks []*database.Todo2Task) []list.Item {
	items := make([]list.Item, len(tasks))
	for i, task := range tasks {
		items[i] = taskItem{task: task, index: i}
	}
	return items
}

// viewTasksList renders the task list using bubble/list.
func (m model) viewTasksList() string {
	if len(m.tasks) == 0 {
		return "\n  No tasks to display\n\n"
	}
	m.taskList.SetSize(m.effectiveWidth()-2, m.effectiveHeight()-10)
	return m.taskList.View()
}

// rowsFromTasks converts database tasks to table rows.
func rowsFromTasks(tasks []*database.Todo2Task) []table.Row {
	rows := make([]table.Row, len(tasks))
	for i, task := range tasks {
		rows[i] = table.Row{
			task.ID,
			task.Status,
			task.Priority,
			truncatePad(task.Content, 38),
			truncatePad(strings.Join(task.Tags, ","), 18),
		}
	}
	return rows
}

// viewTasksTable renders the task list using bubble/table.
func (m model) viewTasksTable() string {
	if len(m.tasks) == 0 {
		return "\n  No tasks to display\n\n"
	}
	m.taskTable.SetRows(rowsFromTasks(m.tasks))
	m.taskTable.SetSize(m.effectiveWidth()-2, m.effectiveHeight()-8)
	return m.taskTable.View()
}

// viewSpinner returns the spinner view with a message.
func (m model) viewSpinner() string {
	m.taskSpinner.SetWidth(m.effectiveWidth() - 4)
	return "\n  " + m.taskSpinner.View() + " " + m.spinnerMessage + "\n\n"
}

func (m model) viewTasks() string {
	if m.loading {
		return m.viewSpinner()
	}

	// Command palette overlay
	if m.showCommandPalette {
		return m.viewCommandPalette()
	}

	if m.err != nil {
		return fmt.Sprintf("\n  Error: %v\n\n  Press q to quit.\n\n", m.err)
	}

	// Use bubble/list if enabled
	if m.useBubbleList {
		return m.viewTasksList()
	}

	// Use bubble/table if enabled
	if m.useBubbleTable {
		return m.viewTasksTable()
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
	} else {
		title += " [ALL]"
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

	// Inline task creation prompt
	if m.createMode {
		b.WriteString(helpStyle.Render("New task: " + m.createInput + "_ (Enter=create Esc=cancel)"))
		b.WriteString("\n")
	}

	// Bulk status update prompt
	if m.bulkStatusPrompt {
		selectedCount := len(m.selected)
		prompt := fmt.Sprintf("Bulk update %d task(s) - Select status: ", selectedCount)
		prompt += helpStyle.Render(bindingList(m.bindingsFor(KeyActionStatusDone)))
		prompt += "=Done "
		prompt += helpStyle.Render(bindingList(m.bindingsFor(KeyActionStatusInProgress)))
		prompt += "=In Progress "
		prompt += helpStyle.Render(bindingList(m.bindingsFor(KeyActionStatusTodo)))
		prompt += "=Todo "
		prompt += helpStyle.Render(bindingList(m.bindingsFor(KeyActionStatusReview)))
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

	// Width-responsive status bar: show more commands on wider terminals
	b.WriteString(m.renderTaskStatusBar(availableWidth))

	return b.String()
}

// renderNarrowTaskList renders tasks in a narrow terminal (< 80 chars).
func (m model) renderNarrowTaskList(b *strings.Builder, width int) {
	vis := m.visibleIndices()
	pageSize := m.taskListPageSize()
	end := m.viewportOffset + pageSize
	if end > len(vis) {
		end = len(vis)
	}

	for idx := m.viewportOffset; idx < end; idx++ {
		realIdx := vis[idx]
		task := m.tasks[realIdx]
		isCursor := m.cursor == idx
		_, isSelected := m.selected[realIdx]
		cursor := cursorMarkerNarrow(isCursor, isSelected)

		line := fmt.Sprintf("%s %s", cursor, task.ID)
		if task.Status != "" {
			line += " " + styleStatus(task.Status)
		}

		maxContentWidth := width - len(line) - 10
		content := task.Content
		if content == "" {
			content = task.LongDescription
		}

		if maxContentWidth > 0 && len(content) > maxContentWidth {
			content = content[:maxContentWidth-3] + "..."
		}

		if content != "" && maxContentWidth > 0 {
			line += " " + content
		}

		b.WriteString(renderRow(line, m.indentForTask(realIdx), m.treeMarkerForTask(realIdx), isCursor, width))
		b.WriteString("\n")

		if m.spaciousMode {
			desc := spaciousDescLine(task.LongDescription, width-4)
			b.WriteString(helpStyle.Render("   " + desc))
			b.WriteString("\n")
		}
	}

	if len(vis) > pageSize {
		scrollInfo := fmt.Sprintf("  [%d-%d of %d]", m.viewportOffset+1, end, len(vis))
		b.WriteString(helpStyle.Render(scrollInfo))
		b.WriteString("\n")
	}
}

// renderMediumTaskList renders tasks in a medium terminal (80-120 chars).
func (m model) renderMediumTaskList(b *strings.Builder, width int) {
	header := fmt.Sprintf("%-*s %-*s %-*s %-*s %-*s %s",
		colCursor, "", colIDMedium, "ID", colStatus, "STATUS", colPriority, "PRIORITY", colPRI, "PRI", "DESCRIPTION")
	b.WriteString(helpStyle.Render(header))
	b.WriteString("\n")
	b.WriteString(borderStyle.Render(strings.Repeat("─", width)))
	b.WriteString("\n")

	descWidth := width - (colCursor + 1 + colIDMedium + 1 + colStatus + 1 + colPriority + 1 + colPRI + 1)
	if descWidth < 10 {
		descWidth = 10
	}

	vis := m.visibleIndices()
	pageSize := m.taskListPageSize()
	end := m.viewportOffset + pageSize
	if end > len(vis) {
		end = len(vis)
	}

	for idx := m.viewportOffset; idx < end; idx++ {
		realIdx := vis[idx]
		task := m.tasks[realIdx]
		isCursor := m.cursor == idx
		_, isSelected := m.selected[realIdx]

		cursor := cursorMarkerWide(isCursor, isSelected)
		taskID := truncatePad(task.ID, colIDMedium)
		statusStr := formatStatus(task.Status, colStatus)
		priFull := formatPriorityFull(task.Priority, colPriority)
		priShort := formatPriorityShort(task.Priority, colPRI)
		content := taskContent(task.Content, task.LongDescription, descWidth)

		line := fmt.Sprintf("%-*s %-*s %-*s %-*s %-*s %s",
			colCursor, cursor, colIDMedium, taskID, colStatus, statusStr, colPriority, priFull, colPRI, priShort, content)
		line = styleStatusInLine(line, task.Status, statusStr)
		line = stylePriorityInLine(line, task.Priority, priShort)
		b.WriteString(renderRow(line, m.indentForTask(realIdx), m.treeMarkerForTask(realIdx), isCursor, width))
		b.WriteString("\n")

		if m.spaciousMode {
			desc := spaciousDescLine(task.LongDescription, width-6)
			b.WriteString(helpStyle.Render("     " + desc))
			b.WriteString("\n")
		}
	}

	if len(vis) > pageSize {
		scrollInfo := fmt.Sprintf("  [%d-%d of %d]", m.viewportOffset+1, end, len(vis))
		b.WriteString(helpStyle.Render(scrollInfo))
		b.WriteString("\n")
	}
}

// renderWideTaskList renders tasks in a wide terminal (>= 120 chars). Uses full width:
// description column grows with terminal width; TAGS column appears when width >= 160.
func (m model) renderWideTaskList(b *strings.Builder, width int) {
	maxDescWidth := width - wideFixedPlusDescSpace
	if maxDescWidth < wideMinDescWidth {
		maxDescWidth = wideMinDescWidth
	}

	tagsWidth := 0
	if width >= wideShowTagsThreshold {
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

	header := fmt.Sprintf("%-*s %-*s %-*s %-*s %-*s %-*s %s",
		wideColCursor, "", wideColID, "ID", wideColStatus, "STATUS", wideColPriority, "PRIORITY", wideColPRI, "PRI", wideColOLD, "OLD", truncatePad("DESCRIPTION", maxDescWidth))
	if tagsWidth > 0 {
		header += " " + truncatePad("TAGS", tagsWidth)
	}

	b.WriteString(helpStyle.Render(header))
	b.WriteString("\n")
	b.WriteString(borderStyle.Render(strings.Repeat("─", width)))
	b.WriteString("\n")

	vis := m.visibleIndices()
	pageSize := m.taskListPageSize()
	end := m.viewportOffset + pageSize
	if end > len(vis) {
		end = len(vis)
	}

	for idx := m.viewportOffset; idx < end; idx++ {
		realIdx := vis[idx]
		task := m.tasks[realIdx]
		isCursor := m.cursor == idx
		_, isSelected := m.selected[realIdx]

		cursor := cursorMarkerWide(isCursor, isSelected)
		taskID := truncatePad(task.ID, wideColID)
		statusStr := formatStatus(task.Status, wideColStatus)
		priFull := formatPriorityFull(task.Priority, wideColPriority)
		priShort := formatPriorityShort(task.Priority, wideColPRI)

		oldIndicator := truncatePad("   ", wideColOLD)
		if isOldSequentialID(task.ID) {
			oldIndicator = truncatePad("OLD", wideColOLD)
		}

		content := truncatePad(taskContent(task.Content, task.LongDescription, maxDescWidth), maxDescWidth)

		line := fmt.Sprintf("%-*s %-*s %-*s %-*s %-*s %-*s %s",
			wideColCursor, cursor, wideColID, taskID, wideColStatus, statusStr, wideColPriority, priFull, wideColPRI, priShort, wideColOLD, oldIndicator, content)

		if tagsWidth > 0 && len(task.Tags) > 0 {
			tagsStr := strings.Join(task.Tags, ",")
			if len(tagsStr) > tagsWidth {
				tagsStr = tagsStr[:tagsWidth-3] + "..."
			}

			line += " " + helpStyle.Render(tagsStr)
		}

		line = styleStatusInLine(line, task.Status, statusStr)
		line = stylePriorityInLine(line, task.Priority, priShort)
		if isOldSequentialID(task.ID) {
			line = strings.Replace(line, "OLD", oldIDStyle.Render("OLD"), 1)
		}

		b.WriteString(renderRow(line, m.indentForTask(realIdx), m.treeMarkerForTask(realIdx), isCursor, width))
		b.WriteString("\n")

		if m.spaciousMode {
			desc := spaciousDescLine(task.LongDescription, width-6)
			b.WriteString(helpStyle.Render("     " + desc))
			b.WriteString("\n")
		}
	}

	if len(vis) > pageSize {
		scrollInfo := fmt.Sprintf("  [%d-%d of %d]", m.viewportOffset+1, end, len(vis))
		b.WriteString(helpStyle.Render(scrollInfo))
		b.WriteString("\n")
	}
}

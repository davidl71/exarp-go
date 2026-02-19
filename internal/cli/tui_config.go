package cli

import (
	"fmt"
	"path/filepath"
	"strings"
)

func (m model) viewConfig() string {
	var b strings.Builder

	availableWidth := m.effectiveWidth() - 2
	if availableWidth < 40 {
		availableWidth = 40
	}

	// Header bar (top/htop style)
	headerLine := strings.Builder{}

	title := "CONFIG"
	if m.projectName != "" {
		title = fmt.Sprintf("%s - %s", strings.ToUpper(m.projectName), title)
	}

	headerLine.WriteString(headerStyle.Render(title))
	headerLine.WriteString(" ")

	// Config path
	if m.projectRoot != "" {
		configPath := filepath.Join(m.projectRoot, ".exarp", "config.pb")
		if len(configPath) > 40 {
			configPath = "..." + configPath[len(configPath)-37:]
		}

		headerLine.WriteString(headerLabelStyle.Render("Config:"))
		headerLine.WriteString(headerValueStyle.Render(fmt.Sprintf(" %s", configPath)))
		headerLine.WriteString(" ")
	}

	// Unsaved changes indicator
	if m.configChanged {
		headerLine.WriteString(headerLabelStyle.Render("Status:"))
		headerLine.WriteString(headerValueStyle.Render(" UNSAVED"))
	}

	// Fill remaining space
	headerText := headerLine.String()
	if len(headerText) < availableWidth {
		padding := strings.Repeat(" ", availableWidth-len(headerText))
		headerText += headerValueStyle.Render(padding)
	}

	b.WriteString(headerText)
	b.WriteString("\n")
	b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
	b.WriteString("\n")

	if m.configSaveMessage != "" {
		msgLine := m.configSaveMessage
		if len(msgLine) > availableWidth {
			msgLine = msgLine[:availableWidth-3] + "..."
		}

		if m.configSaveSuccess {
			b.WriteString(headerValueStyle.Render("  ✅ " + msgLine))
		} else {
			b.WriteString(highPriorityStyle.Render("  ❌ " + msgLine))
		}

		b.WriteString("\n")
		b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
		b.WriteString("\n")
	}

	// Column headers
	header := fmt.Sprintf("%-4s %-20s %s", "PID", "SECTION", "DESCRIPTION")
	b.WriteString(helpStyle.Render(header))
	b.WriteString("\n")
	b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
	b.WriteString("\n")

	// Config sections list
	for i, section := range m.configSections {
		// Cursor indicator
		cursor := "   "
		if m.configCursor == i {
			cursor = " → "
		}

		// Section number (like PID)
		sectionNum := fmt.Sprintf("%-3d", i+1)

		// Section name
		sectionName := section.name
		if len(sectionName) > 20 {
			sectionName = sectionName[:17] + "..."
		}

		// Description
		description := section.description
		maxDescWidth := availableWidth - 30

		if maxDescWidth > 0 && len(description) > maxDescWidth {
			description = description[:maxDescWidth-3] + "..."
		}

		// Build line
		line := fmt.Sprintf("%s%s %-20s %s", cursor, sectionNum, sectionName, description)

		// Apply styling and full-row highlight for current line
		if m.configCursor == i {
			line = highlightRow(line, availableWidth, true)
		} else {
			line = normalStyle.Render(line)
		}

		b.WriteString(line)
		b.WriteString("\n")
	}

	// Status bar
	b.WriteString("\n")
	b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
	b.WriteString("\n")

	statusBar := strings.Builder{}
	statusBar.WriteString(statusBarStyle.Render("Commands:"))
	statusBar.WriteString(" ")
	statusBar.WriteString(helpStyle.Render("↑↓"))
	statusBar.WriteString(" nav  ")
	statusBar.WriteString(helpStyle.Render("Enter"))
	statusBar.WriteString(" section  ")
	statusBar.WriteString(helpStyle.Render("u"))
	statusBar.WriteString(" update (protobuf)  ")
	statusBar.WriteString(helpStyle.Render("s"))
	statusBar.WriteString(" save  ")
	statusBar.WriteString(helpStyle.Render("r"))
	statusBar.WriteString(" reload  ")
	statusBar.WriteString(helpStyle.Render("c"))
	statusBar.WriteString(" tasks  ")
	statusBar.WriteString(helpStyle.Render("q"))
	statusBar.WriteString(" quit")

	statusText := statusBar.String()
	if len(statusText) < availableWidth {
		padding := strings.Repeat(" ", availableWidth-len(statusText))
		statusText += statusBarStyle.Render(padding)
	}

	b.WriteString(statusText)

	return b.String()
}

func (m model) viewConfigSection() string {
	availableWidth := m.effectiveWidth() - 2
	if availableWidth < 40 {
		availableWidth = 40
	}

	maxContentLines := m.effectiveHeight() - 8
	if maxContentLines < 6 {
		maxContentLines = 6
	}

	wrapped := wordWrap(m.configSectionText, availableWidth)
	allLines := strings.Split(wrapped, "\n")

	var b strings.Builder

	title := "CONFIG SECTION"
	if m.projectName != "" {
		title = fmt.Sprintf("%s - %s", strings.ToUpper(m.projectName), title)
	}

	b.WriteString(headerStyle.Render(title))
	b.WriteString(" ")
	b.WriteString(headerLabelStyle.Render("Esc/Enter/Space=close"))
	b.WriteString("\n")
	b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
	b.WriteString("\n")

	for i, line := range allLines {
		if i >= maxContentLines {
			b.WriteString(helpStyle.Render("  ... (truncated)\n"))
			break
		}

		b.WriteString(normalStyle.Render(line))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
	b.WriteString("\n")
	b.WriteString(statusBarStyle.Render("Press Esc, Enter, or Space to close"))

	return b.String()
}

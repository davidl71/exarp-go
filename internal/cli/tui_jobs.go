package cli

import (
	"fmt"
	"strings"
)

func (m model) viewJobs() string {
	availableWidth := m.effectiveWidth() - 2
	if availableWidth < 40 {
		availableWidth = 40
	}

	wrapWidth := availableWidth - 2
	if wrapWidth < 30 {
		wrapWidth = 30
	}

	var b strings.Builder

	title := "BACKGROUND JOBS"
	if m.projectName != "" {
		title = fmt.Sprintf("%s - %s", strings.ToUpper(m.projectName), title)
	}

	b.WriteString(headerStyle.Render(title))
	b.WriteString(" ")
	b.WriteString(headerLabelStyle.Render("Esc/b=back"))
	b.WriteString("\n")
	b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
	b.WriteString("\n")

	if len(m.jobs) == 0 {
		b.WriteString("\n  ")
		b.WriteString(helpStyle.Render("No background jobs. Launch a child agent (E from task/handoff/wave) to add jobs."))
		b.WriteString("\n\n")
		b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
		b.WriteString("\n")
		b.WriteString(statusBarStyle.Render("Commands: Esc or b back  q quit"))

		return b.String()
	}

	// Job detail view
	if m.jobsDetailIndex >= 0 && m.jobsDetailIndex < len(m.jobs) {
		return m.viewJobDetail(m.jobs[m.jobsDetailIndex], availableWidth, wrapWidth)
	}

	// Job list view
	b.WriteString(helpStyle.Render("  #   KIND      PID    STARTED              PROMPT"))
	b.WriteString("\n")
	b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
	b.WriteString("\n")

	for i, job := range m.jobs {
		cursor := "   "
		if m.jobsCursor == i {
			cursor = " → "
		}

		pidStr := "—"
		if job.Pid > 0 {
			pidStr = fmt.Sprintf("%d", job.Pid)
		}

		started := job.StartedAt.Format("2006-01-02 15:04")

		promptPreview := job.Prompt
		if len(promptPreview) > availableWidth-45 {
			promptPreview = promptPreview[:availableWidth-48] + "..."
		}

		line := fmt.Sprintf("%s %-3d  %-6s  %-6s  %s  %s", cursor, i+1, string(job.Kind), pidStr, started, promptPreview)
		if m.jobsCursor == i {
			line = highlightRow(line, availableWidth, true)
		} else {
			line = normalStyle.Render(line)
		}

		b.WriteString("  ")
		b.WriteString(line)
		b.WriteString("\n")
	}

	b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
	b.WriteString("\n")
	b.WriteString(statusBarStyle.Render("Enter view detail  Esc/b back  q quit"))

	return b.String()
}

func (m model) viewJobDetail(job BackgroundJob, availableWidth, wrapWidth int) string {
	var b strings.Builder

	title := "JOB DETAIL"
	if m.projectName != "" {
		title = fmt.Sprintf("%s - %s", strings.ToUpper(m.projectName), title)
	}

	b.WriteString(headerStyle.Render(title))
	b.WriteString(" ")
	b.WriteString(headerLabelStyle.Render("Esc/Enter back"))
	b.WriteString("\n")
	b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
	b.WriteString("\n")

	b.WriteString("  ")
	b.WriteString(headerLabelStyle.Render("Kind:"))
	b.WriteString(headerValueStyle.Render(" " + string(job.Kind)))
	b.WriteString("  ")
	b.WriteString(headerLabelStyle.Render("PID:"))
	b.WriteString(headerValueStyle.Render(fmt.Sprintf(" %d", job.Pid)))
	b.WriteString("  ")
	b.WriteString(headerLabelStyle.Render("Started:"))
	b.WriteString(headerValueStyle.Render(" " + job.StartedAt.Format("2006-01-02 15:04:05")))
	b.WriteString("\n  ")
	b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
	b.WriteString("\n")
	b.WriteString("  ")
	b.WriteString(headerLabelStyle.Render("Prompt:"))
	b.WriteString("\n  ")

	for _, line := range strings.Split(wordWrap(job.Prompt, wrapWidth), "\n") {
		b.WriteString(normalStyle.Render("  "+line) + "\n")
	}

	b.WriteString("  ")
	b.WriteString(headerLabelStyle.Render("Output:"))
	b.WriteString("\n")

	if job.Output != "" {
		for _, line := range strings.Split(wordWrap(job.Output, wrapWidth), "\n") {
			b.WriteString("  ")
			b.WriteString(normalStyle.Render(line))
			b.WriteString("\n")
		}
	} else {
		b.WriteString("  ")
		b.WriteString(helpStyle.Render("(running or interactive job — no capture)"))
		b.WriteString("\n")
	}

	b.WriteString("  ")
	b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
	b.WriteString("\n")
	b.WriteString(statusBarStyle.Render("Esc/Enter back  q quit"))

	return b.String()
}

// report_plan_generate.go — Report plan: plan markdown generation, YAML repair, and display helpers.
// See also: report_plan.go
package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"github.com/davidl71/exarp-go/internal/config"
	"github.com/davidl71/exarp-go/internal/models"
)

// ─── Contents ───────────────────────────────────────────────────────────────
//   generatePlanMarkdown
//   repairPlanFile — repairPlanFile reads an existing .plan.md, recomputes frontmatter and the "## 3. Iterative Milestones" section from Todo2, and writes the file back so Cursor stripping is undone without losing other body content.
//   getTaskFileRefs — getTaskFileRefs extracts file/code references from task metadata (T-1769980664971).
//   planDisplayName — planDisplayName returns a short display name from plan title (e.g. "github.com/davidl71/exarp-go" -> "exarp-go").
//   escapeYAMLString — escapeYAMLString escapes a string for YAML double-quoted value.
//   todo2StatusToCursorStatus — todo2StatusToCursorStatus maps Todo2 status to Cursor plan todo status (pending, in_progress, completed).
//   tableCellSafe — tableCellSafe makes a string safe for a markdown table cell (no |, newlines -> space).
// ────────────────────────────────────────────────────────────────────────────

// ─── generatePlanMarkdown ───────────────────────────────────────────────────
func generatePlanMarkdown(ctx context.Context, projectRoot, planTitle string) (string, error) {
	displayName := planDisplayName(planTitle)
	info, _ := getProjectInfo(projectRoot)

	overview := getStr(info, "description")
	if overview == "" {
		overview = fmt.Sprintf("Deliver and maintain %s with clear milestones and quality gates.", displayName)
	}

	store := NewDefaultTaskStore(projectRoot)
	list, _ := store.ListTasks(ctx, nil)
	tasks := tasksFromPtrs(list)
	taskByID := make(map[string]Todo2Task)

	for _, t := range tasks {
		taskByID[t.ID] = t
	}

	snippet := getPlanningSnippet(projectRoot)
	actions, _ := getNextActions(projectRoot)
	risks, _ := getRisksAndBlockers(projectRoot)

	// Todos for frontmatter (Cursor-buildable: id, content, status so Cursor displays them as Todos)
	todoEntries := make([]struct{ id, content, status string }, 0, len(actions))

	for _, a := range actions {
		id := getStr(a, "task_id")
		if id == "" {
			continue
		}

		content := getStr(a, "name")
		status := "pending"

		if t, ok := taskByID[id]; ok {
			if content == "" {
				content = t.Content
			}

			status = todo2StatusToCursorStatus(t.Status)
		}

		if content == "" {
			content = id
		}

		content = strings.TrimSpace(content)
		if len(content) > 120 {
			content = content[:117] + "..."
		}

		todoEntries = append(todoEntries, struct{ id, content, status string }{id, content, status})
	}

	// Execution order and waves (same wave = can run in parallel); computed once for frontmatter and section 4
	orderedIDs, waves, _, _ := BacklogExecutionOrder(tasks, nil)
	if max := config.MaxTasksPerWave(); max > 0 {
		waves = LimitWavesByMaxTasks(waves, max)
	}

	var sb strings.Builder

	// YAML frontmatter (Cursor plan format; todos = [{ id, content, status }]; waves = parallelizable groups)
	sb.WriteString("---\n")
	sb.WriteString(fmt.Sprintf("name: %s Plan\n", displayName))
	sb.WriteString(fmt.Sprintf("overview: %q\n", escapeYAMLString(overview)))

	if len(todoEntries) > 0 {
		sb.WriteString("todos:\n")

		for _, e := range todoEntries {
			sb.WriteString(fmt.Sprintf("  - id: %s\n", e.id))
			sb.WriteString(fmt.Sprintf("    content: %q\n", escapeYAMLString(e.content)))
			sb.WriteString(fmt.Sprintf("    status: %s\n", e.status))
		}
	} else {
		sb.WriteString("todos: []\n")
	}

	if len(waves) > 0 {
		levelOrder := make([]int, 0, len(waves))
		for k := range waves {
			levelOrder = append(levelOrder, k)
		}

		sort.Ints(levelOrder)
		sb.WriteString("waves:\n")

		for _, level := range levelOrder {
			ids := waves[level]
			if len(ids) == 0 {
				continue
			}

			quoted := make([]string, len(ids))
			for i, id := range ids {
				quoted[i] = fmt.Sprintf("%q", id)
			}

			sb.WriteString("  - [" + strings.Join(quoted, ", ") + "]\n")
		}
	}

	planDate := time.Now().Format("2006-01-02")

	sb.WriteString("isProject: true\n")
	sb.WriteString("status: draft\n")
	sb.WriteString(fmt.Sprintf("last_updated: %q\n", planDate))
	sb.WriteString("tag_hints: [planning]\n")
	// Referenced by / Agents in frontmatter so Cursor UI can show "Referenced by N" (same as body block below)
	sb.WriteString("referenced_by:\n")
	sb.WriteString("  - .cursor/agents/wave-task-runner.md\n")
	sb.WriteString("  - .cursor/agents/wave-verifier.md\n")
	sb.WriteString("  - .cursor/rules/plan-execution.mdc\n")
	sb.WriteString("agents:\n")
	sb.WriteString("  - name: wave-task-runner\n")
	sb.WriteString("    path: .cursor/agents/wave-task-runner.md\n")
	sb.WriteString("    role: Run one task per wave from this plan\n")
	sb.WriteString("  - name: wave-verifier\n")
	sb.WriteString("    path: .cursor/agents/wave-verifier.md\n")
	sb.WriteString("    role: Verify wave outcomes and update status\n")
	sb.WriteString("---\n\n")
	sb.WriteString("*Regenerate with: `exarp-go -tool report -args '{\"action\":\"plan\"}'`. Do not edit frontmatter in Cursor; it may strip fields and break Build.*\n\n")
	sb.WriteString(fmt.Sprintf("# %s Plan\n\n", displayName))
	sb.WriteString(fmt.Sprintf("**Generated:** %s\n\n", planDate))
	sb.WriteString("**Status:** draft\n\n")
	sb.WriteString(fmt.Sprintf("**Last updated:** %s\n\n", planDate))
	sb.WriteString("**Referenced by:** [wave-task-runner](.cursor/agents/wave-task-runner.md), [wave-verifier](.cursor/agents/wave-verifier.md), [plan-execution](.cursor/rules/plan-execution.mdc)\n\n")
	sb.WriteString("**Tag hints:** `#planning`\n\n")
	sb.WriteString("## Agents\n\n")
	sb.WriteString("| Agent | Role |\n")
	sb.WriteString("|-------|------|\n")
	sb.WriteString("| [wave-task-runner](.cursor/agents/wave-task-runner.md) | Run one task per wave from this plan |\n")
	sb.WriteString("| [wave-verifier](.cursor/agents/wave-verifier.md) | Verify wave outcomes and update status |\n\n")
	sb.WriteString("---\n\n")
	sb.WriteString("## Scope\n\n")
	sb.WriteString("**Purpose:** " + overview + "\n\n")
	sb.WriteString("**Success criteria:** Clear milestones and quality gates; backlog aligned with execution order.\n\n")
	sb.WriteString("---\n\n")

	// 1. Technical Foundation
	sb.WriteString("## 1. Technical Foundation\n\n")

	metrics, err := getCodebaseMetrics(projectRoot)
	if err == nil {
		sb.WriteString(fmt.Sprintf("- **Project type:** %s\n", getStr(info, "type")))
		sb.WriteString(fmt.Sprintf("- **Tools:** %v | **Prompts:** %v | **Resources:** %v\n", metrics["tools"], metrics["prompts"], metrics["resources"]))
		sb.WriteString(fmt.Sprintf("- **Codebase:** %v files (Go: %v)\n", metrics["total_files"], metrics["go_files"]))
	}

	sb.WriteString("- **Storage:** Todo2 (SQLite primary, JSON fallback)\n")
	sb.WriteString("- **Invariants:** Use Makefile targets; prefer report/task_workflow over direct file edits\n\n")

	if path, ok := snippet["critical_path"].([]string); ok && len(path) > 0 {
		sb.WriteString("**Critical path (longest dependency chain):** ")
		sb.WriteString(strings.Join(path, " → "))
		sb.WriteString("\n\n")
	}

	sb.WriteString("---\n\n")

	// 2. Backlog Tasks (table: Task | Priority | Description)
	sb.WriteString("## 2. Backlog Tasks\n\n")
	sb.WriteString("| Task | Priority | Description |\n")
	sb.WriteString("|------|----------|-------------|\n")

	if len(actions) > 0 {
		for _, a := range actions {
			taskID := getStr(a, "task_id")
			priority := getStr(a, "priority")
			name := getStr(a, "name")

			desc := name
			if t, ok := taskByID[taskID]; ok && t.LongDescription != "" {
				desc = tableCellSafe(t.LongDescription)
			} else if name != "" {
				desc = tableCellSafe(name)
			} else {
				desc = taskID
			}

			sb.WriteString(fmt.Sprintf("| **%s** | %s | %s |\n", taskID, priority, desc))
		}
	} else {
		if ids, ok := snippet["suggested_backlog_order"].([]string); ok && len(ids) > 0 {
			for _, id := range ids {
				t := taskByID[id]
				desc := t.Content

				if t.LongDescription != "" {
					desc = tableCellSafe(t.LongDescription)
				} else {
					desc = tableCellSafe(t.Content)
				}

				sb.WriteString(fmt.Sprintf("| **%s** | %s | %s |\n", id, t.Priority, desc))
			}
		} else {
			sb.WriteString("| — | — | *(Add Todo2 tasks; run report with include_planning to populate.)* |\n")
		}
	}

	sb.WriteString("\n---\n\n")

	// 3. Iterative Milestones (checkboxes — Cursor buildable)
	sb.WriteString("## 3. Iterative Milestones\n\n")
	sb.WriteString("Each milestone is independently valuable. Check off as done.\n\n")

	if len(actions) > 0 {
		for _, a := range actions {
			name := getStr(a, "name")
			taskID := getStr(a, "task_id")

			if name == "" {
				name = taskID
			}

			line := fmt.Sprintf("- [ ] **%s** (%s)", name, taskID)

			if t, ok := taskByID[taskID]; ok {
				if refs := getTaskFileRefs(&t); len(refs) > 0 {
					line += " — " + strings.Join(refs, ", ")
				}
			}

			sb.WriteString(line + "\n")
		}
	} else {
		if ids, ok := snippet["suggested_backlog_order"].([]string); ok && len(ids) > 0 {
			for _, id := range ids {
				line := fmt.Sprintf("- [ ] %s", id)

				if t, ok := taskByID[id]; ok {
					if refs := getTaskFileRefs(&t); len(refs) > 0 {
						line += " — " + strings.Join(refs, ", ")
					}
				}

				sb.WriteString(line + "\n")
			}
		} else {
			sb.WriteString("- [ ] *(Add Todo2 tasks to populate milestones)*\n")
		}
	}

	sb.WriteString("\n---\n\n")

	// 4. Recommended Execution Order (critical path + waves/parallelizable + optional mermaid)
	cp, _ := snippet["critical_path"].([]string)

	sb.WriteString("## 4. Recommended Execution Order\n\n")

	if len(cp) > 0 {
		sb.WriteString("```mermaid\nflowchart TD\n")

		for i := 0; i < len(cp); i++ {
			node := fmt.Sprintf("CP%d", i+1)
			label := cp[i]
			sb.WriteString(fmt.Sprintf("    %s[%s]\n", node, label))
		}

		for i := 1; i < len(cp); i++ {
			sb.WriteString(fmt.Sprintf("    CP%d --> CP%d\n", i, i+1))
		}

		sb.WriteString("```\n\n")

		for i, taskID := range cp {
			sb.WriteString(fmt.Sprintf("%d. **%s**\n", i+1, taskID))
		}
		// Parallel: next actions not on critical path
		cpSet := make(map[string]bool)
		for _, id := range cp {
			cpSet[id] = true
		}

		var parallel []string

		for _, a := range actions {
			id := getStr(a, "task_id")
			if !cpSet[id] {
				parallel = append(parallel, id)
			}
		}

		if len(parallel) > 0 {
			sb.WriteString("\n**Parallel:** " + strings.Join(parallel, ", ") + "\n")
		}
	} else {
		sb.WriteString("1. Complete backlog in dependency order (see Milestones).\n")

		if len(actions) > 0 {
			for i, a := range actions {
				sb.WriteString(fmt.Sprintf("%d. **%s** (%s)\n", i+1, getStr(a, "name"), getStr(a, "task_id")))
			}
		}
	}

	// Waves: same wave = can run in parallel (from BacklogExecutionOrder dependency levels)
	if len(waves) > 0 {
		sb.WriteString("\n### Waves (same wave = can run in parallel)\n\n")

		levelOrder := make([]int, 0, len(waves))
		for k := range waves {
			levelOrder = append(levelOrder, k)
		}

		sort.Ints(levelOrder)

		for _, level := range levelOrder {
			ids := waves[level]
			if len(ids) == 0 {
				continue
			}

			if len(ids) >= 2 {
				sb.WriteString(fmt.Sprintf("- **Wave %d** (parallel): %s\n", level, strings.Join(ids, ", ")))
			} else {
				sb.WriteString(fmt.Sprintf("- **Wave %d:** %s\n", level, ids[0]))
			}
		}

		sb.WriteString("\n")

		if len(orderedIDs) > 0 {
			sb.WriteString("**Full order:** " + strings.Join(orderedIDs, ", ") + "\n")
		}
	}

	sb.WriteString("\n---\n\n")

	// 5. Open Questions
	sb.WriteString("## 5. Open Questions\n\n")

	if len(risks) > 0 {
		for _, r := range risks {
			sb.WriteString(fmt.Sprintf("- %s", getStr(r, "description")))

			if id := getStr(r, "task_id"); id != "" {
				sb.WriteString(fmt.Sprintf(" (%s)", id))
			}

			sb.WriteString("\n")
		}
	} else {
		sb.WriteString("- *(Add open questions or decisions needed during implementation.)*\n")
	}

	sb.WriteString("\n---\n\n")

	// 6. Out-of-Scope / Deferred
	sb.WriteString("## 6. Out-of-Scope / Deferred\n\n")
	// Low-priority backlog tasks not in next actions, or placeholder
	var deferred []string

	actionIDs := make(map[string]bool)
	for _, a := range actions {
		actionIDs[getStr(a, "task_id")] = true
	}

	for _, t := range tasks {
		if IsPendingStatus(t.Status) && t.Priority == models.PriorityLow && !actionIDs[t.ID] {
			deferred = append(deferred, fmt.Sprintf("**%s** (%s) — low priority", t.ID, t.Content))
		}
	}

	if len(deferred) > 0 {
		for _, line := range deferred {
			sb.WriteString("- " + line + "\n")
		}
	} else {
		sb.WriteString("- *(Add deferred or out-of-scope items as needed.)*\n")
	}

	sb.WriteString("\n---\n\n")

	// 7. Key Files / Implementation Notes
	sb.WriteString("## 7. Key Files / Implementation Notes\n\n")
	sb.WriteString("- *(Add key files or implementation notes; or leave as placeholder.)*\n")
	sb.WriteString("\n---\n\n")

	// 8. References
	sb.WriteString("## 8. References\n\n")
	sb.WriteString("- [docs/BACKLOG_EXECUTION_PLAN.md](docs/BACKLOG_EXECUTION_PLAN.md) — full wave breakdown\n")
	sb.WriteString("- *(Add other plan or doc links as needed.)*\n")

	return sb.String(), nil
}

// ─── repairPlanFile ─────────────────────────────────────────────────────────
// repairPlanFile reads an existing .plan.md, recomputes frontmatter and the "## 3. Iterative Milestones" section from Todo2, and writes the file back so Cursor stripping is undone without losing other body content.
func repairPlanFile(ctx context.Context, projectRoot, planPath, planTitle string) (string, error) {
	data, err := os.ReadFile(planPath)
	if err != nil {
		return "", fmt.Errorf("read plan file: %w", err)
	}
	content := string(data)

	// Locate frontmatter end (after second "---")
	firstDash := strings.Index(content, "---\n")
	if firstDash < 0 {
		return "", fmt.Errorf("plan file has no frontmatter")
	}
	secondDash := strings.Index(content[firstDash+4:], "\n---\n")
	if secondDash < 0 {
		return "", fmt.Errorf("plan file frontmatter not closed")
	}
	frontmatterEnd := firstDash + 4 + secondDash + len("\n---\n")

	// Locate "## 3. Iterative Milestones" and "## 4. Recommended Execution Order"
	marker3 := "## 3. Iterative Milestones"
	marker4 := "## 4. Recommended Execution Order"
	idx3 := strings.Index(content, marker3)
	idx4 := strings.Index(content, marker4)
	if idx3 < 0 || idx4 < 0 || idx4 <= idx3 {
		return "", fmt.Errorf("plan file missing ## 3 or ## 4 section")
	}

	bodyBefore := content[frontmatterEnd:idx3]
	bodyAfter := content[idx4:]

	// Recompute todoEntries and waves from Todo2 (same as generatePlanMarkdown)
	displayName := planDisplayName(planTitle)
	info, _ := getProjectInfo(projectRoot)
	overview := getStr(info, "description")
	if overview == "" {
		overview = fmt.Sprintf("Deliver and maintain %s with clear milestones and quality gates.", displayName)
	}

	store := NewDefaultTaskStore(projectRoot)
	list, _ := store.ListTasks(ctx, nil)
	tasks := tasksFromPtrs(list)
	taskByID := make(map[string]Todo2Task)
	for _, t := range tasks {
		taskByID[t.ID] = t
	}

	actions, _ := getNextActions(projectRoot)
	todoEntries := make([]struct{ id, content, status string }, 0, len(actions))
	for _, a := range actions {
		id := getStr(a, "task_id")
		if id == "" {
			continue
		}
		contentStr := getStr(a, "name")
		status := "pending"
		if t, ok := taskByID[id]; ok {
			if contentStr == "" {
				contentStr = t.Content
			}
			status = todo2StatusToCursorStatus(t.Status)
		}
		if contentStr == "" {
			contentStr = id
		}
		contentStr = strings.TrimSpace(contentStr)
		if len(contentStr) > 120 {
			contentStr = contentStr[:117] + "..."
		}
		todoEntries = append(todoEntries, struct{ id, content, status string }{id, contentStr, status})
	}

	_, waves, _, _ := BacklogExecutionOrder(tasks, nil)
	if max := config.MaxTasksPerWave(); max > 0 {
		waves = LimitWavesByMaxTasks(waves, max)
	}

	planDate := time.Now().Format("2006-01-02")

	// Build new frontmatter (same as generatePlanMarkdown)
	var fm strings.Builder
	fm.WriteString("---\n")
	fm.WriteString(fmt.Sprintf("name: %s Plan\n", displayName))
	fm.WriteString(fmt.Sprintf("overview: %q\n", escapeYAMLString(overview)))
	if len(todoEntries) > 0 {
		fm.WriteString("todos:\n")
		for _, e := range todoEntries {
			fm.WriteString(fmt.Sprintf("  - id: %s\n", e.id))
			fm.WriteString(fmt.Sprintf("    content: %q\n", escapeYAMLString(e.content)))
			fm.WriteString(fmt.Sprintf("    status: %s\n", e.status))
		}
	} else {
		fm.WriteString("todos: []\n")
	}
	if len(waves) > 0 {
		levelOrder := make([]int, 0, len(waves))
		for k := range waves {
			levelOrder = append(levelOrder, k)
		}
		sort.Ints(levelOrder)
		fm.WriteString("waves:\n")
		for _, level := range levelOrder {
			ids := waves[level]
			if len(ids) == 0 {
				continue
			}
			quoted := make([]string, len(ids))
			for i, id := range ids {
				quoted[i] = fmt.Sprintf("%q", id)
			}
			fm.WriteString("  - [" + strings.Join(quoted, ", ") + "]\n")
		}
	}
	fm.WriteString("isProject: true\n")
	fm.WriteString("status: draft\n")
	fm.WriteString(fmt.Sprintf("last_updated: %q\n", planDate))
	fm.WriteString("tag_hints: [planning]\n")
	fm.WriteString("referenced_by:\n")
	fm.WriteString("  - .cursor/agents/wave-task-runner.md\n")
	fm.WriteString("  - .cursor/agents/wave-verifier.md\n")
	fm.WriteString("  - .cursor/rules/plan-execution.mdc\n")
	fm.WriteString("agents:\n")
	fm.WriteString("  - name: wave-task-runner\n")
	fm.WriteString("    path: .cursor/agents/wave-task-runner.md\n")
	fm.WriteString("    role: Run one task per wave from this plan\n")
	fm.WriteString("  - name: wave-verifier\n")
	fm.WriteString("    path: .cursor/agents/wave-verifier.md\n")
	fm.WriteString("    role: Verify wave outcomes and update status\n")
	fm.WriteString("---\n\n")

	// Build new "## 3. Iterative Milestones" section with checkboxes
	reminder := "*Regenerate with: `exarp-go -tool report -args '{\"action\":\"plan\"}'`. Do not edit frontmatter in Cursor; it may strip fields and break Build.*\n\n"
	var milestones strings.Builder
	milestones.WriteString("## 3. Iterative Milestones\n\n")
	milestones.WriteString("Each milestone is independently valuable. Check off as done.\n\n")
	if len(actions) > 0 {
		for _, a := range actions {
			name := getStr(a, "name")
			taskID := getStr(a, "task_id")
			if name == "" {
				name = taskID
			}
			status := "pending"
			if t, ok := taskByID[taskID]; ok {
				status = todo2StatusToCursorStatus(t.Status)
			}
			box := "[ ]"
			if status == "completed" {
				box = "[x]"
			}
			line := fmt.Sprintf("- %s **%s** (%s)", box, name, taskID)
			if t, ok := taskByID[taskID]; ok {
				if refs := getTaskFileRefs(&t); len(refs) > 0 {
					line += " — " + strings.Join(refs, ", ")
				}
			}
			milestones.WriteString(line + "\n")
		}
	} else {
		milestones.WriteString("- [ ] *(Add Todo2 tasks to populate milestones)*\n")
	}
	milestones.WriteString("\n---\n\n")

	repaired := fm.String() + reminder + bodyBefore + milestones.String() + bodyAfter
	if err := os.WriteFile(planPath, []byte(repaired), 0644); err != nil {
		return "", fmt.Errorf("write repaired plan: %w", err)
	}
	return fmt.Sprintf("Plan format repaired: %s (frontmatter and ## 3. Iterative Milestones restored)", planPath), nil
}

// ─── getTaskFileRefs ────────────────────────────────────────────────────────
// getTaskFileRefs extracts file/code references from task metadata (T-1769980664971).
// Returns paths from discovered_from, planning_doc, and planning_links.RelatedDocs.
func getTaskFileRefs(task *Todo2Task) []string {
	var refs []string

	seen := make(map[string]bool)
	add := func(s string) {
		s = strings.TrimSpace(s)
		if s != "" && !seen[s] {
			seen[s] = true

			refs = append(refs, s)
		}
	}

	if task.Metadata != nil {
		if f, ok := task.Metadata["discovered_from"].(string); ok {
			add(f)
		}

		if pd, ok := task.Metadata["planning_doc"].(string); ok {
			add(pd)
		}

		if linkMeta := GetPlanningLinkMetadata(task); linkMeta != nil {
			if linkMeta.PlanningDoc != "" {
				add(linkMeta.PlanningDoc)
			}

			for _, r := range linkMeta.RelatedDocs {
				add(r)
			}
		}
	}

	return refs
}

// ─── planDisplayName ────────────────────────────────────────────────────────
// planDisplayName returns a short display name from plan title (e.g. "github.com/davidl71/exarp-go" -> "exarp-go").
func planDisplayName(title string) string {
	if title == "" {
		return "Project"
	}

	if strings.Contains(title, "/") {
		return filepath.Base(title)
	}

	return title
}

// ─── escapeYAMLString ───────────────────────────────────────────────────────
// escapeYAMLString escapes a string for YAML double-quoted value.
func escapeYAMLString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", " ")

	return strings.TrimSpace(s)
}

// ─── todo2StatusToCursorStatus ──────────────────────────────────────────────
// todo2StatusToCursorStatus maps Todo2 status to Cursor plan todo status (pending, in_progress, completed).
func todo2StatusToCursorStatus(todo2Status string) string {
	switch strings.TrimSpace(strings.ToLower(todo2Status)) {
	case strings.ToLower(models.StatusDone):
		return "completed"
	case strings.ToLower(models.StatusInProgress), "inprogress":
		return "in_progress"
	case strings.ToLower(models.StatusReview):
		return "in_progress"
	default:
		return "pending"
	}
}

// ─── tableCellSafe ──────────────────────────────────────────────────────────
// tableCellSafe makes a string safe for a markdown table cell (no |, newlines -> space).
func tableCellSafe(s string) string {
	s = strings.ReplaceAll(s, "|", "-")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.TrimSpace(s)

	if len(s) > 200 {
		s = s[:197] + "..."
	}

	return s
}

// aggregateProjectData aggregates all project data for overview.

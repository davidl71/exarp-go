package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/internal/security"
	"github.com/davidl71/mcp-go-core/pkg/mcp/response"
	"gopkg.in/yaml.v3"
)

// PlanTodo represents a todo entry from a Cursor plan file's YAML frontmatter.
type PlanTodo struct {
	ID      string `yaml:"id"`
	Content string `yaml:"content"`
	Status  string `yaml:"status"` // pending, in_progress, completed
}

// PlanFrontmatter holds parsed YAML frontmatter from a .plan.md file.
type PlanFrontmatter struct {
	Todos []PlanTodo `yaml:"todos"`
}

// parsePlanFile parses a .plan.md file and returns frontmatter todos and checkbox state from milestones.
// Milestone format: - [ ] **Name** (T-ID) or - [x] **Name** (T-ID)
func parsePlanFile(path string) ([]PlanTodo, map[string]bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("read plan file: %w", err)
	}
	content := string(data)

	// Extract YAML frontmatter
	var fm PlanFrontmatter
	frontmatterRe := regexp.MustCompile(`(?s)^---\r?\n(.*?)\r?\n---\r?\n`)
	matches := frontmatterRe.FindStringSubmatch(content)
	if len(matches) >= 2 {
		if err := yaml.Unmarshal([]byte(matches[1]), &fm); err != nil {
			return nil, nil, fmt.Errorf("parse frontmatter: %w", err)
		}
	}

	// Parse milestone checkboxes: - [x] **Name** (T-ID) or - [ ] **Name** (T-ID)
	// Also match - [x] Name (T-ID) without bold
	checkboxRe := regexp.MustCompile(`^\s*-\s*\[\s*([ xX])\s*\]\s*(?:\*\*[^*]+\*\*|.+?)\s*\(\s*([A-Za-z0-9_-]+)\s*\)`)
	checkboxState := make(map[string]bool)
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		subs := checkboxRe.FindStringSubmatch(line)
		if len(subs) >= 3 {
			taskID := strings.TrimSpace(subs[2])
			checked := strings.ToLower(subs[1]) == "x"
			checkboxState[taskID] = checked
		}
	}

	return fm.Todos, checkboxState, nil
}

// cursorStatusToTodo2 maps Cursor plan status to Todo2 status.
func cursorStatusToTodo2(s string) string {
	switch strings.TrimSpace(strings.ToLower(s)) {
	case "completed", "done":
		return "Done"
	case "in_progress", "inprogress":
		return "In Progress"
	case "review":
		return "Review"
	default:
		return "Todo"
	}
}

// todo2StatusToPlanStatus maps Todo2 status to Cursor plan frontmatter status.
func todo2StatusToPlanStatus(s string) string {
	switch strings.TrimSpace(s) {
	case "Done":
		return "done"
	case "In Progress":
		return "in_progress"
	case "Review":
		return "review"
	default:
		return "pending"
	}
}

// writePlanFileBack updates a .plan.md file so frontmatter todos[].status and milestone checkboxes match Todo2 status.
func writePlanFileBack(planPath string, todos []PlanTodo, statusByID map[string]string) error {
	data, err := os.ReadFile(planPath)
	if err != nil {
		return fmt.Errorf("read plan file: %w", err)
	}
	content := string(data)

	// Build updated frontmatter: same todos but status from Todo2
	for i := range todos {
		if s, ok := statusByID[todos[i].ID]; ok {
			todos[i].Status = todo2StatusToPlanStatus(s)
		}
	}
	fm := PlanFrontmatter{Todos: todos}
	fmYAML, err := yaml.Marshal(&fm)
	if err != nil {
		return fmt.Errorf("marshal frontmatter: %w", err)
	}
	frontmatterRe := regexp.MustCompile(`(?s)^---\r?\n(.*?)\r?\n---\r?\n`)
	content = frontmatterRe.ReplaceAllString(content, "---\n"+strings.TrimSpace(string(fmYAML))+"\n---\n")

	// Replace checkbox lines: [x] if Done, [ ] otherwise
	checkboxRe := regexp.MustCompile(`^(\s*-\s*)\[\s*([ xX])\s*\](\s*(?:\*\*[^*]+\*\*|.+?)\s*\(\s*)([A-Za-z0-9_-]+)(\s*\))$`)
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		subs := checkboxRe.FindStringSubmatch(line)
		if len(subs) >= 5 {
			taskID := strings.TrimSpace(subs[4])
			checked := statusByID[taskID] == "Done"
			box := "[ ]"
			if checked {
				box = "[x]"
			}
			lines[i] = subs[1] + box + subs[3] + taskID + subs[5]
		}
	}
	content = strings.Join(lines, "\n")

	return os.WriteFile(planPath, []byte(content), 0644)
}

// handleTaskWorkflowSyncFromPlan parses a Cursor .plan.md file and creates/updates Todo2 tasks.
// Bidirectional: reads plan file (todos + checkboxes) and syncs to Todo2; optionally writes plan back to match Todo2 status.
// When Cursor "builds" a plan (user checks off milestones), run this to sync status to Todo2 (T-1769980693841).
// Params: planning_doc (path to .plan.md, required), dry_run (optional, default false), write_plan (optional, default true = update plan file from Todo2).
func handleTaskWorkflowSyncFromPlan(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	planningDoc, _ := params["planning_doc"].(string)
	if planningDoc == "" {
		return nil, fmt.Errorf("planning_doc is required for sync_from_plan/sync_plan_status")
	}

	projectRoot, err := security.GetProjectRoot(".")
	if err != nil {
		return nil, fmt.Errorf("sync_from_plan: %w", err)
	}

	// Resolve path (relative to project root)
	planPath := planningDoc
	if !filepath.IsAbs(planPath) {
		planPath = filepath.Join(projectRoot, planPath)
	}
	if err := ValidatePlanningLink(projectRoot, planningDoc); err != nil {
		return nil, fmt.Errorf("sync_from_plan: invalid planning_doc: %w", err)
	}

	todos, checkboxState, err := parsePlanFile(planPath)
	if err != nil {
		return nil, fmt.Errorf("sync_from_plan: %w", err)
	}

	dryRun, _ := params["dry_run"].(bool)

	tasks, err := LoadTodo2Tasks(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("sync_from_plan: load tasks: %w", err)
	}
	taskByID := make(map[string]*Todo2Task)
	for i := range tasks {
		taskByID[tasks[i].ID] = &tasks[i]
	}

	var created, updated []string
	db, dbErr := database.GetDB()
	useDB := dbErr == nil && db != nil

	for _, todo := range todos {
		if todo.ID == "" {
			continue
		}
		todo2Status := cursorStatusToTodo2(todo.Status)
		// Checkbox in milestones overrides frontmatter status for Done/Todo
		if checked, ok := checkboxState[todo.ID]; ok {
			if checked {
				todo2Status = "Done"
			} else if todo2Status == "Done" {
				// Checkbox unchecked overrides completed in frontmatter
				todo2Status = "Todo"
			}
		}

		existing, exists := taskByID[todo.ID]
		if !exists {
			if dryRun {
				created = append(created, todo.ID+" (would create)")
				continue
			}
			// Create new task
			content := todo.Content
			if content == "" {
				content = todo.ID
			}
			newTask := &Todo2Task{
				ID:              todo.ID,
				Content:         content,
				LongDescription: content,
				Status:          todo2Status,
				Priority:        "medium",
			}
			if useDB {
				if err := database.CreateTask(ctx, newTask); err != nil {
					return nil, fmt.Errorf("sync_from_plan: create task %s: %w", todo.ID, err)
				}
			} else {
				tasks = append(tasks, *newTask)
				taskByID[todo.ID] = &tasks[len(tasks)-1]
			}
			created = append(created, todo.ID)
		} else {
			// Update existing task status if different
			if existing.Status != todo2Status {
				if dryRun {
					updated = append(updated, fmt.Sprintf("%s (would update %s -> %s)", todo.ID, existing.Status, todo2Status))
					continue
				}
				existing.Status = todo2Status
				if useDB {
					if err := database.UpdateTask(ctx, existing); err != nil {
						return nil, fmt.Errorf("sync_from_plan: update task %s: %w", todo.ID, err)
					}
				} else {
					for i := range tasks {
						if tasks[i].ID == todo.ID {
							tasks[i].Status = todo2Status
							break
						}
					}
				}
				updated = append(updated, todo.ID)
			}
		}
	}

	if !useDB && (len(created) > 0 || len(updated) > 0) && !dryRun {
		if err := SaveTodo2Tasks(projectRoot, tasks); err != nil {
			return nil, fmt.Errorf("sync_from_plan: save tasks: %w", err)
		}
	}
	if useDB && (len(created) > 0 || len(updated) > 0) && !dryRun {
		if err := SyncTodo2Tasks(projectRoot); err != nil {
			return nil, fmt.Errorf("sync_from_plan: sync to JSON: %w", err)
		}
	}

	// Bidirectional: write plan file back so checkboxes and frontmatter match Todo2 status (default true)
	writePlan := true
	if w, ok := params["write_plan"].(bool); ok {
		writePlan = w
	}
	if writePlan && !dryRun && len(todos) > 0 {
		statusByID := make(map[string]string)
		for _, todo := range todos {
			if todo.ID == "" {
				continue
			}
			if t, ok := taskByID[todo.ID]; ok {
				statusByID[todo.ID] = t.Status
			}
		}
		if useDB {
			// Reload from DB so we have latest status after updates
			for id := range statusByID {
				if t, err := database.GetTask(ctx, id); err == nil {
					statusByID[id] = t.Status
				}
			}
		}
		if err := writePlanFileBack(planPath, todos, statusByID); err != nil {
			return nil, fmt.Errorf("sync_from_plan: write plan back: %w", err)
		}
	}

	result := map[string]interface{}{
		"success":       true,
		"action":        "sync_from_plan",
		"plan_path":     planPath,
		"created":       created,
		"updated":       updated,
		"created_count": len(created),
		"updated_count": len(updated),
		"dry_run":       dryRun,
	}
	return response.FormatResult(result, "")
}

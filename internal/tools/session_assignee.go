// session_assignee.go — session assignee handlers.
package tools

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/davidl71/exarp-go/internal/config"
	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/framework"
	mcpresponse "github.com/davidl71/mcp-go-core/pkg/mcp/response"
	"github.com/spf13/cast"
)

func handleSessionPrompts(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	// Get optional filters
	mode := cast.ToString(params["mode"])
	persona := cast.ToString(params["persona"])
	category := cast.ToString(params["category"])
	keywords := cast.ToString(params["keywords"])

	limit := 50
	if l := cast.ToFloat64(params["limit"]); l > 0 {
		limit = int(l)
	}

	// Get all registered prompts from the prompts registry
	// We'll build a simple list from the known prompts
	allPrompts := []map[string]interface{}{
		// Original prompts (8)
		{"name": "align", "description": "Analyze Todo2 task alignment with project goals.", "category": "tasks"},
		{"name": "discover", "description": "Discover tasks from TODO comments, markdown, and orphaned tasks.", "category": "tasks"},
		{"name": "config", "description": "Generate IDE configuration files.", "category": "config"},
		{"name": "scan", "description": "Scan project dependencies for security vulnerabilities. Supports all languages via tool parameter.", "category": "security"},
		{"name": "scorecard", "description": "Generate comprehensive project health scorecard with all metrics.", "category": "reports"},
		{"name": "overview", "description": "Generate one-page project overview for stakeholders.", "category": "reports"},
		{"name": "dashboard", "description": "Display comprehensive project dashboard with key metrics and status overview.", "category": "reports"},
		{"name": "remember", "description": "Use AI session memory to persist insights.", "category": "memory"},
		// High-value workflow prompts (7)
		{"name": "daily_checkin", "description": "Daily check-in workflow: server status, blockers, git health.", "category": "workflow", "mode": "daily_checkin"},
		{"name": "sprint_start", "description": "Sprint start workflow: clean backlog, align tasks, queue work.", "category": "workflow", "mode": "sprint_planning"},
		{"name": "sprint_end", "description": "Sprint end workflow: test coverage, docs, security check.", "category": "workflow", "mode": "sprint_planning"},
		{"name": "pre_sprint", "description": "Pre-sprint cleanup workflow: duplicates, alignment, documentation.", "category": "workflow", "mode": "sprint_planning"},
		{"name": "post_impl", "description": "Post-implementation review workflow: docs, security, automation.", "category": "workflow"},
		{"name": "sync", "description": "Synchronize tasks between shared TODO table and Todo2.", "category": "tasks"},
		{"name": "dups", "description": "Find and consolidate duplicate Todo2 tasks.", "category": "tasks"},
		// mcp-generic-tools prompts
		{"name": "context", "description": "Manage LLM context with summarization and budget tools.", "category": "workflow"},
		{"name": "mode", "description": "Suggest optimal Cursor IDE mode (Agent vs Ask) for a task.", "category": "workflow"},
		// Task management prompts
		{"name": "task_update", "description": "Update Todo2 task status using proper MCP tools - never edit JSON directly.", "category": "tasks"},
	}

	// Filter prompts
	filtered := []map[string]interface{}{}

	for _, prompt := range allPrompts {
		// Filter by mode
		if mode != "" {
			if promptMode, ok := prompt["mode"].(string); !ok || promptMode != mode {
				continue
			}
		}

		// Filter by category
		if category != "" {
			if promptCategory, ok := prompt["category"].(string); !ok || promptCategory != category {
				continue
			}
		}

		// Filter by keywords (search in name and description)
		if keywords != "" {
			keywordsLower := strings.ToLower(keywords)
			name, _ := prompt["name"].(string)
			desc, _ := prompt["description"].(string)

			if !strings.Contains(strings.ToLower(name), keywordsLower) &&
				!strings.Contains(strings.ToLower(desc), keywordsLower) {
				continue
			}
		}

		// Persona filtering would require persona mapping - skip for now
		// (would need to add persona field to prompts)

		filtered = append(filtered, prompt)
		if len(filtered) >= limit {
			break
		}
	}

	result := map[string]interface{}{
		"success": true,
		"method":  "native_go",
		"prompts": filtered,
		"total":   len(filtered),
		"filters": map[string]interface{}{
			"mode":     mode,
			"persona":  persona,
			"category": category,
			"keywords": keywords,
			"limit":    limit,
		},
	}

	return mcpresponse.FormatResult(result, "")
}

// handleSessionAssignee handles the assignee action - manages task assignments.
func handleSessionAssignee(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	subAction := cast.ToString(params["sub_action"])
	if subAction == "" {
		subAction = "list"
	}

	switch subAction {
	case "list":
		return handleSessionAssigneeList(ctx, params)
	case "assign":
		return handleSessionAssigneeAssign(ctx, params)
	case "unassign":
		return handleSessionAssigneeUnassign(ctx, params)
	default:
		return nil, fmt.Errorf("unknown sub_action: %s (use 'list', 'assign', or 'unassign')", subAction)
	}
}

// handleSessionAssigneeList lists tasks with their assignees.
func handleSessionAssigneeList(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	includeUnassigned := true
	if _, ok := params["include_unassigned"]; ok {
		includeUnassigned = cast.ToBool(params["include_unassigned"])
	}

	statusFilter := cast.ToString(params["status_filter"])

	// Get tasks from database
	var tasks []*database.Todo2Task

	var err error

	if statusFilter != "" {
		filters := &database.TaskFilters{Status: &statusFilter}
		tasks, err = database.ListTasks(ctx, filters)
	} else {
		// Get all pending tasks
		tasks, err = database.ListTasks(ctx, &database.TaskFilters{})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	// Get assignee information from database
	// Note: assignee is stored in tasks_lock table, we'll need to query it
	db, err := database.GetDB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database: %w", err)
	}

	assignments := []map[string]interface{}{}

	for _, task := range tasks {
		// Query assignee from tasks table
		var assignee sql.NullString

		var assignedAt sql.NullInt64 // assigned_at is INTEGER (Unix timestamp)

		err := db.QueryRowContext(ctx, `
			SELECT assignee, assigned_at
			FROM tasks
			WHERE id = ?
		`, task.ID).Scan(&assignee, &assignedAt)

		hasAssignee := err == nil && assignee.Valid && assignee.String != ""

		if !includeUnassigned && !hasAssignee {
			continue
		}

		assignment := map[string]interface{}{
			"task_id":     task.ID,
			"content":     task.Content,
			"status":      task.Status,
			"assignee":    nil,
			"assigned_at": nil,
		}

		if hasAssignee {
			assignment["assignee"] = assignee.String
			if assignedAt.Valid {
				// Convert Unix timestamp to ISO 8601 string
				assignment["assigned_at"] = time.Unix(assignedAt.Int64, 0).Format(time.RFC3339)
			}
		}

		assignments = append(assignments, assignment)
	}

	result := map[string]interface{}{
		"success":     true,
		"method":      "native_go",
		"assignments": assignments,
		"total":       len(assignments),
		"filters": map[string]interface{}{
			"status_filter":      statusFilter,
			"include_unassigned": includeUnassigned,
		},
	}

	return mcpresponse.FormatResult(result, "")
}

// handleSessionAssigneeAssign assigns a task to an agent/human/host.
func handleSessionAssigneeAssign(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	taskID := cast.ToString(params["task_id"])
	if taskID == "" {
		return nil, fmt.Errorf("task_id parameter is required")
	}

	assigneeName := cast.ToString(params["assignee_name"])
	if assigneeName == "" {
		return nil, fmt.Errorf("assignee_name parameter is required")
	}

	assigneeType := "agent"
	if t := strings.TrimSpace(cast.ToString(params["assignee_type"])); t != "" {
		assigneeType = t
	}

	dryRun := false
	if _, ok := params["dry_run"]; ok {
		dryRun = cast.ToBool(params["dry_run"])
	}

	// Get agent ID (for agent type)
	if assigneeType == "agent" {
		agentID, err := database.GetAgentID()
		if err == nil {
			// Use agent ID format: {agent-type}-{hostname}-{pid}
			// But for assignment, we might want just the name
			// For now, use the provided assigneeName
		}

		_ = agentID // Use if needed
	}

	// Use database ClaimTaskForAgent for atomic assignment
	// This handles locking and prevents race conditions
	leaseDuration := config.TaskLockLease()

	result, err := database.ClaimTaskForAgent(ctx, taskID, assigneeName, leaseDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to assign task: %w", err)
	}

	if !result.Success {
		if result.WasLocked {
			response := map[string]interface{}{
				"success":   false,
				"method":    "native_go",
				"error":     fmt.Sprintf("Task already assigned to %s", result.LockedBy),
				"task_id":   taskID,
				"locked_by": result.LockedBy,
			}

			return mcpresponse.FormatResult(response, "")
		}

		return nil, fmt.Errorf("task assignment failed: %w", result.Error)
	}

	response := map[string]interface{}{
		"success":       true,
		"method":        "native_go",
		"task_id":       taskID,
		"assignee":      assigneeName,
		"assignee_type": assigneeType,
		"assigned_at":   time.Now().Format(time.RFC3339),
		"dry_run":       dryRun,
	}

	return mcpresponse.FormatResult(response, "")
}

// handleSessionAssigneeUnassign unassigns a task.
func handleSessionAssigneeUnassign(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	taskID := cast.ToString(params["task_id"])
	if taskID == "" {
		return nil, fmt.Errorf("task_id parameter is required")
	}

	agentID, err := database.GetAgentID()
	if err != nil {
		return nil, fmt.Errorf("failed to get agent ID: %w", err)
	}

	// Release task lock (this unassigns the task)
	err = database.ReleaseTask(ctx, taskID, agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to unassign task: %w", err)
	}

	response := map[string]interface{}{
		"success":       true,
		"method":        "native_go",
		"task_id":       taskID,
		"unassigned_at": time.Now().Format(time.RFC3339),
	}

	return mcpresponse.FormatResult(response, "")
}

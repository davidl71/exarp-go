package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/davidl71/exarp-go/internal/framework"
	mcpcli "github.com/davidl71/mcp-go-core/pkg/mcp/cli"
)

// handleTaskCommand handles task subcommands (list, status, update, create, show) using ParseArgs result.
func handleTaskCommand(server framework.MCPServer, parsed *mcpcli.Args) error {
	subcommand := parsed.Subcommand
	if subcommand == "" && len(parsed.Positional) > 0 {
		subcommand = parsed.Positional[0]
	}

	if subcommand == "" {
		return showTaskUsage()
	}

	switch subcommand {
	case "list":
		return handleTaskListParsed(server, parsed)
	case "status":
		return handleTaskStatus(server, parsed.Positional)
	case "update":
		return handleTaskUpdateParsed(server, parsed)
	case "create":
		return handleTaskCreateParsed(server, parsed)
	case "show":
		return handleTaskShow(server, parsed.Positional)
	case "delete":
		return handleTaskDelete(server, parsed.Positional)
	case "sync":
		return handleTaskSync(server)
	case "help":
		return showTaskUsage()
	default:
		return fmt.Errorf("unknown task command: %s (use: list, status, update, create, show, delete, sync, help)", subcommand)
	}
}

// handleTaskListParsed handles "task list" command using ParseArgs result.
func handleTaskListParsed(server framework.MCPServer, parsed *mcpcli.Args) error {
	status := parsed.GetFlag("status", "")
	priority := parsed.GetFlag("priority", "")
	tag := parsed.GetFlag("tag", "")
	order := parsed.GetFlag("order", "")
	limit, _ := strconv.Atoi(parsed.GetFlag("limit", "0"))

	toolArgs := map[string]interface{}{
		"action":     "sync",
		"sub_action": "list",
	}
	if status != "" {
		toolArgs["status"] = status
	}

	if priority != "" {
		toolArgs["priority"] = priority
	}

	if tag != "" {
		toolArgs["filter_tag"] = tag
	}

	if limit > 0 {
		toolArgs["limit"] = limit
	}

	if order == "execution" || order == "dependency" {
		toolArgs["order"] = order
	}

	return executeTaskWorkflow(server, toolArgs)
}

// handleTaskList handles "task list" command (legacy, used by tests).
func handleTaskList(server framework.MCPServer, args []string) error {
	parsed := mcpcli.ParseArgs(append([]string{"task", "list"}, args...))
	return handleTaskListParsed(server, parsed)
}

// handleTaskStatus handles "task status <task-id>" command.
func handleTaskStatus(server framework.MCPServer, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("task status requires a task ID")
	}

	taskID := args[0]

	// Build task_workflow args to get task details
	toolArgs := map[string]interface{}{
		"action":     "sync",
		"sub_action": "list",
		"task_id":    taskID,
	}

	return executeTaskWorkflow(server, toolArgs)
}

// handleTaskUpdateParsed handles "task update" command using ParseArgs result.
func handleTaskUpdateParsed(server framework.MCPServer, parsed *mcpcli.Args) error {
	oldStatus := parsed.GetFlag("status", "")
	newStatus := parsed.GetFlag("new-status", "")
	newPriority := parsed.GetFlag("new-priority", "")
	autoApply := parsed.GetBoolFlag("auto-apply", false)
	idsStr := parsed.GetFlag("ids", "")

	var taskIDs []string
	if idsStr != "" {
		taskIDs = parseTaskIDs(idsStr)
	} else {
		// Collect positional args that look like task IDs
		for _, p := range parsed.Positional {
			if strings.HasPrefix(p, "T-") {
				taskIDs = append(taskIDs, p)
			}
		}
	}

	if len(taskIDs) == 0 && oldStatus == "" {
		return fmt.Errorf("task update requires task ID(s) or --status flag")
	}

	return handleTaskUpdateWithParams(server, taskIDs, oldStatus, newStatus, newPriority, autoApply)
}

// handleTaskUpdateWithParams executes the update with parsed params.
func handleTaskUpdateWithParams(server framework.MCPServer, taskIDs []string, oldStatus, newStatus, newPriority string, autoApply bool) error {
	if len(taskIDs) == 0 && oldStatus == "" {
		return fmt.Errorf("task update requires task ID(s) or --status flag")
	}

	if newStatus == "" && newPriority == "" {
		return fmt.Errorf("task update requires --new-status and/or --new-priority")
	}
	// Priority update: use action "update" with task_ids and priority
	if newPriority != "" {
		if len(taskIDs) == 0 {
			return fmt.Errorf("task update with --new-priority requires task ID(s) or --ids")
		}

		toolArgs := map[string]interface{}{
			"action":   "update",
			"task_ids": strings.Join(taskIDs, ","),
			"priority": newPriority,
		}
		if newStatus != "" {
			toolArgs["new_status"] = newStatus
		}

		return executeTaskWorkflow(server, toolArgs)
	}
	// Status update: use action "approve"
	toolArgs := map[string]interface{}{
		"action":     "approve",
		"new_status": newStatus,
		"auto_apply": autoApply,
	}
	if oldStatus != "" {
		toolArgs["status"] = oldStatus
	}

	if len(taskIDs) > 0 {
		toolArgs["task_ids"] = strings.Join(taskIDs, ",")
	}

	return executeTaskWorkflow(server, toolArgs)
}

// handleTaskUpdate handles "task update" command (legacy, used by tests).
func handleTaskUpdate(server framework.MCPServer, args []string) error {
	parsed := mcpcli.ParseArgs(append([]string{"task", "update"}, args...))
	return handleTaskUpdateParsed(server, parsed)
}

// handleTaskCreateParsed handles "task create" command using ParseArgs result.
func handleTaskCreateParsed(server framework.MCPServer, parsed *mcpcli.Args) error {
	name := strings.TrimSpace(strings.Join(parsed.Positional, " "))
	if name == "" {
		return fmt.Errorf("task create requires a task name")
	}

	description := parsed.GetFlag("description", "")
	priority := parsed.GetFlag("priority", "")
	tagsStr := parsed.GetFlag("tags", "")
	localAIBackend := parsed.GetFlag("local-ai-backend", "")

	var tags []string

	if tagsStr != "" {
		for _, t := range strings.Split(tagsStr, ",") {
			if trimmed := strings.TrimSpace(t); trimmed != "" {
				tags = append(tags, trimmed)
			}
		}
	}

	toolArgs := map[string]interface{}{
		"action":           "create",
		"name":             name,
		"long_description": description,
	}
	if priority != "" {
		toolArgs["priority"] = priority
	}

	if len(tags) > 0 {
		toolArgs["tags"] = strings.Join(tags, ",")
	}

	if localAIBackend != "" {
		toolArgs["local_ai_backend"] = localAIBackend
	}

	return executeTaskWorkflow(server, toolArgs)
}

// handleTaskCreate handles "task create" command (legacy, used by tests).
func handleTaskCreate(server framework.MCPServer, args []string) error {
	parsed := mcpcli.ParseArgs(append([]string{"task", "create"}, args...))
	return handleTaskCreateParsed(server, parsed)
}

// handleTaskShow handles "task show <task-id>" command.
func handleTaskShow(server framework.MCPServer, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("task show requires a task ID")
	}

	taskID := args[0]

	// Build task_workflow args to get full task details
	toolArgs := map[string]interface{}{
		"action":        "sync",
		"sub_action":    "list",
		"task_id":       taskID,
		"output_format": "text",
	}

	return executeTaskWorkflow(server, toolArgs)
}

// handleTaskDelete handles "task delete <task-id>" command.
func handleTaskDelete(server framework.MCPServer, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("task delete requires a task ID")
	}

	taskID := args[0]

	toolArgs := map[string]interface{}{
		"action":  "delete",
		"task_id": taskID,
	}

	return executeTaskWorkflow(server, toolArgs)
}

// handleTaskSync runs task_workflow action=sync (SQLite ↔ JSON).
func handleTaskSync(server framework.MCPServer) error {
	return executeTaskWorkflow(server, map[string]interface{}{"action": "sync"})
}

// executeTaskWorkflow executes the task_workflow tool with given arguments.
func executeTaskWorkflow(server framework.MCPServer, toolArgs map[string]interface{}) error {
	ctx := context.Background()

	// Convert to json.RawMessage
	argsBytes, err := json.Marshal(toolArgs)
	if err != nil {
		return fmt.Errorf("failed to marshal arguments: %w", err)
	}

	// Execute tool
	result, err := server.CallTool(ctx, "task_workflow", argsBytes)
	if err != nil {
		return fmt.Errorf("task operation failed: %w", err)
	}

	// Display results
	if len(result) == 0 {
		_, _ = fmt.Println("Task operation completed successfully (no output)")
		return nil
	}

	for _, content := range result {
		_, _ = fmt.Println(content.Text)
	}

	return nil
}

// parseTaskIDs parses a comma-separated list of task IDs.
func parseTaskIDs(idsStr string) []string {
	ids := strings.Split(idsStr, ",")
	result := make([]string, 0, len(ids))

	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id != "" {
			result = append(result, id)
		}
	}

	return result
}

// showTaskUsage displays usage information for task commands.
func showTaskUsage() error {
	_, _ = fmt.Println("Task Management Commands")
	_, _ = fmt.Println("========================")
	_, _ = fmt.Println()
	_, _ = fmt.Println("Usage:")
	_, _ = fmt.Println("  exarp-go task <command> [options]")
	_, _ = fmt.Println()
	_, _ = fmt.Println("Commands:")
	_, _ = fmt.Println("  list                    List tasks")
	_, _ = fmt.Println("  status <task-id>        Show task status")
	_, _ = fmt.Println("  update [options]         Update task status")
	_, _ = fmt.Println("  create <name> [options]  Create new task")
	_, _ = fmt.Println("  show <task-id>          Show full task details")
	_, _ = fmt.Println("  delete <task-id>        Delete a task (e.g. wrong project)")
	_, _ = fmt.Println("  sync                    Sync Todo2 (SQLite ↔ JSON)")
	_, _ = fmt.Println("  help                    Show this help")
	_, _ = fmt.Println()
	_, _ = fmt.Println("List Options:")
	_, _ = fmt.Println("  --status <status>       Filter by status (Todo, In Progress, Done, Review)")
	_, _ = fmt.Println("  --priority <priority>   Filter by priority (low, medium, high)")
	_, _ = fmt.Println("  --tag <tag>             Filter by tag")
	_, _ = fmt.Println("  --limit <number>        Limit number of results")
	_, _ = fmt.Println()
	_, _ = fmt.Println("Update Options:")
	_, _ = fmt.Println("  <task-id>               Task ID(s) to update (e.g., T-1 or T-1,T-2)")
	_, _ = fmt.Println("  --status <status>       Current status (for batch updates)")
	_, _ = fmt.Println("  --new-status <status>   New status")
	_, _ = fmt.Println("  --new-priority <pri>    New priority (low, medium, high); requires task ID(s)")
	_, _ = fmt.Println("  --ids <ids>             Comma-separated task IDs")
	_, _ = fmt.Println("  --auto-apply            Auto-apply changes without confirmation")
	_, _ = fmt.Println()
	_, _ = fmt.Println("Create Options:")
	_, _ = fmt.Println("  --description <text>           Task description")
	_, _ = fmt.Println("  --priority <priority>          Task priority (low, medium, high)")
	_, _ = fmt.Println("  --tags <tags>                 Comma-separated tags")
	_, _ = fmt.Println("  --local-ai-backend <backend>  Preferred local AI for estimation (fm|mlx|ollama)")
	_, _ = fmt.Println()
	_, _ = fmt.Println("Examples:")
	_, _ = fmt.Println("  exarp-go task list")
	_, _ = fmt.Println("  exarp-go task list --status \"In Progress\"")
	_, _ = fmt.Println("  exarp-go task status T-123")
	_, _ = fmt.Println("  exarp-go task update T-1 --new-status \"Done\"")
	_, _ = fmt.Println("  exarp-go task update T-1 --new-priority high")
	_, _ = fmt.Println("  exarp-go task update --status \"Todo\" --new-status \"Done\" --ids \"T-1,T-2\"")
	_, _ = fmt.Println("  exarp-go task create \"Fix bug\" --description \"Fix the bug\" --priority \"high\"")
	_, _ = fmt.Println("  exarp-go task create \"AI task\" --local-ai-backend ollama")
	_, _ = fmt.Println("  exarp-go task show T-123")
	_, _ = fmt.Println("  exarp-go task sync")
	_, _ = fmt.Println()

	return nil
}

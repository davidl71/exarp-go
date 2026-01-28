package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/davidl71/exarp-go/internal/framework"
)

// handleTaskCommand handles task subcommands (list, status, update, create, show)
func handleTaskCommand(server framework.MCPServer, args []string) error {
	if len(args) == 0 {
		return showTaskUsage()
	}

	cmd := args[0]
	switch cmd {
	case "list":
		return handleTaskList(server, args[1:])
	case "status":
		return handleTaskStatus(server, args[1:])
	case "update":
		return handleTaskUpdate(server, args[1:])
	case "create":
		return handleTaskCreate(server, args[1:])
	case "show":
		return handleTaskShow(server, args[1:])
	case "help":
		return showTaskUsage()
	default:
		return fmt.Errorf("unknown task command: %s (use: list, status, update, create, show, help)", cmd)
	}
}

// handleTaskList handles "task list" command
func handleTaskList(server framework.MCPServer, args []string) error {
	// Parse flags
	var status, priority, tag string
	var limit int

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--status" && i+1 < len(args):
			status = args[i+1]
			i++
		case arg == "--priority" && i+1 < len(args):
			priority = args[i+1]
			i++
		case arg == "--tag" && i+1 < len(args):
			tag = args[i+1]
			i++
		case arg == "--limit" && i+1 < len(args):
			fmt.Sscanf(args[i+1], "%d", &limit)
			i++
		case strings.HasPrefix(arg, "--status="):
			status = strings.TrimPrefix(arg, "--status=")
		case strings.HasPrefix(arg, "--priority="):
			priority = strings.TrimPrefix(arg, "--priority=")
		case strings.HasPrefix(arg, "--tag="):
			tag = strings.TrimPrefix(arg, "--tag=")
		case strings.HasPrefix(arg, "--limit="):
			fmt.Sscanf(strings.TrimPrefix(arg, "--limit="), "%d", &limit)
		}
	}

	// Build task_workflow args
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

	return executeTaskWorkflow(server, toolArgs)
}

// handleTaskStatus handles "task status <task-id>" command
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

// handleTaskUpdate handles "task update" command
func handleTaskUpdate(server framework.MCPServer, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("task update requires task ID(s) or --status flag")
	}

	var taskIDs []string
	var oldStatus, newStatus string
	var autoApply bool

	// Parse arguments
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--status" && i+1 < len(args):
			oldStatus = args[i+1]
			i++
		case arg == "--new-status" && i+1 < len(args):
			newStatus = args[i+1]
			i++
		case arg == "--ids" && i+1 < len(args):
			idsStr := args[i+1]
			taskIDs = parseTaskIDs(idsStr)
			i++
		case arg == "--auto-apply":
			autoApply = true
		case strings.HasPrefix(arg, "--status="):
			oldStatus = strings.TrimPrefix(arg, "--status=")
		case strings.HasPrefix(arg, "--new-status="):
			newStatus = strings.TrimPrefix(arg, "--new-status=")
		case strings.HasPrefix(arg, "--ids="):
			taskIDs = parseTaskIDs(strings.TrimPrefix(arg, "--ids="))
		case strings.HasPrefix(arg, "T-"):
			// Task ID as positional argument
			taskIDs = append(taskIDs, arg)
		default:
			// Try to parse as task ID
			if strings.HasPrefix(arg, "T-") {
				taskIDs = append(taskIDs, arg)
			}
		}
	}

	if len(taskIDs) == 0 && oldStatus == "" {
		return fmt.Errorf("task update requires task ID(s) or --status flag")
	}

	if newStatus == "" {
		return fmt.Errorf("task update requires --new-status flag")
	}

	// Build task_workflow args
	toolArgs := map[string]interface{}{
		"action":     "approve",
		"new_status": newStatus,
		"auto_apply": autoApply,
	}

	if oldStatus != "" {
		toolArgs["status"] = oldStatus
	}

	if len(taskIDs) > 0 {
		// Convert task IDs to JSON array string
		idsJSON, err := json.Marshal(taskIDs)
		if err != nil {
			return fmt.Errorf("failed to marshal task IDs: %w", err)
		}
		toolArgs["task_ids"] = string(idsJSON)
	}

	return executeTaskWorkflow(server, toolArgs)
}

// handleTaskCreate handles "task create" command
func handleTaskCreate(server framework.MCPServer, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("task create requires a task name")
	}

	name := args[0]
	var description, priority string
	var tags []string

	// Parse flags
	for i := 1; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--description" && i+1 < len(args):
			description = args[i+1]
			i++
		case arg == "--priority" && i+1 < len(args):
			priority = args[i+1]
			i++
		case arg == "--tags" && i+1 < len(args):
			tagsStr := args[i+1]
			tags = strings.Split(tagsStr, ",")
			for j := range tags {
				tags[j] = strings.TrimSpace(tags[j])
			}
			i++
		case strings.HasPrefix(arg, "--description="):
			description = strings.TrimPrefix(arg, "--description=")
		case strings.HasPrefix(arg, "--priority="):
			priority = strings.TrimPrefix(arg, "--priority=")
		case strings.HasPrefix(arg, "--tags="):
			tagsStr := strings.TrimPrefix(arg, "--tags=")
			tags = strings.Split(tagsStr, ",")
			for j := range tags {
				tags[j] = strings.TrimSpace(tags[j])
			}
		}
	}

	// Build task_workflow args
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

	return executeTaskWorkflow(server, toolArgs)
}

// handleTaskShow handles "task show <task-id>" command
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

// executeTaskWorkflow executes the task_workflow tool with given arguments
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

// parseTaskIDs parses a comma-separated list of task IDs
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

// showTaskUsage displays usage information for task commands
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
	_, _ = fmt.Println("  --new-status <status>   New status (required)")
	_, _ = fmt.Println("  --ids <ids>             Comma-separated task IDs")
	_, _ = fmt.Println("  --auto-apply            Auto-apply changes without confirmation")
	_, _ = fmt.Println()
	_, _ = fmt.Println("Create Options:")
	_, _ = fmt.Println("  --description <text>    Task description")
	_, _ = fmt.Println("  --priority <priority>   Task priority (low, medium, high)")
	_, _ = fmt.Println("  --tags <tags>          Comma-separated tags")
	_, _ = fmt.Println()
	_, _ = fmt.Println("Examples:")
	_, _ = fmt.Println("  exarp-go task list")
	_, _ = fmt.Println("  exarp-go task list --status \"In Progress\"")
	_, _ = fmt.Println("  exarp-go task status T-123")
	_, _ = fmt.Println("  exarp-go task update T-1 --new-status \"Done\"")
	_, _ = fmt.Println("  exarp-go task update --status \"Todo\" --new-status \"Done\" --ids \"T-1,T-2\"")
	_, _ = fmt.Println("  exarp-go task create \"Fix bug\" --description \"Fix the bug\" --priority \"high\"")
	_, _ = fmt.Println("  exarp-go task show T-123")
	_, _ = fmt.Println()
	return nil
}

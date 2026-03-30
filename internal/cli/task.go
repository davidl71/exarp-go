// task.go — CLI "task" subcommand: list, create, update, show, estimate, summarize.
package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/internal/taskworkflowspec"
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
	case "review":
		return handleTaskReview(server, parsed.Positional)
	case "delete":
		return handleTaskDelete(server, parsed.Positional)
	case "sync":
		return handleTaskSync(server)
	case "estimate":
		return handleTaskEstimateParsed(server, parsed)
	case "summarize":
		return handleTaskSummarizeParsed(server, parsed)
	case "run-with-ai", "run_with_ai":
		return handleTaskRunWithAIParsed(server, parsed)
	case "help":
		return showTaskUsage()
	default:
		return fmt.Errorf("unknown task command: %s (use: list, status, update, create, show, review, delete, sync, estimate, summarize, run-with-ai, help)", subcommand)
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
		"action": "list",
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
	if CLIOutputOpts.JSON {
		toolArgs["output_format"] = "json"
	}

	return executeTaskWorkflow(server, toolArgs)
}

// handleTaskStatus handles "task status <task-id>" command.
func handleTaskStatus(server framework.MCPServer, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("task status requires a task ID")
	}

	task, err := loadSingleTask(server, strings.TrimSpace(args[0]))
	if err != nil {
		return err
	}

	return printTaskStatus(task)
}

// handleTaskUpdateParsed handles "task update" command using ParseArgs result.
func handleTaskUpdateParsed(server framework.MCPServer, parsed *mcpcli.Args) error {
	oldStatus := parsed.GetFlag("status", "")
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

	input := taskworkflowspec.TaskUpdateInput{TaskIDs: taskIDs}
	if parsed.HasFlag("new-status") {
		input.NewStatus = taskworkflowspec.OptionalString{Set: true, Value: parsed.GetFlag("new-status", "")}
	}
	if parsed.HasFlag("new-priority") {
		input.Priority = taskworkflowspec.OptionalString{Set: true, Value: parsed.GetFlag("new-priority", "")}
	}
	if parsed.HasFlag("tags") {
		input.Tags = taskworkflowspec.OptionalList{Set: true, Values: taskworkflowspec.CSVToList(parsed.GetFlag("tags", ""))}
	}
	if parsed.HasFlag("remove-tags") {
		input.RemoveTags = taskworkflowspec.OptionalList{Set: true, Values: taskworkflowspec.CSVToList(parsed.GetFlag("remove-tags", ""))}
	}
	if parsed.HasFlag("name") {
		input.Name = taskworkflowspec.OptionalString{Set: true, Value: parsed.GetFlag("name", "")}
	}
	if parsed.HasFlag("description") {
		input.LongDescription = taskworkflowspec.OptionalString{Set: true, Value: parsed.GetFlag("description", "")}
	}
	if parsed.HasFlag("dependencies") {
		input.Dependencies = taskworkflowspec.OptionalList{Set: true, Values: taskworkflowspec.CSVToList(parsed.GetFlag("dependencies", ""))}
	}
	if parsed.HasFlag("recommended-tools") {
		input.RecommendedTools = taskworkflowspec.OptionalList{Set: true, Values: taskworkflowspec.CSVToList(parsed.GetFlag("recommended-tools", ""))}
	}
	if parsed.HasFlag("local-ai-backend") {
		input.LocalAIBackend = taskworkflowspec.OptionalString{Set: true, Value: parsed.GetFlag("local-ai-backend", "")}
	}
	if parsed.HasFlag("parent-id") {
		input.ParentID = taskworkflowspec.OptionalString{Set: true, Value: parsed.GetFlag("parent-id", "")}
	}

	return handleTaskUpdateWithParams(server, oldStatus, input, autoApply)
}

// handleTaskUpdateWithParams executes the update with parsed params.
func handleTaskUpdateWithParams(server framework.MCPServer, oldStatus string, input taskworkflowspec.TaskUpdateInput, autoApply bool) error {
	if len(input.TaskIDs) == 0 && oldStatus == "" {
		return fmt.Errorf("task update requires task ID(s) or --status flag")
	}

	if !input.NewStatus.Set && !input.Priority.Set && !input.Tags.Set && !input.RemoveTags.Set && !input.Name.Set &&
		!input.LongDescription.Set && !input.Dependencies.Set && !input.RecommendedTools.Set &&
		!input.LocalAIBackend.Set && !input.ParentID.Set {
		return fmt.Errorf("task update requires at least one task field such as --new-status, --new-priority, --dependencies, --parent-id, --tags, --remove-tags, --name, --description, --recommended-tools, or --local-ai-backend")
	}

	if len(input.TaskIDs) > 0 {
		return executeTaskWorkflow(server, input.ToToolArgs())
	}
	// Status-only batch: use action "approve" (filter by old status, no task_ids required)
	toolArgs := map[string]interface{}{
		"action":     "approve",
		"new_status": input.NewStatus.Value,
		"auto_apply": autoApply,
	}
	if oldStatus != "" {
		toolArgs["status"] = oldStatus
	}
	if len(input.TaskIDs) > 0 {
		toolArgs["task_ids"] = strings.Join(input.TaskIDs, ",")
	}

	return executeTaskWorkflow(server, toolArgs)
}

// handleTaskCreateParsed handles "task create" command using ParseArgs result.
func handleTaskCreateParsed(server framework.MCPServer, parsed *mcpcli.Args) error {
	name := strings.TrimSpace(strings.Join(parsed.Positional, " "))
	if name == "" {
		return fmt.Errorf("task create requires a task name")
	}

	description := parsed.GetFlag("description", "")
	input := taskworkflowspec.TaskCreateInput{
		Name:            name,
		LongDescription: taskworkflowspec.OptionalString{Set: true, Value: description},
	}
	if parsed.HasFlag("priority") {
		input.Priority = taskworkflowspec.OptionalString{Set: true, Value: parsed.GetFlag("priority", "")}
	}
	if parsed.HasFlag("tags") {
		input.Tags = taskworkflowspec.OptionalList{Set: true, Values: taskworkflowspec.CSVToList(parsed.GetFlag("tags", ""))}
	}
	if parsed.HasFlag("dependencies") {
		input.Dependencies = taskworkflowspec.OptionalList{Set: true, Values: taskworkflowspec.CSVToList(parsed.GetFlag("dependencies", ""))}
	}
	if parsed.HasFlag("local-ai-backend") {
		input.LocalAIBackend = taskworkflowspec.OptionalString{Set: true, Value: parsed.GetFlag("local-ai-backend", "")}
	}
	if parsed.HasFlag("recommended-tools") {
		input.RecommendedTools = taskworkflowspec.OptionalList{Set: true, Values: taskworkflowspec.CSVToList(parsed.GetFlag("recommended-tools", ""))}
	}
	if parsed.HasFlag("planning-doc") {
		input.PlanningDoc = taskworkflowspec.OptionalString{Set: true, Value: parsed.GetFlag("planning-doc", "")}
	}
	if parsed.HasFlag("epic-id") {
		input.EpicID = taskworkflowspec.OptionalString{Set: true, Value: parsed.GetFlag("epic-id", "")}
	}
	if parsed.HasFlag("parent-id") {
		input.ParentID = taskworkflowspec.OptionalString{Set: true, Value: parsed.GetFlag("parent-id", "")}
	}

	return executeTaskWorkflow(server, input.ToToolArgs())
}

// handleTaskShow handles "task show <task-id>" command.
func handleTaskShow(server framework.MCPServer, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("task show requires a task ID")
	}

	task, err := loadSingleTask(server, strings.TrimSpace(args[0]))
	if err != nil {
		return err
	}

	return printTaskDetails(task)
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

// handleTaskEstimateParsed handles "task estimate <name>" using the estimation tool.
func handleTaskEstimateParsed(server framework.MCPServer, parsed *mcpcli.Args) error {
	name := strings.TrimSpace(strings.Join(parsed.Positional, " "))
	if name == "" {
		return fmt.Errorf("task estimate requires a task name")
	}
	toolArgs := map[string]interface{}{
		"action": "estimate",
		"name":   name,
	}
	if b := parsed.GetFlag("local-ai-backend", ""); b != "" {
		toolArgs["local_ai_backend"] = b
	}
	if d := parsed.GetFlag("details", ""); d != "" {
		toolArgs["details"] = d
	}
	if p := parsed.GetFlag("priority", ""); p != "" {
		toolArgs["priority"] = p
	}
	if t := parsed.GetFlag("tags", ""); t != "" {
		toolArgs["tags"] = t
	}
	return executeEstimation(server, toolArgs)
}

// executeEstimation calls the estimation tool and prints the result.
func executeEstimation(server framework.MCPServer, toolArgs map[string]interface{}) error {
	ctx := context.Background()
	argsBytes, err := json.Marshal(toolArgs)
	if err != nil {
		return fmt.Errorf("failed to marshal arguments: %w", err)
	}
	result, err := server.CallTool(ctx, "estimation", argsBytes)
	if err != nil {
		return fmt.Errorf("estimation failed: %w", err)
	}
	if len(result) == 0 {
		_, _ = fmt.Println("Estimation completed (no output)")
		return nil
	}
	for _, content := range result {
		_, _ = fmt.Println(content.Text)
	}
	return nil
}

// handleTaskSummarizeParsed handles "task summarize <task-id> [--local-ai-backend]".
func handleTaskSummarizeParsed(server framework.MCPServer, parsed *mcpcli.Args) error {
	taskID := ""
	if len(parsed.Positional) > 0 {
		taskID = strings.TrimSpace(parsed.Positional[0])
	}
	if taskID == "" {
		return fmt.Errorf("task summarize requires a task ID")
	}
	toolArgs := map[string]interface{}{
		"action":  "summarize",
		"task_id": taskID,
	}
	if b := parsed.GetFlag("local-ai-backend", ""); b != "" {
		toolArgs["local_ai_backend"] = b
	}
	return executeTaskWorkflow(server, toolArgs)
}

// handleTaskRunWithAIParsed handles "task run-with-ai <task-id> [--backend] [--instruction]".
func handleTaskRunWithAIParsed(server framework.MCPServer, parsed *mcpcli.Args) error {
	taskID := ""
	if len(parsed.Positional) > 0 {
		taskID = strings.TrimSpace(parsed.Positional[0])
	}
	if taskID == "" {
		return fmt.Errorf("task run-with-ai requires a task ID")
	}
	toolArgs := map[string]interface{}{
		"action":  "run_with_ai",
		"task_id": taskID,
	}
	backend := parsed.GetFlag("backend", "")
	if backend == "" {
		backend = parsed.GetFlag("local-ai-backend", "")
	}
	if backend != "" {
		toolArgs["local_ai_backend"] = backend
	}
	if inst := parsed.GetFlag("instruction", ""); inst != "" {
		toolArgs["instruction"] = inst
	}
	return executeTaskWorkflow(server, toolArgs)
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
		if !CLIOutputOpts.Quiet {
			_, _ = fmt.Println("Task operation completed successfully (no output)")
		}
		return nil
	}

	for _, content := range result {
		text := content.Text
		if CLIOutputOpts.Concise {
			text = ConciseOutput(text)
		}
		if CLIOutputOpts.JSON {
			text = compactJSONIfValid(text)
		}
		if text != "" {
			_, _ = fmt.Println(text)
		}
	}

	return nil
}

func loadSingleTask(server framework.MCPServer, taskID string) (map[string]interface{}, error) {
	if taskID == "" {
		return nil, fmt.Errorf("task ID cannot be empty")
	}

	ctx := context.Background()
	toolArgs := map[string]interface{}{
		"action":        "list",
		"task_id":       taskID,
		"output_format": "json",
		"compact":       true,
	}
	argsBytes, err := json.Marshal(toolArgs)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal task query arguments: %w", err)
	}

	result, err := server.CallTool(ctx, "task_workflow", argsBytes)
	if err != nil {
		return nil, fmt.Errorf("task operation failed: %w", err)
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("task %s not found", taskID)
	}

	var response struct {
		Tasks []map[string]interface{} `json:"tasks"`
	}
	if err := json.Unmarshal([]byte(result[0].Text), &response); err != nil {
		return nil, fmt.Errorf("failed to parse task response: %w", err)
	}
	if len(response.Tasks) == 0 {
		return nil, fmt.Errorf("task %s not found", taskID)
	}

	return response.Tasks[0], nil
}

func printTaskStatus(task map[string]interface{}) error {
	id := taskString(task["id"])
	status := taskString(task["status"])
	if CLIOutputOpts.JSON {
		out := map[string]interface{}{
			"success": true,
			"method":  "status",
			"task": map[string]interface{}{
				"id":      id,
				"status":  status,
				"content": taskString(task["content"]),
			},
		}
		if priority := taskString(task["priority"]); priority != "" {
			outTask := out["task"].(map[string]interface{})
			outTask["priority"] = priority
		}
		return printTaskJSON(out)
	}

	if id != "" {
		_, _ = fmt.Fprintf(os.Stdout, "Task: %s\n", id)
	}
	_, _ = fmt.Fprintf(os.Stdout, "Status: %s\n", status)
	if priority := taskString(task["priority"]); priority != "" {
		_, _ = fmt.Fprintf(os.Stdout, "Priority: %s\n", priority)
	}
	if content := taskString(task["content"]); content != "" {
		_, _ = fmt.Fprintf(os.Stdout, "Content: %s\n", content)
	}
	return nil
}

func printTaskDetails(task map[string]interface{}) error {
	if CLIOutputOpts.JSON {
		return printTaskJSON(map[string]interface{}{
			"success": true,
			"method":  "show",
			"task":    task,
		})
	}

	writeTaskField("ID", taskString(task["id"]))
	writeTaskField("Status", taskString(task["status"]))
	writeTaskField("Priority", taskString(task["priority"]))
	writeTaskField("Content", taskString(task["content"]))
	writeTaskField("Description", taskString(task["long_description"]))
	writeTaskField("Parent ID", taskString(task["parent_id"]))
	writeTaskField("Created", taskString(task["created_at"]))
	writeTaskField("Completed", taskString(task["completed_at"]))
	writeTaskField("Last Modified", taskString(task["last_modified"]))
	writeTaskField("Tags", joinTaskStrings(task["tags"]))
	writeTaskField("Dependencies", joinTaskStrings(task["dependencies"]))
	writeTaskField("Recommended Tools", joinTaskStrings(task["recommended_tools"]))
	if claim, ok := task["active_claim"].(map[string]interface{}); ok {
		writeTaskField("Active Claim", fmt.Sprintf("%s until %s", taskString(claim["assignee"]), taskString(claim["lock_until"])))
	}
	writeTaskField("Recent Runs", formatCount(task["recent_runs"]))
	writeTaskField("Recent Verifications", formatCount(task["recent_verifications"]))
	writeTaskField("Recent Progress", formatCount(task["recent_progress"]))
	return nil
}

func printTaskJSON(v map[string]interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("failed to marshal task output: %w", err)
	}
	_, _ = fmt.Fprintln(os.Stdout, string(data))
	return nil
}

func enrichTaskMap(task map[string]interface{}) map[string]interface{} {
	if task == nil {
		return nil
	}
	taskID := taskString(task["id"])
	if taskID == "" {
		return task
	}
	if activeLocks, err := database.GetActiveLockMapForTasks(context.Background(), []string{taskID}); err == nil {
		if lock, ok := activeLocks[taskID]; ok {
			task["active_claim"] = map[string]interface{}{
				"assignee":    lock.Assignee,
				"assigned_at": lock.AssignedAt.Format(time.RFC3339),
				"lock_until":  lock.LockUntil.Format(time.RFC3339),
				"status":      "active",
			}
		}
	}
	if runs, err := database.ListTaskExecutionRuns(context.Background(), taskID, "", 5); err == nil && len(runs) > 0 {
		items := make([]map[string]interface{}, 0, len(runs))
		for i := range runs {
			items = append(items, map[string]interface{}{
				"run_id":     runs[i].RunID,
				"status":     runs[i].Status,
				"summary":    runs[i].Summary,
				"started_at": runs[i].StartedAt.Format(time.RFC3339),
			})
		}
		task["recent_runs"] = items
	}
	if verifications, err := database.ListTaskVerifications(context.Background(), taskID, "", 5); err == nil && len(verifications) > 0 {
		items := make([]map[string]interface{}, 0, len(verifications))
		for i := range verifications {
			items = append(items, map[string]interface{}{
				"verification_id": verifications[i].VerificationID,
				"kind":            verifications[i].Kind,
				"result":          verifications[i].Result,
				"created_at":      verifications[i].CreatedAt.Format(time.RFC3339),
			})
		}
		task["recent_verifications"] = items
	}
	if progressEntries, err := database.ListTaskProgressEntries(context.Background(), taskID, "", 5); err == nil && len(progressEntries) > 0 {
		items := make([]map[string]interface{}, 0, len(progressEntries))
		for i := range progressEntries {
			items = append(items, map[string]interface{}{
				"progress_id":    progressEntries[i].ProgressID,
				"summary":        progressEntries[i].Summary,
				"remaining_work": progressEntries[i].RemainingWork,
				"created_at":     progressEntries[i].CreatedAt.Format(time.RFC3339),
			})
		}
		task["recent_progress"] = items
	}
	return task
}

func writeTaskField(name, value string) {
	if value == "" {
		return
	}
	_, _ = fmt.Fprintf(os.Stdout, "%s: %s\n", name, value)
}

func taskString(v interface{}) string {
	s, _ := v.(string)
	return strings.TrimSpace(s)
}

func joinTaskStrings(v interface{}) string {
	switch x := v.(type) {
	case []string:
		return strings.Join(x, ", ")
	case []interface{}:
		out := make([]string, 0, len(x))
		for _, item := range x {
			if s := taskString(item); s != "" {
				out = append(out, s)
			}
		}
		return strings.Join(out, ", ")
	default:
		return ""
	}
}

func formatCount(v interface{}) string {
	switch x := v.(type) {
	case []interface{}:
		if len(x) == 0 {
			return ""
		}
		return fmt.Sprintf("%d", len(x))
	default:
		return ""
	}
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
	_, _ = fmt.Println("  update [options]        Update task status")
	_, _ = fmt.Println("  create <name> [options]  Create new task")
	_, _ = fmt.Println("  show <task-id>          Show full task details")
	_, _ = fmt.Println("  review <task-id>        Open local review UI for execution-pack")
	_, _ = fmt.Println("  delete <task-id>        Delete a task (e.g. wrong project)")
	_, _ = fmt.Println("  sync                    Sync Todo2 (SQLite ↔ JSON)")
	_, _ = fmt.Println("  estimate <name>         Estimate task duration (local AI)")
	_, _ = fmt.Println("  summarize <task-id>     Generate AI summary for task")
	_, _ = fmt.Println("  run-with-ai <task-id>   Run task through local LLM (implementation guidance)")
	_, _ = fmt.Println("  help                    Show this help")
	_, _ = fmt.Println()
	_, _ = fmt.Println("List Options:")
	_, _ = fmt.Println("  --status <status>       Filter by status (Todo, In Progress, Done, Review)")
	_, _ = fmt.Println("  --priority <priority>   Filter by priority (low, medium, high)")
	_, _ = fmt.Println("  --tag <tag>             Filter by tag")
	_, _ = fmt.Println("  --limit <number>        Limit number of results")
	_, _ = fmt.Println("  --quiet                 Suppress verbose output (OpenCode/script-friendly)")
	_, _ = fmt.Println("  --json, -j              Machine-readable JSON output")
	_, _ = fmt.Println("  --concise               Strip emojis and decorative lines")
	_, _ = fmt.Println()
	_, _ = fmt.Println("Update Options:")
	_, _ = fmt.Println("  <task-id>               Task ID(s) to update (e.g., T-1 or T-1,T-2)")
	_, _ = fmt.Println("  --status <status>       Current status (for batch updates)")
	_, _ = fmt.Println("  --new-status <status>   New status")
	_, _ = fmt.Println("  --new-priority <pri>    New priority (low, medium, high); requires task ID(s)")
	_, _ = fmt.Println("  --dependencies <ids>    Comma-separated dependency task IDs; replaces dependencies")
	_, _ = fmt.Println("  --parent-id <task-id>   Set parent task ID for hierarchy")
	_, _ = fmt.Println("  --tags <tags>           Comma-separated tags to add")
	_, _ = fmt.Println("  --remove-tags <tags>    Comma-separated tags to remove")
	_, _ = fmt.Println("  --name <text>           Replace task title/content")
	_, _ = fmt.Println("  --description <text>    Replace task description")
	_, _ = fmt.Println("  --recommended-tools <list>  Comma-separated MCP tool IDs; requires task ID(s)")
	_, _ = fmt.Println("  --local-ai-backend <backend>  Set task preferred backend (fm|mlx|ollama); requires task ID(s)")
	_, _ = fmt.Println("  --ids <ids>             Comma-separated task IDs")
	_, _ = fmt.Println("  --auto-apply            Auto-apply changes without confirmation")
	_, _ = fmt.Println()
	_, _ = fmt.Println("Create Options:")
	_, _ = fmt.Println("  --description <text>           Task description")
	_, _ = fmt.Println("  --priority <priority>          Task priority (low, medium, high)")
	_, _ = fmt.Println("  --tags <tags>                  Comma-separated tags")
	_, _ = fmt.Println("  --dependencies <ids>           Comma-separated dependency task IDs")
	_, _ = fmt.Println("  --parent-id <task-id>          Parent task ID for hierarchy")
	_, _ = fmt.Println("  --epic-id <task-id>            Epic task ID")
	_, _ = fmt.Println("  --planning-doc <path>          Linked planning document path")
	_, _ = fmt.Println("  --local-ai-backend <backend>   Preferred local AI (fm|mlx|ollama)")
	_, _ = fmt.Println("  --recommended-tools <list>     Comma-separated MCP tool IDs (e.g. report,task_workflow)")
	_, _ = fmt.Println()
	_, _ = fmt.Println("Estimate Options:")
	_, _ = fmt.Println("  --local-ai-backend <backend>   Backend for estimation (fm|mlx|ollama)")
	_, _ = fmt.Println("  --details <text>               Optional task details")
	_, _ = fmt.Println("  --priority <priority>          Priority (low, medium, high)")
	_, _ = fmt.Println("  --tags <tags>                  Comma-separated tags")
	_, _ = fmt.Println()
	_, _ = fmt.Println("Summarize Options:")
	_, _ = fmt.Println("  --local-ai-backend <backend>   Override task preferred backend (fm|mlx|ollama)")
	_, _ = fmt.Println()
	_, _ = fmt.Println("Run-with-AI Options:")
	_, _ = fmt.Println("  --backend <backend>            Local AI backend (fm|mlx|ollama); alias: --local-ai-backend")
	_, _ = fmt.Println("  --instruction <text>           Extra instruction for the model")
	_, _ = fmt.Println()
	_, _ = fmt.Println("Examples:")
	_, _ = fmt.Println("  exarp-go task list")
	_, _ = fmt.Println("  exarp-go task list --status \"In Progress\"")
	_, _ = fmt.Println("  exarp-go task status T-123")
	_, _ = fmt.Println("  exarp-go task update T-1 --new-status \"Done\"")
	_, _ = fmt.Println("  exarp-go task update T-1 --new-priority high")
	_, _ = fmt.Println("  exarp-go task update T-2 --dependencies \"T-1,T-3\" --parent-id T-10")
	_, _ = fmt.Println("  exarp-go task update --status \"Todo\" --new-status \"Done\" --ids \"T-1,T-2\"")
	_, _ = fmt.Println("  exarp-go task create \"Fix bug\" --description \"Fix the bug\" --priority \"high\"")
	_, _ = fmt.Println("  exarp-go task create \"Subtask\" --parent-id T-10 --dependencies \"T-1\" --planning-doc docs/plan.md")
	_, _ = fmt.Println("  exarp-go task create \"AI task\" --local-ai-backend ollama --recommended-tools report")
	_, _ = fmt.Println("  exarp-go task estimate \"Add tests\" --local-ai-backend fm")
	_, _ = fmt.Println("  exarp-go task summarize T-123")
	_, _ = fmt.Println("  exarp-go task run-with-ai T-123 --backend ollama")
	_, _ = fmt.Println("  exarp-go task show T-123")
	_, _ = fmt.Println("  exarp-go task sync")
	_, _ = fmt.Println()

	return nil
}

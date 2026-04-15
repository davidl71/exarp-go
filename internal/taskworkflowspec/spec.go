package taskworkflowspec

import (
	"strings"
)

type FieldSpec struct {
	CanonicalName string
	MCPName       string
	CLIFlag       string
	Description   string
	Schema        map[string]interface{}
}

var CreateFieldSpecs = []FieldSpec{
	{
		CanonicalName: "long_description",
		MCPName:       "long_description",
		CLIFlag:       "description",
		Description:   "Task description",
		Schema: map[string]interface{}{
			"type":        "string",
			"description": "Task description (for single create; omit when using tasks array)",
		},
	},
	{
		CanonicalName: "priority",
		MCPName:       "priority",
		CLIFlag:       "priority",
		Description:   "Task priority (high, medium, low)",
		Schema: map[string]interface{}{
			"type":        "string",
			"description": "For create/update: task priority (high|medium|low).",
			"enum":        []string{"high", "medium", "low"},
		},
	},
	{
		CanonicalName: "tags",
		MCPName:       "tags",
		CLIFlag:       "tags",
		Description:   "Comma-separated tags",
		Schema: map[string]interface{}{
			"type":        "string",
			"description": "Task tags as comma-separated values (e.g. 'backend,urgent') or JSON array encoded as string (e.g. '[\"backend\",\"urgent\"]')",
		},
	},
	{
		CanonicalName: "dependencies",
		MCPName:       "dependencies",
		CLIFlag:       "dependencies",
		Description:   "Comma-separated dependency task IDs",
		Schema: map[string]interface{}{
			"type":        "string",
			"description": "Task dependencies as comma-separated task IDs or JSON array encoded as string (e.g. '[\"T-1\",\"T-2\"]')",
		},
	},
	{
		CanonicalName: "local_ai_backend",
		MCPName:       "local_ai_backend",
		CLIFlag:       "local-ai-backend",
		Description:   "Preferred local AI backend (fm, mlx, ollama)",
		Schema: map[string]interface{}{
			"type":        "string",
			"description": "For create/update: preferred local LLM for estimation (fm|mlx|ollama). Stored in task metadata as preferred_backend. For summarize/run_with_ai: overrides task metadata to select backend.",
			"enum":        []string{"", "fm", "mlx", "ollama"},
		},
	},
	{
		CanonicalName: "recommended_tools",
		MCPName:       "recommended_tools",
		CLIFlag:       "recommended-tools",
		Description:   "Comma-separated MCP tool IDs",
		Schema: map[string]interface{}{
			"type":        "string",
			"description": "For create/update: comma-separated MCP tool IDs to suggest for this task (e.g. report, task_workflow). Stored in task metadata as recommended_tools; exposed in task show and session prime suggested_next.",
		},
	},
	{
		CanonicalName: "planning_doc",
		MCPName:       "planning_doc",
		CLIFlag:       "planning-doc",
		Description:   "Path to linked planning document",
		Schema: map[string]interface{}{
			"type":        "string",
			"description": "Path to planning document. For link_planning: optional, stored in task metadata. For sync_from_plan/sync_plan_status: required (.plan.md path).",
		},
	},
	{
		CanonicalName: "epic_id",
		MCPName:       "epic_id",
		CLIFlag:       "epic-id",
		Description:   "Epic task ID",
		Schema: map[string]interface{}{
			"type":        "string",
			"description": "Epic task ID if this task is part of an epic (optional, stored in task metadata and parent_id)",
		},
	},
	{
		CanonicalName: "parent_id",
		MCPName:       "parent_id",
		CLIFlag:       "parent-id",
		Description:   "Parent task ID",
		Schema: map[string]interface{}{
			"type":        "string",
			"description": "Parent task ID for hierarchy (optional; separate from blocking dependencies). For create/update/link_planning.",
		},
	},
}

var UpdateFieldSpecs = []FieldSpec{
	{
		CanonicalName: "new_status",
		MCPName:       "new_status",
		CLIFlag:       "new-status",
		Description:   "New task status",
		Schema: map[string]interface{}{
			"type":    "string",
			"default": "Todo",
		},
	},
	{
		CanonicalName: "priority",
		MCPName:       "priority",
		CLIFlag:       "new-priority",
		Description:   "New task priority (high, medium, low)",
		Schema: map[string]interface{}{
			"type":        "string",
			"description": "For create/update: task priority (high|medium|low).",
			"enum":        []string{"high", "medium", "low"},
		},
	},
	{
		CanonicalName: "tags",
		MCPName:       "tags",
		CLIFlag:       "tags",
		Description:   "Comma-separated tags to add",
		Schema: map[string]interface{}{
			"type":        "string",
			"description": "Task tags as comma-separated values (e.g. 'backend,urgent') or JSON array encoded as string (e.g. '[\"backend\",\"urgent\"]')",
		},
	},
	{
		CanonicalName: "remove_tags",
		MCPName:       "remove_tags",
		CLIFlag:       "remove-tags",
		Description:   "Comma-separated tags to remove",
		Schema: map[string]interface{}{
			"type":        "string",
			"description": "Tags to remove from task(s). For action=update: comma-separated values or JSON array encoded as string.",
		},
	},
	{
		CanonicalName: "name",
		MCPName:       "name",
		CLIFlag:       "name",
		Description:   "Replacement task name",
		Schema: map[string]interface{}{
			"type":        "string",
			"description": "Task name (required for single create; omit when using tasks array)",
		},
	},
	{
		CanonicalName: "long_description",
		MCPName:       "long_description",
		CLIFlag:       "description",
		Description:   "Replacement task description",
		Schema: map[string]interface{}{
			"type":        "string",
			"description": "Task description (for single create; omit when using tasks array)",
		},
	},
	{
		CanonicalName: "dependencies",
		MCPName:       "dependencies",
		CLIFlag:       "dependencies",
		Description:   "Comma-separated dependency task IDs",
		Schema: map[string]interface{}{
			"type":        "string",
			"description": "Task dependencies as comma-separated task IDs or JSON array encoded as string (e.g. '[\"T-1\",\"T-2\"]')",
		},
	},
	{
		CanonicalName: "local_ai_backend",
		MCPName:       "local_ai_backend",
		CLIFlag:       "local-ai-backend",
		Description:   "Preferred local AI backend (fm, mlx, ollama)",
		Schema: map[string]interface{}{
			"type":        "string",
			"description": "For create/update: preferred local LLM for estimation (fm|mlx|ollama). Stored in task metadata as preferred_backend. For summarize/run_with_ai: overrides task metadata to select backend.",
			"enum":        []string{"", "fm", "mlx", "ollama"},
		},
	},
	{
		CanonicalName: "recommended_tools",
		MCPName:       "recommended_tools",
		CLIFlag:       "recommended-tools",
		Description:   "Comma-separated MCP tool IDs",
		Schema: map[string]interface{}{
			"type":        "string",
			"description": "For create/update: comma-separated MCP tool IDs to suggest for this task (e.g. report, task_workflow). Stored in task metadata as recommended_tools; exposed in task show and session prime suggested_next.",
		},
	},
	{
		CanonicalName: "parent_id",
		MCPName:       "parent_id",
		CLIFlag:       "parent-id",
		Description:   "Parent task ID",
		Schema: map[string]interface{}{
			"type":        "string",
			"description": "Parent task ID for hierarchy (optional; separate from blocking dependencies). For create/update/link_planning.",
		},
	},
}

type OptionalString struct {
	Set   bool
	Value string
}

type OptionalList struct {
	Set    bool
	Values []string
}

type TaskCreateInput struct {
	Name             string
	LongDescription  OptionalString
	Priority         OptionalString
	Tags             OptionalList
	Dependencies     OptionalList
	LocalAIBackend   OptionalString
	RecommendedTools OptionalList
	PlanningDoc      OptionalString
	EpicID           OptionalString
	ParentID         OptionalString
}

type TaskUpdateInput struct {
	TaskIDs          []string
	NewStatus        OptionalString
	Priority         OptionalString
	Tags             OptionalList
	RemoveTags       OptionalList
	Name             OptionalString
	LongDescription  OptionalString
	Dependencies     OptionalList
	LocalAIBackend   OptionalString
	RecommendedTools OptionalList
	ParentID         OptionalString
}

func cloneSchemaMap(src map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(src))
	for k, v := range src {
		out[k] = v
	}
	return out
}

func AppendTaskFieldSchemaProperties(props map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(props)+len(CreateFieldSpecs)+len(UpdateFieldSpecs))
	for k, v := range props {
		out[k] = v
	}

	seen := map[string]bool{}
	for _, spec := range append(CreateFieldSpecs, UpdateFieldSpecs...) {
		if seen[spec.MCPName] {
			continue
		}
		seen[spec.MCPName] = true
		out[spec.MCPName] = cloneSchemaMap(spec.Schema)
	}
	return out
}

func CSVToList(value string) []string {
	if strings.TrimSpace(value) == "" {
		return []string{}
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func ListToCSV(values []string) string {
	return strings.Join(values, ",")
}

func (in TaskCreateInput) ToToolArgs() map[string]interface{} {
	toolArgs := map[string]interface{}{
		"action": "create",
		"name":   in.Name,
	}
	if in.LongDescription.Set {
		toolArgs["long_description"] = in.LongDescription.Value
	}
	if in.Priority.Set {
		toolArgs["priority"] = in.Priority.Value
	}
	if in.Tags.Set {
		toolArgs["tags"] = ListToCSV(in.Tags.Values)
	}
	if in.Dependencies.Set {
		toolArgs["dependencies"] = ListToCSV(in.Dependencies.Values)
	}
	if in.LocalAIBackend.Set {
		toolArgs["local_ai_backend"] = in.LocalAIBackend.Value
	}
	if in.RecommendedTools.Set {
		toolArgs["recommended_tools"] = ListToCSV(in.RecommendedTools.Values)
	}
	if in.PlanningDoc.Set {
		toolArgs["planning_doc"] = in.PlanningDoc.Value
	}
	if in.EpicID.Set {
		toolArgs["epic_id"] = in.EpicID.Value
	}
	if in.ParentID.Set {
		toolArgs["parent_id"] = in.ParentID.Value
	}
	return toolArgs
}

func (in TaskUpdateInput) ToToolArgs() map[string]interface{} {
	toolArgs := map[string]interface{}{
		"action":   "update",
		"task_ids": ListToCSV(in.TaskIDs),
	}
	if in.NewStatus.Set {
		toolArgs["new_status"] = in.NewStatus.Value
	}
	if in.Priority.Set {
		toolArgs["priority"] = in.Priority.Value
	}
	if in.Tags.Set {
		toolArgs["tags"] = ListToCSV(in.Tags.Values)
	}
	if in.RemoveTags.Set {
		toolArgs["remove_tags"] = ListToCSV(in.RemoveTags.Values)
	}
	if in.Name.Set {
		toolArgs["name"] = in.Name.Value
	}
	if in.LongDescription.Set {
		toolArgs["long_description"] = in.LongDescription.Value
	}
	if in.Dependencies.Set {
		toolArgs["dependencies"] = ListToCSV(in.Dependencies.Values)
	}
	if in.LocalAIBackend.Set {
		toolArgs["local_ai_backend"] = in.LocalAIBackend.Value
	}
	if in.RecommendedTools.Set {
		toolArgs["recommended_tools"] = ListToCSV(in.RecommendedTools.Values)
	}
	if in.ParentID.Set {
		toolArgs["parent_id"] = in.ParentID.Value
	}
	return toolArgs
}

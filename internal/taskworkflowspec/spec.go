// Package taskworkflowspec provides task workflow specification types.
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
		CanonicalName: "priority_rank",
		MCPName:       "priority_rank",
		CLIFlag:       "priority-rank",
		Description:   "Numeric sort key within the same priority (lower first)",
		Schema: map[string]interface{}{
			"type":        "integer",
			"default":     0,
			"description": "For create/update: ordering within priority band (list, claim, backlog).",
		},
	},
	{
		CanonicalName: "tags",
		MCPName:       "tags",
		CLIFlag:       "tags",
		Description:   "Comma-separated tags, JSON array string, or repeated --tag on CLI",
		Schema: map[string]interface{}{
			"type":        "string",
			"description": "Task tags: comma-separated (e.g. 'backend,urgent'), a JSON array string (e.g. '[\"backend\",\"urgent\"]'), or a native JSON array in tool args (required for proto/typed decode when tags include '#'). CLI: --tags CSV and/or repeated --tag.",
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
		Description:   "Preferred local AI backend (fm, ollama)",
		Schema: map[string]interface{}{
			"type":        "string",
			"description": "For create/update: preferred local LLM for estimation (fm|ollama). Stored in task metadata as preferred_backend. For summarize/run_with_ai: overrides task metadata to select backend.",
			"enum":        []string{"", "fm", "ollama"},
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
		CanonicalName: "priority_rank",
		MCPName:       "priority_rank",
		CLIFlag:       "priority-rank",
		Description:   "New priority_rank (integer)",
		Schema: map[string]interface{}{
			"type":        "integer",
			"description": "For update: numeric sort order within the same priority band.",
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
			"examples":    []string{"Add OAuth2 login", "Fix session timeout bug"},
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
		Description:   "Preferred local AI backend (fm, ollama)",
		Schema: map[string]interface{}{
			"type":        "string",
			"description": "For create/update: preferred local LLM for estimation (fm|ollama). Stored in task metadata as preferred_backend. For summarize/run_with_ai: overrides task metadata to select backend.",
			"enum":        []string{"", "fm", "ollama"},
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

// OptionalInt is an optional integer field for CLI → task_workflow mapping.
type OptionalInt struct {
	Set   bool
	Value int
}

type TaskCreateInput struct {
	Name             string
	LongDescription  OptionalString
	Priority         OptionalString
	PriorityRank     OptionalInt
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
	PriorityRank     OptionalInt
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

// toolArgStringSlice returns a defensive copy for task_workflow JSON args.
// Proto fields tags and dependencies are repeated string; protojson requires a JSON array,
// not a single comma-separated string (otherwise unmarshal fails with e.g. unexpected token on "#tag").
func toolArgStringSlice(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	out := make([]string, len(values))
	copy(out, values)
	return out
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
	if in.PriorityRank.Set {
		toolArgs["priority_rank"] = in.PriorityRank.Value
	}
	if in.Tags.Set && len(in.Tags.Values) > 0 {
		toolArgs["tags"] = toolArgStringSlice(in.Tags.Values)
	}
	if in.Dependencies.Set && len(in.Dependencies.Values) > 0 {
		toolArgs["dependencies"] = toolArgStringSlice(in.Dependencies.Values)
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
	if in.PriorityRank.Set {
		toolArgs["priority_rank"] = in.PriorityRank.Value
	}
	if in.Tags.Set && len(in.Tags.Values) > 0 {
		toolArgs["tags"] = toolArgStringSlice(in.Tags.Values)
	}
	if in.RemoveTags.Set && len(in.RemoveTags.Values) > 0 {
		// TaskWorkflowRequest has no remove_tags field yet; kept for non-proto / future wire parity.
		toolArgs["remove_tags"] = toolArgStringSlice(in.RemoveTags.Values)
	}
	if in.Name.Set {
		toolArgs["name"] = in.Name.Value
	}
	if in.LongDescription.Set {
		toolArgs["long_description"] = in.LongDescription.Value
	}
	if in.Dependencies.Set && len(in.Dependencies.Values) > 0 {
		toolArgs["dependencies"] = toolArgStringSlice(in.Dependencies.Values)
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

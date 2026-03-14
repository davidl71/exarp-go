// report_plan_overrides.go — Project-specific plan overrides loaded from .exarp/plan.json.
//
// Drop a .exarp/plan.json in any project to customize plan generation output without
// touching the protobuf config schema. All fields are optional; unset fields fall back
// to the generator defaults.
//
// Example .exarp/plan.json:
//
//	{
//	  "overview": "Ship a production-ready MCP server with AI-powered task management.",
//	  "success_criteria": "All scorecard dimensions green; zero known security issues.",
//	  "storage_note": "Todo2 (SQLite primary, JSON fallback)",
//	  "invariants_note": "Use Makefile targets; prefer MCP tools over direct file edits",
//	  "agents": [
//	    {"name": "wave-task-runner", "path": ".cursor/agents/wave-task-runner.md", "role": "Run one task per wave"},
//	    {"name": "wave-verifier",    "path": ".cursor/agents/wave-verifier.md",    "role": "Verify wave outcomes"}
//	  ],
//	  "referenced_by": [
//	    ".cursor/agents/wave-task-runner.md",
//	    ".cursor/rules/plan-execution.mdc"
//	  ]
//	}
package tools

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// PlanOverrides holds project-specific overrides for plan generation.
// All fields are optional. Empty strings / nil slices mean "use default".
type PlanOverrides struct {
	// Overview is the plan purpose sentence written into the YAML frontmatter and ## Scope section.
	Overview string `json:"overview"`

	// SuccessCriteria appears in the ## Scope section under "Success criteria".
	SuccessCriteria string `json:"success_criteria"`

	// StorageNote replaces the hardcoded "Todo2 (SQLite primary, JSON fallback)" line in ## 1. Technical Foundation.
	StorageNote string `json:"storage_note"`

	// InvariantsNote replaces the hardcoded invariants line in ## 1. Technical Foundation.
	InvariantsNote string `json:"invariants_note"`

	// Agents overrides the agent list in the YAML frontmatter and ## Agents section.
	Agents []PlanAgent `json:"agents"`

	// ReferencedBy overrides the referenced_by list in the YAML frontmatter.
	ReferencedBy []string `json:"referenced_by"`
}

// PlanAgent describes a Cursor agent referenced by the plan.
type PlanAgent struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Role string `json:"role"`
}

// planOverridesPath returns the path to the project's .exarp/plan.json file.
func planOverridesPath(projectRoot string) string {
	return filepath.Join(projectRoot, ".exarp", "plan.json")
}

// loadPlanOverrides reads .exarp/plan.json; returns an empty struct (all defaults) if absent or unreadable.
func loadPlanOverrides(projectRoot string) PlanOverrides {
	data, err := os.ReadFile(planOverridesPath(projectRoot))
	if err != nil {
		return PlanOverrides{}
	}
	var ov PlanOverrides
	if err := json.Unmarshal(data, &ov); err != nil {
		return PlanOverrides{}
	}
	return ov
}

// defaultAgents returns the built-in Cursor agent list used when no override is provided.
func defaultAgents() []PlanAgent {
	return []PlanAgent{
		{Name: "wave-task-runner", Path: ".cursor/agents/wave-task-runner.md", Role: "Run one task per wave from this plan"},
		{Name: "wave-verifier", Path: ".cursor/agents/wave-verifier.md", Role: "Verify wave outcomes and update status"},
	}
}

// defaultReferencedBy returns the built-in referenced_by list.
func defaultReferencedBy() []string {
	return []string{
		".cursor/agents/wave-task-runner.md",
		".cursor/agents/wave-verifier.md",
		".cursor/rules/plan-execution.mdc",
	}
}

// resolvedAgents returns override agents if set, otherwise defaults.
func (o *PlanOverrides) resolvedAgents() []PlanAgent {
	if len(o.Agents) > 0 {
		return o.Agents
	}
	return defaultAgents()
}

// resolvedReferencedBy returns override referenced_by if set, otherwise defaults.
func (o *PlanOverrides) resolvedReferencedBy() []string {
	if len(o.ReferencedBy) > 0 {
		return o.ReferencedBy
	}
	return defaultReferencedBy()
}

// resolvedStorageNote returns the storage note with fallback.
func (o *PlanOverrides) resolvedStorageNote() string {
	if o.StorageNote != "" {
		return o.StorageNote
	}
	return "Todo2 (SQLite primary, JSON fallback)"
}

// resolvedInvariantsNote returns the invariants note with fallback.
func (o *PlanOverrides) resolvedInvariantsNote() string {
	if o.InvariantsNote != "" {
		return o.InvariantsNote
	}
	return "Use Makefile targets; prefer report/task_workflow over direct file edits"
}

// resolvedSuccessCriteria returns the success criteria with fallback.
func (o *PlanOverrides) resolvedSuccessCriteria() string {
	if o.SuccessCriteria != "" {
		return o.SuccessCriteria
	}
	return "Clear milestones and quality gates; backlog aligned with execution order."
}

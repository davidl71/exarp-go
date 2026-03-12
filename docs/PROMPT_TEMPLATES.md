# Prompt Templates Architecture

This document explains the prompt template system in exarp-go.

## Overview

Prompts are defined in two places:
- `internal/prompts/registry.go` - Register prompts with MCP server
- `internal/prompts/templates.go` - Template content

## Template Types

### 1. Standalone Prompts (36)
Registered prompts available to MCP clients:
- Workflow: `daily_checkin`, `sprint_start`, `sprint_end`, `pre_sprint`, `post_impl`, `sync`, `dups`
- Task: `task_update`, `align`, `discover`, `config`, `scan`, `scorecard`, `overview`, `plan`, `dashboard`, `remember`
- Persona: `persona_developer`, `persona_project_manager`, `persona_code_reviewer`, `persona_executive`, `persona_security`, `persona_architect`, `persona_qa`, `persona_tech_writer`
- Advisor: `advisor_consult`, `advisor_briefing`
- Other: `context`, `mode`, `docs`, `automation_discover`, `automation_setup`, `task_review`, `project_health`, `weekly_maintenance`, `tractatus_decompose`

### 2. Placeholder Variables
Templates use `{variable_name}` placeholders that get substituted at runtime:

| Variable | Used By |
|----------|---------|
| `{task_description}` | task_breakdown, task_execution |
| `{acceptance_criteria}` | task_breakdown |
| `{context}` | task_breakdown, task_execution, prompt_optimization_* |
| `{prompt}` | prompt_optimization_* |
| `{task_type}` | prompt_optimization_* |
| `{concept}` | tractatus_decompose |

### 3. Internal Optimization Templates
Used by prompt_tracking tool for prompt quality improvement:
- `prompt_optimization_analysis` - Evaluate prompt on 5 dimensions
- `prompt_optimization_suggestions` - Generate improvement suggestions
- `prompt_optimization_refinement` - Apply suggestions to refine prompt

### 4. Task Workflow Templates
Used by MODEL_ASSISTED_WORKFLOW (see docs/MODEL_ASSISTED_WORKFLOW.md):
- `task_breakdown` - Break task into subtasks
- `task_breakdown_brief` - Quick 3-6 subtask breakdown
- `task_execution` - Execute simple tasks with code changes

## Adding a New Prompt

1. Add to `registry.go`:
```go
var allPrompts = []struct {
    name        string
    description string
}{
    {"my_prompt", "Description of what it does."},
}
```

2. Add template to `templates.go`:
```go
var promptTemplates = map[string]string{
    "my_prompt": `Template content with {placeholders}.`,
}
```

## Template Substitution

Templates support `{variable_name}` substitution:

```go
template, _ := GetPromptTemplate("task_breakdown")
substituted := SubstituteTemplate(template, map[string]interface{}{
    "task_description": "Fix login bug",
    "context": "User can't login with OAuth",
})
```

## Best Practices

1. **Keep prompts standalone** - Each prompt should be independently useful
2. **Use placeholders sparingly** - Only when content varies at runtime
3. **Document tool usage** - Prompts should reference relevant MCP tools
4. **Test templates** - Use `templates_test.go` to verify

## Testing

Run template tests:
```bash
go test ./internal/prompts/... -v -run TestGetPromptTemplate
```

---
name: thinking-workflow
description: Chain tractatus_thinking (structure), sequential-thinking (process), and exarp-go MCP (execution) for complex analysis and task enrichment. Use when analyzing backlogs, decomposing problems, planning sprints, or enriching tasks with dependencies, tags, and priorities.
---

# Thinking Workflow Skill

Apply this skill when you need to **analyze, plan, and execute** complex multi-task work — backlog enrichment, sprint planning, dependency analysis, or problem decomposition.

## Pipeline: Structure → Process → Execute

The three MCP tools form a pipeline. Each stage produces output the next stage consumes.

| Stage | Tool | Purpose | Output |
|-------|------|---------|--------|
| **1. Structure** | `tractatus_thinking` | Decompose problem into clusters, reveal dependencies, identify what must ALL be true | Clusters, priority insights, dependency chains |
| **2. Process** | `sequentialthinking` | Map clusters to specific task IDs and concrete changes, plan per-cluster enrichments | Per-task change list with fields to update |
| **3. Execute** | exarp-go MCP tools | Apply changes via `CallMcpTool` — never shell out to CLI | Updated tasks with tags, deps, priorities, descriptions |

## When to Use

| Scenario | Use this workflow |
|----------|-------------------|
| **Backlog enrichment** | Cluster tasks, add tags/deps/priorities, improve descriptions |
| **Sprint planning** | Decompose scope, identify blockers, create execution waves |
| **Problem decomposition** | Break fuzzy requirements into actionable tasks |
| **Dependency analysis** | Find hidden blocking chains, elevate blocker priorities |
| **Post-completion refinement** | After finishing work, assess and update remaining tasks |

## Stage 1: Tractatus (Structure)

Use `tractatus_thinking` MCP to decompose the problem space.

```
CallMcpTool(
  server="user-tractatus_thinking",
  toolName="tractatus_thinking",
  arguments={
    "operation": "start",
    "concept": "What is the logical structure of [problem]?",
    "style": "analytical",
    "depth_limit": 4
  }
)
```

Then `add` propositions for each cluster discovered:

```
CallMcpTool(
  server="user-tractatus_thinking",
  toolName="tractatus_thinking",
  arguments={
    "operation": "add",
    "session_id": "<from start>",
    "content": "Cluster A: [name] — [tasks] with [dependency relationship]",
    "confidence": 0.9
  }
)
```

Finally `export` the structure:

```
CallMcpTool(
  server="user-tractatus_thinking",
  toolName="tractatus_thinking",
  arguments={
    "operation": "export",
    "session_id": "<from start>",
    "format": "markdown"
  }
)
```

**Tractatus outputs to capture:**
- Cluster names and membership (→ tags)
- Dependency chains between clusters (→ task dependencies)
- Priority insights — which tasks block others (→ priority elevation)
- Multiplicative relationships — what must ALL be true (→ acceptance criteria)

## Stage 2: Sequential Thinking (Process)

Use `sequentialthinking` MCP to plan concrete changes per cluster.

```
CallMcpTool(
  server="user-sequential-thinking",
  toolName="sequentialthinking",
  arguments={
    "thought": "Cluster A has tasks T-xxx, T-yyy. Tag: #cluster-a. T-xxx depends on T-zzz. Priority: keep medium. Description enrichment: add cluster context.",
    "nextThoughtNeeded": true,
    "thoughtNumber": 1,
    "totalThoughts": <num_clusters + 1>
  }
)
```

One thought per cluster, final thought summarizes all changes.

**Sequential outputs to capture (per task):**
- `tags` — cluster tag + category tags
- `dependencies` — blocking task IDs
- `priority` — elevated if task blocks others
- `long_description` — enriched with cluster context and rationale

## Stage 3: Exarp-go MCP (Execute)

**Always use `CallMcpTool` — never shell out to CLI for MCP-available operations.**

### Update tasks

```
CallMcpTool(
  server="project-0-exarp-go-exarp-go",
  toolName="task_workflow",
  arguments={
    "action": "update",
    "task_id": "T-xxx",
    "tags": "#cluster-a #category",
    "dependencies": "T-yyy",
    "priority": "high",
    "long_description": "[Cluster A: Name] Enriched description with context and rationale."
  }
)
```

### Create tasks

```
CallMcpTool(
  server="project-0-exarp-go-exarp-go",
  toolName="task_workflow",
  arguments={
    "action": "create",
    "name": "Task name",
    "long_description": "Description with cluster context",
    "tags": "#cluster-tag",
    "dependencies": "T-blocker-id",
    "priority": "medium"
  }
)
```

### List/filter tasks

```
CallMcpTool(
  server="project-0-exarp-go-exarp-go",
  toolName="task_workflow",
  arguments={
    "action": "sync",
    "sub_action": "list",
    "status": "Todo",
    "filter_tag": "#cluster-a",
    "output_format": "json",
    "compact": true
  }
)
```

### Validate with task_analysis

After enrichment, use `task_analysis` to validate:

```
CallMcpTool(
  server="project-0-exarp-go-exarp-go",
  toolName="task_analysis",
  arguments={
    "action": "dependencies",
    "output_format": "json"
  }
)
```

Other useful validations:
- `action=duplicates` — check for duplicate tasks after enrichment
- `action=execution_plan` — generate waves from dependencies
- `action=suggest_dependencies` — auto-infer missed dependencies
- `action=complexity` — assess task complexity scores

## Mapping: Tractatus → Task Fields

| Tractatus concept | exarp-go field | Example |
|-------------------|----------------|---------|
| Cluster name | `tags` | `#test-infra`, `#mcp-go-core` |
| Dependency chain | `dependencies` | `"T-123,T-456"` |
| Blocker identification | `priority` elevated to `high` | Entry-point task blocks 3+ others |
| Multiplicative requirement | `long_description` | "Must ALL be true: X, Y, Z" |
| Atomic proposition | Task acceptance criteria | Measurable, specific condition |

## Anti-Patterns

| Don't | Do instead |
|-------|-----------|
| Shell out to `./bin/exarp-go` for task ops | Use `CallMcpTool` with `project-0-exarp-go-exarp-go` |
| Skip tractatus and jump to sequential | Structure first — understanding WHAT enables HOW |
| Skip sequential and jump to execution | Plan changes per cluster before batch-applying |
| Apply changes without validation | Run `task_analysis` after enrichment |
| Use tractatus for implementation steps | Use tractatus for structure, sequential for steps |

## Complete Example Flow

1. **Prime session** via exarp-go MCP `session(action=prime)`
2. **Get backlog** via `task_workflow(action=sync, sub_action=list, status=Todo)`
3. **Tractatus start** → decompose into clusters
4. **Tractatus add** → propositions for each cluster + dependencies + priorities
5. **Tractatus export** → capture structure as markdown
6. **Sequential thought 1..N** → map each cluster to task IDs and field changes
7. **Sequential final thought** → summary of all changes
8. **Execute** → `CallMcpTool` task_workflow updates per task
9. **Validate** → `task_analysis(action=dependencies)` + `task_analysis(action=execution_plan)`

## Decision Flow

1. **Complex/fuzzy problem?** → Start with tractatus (Stage 1)
2. **Structure clear, need process?** → Start with sequential (Stage 2)
3. **Changes known, just apply?** → Start with exarp-go MCP (Stage 3)
4. **Quick single-task update?** → Skip pipeline, use `task_workflow` directly

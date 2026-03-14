# Plan Overrides

**Status:** Reference
**Related:** `CONFIGURATION_REFERENCE.md`, `docs/example.plan.md`

---

## Overview

`report action=plan` generates `.cursor/plans/<project>.plan.md` from the current Todo2 backlog. Several fields in the generated output are derived automatically (module name from `go.mod`, hardcoded agents, etc.). For project-specific customization, place a `.exarp/plan.json` file in the repository root.

All fields are **optional**. Omitted fields fall back to built-in defaults. The file is read at generation time; no restart or rebuild is needed.

---

## File location

```
<project-root>/.exarp/plan.json
```

---

## Schema

```json
{
  "overview":          "<string>",
  "success_criteria":  "<string>",
  "storage_note":      "<string>",
  "invariants_note":   "<string>",
  "agents": [
    { "name": "<string>", "path": "<string>", "role": "<string>" }
  ],
  "referenced_by": [
    "<path-or-url>"
  ]
}
```

| Field | Default | Description |
|-------|---------|-------------|
| `overview` | `"MCP Server"` from `go.mod` / package info | Written into `overview:` frontmatter and the **## Scope → Purpose** line. |
| `success_criteria` | `"Clear milestones and quality gates; backlog aligned with execution order."` | Written into **## Scope → Success criteria**. |
| `storage_note` | `"Todo2 (SQLite primary, JSON fallback)"` | Written into **## 1. Technical Foundation → Storage**. |
| `invariants_note` | `"Use Makefile targets; prefer report/task_workflow over direct file edits"` | Written into **## 1. Technical Foundation → Invariants**. |
| `agents` | `wave-task-runner` + `wave-verifier` | Agent table in **## Agents** and the `agents:` frontmatter block. |
| `referenced_by` | Three `.cursor/` paths | `referenced_by:` frontmatter list and **Referenced by** body links. |

---

## Minimal example

Override just the overview and success criteria:

```json
{
  "overview": "Ship a production-ready payment service with zero-downtime deploys.",
  "success_criteria": "All integration tests pass; PCI audit passed; <200 ms p99 latency."
}
```

---

## Full example

```json
{
  "overview": "Ship a production-ready MCP server with AI-powered task management, local LLM inference, and semantic search.",
  "success_criteria": "All scorecard dimensions green; zero known security issues; llamacpp, chromem-go, and Apple FM fully integrated.",
  "storage_note": "Todo2 (SQLite primary, JSON fallback)",
  "invariants_note": "Use Makefile targets; prefer MCP tools over direct file edits; always prefix make/go commands with NO_COLOR=1",
  "agents": [
    {
      "name": "wave-task-runner",
      "path": ".cursor/agents/wave-task-runner.md",
      "role": "Run one task per wave from this plan"
    },
    {
      "name": "wave-verifier",
      "path": ".cursor/agents/wave-verifier.md",
      "role": "Verify wave outcomes and update status"
    }
  ],
  "referenced_by": [
    ".cursor/agents/wave-task-runner.md",
    ".cursor/agents/wave-verifier.md",
    ".cursor/rules/plan-execution.mdc"
  ]
}
```

---

## How it interacts with `repair`

`report action=plan repair=true` rewrites the frontmatter and **## 3. Iterative Milestones** section of an existing plan without touching the body. It reads `.exarp/plan.json` using the same loader, so overridden fields are preserved on repair.

---

## Triggering a regeneration

```bash
# Regenerate plan (reads .exarp/plan.json automatically)
exarp-go -tool report -args '{"action":"plan"}'

# Or via MCP
report  action=plan
```

The generated file is typically `.cursor/plans/<project-name>.plan.md`.

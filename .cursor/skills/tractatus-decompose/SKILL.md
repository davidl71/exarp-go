---
name: tractatus-decompose
description: Use Tractatus Thinking MCP for logical decomposition. Use when the user needs to break down complex concepts, analyze structure before implementation, or decompose problems into atomic propositions.
---

# Tractatus Decompose

Apply this skill when the user needs **logical decomposition** of a concept, research question, or design problem—especially before implementation or when "what must be true" is unclear.

## When to use

- User asks to "decompose", "break down logically", "analyze structure", or "what are the components of X?"
- Complex research or design where bundled ideas hide the real issue.
- Before implementation: understand **what** (structure) before **how** (steps).
- Problems requiring multiplicative understanding (A × B × C must all be true).

## Tractatus MCP (Cursor)

Tractatus is available as the **tractatus_thinking** MCP server in Cursor. Call it directly; exarp-go does not invoke it.

### Key operations

| Operation | Purpose |
|-----------|---------|
| **start** | Begin analysis: `operation="start"`, `concept="What is X?"` |
| **add** | Add propositions (clarification, analysis, cases, implication, negation); use `session_id` from start. |
| **export** | Get result: `operation="export"`, `session_id="..."`, `format="markdown"` or `"json"`. |
| **navigate** | Move in the proposition tree: `target="parent"`, `"child"`, `"sibling"`, or proposition number. |
| **revise** | Update a proposition: `operation="revise"`, `proposition_number`, `new_content`. |

### Example flow

1. **Start:** `tractatus_thinking(operation="start", concept="What is framework-agnostic MCP server design?")` → get `session_id`.
2. **Add** propositions as needed with `operation="add"`, `session_id`, `content`, `decomposition_type`, optional `parent_number`.
3. **Export:** `tractatus_thinking(operation="export", session_id="...", format="markdown")` → use for implementation or docs.

## Summary

Use the **tractatus_thinking** MCP tool for structural analysis and logical decomposition. Prefer it **first** for complex or fuzzy concepts; then use sequential/process thinking for implementation steps.

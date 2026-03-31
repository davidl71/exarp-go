# Context tools: `context` vs `context_budget`

**Canonical audit** (request `context.Context`, FM chain, code changes, follow-ups):  
[CONTEXT_AND_CTX_AUDIT_2026-03.md](CONTEXT_AND_CTX_AUDIT_2026-03.md)

## Quick comparison

| | **`context`** | **`context_budget`** |
|---|---------------|----------------------|
| **Role** | One MCP tool: `summarize`, `budget`, `batch`, `count` | Token budget planning only |
| **Code** | `handlers_ai.go` → `context.go`, `context_shared.go` | `context.go` |
| **AI path** | `summarize` / `batch` use **`DefaultFMProvider()`** (FM chain; not Python-only, not Apple-only) | No AI — estimates only |
| **Budget** | `action=budget` uses native **`handleContextBudget`** (same idea as the standalone tool) | Native **`handleContextBudget`** |
| **Pick this when** | You need summarization, batch processing, or a single entrypoint | You have **many** items and want fast token totals + “what to summarize” hints |

## Workflow

1. **`context_budget`** — measure items against a token limit and read recommendations.  
2. **`context`** (`summarize` / `batch`) — compress what you decided to trim.

## Legacy doc

Longer 2026-01 material (Python-bridge / Apple-only wording, extended examples) lives in  
[archive/context-tools/CONTEXT_TOOLS_COMPARISON_legacy_2026-03-31.md](archive/context-tools/CONTEXT_TOOLS_COMPARISON_legacy_2026-03-31.md).

# Factory and construction candidates

This document complements [`ENUM_CANDIDATES.md`](./ENUM_CANDIDATES.md) (fixed strings / `Parse*`) with **`New*` / registry / switch-on-kind** consolidation opportunities—one place to construct implementations from config or user input.

**Design constraint:** any new `internal/...` package must avoid import cycles (e.g. `tools` must not be imported by a low-level factory package if that package is pulled in from `tools`). Prefer small helpers next to existing interfaces (`TextGenerator`, `FMProvider`) or a thin `internal/llmfactory`-style package that only depends on stable interfaces.

## Priority 1 — Do first (highest ROI)

### Unified LLM / text backend construction

**Problem:** The same backend vocabulary and fallback behavior is spread across:

- [`internal/tools/text_generate.go`](../internal/tools/text_generate.go) — `switch provider` → `DefaultFMProvider`, `DefaultOllamaTextGenerator`, `DefaultReportInsight`, `DefaultLocalAIProvider`, `DefaultGatewayProvider`, `auto`
- [`internal/tools/task_workflow_ai_run.go`](../internal/tools/task_workflow_ai_run.go) — `generateWithBackend` (`ollama` vs `fm` with Ollama fallback; `mlx` → `fm`)
- [`internal/tools/estimation_shared_v2.go`](../internal/tools/estimation_shared_v2.go) and related estimation paths — `switch backend`
- [`internal/tools/estimation_shared.go`](../internal/tools/estimation_shared.go) — `GetPreferredBackend` (`fm`, `ollama`, `mlx` → auto)

**Target:** One documented API, e.g. `NormalizeLLMBackend(string) string` plus `NewTextGeneratorForBackend(...)` (or equivalent using existing `TextGenerator` / `FMProvider`), with table-driven tests. Call sites keep behavior; duplication and drift shrink.

**Acceptance (summary):** existing tests for `text_generate`, task workflow AI, and estimation still pass; new unit tests for normalization and factory errors.

## Priority 2 — Medium

### Agent runner: string → `AgentType`

[`internal/tools/agent_runner.go`](../internal/tools/agent_runner.go) already has **`RegisterAgentRunner` / `GetAgentRunner`**. If CLI, TUI, or tools pass agent kind as a string, add **`ParseAgentType(string) (AgentType, error)`** (or `MustParse`) and route through `RunAgent` so unknown strings fail in one place.

**Acceptance:** parser tests; any string-based dispatch updated to use the parser.

## Priority 3 — Lower

### Child agent: `ChildAgentKind` parsing

[`internal/cli/child_agent.go`](../internal/cli/child_agent.go) defines `ChildAgentKind` consts. If flags or API accept raw strings, add **`ParseChildAgentKind(string) (ChildAgentKind, error)`** and use from TUI/CLI entry points to avoid duplicated `switch` strings.

## Priority 4 — Lower / maintenance

### Database driver registry (non-SQLite)

[`internal/database/driver.go`](../internal/database/driver.go) — `DriverRegistry` + `GetDriver`; SQLite registered in `init`; comment notes MySQL/Postgres on-demand.

**Target:** Audit call sites; ensure `GetDriver` for `mysql` / `postgres` registers or returns a clear error; document in config validation. Not a new “factory” so much as completing the existing pattern.

## Already in good shape (reference only)

| Area | Location |
|------|-----------|
| Task store | `NewDefaultTaskStore` — [`internal/tools/task_store.go`](../internal/tools/task_store.go) |
| MCP server framework | `factory.NewServer` — [`internal/factory/server.go`](../internal/factory/server.go) (today: GoSDK branch; multi-framework is design-only) |

## References

- [`docs/ENUM_CANDIDATES.md`](./ENUM_CANDIDATES.md) — enum / `Parse*` workstreams.
- [`docs/tool-parameter-parsing.md`](./tool-parameter-parsing.md) — handler param patterns.
- [`docs/FRAMEWORK_AGNOSTIC_DESIGN.md`](./FRAMEWORK_AGNOSTIC_DESIGN.md) — future multi-framework server factory.

## Todo2 tracking (exarp-go project)

| Priority | Workstream | Task ID |
|----------|------------|---------|
| P1 | LLM / text backend factory + normalization | `T-1775395318272574000` |
| P2 | AgentType string parser + wiring | `T-1775395318297351000` |
| P3 | ChildAgentKind parser (CLI/TUI) | `T-1775395318321916000` |
| P4 | Database driver registry audit | `T-1775395318347065000` |

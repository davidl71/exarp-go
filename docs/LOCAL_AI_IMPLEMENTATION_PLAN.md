# Local AI Implementation Plan

**Tag hints:** `#local-ai` `#implementation` `#planning`  
**Parent epic:** T-1771164552852 (Local AI task assignment: MLX, AFM, Ollama)  
**Status:** Complete — A1, A2, B1, B2, B3, and Phase C done.

---

## Progress / Session (2026-02-18)

- **A1** T-1771172293634: Done — create, update, summarize, run_with_ai accept/pass `local_ai_backend`; update sets `preferred_backend`; test in task_workflow_test.go.
- **A2** T-1771171137798: Done — `local_ai_backend` on EstimationRequest; estimation uses it.
- **B1** T-1771171136408: Done — `exarp-go task estimate "Name" --local-ai-backend`; test for create with local_ai_backend.
- **B2** T-1771171132414: Done — `exarp-go task summarize T-xxx [--local-ai-backend]`; documented preferred_backend in MODEL_ASSISTED_WORKFLOW.md.
- **B3** T-1771171134703: Done — `exarp-go task run-with-ai T-xxx [--backend] [--instruction]`; documented.
- **Phase C:** Done — MODEL_ASSISTED_WORKFLOW, CURSOR_API_AND_CLI_INTEGRATION, CLAUDE.md, and `exarp-go task help` updated.

---

## 1. Goal

Fully wire **local AI backends** (Apple FM, Ollama, MLX) into task lifecycle: creation, estimation, summarization, and lightweight “run with AI” execution, with clear CLI and MCP surface and no new LLM integrations—only preference plumbing and verification.

---

## 2. Current State

| Area | Status | Notes |
|------|--------|------|
| **Task metadata** | Done | `preferred_backend` (fm\|mlx\|ollama) stored on create; estimation/summarize/run_with_ai read it |
| **task_workflow create** | Done | Accepts `local_ai_backend`, stores as `preferred_backend` |
| **task_workflow summarize** | Done | `task_id` + optional `local_ai_backend`; uses DefaultReportInsight / text_generate |
| **task_workflow run_with_ai** | Done | `task_id` + optional `local_ai_backend`, `instruction`; returns implementation guidance |
| **estimation tool** | Done | Accepts `local_ai_backend`; routes to FM/ollama/MLX per param or task metadata |
| **CLI task create** | Done | `--local-ai-backend fm\|mlx\|ollama` |
| **CLI task estimate** | Done | `exarp-go task estimate "Name" --local-ai-backend fm\|mlx\|ollama` |
| **CLI task summarize** | Done | `exarp-go task summarize T-xxx [--local-ai-backend]` |
| **CLI task run-with-ai** | Done | `exarp-go task run-with-ai T-xxx [--backend] [--instruction]` |
| **Proto EstimationRequest** | Done (A2) | Added `local_ai_backend` to `proto/tools.proto` and `EstimationRequest` in tools.pb.go; estimation reads it via EstimationRequestToParams. Run `make proto` when protoc is available to refresh the full descriptor. |
| **Verification** | Done (A1) | Audit complete: create, update, summarize, run_with_ai accept/pass `local_ai_backend`; update allows only `local_ai_backend` to set `preferred_backend`; clarify uses FM only (no param yet). Test: `update with local_ai_backend sets preferred_backend`. |

References: [CURSOR_API_AND_CLI_INTEGRATION.md §3.6](CURSOR_API_AND_CLI_INTEGRATION.md), [MODEL_ASSISTED_WORKFLOW.md](MODEL_ASSISTED_WORKFLOW.md), [TASK_DEPENDENCIES_AND_PARENTS.md](TASK_DEPENDENCIES_AND_PARENTS.md).

---

## 3. Tasks and Order

All below are under parent **T-1771164552852** and tagged `#local-ai` where applicable.

### Phase A — Verify and complete backend plumbing

| Order | Task ID | Title | Scope |
|-------|---------|--------|--------|
| A1 | T-1771172293634 | Verify task_workflow local_ai_backend for all relevant actions | Audit: create, update, summarize, run_with_ai, link_planning, clarify; add param where it makes sense (e.g. update to set preferred_backend); document and add tests. |
| A2 | T-1771171137798 | Proto: add local_ai_backend to EstimationRequest | Add optional `string local_ai_backend = N` to `EstimationRequest` in `proto/tools.proto`; regenerate; ensure estimation tool reads it when parsing proto. |

### Phase B — CLI and UX

| Order | Task ID | Title | Scope |
|-------|---------|--------|--------|
| B1 | T-1771171136408 | CLI: expose local_ai_backend for estimation and create | Ensure `task create --local-ai-backend` is documented and tested; add CLI path for estimation with `--local-ai-backend` (e.g. `exarp-go estimate "Task name" --local-ai-backend ollama` or via task subcommand). |
| B2 | T-1771171132414 | Local AI: task summarization with preferred backend | Confirm summarize uses task `preferred_backend` when `local_ai_backend` not passed; document in MODEL_ASSISTED_WORKFLOW.md and help; optional: `exarp-go task summarize T-xxx [--local-ai-backend fm]`. |
| B3 | T-1771171134703 | Local AI: lightweight task execution (run task with local LLM) | Confirm run_with_ai is wired and documented; optional: `exarp-go task run-with-ai T-xxx [--backend fm\|ollama\|mlx] [--instruction "..."]` that calls task_workflow run_with_ai. |

### Phase C — Docs and polish

- Update [MODEL_ASSISTED_WORKFLOW.md](MODEL_ASSISTED_WORKFLOW.md) and [CURSOR_API_AND_CLI_INTEGRATION.md](CURSOR_API_AND_CLI_INTEGRATION.md) with final CLI commands and MCP args.
- Ensure [CLAUDE.md](CLAUDE.md) and `exarp-go task help` list `--local-ai-backend` and any new `estimate` / `summarize` / `run-with-ai` subcommands.

---

## 4. Implementation Notes

- **Backends:** Reuse existing tools only: `text_generate` (provider fm\|mlx\|auto), `ollama`, `mlx`, `apple_foundation_models`. See [.cursor/rules/llm-tools.mdc](.cursor/rules/llm-tools.mdc).
- **Preference order:** Explicit param `local_ai_backend` overrides task `preferred_backend`; if unset, use task value or default chain (e.g. FM → Ollama → stub).
- **No new LLM integrations:** Only wiring of task context and backend preference into existing tools.

---

## 5. Acceptance (Epic)

- [x] Every relevant task_workflow action accepts/passes `local_ai_backend` where applicable (A1).
- [x] EstimationRequest proto includes `local_ai_backend` and estimation uses it (A2).
- [x] CLI supports local AI for create, estimation, summarize, and run-with-ai (B1–B3).
- [x] Docs and help reflect all options and examples (Phase C).

---

**Plan file:** [.cursor/plans/local-ai-implementation.plan.md](../.cursor/plans/local-ai-implementation.plan.md) — Cursor Plan snippet with todos and execution order.

**Dependencies set:** A2→A1, B1→A2, B2→A1, B3→B2 (so list --tag local-ai --order execution shows correct sequence).

*Generated for #local-ai task set. Last updated: 2026-02-18. B1–B3 and Phase C implemented; epic complete.*

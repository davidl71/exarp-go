# LLM abstraction: tools, resources, discovery (current)

**Updated:** 2026-04-05  
**Scope:** How local LLM capabilities are exposed in exarp-go after removal of the **`mlx`** MCP tool and the former **direct GGUF / llama.cpp** path. Historical design notes lived in earlier revisions of this file.

---

## Exposed surfaces

| Surface | Role |
|--------|------|
| **`ollama`** MCP tool | Status, models, generate, pull, hardware, docs, etc. Uses `DefaultOllama()`. |
| **`text_generate`** MCP tool | Unified generation: `provider` = `fm`, `ollama`, `insight`, `localai`, `gateway`, `auto`. |
| **`fm_plan_and_execute`** | Planner/executor workflows using FM helpers. |
| **Report / scorecard** | Optional AI insight text via `DefaultReportInsight()` (FM chain); see `report_insights.go`. |
| **`stdio://models`** resource | Model catalog plus **`backends`** from `LLMBackendStatus()`: `fm_available`, `ollama_reachable`, `localai_available`, `gateway_available`, `ollama_tool`, `localai_tool`, `gateway_tool`, `hint`. |

**Not registered:** There is no `mlx` tool and no in-process **llama.cpp** backend; there is no `apple_foundation_models` tool in `registry_ai.go` on all builds—Apple FM is reached via `text_generate` with `provider=fm` and FM helper code paths where CGO/darwin allow.

---

## Internal abstraction (summary)

- **FMProvider / `DefaultFMProvider()`** — FM chain (may include Ollama probe on stock builds); used across tools that need on-device or local generation.
- **ReportInsightProvider / `DefaultReportInsight()`** — Used for report/scorecard AI blurbs (FM chain, not a separate MLX bridge).
- **OllamaProvider / `DefaultOllama()`** — HTTP client to Ollama.
- **TextGenerator** — Implemented by FM, Ollama, insight, LocalAI, gateway providers; selected in `text_generate`.

---

## `text_generate` parameters (reference)

- **`provider`:** `fm` | `ollama` | `insight` | `localai` | `gateway` | `auto`
- **`prompt`**, **`max_tokens`**, **`temperature`**, optional **`model`** (for `localai` / `gateway`), optional task hints for `auto`.

Legacy stored values **`mlx`** in task metadata are normalized away (auto / fm) in task workflow code; do not pass `provider=mlx`.

---

## Optional improvements (unchanged direction)

- Richer **`stdio://models`** hints and catalog alignment with `recommend`.
- Optional dedicated **`stdio://llm/status`** only if non-MCP clients need it (today `stdio://models` is enough).
- See [LLM_NATIVE_ABSTRACTION_PATTERNS.md](LLM_NATIVE_ABSTRACTION_PATTERNS.md) and [GO_AI_ECOSYSTEM.md](GO_AI_ECOSYSTEM.md).

---

## Historical note

Older versions of this document described `mlx` and `apple_foundation_models` as first-class MCP tools, and a **`llamacpp`** tool for direct GGUF. That is **not** the current registration set; grep `RegisterTool` in `internal/tools/registry_ai.go` for the source of truth.

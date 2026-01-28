# LLM Abstraction: Exposed Tools, Prompts, Resources

**Date:** 2026-01-28  
**Scope:** How apple_foundation_models, mlx, and ollama are exposed and how to align them with the LLM abstraction (FMProvider, ReportInsightProvider, OllamaProvider, TextGenerator).

---

## Current Exposure

| Surface | apple_foundation_models | mlx | ollama |
|--------|--------------------------|-----|--------|
| **Tool** | `apple_foundation_models` (darwin/arm64/cgo only) | `mlx` | `ollama` |
| **Tool HINT** | "LLM abstraction (FM). apple_foundation_models. ... Uses DefaultFMProvider()." | "LLM abstraction (MLX). mlx. ... DefaultReportInsight()." | "LLM abstraction. ollama. ... DefaultOllama()." |
| **Resource** | — | — | — |
| **Prompt** | — | — | — |

- **stdio://models** — Returns MODEL_CATALOG (recommend-style model list) and a **backends** object (from LLMBackendStatus()): `fm_available`, tool names for ollama/mlx/apple_foundation_models, and a short hint so clients can discover which LLM backends are available.
- **Prompts** — No prompt specifically for "use an LLM"; the "context" prompt is about context budget/summarization.

---

## Alignment with LLM Abstraction

**Internal abstraction (already in place):**

- **FMProvider / DefaultFMProvider()** — Used by task_analysis, context, estimation, task_workflow, task_discovery; also by `apple_foundation_models` tool.
- **ReportInsightProvider / DefaultReportInsight()** — MLX then FM; used by report/scorecard insights only (no dedicated tool; report tool uses it internally).
- **OllamaProvider / DefaultOllama()** — Native then bridge; used by `ollama` tool.
- **TextGenerator** — Shared interface for FM and ReportInsight (generate text).

**Gap:** The **exposed** tools and resources do not tell clients that these three are part of a unified LLM abstraction or how to discover which backends are available.

---

## Implemented Changes

1. **Tool HINTs** — Updated `apple_foundation_models`, `ollama`, and `mlx` descriptions to reference the LLM abstraction so clients know they are part of the same family.
2. **stdio://models backends** — The `stdio://models` resource now includes a `backends` object: `fm_available` (bool), and tool names for `ollama` and `mlx`, so clients can discover what is available without calling each tool.
3. **Tool catalog** — `apple_foundation_models` added to the catalog (if missing); ollama and mlx descriptions updated to mention the LLM abstraction.

---

## Optional: Unified text_generate tool (design)

If a single entry point for "generate text" is desired:

- **Name:** `text_generate` (or keep using existing tools).
- **Parameters:** `provider: "fm" | "insight"`, `prompt`, `max_tokens`, `temperature`.
- **Behavior:** Call `DefaultFMProvider()` when provider is `"fm"`, or `DefaultReportInsight()` when provider is `"insight"` (MLX then FM). Both implement TextGenerator.
- **Backward compatibility:** Keep `apple_foundation_models`, `ollama`, and `mlx` as-is; `text_generate` would be an alternative for clients that only need generate-text and do not care which backend.

---

## Optional (not implemented)

- **Unified `text_generate` tool** — As designed above; not yet implemented.
- **Prompt** — A prompt such as "generate" or "llm" that directs the AI to use the appropriate LLM tool (fm, ollama, or mlx) based on context. Current prompts are domain-focused; this would be optional.
- **Resource stdio://llm/status** — Dedicated resource for LLM backend status. Currently folded into `stdio://models` as `backends`; a separate URI could be added later if needed.

---

## Files Reference

| File | Role |
|------|------|
| `internal/tools/llm_backends.go` | LLMBackendStatus() for stdio://models and discovery |
| `internal/tools/registry.go` | Tool HINTs for apple_foundation_models, ollama, mlx |
| `internal/tools/tool_catalog.go` | Catalog entries for AI & ML tools |
| `internal/resources/models.go` | stdio://models resource; includes backends from LLMBackendStatus() |

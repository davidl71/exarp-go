# LLM Abstraction: Exposed Tools, Prompts, Resources

**Date:** 2026-01-28 (updated 2026-02-22)  
**Scope:** How apple_foundation_models, mlx, ollama, and llamacpp are exposed and how to align them with the LLM abstraction (FMProvider, ReportInsightProvider, OllamaProvider, LlamaCppProvider, TextGenerator).

---

## Current Exposure

| Surface | apple_foundation_models | mlx | ollama | llamacpp |
|--------|--------------------------|-----|--------|----------|
| **Tool** | `apple_foundation_models` (darwin/arm64/cgo only) | `mlx` | `ollama` | `llamacpp` (build tag `llamacpp` + CGO) |
| **Tool HINT** | "LLM abstraction (FM). apple_foundation_models. ... Uses DefaultFMProvider()." | "LLM abstraction (MLX). mlx. ... DefaultReportInsight()." | "LLM abstraction. ollama. ... DefaultOllama()." | "LLM abstraction (llama.cpp). llamacpp. Direct GGUF inference. Uses DefaultLlamaCppProvider()." |
| **Resource** | — | — | — | — |
| **Prompt** | — | — | — | — |

- **stdio://models** — Returns MODEL_CATALOG (recommend-style model list) and a **backends** object (from LLMBackendStatus()): `fm_available`, `llamacpp_available`, tool names for ollama/mlx/apple_foundation_models/llamacpp, and a short hint so clients can discover which LLM backends are available.
- **Prompts** — No prompt specifically for "use an LLM"; the "context" prompt is about context budget/summarization.

---

## Alignment with LLM Abstraction

**Internal abstraction (already in place):**

- **FMProvider / DefaultFMProvider()** — Used by task_analysis, context, estimation, task_workflow, task_discovery; also by `apple_foundation_models` tool.
- **ReportInsightProvider / DefaultReportInsight()** — MLX then FM; used by report/scorecard insights only (no dedicated tool; report tool uses it internally).
- **OllamaProvider / DefaultOllama()** — Native then bridge; used by `ollama` tool.
- **LlamaCppProvider / DefaultLlamaCppProvider()** — Direct GGUF inference via go-llama.cpp. No server required. Built only with `llamacpp` build tag + CGO. Used by `llamacpp` tool and as `text_generate provider=llamacpp`.
- **TextGenerator** — Shared interface for FM, ReportInsight, and LlamaCpp (generate text).

**Gap:** ~~The **exposed** tools and resources do not tell clients that these are part of a unified LLM abstraction or how to discover which backends are available.~~ Resolved — `stdio://models` backends object now includes all four backends.

---

## Implemented Changes

1. **Tool HINTs** — Updated `apple_foundation_models`, `ollama`, `mlx`, and `llamacpp` descriptions to reference the LLM abstraction so clients know they are part of the same family.
2. **stdio://models backends** — The `stdio://models` resource now includes a `backends` object: `fm_available`, `llamacpp_available` (bool), and tool names for `ollama`, `mlx`, `llamacpp`, so clients can discover what is available without calling each tool.
3. **Tool catalog** — `apple_foundation_models` added to the catalog (if missing); ollama, mlx, and llamacpp descriptions updated to mention the LLM abstraction.
4. **llamacpp backend** — Direct GGUF inference via go-llama.cpp. No HTTP server required. Supports Metal (macOS) and CUDA (Linux). Model management with automatic loading/unloading. See `docs/LLAMACPP_BUILD_REQUIREMENTS.md`.

---

## Unified text_generate tool

The `text_generate` tool provides a single entry point for text generation across all backends:

- **Name:** `text_generate`
- **Parameters:** `provider: "fm" | "ollama" | "mlx" | "llamacpp" | "localai" | "insight" | "auto"`, `prompt`, `max_tokens`, `temperature`.
- **Behavior:** Routes to the appropriate provider. `provider=auto` selects the best available backend.
- **Backward compatibility:** `apple_foundation_models`, `ollama`, `mlx`, and `llamacpp` remain available as dedicated tools with rich actions (status, models, hardware, etc.) beyond generation.

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
| `internal/tools/registry.go` | Tool HINTs for apple_foundation_models, ollama, mlx, llamacpp |
| `internal/tools/tool_catalog.go` | Catalog entries for AI & ML tools |
| `internal/resources/models.go` | stdio://models resource; includes backends from LLMBackendStatus() |
| `internal/tools/llamacpp_provider.go` | LlamaCppProvider — direct GGUF inference via go-llama.cpp |
| `internal/tools/llamacpp_model_manager.go` | ModelManager — load/unload/cache GGUF models with memory limits |
| `internal/tools/llamacpp_ollama.go` | Ollama-format model discovery for llamacpp |
| `internal/tools/llamacpp_nocgo.go` | Stub when llamacpp build tag not set |
| `docs/LLAMACPP_BUILD_REQUIREMENTS.md` | Build prerequisites, Metal/CUDA, environment variables |
| `docs/LLAMACPP_BENCHMARKS.md` | Performance comparison vs Ollama HTTP |

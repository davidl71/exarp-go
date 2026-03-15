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

## llamacpp Backend

### What It Is

The `llamacpp` backend provides **direct GGUF model inference via go-skynet/go-llama.cpp** without any HTTP server overhead. Models are loaded in-process using the llama.cpp C library (linked via CGO). This is the only backend that requires both a build tag (`llamacpp`) and CGO to be compiled in.

Key properties:
- No HTTP round-trips — inference happens in the same process as the MCP server.
- Reads GGUF files directly from disk (or from Ollama's blob storage).
- Supports Metal GPU acceleration on macOS Apple Silicon and CUDA on Linux/Windows.
- Thread-safe `ModelManager` with LRU eviction: up to 3 models cached in memory simultaneously.

### When to Use llamacpp

| Situation | Recommended backend |
|-----------|---------------------|
| Fully offline, no network at all | `llamacpp` |
| Lowest possible latency (no HTTP overhead) | `llamacpp` |
| Already have GGUF files locally | `llamacpp` |
| macOS Apple Silicon + Metal GPU | `llamacpp` or `apple_foundation_models` |
| Ollama server running and accessible | `ollama` (simpler setup) |
| On-device Apple Foundation Models (phi-4-mini, etc.) | `apple_foundation_models` |
| Apple Silicon, research/generation workflows | `mlx` |
| Routing across all available backends | `text_generate provider=auto` |

### Configuration Options

**Environment variables (runtime):**

| Variable | Required | Description |
|----------|----------|-------------|
| `LLAMACPP_MODEL_PATH` | No | Path to a GGUF file to auto-load on startup |
| `EXARP_GPU_MEMORY_MB` | No | Override detected GPU memory (MB) for layer calculation |
| `CUDA_VISIBLE_DEVICES` | No | Enable CUDA backend; set to device index(es) |
| `OLLAMA_MODELS` | No | Custom Ollama model root for GGUF blob discovery |

**Environment variables (build):**

| Variable | When | Description |
|----------|------|-------------|
| `LLAMACPP_DIR` | Build | Path to go-llama.cpp clone with compiled `libbinding.a` |
| `C_INCLUDE_PATH` | Build | Must include `$LLAMACPP_DIR` |
| `LIBRARY_PATH` | Build | Must include `$LLAMACPP_DIR` |

**Tool parameters (per-call):**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `model_path` | string | — | Full path to a `.gguf` file |
| `model` | string | — | Ollama model name or alias (e.g. `llama3.2:latest`) |
| `gpu_layers` | integer | -1 | Layers to offload to GPU (-1 = all) |
| `context_size` | integer | 2048 | Context window size in tokens |
| `max_tokens` | integer | 512 | Maximum tokens to generate |
| `temperature` | number | 0.7 | Sampling temperature |

**Model aliases (`.exarp/llamacpp.json`):**

```json
{
  "aliases": {
    "my-llama": "llama3.2:latest",
    "local-7b": "/path/to/my-model.gguf",
    "deepseek": "deepseek-r1:7b"
  }
}
```

Built-in aliases cover common families: `llama`, `llama3`, `llama3b`, `phi`, `phi3`, `phi4`, `mistral`, `gemma`, `qwen`, `deepseek`, `codellama`, and more.

### Example Usage via the llamacpp MCP Tool

**Check if available:**

```json
{"action": "status"}
```

**List discoverable GGUF models (from Ollama storage and configured paths):**

```json
{"action": "models"}
```

**Load a model by Ollama name:**

```json
{"action": "load", "model": "llama3.2:latest"}
```

**Load a model by file path with custom context:**

```json
{"action": "load", "model_path": "/path/to/mistral-7b-Q4_K_M.gguf", "context_size": 4096}
```

**Generate text:**

```json
{
  "action": "generate",
  "prompt": "Explain quicksort in one sentence.",
  "max_tokens": 128,
  "temperature": 0.7
}
```

**Unload a model to free memory:**

```json
{"action": "unload"}
```

**Via text_generate (unified interface):**

```json
{
  "provider": "llamacpp",
  "prompt": "Summarize this task.",
  "max_tokens": 256
}
```

### Backend Comparison

| Factor | llamacpp | ollama | apple_foundation_models | mlx |
|--------|----------|--------|-------------------------|-----|
| HTTP overhead | None (in-process) | ~1-5ms per request | None (in-process) | Bridge process |
| Server required | No | Yes (`ollama serve`) | No | No |
| Model format | GGUF | Ollama (GGUF internally) | CoreML/on-device | MLX safetensors |
| GPU support | Metal, CUDA, ROCm | Metal, CUDA | Neural Engine | Metal |
| Platform | Any (CGO required) | Any | darwin/arm64 only | Apple Silicon |
| Build complexity | High (CGO + libbinding.a) | Low | High (CGO + Xcode) | Low (Python bridge) |
| Streaming | Direct callback | HTTP SSE | Direct | Bridge |
| Multi-model cache | Yes (LRU, up to 3) | Yes (Ollama manages) | No (single model) | No |
| Cold start | ~2s (first load) | ~2-5s (server start) | Fast (on-device) | Moderate |
| Offline capable | Yes | Yes (no internet) | Yes | Yes |

**Summary of trade-offs:**
- Use `llamacpp` when you need the lowest latency, full offline operation, and direct GGUF control without running a separate server process.
- Use `ollama` when you want simpler model management (pull/run/stop) and don't mind the HTTP overhead.
- Use `apple_foundation_models` on Apple Silicon for on-device models optimized by Apple (phi-4-mini, etc.) — no GGUF files needed.
- Use `mlx` for Apple Silicon-optimized research models in safetensor format.
- Use `text_generate provider=auto` to let the system pick the best available backend automatically.

### Building with llamacpp Support

```bash
# macOS Apple Silicon (Metal)
git clone --recurse-submodules https://github.com/go-skynet/go-llama.cpp
cd go-llama.cpp && make libbinding.a BUILD_TYPE=metal
export LLAMACPP_DIR=/path/to/go-llama.cpp
export C_INCLUDE_PATH=$LLAMACPP_DIR
export LIBRARY_PATH=$LLAMACPP_DIR
CGO_ENABLED=1 go build -tags llamacpp -o bin/exarp-go ./cmd/server

# Linux (CUDA)
CMAKE_ARGS="-DLLAMA_CUBLAS=on" make libbinding.a
```

See `docs/LLAMACPP_BUILD_REQUIREMENTS.md` for full prerequisites and troubleshooting.

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
| `internal/tools/llamacpp_aliases.go` | Built-in and user-defined model alias resolution |
| `internal/tools/llamacpp_gpu.go` | GPU detection: Metal, CUDA, ROCm, CPU fallback |
| `docs/LLAMACPP_BUILD_REQUIREMENTS.md` | Build prerequisites, Metal/CUDA, environment variables |
| `docs/LLAMACPP_BENCHMARKS.md` | Performance comparison vs Ollama HTTP |
| `docs/LLAMACPP_TOOL_SCHEMA.md` | Full tool schema and action documentation |

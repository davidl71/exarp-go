# GGUF Model Compatibility Requirements for go-llama.cpp

**Tag hints:** `#documentation` `#llamacpp` `#gguf`

This document covers which GGUF models are compatible with the go-llama.cpp integration in exarp-go, including format requirements, quantization tradeoffs, context window limits, recommended models per use case, and known incompatible formats.

**See also:**
- [GGUF_COMPATIBILITY.md](GGUF_COMPATIBILITY.md) — hardware, RAM, and model sourcing details
- [LLAMACPP_BUILD_REQUIREMENTS.md](LLAMACPP_BUILD_REQUIREMENTS.md) — build and environment setup
- [LLAMACPP_EVALUATION.md](LLAMACPP_EVALUATION.md) — binding selection rationale
- [LLAMACPP_TOOL_SCHEMA.md](LLAMACPP_TOOL_SCHEMA.md) — tool API reference

---

## 1. GGUF Format Requirements

exarp-go uses [go-skynet/go-llama.cpp](https://github.com/go-skynet/go-llama.cpp) as the binding layer over [llama.cpp](https://github.com/ggml-org/llama.cpp).

**GGUF is the only supported format.** GGUF (Georgi Gerganov Universal Format) replaced the legacy GGML format. The migration happened via go-llama.cpp PR [#180](https://github.com/go-skynet/go-llama.cpp/pull/180).

Key GGUF properties:

- Extensible metadata block: architecture, quantization type, tokenizer vocabulary, and context size are embedded in the file header
- Block-wise structure enables efficient memory mapping (`mmap`) on load
- Self-describing: the loader does not need external config to understand the model

**Version compatibility:** go-llama.cpp tracks upstream llama.cpp. Models produced by any llama.cpp version that uses the current GGUF spec (v1 or later) are compatible. When new architectures ship in llama.cpp, go-llama.cpp may lag until its submodule is updated.

---

## 2. Quantization Levels and Tradeoffs

All GGUF quantization types supported by llama.cpp work with go-llama.cpp. The tradeoff is file size and memory footprint vs. output quality.

| Quantization | Bits/weight | Memory (7B model) | Quality | Recommended use |
|---|---|---|---|---|
| **Q4_0** | 4-bit | ~3.8 GB | Lower | Smallest footprint only |
| **Q4_1** | 4-bit | ~4.1 GB | Slightly better than Q4_0 | Small footprint |
| **Q4_K_M** | 4-bit (K-quant) | ~4.4 GB | Good | **General default** |
| **Q5_0** | 5-bit | ~4.7 GB | Good | Middle ground |
| **Q5_1** | 5-bit | ~5.1 GB | Better than Q5_0 | Middle ground |
| **Q5_K_M** | 5-bit (K-quant) | ~5.3 GB | Very good | Quality-conscious use |
| **Q6_K** | 6-bit (K-quant) | ~6.1 GB | High | Minimal quality loss |
| **Q8_0** | 8-bit | ~7.7 GB | Near-lossless | Best quality / high RAM |
| **F16** | 16-bit | ~14 GB | Reference | Benchmarking, fine-tuning |
| **F32** | 32-bit | ~28 GB | Maximum | Rarely used in practice |

**K-quants (Q4_K_M, Q5_K_M, Q6_K):** These use a mixed-precision approach where certain layers retain higher precision. They deliver noticeably better perplexity than base quants at similar or slightly larger file size. Prefer K-quants when available.

**Practical recommendations:**
- **Q4_K_M** — recommended default. Best balance of quality, memory, and speed for most workloads on 16 GB RAM systems.
- **Q5_K_M** — step up when output quality matters more than the ~25% extra memory.
- **Q8_0** — use when running structured output, code generation, or tasks where hallucination rate matters most. Requires ~2x the RAM of Q4_K_M.
- **F16** — benchmarking and comparison baseline only. Not practical for everyday local inference on consumer hardware.

---

## 3. Context Window Limitations with go-llama.cpp

The context window in exarp-go's llamacpp provider is set at model load time and cannot be changed without unloading and reloading the model.

**Default context size:** `2048` tokens (set via `llama.SetContext(2048)` in `internal/tools/llamacpp_provider.go`).

**Configuring context at load time:** Pass `context_size` in the `load` action:

```json
{"action": "load", "model": "llama3.2:3b", "context_size": 4096}
```

**Practical limits by available RAM (7B Q4_K_M as baseline):**

| Context tokens | Approximate extra RAM | Notes |
|---|---|---|
| 2048 (default) | baseline | Works on 8 GB systems for small models |
| 4096 | +~0.5 GB | Comfortable on 16 GB |
| 8192 | +~2 GB | 16 GB minimum for 7B models |
| 16384 | +~4 GB | 32 GB recommended for 7B |
| 32768 | +~8 GB | 32+ GB required for 7B |

**Important:** The model's trained maximum context is a separate ceiling. A Llama 3.2 model trained with 128K context can be loaded with any context size up to that maximum. Setting `context_size` higher than the model's trained limit produces undefined behavior.

**go-llama.cpp context constraint:** go-llama.cpp allocates the KV cache at load time based on `context_size`. Large context sizes consume significant memory even before any tokens are processed. For most exarp-go tasks (short prompts, tool calls), the 2048 default is sufficient.

---

## 4. Recommended Models per Use Case

These are the built-in aliases shipped with exarp-go (`internal/tools/llamacpp_aliases.go`) and the use cases they map to. Aliases resolve to Ollama model names (which in turn resolve to GGUF blob paths) or to explicit file paths.

### 4.1 General / Chat

| Alias | Resolves to | GGUF recommendation | Notes |
|---|---|---|---|
| `llama`, `llama3` | `llama3.2` | Q4_K_M or Q5_K_M | Best general-purpose; 8B version fits 16 GB |
| `llama3b` | `llama3.2:3b` | Q4_K_M | Fast; fits 8 GB; good for quick tasks |
| `llama1b` | `llama3.2:1b` | Q4_K_M | Fastest; very small; limited capability |
| `mistral`, `mistral7b` | `mistral` / `mistral:7b` | Q4_K_M | Competitive with Llama; good instruction following |
| `gemma`, `gemma2` | `gemma2` | Q4_K_M | Google Gemma 2; strong reasoning |
| `gemma3` | `gemma3` | Q4_K_M | Latest Gemma; larger model family |
| `phi`, `phi3` | `phi3` | Q4_K_M | Microsoft Phi-3; strong for size |
| `phi4` | `phi4` | Q4_K_M | Phi-4; significant quality improvement |
| `phi35` | `phi3.5` | Q4_K_M | Phi-3.5; refined Phi-3 variant |

### 4.2 Code Generation

| Alias | Resolves to | GGUF recommendation | Notes |
|---|---|---|---|
| `codellama`, `code` | `codellama` | Q4_K_M or Q5_K_M | Meta's code-specialized model |
| `qwen-coder`, `qwen2-coder` | `qwen2.5-coder` | Q4_K_M | Qwen2.5-Coder; strong code completion |
| `starcoder` | `starcoder2` | Q4_K_M | BigCode StarCoder2 |

### 4.3 Reasoning / Research

| Alias | Resolves to | GGUF recommendation | Notes |
|---|---|---|---|
| `deepseek`, `deepseek-r1` | `deepseek-r1` | Q5_K_M or Q8_0 | Chain-of-thought reasoning; larger context recommended |
| `deepseek-v3` | `deepseek-v3` | Q4_K_M | DeepSeek V3 general model |
| `qwen`, `qwen2` | `qwen2.5` | Q4_K_M | Qwen2.5; strong multilingual and reasoning |

### 4.4 Small / Fast (Edge Cases)

| Alias | Resolves to | GGUF recommendation | Notes |
|---|---|---|---|
| `smollm` | `smollm2` | Q4_K_M | HuggingFace SmolLM2; fits <4 GB |
| `tinyllama` | `tinyllama` | Q4_K_M | 1.1B; extremely fast; limited quality |
| `orca` | `orca-mini` | Q4_K_M | Small instruction-tuned model |
| `vicuna` | `vicuna` | Q4_K_M | Early RLHF model; mostly legacy |

---

## 5. Known Incompatible Formats

| Format | Status | Details |
|---|---|---|
| **GGML (pre-GGUF)** | Not supported | Old llama.cpp format before August 2023. Not loadable by current go-llama.cpp. Use the [pre-gguf](https://github.com/go-skynet/go-llama.cpp/releases/tag/pre-gguf) tag only if you must use legacy files. |
| **GGML v1–v3** | Not supported | Older quantization formats (e.g. early TheBloke uploads from 2023). Re-download as GGUF. |
| **PyTorch / safetensors** | Not supported | Hugging Face native format. Convert to GGUF using `convert-hf-to-gguf.py` from llama.cpp before use. |
| **ONNX** | Not supported | Microsoft ONNX format is not loadable. |
| **GPTQ** | Not supported | GPU-quantized format (ExLlama, AutoGPTQ). Not the same as GGUF. |
| **AWQ** | Not supported | Activation-aware weight quantization. Requires separate runtime. |
| **EXL2** | Not supported | ExLlamaV2 format. Requires ExLlama runtime. |
| **Corrupted / partial GGUF** | Fails to load | Interrupted downloads produce unreadable files. Re-download. |
| **Architecture not in llama.cpp submodule** | May fail | Very new architectures may not be in the go-llama.cpp submodule version. Update submodule and rebuild. |

**How to identify GGUF files:** Valid GGUF files begin with the magic bytes `GGUF` (0x47 0x47 0x55 0x46). Files from Hugging Face with the `.gguf` extension and files in `~/.ollama/models/blobs/` (from `ollama pull`) are GGUF.

---

## 6. Model Alias Mapping System

exarp-go ships a two-layer alias resolution system for the llamacpp provider (`internal/tools/llamacpp_aliases.go`).

### Resolution order

1. Look up the model name in the merged alias map (built-in aliases + user overrides from `.exarp/llamacpp.json`).
2. If the resolved value is a file path (starts with `/` or `~/`), expand `~/` and use the path directly.
3. Otherwise, treat the resolved value as an Ollama model name and resolve it to the corresponding GGUF blob path via `ResolveOllamaModelPath`.
4. If no alias matches, attempt `ResolveOllamaModelPath` directly on the input.

### User alias overrides

Create `.exarp/llamacpp.json` in the project root to override built-in aliases or add custom ones:

```json
{
  "aliases": {
    "my-llama": "llama3.2:latest",
    "local-7b": "/path/to/my-model-Q4_K_M.gguf",
    "deepseek": "deepseek-r1:7b",
    "work-model": "~/models/company-finetuned-Q5_K_M.gguf"
  }
}
```

User entries override built-in aliases. Built-in aliases not overridden remain available.

### Alias to GGUF path flow

```
alias name
    → merged alias map lookup
    → Ollama model name  →  ResolveOllamaModelPath  →  ~/.ollama/models/blobs/sha256-<hash>
    → file path          →  expandHome              →  /absolute/path/to/model.gguf
```

Ollama blobs are raw GGUF files stored without a `.gguf` extension. The blob path is passed directly to go-llama.cpp for loading.

---

## 7. Quick Reference

| Question | Answer |
|---|---|
| What format is required? | GGUF only (not GGML, GPTQ, ONNX, safetensors) |
| Best quantization for general use | Q4_K_M |
| Best quantization for code/reasoning | Q5_K_M or Q8_0 |
| Default context window | 2048 tokens |
| How to set larger context | `{"action": "load", "model": "...", "context_size": 4096}` |
| Best model for 8 GB RAM | llama3.2:3b Q4_K_M (~2 GB) or smollm2 |
| Best model for 16 GB RAM | llama3.2 (8B) Q4_K_M or mistral:7b Q4_K_M |
| Where are Ollama models stored? | `~/.ollama/models/blobs/sha256-<hash>` |
| Can I use an explicit file path? | Yes — set alias value to `/path/to/file.gguf` in `.exarp/llamacpp.json` |
| Old GGML model? | Not compatible; re-download as GGUF or use `pre-gguf` tag |

---

*Last updated: 2026-03. See [LLAMACPP_BUILD_REQUIREMENTS.md](LLAMACPP_BUILD_REQUIREMENTS.md) for build setup and [LLAMACPP_TOOL_SCHEMA.md](LLAMACPP_TOOL_SCHEMA.md) for tool API.*

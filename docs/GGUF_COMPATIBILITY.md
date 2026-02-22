# GGUF Model Compatibility for exarp-go llamacpp Integration

**Tag hints:** `#documentation` `#llamacpp` `#research`

exarp-go uses [go-skynet/go-llama.cpp](https://github.com/go-skynet/go-llama.cpp) to load and run GGUF models directly. This document describes what models are compatible and practical requirements for exarp-go users.

**See also:** [LLAMACPP_EVALUATION.md](LLAMACPP_EVALUATION.md) for binding selection rationale and integration design.

---

## 1. GGUF Format Overview

**GGUF** (Georgi Gerganov Universal Format) is the successor to GGML. It is the native format used by [llama.cpp](https://github.com/ggml-org/llama.cpp) and **the only format supported by go-skynet/go-llama.cpp** (as of PR [#180](https://github.com/go-skynet/go-llama.cpp/pull/180)). Older GGML models are **not** compatible.

Key properties:

- Extensible metadata (architecture, quantization, tokenizer)
- Block-wise structure for efficient memory mapping
- Content-addressed design (Ollama uses SHA-256 blob names)

---

## 2. Supported Quantization Types

All GGUF quantization types supported by llama.cpp work with go-skynet/go-llama.cpp. Trade-off: smaller files and lower memory vs. quality.

| Type | Bits/param | Use case |
|------|------------|----------|
| **Q4_0** | 4-bit | Smallest; fast; quality loss |
| **Q4_1** | 4-bit | Slightly better than Q4_0 |
| **Q4_K_M** | 4-bit | Recommended balance (quality vs size) |
| **Q5_0** | 5-bit | Middle ground |
| **Q5_1** | 5-bit | Better than Q5_0 |
| **Q5_K_M** | 5-bit | Good quality, larger than Q4 |
| **Q6_K** | 6-bit | High quality |
| **Q8_0** | 8-bit | Near lossless, larger files |
| **F16** | 16-bit | Full precision (reference quality) |
| **F32** | 32-bit | Maximum precision (rare) |

**Practical choice:** Q4_K_M or Q5_K_M for most workloads. Use Q8_0 or F16 when quality matters more than memory.

---

## 3. Model Architecture Support

go-skynet/go-llama.cpp tracks upstream llama.cpp, which supports 100+ model architectures. Commonly used families:

| Architecture | Examples |
|--------------|----------|
| **LLaMA** | llama-3.2, llama-3.1, Llama 2 |
| **Mistral** | Mistral-7B, Mixtral-8x7B |
| **Phi** | Phi-2, Phi-3, Phi-3.5 |
| **Gemma** | Gemma 2 |
| **Qwen** | Qwen2, Qwen2.5 |
| **CodeLlama** | codellama-7b |
| **Others** | Gemma, Yi, DeepSeek, GLM, etc. |

If llama.cpp supports an architecture, go-skynet bindings generally support it. For newer models, ensure your go-llama.cpp submodule is up to date.

---

## 4. Hardware Requirements

### 4.1 Acceleration by Platform

| Platform | Option | Notes |
|----------|--------|-------|
| **macOS / arm64** | Metal | Recommended: `BUILD_TYPE=metal make libbinding.a` |
| **Linux** | CPU | Default (no extra flags) |
| **Linux** | CUDA | `BUILD_TYPE=cublas` for NVIDIA GPUs |
| **Linux** | ROCm | `BUILD_TYPE=hipblas` for AMD GPUs |
| **Linux / cross-platform** | OpenBLAS | CPU acceleration: `BUILD_TYPE=openblas` |
| **Intel / others** | OpenCL | `BUILD_TYPE=clblas` when available |

### 4.2 Memory by Quantization and Model Size

Approximate RAM (model + ~1–2 GB context overhead):

| Model size | Q4_K_M | Q5_K_M | Q6_K | Q8_0 |
|------------|--------|--------|------|------|
| **7–8B** | 5–6 GB | 6–7 GB | 7–8 GB | 9–10 GB |
| **13B** | 9–10 GB | 11–12 GB | 13 GB | 15 GB |
| **70B** | 40 GB | 50 GB | 55 GB | 70 GB |

Context size increases memory: 8K tokens adds ~2 GB for a 7B model; 32K adds ~8 GB.

---

## 5. Model Size Guidelines by RAM

| System RAM | Typical fits |
|------------|--------------|
| **8 GB** | 1–3B models (Q4_K_M) |
| **16 GB** | 7–8B (Q4_K_M or Q5_K_M) |
| **32 GB** | 7–13B (Q5_K_M or Q6_K) |
| **64 GB+** | 70B models at most quantizations |

Add 1–2 GB for system and context. For Metal/CUDA, VRAM is used first; system RAM for overflow.

---

## 6. Where to Get Compatible Models

### Hugging Face

Search for `gguf`:

- [Hugging Face Models (gguf filter)](https://huggingface.co/models?search=gguf)
- Popular repos: `TheBloke`, `bartowski`, `unsloth`, `Qwen`, `meta-llama`

### Ollama Library

Ollama’s catalog uses GGUF internally. Models downloaded with `ollama pull` are GGUF-compatible and can be reused by exarp-go if you point to the correct blob path.

### Local Conversion

Use `convert.py` or `convert-hf-to-gguf.py` from [llama.cpp](https://github.com/ggml-org/llama.cpp) to convert from Hugging Face formats to GGUF.

---

## 7. Ollama Model Reuse

Models pulled with `ollama pull <model>` are stored as GGUF blobs. exarp-go can load them directly if you know the blob path.

**Default Ollama storage:**

```
~/.ollama/models/blobs/sha256-<hash>
```

(or `OLLAMA_MODELS` if customized)

Blob files are content-addressed (SHA-256 filenames). To find the main model blob:

1. Inspect `~/.ollama/models/blobs/` for large files (e.g. >1 GB for 7B Q4).
2. Or parse Ollama manifests (e.g. `~/.ollama/models/manifests/`) to resolve which blob(s) belong to a model.

**Note:** Blobs are raw GGUF files without `.gguf` extension. Pass the full path to exarp-go’s llamacpp provider.

---

## 8. Known Limitations of go-skynet/go-llama.cpp

| Limitation | Description |
|------------|-------------|
| **GGUF only** | GGML format no longer supported; use [pre-gguf](https://github.com/go-skynet/go-llama.cpp/releases/tag/pre-gguf) tag if you must use GGML. |
| **CGO required** | Build needs CGO; no pure-Go fallback. Cross-compilation is nontrivial. |
| **Build complexity** | Requires git submodules, building `libbinding.a`, and `LIBRARY_PATH` / `C_INCLUDE_PATH`. |
| **Backend-specific flags** | Metal, CUDA, ROCm, OpenBLAS, OpenCL need different build flags and linker options. |
| **Version lag** | New architectures and features appear in llama.cpp first; go-llama.cpp may lag until submodule updates. |
| **CuBLAS/build issues** | Some users report CuBLAS compilation errors; OpenBLAS or CPU builds are more reliable on Linux. |

---

## Quick Reference

| Need | Action |
|------|--------|
| **Recommended quant** | Q4_K_M or Q5_K_M |
| **16 GB RAM** | 7B Q4_K_M or Q5_K_M |
| **macOS M1/M2/M3** | `BUILD_TYPE=metal make libbinding.a` |
| **Reuse Ollama models** | Use blob path from `~/.ollama/models/blobs/` |
| **Find models** | Hugging Face (search: gguf), TheBloke, Ollama library |

---

*Last updated: 2026-02. See [LLAMACPP_EVALUATION.md](LLAMACPP_EVALUATION.md) for integration design.*

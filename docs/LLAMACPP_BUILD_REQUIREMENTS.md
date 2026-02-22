# llama.cpp Build Requirements

**Tag hints:** `#llamacpp` `#build`

Build guide for exarp-go with llama.cpp support via [go-llama.cpp](https://github.com/go-skynet/go-llama.cpp).

---

## Prerequisites

| Requirement | Version | Notes |
|-------------|---------|-------|
| Go | 1.25+ | `go version` |
| C compiler | gcc or clang | CGO requires a C toolchain |
| git | any | To clone go-llama.cpp |
| make | any | go-llama.cpp uses Makefile |
| CMake | 3.14+ | Optional; needed for CUDA builds |

### macOS (Metal)

Xcode Command Line Tools provide clang and Metal framework:

```bash
xcode-select --install
```

### Linux (CUDA)

Install CUDA toolkit and a C compiler:

```bash
sudo apt install build-essential
# CUDA toolkit from https://developer.nvidia.com/cuda-downloads
```

---

## Building exarp-go with llamacpp

### Step 1: Build go-llama.cpp

```bash
git clone --recurse-submodules https://github.com/go-skynet/go-llama.cpp
cd go-llama.cpp
make libbinding.a
```

For **Metal** (macOS Apple Silicon):

```bash
make libbinding.a BUILD_TYPE=metal
```

For **CUDA** (Linux/Windows with NVIDIA GPU):

```bash
CMAKE_ARGS="-DLLAMA_CUBLAS=on" make libbinding.a
```

### Step 2: Set environment variables

```bash
export LLAMACPP_DIR=/path/to/go-llama.cpp
export C_INCLUDE_PATH=$LLAMACPP_DIR
export LIBRARY_PATH=$LLAMACPP_DIR
```

### Step 3: Build exarp-go

```bash
CGO_ENABLED=1 go build -tags llamacpp -o bin/exarp-go ./cmd/server
```

Or via Makefile (when target is added):

```bash
make build-llamacpp
```

---

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `LLAMACPP_DIR` | Yes (build) | Path to go-llama.cpp clone with compiled `libbinding.a` |
| `LLAMACPP_MODEL_PATH` | No | Default directory for GGUF model files. Falls back to `~/.cache/llamacpp/models` |
| `C_INCLUDE_PATH` | Yes (build) | Must include `$LLAMACPP_DIR` for CGO headers |
| `LIBRARY_PATH` | Yes (build) | Must include `$LLAMACPP_DIR` for CGO linking |

---

## Model Files

llama.cpp uses GGUF model files. Place them in `$LLAMACPP_MODEL_PATH` or specify the full path when calling the tool.

Popular sources:
- [Hugging Face](https://huggingface.co/models?library=gguf) — search for GGUF quantizations
- TheBloke's quantizations on Hugging Face

Recommended quantizations for local use:
- **Q4_K_M** — Good balance of quality and speed (~4 bits per weight)
- **Q5_K_M** — Higher quality, ~25% more memory
- **Q8_0** — Near-original quality, ~2x memory of Q4

---

## Memory Management

The `ModelManager` handles model loading with configurable limits:

| Setting | Default | Description |
|---------|---------|-------------|
| `max_loaded_models` | 2 | Maximum models held in memory simultaneously |
| `max_memory_bytes` | 0 (unlimited) | Hard memory cap; evicts LRU models when exceeded |

Models are loaded on first use and cached. When limits are reached, the least-recently-used model is unloaded. Configure via tool arguments:

```json
{
  "action": "generate",
  "model": "mistral-7b-q4_k_m.gguf",
  "prompt": "Hello",
  "max_loaded_models": 1,
  "max_memory_bytes": 8589934592
}
```

---

## Verifying the Build

```bash
./bin/exarp-go -tool llamacpp -args '{"action": "status"}'
```

Expected output when built with the tag:

```json
{
  "available": true,
  "build_tag": "llamacpp",
  "model_path": "/path/to/models"
}
```

Without the build tag, the tool returns `available: false`.

---

## Troubleshooting

### "undefined: llama" or CGO linking errors

Ensure `LLAMACPP_DIR`, `C_INCLUDE_PATH`, and `LIBRARY_PATH` all point to the go-llama.cpp directory containing `libbinding.a`.

### "Metal framework not found" (macOS)

Install Xcode Command Line Tools: `xcode-select --install`. Ensure you built go-llama.cpp with `BUILD_TYPE=metal`.

### Model loading fails with "mmap" errors

The GGUF file may be corrupted or incompatible. Re-download from the source. Ensure sufficient RAM for the model size.

### Out of memory

Reduce `max_loaded_models` to 1, or use a smaller quantization (Q4_K_M instead of Q8_0). Set `max_memory_bytes` to limit total model memory.

### Build tag not applied

Verify with: `go build -v -tags llamacpp ./cmd/server 2>&1 | grep llamacpp`. The llamacpp provider files should appear in the build output.

---

## Platform Support

| Platform | GPU Acceleration | Status |
|----------|-----------------|--------|
| macOS (Apple Silicon) | Metal | Supported |
| macOS (Intel) | None (CPU only) | Supported |
| Linux (x86_64) | CUDA | Supported |
| Linux (x86_64) | CPU only | Supported |
| Windows | CUDA | Untested |

---

## See Also

- `docs/LLAMACPP_BENCHMARKS.md` — Performance comparison vs Ollama
- `docs/LLM_EXPOSURE_OPPORTUNITIES.md` — LLM abstraction overview
- `docs/LLAMACPP_TOOL_SCHEMA.md` — Tool schema and actions

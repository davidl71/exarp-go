# llama.cpp Build Requirements

**Tags:** `#llamacpp` `#build` `#cgo`

Build guide for exarp-go with llama.cpp support via [go-llama.cpp](https://github.com/go-skynet/go-llama.cpp).

---

## Prerequisites

| Requirement | Minimum Version | Check | Notes |
|-------------|-----------------|-------|-------|
| Go | 1.22+ | `go version` | CGO support required |
| C compiler | gcc 9+ or clang 10+ | `gcc --version` / `clang --version` | CGO requires a C toolchain |
| git | any | `git --version` | To initialize submodule |
| make | GNU Make 3.81+ | `make --version` | go-llama.cpp uses Makefile |
| CMake | 3.14+ | `cmake --version` | Required for CUDA builds; optional for Metal |

---

## Platform-Specific Prerequisites

### macOS (Apple Silicon — Metal GPU acceleration)

Xcode Command Line Tools provide `clang`, `swiftc`, and the Metal/Foundation frameworks:

```bash
xcode-select --install
```

Verify the install:

```bash
clang --version      # should show Apple clang
xcrun --show-sdk-path  # should print an SDK path
```

Metal support is available on Apple Silicon (M1/M2/M3/M4) and Intel Macs with AMD/Intel GPUs. Apple Silicon is strongly recommended for best performance.

For `make build-apple-fm` (Apple Foundation Models Swift bridge), you additionally need:

- macOS running on arm64 (Apple Silicon)
- `swiftc` (part of Xcode / Xcode Command Line Tools)
- `xcrun` (part of Xcode / Xcode Command Line Tools)

### Linux (CUDA GPU acceleration)

Install a C/C++ toolchain and the CUDA toolkit:

```bash
# Debian/Ubuntu
sudo apt install build-essential cmake

# RHEL/CentOS/Fedora
sudo dnf install gcc gcc-c++ cmake make

# CUDA toolkit — download from https://developer.nvidia.com/cuda-downloads
# Minimum CUDA version: 11.x
```

Verify:

```bash
gcc --version
cmake --version
nvcc --version   # optional, for CUDA verification
```

### Linux (CPU only)

```bash
# Debian/Ubuntu
sudo apt install build-essential

# RHEL/CentOS/Fedora
sudo dnf install gcc gcc-c++ make
```

---

## Environment Variables

### Build-time variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `CGO_ENABLED` | Yes | Auto-detected | Must be `1` for llama.cpp; auto-set by `make build-llamacpp` |
| `LLAMACPP_DIR` | No | `./go-llama.cpp` | Path to go-llama.cpp directory containing `libbinding.a` |
| `CGO_CPPFLAGS` | Auto-set | See below | Include paths for CGO; set by `make build-llamacpp` |
| `CGO_LDFLAGS` | Auto-set | See below | Linker flags for CGO; set by `make build-llamacpp` |
| `BUILD_TYPE` | No | (empty) | Set to `metal` for Metal GPU acceleration on macOS |

The Makefile sets these automatically:

```
CGO_CPPFLAGS="-I<LLAMACPP_DIR> -I<LLAMACPP_DIR>/llama.cpp -I<LLAMACPP_DIR>/llama.cpp/common"
CGO_LDFLAGS="-L<LLAMACPP_DIR> -lbinding -lstdc++ -framework Foundation -framework Metal -framework MetalKit -framework MetalPerformanceShaders"
```

### Runtime variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `LLAMACPP_MODEL_PATH` | No | `~/.cache/llamacpp/models` | Default directory for GGUF model files |

---

## Build Steps

### Step 1: Initialize the submodule

The `go-llama.cpp` submodule must be initialized before building:

```bash
git submodule update --init --recursive
```

If the `go-llama.cpp` directory is empty after cloning, this step was skipped.

### Step 2: Build libbinding.a

Use the Makefile target (auto-detects platform):

```bash
make build-libbinding
```

Or manually:

**macOS (Metal — Apple Silicon recommended):**

```bash
cd go-llama.cpp
BUILD_TYPE=metal make libbinding.a
```

**Linux (CUDA):**

```bash
cd go-llama.cpp
CMAKE_ARGS="-DLLAMA_CUBLAS=on" make libbinding.a
```

**Linux (CPU only):**

```bash
cd go-llama.cpp
make libbinding.a
```

### Step 3: Build exarp-go with llama.cpp support

```bash
make build-llamacpp
```

This is equivalent to:

```bash
CGO_ENABLED=1 \
CGO_CPPFLAGS="-I./go-llama.cpp -I./go-llama.cpp/llama.cpp -I./go-llama.cpp/llama.cpp/common" \
CGO_LDFLAGS="-L./go-llama.cpp -lbinding -lstdc++ -framework Foundation -framework Metal -framework MetalKit -framework MetalPerformanceShaders" \
go build -tags llamacpp -o bin/exarp-go ./cmd/server
```

Note: The `-framework` flags are macOS-only. On Linux, the linker flags will differ.

---

## Apple Foundation Models (build-apple-fm)

`make build-apple-fm` builds a different feature: Apple's on-device Foundation Models API (not llama.cpp). It requires:

- macOS + Apple Silicon (arm64)
- Xcode (not just Command Line Tools) for full Swift compiler
- `swiftc` and `xcrun` in PATH

```bash
make build-apple-fm
```

This first runs `make build-swift-bridge` (compiles `FoundationModelsShim.swift` → `libFMShim.a`), then builds the Go binary with `CGO_ENABLED=1`.

---

## Verifying the Build

```bash
./bin/exarp-go -tool llamacpp -args '{"action": "status"}'
```

Expected output when built with the `llamacpp` tag:

```json
{
  "available": true,
  "build_tag": "llamacpp",
  "model_path": "/path/to/models"
}
```

Without the build tag, the tool returns `"available": false`.

To confirm the build tag was applied:

```bash
go build -v -tags llamacpp ./cmd/server 2>&1 | grep llamacpp
```

The llamacpp provider files should appear in the output.

---

## Model Files

llama.cpp uses GGUF model files. Place them in `$LLAMACPP_MODEL_PATH` (default: `~/.cache/llamacpp/models`) or specify the full path when calling the tool.

Popular sources:
- [Hugging Face](https://huggingface.co/models?library=gguf) — search for GGUF quantizations
- TheBloke's quantizations on Hugging Face

Recommended quantizations for local use:

| Quantization | Memory | Quality | Notes |
|-------------|--------|---------|-------|
| Q4_K_M | ~4 bits/weight | Good | Best balance for most hardware |
| Q5_K_M | ~5 bits/weight | Better | ~25% more memory than Q4_K_M |
| Q8_0 | ~8 bits/weight | Near-original | ~2x memory of Q4_K_M |

---

## Common Build Failures and Fixes

### Submodule not initialized

**Symptom:** `go-llama.cpp directory not found` or `libbinding.a not found`

**Fix:**

```bash
git submodule update --init --recursive
```

### cmake not found

**Symptom:** `cmake: command not found` during `make libbinding.a`

**Fix (macOS):**

```bash
brew install cmake
```

**Fix (Linux):**

```bash
sudo apt install cmake   # Debian/Ubuntu
sudo dnf install cmake   # RHEL/Fedora
```

### CGO not available

**Symptom:** `❌ CGO is not available - required for llama.cpp`

**Fix (macOS):**

```bash
xcode-select --install
```

**Fix (Linux):**

```bash
sudo apt install build-essential   # Debian/Ubuntu
sudo dnf install gcc gcc-c++       # RHEL/Fedora
```

### "undefined: llama" or CGO linking errors

**Symptom:** Link errors referencing undefined symbols from `llama`

**Fix:** Ensure `LLAMACPP_DIR`, `CGO_CPPFLAGS`, and `CGO_LDFLAGS` all point to the go-llama.cpp directory containing the compiled `libbinding.a`. Use `make build-llamacpp` which sets these automatically.

### "Metal framework not found" (macOS)

**Symptom:** Linker error about missing `-framework Metal`

**Fix:** Install Xcode Command Line Tools and ensure you built `libbinding.a` with `BUILD_TYPE=metal`:

```bash
xcode-select --install
cd go-llama.cpp && BUILD_TYPE=metal make libbinding.a
```

### Swift bridge build failed (Apple FM)

**Symptom:** `⚠️ Swift bridge build failed - Apple FM will not be available`

**Fix:** This is non-fatal; the binary still builds, just without Apple Foundation Models. To enable it:

1. Ensure you are on macOS with Apple Silicon (arm64)
2. Install full Xcode (not just Command Line Tools): `xcode-select --install` or via App Store
3. Verify: `swiftc --version` and `xcrun --show-sdk-path`
4. Re-run: `make build-apple-fm`

### Model loading fails with "mmap" errors

**Symptom:** Runtime error when loading a GGUF model

**Fix:** The GGUF file may be corrupted or incompatible. Re-download from the source. Ensure sufficient RAM for the model size.

### Out of memory during model load

**Fix:** Reduce `max_loaded_models` to 1, or use a smaller quantization (Q4_K_M instead of Q8_0). Set `max_memory_bytes` to limit total model memory:

```json
{
  "action": "generate",
  "model": "mistral-7b-q4_k_m.gguf",
  "max_loaded_models": 1,
  "max_memory_bytes": 8589934592
}
```

### Build tag not applied (tool reports available: false)

**Symptom:** Binary built but `llamacpp` tool returns `available: false`

**Fix:** The binary was not built with the `llamacpp` build tag. Use:

```bash
make build-llamacpp
```

Not `make build` or `make b` (those do not include the `llamacpp` tag).

---

## Platform Support Matrix

| Platform | Architecture | GPU Acceleration | Status |
|----------|-------------|-----------------|--------|
| macOS | Apple Silicon (arm64) | Metal | Supported (recommended) |
| macOS | Intel (amd64) | None (CPU only) | Supported |
| Linux | x86_64 | CUDA | Supported |
| Linux | x86_64 | CPU only | Supported |
| Linux | arm64 | CPU only | Untested |
| Windows | x86_64 | CUDA | Untested |

---

## Memory Management

The `ModelManager` handles model loading with configurable limits:

| Setting | Default | Description |
|---------|---------|-------------|
| `max_loaded_models` | 2 | Maximum models held in memory simultaneously |
| `max_memory_bytes` | 0 (unlimited) | Hard memory cap; evicts LRU models when exceeded |

Models are loaded on first use and cached. When limits are reached, the least-recently-used model is unloaded.

---

## See Also

- `docs/LLAMACPP_BENCHMARKS.md` — Performance comparison vs Ollama
- `docs/LLAMACPP_TOOL_SCHEMA.md` — Tool schema and actions
- `docs/GGUF_COMPATIBILITY.md` — GGUF model compatibility notes
- `docs/LLM_EXPOSURE_OPPORTUNITIES.md` — LLM abstraction overview
- `docs/CGO_BUILD_PARITY.md` — CGO build parity across platforms

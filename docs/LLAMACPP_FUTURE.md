# llamacpp Integration - Future Enhancement

## Status: Deferred

The llamacpp integration for local GGUF model inference has been **deferred to future development**.

## Why Deferred?

1. **Complexity**: Requires CMake, C++ compilation, git submodules, and complex CGO configuration
2. **Build Time**: Estimated 8-12 hours of development and testing
3. **Ollama Already Works**: The existing Ollama integration provides excellent local LLM support
4. **Maintenance Burden**: Adds significant build complexity for limited additional value

## Current State

The codebase includes **optional llamacpp support** via build tags:
- Code exists in `internal/tools/llamacpp*.go`
- Disabled by default (uses `llamacpp_nocgo.go` stub)
- Can be enabled with `-tags llamacpp,cgo` when built

## Ollama is Recommended

For local LLM inference, use **Ollama** instead:
- ✅ Easy installation: `brew install ollama`
- ✅ Automatic GPU detection (Metal on Mac, CUDA on Linux)
- ✅ Model management: `ollama pull llama3.2`
- ✅ Already integrated in exarp-go
- ✅ Battle-tested and maintained

## Future Implementation (If Needed)

If llamacpp integration becomes necessary, see the comprehensive task breakdown in `IMPLEMENTATION_PLAN.md` Phase 2.

**Required Steps**:
1. Add llama.cpp as git submodule
2. Create CMake build scripts
3. Build libbinding.a for target platform
4. Configure CGO linker flags
5. Add Makefile targets for conditional build
6. Test on macOS (Metal), Linux (CUDA), CPU-only
7. Document build requirements

**Estimated Effort**: 2-3 days (8-12 hours)

## Build Requirements (For Future)

If implementing llamacpp:
- CMake 3.12+
- C++11 compiler (Xcode CLI tools on Mac)
- Git (for submodules)
- Optional: CUDA toolkit (Linux GPU)
- Optional: Metal Performance Shaders (Mac, automatic)

## Alternative: Direct llama.cpp

Advanced users can run llama.cpp directly:
```bash
git clone https://github.com/ggerganov/llama.cpp
cd llama.cpp
make
./main -m models/llama-7b.gguf -p "Hello"
```

Then use exarp-go with Ollama or HTTP gateway for integration.

---

**Recommendation**: Use Ollama. It provides everything llamacpp would offer with none of the complexity.

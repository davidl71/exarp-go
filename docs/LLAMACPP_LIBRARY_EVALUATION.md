# LlamaCPP Library Evaluation

**Date:** 2026-02-20  
**Task:** T-1771367215872916000  
**Decision:** Choose Go bindings for llama.cpp integration

---

## Executive Summary

**Recommendation:** **tcpipuk/llama-go** (with go-skynet/go-llama.cpp as fallback)

**Rationale:**
- Most actively maintained (last commit: days ago vs months for others)
- Production-ready: 400+ tests, comprehensive API, Model/Context separation
- Advanced features: chat, embeddings, speculative decoding, streaming
- Better documentation and examples
- Clean, idiomatic Go API with functional options pattern
- Active fork of go-skynet (which hasn't been updated since Oct 2023)

---

## Library Comparison

### 1. tcpipuk/llama-go ⭐ **RECOMMENDED**

**Repository:** https://github.com/tcpipuk/llama-go  
**Stars:** 11 (fork of go-skynet)  
**Last Updated:** Active (2026)  
**Maintenance:** Excellent - actively tracking llama.cpp releases

#### Pros
✅ **Most actively maintained** - Regular updates, tracks upstream llama.cpp  
✅ **Production-ready** - 400+ test cases, CI with CUDA builds  
✅ **Complete feature set:**
   - Chat completion with automatic template formatting
   - Text generation (raw)
   - Embeddings for semantic search
   - Speculative decoding (2-3× speedup)
   - Streaming via callbacks or channels
   - Prefix caching (KV cache reuse)
   
✅ **Clean architecture:**
   - Model/Context separation (load weights once, multiple contexts)
   - Thread-safe concurrent inference
   - Functional options pattern (ModelOption, ContextOption)
   - Explicit resource management with finalizers
   
✅ **GPU support:** Metal, CUDA, ROCm, SYCL, Vulkan, OpenCL, RPC  
✅ **Excellent documentation:**
   - Getting started guide
   - API guide
   - Build options
   - Migration guide (v1 → v2)
   - pkg.go.dev with examples
   
✅ **Idiomatic Go API:**
```go
model, _ := llama.LoadModel("/path/to/model.gguf", 
    llama.WithGPULayers(-1))
defer model.Close()

ctx, _ := model.NewContext(llama.WithContext(2048))
defer ctx.Close()

// Chat completion
messages := []llama.ChatMessage{
    {Role: "system", Content: "You are helpful"},
    {Role: "user", Content: "Hello"},
}
response, _ := ctx.Chat(context.Background(), messages)

// Or raw generation
text, _ := ctx.Generate("Hello", llama.WithMaxTokens(50))
```

#### Cons
❌ Smaller community (11 stars vs 871 for go-skynet)  
❌ Forked project (but upstream is abandoned)  
⚠️ API changed from v1 to v2 (migration guide provided)

#### Technical Details
- **Build:** Git submodule for llama.cpp, Makefile build system
- **CGO:** Clean wrapper.cpp/wrapper.h interface
- **Thread Safety:** One Context per goroutine, Model is thread-safe
- **Memory:** Model/Context separation enables efficient VRAM usage
- **Streaming:** Via callbacks or buffered channels (cgo.Handle)

---

### 2. go-skynet/go-llama.cpp

**Repository:** https://github.com/go-skynet/go-llama.cpp  
**Stars:** 871  
**Last Updated:** October 2023 (ABANDONED)  
**Maintenance:** None - no activity for 16 months

#### Pros
✅ Most popular (871 stars)  
✅ Well-known in community  
✅ GGUF format support (post-#180)  
✅ GPU acceleration: Metal, CUDA, ROCM, OpenCL  
✅ High-level bindings (minimal CGO overhead)  
✅ Used by go-skynet/llama-cli (OpenAI-compatible API)

#### Cons
❌ **ABANDONED** - No commits since October 2023  
❌ Stale llama.cpp version (16 months old)  
❌ Missing recent llama.cpp features  
❌ No active maintenance or bug fixes  
❌ API less complete than tcpipuk fork  
❌ Documentation outdated

#### Technical Details
- **Build:** Git submodule, Makefile with BUILD_TYPE=
- **API:** Load model → Predict() loop with callback
- **Acceleration:** Metal, CuBLAS, ROCM (HIPBLAS), OpenCL

**Verdict:** ⚠️ **DO NOT USE** - Superseded by tcpipuk/llama-go fork

---

### 3. swdunlop/llm-go

**Repository:** https://github.com/swdunlop/llm-go  
**Stars:** 4  
**Last Updated:** 2024 (intermittent)  
**Maintenance:** Low - minimal activity

#### Pros
✅ Simplest API - minimal wrapper  
✅ Includes HTTP server example  
✅ Basic load/predict workflow  
✅ Forked from early Ollama codebase  
✅ Nix flake support

#### Cons
❌ Very small community (4 stars)  
❌ Basic feature set (no chat, no embeddings, no speculative decoding)  
❌ Limited documentation  
❌ Minimal test coverage  
❌ Less idiomatic Go (derived from Ollama internals)  
❌ No Model/Context separation  
❌ No streaming support

#### API Example
```go
m, _ := llm.Load("model.gguf")
defer m.Close()

input := m.Encode("Once upon a time..")
s, _ := m.Predict(input)
defer s.Close()

var output []llm.Token
for {
    next, err := s.Next(nil)
    if err != nil { break }
    output = append(output, next)
}
fmt.Println(m.Decode(output))
```

**Verdict:** ⚠️ **Not Recommended** - Too basic, lacks features we need

---

## Feature Matrix

| Feature | tcpipuk/llama-go | go-skynet | swdunlop/llm-go |
|---------|:----------------:|:---------:|:---------------:|
| **Active Maintenance** | ✅ Excellent | ❌ Abandoned | ⚠️ Minimal |
| **Stars/Community** | 11 (new) | 871 (old) | 4 |
| **Last Update** | 2026 (days) | Oct 2023 | 2024 (months) |
| **llama.cpp Version** | Current | 16mo old | Unknown |
| **Test Coverage** | ✅ 400+ tests | ⚠️ Some | ❌ Minimal |
| **Documentation** | ✅ Excellent | ⚠️ Outdated | ❌ Basic |
| **Chat Completion** | ✅ With templates | ❌ | ❌ |
| **Text Generation** | ✅ | ✅ | ✅ |
| **Embeddings** | ✅ | ❌ | ❌ |
| **Speculative Decoding** | ✅ | ❌ | ❌ |
| **Streaming** | ✅ Callbacks + channels | ⚠️ Callback only | ❌ |
| **Prefix Caching** | ✅ KV cache reuse | ❌ | ❌ |
| **Model/Context Separation** | ✅ | ❌ | ❌ |
| **Thread Safety** | ✅ Documented | ⚠️ Unclear | ❌ |
| **GPU: Metal** | ✅ | ✅ | ⚠️ |
| **GPU: CUDA** | ✅ | ✅ | ⚠️ |
| **GPU: ROCm** | ✅ | ✅ | ❌ |
| **GPU: SYCL** | ✅ | ❌ | ❌ |
| **GPU: Vulkan** | ✅ | ❌ | ❌ |
| **Build Complexity** | Medium | Medium | Low |
| **API Cleanliness** | ✅ Idiomatic Go | ⚠️ Functional | ⚠️ Low-level |
| **Migration Path** | ✅ Guide provided | N/A | N/A |

---

## Decision Criteria

### 1. Maintenance & Longevity ⭐ CRITICAL
**Winner:** tcpipuk/llama-go
- Active development (commits within days)
- Tracks llama.cpp releases
- go-skynet abandoned 16 months ago
- swdunlop minimal activity

### 2. Feature Completeness
**Winner:** tcpipuk/llama-go
- Chat, generation, embeddings, speculative decoding
- Streaming, prefix caching, thread safety
- go-skynet lacks chat/embeddings
- swdunlop very basic

### 3. Production Readiness
**Winner:** tcpipuk/llama-go
- 400+ test cases with CI
- Documented thread safety
- Clean resource management
- Others lack comprehensive testing

### 4. API Quality
**Winner:** tcpipuk/llama-go
- Idiomatic Go (functional options, context.Context)
- Model/Context separation
- Clear error handling
- Migration guide for API changes

### 5. GPU Support
**Winner:** tcpipuk/llama-go
- 8 backends (Metal, CUDA, ROCm, SYCL, Vulkan, OpenCL, OpenBLAS, RPC)
- go-skynet has 4 backends
- swdunlop unclear support

### 6. Documentation
**Winner:** tcpipuk/llama-go
- Getting started guide, API guide, build docs, migration guide
- Complete pkg.go.dev with examples
- go-skynet docs outdated
- swdunlop minimal docs

---

## Recommended Decision

### Primary Choice: **tcpipuk/llama-go**

**Rationale:**
1. **Active maintenance** - Critical for production use
2. **Complete feature set** - Everything we need (chat, embeddings, speculative)
3. **Production-ready** - 400+ tests, documented thread safety
4. **Best API** - Idiomatic Go, Model/Context separation
5. **GPU support** - 8 backends including Metal for macOS
6. **Documentation** - Excellent guides and examples

**Trade-off:** Smaller community (11 stars), but fork of 871-star project that's abandoned.

### Fallback: **go-skynet/go-llama.cpp**

**Only if:** tcpipuk fork causes issues (unlikely)

**Why keep as fallback:**
- Large community (871 stars)
- Well-tested API (though old)
- Known to work (if we don't need recent features)

**Recommended:** Start with tcpipuk, reference go-skynet if needed

### NOT Recommended: **swdunlop/llm-go**

**Reasons:**
- Too basic (no chat, embeddings, speculative decoding)
- Minimal community and maintenance
- Lacks features we need for production

---

## Implementation Plan

### Phase 1: Prototype with tcpipuk/llama-go (2-3 hours)

1. **Clone and build:**
```bash
git clone --recurse-submodules https://github.com/tcpipuk/llama-go
cd llama-go
make libbinding.a
```

2. **Test basic generation:**
```bash
export LIBRARY_PATH=$PWD C_INCLUDE_PATH=$PWD LD_LIBRARY_PATH=$PWD
go run ./examples/simple -m /path/to/model.gguf -p "Hello" -n 50
```

3. **Test with Ollama model:**
   - Load GGUF from `~/.ollama/models/blobs/`
   - Verify Metal acceleration on macOS
   - Check performance vs Ollama HTTP

4. **Create minimal integration:**
```go
// internal/llm/llamacpp.go
package llm

import (
    llama "github.com/tcpipuk/llama-go"
)

type LlamaCPPProvider struct {
    model *llama.Model
}

func (p *LlamaCPPProvider) Generate(ctx context.Context, prompt string) (string, error) {
    // Implement TextGenerator interface
}
```

### Phase 2: Evaluate vs Ollama (1 hour)

1. **Performance comparison:**
   - Load time (cold start)
   - Generation latency (first token, tokens/sec)
   - Memory usage
   - GPU utilization (Metal)

2. **Feature parity:**
   - Chat completion
   - Streaming
   - Multiple concurrent requests

3. **Decision point:** Proceed if:
   - Performance within 20% of Ollama
   - Build complexity acceptable
   - API meets our needs

### Phase 3: Proceed with Integration (IF prototype successful)

Continue with Phase 2-6 from llamacpp-integration.plan.md

---

## Risk Assessment

### Low Risk ✅
- **CGO build** - Well-documented, tested on macOS/arm64
- **API stability** - v2 API is stable, migration guide provided
- **Thread safety** - Clearly documented, one Context per goroutine
- **GPU support** - Metal support confirmed, widely used

### Medium Risk ⚠️
- **Community size** - Small (11 stars), but active maintainer
- **Fork maintenance** - Depends on one maintainer (tcpipuk)
- **API changes** - v1→v2 broke compatibility (but rare)

### Mitigation Strategies
1. **Vendor the dependency** - Lock to known-good version
2. **Test thoroughly** - Comprehensive test suite before production
3. **Monitor upstream** - Watch for llama.cpp and fork updates
4. **Fallback plan** - go-skynet available if fork abandoned

---

## Questions & Answers

### Q: Why fork instead of upstream?
**A:** Upstream (go-skynet) abandoned 16 months ago. Fork is only maintained option.

### Q: What if tcpipuk abandons the fork?
**A:** Fallback to go-skynet (older but stable), or fork ourselves (maintainable).

### Q: Build complexity vs Ollama HTTP?
**A:** More complex (CGO, C++ compiler), but no external dependency, better performance.

### Q: Windows support?
**A:** Start with macOS/Linux. Windows CGO is complex, evaluate later if needed.

### Q: Can we switch libraries later?
**A:** Yes - internal abstraction (TextGenerator interface) allows swapping implementations.

---

## References

- **tcpipuk/llama-go:** https://github.com/tcpipuk/llama-go
- **go-skynet/go-llama.cpp:** https://github.com/go-skynet/go-llama.cpp
- **swdunlop/llm-go:** https://github.com/swdunlop/llm-go
- **llama.cpp upstream:** https://github.com/ggerganov/llama.cpp
- **exarp-go ecosystem doc:** `docs/GO_AI_ECOSYSTEM.md` Section 9.1
- **Integration plan:** `.cursor/plans/llamacpp-integration.plan.md`

---

## Approval

**Recommended Action:** Proceed with tcpipuk/llama-go prototype (Task 2: Test build)

**Decision Maker:** Project maintainer  
**Timeline:** Prototype in 2-3 hours, decision within 1 day  
**Next Step:** Mark Task 1 Done, start Task 2 (Test build on macOS/arm64)

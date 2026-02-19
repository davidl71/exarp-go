# Go AI Ecosystem

**Purpose:** Reference for the Go AI landscape (frameworks, provider SDKs, ML/numerics), what exarp-go uses today, and how exarp-go could improve by aligning with or adopting from this ecosystem.

**See also:** [LLM_NATIVE_ABSTRACTION_PATTERNS.md](LLM_NATIVE_ABSTRACTION_PATTERNS.md), [LLM_EXPOSURE_OPPORTUNITIES.md](LLM_EXPOSURE_OPPORTUNITIES.md). Section 9 covers Transformers-style equivalents (Go/Swift) and LLM middleware.

---

## 1. General AI Application Frameworks

Frameworks that orchestrate AI models and services for complex applications and agentic workflows.

| Framework | Description |
|-----------|-------------|
| **langchaingo** | Go implementation of the Python LangChain framework; applications powered by LLMs (chains, agents, tools). |
| **Genkit Go** | Open-source framework by Google; unified API for AI apps across multiple model providers, multimodal content, tool calling. |
| **Ollama** | Run large language models locally; useful for offline or self-hosted AI. (exarp-go integrates with Ollama via HTTP API + optional Python bridge.) |
| **LocalAI** | Open-source, self-hosted alternative to the OpenAI API; run AI models on your own infrastructure (OpenAI-compatible API). |
| **Jetify AI** (Go AI SDK) | Unified AI SDK for Go ([github.com/jetify-com/ai](https://github.com/jetify-com/ai)); multi-provider (OpenAI, Anthropic), streaming, tool calling, embeddings, structured output; Vercel AI SDK–style API. Public alpha, Apache-2.0. |

---

## 2. Provider-Specific SDKs

Official and community SDKs for specific AI service providers.

| SDK | Description |
|-----|-------------|
| **Google Generative AI Go SDK** | Official Go package for Google generative models (e.g. Gemini); text and multimodal inputs. |
| **Go OpenAI** | Community-driven Go client for the OpenAI APIs. |
| **Go Anthropic** | Community-driven wrapper for the Anthropic Claude API. |
| **Go HuggingFace** | Inference client for the Hugging Face model repository. |

---

## 3. Machine Learning & Numerical Computation

Libraries for core ML algorithms, numerical computation, and related tasks.

| Library | Description |
|---------|-------------|
| **Gorgonia** | Deep learning and numerical computation with a tensor-based engine. |
| **GoLearn** | Classical machine learning algorithms and data handling. |
| **Gonum** | Numerical computing: matrices, linear algebra, statistics, graphs. **Used by exarp-go** for task graphs and statistics. |
| **GoCV** | Computer vision using the OpenCV framework. |
| **TensorFlow Go bindings** | Load and execute pre-trained TensorFlow models in Go. |

---

## 4. What exarp-go Uses Today

### 4.1 LLM / generative text

exarp-go uses a **custom LLM abstraction** (no langchaingo, Genkit, or Go AI SDK). Backends:

| Backend | Role | Implementation |
|--------|------|-----------------|
| **Apple Foundation Models (FM)** | On-device generation (darwin/arm64, CGO). Used by task_analysis, context, estimation, task_workflow, task_discovery, report fallback. | `internal/tools/fm_apple.go`, `fm_provider.go`; `DefaultFMProvider()`. |
| **Ollama** | Local LLM server (status, models, generate, pull, etc.). | Native Go HTTP client first; Python bridge fallback. `internal/tools/ollama_native.go`, `ollama_provider.go`; `DefaultOllama()`. |
| **MLX** | Report/scorecard insights (long-form text). | Python bridge; `DefaultReportInsight()` tries MLX then FM. `internal/tools/mlx_invoke.go`, `insight_provider.go`. |
| **LocalAI** | Self-hosted OpenAI-compatible API (optional). | Env: `LOCALAI_BASE_URL` (required), `LOCALAI_MODEL` (optional). `internal/tools/localai_provider.go`; `text_generate` with `provider=localai`. |

- **Unified entry:** `text_generate` tool with `provider: fm | insight | mlx | localai`.
- **Discovery:** `stdio://models` exposes `backends` (e.g. `fm_available`, tool names). See [LLM_EXPOSURE_OPPORTUNITIES.md](LLM_EXPOSURE_OPPORTUNITIES.md).

### 4.2 Numerical / graph

| Component | Use in exarp-go |
|-----------|------------------|
| **Gonum** | Task dependency graphs (`graph/simple`, `graph/topo`), cycle detection, critical path; statistics (`gonum.org/v1/gonum/stat`). See `internal/tools/graph_helpers.go`, `task_analysis_shared.go`, `statistics.go`. |

exarp-go does **not** use Gorgonia, GoLearn, GoCV, or TensorFlow.

---

## 5. Improvement Opportunities

Concrete ways exarp-go could improve by leveraging or aligning with the Go AI ecosystem.

### 5.1 Orchestration and agentic flows

- **Current:** Custom provider abstraction (FM, Ollama, MLX) and MCP tools; no chain/agent framework.
- **Opportunity:** If exarp-go adds more multi-step agentic flows (e.g. plan → execute → reflect), consider:
  - **langchaingo** for chains, agents, and tool use patterns.
  - **Go AI SDK** for a standard agent interface and streaming, if the project wants parity with the Vercel AI SDK.
- **Recommendation:** Keep the current abstraction for “single-call” generation; evaluate langchaingo or Go AI SDK only when adding explicit agent workflows (and prefer minimal dependency).

**Evaluation (langchaingo / Jetify AI):** As of 2026-02, exarp-go uses a minimal custom abstraction (FM, Ollama, MLX, LocalAI) and MCP tools for orchestration. **langchaingo** would add chains/agents and Python LangChain parity but increases dependency surface and overlaps with existing MCP tool flows. **Jetify AI** ([github.com/jetify-com/ai](https://github.com/jetify-com/ai)) provides Vercel AI SDK–style unified API and streaming. **Decision:** Do not adopt either until we add explicit multi-step agent workflows (plan → execute → reflect) that benefit from a framework; then re-evaluate with a small proof-of-concept (Jetify AI is the primary candidate for a unified Go SDK). See task T-1771252268378.

### 5.2 Cloud and self-hosted backends

- **Current:** Local-only (Apple FM, Ollama, MLX) plus optional **LocalAI** (OpenAI-compatible). No direct use of Google, OpenAI, or Anthropic SDKs.
- **Opportunity:**
  - **LocalAI:** Add an optional backend that talks to a LocalAI server (OpenAI-compatible API). Gives users a self-hosted “OpenAI-style” option alongside Ollama.
  - **Provider SDKs (optional):** For future “cloud fallback” or optional features, consider **Go OpenAI**, **Google Generative AI Go SDK**, or **Go Anthropic** behind the same provider interface. Not required for current scope.
- **LocalAI (implemented):** Optional backend via `LOCALAI_BASE_URL` and optional `LOCALAI_MODEL`; use `text_generate` with `provider=localai`. Discovery: `stdio://models` includes `localai_available` and `localai_tool`.
- **Opportunity:** Provider SDKs (optional) for future cloud fallback: **Go OpenAI**, **Google Generative AI Go SDK**, or **Go Anthropic** behind the same provider interface. Not required for current scope.
- **Recommendation:** Defer cloud SDKs until there is a concrete need.

### 5.3 LLM abstraction and discovery

- **Current:** `stdio://models` and tool hints already expose backends; `text_generate` provides a single entry point.
- **Opportunity:**
  - Reference this ecosystem doc and [LLM_NATIVE_ABSTRACTION_PATTERNS.md](LLM_NATIVE_ABSTRACTION_PATTERNS.md) from the main README or “Architecture” doc.
  - Optional dedicated resource (e.g. `stdio://llm/status`) if clients need a separate LLM-status endpoint.
- **Recommendation:** Add a short “AI/LLM” section in the main docs that links to this file and the LLM docs.

**Decision (stdio://llm/status):** Deferred. The existing `stdio://models` resource already exposes backend status (`fm_available`, `localai_available`, tool names, hint). A separate `stdio://llm/status` resource would duplicate this unless we need a distinct endpoint for non-MCP clients. Revisit if such a requirement appears (see task T-1771252280227).

### 5.4 Numerical and graph stack

- **Current:** Gonum for task graphs and statistics; no other ML libraries.
- **Opportunity:** Gonum is already part of the ecosystem; no change needed for current use cases.
- **Recommendation:** Keep using Gonum; if future features need classical ML (e.g. clustering, regression beyond current stats), consider **GoLearn** or more Gonum packages before adding Gorgonia/TensorFlow.

### 5.5 Token estimation (nice-to-have)

- **Current:** Token counts are estimated with a **character ratio** (`tokens_per_char`, default `0.25`). Used by the **context** tool (budget analysis, summarization token estimates) and config (`thresholds.tokens_per_char`, `context.tokens_per_char`). See [pkg.go.dev](https://pkg.go.dev/github.com/tiktoken-go/tokenizer#section-documentation) and `internal/tools/context.go` (`estimateTokens`).
- **Nice-to-have:** **[tiktoken-go](https://pkg.go.dev/github.com/tiktoken-go/tokenizer)** — pure Go port of OpenAI's tokenizer. Provides:
  - **Accurate counts** for OpenAI-style APIs: `Codec.Count(text)` or `Encode`/`Decode` with encodings such as `Cl100kBase`, `O200kBase`, `R50kBase`. Useful when context or report payloads are sent to **OpenAI-compatible** endpoints (e.g. cloud API, LocalAI with same tokenization).
  - **Model-aware:** `tokenizer.ForModel(model)` for GPT-4, o1, etc.; aligns estimates with what the API actually counts.
- **Usefulness for exarp-go:**
  - **High** when integrating with **OpenAI API or OpenAI-compatible** services: budget and truncation decisions would match server-side limits.
  - **Low** for **local-only** (Apple FM, Ollama, MLX): those backends use different tokenization; the current ratio is a reasonable heuristic.
  - **Optional/feature-flag:** Use tiktoken when a config flag or provider indicates “OpenAI-compatible”; otherwise keep ratio-based estimation to avoid dependency and binary size impact.
- **Tradeoffs:** Library embeds ~4MB of vocabularies (compiled into the binary); MIT license; module at v0.7.0. Adds a direct dependency. Consider as an **optional** dependency or build tag so vendors who never call OpenAI-compatible APIs don’t pay the cost.
- **Recommendation:** Document as nice-to-have; consider adding when exarp-go gains first-class support for OpenAI (or compatible) API and context limits must match server-side token counts.

- **Would tiktoken help?** Yes, when the consumer is an **OpenAI or OpenAI-compatible** API. It would:
  - Make **pre-tokenized output** (e.g. `token_estimate` on report/prime/task list) match what the API actually counts, so AIs can budget and truncate accurately.
  - Improve **context** tool budget/summarize decisions and **scorecard** “tokenizability” and large-file thresholds when the user sends that data to an OpenAI-style endpoint.
  - For **local-only** (FM, Ollama, MLX), the ratio remains sufficient; tiktoken is optional.

- **What else could tiktoken give?**
  - **Exact truncation:** Truncate to exactly N tokens (Encode then take first N, Decode) so payloads never overshoot context limits; ratio-based truncation can overshoot.
  - **Token-boundary chunking:** Split text into fixed-size token chunks (e.g. 512 tokens) for RAG or batching, so boundaries match what the model sees (no mid-token cuts).
  - **Cost estimation:** Combine token count with $/token (per model) to estimate API cost before calling (e.g. for report or sprint flows).
  - **Multi-encoding / model-specific counts:** Report “tokens for GPT-4” vs “tokens for o1” (different encodings) so the AI can choose the right budget per model; `ForModel(model)` and encodings like `Cl100kBase`, `O200kBase` support this.
  - **Decode/inspect:** Decode token IDs to see token boundaries and special tokens; useful for debugging or “first N tokens” preview that respects tokenization.
  - **Special tokens:** Count or flag special tokens (BOS, EOS, etc.) when present, for prompt-engineering or prompt-size accuracy.

- **Scorecard use (tokenizability of code):** The **project scorecard** reports **estimated tokens** for the codebase (Go + Go test + bridge/tests Python files) using the same ratio (`tokens_per_char`). That gives a “tokenizability” metric: how many tokens the code would consume if sent to an LLM (e.g. for context budgeting or RAG). See `GoProjectMetrics.EstimatedTokens` and `TotalCodeBytes` in `internal/tools/scorecard_go.go`. **Optional:** If tiktoken-go is added (e.g. behind a build tag), the scorecard could expose an **OpenAI token count** for the same files for users who send code to OpenAI-compatible APIs.

- **Multi-stage logic for split/refactor candidates:** The scorecard uses a **three-stage pipeline** to find files that would benefit from splitting or refactoring:
  1. **Per-file stats:** `collectPerFileCodeStats(projectRoot)` walks all `.go`, `_test.go`, and bridge/tests `.py` files and records path, lines, bytes, and ratio-based estimated tokens per file.
  2. **Threshold filter:** `filterLargeFileCandidates(files, tokenThreshold, lineThreshold)` keeps only files that exceed the token threshold (default 6000 ≈ ~24KB) or the line threshold (default 500). These thresholds are tunable constants in `scorecard_go.go`.
  3. **Output:** The scorecard result includes `LargeFileCandidates` (list of `FileSizeInfo`). The formatted scorecard has a “Large files (consider splitting/refactoring for context fit)” section listing path, lines, and est. tokens; a recommendation is added when the list is non-empty. JSON/MLX output includes `large_file_candidates` for automation or tooling.

  This helps identify files that (a) would consume a large share of context if sent to an LLM, and (b) often indicate modules that have grown too large and would benefit from splitting for maintainability.

### 5.6 Summary table

| Area | Current | Improvement |
|------|---------|-------------|
| Orchestration | Custom tools + providers | Consider langchaingo or Jetify AI only if adding agentic workflows. |
| Cloud / self-hosted | Local (FM, Ollama, MLX) + optional LocalAI | Optional cloud SDKs later. |
| Discovery / docs | stdio://models, tool hints | Link this doc and LLM patterns from main docs; optional stdio://llm/status. |
| Numerics / ML | Gonum (graph + stat) | Keep; add GoLearn/Gonum only if new ML features are needed. |
| **Token estimation** | Chars × `tokens_per_char` (0.25) | **Nice-to-have:** [tiktoken-go](https://pkg.go.dev/github.com/tiktoken-go/tokenizer) for OpenAI-compatible API accuracy; optional dep or build tag. |

---

## 6. Follow-up tasks

Created from the improvement opportunities above (see `.todo2` or `exarp-go task list`):

| Task ID | Summary |
|---------|---------|
| T-1771252268378 | Evaluate langchaingo or Go AI SDK for agentic workflows |
| T-1771252272139 | Optional LocalAI backend (OpenAI-compatible) |
| T-1771252276374 | Document AI/LLM stack in main docs |
| T-1771252280227 | Optional stdio://llm/status resource |
| T-1771252286533 | **General agent abstraction** for run-agent-in-task (Cursor CLI, Cloud Agent, etc.) |

The **general agent abstraction** (T-1771252286533) defines a single interface so that “run an agent in another task” can use different runtimes (Cursor CLI `agent -p`, Cursor Cloud Agents API, or future MCP/sub-agents) behind one contract. See [CURSOR_API_AND_CLI_INTEGRATION.md](CURSOR_API_AND_CLI_INTEGRATION.md) for existing Cursor agent integration ideas.

---

## 7. References

- **Internal:** [LLM_NATIVE_ABSTRACTION_PATTERNS.md](LLM_NATIVE_ABSTRACTION_PATTERNS.md), [LLM_EXPOSURE_OPPORTUNITIES.md](LLM_EXPOSURE_OPPORTUNITIES.md), [LLM_ABSTRACTIONS_REDUCE_PYTHON.md](LLM_ABSTRACTIONS_REDUCE_PYTHON.md).
- **Code:** `internal/tools/llm_backends.go`, `internal/tools/graph_helpers.go`, `internal/tools/statistics.go`, `internal/tools/*provider*.go`.
- **Cursor rules:** `.cursor/rules/llm-tools.mdc` (when to use which LLM tool). See also §9 for Transformers-style equivalents and LLM middleware.

---

## 8. External resources and curated lists

Ecosystem overviews, official references, and third-party SDKs useful when evaluating or documenting the Go AI stack.

| Resource | Description |
|----------|-------------|
| **[Go Wiki: AI](https://go.dev/wiki/AI)** | Official Go wiki: calling hosted vs local (e.g. Ollama) services, prompt management with `text/template`, and pointers to Genkit Go, langchaingo, Ollama. |
| **[Jetify AI](https://github.com/jetify-com/ai)** | Unified Go AI SDK (OpenAI, Anthropic; streaming, tools, embeddings). See §1 and §5.1. |
| **[The State of AI in Go](https://captainnobody1.medium.com/the-state-of-ai-in-go-ceb3d029664a)** (Medium) | Article on the current Go AI ecosystem: libraries, patterns, and gaps. |
| **[awesome-golang-ai](https://github.com/promacanthus/awesome-golang-ai)** | Curated list: benchmarks, LLM tools, RAG, MCP, ML libs, neural networks, learning resources. |
| **[CodeGPT – Go AI](https://www.codegpt.co/agents/go)** | IDE-focused AI assistant for Go development (concurrency, web services, testing); product reference, not a library. |
| **[LaunchDarkly Go AI SDK](https://launchdarkly.com/docs/sdk/ai/go)** | LaunchDarkly’s Go SDK for **AI Configs**: context-based model/config selection, metrics and tracking, feature-flag-style control of AI behaviour (operational layer, not a provider SDK). |

**Relation to exarp-go:** The wiki and awesome list support “state of the ecosystem” and discovery. Jetify AI is the main candidate if we later adopt a unified Go SDK. LaunchDarkly is relevant only if we add feature flags or external metrics for LLM calls.

---

## 9. Transformers-style equivalents (Go / Swift) and LLM middleware

Reference for developers looking for Hugging Face Transformers–like options in Go or Swift, and for middleware that unifies multiple LLM backends behind a single API.

### 9.1 Go: no full Transformers port; inference via bindings

There is no Go library that replicates the full scope of Hugging Face Transformers (model definitions, training, Hub). Inference is done via bindings to C++ runtimes:

| Project | Role | Notes |
|---------|------|--------|
| **[llama.cpp](https://github.com/ggerganov/llama.cpp)** | C++ engine | GGUF models, GPU (Metal, CUDA, ROCm, etc.); used by many wrappers. |
| **[tcpipuk/llama-go](https://github.com/tcpipuk/llama-go)** | Go bindings | Idiomatic Go: chat, generation, embeddings, speculative decoding; actively maintained. |
| **[swdunlop/llm-go](https://github.com/swdunlop/llm-go)** | Go bindings | Simpler wrapper around llama.cpp; basic load/predict + optional HTTP server. |
| **Ollama** | App + Go backend | Uses its own [ggml backend in Go](https://pkg.go.dev/github.com/ollama/ollama/ml/backend/ggml); consumed via HTTP API. exarp-go uses Ollama via native client + optional Python bridge. |

**Relation to exarp-go:** We use **Ollama** (HTTP API) for llama-style models; no need for a separate Transformers-in-Go unless we want to embed **llama-go** for in-process GGUF inference.

### 9.2 Swift: MLX Swift LM as Transformers analogue

| Project | Role | Notes |
|---------|------|--------|
| **[ml-explore/mlx-swift](https://github.com/ml-explore/mlx-swift)** | Core | Array/ML framework for Apple Silicon. |
| **[ml-explore/mlx-swift-lm](https://github.com/ml-explore/mlx-swift-lm)** | LLM/VLM layer | Closest to Transformers in Swift: many architectures (Llama, Mistral, Gemma, Phi, Qwen), Hugging Face Hub, LoRA/fine-tuning, VLMs, embedders. [Swift Package Registry](https://swiftpackageregistry.com/ml-explore/mlx-swift-lm). |
| **[ml-explore/mlx-swift-examples](https://github.com/ml-explore/mlx-swift-examples)** | Examples | Chat, CLI, fine-tuning. |
| **Apple Foundation Models** | System API | `SystemLanguageModel`; on-device Apple models. exarp-go uses this via Swift/CGO. |

**Relation to exarp-go:** We use **Apple FM** (Swift/CGO) and an **MLX bridge**. For arbitrary HF/MLX-community models from Swift, **mlx-swift-lm** is the natural option; same MLX ecosystem the bridge targets.

### 9.3 LLM middleware (unified API, routing, observability)

Middleware that sits in front of multiple LLM backends and exposes a single API (often OpenAI-compatible) for routing, keys, and observability.

| Middleware | Stack | Role | Relation to exarp-go |
|------------|--------|------|----------------------|
| **[inference-gateway/inference-gateway](https://github.com/inference-gateway/inference-gateway)** | Go | Proxy: Ollama, OpenAI, Groq, Cohere, Anthropic, Cloudflare, DeepSeek. OpenAI-compatible API, MCP integration, OpenTelemetry, K8s. | **Best fit:** Run as a sidecar or service; exarp-go can add an optional backend that calls the gateway URL (like LocalAI). Single place for cloud routing and MCP tool exposure. |
| **[mozilla-ai/any-llm](https://github.com/mozilla-ai/any-llm)** | Python | Unified SDK + optional FastAPI gateway; budgets, API keys, analytics. | Useful for Python services; exarp-go unchanged; Python can call Inference Gateway instead. |
| **[tluyben/llm-router](https://pkg.go.dev/github.com/tluyben/llm-router)** | Go | Intercepts OpenAI/Anthropic calls, forwards to OpenRouter; system prompt injection, streaming, Docker. | Use when you want one API key and many models (OpenRouter) or a drop-in proxy. |
| **[llmrooter/router](https://github.com/llmrooter/router)** | Go | Self-hosted Open Router–like gateway with React admin UI. | Alternative to OpenRouter for self-hosted routing. |
| **LocalAI** | Go | Self-hosted OpenAI-compatible API for local models (GGUF, etc.). | **Already in exarp-go** as optional backend via `LOCALAI_BASE_URL`; middleware for anything OpenAI-compatible to use local models. |

**Recommendation:** For a single HTTP API across local + cloud with optional MCP and observability, add **Inference Gateway** as an optional backend (e.g. `provider=inference-gateway` or new backend pointing at the gateway URL). See §4.1 and [LLM_NATIVE_ABSTRACTION_PATTERNS.md](LLM_NATIVE_ABSTRACTION_PATTERNS.md).

### 9.4 Related ecosystem (Node, skills)

| Resource | Description |
|----------|-------------|
| **[Meridius-Labs/apple-on-device-ai](https://github.com/Meridius-Labs/apple-on-device-ai)** | Node/TypeScript bindings for Apple on-device models; Vercel AI SDK provider. Same capability as exarp-go Apple FM in the JS/TS ecosystem. |
| **[MCP Market: Apple Foundation Models skill](https://mcpmarket.com/tools/skills/apple-foundation-models)** | Claude Code skill (guidance) for SystemLanguageModel, guided generation, tool calling, guardrails. Complementary to exarp-go implementation. |
| **[Hugging Face Transformers](https://github.com/huggingface/transformers)** | Python model-definition and inference framework; 1M+ Hub checkpoints. exarp-go does not use it; Go/Swift alternatives above. |

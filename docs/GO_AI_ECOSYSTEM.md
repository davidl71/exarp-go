# Go AI Ecosystem

**Purpose:** Reference for the Go AI landscape (frameworks, provider SDKs, ML/numerics), what exarp-go uses today, and how exarp-go could improve by aligning with or adopting from this ecosystem.

**See also:** [LLM_NATIVE_ABSTRACTION_PATTERNS.md](LLM_NATIVE_ABSTRACTION_PATTERNS.md), [LLM_EXPOSURE_OPPORTUNITIES.md](LLM_EXPOSURE_OPPORTUNITIES.md).

---

## 1. General AI Application Frameworks

Frameworks that orchestrate AI models and services for complex applications and agentic workflows.

| Framework | Description |
|-----------|-------------|
| **langchaingo** | Go implementation of the Python LangChain framework; applications powered by LLMs (chains, agents, tools). |
| **Genkit Go** | Open-source framework by Google; unified API for AI apps across multiple model providers, multimodal content, tool calling. |
| **Ollama** | Run large language models locally; useful for offline or self-hosted AI. (exarp-go integrates with Ollama via HTTP API + optional Python bridge.) |
| **LocalAI** | Open-source, self-hosted alternative to the OpenAI API; run AI models on your own infrastructure (OpenAI-compatible API). |
| **Go AI SDK** | Toolkit for building AI-powered applications and agents in Go; feature parity to the Vercel AI SDK. |

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

### 5.4 Numerical and graph stack

- **Current:** Gonum for task graphs and statistics; no other ML libraries.
- **Opportunity:** Gonum is already part of the ecosystem; no change needed for current use cases.
- **Recommendation:** Keep using Gonum; if future features need classical ML (e.g. clustering, regression beyond current stats), consider **GoLearn** or more Gonum packages before adding Gorgonia/TensorFlow.

### 5.5 Summary table

| Area | Current | Improvement |
|------|---------|-------------|
| Orchestration | Custom tools + providers | Consider langchaingo or Go AI SDK only if adding agentic workflows. |
| Cloud / self-hosted | Local (FM, Ollama, MLX) + optional LocalAI | Optional cloud SDKs later. |
| Discovery / docs | stdio://models, tool hints | Link this doc and LLM patterns from main docs; optional stdio://llm/status. |
| Numerics / ML | Gonum (graph + stat) | Keep; add GoLearn/Gonum only if new ML features are needed. |

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
- **Cursor rules:** `.cursor/rules/llm-tools.mdc` (when to use which LLM tool).

# LLM Router & RouteLLM Research

**Date:** 2026-02-27  
**Purpose:** Document research on LLM routing solutions, workflow/agent platforms (n8n, MindStudio), and Golang LLM routers. Covers radlab llm-router (gateway), RouteLLM (ML cost routing), n8n/MindStudio (not routers), and Go-based gateways (inference-gateway, pLLM, kcolemangt/llm-router, etc.).

---

## Executive Summary

| Project | Type | Primary Goal | exarp-go Fit |
|---------|------|--------------|--------------|
| **radlab-dev-group/llm-router** | Infrastructure gateway | Deployable proxy: load balancing, security (masking, audit), multi-provider routing | Optional backend like LocalAI—point at router URL for unified local/cloud access |
| **lm-sys/RouteLLM** | ML-based query router | Per-query routing: send simple queries to cheaper models, complex to strong models; reduce cost 40–85% | Optional backend when exarp-go calls cloud APIs; adds intelligent cost-quality tradeoff |

Both expose **OpenAI-compatible** REST APIs, so exarp-go can integrate either as an optional `provider` (similar to LocalAI) without changing the existing abstraction.

---

## 1. radlab-dev-group/llm-router

**Repository:** [radlab-dev-group/llm-router](https://github.com/radlab-dev-group/llm-router)  
**License:** Apache 2.0  
**Stack:** Python 3.10+

### What It Is

An on-premises or cloud-deployable service that sits between applications and LLM providers. It provides a unified REST interface with load balancing, security features, and observability.

### Components

| Component | Role |
|-----------|------|
| **llm_router_api** | REST proxy; routes to OpenAI-compatible, Ollama, vLLM, LM Studio, etc. |
| **llm_router_lib** | Python SDK with typed models, retries, token handling |
| **llm_router_web** | Flask UIs for anonymizer and config management |
| **llm_router_plugins** | Rule-based anonymisation (e.g., fast_masker: PII, Polish identifiers) |
| **llm_router_services** | HTTP services for guardrails and masking |

### Key Features

- **Unified REST interface** — Same schema for multiple backends
- **Load balancing** — LoadBalancedStrategy, round-robin, weighted-random
- **Streaming** — Provider-agnostic chunked or aggregated responses
- **Security** — Masking, anonymization, guardrails, prohibited content
- **Auditing** — GPG-encrypted logs under `logs/auditor/`
- **Prometheus metrics** — `/metrics` when `LLM_ROUTER_USE_PROMETHEUS=1`
- **Embeddings** — Dedicated endpoints across providers
- **Dynamic config** — JSON `models-config.json` for providers, models, overrides

### Quick Start

```bash
# Docker
docker run -p 5555:8080 quay.io/radlab/llm-router:rc1

# Or Python
LLM_ROUTER_MINIMUM=1 python3 -m llm_router_api.rest_api
```

### exarp-go Integration Path

Add an optional backend (e.g. `provider=llm-router`) that calls the llm-router base URL via HTTP, similar to LocalAI. Config:

- `LLM_ROUTER_BASE_URL` (e.g. `http://localhost:5555/api`)
- Optional: `LLM_ROUTER_MODEL` for model override

**Benefits:** Single endpoint for local + cloud; built-in masking/audit; load balancing across Ollama/vLLM/OpenAI.

---

## 2. lm-sys/RouteLLM

**Repository:** [lm-sys/RouteLLM](https://github.com/lm-sys/RouteLLM)  
**License:** Apache 2.0  
**Stack:** Python (LiteLLM-backed)  
**Paper:** [RouteLLM: Learning to Route LLMs with Preference Data](https://arxiv.org/abs/2406.18665) (arXiv:2406.18665)  
**Blog:** [LMSYS Blog](http://lmsys.org/blog/2024-07-01-routellm/)

### What It Is

A framework for **per-query routing** between a strong (expensive) model and a weak (cheap) model. Routers use ML (matrix factorization, BERT, etc.) or heuristics to decide which model handles each request. Goal: reduce cost while keeping quality high.

### Claimed Results

- **Up to 85% cost reduction** while maintaining **~95% GPT-4 performance** on benchmarks (e.g. MT Bench)
- **>40% cheaper** than commercial routing offerings at similar performance

### Routers (Out of the Box)

| Router | Mechanism | Recommendation |
|--------|-----------|----------------|
| **mf** | Matrix factorization on preference data | **Recommended** — strong, lightweight |
| **sw_ranking** | Weighted Elo, votes weighted by prompt similarity | Good alternative |
| **bert** | BERT classifier on preference data | Benchmark/eval |
| **causal_llm** | LLM-based classifier tuned on preference data | Higher latency |
| **random** | Random choice | Baseline |

Routers are trained on `gpt-4-1106-preview` / `mixtral-8x7b-instruct-v0.1` but generalize to other strong/weak pairs.

### Model Selection Flow

1. User provides **cost threshold** (0–1).
2. Router computes **strong-model win rate** for the query.
3. If win rate > threshold → route to **strong** model; else → **weak** model.

### Model Format in Requests

Clients pass router + threshold in the `model` field:

```
model="router-mf-0.11593"
```

### Installation & Server

```bash
pip install "routellm[serve,eval]"

# Launch OpenAI-compatible server (port 6060 by default)
python -m routellm.openai_server \
  --routers mf \
  --strong-model gpt-4-1106-preview \
  --weak-model anyscale/mistralai/Mixtral-8x7B-Instruct-v0.1
```

### Model Support

Uses **LiteLLM** for providers (Anthropic, Gemini, Bedrock, Together, Anyscale, etc.) and **OpenAI-compatible** endpoints. Supports routing to local models (e.g. Ollama) via [routing_to_local_models](https://github.com/lm-sys/RouteLLM/blob/main/examples/routing_to_local_models.md).

### Threshold Calibration

Thresholds can be calibrated from the Chatbot Arena dataset:

```bash
python -m routellm.calibrate_threshold \
  --routers mf \
  --strong-model-pct 0.5 \
  --config config.example.yaml
# → threshold = 0.11593 for ~50% strong-model calls
```

### exarp-go Integration Path

When exarp-go calls cloud APIs (or an OpenAI-compatible gateway):

1. Point `text_generate` / LocalAI-style backend at RouteLLM server URL.
2. Use `model=router-mf-<threshold>` instead of a fixed model name.
3. Optionally add config for `ROUTELLM_BASE_URL`, `ROUTELLM_ROUTER`, `ROUTELLM_THRESHOLD`.

**Benefits:** Automatic cost-quality tradeoff; no retraining for new model pairs; evaluated on MMLU, GSM8K, MT-Bench.

---

## 3. Comparison

| Aspect | radlab llm-router | RouteLLM |
|--------|------------------|----------|
| **Routing logic** | Config-driven (load balance, round-robin) | ML/heuristic (per-query to strong vs weak) |
| **Primary goal** | Reliability, security, multi-provider | Cost reduction with quality preservation |
| **Security** | Masking, anonymization, audit | None built-in |
| **Local models** | Ollama, vLLM, LM Studio | Via LiteLLM / Ollama guide |
| **Cloud** | OpenAI-compatible backends | All LiteLLM providers |
| **Evaluation** | N/A | Benchmarks (MMLU, GSM8K, MT-Bench) |
| **API** | OpenAI-compatible REST | Drop-in OpenAI client or server |

### Complementary Use

- **llm-router** → operational layer (where requests go, how they’re secured, how load is spread).
- **RouteLLM** → cost layer (which model handles each query).

They can be composed: exarp-go → llm-router (security + load balance) → RouteLLM (cost routing) → cloud/local providers.

### 3.1 Composed deployment pattern

You can chain radlab llm-router and RouteLLM so that exarp-go gets both **security/load-balancing** and **cost-aware routing** through a single gateway URL:

1. **Run radlab llm-router** (e.g. Docker or Python) — handles masking, audit, load balancing across Ollama/vLLM/OpenAI.
2. **Run RouteLLM server** behind it (or as the upstream of llm-router, depending on where you want cost routing) — routes each request to strong vs weak model by threshold.
3. **Point exarp-go at the front of the chain** with `provider=gateway` and `OPENAI_GATEWAY_BASE_URL` set to the first service. Use optional `model` (e.g. `router-mf-0.11593`) when the front is RouteLLM.

**Example (exarp-go → RouteLLM → providers):** Set `OPENAI_GATEWAY_BASE_URL=http://localhost:6060` and pass `model=router-mf-0.11593` in `text_generate`; RouteLLM does cost routing, then calls your configured strong/weak models.

**Example (exarp-go → llm-router → RouteLLM → providers):** Set `OPENAI_GATEWAY_BASE_URL` to the radlab llm-router URL; configure llm-router’s `models-config.json` so that the chosen model is served by a RouteLLM instance. Then exarp-go sees one gateway; llm-router handles security and load; RouteLLM (if in the chain) handles cost.

---

## 4. References

- [radlab-dev-group/llm-router](https://github.com/radlab-dev-group/llm-router)
- [lm-sys/RouteLLM](https://github.com/lm-sys/RouteLLM)
- [RouteLLM Paper (arXiv:2406.18665)](https://arxiv.org/abs/2406.18665)
- [LMSYS Blog: RouteLLM](http://lmsys.org/blog/2024-07-01-routellm/)
- [Anyscale: Building an LLM Router](https://www.anyscale.com/blog/building-an-llm-router-for-high-quality-and-cost-effective-responses)
- [GO_AI_ECOSYSTEM.md](../GO_AI_ECOSYSTEM.md) §9.3 — LLM middleware table
- [LLM_NATIVE_ABSTRACTION_PATTERNS.md](../LLM_NATIVE_ABSTRACTION_PATTERNS.md)

---

## 5. n8n & MindStudio (Workflow/Agent Platforms — Not Routers)

These platforms orchestrate AI workflows and agent logic; they are **not** LLM gateways or routers in the radlab/RouteLLM sense. They call LLMs as part of larger workflows.

### 5.1 n8n

**Repository:** [n8n-io/n8n](https://github.com/n8n-io/n8n)  
**License:** Fair-code (Sustainable Use License)  
**Stack:** TypeScript, Node.js

| Aspect | Details |
|--------|---------|
| **Role** | Workflow automation platform with AI nodes (ChatGPT, Claude, LangChain, HuggingFace, Ollama) |
| **Integrations** | 400+ apps; 422+ with AI; custom code nodes (JavaScript, Python) |
| **AI features** | AI Agent nodes, LLM Chain nodes, Chat Trigger, tools, output parsers |
| **Deployment** | Self-hosted (free), n8n Cloud, Embed (white-label) |
| **Router-like?** | **No** — connects apps and calls LLMs per workflow step; no gateway/routing layer |

**Use case:** Build workflows that call multiple LLMs (e.g. extract invoice → summarize → send to CRM). Can integrate with [LLM Gateway](https://docs.llmgateway.io/guides/n8n) by setting a custom OpenAI base URL.

### 5.2 MindStudio

**URL:** [mindstudio.ai](https://mindstudio.ai)  
**Role:** No-code AI agent builder (commercial)

| Aspect | Details |
|--------|---------|
| **Role** | Visual agent builder; 200+ models (OpenAI, Anthropic, Google, Mistral, Meta) without separate API keys |
| **Features** | Dynamic tool use, multi-model workflows, multimodal (text, images, audio, video), Architect (auto-generate from text) |
| **Automation** | Blocks for text/image gen, data analysis, API calls, logic, Slack, etc. |
| **Deployment** | SaaS; trusted by Adobe, Google, Meta, IBM |
| **Router-like?** | **No** — agent orchestration and model selection within agents; not a proxy/gateway |

**Vs n8n:** MindStudio is AI-native (agents, reasoning, tool use); n8n is automation-first (triggers, integrations). MindStudio easier for non-technical users; n8n offers self-hosting and 400+ app connectors.

---

## 6. Golang LLM Routers

Go-based projects that provide router/gateway behavior comparable to radlab llm-router (infrastructure) or RouteLLM (client-side routing).

### 6.1 Summary Table

| Project | Stars (approx) | Role | exarp-go fit |
|---------|----------------|------|--------------|
| **[inference-gateway/inference-gateway](https://github.com/inference-gateway/inference-gateway)** | ~1k | Proxy: Ollama, OpenAI, Groq, Cohere, Anthropic, Cloudflare, DeepSeek; MCP, OpenTelemetry, K8s | **Already recommended** in GO_AI_ECOSYSTEM |
| **[pLLM](https://pllm.dev/)** | — | Enterprise gateway: multi-provider, adaptive routing, failover, RBAC, audit, cost management | Strong for production/enterprise |
| **[kcolemangt/llm-router](https://github.com/kcolemangt/llm-router)** | ~380 | Reverse proxy for Cursor: prefix-based routing (ollama/, groq/, openai/); model aliases, role rewrites | **Best for Cursor users** — use as Base URL |
| **[llmrooter/router](https://github.com/llmrooter/router)** | — | Self-hosted OpenRouter-like; Go backend, React admin UI | Self-hosted multi-provider routing |
| **[tluyben/llm-router](https://pkg.go.dev/github.com/tluyben/llm-router)** | — | Redirects OpenAI/Anthropic to OpenRouter; Docker, system prompt injection | Single OpenRouter API key |
| **[paularlott/llmrouter](https://github.com/paularlott/llmrouter)** | — | Simple aggregator across multiple OpenAI providers | Minimal multi-provider |

### 6.2 kcolemangt/llm-router (Cursor-Oriented)

Designed specifically for [Cursor](https://cursor.sh) as the Override OpenAI Base URL:

- **Prefix routing:** `ollama/phi4`, `openai/gpt-4o-mini`, `groq/deepseek-r1-distill-llama-70b-specdec`
- **Model aliases:** Map `o1` → `groq/deepseek-r1-distill-qwen-32b` so Cursor reasoning prompts work with Groq
- **Role rewrites:** Map `developer` → `system` for provider compatibility
- **Local + cloud:** Works with local Ollama and cloud providers; typically exposed via ngrok for Cursor

#### 6.2.1 Using kcolemangt/llm-router with Cursor and exarp-go

- **Cursor:** Run [kcolemangt/llm-router](https://github.com/kcolemangt/llm-router) locally (e.g. `./llm-router-darwin-arm64`), expose it via **ngrok** (`ngrok http 11411`), then in Cursor set **Override OpenAI Base URL** to the ngrok HTTPS URL and put the router’s API key in the OpenAI API Key field. Cursor then uses prefix-based models (e.g. `ollama/phi4`, `groq/...`).
- **exarp-go:** Use the **same** base URL for MCP tools: set `OPENAI_GATEWAY_BASE_URL` to your ngrok URL (or the local URL if exarp-go runs on the same machine) and optional `OPENAI_GATEWAY_API_KEY` to the same key. Then `text_generate` with `provider=gateway` and optional `model` (e.g. `ollama/phi4`) uses the same router. Cursor and exarp-go share one gateway for local + cloud models.

### 6.3 pLLM (Enterprise)

[pLLM](https://pllm.dev/) — enterprise-grade Go gateway:

- 100% OpenAI-compatible API
- Multi-provider (OpenAI, Anthropic, Azure, AWS Bedrock, Vertex AI, Llama, Cohere)
- Adaptive routing, failover, health-based load balancing
- JWT auth, RBAC, audit logging
- Budget management, intelligent caching
- Lower memory footprint than typical gateways

### 6.4 inference-gateway (Already Documented)

See GO_AI_ECOSYSTEM §9.3. Go-based, cloud-native, supports Ollama and major cloud providers; MCP integration, OpenTelemetry, Docker/K8s.

### 6.5 Integration opportunities and patterns

The Section 6 Go routers (and the Python radlab llm-router / RouteLLM) all expose **OpenAI-compatible** `/v1/chat/completions`. exarp-go already implements this pattern for LocalAI. The same patterns can support any of these gateways with minimal new code.

#### 6.5.1 Integration opportunities

| Opportunity | Description |
|-------------|-------------|
| **Unified “gateway” provider** | Add one configurable provider (e.g. `provider=gateway`) that calls a base URL from env. **One** implementation works for inference-gateway, pLLM, llmrooter, kcolemangt/llm-router, radlab llm-router, and RouteLLM. Users run the router they want and set `OPENAI_GATEWAY_BASE_URL` (or similar). No need for a separate provider per project. |
| **Discovery** | Extend `LLMBackendStatus()` in `llm_backends.go` with `gateway_available` (or `openai_gateway_available`) when the gateway base URL is set. `stdio://models` and tool hints then advertise the gateway so clients can use `provider=gateway`. |
| **ModelRouter / provider=auto** | Optionally add a `ModelType` (e.g. `ModelGateway`) and have `ResolveModelForTask` or the router select it when a gateway is configured and cost/quality suggests using it. Alternatively keep gateway as an explicit `provider=gateway` only and leave `auto` for local FM/Ollama/insight paths. |
| **Cursor coexistence** | **kcolemangt/llm-router** is used as Cursor’s “Override OpenAI Base URL”. exarp-go does not need to implement it: document that Cursor users can point Cursor at kcolemangt/llm-router (e.g. via ngrok) for Cursor’s models, while exarp-go tools use `provider=fm|ollama|localai|gateway` independently. Optional advanced: a “cursor_proxy” provider that forwards to the same base URL Cursor uses (from env) for parity. |

#### 6.5.2 Patterns we can use

| Pattern | Current use | Apply to Section 6 routers |
|---------|-------------|----------------------------|
| **OpenAI-compatible HTTP client** | `localai_provider.go`: POST to `baseURL/v1/chat/completions` with `model`, `messages`, `max_tokens`, `temperature`; parse `choices[0].message.content`. | **Same contract.** inference-gateway, pLLM, llmrooter, kcolemangt, radlab, RouteLLM all accept this. A single `openai_gateway_provider.go` (or generic “gateway” provider) parameterized by env is enough for all of them. |
| **Env-based configuration** | LocalAI: `LOCALAI_BASE_URL`, `LOCALAI_MODEL`; `Supported()` = base URL set. | Gateway: `OPENAI_GATEWAY_BASE_URL` (required), `OPENAI_GATEWAY_MODEL` (optional default), `OPENAI_GATEWAY_API_KEY` (optional). `Supported()` = base URL set. Same pattern, different env names. |
| **Optional `model` in tool** | `text_generate` today uses env model only for LocalAI. | Add optional **`model`** param to `text_generate` for `provider=localai` and `provider=gateway`. When present, send it in the request body. Enables: **RouteLLM** (`model=router-mf-0.11593`), **kcolemangt** (`model=ollama/phi4`, `model=groq/...`), **radlab** (model from config). No new providers needed for model-specific routing. |
| **TextGenerator interface** | `Supported() bool` + `Generate(ctx, prompt, maxTokens, temperature) (string, error)`. LocalAI, FM, Ollama, insight, gateway implement it via `text_generate`. | New gateway provider implements `TextGenerator`; `text_generate` switch includes `gateway`. Same pattern as LocalAI. |
| **Single gateway, many backends** | One LocalAI backend = one URL. | One **gateway** provider = one URL. Which router (inference-gateway, pLLM, kcolemangt, etc.) is chosen by the user at deploy time; exarp-go only needs to call the configured base URL. |

#### 6.5.3 Recommended implementation order

1. **Add `provider=gateway`** — New `gateway_provider.go` (or `openai_gateway_provider.go`): env `OPENAI_GATEWAY_BASE_URL`, optional `OPENAI_GATEWAY_API_KEY`, `OPENAI_GATEWAY_MODEL`; same HTTP logic as `localai_provider.go`. Register in `text_generate` and `LLMBackendStatus()`.
2. **Optional `model` param** — In `handleTextGenerate`, accept optional `params["model"]`; when provider is `localai` or `gateway`, pass it in the POST body instead of defaulting to env model. Enables RouteLLM and prefix-based routers.
3. **Document Cursor + kcolemangt** — In docs (e.g. GO_AI_ECOSYSTEM or this research doc), add a short “Using kcolemangt/llm-router with Cursor” note: set as Cursor Base URL via ngrok; exarp-go can use `provider=gateway` pointing at the same URL if desired, or use other providers independently.
4. **ModelRouter + gateway (optional)** — If we want `provider=auto` to use the gateway for some tasks, add `ModelGateway` and wire it in `ResolveModelForTask` / `defaultModelRouter.Generate`; otherwise keep gateway explicit-only.

---

## 7. References (Updated)

**Routers & Gateways:**
- [radlab-dev-group/llm-router](https://github.com/radlab-dev-group/llm-router)
- [lm-sys/RouteLLM](https://github.com/lm-sys/RouteLLM)
- [RouteLLM Paper (arXiv:2406.18665)](https://arxiv.org/abs/2406.18665)
- [GO_AI_ECOSYSTEM.md](../GO_AI_ECOSYSTEM.md) §9.3 — LLM middleware table

**Workflow/Agent Platforms (not routers):**
- [n8n](https://github.com/n8n-io/n8n) — [n8n AI Tutorial](https://docs.n8n.io/advanced-ai/intro-tutorial)
- [MindStudio](https://mindstudio.ai) — [MindStudio vs n8n](https://www.mindstudio.ai/blog/mindstudio-vs-n8n)

**Go Routers:**
- [inference-gateway/inference-gateway](https://github.com/inference-gateway/inference-gateway)
- [pLLM](https://pllm.dev/)
- [kcolemangt/llm-router](https://github.com/kcolemangt/llm-router)
- [llmrooter/router](https://github.com/llmrooter/router)
- [tluyben/llm-router](https://pkg.go.dev/github.com/tluyben/llm-router)
- [paularlott/llmrouter](https://github.com/paularlott/llmrouter)

---

## 8. Follow-up

| Task | Description |
|------|-------------|
| Add gateway provider | Single `provider=gateway` with `OPENAI_GATEWAY_BASE_URL` (covers inference-gateway, pLLM, kcolemangt, llmrooter, radlab, RouteLLM) — see §6.5 |
| Add optional `model` param | In `text_generate` for localai/gateway to support RouteLLM (`router-mf-0.11`) and prefix routing (`ollama/phi4`) |
| Add llm-router backend | Alternative: optional `provider=llm-router` with `LLM_ROUTER_BASE_URL` (or use gateway provider) |
| Add RouteLLM backend | Alternative: optional `provider=routellm` with base URL + router/threshold (or use gateway + model param) |
| Update GO_AI_ECOSYSTEM.md | Add both to §9.3 middleware table |
| Compose both | Document pattern: exarp-go → llm-router → RouteLLM → providers | **Done** — §3.1 |
| Cursor + kcolemangt/llm-router | Document using Go llm-router as Cursor Base URL for local + cloud models | **Done** — §6.2.1 |

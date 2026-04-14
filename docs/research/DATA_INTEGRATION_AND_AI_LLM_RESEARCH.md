# Data integration frameworks, AI/LLM, and exarp-go tasks

**Status:** Research note  
**Date:** 2026-04-05  
**Scope:** Relate [awesome-go](https://github.com/avelino/awesome-go) categories *Data Integration Frameworks* and *Artificial Intelligence* to exarp-go’s task stack, persistence, and MCP/LLM surfaces.

---

## 1. exarp-go task system (project-local)

### Canonical tool surface

Prefer **`task_workflow`**, **`task_analysis`**, **`task_discovery`** (with **`report`**, **`health`**, **`session`**, **`automation`**, etc.) as the primary operator API—see [`docs/INDEX.md`](../INDEX.md) and [`docs/TASK_TOOLS_GUIDE.md`](../TASK_TOOLS_GUIDE.md).

### Persistence and patterns

- **SQLite** (e.g. `modernc.org/sqlite`) backs Todo2 task state; config and protos are documented under Database & Storage in the index.
- **`TaskStore`** (DB-first, JSON fallback) is the recommended access path for tools. **`Direct DB`** bypasses fallback and is legacy in places—see [`docs/TASK_BACKEND_USAGE.md`](../TASK_BACKEND_USAGE.md) for the full tool-by-tool map.
- **Task-adjacent docs:** [`docs/TASK_WORKFLOW_LIST_SQL_AUDIT.md`](../TASK_WORKFLOW_LIST_SQL_AUDIT.md), [`docs/TASK_DISCOVERY.md`](../TASK_DISCOVERY.md), [`docs/MODEL_ASSISTED_WORKFLOW.md`](../MODEL_ASSISTED_WORKFLOW.md), [`docs/CLI_TASK_STATUS_SUPPORT.md`](../CLI_TASK_STATUS_SUPPORT.md).

### Takeaway

New “data integration” or “AI” features should respect **TaskStore**, **explicit health/db operations** (see root `README.md` for `health action=database`), and existing **MCP resources** (`stdio://tasks`, `stdio://ready-tasks`, etc.) rather than introducing parallel task pipelines.

---

## 2. Data integration frameworks (awesome-go + ecosystem)

[awesome-go § Data Integration Frameworks](https://github.com/avelino/awesome-go#data-integration-frameworks) lists three Go-oriented ELT/ETL entries:

| Project | Role | Relevance to exarp-go |
|--------|------|------------------------|
| **[Benthos](https://github.com/benthosdev/benthos)** (upstream; product lineage continues as **Redpanda Connect**) | Declarative stream bridge—many inputs/outputs, YAML pipelines, Bloblang mapping | **Sidecar / ops**, not a library to embed inside exarp-go. Useful if you need durable **NATS/Kafka/SQL/HTTP** movement of *telemetry or exported task snapshots* at scale. See [Redpanda Connect](https://github.com/redpanda-data/connect) and [Introduction to Redpanda Connect](https://www.benthos.dev/) (historical naming); migration notes: [Migrating to v4](https://docs.redpanda.com/redpanda-connect/guides/migration/v4). |
| **[CloudQuery](https://github.com/cloudquery/cloudquery)** | High-performance **ELT** with **plugin** sources/destinations, **Apache Arrow** type system, gRPC plugin protocol | **Analytics warehouse** or **multi-cloud asset sync** use case—orthogonal to per-repo Todo2, but relevant if exarp-go ever **exports** tasks/metadata into Postgres/BigQuery/Snowflake for fleet reporting. Architecture: [CloudQuery developers](https://www.cloudquery.io/docs/developers/creating-new-plugin), [Building CloudQuery](https://cloudquery.io/blog/building-cloudquery). |
| **[omniparser](https://github.com/jf-tech/omniparser)** | Streaming parser for CSV/JSON/XML/EDI/etc. → JSON via schemas | **Narrow, high-value** if exarp-go ingests **legacy flat files** (exports from Jira/CSV dumps/EDIF-ish logs) into structured task imports **without** standing up a full Connect pipeline. |

### Synthesis

- **Default:** keep task ingestion on **`task_discovery`**, **`task_workflow` import paths**, and SQLite—do not pull a full ELT framework into the binary unless requirements are clearly cross-system replication or warehouse load.
- **When to adopt Connect/Benthos:** bounded contexts needing **reliable streaming**, many connectors, and operational isolation (separate process).
- **When to adopt CloudQuery:** “sync many APIs/dbs → one warehouse” product bet, not single-repo task UX.
- **When to adopt omniparser:** one-off **file-format** adapters with schema-driven parsing.

---

## 3. AI / LLM (awesome-go + exarp-go alignment)

### awesome-go Artificial Intelligence section

Curated entries (abbreviated) are listed under [awesome-go § Artificial Intelligence](https://github.com/avelino/awesome-go#artificial-intelligence). Examples with direct overlap to exarp-go’s domain:

| Library | Notes | vs exarp-go today |
|--------|--------|-------------------|
| [ai](https://github.com/joakimcarlsson/ai) | Agents, embeddings, tool calling, **MCP** | exarp-go is already an **MCP server** with [`modelcontextprotocol/go-sdk`](https://github.com/modelcontextprotocol/go-sdk); adopt only if consolidating *client-side* agent loops in-process. Official SDK overview: [MCP SDKs](https://modelcontextprotocol.org/docs/sdk). |
| [chromem-go](https://github.com/philippgille/chromem-go) | Embeddable vector store | **Already a dependency** in exarp-go (`go.mod`) where vector memory applies. |
| [langchaingo](https://github.com/tmc/langchaingo) | Chains, prompts, tool abstractions | Heavier framework; useful if *many* chained LLM steps with shared memory—exarp-go currently centralizes generation on **`text_generate`** and backend-specific tools—see [`docs/AI_LLM_INTEGRATION.md`](../AI_LLM_INTEGRATION.md). |
| [langgraphgo](https://github.com/smallnest/langgraphgo) | Stateful multi-actor graphs | Same trade-off: adopt only for explicit graph workflows; risks overlap with existing task state machine + MCP. |
| [hotplex](https://github.com/hrygo/hotplex), [routex](https://github.com/Ad3bay0c/routex) | Agent runtime / YAML MCP tool servers | **Reference architectures** for multi-session CLI agents; merge carefully with exarp-go’s stdio MCP and `session` tooling. |
| [LocalAI](https://github.com/mudler/LocalAI), [Ollama](https://github.com/ollama/ollama) | Self-hosted OpenAI-compatible / local models | exarp-go already integrates **Ollama**, **LocalAI**, and gateway-style backends via `text_generate`; see [`docs/GO_AI_ECOSYSTEM.md`](../GO_AI_ECOSYSTEM.md). |
| [otellix](https://github.com/oluwajubelo1/otellix) | OTel + LLM cost guardrails | Consider if production **tracing/budget** around `text_generate` becomes a requirement. |
| [AegisFlow](https://github.com/saivedant169/AegisFlow) | AI gateway, policies, multi-provider | fits if exarp-go ever exposes a **shared gateway** in front of many org models—not required for single-user MCP. |

### Synthesis

- **Stay course:** **`text_generate`** + **`ollama`** + FM/insight paths + **official Go MCP SDK** cover most needs documented in [`docs/AI_LLM_INTEGRATION.md`](../AI_LLM_INTEGRATION.md).
- **Add langchaingo/langgraphgo** only with a concrete workflow (e.g. multi-step task decomposition graph) that is awkward in current handlers.
- **Vector memory:** prefer **chromem-go** patterns already in-repo over new vector DB deps unless a feature gap is identified.

---

## 4. Recommendations (prioritized)

1. **Tasks:** When touching persistence, align with **`TaskStore`** and [`TASK_BACKEND_USAGE.md`](../TASK_BACKEND_USAGE.md); avoid new “shadow” task stores.
2. **Data integration:** Treat **Redpanda Connect / CloudQuery** as **sibling services** for fleet/analytics exports, not core MCP binary deps.
3. **File ingestion:** Evaluate **omniparser** only for defined import formats; otherwise keep imports SQL/JSON/API-driven.
4. **AI:** Prefer **`text_generate` + MCP**; add **langchaingo** only for maintainability-tested chain graphs; consider **otellix**-class observability if cost/latency SLAs appear.
5. **Index:** Keep [`docs/INDEX.md`](../INDEX.md) and [`docs/research/LLM_ROUTER_AND_ROUTELLM_RESEARCH.md`](LLM_ROUTER_AND_ROUTELLM_RESEARCH.md) updated when adopting any of the above.

---

## 5. References

- [avelino/awesome-go](https://github.com/avelino/awesome-go) — curated list (Data Integration Frameworks; Artificial Intelligence).
- [redpanda-data/connect](https://github.com/redpanda-data/connect) — Redpanda Connect (Benthos lineage).
- [docs.redpanda.com — Redpanda Connect migration v4](https://docs.redpanda.com/redpanda-connect/guides/migration/v4).
- [cloudquery/cloudquery](https://github.com/cloudquery/cloudquery); [CloudQuery plugin developer docs](https://www.cloudquery.io/docs/developers/creating-new-plugin); [Building CloudQuery](https://cloudquery.io/blog/building-cloudquery).
- [jf-tech/omniparser](https://github.com/jf-tech/omniparser).
- [modelcontextprotocol/go-sdk](https://github.com/modelcontextprotocol/go-sdk); [MCP SDK documentation](https://modelcontextprotocol.org/docs/sdk).

---

*This note is descriptive research, not a commitment to implement listed dependencies.*

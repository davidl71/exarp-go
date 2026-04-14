# awesome-go: queues, text, trees, pipelines, caches, time, errors, codegen, ML, NLP, reflection, templates

**Status:** Research note  
**Date:** 2026-04-05  
**Source:** [avelino/awesome-go](https://github.com/avelino/awesome-go) sections the operator asked to cross-check against exarp-go.

Cross-links: [`DATA_INTEGRATION_AND_AI_LLM_RESEARCH.md`](DATA_INTEGRATION_AND_AI_LLM_RESEARCH.md), [`TASK_BACKEND_USAGE.md`](../TASK_BACKEND_USAGE.md).

---

## Legend

- **In-repo today:** already in `go.mod` (direct or important indirect).
- **Consider:** plausible next step if a concrete feature needs it.
- **Defer:** prefer `stdlib` / existing patterns unless requirements sharpen.

---

## Queues (incl. [dqueue](https://github.com/vodolaz095/dqueue))

| Entry (examples) | Role | exarp-go |
|------------------|------|----------|
| [dqueue](https://github.com/vodolaz095/dqueue) | Thread-safe deferred in-memory queue | **Defer.** No evidence exarp-go needs a separate deferred queue; tool orchestration uses channels, DB, and **asynq** for Redis-backed work. |
| [deque](https://github.com/gammazero/deque), [deque](https://github.com/edwingeng/deque) | Ring-buffer / high perf deques | **Defer** unless a hot path needs O(1) two-ended ops at scale. |
| [hatchet](https://github.com/hatchet-dev/hatchet) | Distributed task queue / durable workflows | **Heavy overlap** with **hibiken/asynq** already in `go.mod`. Adoption would be a product/architecture change, not a drop-in. |

**Synthesis:** Keep **asynq** + SQLite task state; add **dqueue**/deque only for a measured in-process backlog pattern.

---

## Text Analysis

| Entry (examples) | Role | exarp-go |
|------------------|------|----------|
| [bleve](https://github.com/blevesearch/bleve) | Full-text index | **Consider** only for local search over large task bodies or docs without an external engine. |
| [go-edlib](https://github.com/hbollon/go-edlib), [levenshtein](https://github.com/agext/levenshtein) | Edit distance, fuzzy match | **Consider** for duplicate-task detection, tag suggestion, or merge UX (see also task_analysis tooling). |
| [ptrie](https://github.com/viant/ptrie), [trie](https://github.com/derekparker/trie) | Prefix structures | **Defer** unless implementing prefix-heavy autocomplete over tags/commands. |

**Synthesis:** Task dedup already has dedicated flows; **go-edlib**-style metrics are the most likely *small* addition for fuzzy matching.

---

## Trees

| Entry (examples) | Role | exarp-go |
|------------------|------|----------|
| [graphlib](https://github.com/aio-arch/graphlib) | DAG topo sort / prune | **Consider** for **task dependency graphs** and wave planning (aligns with `task_analysis` / `task wave` concepts). |
| [treemap](https://github.com/igrmk/treemap) | Ordered map (red-black) | **Defer**; `stdlib` maps or SQLite ordering usually suffice. |
| [merkle](https://github.com/bobg/merkle), [hashsplit](https://github.com/bobg/hashsplit) | Content-defined chunking / Merkle | **Defer** unless artifact or blob integrity becomes a requirement. |

**Synthesis:** **graphlib**-class DAG helpers are the best thematic fit for dependency-ready task waves.

---

## Pipes

| Entry (examples) | Role | exarp-go |
|------------------|------|----------|
| [parapipe](https://github.com/nazar256/parapipe), [pipeline](https://github.com/hyfather/pipeline), [pipelines](https://github.com/nxdir-s/pipelines) | Fan-out/fan-in ordered pipelines | **Optional pattern** for multi-stage tool pipelines *inside* one handler; idiomatic Go often uses `errgroup` + channels instead. |
| [ordered-concurrently](https://github.com/tejzpr/ordered-concurrently) | Ordered concurrent map | Same trade-off: adopt if pipeline code gets unwieldy. |

**Synthesis:** Prefer **`golang.org/x/sync/errgroup`** and small functions first; add a pipe library only when stages multiply.

---

## Caches

| Entry (examples) | Role | exarp-go |
|------------------|------|----------|
| [ristretto](https://github.com/dgraph-io/ristretto), [otter](https://github.com/maypok86/otter), [theine](https://github.com/Yiling-J/theine-go) | High-perf in-memory TTL/LRU-style | **Consider** for hot read paths (reports, list views) if profiling shows repeated DB hits. |
| [groupcache](https://github.com/golang/groupcache) | Fill-on-miss distributed cache | **Rare** for single-user MCP; more for multi-tenant HTTP. |
| **In-repo:** [ctxcache](https://github.com/lawlielt/ctxcache) | Context-scoped cache | **Already used** (`go.mod`). |

**Synthesis:** Extend **ctxcache** / query tuning before adding **ristretto**/**otter**.

---

## Date and Time

| Entry (examples) | Role | exarp-go |
|------------------|------|----------|
| [iso8601](https://github.com/relvacode/iso8601), [dateparse](https://github.com/araddon/dateparse) | Fast / loose parsing | **Consider** for ingesting heterogeneous timestamps from exports or APIs. |
| [cronrange](https://github.com/1set/cronrange) | Cron-style time windows | **Consider** next to **robfig/cron** (already indirect via asynq) for automation rules. |
| [durafmt](https://github.com/hako/durafmt) | Human durations | Overlaps with **dustin/go-humanize** (already in `go.mod`). |

**Synthesis:** Prefer **`time`**, **humanize**, and existing cron wiring; add **dateparse**/ISO helpers when parsers become a bug source.

---

## Error Handling

| Entry (examples) | Role | exarp-go |
|------------------|------|----------|
| [eris](https://github.com/rotisserie/eris), [oops](https://github.com/samber/oops), [Fault](https://github.com/Southclaws/fault) | Wrapped errors + stack / metadata | **Consider** if MCP tool errors need uniform codes and stack traces in logs. |
| [multierr](https://github.com/uber-go/multierr), [go-multierror](https://github.com/hashicorp/go-multierror) | Aggregate errors | **Consider** when one tool runs many validations and should return *all* failures. |
| [cockroachdb/errors](https://github.com/cockroachdb/errors) | Network-portable errors | **Heavy**; only for cross-RPC error chains. |

**Synthesis:** Stay on **`errors.Is` / `errors.As` / `%w`** until observability requirements justify a wrapper lib.

---

## Generators (code generation)

| Entry (examples) | Role | exarp-go |
|------------------|------|----------|
| [jennifer](https://github.com/dave/jennifer) | Generate Go source programmatically | **Consider** for protobuf-adjacent or enum-heavy codegen. |
| [go-enum](https://github.com/abice/go-enum) | Enums from comments | **Consider** if string enums in tools proliferate. |
| [goverter](https://github.com/jmattheis/goverter) | Interface-driven converters | **Consider** when DTO ↔ domain mapping becomes noisy. |

**Synthesis:** Use **`go generate`** + small scripts today; adopt **jennifer**/**goverter** when generated surface area grows.

---

## Machine Learning

| Entry (examples) | Role | exarp-go |
|------------------|------|----------|
| [gonum](https://github.com/gonum/gonum) | Linear algebra, stats, optimization | **In-repo** (`go.mod`) for numeric/statistics-style code paths. |
| [Goptuna](https://github.com/c-bata/goptuna), [eaopt](https://github.com/MaxHalford/eaopt) | Black-box / evolutionary optimization | **Defer** unless task scheduling or estimation becomes an optimization problem. |
| [gorse](https://github.com/zhenghaoz/gorse), [goRecommend](https://github.com/timkaye11/goRecommend) | Recommendation / CF | **Defer** unless building “suggested next task” from interaction history beyond heuristics. |
| [hugot](https://github.com/knights-analytics/hugot) | HF transformers via ONNX | **Heavy**; overlaps with **LLM backends** (`text_generate`, Ollama). Use when you need *specific* small ONNX models in-process. |

**Synthesis:** Keep **LLM work** on **`text_generate` + chromem-go**; use **gonum** for classical metrics; avoid shipping full ML stacks without a scoped feature.

---

## Morphological analyzers (NLP)

| Entry (examples) | Role | exarp-go |
|------------------|------|----------|
| [porter2](https://github.com/zhenjl/porter2), [go-stem](https://github.com/agonopol/go-stem), [stemmer](https://github.com/dchest/stemmer) | English (etc.) stemming | **Consider** for English keyword extraction / search normalization on task text. |
| [kagome](https://github.com/ikawaha/kagome) | Japanese morphology | **Only if** JP locale/fuzzy search is required. |
| [spaGO](https://github.com/nlpodyssey/spago) | Self-contained NLP | **Large**; defer in favor of LLM or lighter stemmers. |
| [snowball](https://github.com/goodsign/snowball) | CGO snowball | **Avoid** in portability-sensitive builds unless necessary. |

**Synthesis:** For **search/index** in exarp-go, stemming is cheaper than full ML; locale drives choice (porter2 vs kagome).

---

## Tokenizers

| Entry (examples) | Role | exarp-go |
|------------------|--------|----------|
| [segment](https://github.com/blevesearch/segment) | Unicode UAX #29 grapheme/word boundaries | **High leverage** for correct truncation/display in TUI/MCP text (complements **clipperhouse/uax29** already in Charm/Bubble Tea deps). |
| [sentences](https://github.com/neurosnap/sentences) | Sentence splits | **Consider** for chunking long descriptions for LLM context or summaries. |
| [gojieba](https://github.com/yanyiwu/gojieba), [gse](https://github.com/go-ego/gse) | CJK segmentation | **Only if** product needs CJK tokenization (CGO/deps trade-offs). |

**Synthesis:** Prefer **stdlib + Unicode segment** over custom splits; align TUI width with **runewidth** / UAX #29 (already in stack).

---

## Reflection

| Entry (examples) | Role | exarp-go |
|------------------|------|----------|
| [reflectutils](https://github.com/muir/reflectutils) | Struct tags, walk, fill from string | **Consider** for generic tool arg mapping or config reflection. |
| [gopath](https://github.com/tenntenn/gpath) | Field access by expression | **Defer**; easy to overuse vs explicit types. |
| [go-deepcopy](https://github.com/tiendc/go-deepcopy), [copy](https://github.com/gotidy/copy) | Cross-type copy | **Consider** for snapshotting large structs; watch performance. |

**Synthesis:** Minimize reflection in hot MCP paths; use codegen or hand-written mappers at boundaries.

---

## Template engines

| Entry (examples) | Role | exarp-go |
|------------------|------|----------|
| [text/template](https://pkg.go.dev/text/template), [html/template](https://pkg.go.dev/html/template) | stdlib | **Default** for prompts, reports, and safe HTML if needed. |
| [fasttemplate](https://github.com/valyala/fasttemplate), [quicktemplate](https://github.com/valyala/quicktemplate) | Speed-focused replacements | **Consider** if prompt rendering is hot or allocations matter in benchmarks. |
| [sprout](https://github.com/go-sprout/sprout) | Template func library | **Consider** with stdlib templates to share helpers. |
| [templ](https://github.com/a-h/templ) | Type-safe HTML components | **Web UI** direction only; not required for MCP stdio. |

**Synthesis:** **Stay on stdlib** for execution packs and briefing strings; optimize only after proof.

---

## Quick priority table (exarp-go)

| Category | Default action |
|----------|----------------|
| Queues | Keep **asynq**; no **dqueue** unless profiling says so |
| Text / fuzzy | **go-edlib** if duplicate/suggest UX needs it |
| Trees / DAG | **graphlib**-style topo for waves/deps |
| Pipelines | **errgroup** first |
| Caches | **ctxcache** + DB; then **ristretto**/otter if hot |
| Time | **humanize** + **time**; **dateparse** if messy inputs |
| Errors | stdlib wraps; **multierr** if batch validations |
| Codegen | **jennifer** / **goverter** when generated code grows |
| ML | **gonum** + **text_generate**; avoid heavy ONNX unless scoped |
| Morphology | **porter2** for EN search; **kagome** for JP |
| Tokenizers | **segment** / UAX #29; **sentences** for chunking |
| Reflection | Sparingly; **reflectutils** if tag-driven tooling expands |
| Templates | **text/template**; **fasttemplate** if hot |

---

## References

- [awesome-go](https://github.com/avelino/awesome-go) — full categorized list.
- Individual package links appear inline above.

---

*Research only; no new dependencies implied.*

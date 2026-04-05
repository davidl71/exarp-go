## Coding standards (exarp-go)

This repo’s “how we build” guidance is split across:

- `AGENTS.md` — agent workflow rules (build/test/lint via `make`, task DB rules, scope discipline)
- `CLAUDE.md` — developer/agent operational guide (key patterns, where things live)

This document adds project-specific guidance on **factories vs helpers vs readonly globals**, and a default rule:
**outputs at boundaries should be immutable to callers**.

### Protobuf standards

See `docs/PROTOBUF_BEST_PRACTICES.md` (source: [protobuf.dev dos/don’ts](https://protobuf.dev/best-practices/dos-donts/)).

### Factories

Use a **factory** when you need **pluggability + discovery + consistent construction**.

- **Use when**:
  - There are multiple implementations behind an interface/trait (providers, drivers, backends)
  - Selection happens by name/config and you need validation (`"ollama"`, `"fm"`, `"sqlite"`, …)
  - You want a single place to register/construct variants and keep wiring predictable
- **Avoid when**:
  - There is only one implementation and branching is not expected
  - Construction has no meaningful configuration (prefer a plain constructor)

### Helpers

Use **helpers** for small, composable, deterministic logic shared in 2+ places.

- **Prefer**:
  - Pure functions (formatting, parsing, validation, normalization)
  - Helpers that are obviously “leaf” utilities and don’t hide domain control flow
- **Avoid**:
  - “helpers” that become dumping grounds for business logic
  - helpers that introduce hidden state

### Readonly package-level vars (globals)

Use package-level “readonly” vars only for **static data** reused across calls:
lookups, templates, allowlists/denylists, small registries.

- **Rules**:
  - Keep them **unexported**
  - If callers can obtain the map/slice, provide an accessor that returns a **defensive copy**
  - For nested maps/slices, do a **deep clone**
- **Why**: avoids repeated allocations while preserving immutability at the boundary.

### Mutable-by-design globals (explicit exceptions)

Some values are intentionally mutable because they represent runtime coordination or caches:

- Caches (TTL/file cache singletons)
- `singleflight.Group`
- `sync.Once`, `sync.Map`, rate limiters, semaphores
- `*sql.DB` pools/handles
- CLI invocation option structs (per-process invocation state)

When using these, add a short comment: **“Mutable by design: …”** describing why.

### Boundary immutability (default rule)

For **MCP tool results**, `stdio://` resources, and CLI JSON payloads:

- Do **not** return references to internal slices or maps.
- Clone/copy slices and maps at the boundary:
  - Slices: `append([]T(nil), src...)` (or `slices.Clone` when appropriate)
  - Maps: allocate a new map and copy keys/values (deep clone if nested)

**Rationale**: prevents accidental aliasing and makes consumers safer (especially when maps/slices are passed through multiple adapters).

### Known divergences

When we intentionally diverge from “immutable everywhere,” document it in `docs/IMMUTABILITY_POLICY.md`
and add local “mutable by design” comments at the definition site.

### Error handling

- Prefer `fmt.Errorf("context: %w", err)` for wrapping.
- Don’t silently drop errors. If a failure is ignorable, document why.
- Tool handlers should generally **return** errors (MCP clients decide how to display).

### Context usage

- `context.Context` should be the first arg in request paths and should be propagated down-call.
- Avoid `context.TODO()` in non-test code. If you truly have no context, add a short reason.

### Logging

- Prefer structured logging with stable key names (`"task_id"`, `"project_root"`, `"operation"`).
- Never log secrets (tokens, DSNs with credentials, auth headers).
- Prefer returning an error over logging-and-continuing in tool handlers.

### Output shape stability

- Prefer typed structs → marshal once for JSON outputs.
- If returning `map[string]interface{}`, keep keys stable and document required fields.
- Keep “compact vs pretty” behavior explicit via params/flags; don’t surprise callers.
- Task IDs are identifiers: **never truncate them** in UI or JSON/text outputs.

### Testing expectations

- Prefer table-driven tests (`t.Run`) and explicit fixtures for tricky cases.
- If you change tool/resource output shapes, add/adjust tests that assert the shape.
- Before committing behavior changes, run `make test-go`.

### Concurrency & shared state

- Any global cache/singleflight/semaphore/Once is **mutable by design**: add a short comment at the definition.
- Don’t mutate shared slices/maps after publishing them to other goroutines; clone at boundaries.

### File/package organization

- Keep tool handlers in `internal/tools/` and resource handlers in `internal/resources/`.
- Split files when they become multi-concern; prefer `<tool>_*.go` naming patterns already used in this repo.

### Performance

- Preallocate in hot paths (`make([]T, 0, n)`, `make(map[K]V, n)`).
- Avoid repeated marshal/unmarshal in tight loops; build typed structures and marshal once.
- Optimize with evidence (bench/pprof) rather than “because it feels hot.”

### Security

- Validate paths and external inputs; prefer allowlists where possible.
- Don’t write credentials to disk; don’t echo secrets in errors/logs.


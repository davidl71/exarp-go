## Immutability policy (exarp-go)

exarp-go prefers **immutable outputs at API boundaries** (MCP tool results, stdio:// resources, CLI JSON/text payloads). In practice that means:

- Do **not** return references to internal slices or maps that callers could mutate.
- When returning `[]string` or `map[...]...` derived from internal structs, return **defensive copies** (clone/copy-on-write).
- Internal computation can be mutable for simplicity/performance, but the final payload should be immutable-from-the-caller’s-point-of-view.

### Mutable by design (intentional exceptions)

Some state must remain mutable because it represents **process-local runtime state** or **synchronization primitives**, not domain data.

- **Caches / memoization**
  - Example: `internal/cache/cache.go` `GetScorecardCache()`
  - Rationale: process-local optimization; avoids recomputation and reduces I/O

- **Singleflight / deduplication state**
  - Example: `internal/tools/report.go` `reportOverviewFlight`
  - Rationale: internal coordination for concurrent calls

- **Global CLI invocation options**
  - Example: `internal/cli/cli.go` `CLIOutputOpts`
  - Rationale: per-invocation output mode shared across CLI subcommands

- **Database connection pools / handles**
  - Example: `internal/database/sqlite.go` `GetDB()` returns `*sql.DB`
  - Rationale: connection pools are intentionally mutable (stats, connections, prepared statements)

- **sync.Once / sync.Map / rate limiters / semaphores**
  - Rationale: concurrency primitives necessarily mutate internal state

### Guardrails

- **Prefer immutability at the boundary**: tools/resources should clone slices/maps.
- **Do not expose mutable internals**: avoid exporting variables that are meant to be treated as read-only.
- **If you must use a package-level “readonly” map**: keep it unexported and return a defensive copy for any accessor that hands it to callers.


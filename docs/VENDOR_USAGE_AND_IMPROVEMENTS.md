# Vendor usage analysis – duplication and improvements

This document reviews vendored packages (see [VENDOR_LICENSES.md](VENDOR_LICENSES.md)) for possible **duplication** of functionality, **underused** packages, and **improvements**. It is hand-maintained; regenerate the license table with `make vendor-licenses`.

---

## 1. Possible duplication

### Rate limiting vs `golang.org/x/time/rate`

- **Current:** `internal/security/ratelimit.go` implements a **per-client sliding-window** rate limiter (map of client ID → request timestamps, configurable window and max requests).
- **Vendored:** `golang.org/x/time/rate` provides a **token-bucket** `rate.Limiter` (single limiter per resource, no per-client map).
- **Conclusion:** Not a direct duplicate. Our use case is per-client API limiting; `rate` is better for single-resource or global limits. **No change required** unless we add a simple global limit, in which case using `rate.Limiter` would be appropriate.

### Type conversion vs `github.com/spf13/cast`

- **Current:** We do manual type assertions and switches on `map[string]interface{}` and `interface{}` in many places (e.g. `internal/tools/estimation_shared.go`, params parsing, metadata handling).
- **Vendored:** `github.com/spf13/cast` is present (used by **hibiken/asynq** in vendor) and provides `cast.ToString`, `cast.ToInt`, `cast.ToFloat64`, `cast.ToBool`, `cast.ToStringSlice`, etc. from `interface{}`.
- **Improvement:** Consider using `cast` in our own code where we convert `interface{}` or JSON-like maps to string/int/float64/bool. This would reduce boilerplate and edge-case bugs. We do **not** currently import `cast` in `internal/` or `cmd/`.

---

## 2. Underused or optional packages

### `github.com/dustin/go-humanize`

- **Vendored:** Only used inside **modernc.org/libc** (transitive). We do not import it in exarp-go code.
- **Optional use:** If we ever format byte sizes, durations, or relative times for user-facing output (e.g. “2.3 MB”, “3 days ago”), we could use `humanize.Bytes`, `humanize.Time`, etc. instead of custom formatting. Low priority unless we add such UI.

### `github.com/robfig/cron/v3`

- **Vendored:** Pulled in by **hibiken/asynq** for scheduled tasks. We do not use cron directly.
- **Optional use:** If we add a “dispatcher” or scheduled job (e.g. “enqueue waves on a schedule” as in plan M5), we could use `cron` for cron expressions. Asynq already has scheduling; use cron only if we need cron syntax outside Asynq.

### `golang.org/x/time/rate`

- **Vendored:** Transitive (e.g. via Asynq or other deps). We do not import it.
- **Optional use:** For simple global or per-resource rate limits (token bucket), `rate.Limiter` is a good fit. Our current security need is per-client sliding window, which we already implement.

---

## 3. Packages we use as intended (no duplication)

| Package | Where used | Purpose |
|--------|------------|--------|
| **gonum.org/v1/gonum** | `internal/tools/graph_helpers.go` | Task DAG, topological sort, critical path (no duplication). |
| **golang.org/x/term** | `internal/cli/tui.go` | Terminal raw mode / size (correct use). |
| **gopkg.in/yaml.v3** | `internal/config/` | Config loading (correct use). |
| **google.golang.org/protobuf** | `internal/models/`, `internal/config/`, proto code | Protobuf (correct use). |
| **modernc.org/sqlite** | `internal/database/sqlite.go` | SQLite driver (correct use). |
| **github.com/hibiken/asynq** | `internal/queue/` | Task queue over Redis (correct use). |
| **github.com/charmbracelet/bubbletea**, **lipgloss** | `internal/cli/tui*.go` | TUI (correct use). |
| **github.com/racingmars/go3270** | `internal/cli/tui3270.go` | 3270 terminal (correct use). |
| **github.com/davidl71/mcp-go-core** | Many tools | MCP framework, request/response, CLI (correct use). |
| **github.com/davidl71/devwisdom-go** | `internal/tools/recommend.go` | Wisdom/recommendations (correct use). |
| **github.com/joshgarnett/agent-client-protocol-go** | `internal/acp/` | ACP server (correct use). |
| **github.com/google/go-github/v60** | `internal/tasksync/github_issues.go` | GitHub API (correct use). |

---

## 4. Summary and recommendations

| Item | Action |
|------|--------|
| **Rate limiter** | Keep current per-client sliding-window impl; consider `golang.org/x/time/rate` only for future global/simple limits. |
| **Type conversion** | Consider adopting **spf13/cast** in `internal/` for `interface{}`/map → string, int, float64, bool to reduce duplication and bugs. |
| **Humanize** | Optional: use **dustin/go-humanize** if we add user-facing size/duration/relative-time formatting. |
| **Cron** | Optional: use **robfig/cron/v3** only if we add cron-based scheduling outside Asynq (e.g. wave dispatcher). |
| **Other vendor** | No other duplication or redundant packages identified; usage aligns with each package’s purpose. |

---

*Last updated from codebase review; adjust as dependencies and features change.*

# Dependency Analysis

**Tag hints:** `#docs` `#analysis`

**Generated:** 2026-02-17

---

## 1. Go module dependencies

### 1.1 Direct dependencies (`go.mod`)

| Module | Version | Purpose |
|--------|---------|---------|
| `github.com/blacktop/go-foundationmodels` | v0.1.8 | Apple Foundation Models (CGO, darwin/arm64) |
| `github.com/charmbracelet/bubbletea` | v1.3.10 | TUI framework |
| `github.com/charmbracelet/lipgloss` | v1.1.0 | TUI styling |
| `github.com/davidl71/devwisdom-go` | v0.1.2 | DevWisdom MCP (report briefings) |
| `github.com/davidl71/mcp-go-core` | v0.3.1 | MCP core |
| `github.com/go-sql-driver/mysql` | v1.9.3 | MySQL driver (Todo2) |
| `github.com/google/go-github/v60` | v60.0.0 | GitHub API (security, health) |
| `github.com/lib/pq` | v1.11.2 | PostgreSQL driver (Todo2) |
| `github.com/modelcontextprotocol/go-sdk` | v1.3.0 | MCP SDK |
| `github.com/racingmars/go3270` | v0.9.13 | 3270 terminal (optional) |
| `golang.org/x/oauth2` | v0.35.0 | OAuth2 (GitHub, Google) |
| `golang.org/x/sys` | v0.41.0 | Syscall / platform |
| `golang.org/x/term` | v0.40.0 | Terminal handling |
| `gonum.org/v1/gonum` | v0.17.0 | Graph algorithms (task deps, critical path) |
| `google.golang.org/api` | v0.266.0 | Google APIs |
| `google.golang.org/protobuf` | v1.36.11 | Protobuf (memory tool) |
| `gopkg.in/yaml.v3` | v3.0.1 | YAML config |
| `modernc.org/sqlite` | v1.45.0 | Pure-Go SQLite (Todo2 default, no CGO) |

### 1.2 Replace directives

| Replaced module | Local path | Note |
|-----------------|------------|------|
| `github.com/davidl71/devwisdom-go` | `/Users/davidl/Projects/devwisdom-go` | Local dev; `go list -m all` fails if path missing |
| `github.com/davidl71/mcp-go-core` | `/Users/davidl/Projects/mcp-go-core` | Local dev; same caveat |

**Impact:** Builds and `go list -m all` require these paths to exist on the machine. CI or other devs should use published versions or set `replace` to a valid path.

### 1.3 CGO vs pure Go

- **CGO required:** `github.com/blacktop/go-foundationmodels` (Swift bridge on darwin/arm64 only; behind build tags).
- **Pure Go (no CGO):** All other direct deps. `modernc.org/sqlite` is pure Go (no cgo-sqlite).

### 1.4 Notable indirect dependency chains

- **Google API / auth:** `google.golang.org/api` → `cloud.google.com/go/auth`, `googleapis/gax-go/v2`, `go.opentelemetry.io/...`
- **Charm TUI:** `bubbletea` / `lipgloss` → `charmbracelet/x/ansi`, `muesli/termenv`, `rivo/uniseg`, etc.
- **SQLite stack:** `modernc.org/sqlite` → `modernc.org/libc`, `modernc.org/memory`, `modernc.org/mathutil`, `ncruces/go-strftime`

---

## 2. Task dependency system (Todo2)

### 2.1 Data model

- **Task fields:** `Dependencies` (blocking task IDs), `ParentID` (hierarchy).
- **Storage:** `task_dependencies` table (task_id, depends_on_id); loaded with tasks in `internal/database` (e.g. `loadTaskDependencies`).

### 2.2 Graph and analysis

| Component | Location | Role |
|-----------|----------|------|
| `TaskGraph` | `internal/tools/graph_helpers.go` | Gonum `DirectedGraph` + task ID ↔ node ID maps |
| `BuildTaskGraph` | `graph_helpers.go` | Builds graph from tasks; edges from `Dependencies` and `ParentID` |
| `AddDependency(dep, task)` | `graph_helpers.go` | Edge dep → task (dep blocks task) |
| `HasCycles` / `DetectCycles` | `graph_helpers.go` | Cycle detection (e.g. via topo sort) |
| `GetDependencyAnalysisFromTasks` | `task_analysis_shared.go` | Returns cycles + missing refs from task list |
| `findMissingDependencies` | `task_analysis_shared.go` | Tasks that reference non-existent dependency IDs |
| `FindCriticalPath` | `graph_helpers.go` | Longest path in DAG (backlog-only graph) |
| `GetTaskLevels` | `graph_helpers.go` | Dependency depth (topo or iterative for cyclic) |
| `TopoSortTasks` | `graph_helpers.go` | Execution order (deps before dependents) |

### 2.3 Who consumes dependency analysis

- **task_analysis (dependencies):** `handleTaskAnalysisDependencies` — full analysis, critical path, levels, cycles, missing; outputs JSON/text.
- **task_discovery (orphans):** `findOrphanTasks` / `findOrphanTasksBasic` — use `GetDependencyAnalysisFromTasks` for cycles and missing; then parent/orphan checks.
- **todo2_db_adapter:** `sortTasksByDependencies` — topo sort for save order (deps before dependents).
- **task_workflow / execution_plan:** Uses graph for ordering and validation.

### 2.4 Dependency flow summary

```
Todo2 tasks (DB/JSON)
    → BuildTaskGraph (Dependencies + ParentID)
    → TaskGraph (gonum)
    → HasCycles / DetectCycles, findMissingDependencies
    → GetDependencyAnalysisFromTasks(cycles, missing)
    → FindCriticalPath (backlog), GetTaskLevels, TopoSortTasks
    → MCP response (dependencies action), orphan discovery, save order
```

---

## 3. Database layer dependencies

- **Drivers:** SQLite (default, `modernc.org/sqlite`), MySQL (`go-sql-driver/mysql`), Postgres (`lib/pq`). Rqlite planned (see `docs/RQLITE_BACKEND_PLAN.md`).
- **Abstraction:** `internal/database` — `Driver` interface, config (env/file), init per driver; tasks, tags, dependencies, locking in DB.
- **Invariant:** No direct edits to `.todo2/state.todo2.json` for backend storage; all access via `database` package.

---

## 4. Suggested follow-ups

1. **Go modules:** Run `go list -m -mod=readonly all` (or temporarily remove/override `replace` for devwisdom-go and mcp-go-core) to list full tree and run `govulncheck`/updates.
2. **Task dependencies:** Use MCP `task_analysis` with `action=dependencies` (and optional `action=dependencies_summary`) to inspect cycles, missing refs, and critical path for current backlog.
3. **Rqlite:** Add `gorqlite` and `DriverRqlite` per `docs/RQLITE_BACKEND_PLAN.md` without adding CGO.

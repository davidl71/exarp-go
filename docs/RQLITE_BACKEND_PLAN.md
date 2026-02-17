# Plan: Self-Hosted rqlite Backend for exarp-go

**Tag hints:** `#database` `#feature` `#config` `#docs`

**Status:** Draft  
**Created:** 2026-02-17

---

## 1. Purpose & Success Criteria

**Purpose:** Enable exarp-go to use a self-hosted [rqlite](https://github.com/rqlite/rqlite) node as the Todo2 database backend so that multiple development machines (macOS, Linux, FreeBSD) share one source of truth without cloud dependencies (only GitHub in the cloud).

**Success criteria:**

- Users can set `DB_DRIVER=rqlite` and `DB_DSN=http://host:4001` (or equivalent) and run exarp-go against a remote rqlite node.
- No change to existing SQLite/MySQL/Postgres usage when rqlite is not configured.
- One planning doc and Todo2 tasks track the work; implementation uses existing driver abstraction.

---

## 2. Technical Foundation

- **Client:** [gorqlite](https://github.com/rqlite/gorqlite) stdlib driver — `database/sql` compatible, pure Go, no CGO. DSN format: `http://host:4001` or `https://user:pass@host:4001/?level=strong`.
- **Server:** User runs rqlite themselves (single node or cluster); exarp-go does not ship or start rqlite.
- **Integration:** Add `DriverRqlite` and `driver_rqlite.go` implementing the existing `Driver` interface; reuse SQLite-like dialect (rqlite speaks SQLite SQL). Register driver and extend config/env to accept `rqlite` and rqlite DSN default.
- **Invariants:** No direct edits to `.todo2/state.todo2.json` for rqlite path; all access via `database` package. Existing tests for SQLite continue to use local SQLite unless explicitly switched.

---

## 3. Implementation Phases

**Tag hints:** `#database` `#feature`

### Phase 1: Add rqlite driver

- [ ] Add dependency: `github.com/rqlite/gorqlite` (or `gorqlite/stdlib` for `database/sql` driver).
- [ ] Add `DriverRqlite` constant in `internal/database/driver.go`.
- [ ] Implement `internal/database/driver_rqlite.go`: `RqliteDriver` (Type, Open, Configure, Dialect, Close) and `RqliteDialect` (SQLite-compatible: `?` placeholders, same types). Use `sql.Open("rqlite", dsn)` after blank-importing gorqlite stdlib.
- [ ] Register rqlite in `init()` or on first use (consistent with MySQL/Postgres lazy registration if desired).

### Phase 2: Config and env wiring

**Tag hints:** `#config` `#database`

- [ ] In `internal/database/config.go`: accept `DriverRqlite` in `LoadConfig` switch; add default DSN for rqlite (e.g. `http://localhost:4001`). Update `GetDefaultDSN` for rqlite.
- [ ] In `LoadConfigFromCentralizedFields`: when driver is rqlite, set DSN from env or default (no filepath join).
- [ ] In `internal/database/sqlite.go` `InitWithConfig`: add case for `DriverRqlite` to register and get rqlite driver (same pattern as MySQL/Postgres).

### Phase 3: Documentation

**Tag hints:** `#docs`

- [ ] Add `docs/RQLITE_SETUP.md` (or section in existing docs): how to install and run rqlite (single node), set `DB_DRIVER=rqlite` and `DB_DSN`, run exarp-go on multiple machines. Note macOS/Linux/FreeBSD and self-host only.

### Phase 4: Tests and sanity

**Tag hints:** `#testing`

- [ ] Unit test for rqlite driver: either use a test that mocks/skips when no rqlite server (e.g. integration tag) or document that rqlite tests require a local node. Ensure existing SQLite tests still pass.

---

## 4. Open Questions

- Whether to register rqlite in `init()` (always) vs on first use (lazy like MySQL/Postgres) to avoid pulling gorqlite when only SQLite is used.
- Whether to add a `make` target or script to start a local rqlite node for dev (optional).

---

## 5. Todo2 Tasks

Created via exarp-go (see `.todo2/`):

| Phase | Task ID | Name |
|-------|---------|------|
| 1 | T-1771353163496878000 | Add rqlite driver (driver_rqlite.go, gorqlite) |
| 2 | T-1771353168386164000 | Wire config and env for rqlite (DB_DRIVER, DB_DSN) |
| 3 | T-1771353169844112000 | Document self-hosted rqlite setup (docs) |
| 4 | T-1771353170943219000 | Tests for rqlite driver and config |

Optional: set task dependencies (e.g. Phase 2 → Phase 1, Phase 4 → Phase 1 & 2) via `exarp-go task update` or task_workflow.

---

## 6. References

- [rqlite](https://rqlite.io/) — distributed SQLite, Raft.
- [gorqlite](https://github.com/rqlite/gorqlite) — Go client; stdlib subpackage for `database/sql`.
- [rqlite client libs](https://rqlite.io/docs/api/client-libraries/) — DSN and options.
- exarp-go: `internal/database/driver.go`, `driver_sqlite.go`, `config.go`, `sqlite.go`.

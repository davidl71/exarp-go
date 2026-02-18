# rqlite Backend Setup (exarp-go)

Use a self-hosted [rqlite](https://rqlite.io/) node as the Todo2 database so multiple machines share one source of truth without cloud dependencies.

**Platforms:** macOS, Linux, FreeBSD (self-hosted only).

---

## 1. Install and run rqlite

- **macOS (Homebrew):** `brew install rqlite`
- **Linux/FreeBSD:** Download from [rqlite releases](https://github.com/rqlite/rqlite/releases) or build from source.

Start a single node (default HTTP on 4001):

```bash
rqlited -node-id 1 ~/rqlite/data
```

Or use the [official docs](https://rqlite.io/docs/install-rqlite/) for cluster setup.

---

## 2. Configure exarp-go

Set environment variables (or export in your shell):

```bash
export DB_DRIVER=rqlite
export DB_DSN=http://localhost:4001
```

For a remote node or auth:

```bash
export DB_DSN=https://user:password@host:4001/?level=strong
```

Then run exarp-go as usual (CLI, MCP server, etc.). All Todo2 access goes to the rqlite node.

---

## 3. Multiple machines

- Run one rqlite node (or cluster) on a host reachable from all dev machines.
- On each machine set `DB_DRIVER=rqlite` and `DB_DSN=http://that-host:4001`.
- No change to existing SQLite/JSON usage when rqlite is not configured.

---

## 4. References

- [RQLITE_BACKEND_PLAN.md](RQLITE_BACKEND_PLAN.md) — implementation plan and phases
- [rqlite](https://rqlite.io/) — distributed SQLite
- [gorqlite](https://github.com/rqlite/gorqlite) — Go client (stdlib driver used by exarp-go)

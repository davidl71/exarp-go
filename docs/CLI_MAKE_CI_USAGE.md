# exarp-go CLI Usage from Make and CI

**Purpose:** Run exarp-go tools directly from Makefiles, shell scripts, or CI (e.g. GitHub Actions) via the CLI. No MCP client required.

---

## Invocation pattern

```bash
exarp-go -tool <tool_name> -args '<json>'
```

- Use the built binary: `./bin/exarp-go` or `bin/exarp-go` from repo root.
- Or run from source: `go run ./cmd/server -tool <name> -args '<json>'`.
- `-list`: list tools (no `-args`).
- `-test <tool>`: run tool with example args (smoke test).

---

## Make targets

Ensure `make build` has been run, then:

| Target | Description |
|--------|-------------|
| `make exarp-list` | List all exarp-go tools |
| `make exarp-report-scorecard` | Report scorecard |
| `make exarp-report-overview` | Report overview |
| `make exarp-health-server` | Health check (server) |
| `make exarp-health-docs` | Health check (docs) |
| `make exarp-context-budget` | Context budget (default items). Override with `CONTEXT_JSON='{"action":"budget","items":["a","b"],"budget_tokens":4000}'` |
| `make exarp-test TOOL=health` | Smoke-test a tool via `-test` |

Example override:

```bash
make exarp-context-budget CONTEXT_JSON='{"action":"budget","items":["file1","file2"],"budget_tokens":8000}'
```

---

## CI examples

### GitHub Actions

Build the binary, then run tools:

```yaml
- run: go build -o bin/exarp-go ./cmd/server
- run: ./bin/exarp-go -tool health -args '{"action":"server"}'
- run: ./bin/exarp-go -tool report -args '{"action":"scorecard","include_metrics":true}'
```

See `.github/workflows/agentic-ci.yml` and `.github/workflows/go.yml` for full examples.

### Shell scripts

```bash
./bin/exarp-go -tool context -args '{"action":"budget","items":["a","b"],"budget_tokens":4000}'
./bin/exarp-go -tool text_generate -args '{"prompt":"Summarize this repo","max_tokens":100}'
```

---

## Optional: `tool_name key=value` form

Scripts can use `exarp-go tool_name key=value ...`. The CLI rewrites this to `-tool tool_name -args '{"key":"value",...}'` when the first argument is a tool name (not a reserved subcommand). Prefer `-tool` / `-args` for clarity and to avoid quoting issues.

---

## References

- **CURSOR_MCP_SETUP.md** — MCP server setup for Cursor.
- **Makefile** — `exarp-*` targets and existing ones (`fmt`, `lint`, `sprint`, `pre-sprint`, etc.) that call exarp-go.
- **Context token reduction** — Use `context` (budget, summarize, batch) to reduce Cursor token usage; see previous discussion.

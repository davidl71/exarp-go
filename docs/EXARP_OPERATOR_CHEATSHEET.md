# exarp-go operator cheat sheet

Quick reference for **humans and agents**: task lifecycle, batch approval, and footguns. For Makefile aliases and env vars, see [EXARP_CLI_SHORTCUTS.md](EXARP_CLI_SHORTCUTS.md).

## Project root

- **Canonical:** directory containing `.exarp` and/or `.todo2` for the project you intend to operate on.
- **`PROJECT_ROOT`:** set when running from another tree (MCP, CI, portable wrapper). Client repos often use `scripts/run_exarp_go.sh` with `PROJECT_ROOT` so tasks stay in **that** repo, not the exarp-go clone.

```bash
cd /path/to/target-repo
export PROJECT_ROOT="$(pwd)"   # if your entrypoint does not set it
/path/to/exarp-go/bin/exarp-go task list --status Todo
```

## Task lifecycle (CLI first)

| Step | Command |
|------|---------|
| List Todo / Review | `exarp-go task list --status Todo` / `--status Review` |
| Show one | `exarp-go task show T-123` |
| Create | `exarp-go task create "Title" --description "..." --priority medium` |
| Comment | `exarp-go -tool task_workflow -args '{"action":"add_comment","task_id":"T-123","comment_type":"result","content":"..."}'` |
| Update status | `exarp-go task update T-123 --new-status Done` |
| Batch update | `exarp-go task update --status Todo --new-status "In Progress" --ids "T-1,T-2"` |

**Prefer** these convenience subcommands **before** raw `-tool task_workflow` JSON for list/update/create/show.

## Batch approve (`task_workflow` action `approve`)

Used to move **many** tasks from one status to another (e.g. all **Review** → **Done**).

```bash
exarp-go -tool task_workflow -args '{
  "action": "approve",
  "status": "Review",
  "new_status": "Done",
  "output_format": "json",
  "compact": true
}'
```

### Critical details

1. **`new_status` is not optional in practice** — If omitted, the server default may be **`Todo`**, not `Done`. Always pass `"new_status":"Done"` (or your target) explicitly for closing reviews.
2. **`task_ids`** — Comma-separated string or JSON array. If **omitted**, every task matching ** `status` ** (and optional **`filter_tag`**) is a candidate — high blast radius. Preview first.
3. **`dry_run`: true** — Lists who would be updated without writing. Use before large batches.

```bash
exarp-go -tool task_workflow -args '{
  "action":"approve",
  "status":"Review",
  "new_status":"Done",
  "dry_run": true
}'
```

4. **`clarification_none`: true** — Skips tasks whose description is “too short” for approval filters (see config). Omit unless you know you need it.

## MCP vs CLI

- **MCP:** same tool params; include **`output_format":"json"`** (and **`compact": true`**) when the consumer is code or log pipelines. See [exarp-mcp-output.mdc](../.cursor/rules/exarp-mcp-output.mdc) (repo rule) or [MCP tool docs](MCP_CLIENT_SOLUTION.md).
- **Discovery:** `tool_catalog` with `action=help`, `tool_name=task_workflow`.

## Common failures

| Symptom | Likely cause |
|--------|----------------|
| `proto: syntax error` / unexpected token on **create** | **`--tag`** values: avoid leading **`#`** in flags that flow through protobuf; use `foo,bar` or separate flags per tool support. |
| Tasks updated “wrong” after approve | **`new_status` omitted** → check default; always set explicitly. |
| JSON mirror out of date | SQLite is primary; run **`task sync`** for `.todo2/state.todo2.json` when your editor relies on it. |
| Wrong project’s tasks | **`PROJECT_ROOT`** pointed at the wrong repo; fix entrypoint or cwd. |

## Result comments and “Done”

Some integrations require a **result** comment before **Done**. If `task update --new-status Done` fails, add a **result** comment via `task_workflow` **`add_comment`**, then update again.

## Verification one-liners

```bash
# From exarp-go repo (or with PATH to binary)
exarp-go task list --status Review | head
make sanity-check
```

## See also

- [EXARP_CLI_SHORTCUTS.md](EXARP_CLI_SHORTCUTS.md) — Make targets, env, shell aliases
- [CLI_TASK_STATUS_SUPPORT.md](CLI_TASK_STATUS_SUPPORT.md) — Status strings and CLI behavior
- [CURSOR_API_AND_CLI_INTEGRATION.md](CURSOR_API_AND_CLI_INTEGRATION.md) — Session prime and status contract
- [AGENTS.md](../AGENTS.md) — Repo-wide agent guidelines
- `exarp-go task --help` / `exarp-go task create --help` after `make build`

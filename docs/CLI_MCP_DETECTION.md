# CLI vs MCP Mode Detection

**Location:** `cmd/server/main.go` (entry) + `internal/cli/mode.go` (logic)  
**Purpose:** Decide whether to run as CLI (interactive/script) or MCP stdio server.

---

## Decision order

1. **Normalize** – `cli.NormalizeToolArgs(os.Args)` rewrites `exarp-go tool_name key=value ...` to `-tool <name> -args <json>` when first arg is a single word (not a reserved subcommand, no `/`, no leading `-`). Used so hooks/scripts with non-TTY stdin run CLI instead of MCP.
2. **CLI flags** – `cli.HasCLIFlags(os.Args)` is true if first arg is a reserved subcommand or any arg is a CLI flag → **CLI**.
3. **Mode** – `cli.DetectMode() == cli.ModeCLI` (stdin is a TTY) → **CLI**. Uses mcp-go-core `DetectMode` / `IsTTY`.
4. **Else** → **MCP** (stdio JSON-RPC server).

---

## API (internal/cli)

### Mode (mode.go)

| Symbol | Purpose |
|--------|--------|
| **ReservedSubcommands** | `[]string{"task", "config", "tui", "tui3270"}` – first-args that are subcommands, not tool names. |
| **CLIFlags** | `[]string` – flags that imply CLI (`-tool`, `-args`, `-list`, `-completion`, `-test`, `-i`, `-h`, `help`, etc.). |
| **NormalizeToolArgs(args)** | Rewrites `args` to `-tool`/`-args` form when first arg is a tool name; returns `(newArgs, true)` if changed. |
| **HasCLIFlags(args)** | Returns true if args indicate CLI mode (reserved subcommand or any CLI flag). |

### mcp-go-core CLI (cli.go)

| Symbol | Purpose |
|--------|--------|
| **IsTTY()** | stdin is a terminal. Delegates to mcp-go-core `IsTTY`. |
| **IsTTYStdout()** | stdout is a terminal. Uses mcp-go-core `IsTTYFile(os.Stdout)`. Use for **colored output** detection. |
| **ColorEnabled()** | `IsTTYStdout()`. Enable ANSI colors only when true (e.g. `-list` tool names). |
| **DetectMode()** | `ModeCLI` if TTY, else `ModeMCP`. Uses mcp-go-core `DetectMode`. |
| **ModeCLI / ModeMCP** | Execution modes (re-exported from mcp-go-core). |

### Subcommand dispatch (mcp-go-core ParseArgs)

`Run()` uses `mcpcli.ParseArgs(os.Args[1:])` to branch on `parsed.Command` for **config**, **task**, **tui**, **tui3270**. Flag-based modes (`-tool`, `-list`, `-test`, `-i`, `-completion`) use the `flag` package (ParseArgs does not skip consumed `-flag value` args).

**Tests:** `internal/cli/mode_test.go` – `TestNormalizeToolArgs`, `TestHasCLIFlags`, `TestReservedSubcommands`, `TestCLIFlags`.

---

## References

- **Pre-push / non-JSON:** `docs/NEXT_MIGRATION_STEPS.md` (§ Pre-push hook and non-JSON stdin)
- **TTY / mode:** `internal/cli/cli.go` → `IsTTY`, `IsTTYStdout`, `DetectMode` → mcp-go-core `pkg/mcp/cli`
- **Subcommands:** `internal/cli/cli.go` (config, task, tui, tui3270 via ParseArgs)

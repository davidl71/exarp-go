# iTerm2 Integration: TUI and CLI

How exarp-go’s TUI and CLI can interact with iTerm2, and where to hook it in.

## Current TUI/CLI flow

- **CLI**: `internal/cli/cli.go` — `Run()` dispatches subcommands (`config`, `task`, **`tui`**, `tui3270`, etc.). TUI is invoked with `exarp-go tui [status]` → `RunTUI(server, status)`.
- **TUI**: `internal/cli/tui.go` — Bubble Tea program; uses `os.Stdout` for rendering and `term.GetSize(int(os.Stdout.Fd()))` for dimensions; SIGWINCH is handled by Bubble Tea.
- **TTY/color**: `IsTTYStdout()` / `ColorEnabled()` in `cli.go` use `os.Stdout` to decide if output is to a terminal (and thus safe for color/TUI).

So both TUI and CLI already run in the terminal and write to stdout; when that terminal is iTerm2, we can add iTerm2-specific behavior in two ways.

## Two ways to interact with iTerm2

### 1. Escape sequences (no extra dependency)

When exarp-go runs **inside** iTerm2, it can emit [iTerm2 proprietary escape codes](https://iterm2.com/documentation-escape-codes.html) on stdout. No extra Go dependency; only use when `TERM_PROGRAM=iTerm.app` (or similar) so other terminals are unaffected.

**Useful sequences:**

| Goal | Sequence (concept) | Notes |
|------|--------------------|--------|
| Set tab/session title | `OSC 0 ; <title> BEL` or `OSC 1 ; <title> BEL` | e.g. "exarp-go TUI" or task filter |
| Badge text | `OSC 1337 ; SetBadgeFormat=<base64> ST` | Badge in corner of window |
| Notifications | `OSC 1337 ; Custom=... ST` | Requires specific payload format |
| Hyperlinks | OSC 8 | For task IDs linking to Cursor/plan |

**Integration points:**

- **TUI startup** (`RunTUI` in `tui.go`): Before `tea.NewProgram(…)`, if `os.Getenv("TERM_PROGRAM") == "iTerm.app"`, print OSC to set session title (e.g. `exarp-go TUI` or `exarp-go TUI: In Progress`). Optionally set badge to project name.
- **CLI** (e.g. after `task list` or `task update`): When stdout is TTY and iTerm2, optionally print OSC to set title to last command or project (e.g. `exarp-go task list`).

Implement a small helper (e.g. `internal/cli/iterm2.go`) that only runs when `TERM_PROGRAM` indicates iTerm2 and stdout is a TTY, and only writes OSC sequences; keep all existing behavior unchanged when not iTerm2.

### 2. goiterm / marwan.io/iterm2 (WebSocket automation)

[marwan.io/iterm2](https://pkg.go.dev/marwan.io/iterm2) (and the `goiterm` binary from `marwan.io/iterm2/cmd/goiterm@latest`) use the **iTerm2 WebSocket API** to control iTerm2 from outside: create windows, tabs, sessions, run commands in them. This is separate from “we are running inside an iTerm2 session.”

**Ways to use it with exarp-go:**

- **Optional dependency**: Add `marwan.io/iterm2` as an optional import; only call it when the user requests “open TUI in new iTerm2 tab” (e.g. a flag or subcommand). Then:
  - Create an `iterm2.App` (e.g. `iterm2.NewApp("exarp-go")`), open a new tab or window, and run `exarp-go tui` (or the full path to the binary) in that session.
- **Shell out to goiterm**: If `goiterm` is on PATH, you can `exec.Command("goiterm", …)` to run iTerm2 automation (e.g. “new tab”, “run exarp-go tui”) without adding the library to exarp-go. Exact CLI of `goiterm` would need to be checked in [cmd/goiterm](https://github.com/marwan-at-work/iterm2/tree/master/cmd/goiterm).

**Integration points:**

- **CLI**: New subcommand or flag, e.g. `exarp-go tui --iterm2-new-tab` that either:
  - Uses `marwan.io/iterm2` to create a tab and run `exarp-go tui [status]`, or
  - Executes `goiterm` with the right arguments so the user’s iTerm2 opens a new tab running `exarp-go tui`.
- **TUI**: The TUI itself stays unchanged; “interaction with iTerm2” here means “how we launch the TUI inside iTerm2” (escape sequences) or “how we ask iTerm2 to open the TUI in a new tab” (WebSocket/goiterm).

## Recommended order

1. **Escape sequences**: Add optional iTerm2 detection and OSC for title (and optionally badge) in `RunTUI` and, if desired, for CLI commands. No new deps; works whenever exarp-go is run inside iTerm2.
2. **goiterm / WebSocket**: If you want “open TUI in new tab” from outside iTerm2 or from a script, add a small path that shells out to `goiterm` or uses `marwan.io/iterm2` behind a flag or subcommand.

## References

- [iTerm2 escape codes](https://iterm2.com/documentation-escape-codes.html)
- [Session title (OSC 0/1)](https://iterm2.com/documentation-session-title.html)
- [marwan.io/iterm2](https://pkg.go.dev/marwan.io/iterm2) and [GitHub marwan-at-work/iterm2](https://github.com/marwan-at-work/iterm2)
- exarp-go: `internal/cli/cli.go` (dispatch, `RunTUI`), `internal/cli/tui.go` (`RunTUI`, Bubble Tea, `term.GetSize`)

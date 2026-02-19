# Using universal-ctags in CI, Make, TUI, and Other Features

**Scope:** [universal-ctags](https://github.com/universal-ctags/ctags) (not the Sourcegraph [go-ctags](https://github.com/sourcegraph/go-ctags) wrapper). This doc covers non-Cursor uses: CI, Make, TUI, and other tooling.

**Dependency:** Install `universal-ctags` (e.g. `brew install universal-ctags` on macOS, or in CI image). The project does not vendor a ctags library; Make/scripts shell out to the `ctags` binary.

**Ansible:** The linters role installs universal-ctags when `install_linters` is true and `universal-ctags` is in the `linters` list (default). Run the development playbook with `install_linters=true` and ensure the linters role runs; see `ansible/roles/linters/`.

**exarp-go:** When ctags or a tags file is present, the **health** tool can report it: use `health` with `action=ctags` to get `ctags_available`, `tags_file_present`, and `tags_file_path`. Useful for automation or agents to know whether a symbol index exists.

---

## 1. Make

**Use:** Generate a `tags` (or `TAGS`) file for the repo so editors and scripts can resolve symbols without parsing Go.

**Example target:**

```makefile
# Optional: requires universal-ctags on PATH
CTAGS ?= ctags
TAGS_FILE ?= tags

tags: ## Generate tags file (universal-ctags) for Go + scripts. Requires: universal-ctags on PATH.
	@command -v $(CTAGS) >/dev/null 2>&1 || { echo "universal-ctags not found; install with e.g. brew install universal-ctags"; exit 1; }
	$(CTAGS) -R -o $(TAGS_FILE) --languages=Go,Sh --exclude=.git --exclude=vendor --exclude=bin .
	@echo "Generated $(TAGS_FILE)"
```

**Typical usage:**

- `make tags` — regenerate after big changes.
- Editors (Vim, Emacs, etc.) that read `tags` get "go to definition" for Go and shell.
- Scripts can `grep` or parse the tags file for "list all functions in package X" or "find symbol Y."

**Note:** Exclude `vendor/` and `.git/` to keep the file small and avoid indexing dependencies.

---

## 2. CI

**Uses:**

| Use | How |
|-----|-----|
| **Generate and archive tags** | Run `ctags -R ...`, upload `tags` as an artifact. Downstream: diff symbols between commits, or serve for "browse code" / search. |
| **Sanity check** | Run ctags; if it fails (e.g. parse error on weird syntax), the job fails. Acts as a light structural check. |
| **Symbol diff** | Generate tags on main and on PR branch; diff tag lines to see "which symbols were added/removed" for review or changelog. |

**Example (GitHub Actions):**

```yaml
- name: Install universal-ctags
  run: sudo apt-get install -y universal-ctags   # or use a image that has it

- name: Generate tags
  run: ctags -R -o tags --languages=Go,Sh --exclude=.git --exclude=vendor --exclude=bin .

- name: Upload tags
  uses: actions/upload-artifact@v4
  with:
    name: tags
    path: tags
```

**Optional:** Only run when Go or script files change; no need on doc-only PRs.

---

## 3. TUI (exarp-go)

**Current:** The TUI (`internal/cli/tui.go`) focuses on tasks, waves, and Cursor CLI integration. It does not currently have "jump to symbol" or "search symbols."

**Possible addition:** If you want symbol navigation inside the TUI:

- **Option A:** Run `make tags` (or equivalent) so a `tags` file exists. The TUI (or a subprocess) can parse the tags file and offer a "search symbol" / "open at definition" flow (e.g. open user’s `$EDITOR` at `file:line`).
- **Option B:** Call universal-ctags from Go via `exec.Command` when the user invokes "symbol search," and parse stdout. No need for the [go-ctags](https://github.com/sourcegraph/go-ctags) wrapper unless you want a long-lived ctags process; a one-shot `ctags -R ...` or `ctags -L -` (from file list) is enough.

**Effort:** Medium (UI for query + list results + open editor or show location). Low if you only "run ctags and dump to a file" for external editor use.

---

## 4. Other Non-Cursor Uses

| Use | Description |
|-----|-------------|
| **Vim / Emacs** | Point editor to repo `tags` (or `TAGS`) for "go to definition" and symbol search. `make tags` keeps it fresh. |
| **Scripts** | Grep or parse `tags` to list exported functions, find call sites, or generate a simple "symbol index" page. |
| **Doc generation** | Use tag names/kinds to build a minimal "list of functions/types per package" without full Go AST (crude but no Go parser needed). |
| **Dead code (crude)** | Compare "symbols in tags" vs "symbols referenced in code" (e.g. via grep); requires more logic and is best done with proper static analysis later. |

---

## 5. Recommendation Summary

| Area | Recommendation |
|------|----------------|
| **Make** | Add a `tags` target that runs universal-ctags over the repo (Go + shell, exclude vendor/git/bin). Optional, on-demand. |
| **CI** | Optional: install universal-ctags, generate tags, optionally upload as artifact or use for symbol-diff. Low priority unless you have a consumer. |
| **TUI** | Optional: add "symbol search" that uses a pre-generated tags file or one-shot ctags; open editor at file:line. Nice-to-have. |
| **Other** | Use the same `tags` file for Vim/Emacs and ad-hoc scripts; no extra infrastructure. |

**Bottom line:** universal-ctags is useful for Make (generate tags for editors/scripts), and optionally in CI (artifact or symbol diff) and TUI (symbol search). Add the Make target first; CI and TUI only if you have a concrete consumer.

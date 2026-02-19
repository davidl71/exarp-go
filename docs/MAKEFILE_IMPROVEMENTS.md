# Makefile improvement suggestions

**Status:** Living doc. Many items implemented; optional follow-ups listed below.

---

## Implemented (this pass)

- **task-show** — `make task-show TASK_ID=T-123` runs `exarp-go task show` (show one task). Aligns with CLI and task-workflow skill.
- **task-run-with-ai** — `make task-run-with-ai TASK_ID=T-123 [BACKEND=ollama]` runs `exarp-go task run-with-ai` for local LLM guidance. Default backend is `fm`.
- **check-security** — Standalone target that runs only the exarp-go security scan (`security action=scan`). Use for quick checks without full pre-release (govulncheck + vendor-licenses).
- **Split .PHONY** — Grouped `.PHONY` lines by section (main, test, config, task, queue, ansible, etc.).
- **help “Quick reference”** — `make help` includes a fixed Quick reference block (`make b`, `make test`, `make fmt`, `make task-list`, `make task-show TASK_ID=...`, `make task-run-with-ai`, `make check-security`).
- **exarp_ensure_binary** — Shared macro used by `fmt`, `lint`, `lint-fix`, `lint-all`, `lint-all-fix`; builds binary if missing instead of failing.
- **queue-status** — `make queue-status` prints REDIS_ADDR and usage hints (enqueue, worker, dispatcher). No CLI “queue status” command yet; this is hint-only.
- **clean-binary** — `make clean-binary` removes `$(BINARY_PATH)`. Use with `clean` for a full rebuild.
- **clean** — Comment added: `clean` does not remove the main binary by design; use `clean-binary` to remove it.
- **ansible-docker-smoke** — `make ansible-docker-smoke` runs Ansible `--syntax-check` inside `cytopia/ansible:latest-tools` (Ubuntu) for cross-platform smoke. Supports task T-1771457731564409000.

---

## Suggested (optional)

### Structure and readability

- **DRY for build** — `build`, `build-debug`, and `build-race` repeat the same Darwin/arm64 + CGO detection. A macro or small helper script could reduce duplication and make adding new build variants easier.

### Ansible

- **ansible-check deps** — Ensure `ansible-galaxy` is run before syntax-check when using ansible-lint (e.g. `ansible-check` could depend on `ansible-galaxy` for environments where community.general is required). See task T-1771459746925000000.

### Documentation

- **make-shortcuts.mdc** — Keep `.cursor/rules/make-shortcuts.mdc` in sync when adding shortcuts (e.g. `task-show`, `task-run-with-ai`, `check-security`, `queue-status`, `clean-binary`, `ansible-docker-smoke`).
- **help target** — Ensure every new target has a `## Description` comment so `make help` stays complete.

---

## References

- Makefile: repo root `Makefile`
- Shortcuts: `.cursor/rules/make-shortcuts.mdc`
- Task CLI: `exarp-go task show`, `exarp-go task run-with-ai` (see CLAUDE.md, MODEL_ASSISTED_WORKFLOW.md)
- Ansible: `docs/ANSIBLE_SETUP.md`, `ansible/README.md`

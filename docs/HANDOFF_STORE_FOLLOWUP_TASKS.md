# Handoff store and related follow-up tasks

**Status (2026-03-30):** Todo2 items T-1774898687671700000, T-1774898687688479000, T-1774898687700802000 are **Done** (verification run; parity test added; `GatherEvidence` uses `WalkDir`). Remaining `filepath.Walk` call sites and a batch `GetTasksByIDs` API are optional follow-ups if profiling warrants it.

Created from the handoff_store zero-copy / `handoffEntryToMap` work and the broader DB/perf backlog. **Todo2 IDs** below were created in this repo; if your DB diverges, recreate with `exarp-go task create` or `task_workflow` action=create.

## Tasks (created)

| Todo2 ID | Title | Priority | Tags |
|----------|--------|----------|------|
| T-1774898687671700000 | Verify handoff_store changes: make test-go + make lint | high | testing, handoff, quality |
| T-1774898687688479000 | Optional: parity test handoffEntryToMap vs JSON round-trip | low | testing, handoff |
| T-1774898687700802000 | Perf backlog: GetTasksByIDs batch + WalkDir audit | medium | performance, database |

## Details

### T-1774898687671700000 — Verification gate

- Run from repo root: `make test-go`, then `make lint`.
- Scope: `internal/tools/handoff_store.go`, `handoff_store_test.go`.
- If `make fmt` fails with permission errors under `.cache/go-mod`, fix `GOMODCACHE` / permissions locally; do not treat as code defects.

### T-1774898687688479000 — Map vs JSON contract

- `handoffEntryToMap` may emit `int` for some numeric fields where `json.Unmarshal` into `map[string]any` yields `float64`.
- Existing coverage: `TestHandoffEntryToMapJSONNumbers` (accepts either for snapshot count).
- Optional parity test: compare `handoffEntryToMap` to `json.Marshal` → `json.Unmarshal` into `map[string]any` with normalization.

### T-1774898687700802000 — Performance backlog

- Batch path: `GetTasksByIDs` (or equivalent) to avoid N× `GetTask` where lists are built in a loop.
- I/O: prefer `filepath.WalkDir` over `filepath.Walk` where directory trees are still walked.
- Tooling: reduce redundant full task loads in `infer_task_progress` / plan sync paths where profiling shows cost.

## Related code

- `internal/tools/handoff_store.go` — `saveHandoffStore`, `handoffEntryToMap`
- `internal/tools/handoff_store_test.go` — round-trip and numeric-type tests

After DB task changes, run `./scripts/run_exarp_go.sh task sync` from a project that uses this `.todo2` (or your wrapper) so `state.todo2.json` stays aligned.

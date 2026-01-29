# Python Deprecated – MLX-Only Retention

**Date:** 2026-01-29  
**Purpose:** After deprecating Python, retain only what is needed to **test and develop MLX** via the bridge. Remove all other manual/CI Python.

---

## Keep (required for MLX test/develop)

| Location | Purpose |
|----------|---------|
| **bridge/execute_tool.py** | Bridge entry; routes `mlx` only. |
| **bridge/proto/bridge_pb2.py** | Protobuf for Go ↔ Python bridge. |
| **project_management_automation/__init__.py** | Package root. |
| **project_management_automation/tools/__init__.py** | Package root. |
| **project_management_automation/tools/consolidated_mlx_only.py** | Slim re-export; bridge imports this. |
| **project_management_automation/tools/consolidated_ai.py** | Defines `mlx()`; imports mlx_integration. |
| **project_management_automation/tools/mlx_integration.py** | MLX status/hardware/models/generate. |

Tests that hit the bridge (e.g. `tests/unit/python/test_execute_tool.py`, `tests/integration/bridge/test_bridge_integration.py`) only need the bridge + the above MLX path.

---

## Removed (2026-01-29; manual/CI, not needed for MLX)

- **project_management_automation/scripts/** – Removed (Python daily removed; duplicate_test_names etc. deprecated).
- **project_management_automation/resources/** – Removed (not on MLX path).
- **project_management_automation/utils/** – Removed (mlx_integration uses inline fallback for error_handler).
- **project_management_automation/tools/** – All files except `__init__.py`, `consolidated_mlx_only.py`, `consolidated_ai.py`, `mlx_integration.py` removed.

`project_management_automation` now contains only the MLX chain above; bridge folder is MLX-only.

---

## What manual/CI can be removed (MLX still testable)

**CI:** Neither `go.yml` nor `agentic-ci.yml` run Python. There is no Python in CI to remove.

**Manual / Makefile (removed or optional so MLX stays testable):**

| Change | Effect |
|--------|--------|
| **`make test` = Go only** | Default `make test` no longer runs Python tests. Run `make test-python` or `make test-integration` when developing MLX/bridge. |
| **`make test-coverage` = Go coverage only** | Python coverage is optional: run `make test-coverage-python` when working on MLX. |
| **`make bench` = Go benchmarks only** | Python `tests/benchmarks/` does not exist; `bench` runs `go-bench` only. |

**Keep for MLX test/develop:**

- `make test-python` – run when changing bridge or MLX code.
- `make test-integration` – includes bridge integration tests.
- `make install` / `make install-dev` – `uv sync` so bridge and pytest work when you run the above.

---

## References

- **End state:** docs/PYTHON_MIGRATION_END_STATE.md  
- **What can remove:** docs/PYTHON_WHAT_MORE_CAN_REMOVE.md  

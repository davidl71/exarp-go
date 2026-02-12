# Symlink Test Case Design for Path Validation

**Date:** 2026-02-12  
**Status:** Design (informs T-284 and related implementation)  
**Purpose:** Capture test case design for symlink handling in `ValidatePath` / `ValidatePathExists` (exarp-go and mcp-go-core).

---

## Current Behavior

**mcp-go-core `ValidatePath`** does **not** call `filepath.EvalSymlinks`. It:

1. Cleans and resolves the input path relative to `projectRoot`
2. Ensures the resulting path string is within the project root via `filepath.Rel` and `..` checks
3. Returns the absolute path if the **path string** does not escape root

Therefore:

- A symlink **inside** the project that **points to a path inside** the project: the path string stays within root → **accepted** by `ValidatePath`.
- A symlink **inside** the project that **points outside** the project: the path **string** of the link (e.g. `evil_link`) is still within root → **currently accepted** by `ValidatePath` (no resolution of the target). This is a **security gap** if symlink targets should be validated.
- `ValidatePathExists` calls `ValidatePath` then `os.Stat(absPath)`. For a symlink, `os.Stat` follows the link, so a link to a missing file will fail at "path does not exist".

---

## Test Cases

### Case 1: Symlink within project → target within project (accept)

| Item | Value |
|------|--------|
| **Setup** | `projectRoot/subdir/link_in` → `projectRoot/real_file.txt` (real file exists) |
| **Input** | `subdir/link_in` |
| **ValidatePath** | Accept (path string within root) |
| **ValidatePathExists** | Accept (target exists) |
| **Platform** | Unix/macOS; `os.Symlink` may need privileges on Windows |

### Case 2: Symlink within project → target outside project root (reject for strict security)

| Item | Value |
|------|--------|
| **Setup** | `projectRoot/evil_link` → `/etc/passwd` or `filepath.Join(tmpDir,"outside","secret")` |
| **Input** | `evil_link` |
| **Current ValidatePath** | Accepts (path string `projectRoot/evil_link` is within root; no `EvalSymlinks`) |
| **Desired (if using EvalSymlinks)** | Reject (resolved path outside root) |
| **ValidatePathExists** | Depends: if target exists, may return abs path of target (outside root) — should reject for security |
| **Platform** | Unix/macOS; use portable target for CI (e.g. temp dir sibling) |

### Case 3: Broken symlink

| Item | Value |
|------|--------|
| **Setup** | `projectRoot/broken_link` → `nonexistent_target` |
| **Input** | `broken_link` |
| **ValidatePath** | Accept (path string within root) |
| **ValidatePathExists** | Reject (path does not exist after Stat) |
| **Platform** | All |

### Case 4: Symlink in path component (nested)

| Item | Value |
|------|--------|
| **Setup** | `projectRoot/a/link` → `projectRoot/b/file`; resolve `a/link` |
| **ValidatePath** | Accept if resolved path stays within root (current impl: accept, no resolution) |
| **ValidatePathExists** | Accept if target exists |
| **Platform** | Unix/macOS |

---

## Implementation Notes

- Use `t.TempDir()`, `os.MkdirAll`, and `os.Symlink` for test setup.
- On Windows, `os.Symlink` may require elevated privileges or developer mode; skip symlink tests with `runtime.GOOS == "windows"` if needed, or use build tags.
- **Assert current behavior first**: add tests that document existing behavior (e.g. Case 2 currently accepted by ValidatePath), then in a follow-up (T-284) optionally add `filepath.EvalSymlinks` and update expected results.
- Reference: exarp-go [internal/security/path_test.go](internal/security/path_test.go), mcp-go-core [pkg/mcp/security/path_test.go](https://github.com/davidl71/mcp-go-core) (or vendor path).

---

## Epic / Related Tasks

- **T-284:** Test symlink handling (implementation of these cases)
- **T-1770829363301:** Implement symlink tests in exarp-go `internal/security/path_test.go`
- **T-1770829364114:** Implement symlink tests in mcp-go-core `pkg/mcp/security/path_test.go`

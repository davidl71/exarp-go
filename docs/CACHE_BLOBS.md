## SQLite `cache_blobs` (derived payload cache)

This project includes a small SQLite table intended for caching **derived payloads** that are:
- expensive to compute (extra DB queries, joins, aggregation)
- requested frequently (polling MCP clients / resources)
- safe to invalidate via **version tokens**

### Table

Created by migration `migrations/015_add_cache_blobs.sql` (schema version 15):

- `cache_key` (TEXT): logical identity (e.g. `task_id`)
- `cache_version` (INTEGER): version token for invalidation (e.g. `tasks.version`)
- `cache_kind` (TEXT): payload type discriminator (e.g. `task_execution_pack_v1`)
- `payload` (BLOB): cached bytes (often protobuf-encoded)
- `created_at` (INTEGER epoch seconds): insertion time

Primary key: `(cache_key, cache_version, cache_kind)`

### Invalidation model (version-based)

This cache is **versioned**, not TTL-based:
- when the underlying record changes, its version token changes
- clients will naturally miss the cache (new `(key, version, kind)` tuple)

This keeps invalidation simple and avoids “stale-but-fresh TTL” bugs.

### Current usage

- `stdio://agent/task/{task_id}/execution-pack` uses `task.version` as the token.
  - Cached payload is stored as a protobuf wrapper containing the JSON payload bytes.

### Guardrails (recommended)

- Treat this as a **small cache** (derived resources only).
- Prefer version tokens (task `version`, timestamps) over content hashes.
- If payloads get large or write-heavy, add eviction (e.g. keep latest N versions per key/kind).


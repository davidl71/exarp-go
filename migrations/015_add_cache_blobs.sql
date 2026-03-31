-- 015_add_cache_blobs.sql
-- Version: 15
-- Add a small SQLite BLOB cache for derived payloads keyed by (key, version, kind).
-- Use version-based invalidation so callers can skip recomputation when inputs don't change.

CREATE TABLE IF NOT EXISTS cache_blobs (
    cache_key TEXT NOT NULL,
    cache_version INTEGER NOT NULL,
    cache_kind TEXT NOT NULL,
    payload BLOB NOT NULL,
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    PRIMARY KEY (cache_key, cache_version, cache_kind)
);

CREATE INDEX IF NOT EXISTS idx_cache_blobs_key_kind_version
    ON cache_blobs(cache_key, cache_kind, cache_version DESC);


// cache_blobs.go — Small SQLite blob cache for derived payloads keyed by (key, version, kind).
package database

import (
	"context"
	"fmt"
)

// GetCacheBlob returns a cached payload for (key, version, kind) if present.
// A nil payload with nil error indicates a cache miss.
func GetCacheBlob(ctx context.Context, key string, version int64, kind string) ([]byte, error) {
	ctx = ensureContext(ctx)
	queryCtx, cancel, db, err := QueryContextDB(ctx)
	if err != nil {
		return nil, err
	}
	defer cancel()

	var payload []byte
	err = db.GetContext(
		queryCtx,
		&payload,
		`SELECT payload FROM cache_blobs WHERE cache_key = ? AND cache_version = ? AND cache_kind = ?`,
		key,
		version,
		kind,
	)
	if err != nil {
		// sqlite miss => treat as cache miss (do not fail callers)
		return nil, nil
	}

	return payload, nil
}

// PutCacheBlob stores a cached payload for (key, version, kind).
func PutCacheBlob(ctx context.Context, key string, version int64, kind string, payload []byte) error {
	ctx = ensureContext(ctx)
	queryCtx, cancel, db, err := QueryContextDB(ctx)
	if err != nil {
		return err
	}
	defer cancel()

	if key == "" || kind == "" {
		return fmt.Errorf("key and kind are required")
	}
	if payload == nil {
		return fmt.Errorf("payload is required")
	}

	_, err = db.ExecContext(
		queryCtx,
		`INSERT OR REPLACE INTO cache_blobs(cache_key, cache_version, cache_kind, payload) VALUES (?, ?, ?, ?)`,
		key,
		version,
		kind,
		payload,
	)
	if err != nil {
		return fmt.Errorf("failed to store cache blob: %w", err)
	}
	return nil
}

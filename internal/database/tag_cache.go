// tag_cache.go — In-memory tag cache for canonical tag lookups and discovery.
package database

import (
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"golang.org/x/sync/singleflight"
)

var tagCacheFlight singleflight.Group

// DiscoveredTag represents a tag discovered from a file.
type DiscoveredTag struct {
	ID           int64  `db:"id"`
	FilePath     string `db:"file_path"`
	FileHash     string `db:"file_hash"`
	Tag          string `db:"tag"`
	Source       string `db:"source"`
	LLMSuggested bool   `db:"llm_suggested"`
	CreatedAt    int64  `db:"created_at"`
	UpdatedAt    int64  `db:"updated_at"`
}

// TagFrequency represents tag usage statistics.
type TagFrequency struct {
	Tag         string `db:"tag"`
	Count       int    `db:"count"`
	LastSeenAt  *int64 `db:"last_seen_at"`
	IsCanonical bool   `db:"is_canonical"`
	CreatedAt   int64  `db:"created_at"`
	UpdatedAt   int64  `db:"updated_at"`
}

// FileTaskTag represents a file-to-task tag match.
type FileTaskTag struct {
	ID        int64
	FilePath  string
	TaskID    string
	Tag       string
	Applied   bool
	CreatedAt int64
}

// GetDiscoveredTagsForFile retrieves cached discovered tags for a file.
// Uses singleflight to deduplicate concurrent queries for the same file path.
func GetDiscoveredTagsForFile(filePath string) ([]DiscoveredTag, error) {
	v, err, _ := tagCacheFlight.Do("tags:"+filePath, func() (interface{}, error) {
		return getDiscoveredTagsForFileDB(filePath)
	})
	if err != nil {
		return nil, err
	}
	return v.([]DiscoveredTag), nil
}

func getDiscoveredTagsForFileDB(filePath string) ([]DiscoveredTag, error) {
	queryCtx, cancel, db, err := QueryContextDB(context.Background())
	if err != nil {
		return nil, err
	}
	defer cancel()

	var tags []DiscoveredTag
	if err := db.SelectContext(queryCtx, &tags, `
		SELECT id, file_path, file_hash, tag, source, llm_suggested, created_at, updated_at
		FROM discovered_tags
		WHERE file_path = ?
		ORDER BY id
	`, filePath); err != nil {
		return nil, fmt.Errorf("failed to query discovered tags: %w", err)
	}

	return tags, nil
}

// GetDiscoveredTagsWithHash retrieves cached discovered tags if file hash matches.
func GetDiscoveredTagsWithHash(filePath, currentHash string) ([]DiscoveredTag, bool, error) {
	queryCtx, cancel, db, err := QueryContextDB(context.Background())
	if err != nil {
		return nil, false, err
	}
	defer cancel()

	var storedHash string
	err = db.GetContext(queryCtx, &storedHash, `
		SELECT file_hash
		FROM discovered_tags
		WHERE file_path = ?
		LIMIT 1
	`, filePath)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("failed to query discovered hash: %w", err)
	}

	if storedHash != currentHash {
		return nil, false, nil
	}

	tags, err := GetDiscoveredTagsForFile(filePath)
	return tags, true, err
}

// SaveDiscoveredTags saves discovered tags for a file.
func SaveDiscoveredTags(filePath, fileHash string, tags []DiscoveredTag) (err error) {
	queryCtx, cancel, db, dbErr := QueryContextDB(context.Background())
	if dbErr != nil {
		return dbErr
	}
	defer cancel()

	tx, err := db.BeginTxx(queryCtx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.ExecContext(queryCtx, "DELETE FROM discovered_tags WHERE file_path = ?", filePath); err != nil {
		return fmt.Errorf("failed to delete existing tags: %w", err)
	}

	now := time.Now().Unix()

	for _, tag := range tags {
		llmSuggested := 0
		if tag.LLMSuggested {
			llmSuggested = 1
		}

		if _, err = tx.ExecContext(queryCtx, `
			INSERT INTO discovered_tags (file_path, file_hash, tag, source, llm_suggested, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, filePath, fileHash, tag.Tag, tag.Source, llmSuggested, now, now); err != nil {
			return fmt.Errorf("failed to insert tag: %w", err)
		}
	}

	return tx.Commit()
}

// UpdateTagFrequency updates the frequency count for a tag.
func UpdateTagFrequency(tag string, count int, isCanonical bool) error {
	queryCtx, cancel, db, err := QueryContextDB(context.Background())
	if err != nil {
		return err
	}
	defer cancel()

	now := time.Now().Unix()

	canonical := 0
	if isCanonical {
		canonical = 1
	}

	_, err = db.ExecContext(queryCtx, `
		INSERT INTO tag_frequency (tag, count, last_seen_at, is_canonical, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(tag) DO UPDATE SET
			count = excluded.count,
			last_seen_at = excluded.last_seen_at,
			is_canonical = excluded.is_canonical,
			updated_at = excluded.updated_at
	`, tag, count, now, canonical, now, now)

	return err
}

// GetTagFrequencies retrieves tag frequencies.
func GetTagFrequencies() ([]TagFrequency, error) {
	queryCtx, cancel, db, err := QueryContextDB(context.Background())
	if err != nil {
		return nil, err
	}
	defer cancel()

	var frequencies []TagFrequency
	if err := db.SelectContext(queryCtx, &frequencies, `
		SELECT tag, count, last_seen_at, is_canonical, created_at, updated_at
		FROM tag_frequency
		ORDER BY count DESC
	`); err != nil {
		return nil, fmt.Errorf("failed to query tag frequencies: %w", err)
	}

	return frequencies, nil
}

// SaveFileTaskTag saves a file-to-task tag match.
func SaveFileTaskTag(filePath, taskID, tag string, applied bool) error {
	queryCtx, cancel, db, err := QueryContextDB(context.Background())
	if err != nil {
		return err
	}
	defer cancel()

	appliedInt := 0
	if applied {
		appliedInt = 1
	}

	now := time.Now().Unix()

	_, err = db.ExecContext(queryCtx, `
		INSERT INTO file_task_tags (file_path, task_id, tag, applied, created_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(file_path, task_id, tag) DO UPDATE SET
			applied = excluded.applied
	`, filePath, taskID, tag, appliedInt, now)

	return err
}

// ComputeFileHash computes MD5 hash of file content.
func ComputeFileHash(content []byte) string {
	hash := md5.Sum(content)
	return hex.EncodeToString(hash[:])
}

// ClearDiscoveredTagsCache clears all discovered tag cache entries.
func ClearDiscoveredTagsCache() error {
	queryCtx, cancel, db, err := QueryContextDB(context.Background())
	if err != nil {
		return err
	}
	defer cancel()

	_, err = db.ExecContext(queryCtx, "DELETE FROM discovered_tags")

	return err
}

// SaveTaskTagSuggestion saves a task-level tag suggestion (from action=tags) for reuse as LLM hint.
func SaveTaskTagSuggestion(taskID, tag, source string, applied bool) error {
	queryCtx, cancel, db, err := QueryContextDB(context.Background())
	if err != nil {
		return err
	}
	defer cancel()

	appliedInt := 0
	if applied {
		appliedInt = 1
	}

	now := time.Now().Unix()
	_, err = db.ExecContext(queryCtx, `
		INSERT INTO task_tag_suggestions (task_id, tag, source, applied, created_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(task_id, tag) DO UPDATE SET source = excluded.source, applied = excluded.applied
	`, taskID, tag, source, appliedInt, now)

	return err
}

// GetTaskTagSuggestions returns cached tag suggestions for a task (for LLM hints).
func GetTaskTagSuggestions(taskID string) ([]string, error) {
	queryCtx, cancel, db, err := QueryContextDB(context.Background())
	if err != nil {
		return nil, err
	}
	defer cancel()

	var tags []string
	if err := db.SelectContext(queryCtx, &tags, `
		SELECT tag FROM task_tag_suggestions
		WHERE task_id = ?
		ORDER BY created_at
	`, taskID); err != nil {
		return nil, fmt.Errorf("failed to query task tag suggestions: %w", err)
	}

	return tags, nil
}

// GetTopTagFrequencies returns the top N tag names by count from the cache (for LLM hint list).
func GetTopTagFrequencies(limit int) ([]string, error) {
	freqs, err := GetTagFrequencies()
	if err != nil {
		return nil, err
	}

	if limit <= 0 {
		limit = 30
	}

	tags := make([]string, 0, limit)
	for i := 0; i < len(freqs) && i < limit; i++ {
		tags = append(tags, freqs[i].Tag)
	}

	return tags, nil
}

// ClearTaskTagSuggestions clears tag suggestions for a specific task (call on task delete/update).
func ClearTaskTagSuggestions(taskID string) error {
	queryCtx, cancel, db, err := QueryContextDB(context.Background())
	if err != nil {
		return err
	}
	defer cancel()

	_, err = db.ExecContext(queryCtx, "DELETE FROM task_tag_suggestions WHERE task_id = ?", taskID)
	return err
}

// ClearFileTaskTags clears file-to-task tag mappings for a specific task (call on task delete/update).
func ClearFileTaskTags(taskID string) error {
	queryCtx, cancel, db, err := QueryContextDB(context.Background())
	if err != nil {
		return err
	}
	defer cancel()

	_, err = db.ExecContext(queryCtx, "DELETE FROM file_task_tags WHERE task_id = ?", taskID)
	return err
}

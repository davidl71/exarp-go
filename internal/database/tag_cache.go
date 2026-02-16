package database

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"time"
)

// CanonicalTag represents a canonical tag mapping.
type CanonicalTag struct {
	OldTag    string
	NewTag    string
	Category  string
	CreatedAt int64
	UpdatedAt int64
}

// DiscoveredTag represents a tag discovered from a file.
type DiscoveredTag struct {
	ID           int64
	FilePath     string
	FileHash     string
	Tag          string
	Source       string
	LLMSuggested bool
	CreatedAt    int64
	UpdatedAt    int64
}

// TagFrequency represents tag usage statistics.
type TagFrequency struct {
	Tag         string
	Count       int
	LastSeenAt  *int64
	IsCanonical bool
	CreatedAt   int64
	UpdatedAt   int64
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

// GetCanonicalTags retrieves all canonical tag mappings from the database.
func GetCanonicalTags() (map[string]string, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}

	rows, err := db.Query("SELECT old_tag, new_tag FROM canonical_tags")
	if err != nil {
		return nil, fmt.Errorf("failed to query canonical tags: %w", err)
	}
	defer rows.Close()

	tags := make(map[string]string)

	for rows.Next() {
		var oldTag, newTag string
		if err := rows.Scan(&oldTag, &newTag); err != nil {
			continue
		}

		tags[oldTag] = newTag
	}

	return tags, nil
}

// GetCanonicalTagsByCategory retrieves canonical tags by category.
func GetCanonicalTagsByCategory(category string) ([]CanonicalTag, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}

	rows, err := db.Query("SELECT old_tag, new_tag, category, created_at, updated_at FROM canonical_tags WHERE category = ?", category)
	if err != nil {
		return nil, fmt.Errorf("failed to query canonical tags: %w", err)
	}
	defer rows.Close()

	var tags []CanonicalTag

	for rows.Next() {
		var tag CanonicalTag
		if err := rows.Scan(&tag.OldTag, &tag.NewTag, &tag.Category, &tag.CreatedAt, &tag.UpdatedAt); err != nil {
			continue
		}

		tags = append(tags, tag)
	}

	return tags, nil
}

// UpsertCanonicalTag inserts or updates a canonical tag mapping.
func UpsertCanonicalTag(oldTag, newTag, category string) error {
	db, err := GetDB()
	if err != nil {
		return err
	}

	now := time.Now().Unix()
	_, err = db.Exec(`
		INSERT INTO canonical_tags (old_tag, new_tag, category, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(old_tag) DO UPDATE SET
			new_tag = excluded.new_tag,
			category = excluded.category,
			updated_at = excluded.updated_at
	`, oldTag, newTag, category, now, now)

	return err
}

// GetDiscoveredTagsForFile retrieves cached discovered tags for a file.
func GetDiscoveredTagsForFile(filePath string) ([]DiscoveredTag, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(`
		SELECT id, file_path, file_hash, tag, source, llm_suggested, created_at, updated_at
		FROM discovered_tags WHERE file_path = ?
	`, filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to query discovered tags: %w", err)
	}
	defer rows.Close()

	var tags []DiscoveredTag

	for rows.Next() {
		var tag DiscoveredTag

		var llmSuggested int
		if err := rows.Scan(&tag.ID, &tag.FilePath, &tag.FileHash, &tag.Tag, &tag.Source, &llmSuggested, &tag.CreatedAt, &tag.UpdatedAt); err != nil {
			continue
		}

		tag.LLMSuggested = llmSuggested == 1
		tags = append(tags, tag)
	}

	return tags, nil
}

// GetDiscoveredTagsWithHash retrieves cached discovered tags if file hash matches.
func GetDiscoveredTagsWithHash(filePath, currentHash string) ([]DiscoveredTag, bool, error) {
	db, err := GetDB()
	if err != nil {
		return nil, false, fmt.Errorf("database not initialized")
	}

	// Check if any tags exist for this file with matching hash
	var storedHash string

	err = db.QueryRow("SELECT file_hash FROM discovered_tags WHERE file_path = ? LIMIT 1", filePath).Scan(&storedHash)
	if err != nil {
		return nil, false, nil // No cache exists
	}

	if storedHash != currentHash {
		return nil, false, nil // Cache is stale
	}

	// Cache is valid, return cached tags
	tags, err := GetDiscoveredTagsForFile(filePath)

	return tags, true, err
}

// SaveDiscoveredTags saves discovered tags for a file.
func SaveDiscoveredTags(filePath, fileHash string, tags []DiscoveredTag) error {
	db, err := GetDB()
	if err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete existing tags for this file
	_, err = tx.Exec("DELETE FROM discovered_tags WHERE file_path = ?", filePath)
	if err != nil {
		return fmt.Errorf("failed to delete existing tags: %w", err)
	}

	// Insert new tags
	now := time.Now().Unix()

	for _, tag := range tags {
		llmSuggested := 0
		if tag.LLMSuggested {
			llmSuggested = 1
		}

		_, err = tx.Exec(`
			INSERT INTO discovered_tags (file_path, file_hash, tag, source, llm_suggested, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, filePath, fileHash, tag.Tag, tag.Source, llmSuggested, now, now)
		if err != nil {
			return fmt.Errorf("failed to insert tag: %w", err)
		}
	}

	return tx.Commit()
}

// GetAllDiscoveredTags retrieves all discovered tags.
func GetAllDiscoveredTags() ([]DiscoveredTag, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(`
		SELECT id, file_path, file_hash, tag, source, llm_suggested, created_at, updated_at
		FROM discovered_tags ORDER BY file_path, tag
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query discovered tags: %w", err)
	}
	defer rows.Close()

	var tags []DiscoveredTag

	for rows.Next() {
		var tag DiscoveredTag

		var llmSuggested int
		if err := rows.Scan(&tag.ID, &tag.FilePath, &tag.FileHash, &tag.Tag, &tag.Source, &llmSuggested, &tag.CreatedAt, &tag.UpdatedAt); err != nil {
			continue
		}

		tag.LLMSuggested = llmSuggested == 1
		tags = append(tags, tag)
	}

	return tags, nil
}

// UpdateTagFrequency updates the frequency count for a tag.
func UpdateTagFrequency(tag string, count int, isCanonical bool) error {
	db, err := GetDB()
	if err != nil {
		return err
	}

	now := time.Now().Unix()

	canonical := 0
	if isCanonical {
		canonical = 1
	}

	_, err = db.Exec(`
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
	db, err := GetDB()
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(`
		SELECT tag, count, last_seen_at, is_canonical, created_at, updated_at
		FROM tag_frequency ORDER BY count DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query tag frequencies: %w", err)
	}
	defer rows.Close()

	var frequencies []TagFrequency

	for rows.Next() {
		var freq TagFrequency

		var isCanonical int
		if err := rows.Scan(&freq.Tag, &freq.Count, &freq.LastSeenAt, &isCanonical, &freq.CreatedAt, &freq.UpdatedAt); err != nil {
			continue
		}

		freq.IsCanonical = isCanonical == 1
		frequencies = append(frequencies, freq)
	}

	return frequencies, nil
}

// SaveFileTaskTag saves a file-to-task tag match.
func SaveFileTaskTag(filePath, taskID, tag string, applied bool) error {
	db, err := GetDB()
	if err != nil {
		return err
	}

	appliedInt := 0
	if applied {
		appliedInt = 1
	}

	now := time.Now().Unix()

	_, err = db.Exec(`
		INSERT INTO file_task_tags (file_path, task_id, tag, applied, created_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(file_path, task_id, tag) DO UPDATE SET
			applied = excluded.applied
	`, filePath, taskID, tag, appliedInt, now)

	return err
}

// GetFileTaskTags retrieves file-to-task tag matches for a task.
func GetFileTaskTags(taskID string) ([]FileTaskTag, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(`
		SELECT id, file_path, task_id, tag, applied, created_at
		FROM file_task_tags WHERE task_id = ?
	`, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to query file task tags: %w", err)
	}
	defer rows.Close()

	var tags []FileTaskTag

	for rows.Next() {
		var tag FileTaskTag

		var applied int
		if err := rows.Scan(&tag.ID, &tag.FilePath, &tag.TaskID, &tag.Tag, &applied, &tag.CreatedAt); err != nil {
			continue
		}

		tag.Applied = applied == 1
		tags = append(tags, tag)
	}

	return tags, nil
}

// ComputeFileHash computes MD5 hash of file content.
func ComputeFileHash(content []byte) string {
	hash := md5.Sum(content)
	return hex.EncodeToString(hash[:])
}

// ClearDiscoveredTagsCache clears all discovered tag cache entries.
func ClearDiscoveredTagsCache() error {
	db, err := GetDB()
	if err != nil {
		return err
	}

	_, err = db.Exec("DELETE FROM discovered_tags")

	return err
}

// ClearFileTaskTagsCache clears all file-task tag matches.
func ClearFileTaskTagsCache() error {
	db, err := GetDB()
	if err != nil {
		return err
	}

	_, err = db.Exec("DELETE FROM file_task_tags")

	return err
}

// SaveTaskTagSuggestion saves a task-level tag suggestion (from action=tags) for reuse as LLM hint.
func SaveTaskTagSuggestion(taskID, tag, source string, applied bool) error {
	db, err := GetDB()
	if err != nil {
		return err
	}

	appliedInt := 0
	if applied {
		appliedInt = 1
	}

	now := time.Now().Unix()
	_, err = db.Exec(`
		INSERT INTO task_tag_suggestions (task_id, tag, source, applied, created_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(task_id, tag) DO UPDATE SET source = excluded.source, applied = excluded.applied
	`, taskID, tag, source, appliedInt, now)

	return err
}

// GetTaskTagSuggestions returns cached tag suggestions for a task (for LLM hints).
func GetTaskTagSuggestions(taskID string) ([]string, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}

	rows, err := db.Query("SELECT tag FROM task_tag_suggestions WHERE task_id = ? ORDER BY created_at", taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to query task tag suggestions: %w", err)
	}

	defer rows.Close()

	var tags []string

	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			continue
		}

		tags = append(tags, tag)
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

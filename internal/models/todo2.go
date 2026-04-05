// Package models provides shared types, constants, and task ID utilities used across packages.
package models

import (
	"math"
	"strings"
	"time"
)

// ClampPriorityRankForProto coerces priority_rank to protobuf int32 range.
func ClampPriorityRankForProto(n int) int32 {
	if n > math.MaxInt32 {
		return math.MaxInt32
	}
	if n < math.MinInt32 {
		return math.MinInt32
	}
	return int32(n)
}

// IsEpochDate returns true if s is empty or is the Unix epoch (1970-01-01 or numeric 0).
// Used to avoid displaying or persisting 1/1/1970 when the real date is unknown.
func IsEpochDate(s string) bool {
	if s == "" {
		return true
	}

	s = strings.TrimSpace(s)
	// Numeric zero (from JSON "created_at": 0 or DB default)
	if s == "0" || s == "0.0" {
		return true
	}
	// 1970-01-01 in any common format (UTC Z, or with time)
	return strings.HasPrefix(s, "1970-01-01")
}

// NormalizeEpochDates sets CreatedAt, LastModified, and CompletedAt to empty
// when they are the Unix epoch (1970-01-01), so we never display or persist 1/1/1970.
func (t *Todo2Task) NormalizeEpochDates() {
	if IsEpochDate(t.CreatedAt) {
		t.CreatedAt = ""
	}

	if IsEpochDate(t.LastModified) {
		t.LastModified = ""
	}

	if IsEpochDate(t.CompletedAt) {
		t.CompletedAt = ""
	}
}

// FillRFC3339FromUnix sets CreatedAt, LastModified, and CompletedAt from Unix
// seconds when the string field is empty or epoch-like and the corresponding
// *_ts value is positive. Used when loading from SQLite so JSON/export matches
// the integer timeline columns if legacy text fields were never backfilled.
func (t *Todo2Task) FillRFC3339FromUnix(createdTS, lastModifiedTS, completedAtTS int64) {
	if IsEpochDate(t.CreatedAt) && createdTS > 0 {
		t.CreatedAt = time.Unix(createdTS, 0).UTC().Format(time.RFC3339)
	}
	if IsEpochDate(t.LastModified) && lastModifiedTS > 0 {
		t.LastModified = time.Unix(lastModifiedTS, 0).UTC().Format(time.RFC3339)
	}
	if IsEpochDate(t.CompletedAt) && completedAtTS > 0 {
		t.CompletedAt = time.Unix(completedAtTS, 0).UTC().Format(time.RFC3339)
	}
}

// FillRFC3339FromSQLiteTimes fills display timestamps like FillRFC3339FromUnix, but when
// created_ts or last_modified_ts is zero uses the legacy integer columns created_at /
// updated_at on the same row (schema v9 / migration 016). Some older writes only bumped
// those columns or left *_ts at default 0.
func (t *Todo2Task) FillRFC3339FromSQLiteTimes(createdTS, lastModifiedTS, completedAtTS, legacyCreatedAt, legacyUpdatedAt int64) {
	created := createdTS
	if created <= 0 && legacyCreatedAt > 0 {
		created = legacyCreatedAt
	}
	lastMod := lastModifiedTS
	if lastMod <= 0 && legacyUpdatedAt > 0 {
		lastMod = legacyUpdatedAt
	}
	t.FillRFC3339FromUnix(created, lastMod, completedAtTS)
}

// Todo2Task represents a Todo2 task.
type Todo2Task struct {
	ID string `json:"id"`
	// Name is a short summary/title intended for list views and external consumers.
	// Content remains the canonical task title used throughout the app and CLI.
	Name            string `json:"name,omitempty"`
	Content         string `json:"content"`
	LongDescription string `json:"long_description,omitempty"`
	Status          string `json:"status"`
	// StatusEnum is the internal typed status; it is derived from Status on load.
	// It is not part of the canonical JSON shape.
	StatusEnum TaskStatus `json:"-"`
	Priority   string     `json:"priority,omitempty"`
	// PriorityRank is a numeric sort key within the same named priority (lower = earlier).
	// Default 0 when unset.
	PriorityRank int `json:"priority_rank,omitempty"`
	// PriorityEnum is the internal typed priority; it is derived from Priority on load.
	// It is not part of the canonical JSON shape.
	PriorityEnum TaskPriority           `json:"-"`
	Tags         []string               `json:"tags,omitempty"`
	Dependencies []string               `json:"dependencies,omitempty"`
	ParentID     string                 `json:"parent_id,omitempty"` // Parent task ID (epic or container); hierarchy, not blocking
	Completed    bool                   `json:"completed,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	// CreatedAt, LastModified, CompletedAt are RFC3339 timestamps from DB/JSON; preserved on load/save.
	CreatedAt    string `json:"created_at,omitempty"`
	LastModified string `json:"last_modified,omitempty"`
	CompletedAt  string `json:"completed_at,omitempty"`
	// Distributed tracking (for aggregation across projects/hosts/agents)
	ProjectID  string `json:"project_id,omitempty"`  // Logical project identifier (e.g. "exarp-go")
	AssignedTo string `json:"assigned_to,omitempty"` // Persistent assignee (owner); distinct from lock assignee
	Host       string `json:"host,omitempty"`        // Hostname where task was created or last modified
	Agent      string `json:"agent,omitempty"`       // Agent ID that created or last modified (e.g. general-host-pid)
	// Version is the SQLite optimistic-lock version when the task was loaded from the DB.
	// When provided back to UpdateTask, it skips the pre-UPDATE SELECT round-trip.
	Version int64 `json:"version,omitempty"`
}

// EnsureName populates Name when empty using Content (fallback: LongDescription).
// It normalizes whitespace and truncates to a small, list-friendly length.
func (t *Todo2Task) EnsureName() {
	if strings.TrimSpace(t.Name) != "" {
		return
	}

	base := strings.TrimSpace(t.Content)
	if base == "" {
		base = strings.TrimSpace(t.LongDescription)
	}
	if base == "" {
		return
	}

	// Collapse whitespace.
	base = strings.Join(strings.Fields(base), " ")

	// Truncate to ~120 runes (UTF-8 safe).
	const maxRunes = 120
	if runeCount(base) > maxRunes {
		base = truncateRunes(base, maxRunes-3) + "..."
	}

	t.Name = base
}

func runeCount(s string) int {
	n := 0
	for range s {
		n++
	}
	return n
}

func truncateRunes(s string, limit int) string {
	if limit <= 0 {
		return ""
	}
	n := 0
	for i := range s {
		if n == limit {
			return s[:i]
		}
		n++
	}
	return s
}

// Todo2State represents the Todo2 state file structure.
type Todo2State struct {
	Todos []Todo2Task `json:"todos"`
}

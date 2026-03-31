// Package models provides shared types, constants, and task ID utilities used across packages.
package models

import "strings"

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

// Todo2Task represents a Todo2 task.
type Todo2Task struct {
	ID              string                 `json:"id"`
	// Name is a short summary/title intended for list views and external consumers.
	// Content remains the canonical task title used throughout the app and CLI.
	Name            string                 `json:"name,omitempty"`
	Content         string                 `json:"content"`
	LongDescription string                 `json:"long_description,omitempty"`
	Status          string                 `json:"status"`
	// StatusEnum is the internal typed status; it is derived from Status on load.
	// It is not part of the canonical JSON shape.
	StatusEnum      TaskStatus             `json:"-"`
	Priority        string                 `json:"priority,omitempty"`
	// PriorityEnum is the internal typed priority; it is derived from Priority on load.
	// It is not part of the canonical JSON shape.
	PriorityEnum    TaskPriority           `json:"-"`
	Tags            []string               `json:"tags,omitempty"`
	Dependencies    []string               `json:"dependencies,omitempty"`
	ParentID        string                 `json:"parent_id,omitempty"` // Parent task ID (epic or container); hierarchy, not blocking
	Completed       bool                   `json:"completed,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
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

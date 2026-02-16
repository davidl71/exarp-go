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
	Content         string                 `json:"content"`
	LongDescription string                 `json:"long_description,omitempty"`
	Status          string                 `json:"status"`
	Priority        string                 `json:"priority,omitempty"`
	Tags            []string               `json:"tags,omitempty"`
	Dependencies    []string               `json:"dependencies,omitempty"`
	ParentID        string                 `json:"parent_id,omitempty"` // Parent task ID (epic or container); hierarchy, not blocking
	Completed       bool                   `json:"completed,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	// CreatedAt, LastModified, CompletedAt are RFC3339 timestamps from DB/JSON; preserved on load/save.
	CreatedAt    string `json:"created_at,omitempty"`
	LastModified string `json:"last_modified,omitempty"`
	CompletedAt  string `json:"completed_at,omitempty"`
}

// Todo2State represents the Todo2 state file structure.
type Todo2State struct {
	Todos []Todo2Task `json:"todos"`
}

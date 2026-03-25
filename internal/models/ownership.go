// Package models provides shared types, constants, and task ID utilities used across packages.
package models

// ============================================================================
// Task Ownership Metadata
// ============================================================================
//
// This file defines helpers for reading/writing task ownership metadata.
// Ownership metadata lives in Task.Metadata under the "ownership" key.
//
// Structure stored in metadata["ownership"]:
//
//	{
//	  "owned_files": ["path/to/file.go", "path/to/other.go"],
//	  "owned_globs": ["internal/tools/*.go"],
//	  "forbidden_files": ["path/to/critical.go"],
//	  "ownership_confidence": "explicit", // "explicit" | "inferred" | "unknown"
//	  "lane": "backend-health"
//	}
//
// Reference: docs/TASK_LANES_AND_FILE_OWNERSHIP_PLAN.md

// OwnershipConfidence levels for task ownership inference.
const (
	OwnershipConfidenceExplicit = "explicit" // User or tool declared exact ownership
	OwnershipConfidenceInferred = "inferred" // Heuristics from task text/tags
	OwnershipConfidenceUnknown  = "unknown"  // No usable signal
)

// TaskOwnership holds file ownership metadata for a task.
// All fields are optional; empty/nil means "not set".
type TaskOwnership struct {
	OwnedFiles          []string `json:"owned_files,omitempty"`
	OwnedGlobs          []string `json:"owned_globs,omitempty"`
	ForbiddenFiles      []string `json:"forbidden_files,omitempty"`
	OwnershipConfidence string   `json:"ownership_confidence,omitempty"` // "explicit" | "inferred" | "unknown"
	Lane                string   `json:"lane,omitempty"`                 // e.g. "tui-shell", "backend-health", "docs"
}

// ownershipMetaKey is the metadata key under which ownership data is stored.
const ownershipMetaKey = "ownership"

// GetTaskOwnership returns the ownership metadata for a task, or nil if not set.
func GetTaskOwnership(task *Todo2Task) *TaskOwnership {
	if task == nil || task.Metadata == nil {
		return nil
	}

	raw, ok := task.Metadata[ownershipMetaKey]
	if !ok {
		return nil
	}

	// Handle map[string]interface{} (JSON unmarshaled)
	if m, ok := raw.(map[string]interface{}); ok {
		own := &TaskOwnership{}
		if v, ok := m["owned_files"].([]interface{}); ok {
			for _, item := range v {
				if s, ok := item.(string); ok {
					own.OwnedFiles = append(own.OwnedFiles, s)
				}
			}
		}
		if v, ok := m["owned_globs"].([]interface{}); ok {
			for _, item := range v {
				if s, ok := item.(string); ok {
					own.OwnedGlobs = append(own.OwnedGlobs, s)
				}
			}
		}
		if v, ok := m["forbidden_files"].([]interface{}); ok {
			for _, item := range v {
				if s, ok := item.(string); ok {
					own.ForbiddenFiles = append(own.ForbiddenFiles, s)
				}
			}
		}
		if v, ok := m["ownership_confidence"].(string); ok {
			own.OwnershipConfidence = v
		}
		if v, ok := m["lane"].(string); ok {
			own.Lane = v
		}
		return own
	}

	return nil
}

// SetTaskOwnership sets the ownership metadata on a task.
// Initializes task.Metadata if nil.
func SetTaskOwnership(task *Todo2Task, own *TaskOwnership) {
	if task == nil {
		return
	}

	if task.Metadata == nil {
		task.Metadata = make(map[string]interface{})
	}

	if own == nil {
		delete(task.Metadata, ownershipMetaKey)
		return
	}

	m := make(map[string]interface{})
	if len(own.OwnedFiles) > 0 {
		m["owned_files"] = own.OwnedFiles
	}
	if len(own.OwnedGlobs) > 0 {
		m["owned_globs"] = own.OwnedGlobs
	}
	if len(own.ForbiddenFiles) > 0 {
		m["forbidden_files"] = own.ForbiddenFiles
	}
	if own.OwnershipConfidence != "" {
		m["ownership_confidence"] = own.OwnershipConfidence
	}
	if own.Lane != "" {
		m["lane"] = own.Lane
	}

	task.Metadata[ownershipMetaKey] = m
}

// GetOwnershipConfidence returns the ownership confidence level for a task.
// Returns "unknown" if ownership metadata is not set.
func GetOwnershipConfidence(task *Todo2Task) string {
	own := GetTaskOwnership(task)
	if own == nil {
		return OwnershipConfidenceUnknown
	}
	if own.OwnershipConfidence == "" {
		return OwnershipConfidenceUnknown
	}
	return own.OwnershipConfidence
}

// GetTaskLane returns the lane label for a task, or empty string if not set.
func GetTaskLane(task *Todo2Task) string {
	own := GetTaskOwnership(task)
	if own == nil {
		return ""
	}
	return own.Lane
}

// GetOwnedFiles returns the list of files owned by a task.
func GetOwnedFiles(task *Todo2Task) []string {
	own := GetTaskOwnership(task)
	if own == nil {
		return nil
	}
	return own.OwnedFiles
}

// GetOwnedGlobs returns the glob patterns for files owned by a task.
func GetOwnedGlobs(task *Todo2Task) []string {
	own := GetTaskOwnership(task)
	if own == nil {
		return nil
	}
	return own.OwnedGlobs
}

// GetForbiddenFiles returns the list of files the task should avoid.
func GetForbiddenFiles(task *Todo2Task) []string {
	own := GetTaskOwnership(task)
	if own == nil {
		return nil
	}
	return own.ForbiddenFiles
}

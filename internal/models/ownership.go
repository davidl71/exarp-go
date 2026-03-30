// Package models provides shared types, constants, and task ID utilities used across packages.
package models

import "slices"

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
		m["owned_files"] = slices.Clone(own.OwnedFiles)
	}
	if len(own.OwnedGlobs) > 0 {
		m["owned_globs"] = slices.Clone(own.OwnedGlobs)
	}
	if len(own.ForbiddenFiles) > 0 {
		m["forbidden_files"] = slices.Clone(own.ForbiddenFiles)
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

// ============================================================================
// Hotspot and File Lease Tracking
// ============================================================================

// HotspotFile represents a file that is frequently edited or contested.
type HotspotFile struct {
	Path        string `json:"path"`
	EditCount   int    `json:"edit_count"`             // Number of recent edits
	TaskCount   int    `json:"task_count"`             // Number of tasks that touch this file
	IsLeased    bool   `json:"is_leased"`              // Currently leased by an active run
	LeasedBy    string `json:"leased_by,omitempty"`    // Task/run ID that holds the lease
	LeaseExpiry string `json:"lease_expiry,omitempty"` // When lease expires (RFC3339)
}

// FileLease represents an exclusive lock on a file by an active task run.
type FileLease struct {
	FilePath    string `json:"file_path"`
	TaskID      string `json:"task_id"`
	RunID       string `json:"run_id"`
	AgentID     string `json:"agent_id"`
	AcquiredAt  string `json:"acquired_at"` // RFC3339
	ExpiresAt   string `json:"expires_at"`  // RFC3339
	Description string `json:"description,omitempty"`
}

// ProjectHotspots holds the hotspot analysis for a project.
type ProjectHotspots struct {
	ProjectRoot string        `json:"project_root"`
	AnalyzedAt  string        `json:"analyzed_at"`
	Hotspots    []HotspotFile `json:"hotspots"`
	TotalFiles  int           `json:"total_files"`
	HighRisk    []string      `json:"high_risk_files"` // Files with edit_count > threshold
}

// metadata keys for hotspots and leases
const (
	MetadataKeyHotspots   = "hotspots"
	MetadataKeyFileLeases = "file_leases"
)

// GetProjectHotspots returns hotspot analysis from task metadata.
func GetProjectHotspots(metadata map[string]interface{}) *ProjectHotspots {
	if metadata == nil {
		return nil
	}
	raw, ok := metadata[MetadataKeyHotspots]
	if !ok {
		return nil
	}
	if m, ok := raw.(map[string]interface{}); ok {
		hp := &ProjectHotspots{}
		if v, ok := m["project_root"].(string); ok {
			hp.ProjectRoot = v
		}
		if v, ok := m["analyzed_at"].(string); ok {
			hp.AnalyzedAt = v
		}
		if v, ok := m["total_files"].(float64); ok {
			hp.TotalFiles = int(v)
		}
		if arr, ok := m["high_risk_files"].([]interface{}); ok {
			for _, item := range arr {
				if s, ok := item.(string); ok {
					hp.HighRisk = append(hp.HighRisk, s)
				}
			}
		}
		if arr, ok := m["hotspots"].([]interface{}); ok {
			for _, item := range arr {
				if m2, ok := item.(map[string]interface{}); ok {
					hf := HotspotFile{}
					if v, ok := m2["path"].(string); ok {
						hf.Path = v
					}
					if v, ok := m2["edit_count"].(float64); ok {
						hf.EditCount = int(v)
					}
					if v, ok := m2["task_count"].(float64); ok {
						hf.TaskCount = int(v)
					}
					if v, ok := m2["is_leased"].(bool); ok {
						hf.IsLeased = v
					}
					if v, ok := m2["leased_by"].(string); ok {
						hf.LeasedBy = v
					}
					hp.Hotspots = append(hp.Hotspots, hf)
				}
			}
		}
		return hp
	}
	return nil
}

// SetProjectHotspots stores hotspot analysis in metadata.
func SetProjectHotspots(metadata map[string]interface{}, hp *ProjectHotspots) {
	if metadata == nil || hp == nil {
		return
	}
	m := map[string]interface{}{
		"project_root": hp.ProjectRoot,
		"analyzed_at":  hp.AnalyzedAt,
		"total_files":  hp.TotalFiles,
	}
	if len(hp.HighRisk) > 0 {
		m["high_risk_files"] = hp.HighRisk
	}
	if len(hp.Hotspots) > 0 {
		hotspots := make([]map[string]interface{}, 0, len(hp.Hotspots))
		for _, hf := range hp.Hotspots {
			hm := map[string]interface{}{
				"path":       hf.Path,
				"edit_count": hf.EditCount,
				"task_count": hf.TaskCount,
				"is_leased":  hf.IsLeased,
			}
			if hf.LeasedBy != "" {
				hm["leased_by"] = hf.LeasedBy
			}
			if hf.LeaseExpiry != "" {
				hm["lease_expiry"] = hf.LeaseExpiry
			}
			hotspots = append(hotspots, hm)
		}
		m["hotspots"] = hotspots
	}
	metadata[MetadataKeyHotspots] = m
}

// GetFileLeases returns active file leases from metadata.
func GetFileLeases(metadata map[string]interface{}) []FileLease {
	if metadata == nil {
		return nil
	}
	raw, ok := metadata[MetadataKeyFileLeases]
	if !ok {
		return nil
	}
	arr, ok := raw.([]interface{})
	if !ok {
		return nil
	}
	var leases []FileLease
	for _, item := range arr {
		if m, ok := item.(map[string]interface{}); ok {
			lease := FileLease{}
			if v, ok := m["file_path"].(string); ok {
				lease.FilePath = v
			}
			if v, ok := m["task_id"].(string); ok {
				lease.TaskID = v
			}
			if v, ok := m["run_id"].(string); ok {
				lease.RunID = v
			}
			if v, ok := m["agent_id"].(string); ok {
				lease.AgentID = v
			}
			if v, ok := m["acquired_at"].(string); ok {
				lease.AcquiredAt = v
			}
			if v, ok := m["expires_at"].(string); ok {
				lease.ExpiresAt = v
			}
			leases = append(leases, lease)
		}
	}
	return leases
}

// SetFileLeases stores file leases in metadata.
func SetFileLeases(metadata map[string]interface{}, leases []FileLease) {
	if metadata == nil {
		return
	}
	if len(leases) == 0 {
		delete(metadata, MetadataKeyFileLeases)
		return
	}
	arr := make([]map[string]interface{}, 0, len(leases))
	for _, lease := range leases {
		m := map[string]interface{}{
			"file_path":   lease.FilePath,
			"task_id":     lease.TaskID,
			"run_id":      lease.RunID,
			"agent_id":    lease.AgentID,
			"acquired_at": lease.AcquiredAt,
			"expires_at":  lease.ExpiresAt,
		}
		if lease.Description != "" {
			m["description"] = lease.Description
		}
		arr = append(arr, m)
	}
	metadata[MetadataKeyFileLeases] = arr
}

// IsFileLeased checks if a file has an active lease.
func IsFileLeased(metadata map[string]interface{}, filePath string) (bool, string) {
	leases := GetFileLeases(metadata)
	for _, lease := range leases {
		if lease.FilePath == filePath {
			// Check if lease is still valid (simplified - no time parsing)
			return true, lease.TaskID
		}
	}
	return false, ""
}

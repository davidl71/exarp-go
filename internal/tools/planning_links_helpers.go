package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/davidl71/exarp-go/internal/models"
	"github.com/davidl71/exarp-go/internal/security"
)

// PlanningLinkMetadata represents planning document link metadata stored in task metadata.
type PlanningLinkMetadata struct {
	PlanningDoc string   `json:"planning_doc,omitempty"` // Path to planning document
	EpicID      string   `json:"epic_id,omitempty"`      // Epic task ID if part of an epic
	RelatedDocs []string `json:"related_docs,omitempty"` // Related planning document paths
	TaskRefs    []string `json:"task_refs,omitempty"`    // Task IDs referenced in planning doc
	EpicRefs    []string `json:"epic_refs,omitempty"`    // Epic IDs referenced in planning doc
	DocType     string   `json:"doc_type,omitempty"`     // Planning document type
	LastSynced  string   `json:"last_synced,omitempty"`  // Last sync timestamp
}

// SetPlanningLinkMetadata stores planning document link metadata in task metadata.
func SetPlanningLinkMetadata(task *models.Todo2Task, linkMeta *PlanningLinkMetadata) {
	if task.Metadata == nil {
		task.Metadata = make(map[string]interface{})
	}

	// Serialize planning link metadata to JSON string (for protobuf compatibility)
	linkJSON, err := json.Marshal(linkMeta)
	if err == nil {
		task.Metadata["planning_links"] = string(linkJSON)
	}

	// Also store individual fields for easy access
	if linkMeta.PlanningDoc != "" {
		task.Metadata["planning_doc"] = linkMeta.PlanningDoc
	}

	if linkMeta.EpicID != "" {
		task.Metadata["epic_id"] = linkMeta.EpicID
	} else {
		delete(task.Metadata, "epic_id")
	}

	if linkMeta.DocType != "" {
		task.Metadata["planning_doc_type"] = linkMeta.DocType
	}
}

// GetPlanningLinkMetadata retrieves planning document link metadata from task metadata.
func GetPlanningLinkMetadata(task *models.Todo2Task) *PlanningLinkMetadata {
	if task.Metadata == nil {
		return nil
	}

	// Try to get from JSON string first (preferred format)
	if linkJSON, ok := task.Metadata["planning_links"].(string); ok && linkJSON != "" {
		var linkMeta PlanningLinkMetadata
		if err := json.Unmarshal([]byte(linkJSON), &linkMeta); err == nil {
			return &linkMeta
		}
	}

	// Fallback: reconstruct from individual fields
	linkMeta := &PlanningLinkMetadata{}
	if doc, ok := task.Metadata["planning_doc"].(string); ok {
		linkMeta.PlanningDoc = doc
	}

	if epicID, ok := task.Metadata["epic_id"].(string); ok {
		linkMeta.EpicID = epicID
	}

	if docType, ok := task.Metadata["planning_doc_type"].(string); ok {
		linkMeta.DocType = docType
	}

	// Only return if we have at least one field
	if linkMeta.PlanningDoc != "" || linkMeta.EpicID != "" || linkMeta.DocType != "" {
		return linkMeta
	}

	return nil
}

// ValidatePlanningLink validates that a planning document link is valid.
// Rejects paths outside project root (relative or absolute).
func ValidatePlanningLink(projectRoot string, planningDocPath string) error {
	if planningDocPath == "" {
		return fmt.Errorf("planning document path is empty")
	}

	// Ensure path is within project root (rejects absolute paths outside root and traversal)
	fullPath, err := security.ValidatePath(planningDocPath, projectRoot)
	if err != nil {
		return err
	}

	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return fmt.Errorf("planning document not found: %s", planningDocPath)
	}

	// Check if it's a markdown file
	if ext := filepath.Ext(fullPath); ext != ".md" && ext != ".markdown" {
		return fmt.Errorf("planning document must be a markdown file: %s", planningDocPath)
	}

	return nil
}

// ValidateTaskReference validates that a task ID exists in the task list.
func ValidateTaskReference(taskID string, tasks []models.Todo2Task) error {
	if taskID == "" {
		return fmt.Errorf("task ID is empty")
	}

	// Validate task ID format (T- followed by digits)
	taskIDPattern := regexp.MustCompile(`^T-\d+$`)
	if !taskIDPattern.MatchString(taskID) {
		return fmt.Errorf("invalid task ID format: %s (expected T-123 or T-1234567890)", taskID)
	}

	// Check if task exists
	for _, task := range tasks {
		if task.ID == taskID {
			return nil
		}
	}

	return fmt.Errorf("task %s not found", taskID)
}

// PlanningDependencyHint is a dependency hint extracted from a planning document (e.g. "Depends on T-XXX" or order).
type PlanningDependencyHint struct {
	TaskID      string `json:"task_id"`
	DependsOnID string `json:"depends_on_id"`
	Reason      string `json:"reason"`
}

// ExtractDependencyHintsFromPlan reads a single plan file and returns dependency hints.
// Looks for: "Depends on: T-XXX", "dependencies: [T-XXX]", "After: T-XXX", and milestone order (earlier (T-ID) suggests dep for later).
func ExtractDependencyHintsFromPlan(planPath string) ([]PlanningDependencyHint, error) {
	data, err := os.ReadFile(planPath)
	if err != nil {
		return nil, err
	}

	content := string(data)

	var hints []PlanningDependencyHint

	// Explicit patterns (case-insensitive)
	depOnRe := regexp.MustCompile(`(?i)(?:depends?\s+on|after|requires?)\s*:\s*(T-\d+(?:-\d+)?|T-\d+)`)
	depListRe := regexp.MustCompile(`(?i)dependencies?\s*:\s*\[([^\]]+)\]`)
	taskIDRe := regexp.MustCompile(`T-\d+(?:-\d+)?|T-\d+`)

	lines := strings.Split(content, "\n")
	for i, line := range lines {
		// In same line: "Task X (T-123). Depends on: T-456" or "Depends on: T-456"
		if depOnRe.MatchString(line) {
			ids := taskIDRe.FindAllString(line, -1)
			if len(ids) >= 2 {
				// Last ID is the dependency; one before can be the task (if we have context)
				depID := ids[len(ids)-1]
				for j := 0; j < len(ids)-1; j++ {
					if ids[j] != depID {
						hints = append(hints, PlanningDependencyHint{
							TaskID:      ids[j],
							DependsOnID: depID,
							Reason:      fmt.Sprintf("plan line %d: depends on", i+1),
						})
					}
				}
			}

			if len(ids) == 1 {
				// "Depends on: T-456" without task context - try previous line for (T-XXX)
				if i > 0 {
					prevIDs := taskIDRe.FindAllString(lines[i-1], -1)
					if len(prevIDs) > 0 {
						hints = append(hints, PlanningDependencyHint{
							TaskID:      prevIDs[len(prevIDs)-1],
							DependsOnID: ids[0],
							Reason:      fmt.Sprintf("plan line %d: depends on (from previous line)", i+1),
						})
					}
				}
			}
		}

		if depListRe.MatchString(line) {
			subs := depListRe.FindStringSubmatch(line)
			if len(subs) >= 2 {
				depIDs := taskIDRe.FindAllString(subs[1], -1)

				if i > 0 {
					prevIDs := taskIDRe.FindAllString(lines[i-1], -1)
					if len(prevIDs) > 0 {
						taskID := prevIDs[len(prevIDs)-1]
						for _, depID := range depIDs {
							if depID != taskID {
								hints = append(hints, PlanningDependencyHint{
									TaskID:      taskID,
									DependsOnID: depID,
									Reason:      fmt.Sprintf("plan line %d: dependencies list", i+1),
								})
							}
						}
					}
				}
			}
		}
	}

	// Milestone order: only from checkbox lines (- [ ] **Name** (T-ID)) so order reflects sequence
	checkboxRe := regexp.MustCompile(`^\s*-\s*\[\s*[ xX]\s*\]\s*(?:\*\*[^*]+\*\*|.+?)\s*\(\s*([A-Za-z0-9_-]+)\s*\)`)

	var orderedIDs []string

	for _, line := range lines {
		if subs := checkboxRe.FindStringSubmatch(line); len(subs) >= 2 {
			orderedIDs = append(orderedIDs, strings.TrimSpace(subs[1]))
		}
	}

	for k := 1; k < len(orderedIDs); k++ {
		curr, prev := orderedIDs[k], orderedIDs[k-1]
		if curr != prev {
			hints = append(hints, PlanningDependencyHint{
				TaskID:      curr,
				DependsOnID: prev,
				Reason:      "plan milestone order (earlier checkbox in doc)",
			})
		}
	}

	return hints, nil
}

// ExtractDependencyHintsFromPlanDir scans .cursor/plans and docs for *.plan.md / *_PLAN*.md and returns merged hints.
func ExtractDependencyHintsFromPlanDir(projectRoot string) []PlanningDependencyHint {
	var all []PlanningDependencyHint

	seen := make(map[string]struct{})

	dirs := []string{
		filepath.Join(projectRoot, ".cursor", "plans"),
		filepath.Join(projectRoot, "docs"),
	}
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, e := range entries {
			if e.IsDir() {
				continue
			}

			name := e.Name()
			if !strings.HasSuffix(name, ".md") {
				continue
			}

			if !strings.Contains(strings.ToLower(name), "plan") {
				continue
			}

			full := filepath.Join(dir, name)

			hints, err := ExtractDependencyHintsFromPlan(full)
			if err != nil {
				continue
			}

			for _, h := range hints {
				key := h.TaskID + "|" + h.DependsOnID
				if _, ok := seen[key]; !ok {
					seen[key] = struct{}{}

					all = append(all, h)
				}
			}
		}
	}

	return all
}

// ExtractTaskIDsFromPlanningDoc extracts task IDs referenced in a planning document.
func ExtractTaskIDsFromPlanningDoc(content string) []string {
	taskIDs := []string{}

	// Pattern to match task/epic IDs: T-123 or T-1234567890
	taskRefPattern := regexp.MustCompile(`T-(\d+)`)
	matches := taskRefPattern.FindAllStringSubmatch(content, -1)

	seen := make(map[string]bool)

	for _, match := range matches {
		if len(match) > 0 {
			taskID := match[0] // Full match (T-123)
			if !seen[taskID] {
				taskIDs = append(taskIDs, taskID)
				seen[taskID] = true
			}
		}
	}

	return taskIDs
}

// UpdatePlanningDocWithTaskRefs updates a planning document with task references
// Adds task/epic references in a standardized format if they don't exist.
func UpdatePlanningDocWithTaskRefs(projectRoot string, docPath string, taskIDs []string, epicIDs []string) (bool, error) {
	fullPath := filepath.Join(projectRoot, docPath)

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return false, fmt.Errorf("failed to read planning doc: %w", err)
	}

	contentStr := string(content)
	modified := false

	// Check if task references section exists
	taskRefsSection := regexp.MustCompile(`(?i)##\s*Related\s+Tasks|##\s*Task\s+References`)
	hasTaskRefsSection := taskRefsSection.MatchString(contentStr)

	// Build task references text
	var refsText strings.Builder
	if len(taskIDs) > 0 || len(epicIDs) > 0 {
		refsText.WriteString("\n## Related Tasks\n\n")

		if len(epicIDs) > 0 {
			refsText.WriteString("### Epics\n\n")

			for _, epicID := range epicIDs {
				refsText.WriteString(fmt.Sprintf("- **Epic ID**: `%s`\n", epicID))
			}

			refsText.WriteString("\n")
		}

		if len(taskIDs) > 0 {
			refsText.WriteString("### Tasks\n\n")

			for _, taskID := range taskIDs {
				refsText.WriteString(fmt.Sprintf("- **Task ID**: `%s`\n", taskID))
			}

			refsText.WriteString("\n")
		}
	}

	// If section doesn't exist and we have references, add it
	if !hasTaskRefsSection && refsText.Len() > 0 {
		// Try to insert before "## Next Steps" or at the end
		nextStepsPattern := regexp.MustCompile(`(?i)\n##\s*Next\s+Steps`)
		if nextStepsPattern.MatchString(contentStr) {
			contentStr = nextStepsPattern.ReplaceAllString(contentStr, refsText.String()+"$0")
			modified = true
		} else {
			// Append at the end
			contentStr += refsText.String()
			modified = true
		}
	}

	if modified {
		if err := os.WriteFile(fullPath, []byte(contentStr), 0644); err != nil {
			return false, fmt.Errorf("failed to write planning doc: %w", err)
		}
	}

	return modified, nil
}

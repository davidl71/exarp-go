package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/davidl71/exarp-go/internal/models"
)

// PlanningLinkMetadata represents planning document link metadata stored in task metadata
type PlanningLinkMetadata struct {
	PlanningDoc   string   `json:"planning_doc,omitempty"`   // Path to planning document
	EpicID        string   `json:"epic_id,omitempty"`        // Epic task ID if part of an epic
	RelatedDocs   []string `json:"related_docs,omitempty"`   // Related planning document paths
	TaskRefs      []string `json:"task_refs,omitempty"`      // Task IDs referenced in planning doc
	EpicRefs      []string `json:"epic_refs,omitempty"`      // Epic IDs referenced in planning doc
	DocType       string   `json:"doc_type,omitempty"`       // Planning document type
	LastSynced    string   `json:"last_synced,omitempty"`    // Last sync timestamp
}

// SetPlanningLinkMetadata stores planning document link metadata in task metadata
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
	}
	if linkMeta.DocType != "" {
		task.Metadata["planning_doc_type"] = linkMeta.DocType
	}
}

// GetPlanningLinkMetadata retrieves planning document link metadata from task metadata
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

// ValidatePlanningLink validates that a planning document link is valid
func ValidatePlanningLink(projectRoot string, planningDocPath string) error {
	if planningDocPath == "" {
		return fmt.Errorf("planning document path is empty")
	}
	
	// Resolve path relative to project root
	fullPath := planningDocPath
	if !filepath.IsAbs(planningDocPath) {
		fullPath = filepath.Join(projectRoot, planningDocPath)
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

// ValidateTaskReference validates that a task ID exists in the task list
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

// ExtractTaskIDsFromPlanningDoc extracts task IDs referenced in a planning document
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
// Adds task/epic references in a standardized format if they don't exist
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

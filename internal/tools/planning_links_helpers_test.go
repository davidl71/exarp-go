package tools

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/davidl71/exarp-go/internal/models"
)

func TestSetPlanningLinkMetadata(t *testing.T) {
	task := &models.Todo2Task{
		ID:       "T-1234567890",
		Content:  "Test Task",
		Metadata: make(map[string]interface{}),
	}

	linkMeta := &PlanningLinkMetadata{
		PlanningDoc: "docs/README.md",
		EpicID:      "T-9876543210",
		DocType:     "planning",
	}

	SetPlanningLinkMetadata(task, linkMeta)

	// Verify metadata was stored
	if task.Metadata == nil {
		t.Fatal("Metadata should not be nil")
	}

	// Check individual fields
	if doc, ok := task.Metadata["planning_doc"].(string); !ok || doc != "docs/README.md" {
		t.Errorf("Expected planning_doc to be 'docs/README.md', got %v", task.Metadata["planning_doc"])
	}

	if epicID, ok := task.Metadata["epic_id"].(string); !ok || epicID != "T-9876543210" {
		t.Errorf("Expected epic_id to be 'T-9876543210', got %v", task.Metadata["epic_id"])
	}

	// Check JSON string exists
	if linkJSON, ok := task.Metadata["planning_links"].(string); !ok || linkJSON == "" {
		t.Error("Expected planning_links JSON string to be stored")
	}
}

func TestGetPlanningLinkMetadata(t *testing.T) {
	task := &models.Todo2Task{
		ID:       "T-1234567890",
		Content:  "Test Task",
		Metadata: make(map[string]interface{}),
	}

	// Set metadata first
	linkMeta := &PlanningLinkMetadata{
		PlanningDoc: "docs/README.md",
		EpicID:      "T-9876543210",
		DocType:     "planning",
	}
	SetPlanningLinkMetadata(task, linkMeta)

	// Retrieve metadata
	retrieved := GetPlanningLinkMetadata(task)
	if retrieved == nil {
		t.Fatal("Expected to retrieve planning link metadata, got nil")
	}

	if retrieved.PlanningDoc != "docs/README.md" {
		t.Errorf("Expected PlanningDoc to be 'docs/README.md', got %s", retrieved.PlanningDoc)
	}

	if retrieved.EpicID != "T-9876543210" {
		t.Errorf("Expected EpicID to be 'T-9876543210', got %s", retrieved.EpicID)
	}

	if retrieved.DocType != "planning" {
		t.Errorf("Expected DocType to be 'planning', got %s", retrieved.DocType)
	}
}

func TestGetPlanningLinkMetadataFromIndividualFields(t *testing.T) {
	// Test fallback to individual fields
	task := &models.Todo2Task{
		ID:      "T-1234567890",
		Content: "Test Task",
		Metadata: map[string]interface{}{
			"planning_doc": "docs/test.md",
			"epic_id":      "T-999",
		},
	}

	retrieved := GetPlanningLinkMetadata(task)
	if retrieved == nil {
		t.Fatal("Expected to retrieve planning link metadata, got nil")
	}

	if retrieved.PlanningDoc != "docs/test.md" {
		t.Errorf("Expected PlanningDoc to be 'docs/test.md', got %s", retrieved.PlanningDoc)
	}

	if retrieved.EpicID != "T-999" {
		t.Errorf("Expected EpicID to be 'T-999', got %s", retrieved.EpicID)
	}
}

func TestValidatePlanningLink(t *testing.T) {
	// Get project root (assuming we're in the project directory)
	projectRoot, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	// Test valid path (README.md should exist)
	validPath := "README.md"
	fullPath := filepath.Join(projectRoot, validPath)
	if _, err := os.Stat(fullPath); err == nil {
		err := ValidatePlanningLink(projectRoot, validPath)
		if err != nil {
			t.Errorf("Expected valid path to pass validation, got error: %v", err)
		}
	}

	// Test invalid path (non-existent file)
	invalidPath := "docs/NONEXISTENT_FILE_12345.md"
	err = ValidatePlanningLink(projectRoot, invalidPath)
	if err == nil {
		t.Error("Expected invalid path to fail validation, but got no error")
	}

	// Test empty path
	err = ValidatePlanningLink(projectRoot, "")
	if err == nil {
		t.Error("Expected empty path to fail validation, but got no error")
	}
}

func TestValidateTaskReference(t *testing.T) {
	testTasks := []models.Todo2Task{
		{ID: "T-1234567890", Content: "Test Task 1"},
		{ID: "T-9876543210", Content: "Test Task 2"},
		{ID: "T-1111111111", Content: "Test Task 3"},
	}

	// Test valid task ID
	validTaskID := "T-1234567890"
	err := ValidateTaskReference(validTaskID, testTasks)
	if err != nil {
		t.Errorf("Expected valid task ID to pass validation, got error: %v", err)
	}

	// Test invalid task ID (doesn't exist)
	invalidTaskID := "T-9999999999"
	err = ValidateTaskReference(invalidTaskID, testTasks)
	if err == nil {
		t.Error("Expected invalid task ID to fail validation, but got no error")
	}

	// Test invalid format (missing T- prefix)
	invalidFormat := "1234567890"
	err = ValidateTaskReference(invalidFormat, testTasks)
	if err == nil {
		t.Error("Expected invalid format to fail validation, but got no error")
	}

	// Test empty task ID
	err = ValidateTaskReference("", testTasks)
	if err == nil {
		t.Error("Expected empty task ID to fail validation, but got no error")
	}
}

func TestExtractTaskIDsFromPlanningDoc(t *testing.T) {
	docContent := `# Planning Document

This document references several tasks:
- Task ID: T-1234567890
- Epic ID: T-9876543210
- Another task: T-1111111111
- Duplicate: T-1234567890 (should be deduplicated)

## Summary
Tasks T-1234567890 and T-9876543210 are related.
`

	taskIDs := ExtractTaskIDsFromPlanningDoc(docContent)

	expectedCount := 3 // Should be 3 unique task IDs
	if len(taskIDs) != expectedCount {
		t.Errorf("Expected %d task IDs, got %d: %v", expectedCount, len(taskIDs), taskIDs)
	}

	// Check that all expected IDs are present
	expectedIDs := map[string]bool{
		"T-1234567890": true,
		"T-9876543210": true,
		"T-1111111111": true,
	}

	for _, taskID := range taskIDs {
		if !expectedIDs[taskID] {
			t.Errorf("Unexpected task ID found: %s", taskID)
		}
	}

	// Test empty content
	emptyTaskIDs := ExtractTaskIDsFromPlanningDoc("")
	if len(emptyTaskIDs) != 0 {
		t.Errorf("Expected empty content to return no task IDs, got %d", len(emptyTaskIDs))
	}
}

func TestPlanningLinkMetadataJSON(t *testing.T) {
	// Test JSON serialization/deserialization
	linkMeta := &PlanningLinkMetadata{
		PlanningDoc: "docs/README.md",
		EpicID:      "T-9876543210",
		DocType:     "planning",
		TaskRefs:    []string{"T-123", "T-456"},
		EpicRefs:    []string{"T-789"},
	}

	// Serialize
	jsonData, err := json.Marshal(linkMeta)
	if err != nil {
		t.Fatalf("Failed to marshal PlanningLinkMetadata: %v", err)
	}

	// Deserialize
	var deserialized PlanningLinkMetadata
	err = json.Unmarshal(jsonData, &deserialized)
	if err != nil {
		t.Fatalf("Failed to unmarshal PlanningLinkMetadata: %v", err)
	}

	// Verify fields
	if deserialized.PlanningDoc != linkMeta.PlanningDoc {
		t.Errorf("PlanningDoc mismatch: expected %s, got %s", linkMeta.PlanningDoc, deserialized.PlanningDoc)
	}

	if deserialized.EpicID != linkMeta.EpicID {
		t.Errorf("EpicID mismatch: expected %s, got %s", linkMeta.EpicID, deserialized.EpicID)
	}

	if len(deserialized.TaskRefs) != len(linkMeta.TaskRefs) {
		t.Errorf("TaskRefs length mismatch: expected %d, got %d", len(linkMeta.TaskRefs), len(deserialized.TaskRefs))
	}
}

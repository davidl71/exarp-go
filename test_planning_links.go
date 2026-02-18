package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/davidl71/exarp-go/internal/models"
	"github.com/davidl71/exarp-go/internal/tools"
)

func main() {
	projectRoot, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Testing Planning Links Helpers")
	fmt.Println("==============================")
	fmt.Println()

	// Test 1: Create a task and set planning link metadata
	fmt.Println("Test 1: Set Planning Link Metadata")
	fmt.Println("-----------------------------------")

	task := &models.Todo2Task{
		ID:       "T-1234567890",
		Content:  "Test Task",
		Metadata: make(map[string]interface{}),
	}

	linkMeta := &tools.PlanningLinkMetadata{
		PlanningDoc: "docs/README.md",
		EpicID:      "T-9876543210",
		DocType:     "planning",
	}

	tools.SetPlanningLinkMetadata(task, linkMeta)
	fmt.Printf("✓ Set planning link metadata\n")
	fmt.Printf("  PlanningDoc: %s\n", linkMeta.PlanningDoc)
	fmt.Printf("  EpicID: %s\n", linkMeta.EpicID)
	fmt.Println()

	// Test 2: Retrieve planning link metadata
	fmt.Println("Test 2: Get Planning Link Metadata")
	fmt.Println("-----------------------------------")

	retrieved := tools.GetPlanningLinkMetadata(task)
	if retrieved != nil {
		fmt.Printf("✓ Retrieved planning link metadata\n")
		fmt.Printf("  PlanningDoc: %s\n", retrieved.PlanningDoc)
		fmt.Printf("  EpicID: %s\n", retrieved.EpicID)
		fmt.Printf("  DocType: %s\n", retrieved.DocType)
	} else {
		fmt.Printf("✗ Failed to retrieve planning link metadata\n")
	}

	fmt.Println()

	// Test 3: Validate planning link (valid path)
	fmt.Println("Test 3: Validate Planning Link (Valid Path)")
	fmt.Println("-------------------------------------------")

	validPath := "docs/README.md"
	fullPath := filepath.Join(projectRoot, validPath)

	if _, err := os.Stat(fullPath); err == nil {
		err := tools.ValidatePlanningLink(projectRoot, validPath)
		if err == nil {
			fmt.Printf("✓ Valid planning link: %s\n", validPath)
		} else {
			fmt.Printf("✗ Validation failed: %v\n", err)
		}
	} else {
		fmt.Printf("⚠ Test file not found: %s (skipping validation test)\n", fullPath)
	}

	fmt.Println()

	// Test 4: Validate planning link (invalid path)
	fmt.Println("Test 4: Validate Planning Link (Invalid Path)")
	fmt.Println("---------------------------------------------")

	invalidPath := "docs/NONEXISTENT.md"

	err = tools.ValidatePlanningLink(projectRoot, invalidPath)
	if err != nil {
		fmt.Printf("✓ Correctly rejected invalid path: %s\n", invalidPath)
		fmt.Printf("  Error: %v\n", err)
	} else {
		fmt.Printf("✗ Validation should have failed for: %s\n", invalidPath)
	}

	fmt.Println()

	// Test 5: Validate task reference format
	fmt.Println("Test 5: Validate Task Reference Format")
	fmt.Println("--------------------------------------")

	testTasks := []models.Todo2Task{
		{ID: "T-1234567890", Content: "Test Task 1"},
		{ID: "T-9876543210", Content: "Test Task 2"},
	}

	validTaskID := "T-1234567890"

	err = tools.ValidateTaskReference(validTaskID, testTasks)
	if err == nil {
		fmt.Printf("✓ Valid task reference: %s\n", validTaskID)
	} else {
		fmt.Printf("✗ Validation failed: %v\n", err)
	}

	invalidTaskID := "T-9999999999"

	err = tools.ValidateTaskReference(invalidTaskID, testTasks)
	if err != nil {
		fmt.Printf("✓ Correctly rejected invalid task ID: %s\n", invalidTaskID)
		fmt.Printf("  Error: %v\n", err)
	} else {
		fmt.Printf("✗ Validation should have failed for: %s\n", invalidTaskID)
	}

	fmt.Println()

	// Test 6: Extract task IDs from planning doc content
	fmt.Println("Test 6: Extract Task IDs from Planning Doc")
	fmt.Println("------------------------------------------")

	docContent := `# Planning Document

This document references several tasks:
- Task ID: T-1234567890
- Epic ID: T-9876543210
- Another task: T-1111111111

## Summary
Tasks T-1234567890 and T-9876543210 are related.
`

	taskIDs := tools.ExtractTaskIDsFromPlanningDoc(docContent)
	fmt.Printf("✓ Extracted %d task IDs from planning doc\n", len(taskIDs))
	fmt.Printf("  Task IDs: %v\n", taskIDs)
	fmt.Println()

	// Test 7: Display metadata structure
	fmt.Println("Test 7: Metadata Structure")
	fmt.Println("--------------------------")

	if task.Metadata != nil {
		metadataJSON, _ := json.MarshalIndent(task.Metadata, "  ", "  ")
		fmt.Printf("Task Metadata:\n%s\n", metadataJSON)
	}

	fmt.Println()

	fmt.Println("==============================")
	fmt.Println("All tests completed!")
}

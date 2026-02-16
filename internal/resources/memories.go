package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/davidl71/exarp-go/internal/tools"
)

// handleMemories handles the stdio://memories resource
// Returns all memories with statistics.
func handleMemories(ctx context.Context, uri string) ([]byte, string, error) {
	projectRoot, err := tools.FindProjectRoot()
	if err != nil {
		return nil, "", fmt.Errorf("failed to find project root: %w", err)
	}

	// Load all memories using native Go implementation
	memories, err := tools.LoadAllMemories(projectRoot)
	if err != nil {
		return nil, "", fmt.Errorf("failed to load memories: %w", err)
	}

	limit := 50
	if len(memories) > limit {
		memories = memories[:limit]
	}

	// Calculate statistics
	categories := make(map[string]int)
	allMemories, _ := tools.LoadAllMemories(projectRoot)

	for _, m := range allMemories {
		categories[m.Category]++
	}

	result := map[string]interface{}{
		"memories":             formatMemoriesForResource(memories),
		"total":                len(allMemories),
		"returned":             len(memories),
		"categories":           categories,
		"available_categories": tools.MemoryCategories(),
		"timestamp":            time.Now().Format(time.RFC3339),
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal memories: %w", err)
	}

	return jsonData, "application/json", nil
}

// handleMemoriesByCategory handles the stdio://memories/category/{category} resource.
func handleMemoriesByCategory(ctx context.Context, uri string) ([]byte, string, error) {
	// Parse category from URI: stdio://memories/category/{category}
	category, err := parseURIVariableByIndexWithValidation(uri, 3, "category", "stdio://memories/category/{category}")
	if err != nil {
		return nil, "", err
	}

	projectRoot, err := tools.FindProjectRoot()
	if err != nil {
		return nil, "", fmt.Errorf("failed to find project root: %w", err)
	}

	// Load and filter memories
	memories, err := tools.LoadAllMemories(projectRoot)
	if err != nil {
		return nil, "", fmt.Errorf("failed to load memories: %w", err)
	}

	// Filter by category
	filtered := []tools.Memory{}

	for _, m := range memories {
		if m.Category == category {
			filtered = append(filtered, m)
		}
	}

	limit := 50
	if len(filtered) > limit {
		filtered = filtered[:limit]
	}

	result := map[string]interface{}{
		"category":  category,
		"memories":  formatMemoriesForResource(filtered),
		"total":     len(filtered),
		"timestamp": time.Now().Format(time.RFC3339),
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal memories: %w", err)
	}

	return jsonData, "application/json", nil
}

// handleMemoriesByTask handles the stdio://memories/task/{task_id} resource.
func handleMemoriesByTask(ctx context.Context, uri string) ([]byte, string, error) {
	// Parse task_id from URI: stdio://memories/task/{task_id}
	taskID, err := parseURIVariableByIndexWithValidation(uri, 3, "task ID", "stdio://memories/task/{task_id}")
	if err != nil {
		return nil, "", err
	}

	projectRoot, err := tools.FindProjectRoot()
	if err != nil {
		return nil, "", fmt.Errorf("failed to find project root: %w", err)
	}

	// Load and filter memories
	memories, err := tools.LoadAllMemories(projectRoot)
	if err != nil {
		return nil, "", fmt.Errorf("failed to load memories: %w", err)
	}

	// Filter by task_id
	filtered := []tools.Memory{}

	for _, m := range memories {
		for _, linkedTask := range m.LinkedTasks {
			if linkedTask == taskID {
				filtered = append(filtered, m)
				break
			}
		}
	}

	result := map[string]interface{}{
		"task_id":   taskID,
		"memories":  formatMemoriesForResource(filtered),
		"total":     len(filtered),
		"timestamp": time.Now().Format(time.RFC3339),
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal memories: %w", err)
	}

	return jsonData, "application/json", nil
}

// handleRecentMemories handles the stdio://memories/recent resource.
func handleRecentMemories(ctx context.Context, uri string) ([]byte, string, error) {
	projectRoot, err := tools.FindProjectRoot()
	if err != nil {
		return nil, "", fmt.Errorf("failed to find project root: %w", err)
	}

	// Load all memories
	memories, err := tools.LoadAllMemories(projectRoot)
	if err != nil {
		return nil, "", fmt.Errorf("failed to load memories: %w", err)
	}

	// Filter to last 24 hours
	cutoff := time.Now().Add(-24 * time.Hour)
	recent := []tools.Memory{}

	for _, m := range memories {
		createdAt, err := time.Parse(time.RFC3339, m.CreatedAt)
		if err != nil {
			// Try alternative format
			createdAt, err = time.Parse("2006-01-02T15:04:05.999999", m.CreatedAt)
			if err != nil {
				continue
			}
		}

		if createdAt.After(cutoff) {
			recent = append(recent, m)
		}
	}

	result := map[string]interface{}{
		"hours":     24,
		"memories":  formatMemoriesForResource(recent),
		"total":     len(recent),
		"timestamp": time.Now().Format(time.RFC3339),
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal memories: %w", err)
	}

	return jsonData, "application/json", nil
}

// handleSessionMemories handles the stdio://memories/session/{date} resource.
func handleSessionMemories(ctx context.Context, uri string) ([]byte, string, error) {
	// Parse date from URI: stdio://memories/session/{date}
	date, err := parseURIVariableByIndexWithValidation(uri, 3, "date", "stdio://memories/session/{date}")
	if err != nil {
		return nil, "", err
	}

	// Validate date format (YYYY-MM-DD)
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return nil, "", fmt.Errorf("invalid date format: %s (expected YYYY-MM-DD)", date)
	}

	projectRoot, err := tools.FindProjectRoot()
	if err != nil {
		return nil, "", fmt.Errorf("failed to find project root: %w", err)
	}

	// Load all memories
	memories, err := tools.LoadAllMemories(projectRoot)
	if err != nil {
		return nil, "", fmt.Errorf("failed to load memories: %w", err)
	}

	// Filter by session date
	filtered := []tools.Memory{}

	for _, m := range memories {
		if m.SessionDate == date {
			filtered = append(filtered, m)
		}
	}

	result := map[string]interface{}{
		"session_date": date,
		"memories":     formatMemoriesForResource(filtered),
		"total":        len(filtered),
		"timestamp":    time.Now().Format(time.RFC3339),
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal memories: %w", err)
	}

	return jsonData, "application/json", nil
}

// Helper functions - all use exported functions from tools package

// formatMemoriesForResource formats memories for resource output.
func formatMemoriesForResource(memories []tools.Memory) []map[string]interface{} {
	result := make([]map[string]interface{}, len(memories))
	for i, m := range memories {
		result[i] = map[string]interface{}{
			"id":           m.ID,
			"title":        m.Title,
			"content":      m.Content,
			"category":     m.Category,
			"linked_tasks": m.LinkedTasks,
			"metadata":     m.Metadata,
			"created_at":   m.CreatedAt,
			"session_date": m.SessionDate,
		}
	}

	return result
}

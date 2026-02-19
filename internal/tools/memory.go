// memory.go — MCP "memory" tool: store, recall, search, list, forget project memories.
package tools

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/davidl71/exarp-go/internal/config"
	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/proto"
	"github.com/davidl71/mcp-go-core/pkg/mcp/response"
)

// Memory represents a stored memory.
type Memory struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Content     string                 `json:"content"`
	Category    string                 `json:"category"`
	LinkedTasks []string               `json:"linked_tasks,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   string                 `json:"created_at"`
	SessionDate string                 `json:"session_date"`
}

// MemoryCategories returns the valid memory categories (from config when available).
// Exported for use by resources and other packages.
func MemoryCategories() []string {
	return config.MemoryCategories()
}

func memoryCategories() []string {
	return config.MemoryCategories()
}

// handleMemoryNative handles the memory tool with native Go CRUD operations.
// All success responses use proto.MemoryResponse and are formatted via MemoryResponseToMap;
// do not return ad-hoc maps so that responses stay aligned to the MemoryResponse proto.
func handleMemoryNative(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	// Try protobuf first, fall back to JSON for backward compatibility
	req, params, err := ParseMemoryRequest(args)
	if err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Convert protobuf request to params map if needed (for compatibility with existing functions)
	if req != nil {
		params = MemoryRequestToParams(req)
		// Set defaults for protobuf request
		if req.Action == "" {
			params["action"] = "search"
		}

		if req.Category == "" {
			params["category"] = "insight"
		}

		if req.Limit == 0 {
			params["limit"] = 10
		}

		if !req.IncludeRelated {
			params["include_related"] = true // Default is true
		}
	}

	action, _ := params["action"].(string)
	if action == "" {
		action = "search"
	}

	switch action {
	case "save":
		return handleMemorySave(ctx, params)
	case "recall":
		return handleMemoryRecall(ctx, params)
	case "search":
		// Try basic text search in Go first, fall back to Python bridge for semantic search
		return handleMemorySearch(ctx, params)
	case "list":
		return handleMemoryList(ctx, params)
	default:
		// Unknown action, fall back to Python bridge
		return nil, fmt.Errorf("unknown action: %s (use 'save', 'recall', 'search', or 'list')", action)
	}
}

// handleMemorySave handles save action.
func handleMemorySave(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	title, _ := params["title"].(string)
	content, _ := params["content"].(string)

	if title == "" || content == "" {
		return nil, fmt.Errorf("title and content are required for save action")
	}

	category := "insight"
	if cat, ok := params["category"].(string); ok && cat != "" {
		category = cat
	}

	// Validate category
	validCategory := false

	for _, c := range memoryCategories() {
		if category == c {
			validCategory = true
			break
		}
	}

	if !validCategory {
		return nil, fmt.Errorf("invalid category '%s'. Must be one of: %s", category, strings.Join(memoryCategories(), ", "))
	}

	// Truncate title if too long
	if len(title) > 100 {
		title = title[:97] + "..."
	}

	var taskID string
	if tid, ok := params["task_id"].(string); ok {
		taskID = tid
	}

	var metadata map[string]interface{}
	if metaStr, ok := params["metadata"].(string); ok && metaStr != "" {
		if err := json.Unmarshal([]byte(metaStr), &metadata); err != nil {
			return nil, fmt.Errorf("invalid metadata JSON: %w", err)
		}
	}

	// Generate UUID v4 format
	id, err := generateUUID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate memory ID: %w", err)
	}

	// Create memory
	memory := Memory{
		ID:          id,
		Title:       title,
		Content:     content,
		Category:    category,
		LinkedTasks: []string{},
		Metadata:    metadata,
		CreatedAt:   time.Now().Format(time.RFC3339),
		SessionDate: time.Now().Format("2006-01-02"),
	}

	if taskID != "" {
		memory.LinkedTasks = []string{taskID}
	}

	// Save to file
	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	if err := saveMemory(projectRoot, memory); err != nil {
		return nil, fmt.Errorf("failed to save memory: %w", err)
	}

	pbMem, _ := MemoryToProto(&memory)
	resp := &proto.MemoryResponse{
		Success:  true,
		Method:   "native_go",
		MemoryId: memory.ID,
		Message:  fmt.Sprintf("✅ Memory saved: %s", title),
	}

	if pbMem != nil {
		resp.Memories = []*proto.Memory{pbMem}
	}

	return response.FormatResult(MemoryResponseToMap(resp), "")
}

// handleMemoryRecall handles recall action.
func handleMemoryRecall(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	taskID, _ := params["task_id"].(string)
	if taskID == "" {
		return nil, fmt.Errorf("task_id is required for recall action")
	}

	includeRelated := true
	if ir, ok := params["include_related"].(bool); ok {
		includeRelated = ir
	}

	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	memories, err := LoadAllMemories(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load memories: %w", err)
	}

	// Filter by task_id
	related := []Memory{}

	for _, m := range memories {
		for _, linkedTask := range m.LinkedTasks {
			if linkedTask == taskID {
				related = append(related, m)
				break
			}
		}
	}

	// If include_related, find memories from related tasks (dependencies)
	if includeRelated {
		// Get task dependencies from database
		dependencies, err := database.GetDependencies(taskID)
		if err == nil {
			for _, depID := range dependencies {
				// Find memories linked to dependency tasks
				for _, m := range memories {
					for _, linkedTask := range m.LinkedTasks {
						if linkedTask == depID {
							// Check if already in related list
							found := false

							for _, existing := range related {
								if existing.ID == m.ID {
									found = true
									break
								}
							}

							if !found {
								related = append(related, m)
							}

							break
						}
					}
				}
			}
		}
	}

	pbMemories := make([]*proto.Memory, 0, len(related))

	for i := range related {
		pb, err := MemoryToProto(&related[i])
		if err == nil && pb != nil {
			pbMemories = append(pbMemories, pb)
		}
	}

	resp := &proto.MemoryResponse{
		Success:        true,
		Method:         "native_go",
		TaskId:         taskID,
		Memories:       pbMemories,
		Count:          int32(len(related)),
		IncludeRelated: includeRelated,
	}

	return response.FormatResult(MemoryResponseToMap(resp), "")
}

// handleMemorySearch handles search action (basic text search in Go).
func handleMemorySearch(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	query, _ := params["query"].(string)
	if query == "" {
		return nil, fmt.Errorf("query is required for search action")
	}

	limit := 10
	if l, ok := params["limit"].(float64); ok {
		limit = int(l)
	}

	var category string
	if cat, ok := params["category"].(string); ok && cat != "" && cat != "insight" {
		category = cat
	}

	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	memories, err := LoadAllMemories(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load memories: %w", err)
	}

	// Basic text search (semantic search would use Python bridge)
	queryLower := strings.ToLower(query)
	scored := []struct {
		score  int
		memory Memory
	}{}

	for _, m := range memories {
		// Filter by category if specified
		if category != "" && m.Category != category {
			continue
		}

		score := 0
		titleLower := strings.ToLower(m.Title)
		contentLower := strings.ToLower(m.Content)
		categoryLower := strings.ToLower(m.Category)

		// Title match scores highest
		if strings.Contains(titleLower, queryLower) {
			score += 10
		}

		// Content match
		if strings.Contains(contentLower, queryLower) {
			score += 5
			// Bonus for multiple occurrences
			score += strings.Count(contentLower, queryLower)
		}

		// Category match
		if strings.Contains(categoryLower, queryLower) {
			score += 3
		}

		if score > 0 {
			scored = append(scored, struct {
				score  int
				memory Memory
			}{score: score, memory: m})
		}
	}

	// Sort by score descending
	for i := 0; i < len(scored)-1; i++ {
		for j := i + 1; j < len(scored); j++ {
			if scored[i].score < scored[j].score {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}

	// Take top results
	results := []Memory{}

	for i, s := range scored {
		if i >= limit {
			break
		}

		results = append(results, s.memory)
	}

	pbMemories := make([]*proto.Memory, 0, len(results))

	for i := range results {
		pb, err := MemoryToProto(&results[i])
		if err == nil && pb != nil {
			pbMemories = append(pbMemories, pb)
		}
	}

	resp := &proto.MemoryResponse{
		Success:    true,
		Method:     "native_go",
		Query:      query,
		Memories:   pbMemories,
		Count:      int32(len(results)),
		TotalFound: int32(len(scored)),
	}

	return response.FormatResult(MemoryResponseToMap(resp), "")
}

// handleMemoryList handles list action.
func handleMemoryList(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	var category string
	if cat, ok := params["category"].(string); ok && cat != "" {
		category = cat
	}

	limit := 50
	if l, ok := params["limit"].(float64); ok {
		limit = int(l)
	}

	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	memories, err := LoadAllMemories(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load memories: %w", err)
	}

	// Filter by category if specified
	if category != "" {
		filtered := []Memory{}

		for _, m := range memories {
			if m.Category == category {
				filtered = append(filtered, m)
			}
		}

		memories = filtered
	}

	// Limit results
	if len(memories) > limit {
		memories = memories[:limit]
	}

	// Calculate statistics
	categories := make(map[string]int)
	allMemories, _ := LoadAllMemories(projectRoot)

	for _, m := range allMemories {
		categories[m.Category]++
	}

	pbMemories := make([]*proto.Memory, 0, len(memories))

	for i := range memories {
		pb, err := MemoryToProto(&memories[i])
		if err == nil && pb != nil {
			pbMemories = append(pbMemories, pb)
		}
	}

	catProto := make(map[string]int32)
	for k, v := range categories {
		catProto[k] = int32(v)
	}

	resp := &proto.MemoryResponse{
		Success:             true,
		Method:              "native_go",
		Memories:            pbMemories,
		Total:               int32(len(allMemories)),
		Returned:            int32(len(memories)),
		Categories:          catProto,
		AvailableCategories: memoryCategories(),
	}

	return response.FormatResult(MemoryResponseToMap(resp), "")
}

// Helper functions

func getMemoriesDir(projectRoot string) (string, error) {
	storagePath := config.MemoryStoragePath()

	memoriesDir := filepath.Join(projectRoot, filepath.FromSlash(storagePath))
	if err := os.MkdirAll(memoriesDir, 0755); err != nil {
		return "", err
	}

	return memoriesDir, nil
}

// deleteMemoryFile deletes a memory file, trying both .pb and .json formats
// Returns true if a file was deleted, false otherwise.
func deleteMemoryFile(projectRoot, memoryID string) bool {
	memoriesDir, err := getMemoriesDir(projectRoot)
	if err != nil {
		return false
	}

	// Try protobuf format first (.pb)
	pbPath := filepath.Join(memoriesDir, memoryID+".pb")
	if err := os.Remove(pbPath); err == nil {
		return true
	}

	// Fall back to JSON format (backward compatibility)
	jsonPath := filepath.Join(memoriesDir, memoryID+".json")
	if err := os.Remove(jsonPath); err == nil {
		return true
	}

	return false
}

// LoadAllMemories loads all memories from the project root
// Supports both protobuf (.pb) and JSON (.json) formats for backward compatibility
// Exported for use by resource handlers.
func LoadAllMemories(projectRoot string) ([]Memory, error) {
	memoriesDir, err := getMemoriesDir(projectRoot)
	if err != nil {
		return nil, err
	}

	memories := []Memory{}

	entries, err := os.ReadDir(memoriesDir)
	if err != nil {
		return memories, nil // Return empty if directory doesn't exist yet
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		memoryPath := filepath.Join(memoriesDir, entry.Name())

		var memory Memory

		var shouldMigrate bool

		// Try protobuf format first (.pb)
		if strings.HasSuffix(entry.Name(), ".pb") {
			data, err := os.ReadFile(memoryPath)
			if err != nil {
				continue // Skip corrupted files
			}

			loadedMemory, err := DeserializeMemoryFromProtobuf(data)
			if err != nil {
				continue // Skip invalid protobuf
			}

			memory = *loadedMemory
		} else if strings.HasSuffix(entry.Name(), ".json") {
			// Fall back to JSON format (backward compatibility)
			data, err := os.ReadFile(memoryPath)
			if err != nil {
				continue // Skip corrupted files
			}

			if err := json.Unmarshal(data, &memory); err != nil {
				continue // Skip invalid JSON
			}

			// Mark for migration to protobuf
			shouldMigrate = true
		} else {
			// Skip files that don't match expected formats
			continue
		}

		memories = append(memories, memory)

		// Migrate JSON to protobuf format (async, non-blocking)
		if shouldMigrate {
			// Convert and save as protobuf
			if err := saveMemory(projectRoot, memory); err == nil {
				// Remove old JSON file after successful protobuf save
				_ = os.Remove(memoryPath)
			}
		}
	}

	// Sort by created_at descending (newest first)
	for i := 0; i < len(memories)-1; i++ {
		for j := i + 1; j < len(memories); j++ {
			if memories[i].CreatedAt < memories[j].CreatedAt {
				memories[i], memories[j] = memories[j], memories[i]
			}
		}
	}

	return memories, nil
}

// saveMemory saves a memory to file using protobuf binary format
// Also removes any old JSON file with the same ID for cleanup.
func saveMemory(projectRoot string, memory Memory) error {
	memoriesDir, err := getMemoriesDir(projectRoot)
	if err != nil {
		return err
	}

	// Save as protobuf binary (.pb)
	memoryPath := filepath.Join(memoriesDir, memory.ID+".pb")

	data, err := SerializeMemoryToProtobuf(&memory)
	if err != nil {
		return fmt.Errorf("failed to serialize memory to protobuf: %w", err)
	}

	if err := os.WriteFile(memoryPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write memory file: %w", err)
	}

	// Remove old JSON file if it exists (cleanup during migration)
	oldJSONPath := filepath.Join(memoriesDir, memory.ID+".json")
	if _, err := os.Stat(oldJSONPath); err == nil {
		_ = os.Remove(oldJSONPath) // Best effort cleanup
	}

	return nil
}

func formatMemories(memories []Memory) []map[string]interface{} {
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

// generateUUID generates a UUID v4 format string.
func generateUUID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	// Set version (4) and variant bits
	b[6] = (b[6] & 0x0f) | 0x40 // Version 4
	b[8] = (b[8] & 0x3f) | 0x80 // Variant 10

	// Format as UUID: xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx
	return fmt.Sprintf("%s-%s-4%s-%s-%s",
		hex.EncodeToString(b[0:4]),
		hex.EncodeToString(b[4:6]),
		hex.EncodeToString(b[6:8])[1:],
		hex.EncodeToString(b[8:10]),
		hex.EncodeToString(b[10:16])), nil
}

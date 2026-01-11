package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/davidl71/exarp-go/internal/tools"
)

// handleAllTools handles the stdio://tools resource
// Returns all tools in the catalog
func handleAllTools(ctx context.Context, uri string) ([]byte, string, error) {
	catalog := tools.GetToolCatalog()

	// Convert catalog map to slice (matching tool format)
	var toolList []tools.ToolCatalogEntry
	categories := make(map[string]bool)

	for toolID, tool := range catalog {
		// Create entry with tool ID
		entry := tool
		entry.Tool = toolID
		toolList = append(toolList, entry)
		categories[tool.Category] = true
	}

	// Build category list
	catList := make([]string, 0, len(categories))
	for cat := range categories {
		catList = append(catList, cat)
	}

	// Sort categories (simple alphabetical)
	for i := 0; i < len(catList)-1; i++ {
		for j := i + 1; j < len(catList); j++ {
			if catList[i] > catList[j] {
				catList[i], catList[j] = catList[j], catList[i]
			}
		}
	}

	// Build response (matching tool format)
	response := tools.ToolCatalogResponse{
		Tools:               toolList,
		Count:               len(toolList),
		AvailableCategories: catList,
		AvailablePersonas:   []string{}, // Not currently supported
		FiltersApplied: map[string]interface{}{
			"category": "",
			"persona":  "",
		},
	}

	result, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal tools: %w", err)
	}

	return result, "application/json", nil
}

// handleToolsByCategory handles the stdio://tools/{category} resource
// Returns tools filtered by category
func handleToolsByCategory(ctx context.Context, uri string) ([]byte, string, error) {
	// Parse category from URI: stdio://tools/{category}
	category, err := parseURIVariableByIndexWithValidation(uri, 3, "category", "stdio://tools/{category}")
	if err != nil {
		return nil, "", err
	}

	catalog := tools.GetToolCatalog()

	// Filter tools by category (case-insensitive)
	var filtered []tools.ToolCatalogEntry
	categories := make(map[string]bool)

	for toolID, tool := range catalog {
		// Apply category filter (case-insensitive)
		if !strings.EqualFold(tool.Category, category) {
			continue
		}

		// Create entry
		entry := tool
		entry.Tool = toolID
		filtered = append(filtered, entry)
		categories[tool.Category] = true
	}

	// Build category list
	catList := make([]string, 0, len(categories))
	for cat := range categories {
		catList = append(catList, cat)
	}

	// Sort categories (simple alphabetical)
	for i := 0; i < len(catList)-1; i++ {
		for j := i + 1; j < len(catList); j++ {
			if catList[i] > catList[j] {
				catList[i], catList[j] = catList[j], catList[i]
			}
		}
	}

	// Build response (matching tool format)
	response := tools.ToolCatalogResponse{
		Tools:               filtered,
		Count:               len(filtered),
		AvailableCategories: catList,
		AvailablePersonas:   []string{}, // Not currently supported
		FiltersApplied: map[string]interface{}{
			"category": category,
			"persona":  "",
		},
	}

	result, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal tools: %w", err)
	}

	return result, "application/json", nil
}

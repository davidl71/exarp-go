package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/davidl71/exarp-go/internal/framework"
)

// handleMemoryMaintNative handles the memory_maint tool with native Go maintenance operations
func handleMemoryMaintNative(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	action, _ := params["action"].(string)
	if action == "" {
		action = "health"
	}

	switch action {
	case "health":
		return handleMemoryMaintHealth(ctx, params)
	case "gc":
		return handleMemoryMaintGC(ctx, params)
	case "prune":
		return handleMemoryMaintPrune(ctx, params)
	case "consolidate", "dream":
		// These actions require ML/AI - fall back to Python bridge
		return nil, fmt.Errorf("action %s requires ML/AI capabilities, falling back to Python bridge", action)
	default:
		return nil, fmt.Errorf("unknown action: %s (use 'health', 'gc', 'prune', 'consolidate', or 'dream')", action)
	}
}

// handleMemoryMaintHealth handles health check action
func handleMemoryMaintHealth(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	memories, err := LoadAllMemories(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load memories: %w", err)
	}

	// Calculate statistics
	byCategory := make(map[string]int)
	ageDistribution := make(map[string]int)
	now := time.Now()

	for _, m := range memories {
		byCategory[m.Category]++

		// Parse created_at to determine age
		createdAt, err := time.Parse(time.RFC3339, m.CreatedAt)
		if err != nil {
			// Try alternative format
			createdAt, err = time.Parse("2006-01-02T15:04:05.999999", m.CreatedAt)
			if err != nil {
				continue
			}
		}

		age := now.Sub(createdAt)
		days := int(age.Hours() / 24)

		switch {
		case days < 1:
			ageDistribution["<1 day"]++
		case days < 7:
			ageDistribution["<1 week"]++
		case days < 30:
			ageDistribution["<1 month"]++
		case days < 90:
			ageDistribution["<3 months"]++
		default:
			ageDistribution[">3 months"]++
		}
	}

	// Calculate health score (simplified)
	healthScore := 100.0
	issues := []string{}

	if len(memories) == 0 {
		healthScore = 0
		issues = append(issues, "No memories found")
	} else {
		// Check for very old memories
		if ageDistribution[">3 months"] > len(memories)/2 {
			healthScore -= 20
			issues = append(issues, "More than 50% of memories are older than 3 months")
		}

		// Check for category imbalance
		maxCategoryCount := 0
		for _, count := range byCategory {
			if count > maxCategoryCount {
				maxCategoryCount = count
			}
		}
		if maxCategoryCount > len(memories)*3/4 {
			healthScore -= 10
			issues = append(issues, "One category dominates (>75% of memories)")
		}
	}

	recommendations := []string{}
	if healthScore < 80 {
		recommendations = append(recommendations, "Consider running gc to clean up old memories")
	}
	if len(issues) > 0 {
		recommendations = append(recommendations, "Review memory categories for better organization")
	}

	result := map[string]interface{}{
		"success":          true,
		"method":           "native_go",
		"total_memories":   len(memories),
		"health_score":     healthScore,
		"by_category":      byCategory,
		"age_distribution": ageDistribution,
		"issues":           issues,
		"recommendations":  recommendations,
	}

	output, _ := json.MarshalIndent(result, "", "  ")
	return []framework.TextContent{
		{Type: "text", Text: string(output)},
	}, nil
}

// handleMemoryMaintGC handles garbage collection action
func handleMemoryMaintGC(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	maxAgeDays := 90
	if age, ok := params["max_age_days"].(float64); ok {
		maxAgeDays = int(age)
	}

	deleteOrphaned := true
	if do, ok := params["delete_orphaned"].(bool); ok {
		deleteOrphaned = do
	}

	deleteDuplicates := true
	if dd, ok := params["delete_duplicates"].(bool); ok {
		deleteDuplicates = dd
	}

	scorecardMaxAgeDays := 7
	if age, ok := params["scorecard_max_age_days"].(float64); ok {
		scorecardMaxAgeDays = int(age)
	}

	dryRun := true
	if dr, ok := params["dry_run"].(bool); ok {
		dryRun = dr
	}

	memories, err := LoadAllMemories(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load memories: %w", err)
	}

	now := time.Now()
	toDelete := []string{}
	deletedCounts := map[string]int{
		"old":        0,
		"orphaned":   0,
		"duplicates": 0,
		"scorecard":  0,
	}

	for _, m := range memories {
		shouldDelete := false

		// Parse created_at
		createdAt, err := time.Parse(time.RFC3339, m.CreatedAt)
		if err != nil {
			createdAt, err = time.Parse("2006-01-02T15:04:05.999999", m.CreatedAt)
			if err != nil {
				continue
			}
		}

		age := now.Sub(createdAt)
		days := int(age.Hours() / 24)

		// Check for old memories
		if days > maxAgeDays {
			// Special handling for scorecard memories
			if m.Metadata != nil {
				if mType, ok := m.Metadata["type"].(string); ok && mType == "scorecard" {
					if days > scorecardMaxAgeDays {
						shouldDelete = true
						deletedCounts["scorecard"]++
					}
				} else if days > maxAgeDays {
					shouldDelete = true
					deletedCounts["old"]++
				}
			} else if days > maxAgeDays {
				shouldDelete = true
				deletedCounts["old"]++
			}
		}

		// Check for orphaned memories (no linked tasks, old)
		if deleteOrphaned && len(m.LinkedTasks) == 0 && days > 30 {
			shouldDelete = true
			deletedCounts["orphaned"]++
		}

		// Check for duplicates (simplified - exact title match)
		if deleteDuplicates {
			for _, other := range memories {
				if other.ID != m.ID && other.Title == m.Title && other.Category == m.Category {
					// Found duplicate - keep the newer one
					otherCreatedAt, err := time.Parse(time.RFC3339, other.CreatedAt)
					if err != nil {
						otherCreatedAt, err = time.Parse("2006-01-02T15:04:05.999999", other.CreatedAt)
						if err != nil {
							continue
						}
					}
					if createdAt.Before(otherCreatedAt) {
						shouldDelete = true
						deletedCounts["duplicates"]++
						break
					}
				}
			}
		}

		if shouldDelete {
			toDelete = append(toDelete, m.ID)
		}
	}

	if !dryRun {
		// Delete memory files
		memoriesDir, _ := getMemoriesDir(projectRoot)
		for _, id := range toDelete {
			memoryPath := filepath.Join(memoriesDir, id+".json")
			os.Remove(memoryPath)
		}
	}

	result := map[string]interface{}{
		"success":         true,
		"method":          "native_go",
		"dry_run":         dryRun,
		"deleted_count":   len(toDelete),
		"deleted_by_type": deletedCounts,
		"total_memories":  len(memories),
		"remaining":       len(memories) - len(toDelete),
	}

	if dryRun {
		result["would_delete"] = toDelete
	} else {
		result["deleted_ids"] = toDelete
	}

	output, _ := json.MarshalIndent(result, "", "  ")
	return []framework.TextContent{
		{Type: "text", Text: string(output)},
	}, nil
}

// handleMemoryMaintPrune handles prune action
func handleMemoryMaintPrune(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	valueThreshold := 0.3
	if threshold, ok := params["value_threshold"].(float64); ok {
		valueThreshold = threshold
	}

	keepMinimum := 50
	if min, ok := params["keep_minimum"].(float64); ok {
		keepMinimum = int(min)
	}

	dryRun := true
	if dr, ok := params["dry_run"].(bool); ok {
		dryRun = dr
	}

	memories, err := LoadAllMemories(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load memories: %w", err)
	}

	// Simplified value calculation (Python version may use ML)
	// For now, use simple heuristics:
	// - Scorecard memories have lower value (temporary)
	// - Memories with linked tasks have higher value
	// - Recent memories have higher value
	now := time.Now()
	scored := []struct {
		memory Memory
		value  float64
	}{}

	for _, m := range memories {
		value := 0.5 // Base value

		// Linked tasks increase value
		if len(m.LinkedTasks) > 0 {
			value += 0.3
		}

		// Recent memories have higher value
		createdAt, err := time.Parse(time.RFC3339, m.CreatedAt)
		if err != nil {
			createdAt, err = time.Parse("2006-01-02T15:04:05.999999", m.CreatedAt)
			if err != nil {
				continue
			}
		}
		age := now.Sub(createdAt)
		days := int(age.Hours() / 24)
		if days < 7 {
			value += 0.2
		} else if days < 30 {
			value += 0.1
		}

		// Scorecard memories have lower value
		if m.Metadata != nil {
			if mType, ok := m.Metadata["type"].(string); ok && mType == "scorecard" {
				value -= 0.3
			}
		}

		scored = append(scored, struct {
			memory Memory
			value  float64
		}{memory: m, value: value})
	}

	// Sort by value ascending (lowest value first)
	for i := 0; i < len(scored)-1; i++ {
		for j := i + 1; j < len(scored); j++ {
			if scored[i].value > scored[j].value {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}

	// Determine which to prune (keep at least keepMinimum)
	toPrune := []string{}
	for i, s := range scored {
		if i < len(scored)-keepMinimum && s.value < valueThreshold {
			toPrune = append(toPrune, s.memory.ID)
		}
	}

	if !dryRun {
		// Delete memory files
		memoriesDir, _ := getMemoriesDir(projectRoot)
		for _, id := range toPrune {
			memoryPath := filepath.Join(memoriesDir, id+".json")
			os.Remove(memoryPath)
		}
	}

	result := map[string]interface{}{
		"success":         true,
		"method":          "native_go",
		"dry_run":         dryRun,
		"pruned_count":    len(toPrune),
		"total_memories":  len(memories),
		"remaining":       len(memories) - len(toPrune),
		"value_threshold": valueThreshold,
		"keep_minimum":    keepMinimum,
	}

	if dryRun {
		result["would_prune"] = toPrune
	} else {
		result["pruned_ids"] = toPrune
	}

	output, _ := json.MarshalIndent(result, "", "  ")
	return []framework.TextContent{
		{Type: "text", Text: string(output)},
	}, nil
}

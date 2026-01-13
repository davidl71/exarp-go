package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/davidl71/exarp-go/internal/config"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/internal/security"
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
	case "consolidate":
		return handleMemoryMaintConsolidate(ctx, params)
	case "dream":
		return handleMemoryMaintDream(ctx, params)
	default:
		return nil, fmt.Errorf("unknown action: %s (use 'health', 'gc', 'prune', 'consolidate', or 'dream')", action)
	}
}

// handleMemoryMaintHealth handles health check action
func handleMemoryMaintHealth(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	projectRoot, err := security.GetProjectRoot(".")
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
	projectRoot, err := security.GetProjectRoot(".")
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
	projectRoot, err := security.GetProjectRoot(".")
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

// handleMemoryMaintConsolidate handles consolidate action
func handleMemoryMaintConsolidate(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	projectRoot, err := security.GetProjectRoot(".")
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	similarityThreshold := config.SimilarityThreshold()
	if threshold, ok := params["similarity_threshold"].(float64); ok {
		similarityThreshold = threshold
	}

	mergeStrategy := "newest"
	if strategy, ok := params["merge_strategy"].(string); ok && strategy != "" {
		mergeStrategy = strategy
	}

	dryRun := true
	if dr, ok := params["dry_run"].(bool); ok {
		dryRun = dr
	}

	memories, err := LoadAllMemories(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load memories: %w", err)
	}

	// Group memories by category for efficiency
	byCategory := make(map[string][]Memory)
	for _, m := range memories {
		byCategory[m.Category] = append(byCategory[m.Category], m)
	}

	// Find duplicate groups (similar titles within same category)
	groups := findDuplicateGroups(byCategory, similarityThreshold)

	mergedCount := 0
	deletedIDs := []string{}
	mergedResults := []map[string]interface{}{}

	for _, group := range groups {
		if len(group) < 2 {
			continue
		}

		if !dryRun {
			// Merge the group
			merged := mergeMemories(group, mergeStrategy)

			// Save the merged memory
			if err := saveMemory(projectRoot, merged); err != nil {
				return nil, fmt.Errorf("failed to save merged memory: %w", err)
			}

			// Delete the others
			baseID := merged.ID
			memoriesDir, _ := getMemoriesDir(projectRoot)
			for _, m := range group {
				if m.ID != baseID {
					memoryPath := filepath.Join(memoriesDir, m.ID+".json")
					if err := os.Remove(memoryPath); err == nil {
						deletedIDs = append(deletedIDs, m.ID)
					}
				}
			}

			mergedCount++
		}

		// Track which will be merged
		titles := make([]string, len(group))
		for i, m := range group {
			titles[i] = m.Title
		}

		// Determine merge target ID
		var mergeTargetID string
		switch mergeStrategy {
		case "newest":
			latest := group[0]
			for _, m := range group[1:] {
				if m.CreatedAt > latest.CreatedAt {
					latest = m
				}
			}
			mergeTargetID = latest.ID
		case "oldest":
			oldest := group[0]
			for _, m := range group[1:] {
				if m.CreatedAt < oldest.CreatedAt {
					oldest = m
				}
			}
			mergeTargetID = oldest.ID
		case "longest":
			longest := group[0]
			for _, m := range group[1:] {
				if len(m.Content) > len(longest.Content) {
					longest = m
				}
			}
			mergeTargetID = longest.ID
		default:
			mergeTargetID = group[0].ID
		}

		mergedResults = append(mergedResults, map[string]interface{}{
			"group_size": len(group),
			"titles":     titles,
			"category":   group[0].Category,
			"base_id":    group[0].ID,
			"merge_into": mergeTargetID,
		})
	}

	result := map[string]interface{}{
		"success":              true,
		"method":               "native_go",
		"dry_run":              dryRun,
		"total_memories":       len(memories),
		"similarity_threshold": similarityThreshold,
		"merge_strategy":       mergeStrategy,
		"groups_found":         len(groups),
		"merged_count":         mergedCount,
		"deleted_count":        len(deletedIDs),
		"merged_results":       mergedResults,
	}

	if dryRun {
		result["would_merge"] = mergedResults
	} else {
		result["deleted_ids"] = deletedIDs
	}

	output, _ := json.MarshalIndent(result, "", "  ")
	return []framework.TextContent{
		{Type: "text", Text: string(output)},
	}, nil
}

// handleMemoryMaintDream handles the dream action for memory_maint tool
// Uses devwisdom-go wisdom engine directly (no MCP client needed)
func handleMemoryMaintDream(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	// Get score (default: 50)
	var score float64 = 50.0
	if sc, ok := params["score"].(float64); ok {
		score = sc
	} else if sc, ok := params["score"].(int); ok {
		score = float64(sc)
	}

	// Validate and clamp score
	if score < 0 {
		score = 0
	} else if score > 100 {
		score = 100
	}

	// Get wisdom engine
	engine, err := getWisdomEngine()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize wisdom engine: %w", err)
	}

	// Get wisdom quote (use "random" source for variety)
	quote, err := engine.GetWisdom(score, "random")
	if err != nil {
		return nil, fmt.Errorf("failed to get wisdom quote: %w", err)
	}

	// Build dream result (simple format for now)
	dream := map[string]interface{}{
		"quote":         quote.Quote,
		"source":        quote.Source,
		"encouragement": quote.Encouragement,
		"score":         score,
		"timestamp":     time.Now().Format(time.RFC3339),
	}

	// Convert to JSON
	resultJSON, err := json.MarshalIndent(dream, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal dream: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: string(resultJSON)},
	}, nil
}

// similarityRatio calculates similarity ratio between two strings (0.0 - 1.0)
// Simple implementation using character matching
func similarityRatio(a, b string) float64 {
	a = strings.ToLower(a)
	b = strings.ToLower(b)

	if a == b {
		return 1.0
	}

	if len(a) == 0 || len(b) == 0 {
		return 0.0
	}

	// Simple character-based similarity
	charCountsA := make(map[rune]int)
	charCountsB := make(map[rune]int)

	for _, char := range a {
		charCountsA[char]++
	}
	for _, char := range b {
		charCountsB[char]++
	}

	matches := 0
	for char, countA := range charCountsA {
		countB := charCountsB[char]
		if countA < countB {
			matches += countA
		} else {
			matches += countB
		}
	}

	maxLen := len(a)
	if len(b) > maxLen {
		maxLen = len(b)
	}

	if maxLen == 0 {
		return 0.0
	}

	return float64(matches) / float64(maxLen)
}

// findDuplicateGroups finds groups of similar memories based on title similarity
func findDuplicateGroups(byCategory map[string][]Memory, threshold float64) [][]Memory {
	allGroups := [][]Memory{}

	for _, categoryMemories := range byCategory {
		grouped := make(map[string]bool) // Track which memories have been grouped

		for i, m1 := range categoryMemories {
			if grouped[m1.ID] {
				continue
			}

			group := []Memory{m1}
			grouped[m1.ID] = true

			for j := i + 1; j < len(categoryMemories); j++ {
				m2 := categoryMemories[j]
				if grouped[m2.ID] {
					continue
				}

				// Compare titles
				titleSim := similarityRatio(m1.Title, m2.Title)
				if titleSim >= threshold {
					group = append(group, m2)
					grouped[m2.ID] = true
				}
			}

			if len(group) > 1 {
				allGroups = append(allGroups, group)
			}
		}
	}

	return allGroups
}

// mergeMemories merges multiple similar memories into one
func mergeMemories(memories []Memory, strategy string) Memory {
	if len(memories) == 0 {
		return Memory{}
	}
	if len(memories) == 1 {
		return memories[0]
	}

	// Sort by strategy
	var base Memory
	switch strategy {
	case "newest":
		base = memories[0]
		for _, m := range memories[1:] {
			if m.CreatedAt > base.CreatedAt {
				base = m
			}
		}
	case "oldest":
		base = memories[0]
		for _, m := range memories[1:] {
			if m.CreatedAt < base.CreatedAt {
				base = m
			}
		}
	case "longest":
		base = memories[0]
		for _, m := range memories[1:] {
			if len(m.Content) > len(base.Content) {
				base = m
			}
		}
	default:
		base = memories[0]
	}

	// Combine unique content from all memories
	allContent := []string{base.Content}
	allTasks := make(map[string]bool)
	for _, taskID := range base.LinkedTasks {
		allTasks[taskID] = true
	}

	for _, m := range memories {
		if m.ID == base.ID {
			continue
		}

		// Only add if meaningfully different
		if m.Content != "" && similarityRatio(m.Content, base.Content) < 0.9 {
			allContent = append(allContent, fmt.Sprintf("\n--- Merged from %s ---\n%s", m.Title, m.Content))
		}

		// Combine linked tasks
		for _, taskID := range m.LinkedTasks {
			allTasks[taskID] = true
		}
	}

	// Create merged memory
	mergedTasks := []string{}
	for taskID := range allTasks {
		mergedTasks = append(mergedTasks, taskID)
	}

	mergedFrom := []string{}
	for _, m := range memories {
		if m.ID != base.ID {
			mergedFrom = append(mergedFrom, m.ID)
		}
	}

	mergedMetadata := base.Metadata
	if mergedMetadata == nil {
		mergedMetadata = make(map[string]interface{})
	}
	mergedMetadata["merged_from"] = mergedFrom
	mergedMetadata["merged_at"] = time.Now().Format(time.RFC3339)

	return Memory{
		ID:          base.ID,
		Title:       base.Title,
		Content:     strings.Join(allContent, "\n"),
		Category:    base.Category,
		LinkedTasks: mergedTasks,
		Metadata:    mergedMetadata,
		CreatedAt:   base.CreatedAt,
		SessionDate: base.SessionDate,
	}
}

// memory_maint_utils.go — Memory maintenance: dream, similarity scoring, duplicate groups, and memory merge.
// See also: memory_maint.go
package tools

import (
	"context"
	"fmt"
	"strings"
	"time"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/mcp-go-core/pkg/mcp/response"
)

// ─── Contents ───────────────────────────────────────────────────────────────
//   handleMemoryMaintDream
//   similarityRatio — similarityRatio calculates similarity ratio between two strings (0.0 - 1.0)
//   findDuplicateGroups — findDuplicateGroups finds groups of similar memories based on title similarity.
//   mergeMemories — mergeMemories merges multiple similar memories into one.
// ────────────────────────────────────────────────────────────────────────────

// ─── handleMemoryMaintDream ─────────────────────────────────────────────────
func handleMemoryMaintDream(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	// Get score (default: 50)
	var score = 50.0
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

	return response.FormatResult(dream, "")
}

// ─── similarityRatio ────────────────────────────────────────────────────────
// similarityRatio calculates similarity ratio between two strings (0.0 - 1.0)
// Simple implementation using character matching.
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

// ─── findDuplicateGroups ────────────────────────────────────────────────────
// findDuplicateGroups finds groups of similar memories based on title similarity.
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

// ─── mergeMemories ──────────────────────────────────────────────────────────
// mergeMemories merges multiple similar memories into one.
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

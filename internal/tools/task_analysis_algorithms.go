// task_analysis_algorithms.go — task_analysis algorithms helpers.
package tools

import (
	"fmt"
	"strings"
	"sync"

	"github.com/davidl71/exarp-go/internal/models"
)

func truncateStr(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

// Helper functions for duplicates detection

func findDuplicateTasks(tasks []Todo2Task, threshold float64) [][]string {
	// Fast pass: exact duplicates by content hash (O(n) vs O(n²) similarity)
	hashGroups := make(map[string][]string)
	for i := range tasks {
		hash := models.GetContentHash(&tasks[i])
		if hash != "" {
			hashGroups[hash] = append(hashGroups[hash], tasks[i].ID)
		}
	}

	var groups [][]string
	exactIDs := make(map[string]bool)
	for _, ids := range hashGroups {
		if len(ids) > 1 {
			groups = append(groups, ids)
			for _, id := range ids {
				exactIDs[id] = true
			}
		}
	}

	// Filter out exact-match tasks before running expensive similarity check
	remaining := make([]Todo2Task, 0, len(tasks))
	for _, t := range tasks {
		if !exactIDs[t.ID] {
			remaining = append(remaining, t)
		}
	}

	// Similarity pass on remaining tasks
	var similarGroups [][]string
	if len(remaining) < 100 {
		similarGroups = findDuplicateTasksSequential(remaining, threshold)
	} else {
		similarGroups = findDuplicateTasksParallel(remaining, threshold)
	}

	return append(groups, similarGroups...)
}

// findDuplicateTasksSequential is the original sequential implementation
// Optimized for small datasets where parallel overhead isn't worth it.
func findDuplicateTasksSequential(tasks []Todo2Task, threshold float64) [][]string {
	duplicates := [][]string{}
	processed := make(map[string]bool)

	for i, task1 := range tasks {
		if processed[task1.ID] {
			continue
		}

		group := []string{task1.ID}

		for j := i + 1; j < len(tasks); j++ {
			task2 := tasks[j]
			if processed[task2.ID] {
				continue
			}

			similarity := calculateSimilarity(task1, task2)
			if similarity >= threshold {
				group = append(group, task2.ID)
				processed[task2.ID] = true
			}
		}

		if len(group) > 1 {
			duplicates = append(duplicates, group)
			processed[task1.ID] = true
		}
	}

	return duplicates
}

// findDuplicateTasksParallel uses worker pool for parallel duplicate detection
// Optimized for large datasets (>100 tasks).
func findDuplicateTasksParallel(tasks []Todo2Task, threshold float64) [][]string {
	const numWorkers = 4 // Adjust based on CPU cores

	if len(tasks) == 0 {
		return [][]string{}
	}

	type taskPair struct {
		i     int
		task1 Todo2Task
	}

	type similarityResult struct {
		i          int
		j          int
		similarity float64
	}

	// Channel for task pairs to process
	taskPairs := make(chan taskPair, len(tasks))
	results := make(chan similarityResult, len(tasks)*len(tasks))

	// Start worker goroutines
	var wg sync.WaitGroup
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for pair := range taskPairs {
				task1 := pair.task1

				for j := pair.i + 1; j < len(tasks); j++ {
					task2 := tasks[j]

					similarity := calculateSimilarity(task1, task2)
					if similarity >= threshold {
						results <- similarityResult{
							i:          pair.i,
							j:          j,
							similarity: similarity,
						}
					}
				}
			}
		}()
	}

	// Send all task pairs to workers
	go func() {
		defer close(taskPairs)

		for i, task1 := range tasks {
			taskPairs <- taskPair{i: i, task1: task1}
		}
	}()

	// Close results channel when all workers done
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	processed := make(map[string]bool)
	duplicateMap := make(map[int][]int) // task index -> list of duplicate indices

	for result := range results {
		if !processed[tasks[result.i].ID] && !processed[tasks[result.j].ID] {
			if duplicateMap[result.i] == nil {
				duplicateMap[result.i] = []int{result.i}
			}

			duplicateMap[result.i] = append(duplicateMap[result.i], result.j)
			processed[tasks[result.j].ID] = true
		}
	}

	// Build duplicate groups
	duplicates := [][]string{}

	for i, group := range duplicateMap {
		if len(group) > 1 && !processed[tasks[i].ID] {
			groupIDs := make([]string, len(group))
			for idx, taskIdx := range group {
				groupIDs[idx] = tasks[taskIdx].ID
			}

			duplicates = append(duplicates, groupIDs)
			processed[tasks[i].ID] = true
		}
	}

	return duplicates
}

func calculateSimilarity(task1, task2 Todo2Task) float64 {
	// Simple word-based similarity (can be enhanced with Levenshtein)
	text1 := strings.ToLower(task1.Content + " " + task1.LongDescription)
	text2 := strings.ToLower(task2.Content + " " + task2.LongDescription)

	words1 := strings.Fields(text1)
	words2 := strings.Fields(text2)

	if len(words1) == 0 && len(words2) == 0 {
		return 1.0
	}

	if len(words1) == 0 || len(words2) == 0 {
		return 0.0
	}

	// Count common words
	wordSet1 := make(map[string]bool)
	for _, word := range words1 {
		wordSet1[word] = true
	}

	common := 0

	for _, word := range words2 {
		if wordSet1[word] {
			common++
		}
	}

	// Jaccard similarity
	union := len(wordSet1) + len(words2) - common
	if union == 0 {
		return 0.0
	}

	return float64(common) / float64(union)
}

func mergeDuplicateTasks(tasks []Todo2Task, duplicates [][]string) []Todo2Task {
	// Keep first task in each group, merge others into it
	keepMap := make(map[string]bool)
	removeMap := make(map[string]bool)

	for _, group := range duplicates {
		if len(group) == 0 {
			continue
		}

		keepMap[group[0]] = true

		for i := 1; i < len(group); i++ {
			removeMap[group[i]] = true
		}
	}

	// Merge task data
	taskMap := make(map[string]*Todo2Task)
	for i := range tasks {
		taskMap[tasks[i].ID] = &tasks[i]
	}

	for _, group := range duplicates {
		if len(group) < 2 {
			continue
		}

		primary := taskMap[group[0]]

		for i := 1; i < len(group); i++ {
			secondary := taskMap[group[i]]
			// Merge tags
			tagSet := make(map[string]bool)
			for _, tag := range primary.Tags {
				tagSet[tag] = true
			}

			for _, tag := range secondary.Tags {
				if !tagSet[tag] {
					primary.Tags = append(primary.Tags, tag)
				}
			}
			// Merge dependencies
			depSet := make(map[string]bool)
			for _, dep := range primary.Dependencies {
				depSet[dep] = true
			}

			for _, dep := range secondary.Dependencies {
				if !depSet[dep] {
					primary.Dependencies = append(primary.Dependencies, dep)
				}
			}
		}
	}

	// Remove duplicate tasks
	result := []Todo2Task{}

	for _, task := range tasks {
		if !removeMap[task.ID] {
			result = append(result, task)
		}
	}

	return result
}

// Helper functions for tag analysis

type TagAnalysis struct {
	TagFrequency     map[string]int      `json:"tag_frequency"`
	TagSuggestions   map[string][]string `json:"tag_suggestions"`
	InconsistentTags map[string][]string `json:"inconsistent_tags"`
	UnusedTags       []string            `json:"unused_tags"`
	RenameRules      map[string]string   `json:"rename_rules,omitempty"`   // oldTag -> newTag
	TagsToRemove     []string            `json:"tags_to_remove,omitempty"` // tags to delete
}

func analyzeTags(tasks []Todo2Task) TagAnalysis {
	analysis := TagAnalysis{
		TagFrequency:     make(map[string]int),
		TagSuggestions:   make(map[string][]string),
		InconsistentTags: make(map[string][]string),
	}

	tagSet := make(map[string]bool)

	for _, task := range tasks {
		for _, tag := range task.Tags {
			analysis.TagFrequency[tag]++
			tagSet[tag] = true
		}
	}

	// Suggest tags from semantic analysis of title + content + existing tags (all tasks)
	existingSet := make(map[string]bool)

	for _, task := range tasks {
		suggestions := suggestTagsForTask(task)
		if len(suggestions) > 0 {
			analysis.TagSuggestions[task.ID] = suggestions
		}

		_ = existingSet
	}

	return analysis
}

func suggestTagsForTask(task Todo2Task) []string {
	title := strings.ToLower(task.Content)
	body := strings.ToLower(task.LongDescription)
	content := title + " " + body

	existingSet := make(map[string]bool)
	for _, t := range task.Tags {
		existingSet[strings.ToLower(t)] = true
	}

	keywords := map[string]string{
		"bug": "bug", "fix": "bug", "error": "bug",
		"feature": "feature", "add": "feature", "implement": "feature",
		"refactor": "refactor", "cleanup": "refactor", "improve": "refactor",
		"test": "testing", "testing": "testing",
		"doc": "docs", "documentation": "docs", "docs": "docs",
		"migration": "migration", "migrate": "migration",
		"config": "config", "cli": "cli", "mcp": "mcp",
		"security": "security", "performance": "performance", "database": "database",
	}

	seen := make(map[string]bool)
	suggestions := []string{}

	for keyword, tag := range keywords {
		if existingSet[tag] || seen[tag] {
			continue
		}

		if strings.Contains(content, keyword) {
			seen[tag] = true

			suggestions = append(suggestions, tag)
		}
	}

	return suggestions
}

func applyTagRules(tasks []Todo2Task, rules map[string]string, analysis TagAnalysis) TagAnalysis {
	// Store rename rules for later application
	if analysis.RenameRules == nil {
		analysis.RenameRules = make(map[string]string)
	}
	// Apply rename rules to frequency map and store for task updates
	for oldTag, newTag := range rules {
		if count, ok := analysis.TagFrequency[oldTag]; ok {
			delete(analysis.TagFrequency, oldTag)
			analysis.TagFrequency[newTag] += count
			analysis.RenameRules[oldTag] = newTag
		}
	}

	return analysis
}

func removeTags(tasks []Todo2Task, tagsToRemove []string, analysis TagAnalysis) TagAnalysis {
	// Store tags to remove for later application
	analysis.TagsToRemove = append(analysis.TagsToRemove, tagsToRemove...)
	for _, tag := range tagsToRemove {
		delete(analysis.TagFrequency, tag)
	}

	return analysis
}

func applyTagChanges(tasks []Todo2Task, analysis TagAnalysis) []Todo2Task {
	removeSet := make(map[string]bool)
	for _, tag := range analysis.TagsToRemove {
		removeSet[tag] = true
	}

	for i := range tasks {
		// Start with current tags (after renames/removals)
		newTags := make([]string, 0, len(tasks[i].Tags))
		seen := make(map[string]bool)

		for _, tag := range tasks[i].Tags {
			if removeSet[tag] {
				continue
			}

			finalTag := tag
			if newTag, ok := analysis.RenameRules[tag]; ok {
				finalTag = newTag
			}

			if !seen[finalTag] {
				seen[finalTag] = true

				newTags = append(newTags, finalTag)
			}
		}

		// Quick tag addition: merge semantic/keyword suggestions (don't duplicate)
		for _, tag := range analysis.TagSuggestions[tasks[i].ID] {
			if removeSet[tag] {
				continue
			}

			finalTag := tag
			if newTag, ok := analysis.RenameRules[tag]; ok {
				finalTag = newTag
			}

			if !seen[finalTag] {
				seen[finalTag] = true

				newTags = append(newTags, finalTag)
			}
		}

		tasks[i].Tags = newTags
	}

	return tasks
}

func buildTagRecommendations(analysis TagAnalysis) []map[string]interface{} {
	recommendations := []map[string]interface{}{}

	// Recommend consolidating low-frequency tags
	for tag, count := range analysis.TagFrequency {
		if count < 2 {
			recommendations = append(recommendations, map[string]interface{}{
				"type":    "consolidate",
				"tag":     tag,
				"count":   count,
				"message": fmt.Sprintf("Tag '%s' is only used %d time(s), consider consolidating", tag, count),
			})
		}
	}

	return recommendations
}

// Helper functions for dependency analysis

// DependencyGraph is the legacy format for backward compatibility.
type DependencyGraph map[string][]string

// buildLegacyGraphFormat converts TaskGraph to legacy map format for backward compatibility.

// task_analysis_tags_discover.go — Tag discovery: markdown docs, project tag caching, and batch size helpers.
// See also: task_analysis_tags.go, task_analysis_tags_llm.go
package tools

import (
	"context"
	"errors"
	"github.com/davidl71/exarp-go/internal/database"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// ─── Contents ───────────────────────────────────────────────────────────────
//   discoverTagsFromMarkdown
//   discoverTagsFromMarkdownWithCache — discoverTagsFromMarkdownWithCache scans markdown files with SQLite caching.
//   saveDiscoveriesToCache — saveDiscoveriesToCache saves tag discoveries to SQLite cache.
//   saveFileTaskTagsToCache — saveFileTaskTagsToCache saves file-task tag matches to cache.
//   updateTagFrequencyCache — updateTagFrequencyCache updates tag frequency statistics in cache.
//   collectProjectTags — collectProjectTags extracts all unique tags from existing tasks.
//   projectTagsWithCacheHints — projectTagsWithCacheHints merges project tags from tasks with top tag frequencies from cache (for LLM hints everywhere).
//   getCanonicalTagsList — getCanonicalTagsList returns the list of canonical tag categories.
//   effectiveLLMBatchSize
//   suggestNextLLMBatchSize — suggestNextLLMBatchSize returns a suggested llm_batch_size for the next run (feedback loop).
//   batchTimeout — batchTimeout returns per-batch timeout; scales slightly with batch size for large batches.
//   enrichTagsWithLLM — enrichTagsWithLLM uses Apple FM to infer additional tags, in batches of files per call to speed up.
// ────────────────────────────────────────────────────────────────────────────

// ─── discoverTagsFromMarkdown ───────────────────────────────────────────────
func discoverTagsFromMarkdown(projectRoot, docPath string) []map[string]interface{} {
	discoveries := []map[string]interface{}{}

	searchPath := filepath.Join(projectRoot, docPath)
	if _, err := os.Stat(searchPath); os.IsNotExist(err) {
		return discoveries
	}

	// Patterns to match tags in markdown
	hashtagPattern := regexp.MustCompile(`#([a-zA-Z][a-zA-Z0-9_-]+)`)                    // #tag
	bracketPattern := regexp.MustCompile(`\[([a-zA-Z][a-zA-Z0-9_-]+)\]`)                 // [tag]
	tagsLinePattern := regexp.MustCompile(`(?i)^(?:tags?|labels?|categories?):\s*(.+)$`) // Tags: tag1, tag2

	if err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			if strings.Contains(path, "/archive/") {
				return filepath.SkipDir
			}

			return nil
		}

		if filepath.Ext(path) != ".md" && filepath.Ext(path) != ".markdown" {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		relPath := strings.TrimPrefix(path, projectRoot+"/")
		fileContent := string(content)
		lines := strings.Split(fileContent, "\n")

		discoveredTags := []string{}
		tagSources := []string{}

		// Find hashtags
		for _, match := range hashtagPattern.FindAllStringSubmatch(fileContent, -1) {
			if len(match) >= 2 {
				tag := strings.ToLower(match[1])
				// Skip common false positives
				if tag != "todo" && tag != "fixme" && tag != "note" && tag != "warning" &&
					tag != "deprecated" && tag != "see" && tag != "param" && tag != "return" &&
					len(tag) > 2 {
					discoveredTags = append(discoveredTags, tag)
					tagSources = append(tagSources, "hashtag")
				}
			}
		}

		// Find bracket tags (less common, skip if looks like link or code type)
		bracketFalsePositives := map[string]bool{
			"string": true, "int": true, "bool": true, "float": true, "number": true,
			"object": true, "array": true, "map": true, "slice": true, "struct": true,
			"interface": true, "error": true, "any": true, "void": true, "null": true,
			"true": true, "false": true, "nil": true, "none": true,
			"optional": true, "required": true, "default": true,
			"link": true, "url": true, "path": true, "file": true,
			"date": true, "time": true, "timestamp": true,
			"tid": true, "pid": true, "uid": true, "gid": true,
			"key": true, "value": true, "name": true, "type": true,
			"status": true, "state": true, "mode": true, "action": true,
			"observations": true, "example": true, "examples": true,
		}

		for _, match := range bracketPattern.FindAllStringSubmatch(fileContent, -1) {
			if len(match) >= 2 {
				tag := strings.ToLower(match[1])
				// Skip if looks like markdown link text, code type, or common words
				if len(tag) > 2 && len(tag) < 30 && !strings.Contains(tag, " ") && !bracketFalsePositives[tag] {
					discoveredTags = append(discoveredTags, tag)
					tagSources = append(tagSources, "bracket")
				}
			}
		}

		// Find explicit tags lines
		for _, line := range lines {
			if matches := tagsLinePattern.FindStringSubmatch(line); len(matches) >= 2 {
				tagList := strings.Split(matches[1], ",")
				for _, t := range tagList {
					tag := strings.ToLower(strings.TrimSpace(t))
					tag = strings.Trim(tag, "#[]`")

					if len(tag) > 1 {
						discoveredTags = append(discoveredTags, tag)
						tagSources = append(tagSources, "explicit")
					}
				}
			}
		}

		// Dedupe tags
		seen := make(map[string]bool)
		uniqueTags := []string{}

		for _, tag := range discoveredTags {
			if !seen[tag] {
				seen[tag] = true

				uniqueTags = append(uniqueTags, tag)
			}
		}

		if len(uniqueTags) > 0 {
			discoveries = append(discoveries, map[string]interface{}{
				"file":    relPath,
				"tags":    uniqueTags,
				"sources": tagSources,
				"type":    "markdown_discovery",
			})
		}

		return nil
	}); err != nil {
		return discoveries
	}

	return discoveries
}

// ─── discoverTagsFromMarkdownWithCache ──────────────────────────────────────
// discoverTagsFromMarkdownWithCache scans markdown files with SQLite caching.
func discoverTagsFromMarkdownWithCache(projectRoot, docPath string, useCache bool) ([]map[string]interface{}, int, int) {
	discoveries := []map[string]interface{}{}
	cacheHits := 0
	cacheMisses := 0

	searchPath := filepath.Join(projectRoot, docPath)
	if _, err := os.Stat(searchPath); os.IsNotExist(err) {
		return discoveries, cacheHits, cacheMisses
	}

	// Patterns to match tags in markdown
	hashtagPattern := regexp.MustCompile(`#([a-zA-Z][a-zA-Z0-9_-]+)`)
	bracketPattern := regexp.MustCompile(`\[([a-zA-Z][a-zA-Z0-9_-]+)\]`)
	tagsLinePattern := regexp.MustCompile(`(?i)^(?:tags?|labels?|categories?):\s*(.+)$`)

	bracketFalsePositives := map[string]bool{
		"string": true, "int": true, "bool": true, "float": true, "number": true,
		"object": true, "array": true, "map": true, "slice": true, "struct": true,
		"interface": true, "error": true, "any": true, "void": true, "null": true,
		"true": true, "false": true, "nil": true, "none": true,
		"optional": true, "required": true, "default": true,
		"link": true, "url": true, "path": true, "file": true,
		"date": true, "time": true, "timestamp": true,
		"tid": true, "pid": true, "uid": true, "gid": true,
		"key": true, "value": true, "name": true, "type": true,
		"status": true, "state": true, "mode": true, "action": true,
		"observations": true, "example": true, "examples": true,
	}

	if err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			if strings.Contains(path, "/archive/") {
				return filepath.SkipDir
			}

			return nil
		}

		if filepath.Ext(path) != ".md" && filepath.Ext(path) != ".markdown" {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		relPath := strings.TrimPrefix(path, projectRoot+"/")
		fileHash := database.ComputeFileHash(content)

		// Check cache first
		if useCache {
			cachedTags, valid, _ := database.GetDiscoveredTagsWithHash(relPath, fileHash)
			if valid && len(cachedTags) > 0 {
				cacheHits++
				// Convert cached tags to discovery format
				tags := []string{}
				sources := []string{}

				for _, ct := range cachedTags {
					tags = append(tags, ct.Tag)
					sources = append(sources, ct.Source)
				}

				discoveries = append(discoveries, map[string]interface{}{
					"file":       relPath,
					"tags":       tags,
					"sources":    sources,
					"type":       "markdown_discovery",
					"from_cache": true,
				})

				return nil
			}
		}

		cacheMisses++
		fileContent := string(content)
		lines := strings.Split(fileContent, "\n")

		discoveredTags := []string{}
		tagSources := []string{}

		// Find hashtags
		for _, match := range hashtagPattern.FindAllStringSubmatch(fileContent, -1) {
			if len(match) >= 2 {
				tag := strings.ToLower(match[1])
				if tag != "todo" && tag != "fixme" && tag != "note" && tag != "warning" &&
					tag != "deprecated" && tag != "see" && tag != "param" && tag != "return" &&
					len(tag) > 2 {
					discoveredTags = append(discoveredTags, tag)
					tagSources = append(tagSources, "hashtag")
				}
			}
		}

		// Find bracket tags
		for _, match := range bracketPattern.FindAllStringSubmatch(fileContent, -1) {
			if len(match) >= 2 {
				tag := strings.ToLower(match[1])
				if len(tag) > 2 && len(tag) < 30 && !strings.Contains(tag, " ") && !bracketFalsePositives[tag] {
					discoveredTags = append(discoveredTags, tag)
					tagSources = append(tagSources, "bracket")
				}
			}
		}

		// Find explicit tags lines
		for _, line := range lines {
			if matches := tagsLinePattern.FindStringSubmatch(line); len(matches) >= 2 {
				tagList := strings.Split(matches[1], ",")
				for _, t := range tagList {
					tag := strings.ToLower(strings.TrimSpace(t))
					tag = strings.Trim(tag, "#[]`")

					if len(tag) > 1 {
						discoveredTags = append(discoveredTags, tag)
						tagSources = append(tagSources, "explicit")
					}
				}
			}
		}

		// Dedupe tags
		seen := make(map[string]bool)
		uniqueTags := []string{}
		uniqueSources := []string{}

		for i, tag := range discoveredTags {
			if !seen[tag] {
				seen[tag] = true

				uniqueTags = append(uniqueTags, tag)

				if i < len(tagSources) {
					uniqueSources = append(uniqueSources, tagSources[i])
				}
			}
		}

		if len(uniqueTags) > 0 {
			discoveries = append(discoveries, map[string]interface{}{
				"file":      relPath,
				"tags":      uniqueTags,
				"sources":   uniqueSources,
				"type":      "markdown_discovery",
				"file_hash": fileHash,
			})
		}

		return nil
	}); err != nil {
		return discoveries, cacheHits, cacheMisses
	}

	return discoveries, cacheHits, cacheMisses
}

// ─── saveDiscoveriesToCache ─────────────────────────────────────────────────
// saveDiscoveriesToCache saves tag discoveries to SQLite cache.
func saveDiscoveriesToCache(discoveries []map[string]interface{}) {
	for _, discovery := range discoveries {
		filePath, _ := discovery["file"].(string)
		fileHash, _ := discovery["file_hash"].(string)
		tags, _ := discovery["tags"].([]string)
		sources, _ := discovery["sources"].([]string)
		llmSuggestions, _ := discovery["llm_suggestions"].([]string)

		if filePath == "" || len(tags) == 0 {
			continue
		}

		// Build cache entries
		var cacheEntries []database.DiscoveredTag

		for i, tag := range tags {
			source := "unknown"
			if i < len(sources) {
				source = sources[i]
			}

			cacheEntries = append(cacheEntries, database.DiscoveredTag{
				FilePath:     filePath,
				FileHash:     fileHash,
				Tag:          tag,
				Source:       source,
				LLMSuggested: false,
			})
		}

		// Add LLM suggestions
		for _, tag := range llmSuggestions {
			cacheEntries = append(cacheEntries, database.DiscoveredTag{
				FilePath:     filePath,
				FileHash:     fileHash,
				Tag:          tag,
				Source:       "llm",
				LLMSuggested: true,
			})
		}

		// Save to database
		if err := database.SaveDiscoveredTags(filePath, fileHash, cacheEntries); err != nil {
			continue
		}
	}
}

// ─── saveFileTaskTagsToCache ────────────────────────────────────────────────
// saveFileTaskTagsToCache saves file-task tag matches to cache.
func saveFileTaskTagsToCache(tagMatches map[string][]string, discoveries []map[string]interface{}) {
	// Build file-to-tags map from discoveries
	fileTags := make(map[string][]string)

	for _, discovery := range discoveries {
		filePath, _ := discovery["file"].(string)
		tags, _ := discovery["tags"].([]string)

		if filePath != "" && len(tags) > 0 {
			fileTags[filePath] = tags
		}
	}

	// For each task match, save the file-task-tag relationship
	for taskID, tags := range tagMatches {
		for filePath, ftags := range fileTags {
			for _, tag := range tags {
				// Check if this tag came from this file
				for _, ftag := range ftags {
					if ftag == tag {
						if err := database.SaveFileTaskTag(filePath, taskID, tag, true); err != nil {
							continue
						}

						break
					}
				}
			}
		}
	}
}

// ─── updateTagFrequencyCache ────────────────────────────────────────────────
// updateTagFrequencyCache updates tag frequency statistics in cache.
func updateTagFrequencyCache(tasks []Todo2Task) {
	tagCounts := make(map[string]int)
	canonicalTags := getCanonicalTagsList()
	canonicalSet := make(map[string]bool)

	for _, tag := range canonicalTags {
		canonicalSet[tag] = true
	}

	for _, task := range tasks {
		for _, tag := range task.Tags {
			tagCounts[tag]++
		}
	}

	for tag, count := range tagCounts {
		isCanonical := canonicalSet[tag]
		if err := database.UpdateTagFrequency(tag, count, isCanonical); err != nil {
			continue
		}
	}
}

// ─── collectProjectTags ─────────────────────────────────────────────────────
// collectProjectTags extracts all unique tags from existing tasks.
func collectProjectTags(tasks []Todo2Task) []string {
	tagSet := make(map[string]bool)

	for _, task := range tasks {
		for _, tag := range task.Tags {
			// Skip auto-discovered tags
			if tag != "markdown" && tag != "discovered" {
				tagSet[tag] = true
			}
		}
	}

	tags := make([]string, 0, len(tagSet))
	for tag := range tagSet {
		tags = append(tags, tag)
	}

	sort.Strings(tags)

	return tags
}

// ─── projectTagsWithCacheHints ──────────────────────────────────────────────
// projectTagsWithCacheHints merges project tags from tasks with top tag frequencies from cache (for LLM hints everywhere).
func projectTagsWithCacheHints(tasks []Todo2Task, cacheLimit int) []string {
	tagSet := make(map[string]bool)
	for _, tag := range collectProjectTags(tasks) {
		tagSet[tag] = true
	}

	cacheTags, err := database.GetTopTagFrequencies(cacheLimit)
	if err == nil {
		for _, tag := range cacheTags {
			tagSet[tag] = true
		}
	}

	tags := make([]string, 0, len(tagSet))
	for tag := range tagSet {
		tags = append(tags, tag)
	}

	sort.Strings(tags)

	return tags
}

// ─── getCanonicalTagsList ───────────────────────────────────────────────────
// getCanonicalTagsList returns the list of canonical tag categories.
func getCanonicalTagsList() []string {
	return []string{
		"testing", "docs", "security", "build", "performance", "linting",
		"bug", "feature", "refactor", "migration", "config", "cli",
		"mcp", "llm", "database", "concurrency", "git", "planning",
		"batch", "phase", "epic", "workflow", "research", "analysis",
	}
}

// LLM timeout constants.
const (
	llmPerFileTimeout  = 10 * time.Second                             // Timeout for each file when not batching
	llmPerBatchTimeout = 45 * time.Second                             // Timeout per batch when batching (one call per batch)
	llmTotalTimeout    = 2 * time.Minute                              // Total timeout for all LLM processing
	discoverTagTimeout = 5 * time.Minute                              // Total timeout for discover_tags action
	llmTagBatchSize    = 25                                           // Default max tasks per LLM call for tag inference (tune via llm_batch_size; bigger = fewer calls, faster)
	llmTagMaxTokens    = 320                                          // Max tokens for tag-inference response (short JSON; lower = faster generation)
	ollamaTinyTagModel = "tinyllama"                                  // Small Ollama model for quick tag inference (use_tiny_tag_model)
	mlxTinyTagModel    = "mlx-community/TinyLlama-1.1B-Chat-v1.0-mlx" // Small MLX model for quick tag inference
)

// ─── effectiveLLMBatchSize ──────────────────────────────────────────────────
func effectiveLLMBatchSize(param int) int {
	if param > 0 {
		return param
	}

	return llmTagBatchSize
}

// ─── suggestNextLLMBatchSize ────────────────────────────────────────────────
// suggestNextLLMBatchSize returns a suggested llm_batch_size for the next run (feedback loop).
// If there were timeouts, suggest halving; if none, suggest trying a larger batch up to total files or 50.
func suggestNextLLMBatchSize(batchSizeUsed, llmTimeouts, totalDiscoveries int) int {
	const maxSuggested = 50

	if llmTimeouts > 0 && batchSizeUsed > llmTagBatchSize {
		next := batchSizeUsed / 2
		if next < llmTagBatchSize {
			return llmTagBatchSize
		}

		return next
	}

	if llmTimeouts == 0 && totalDiscoveries > 0 {
		// No timeouts: suggest trying larger to find ideal size
		next := batchSizeUsed + 10
		if next > totalDiscoveries {
			next = totalDiscoveries
		}

		if next > maxSuggested {
			next = maxSuggested
		}

		return next
	}

	return llmTagBatchSize
}

// ─── batchTimeout ───────────────────────────────────────────────────────────
// batchTimeout returns per-batch timeout; scales slightly with batch size for large batches.
func batchTimeout(batchSize int) time.Duration {
	base := llmPerBatchTimeout
	if batchSize <= llmTagBatchSize {
		return base
	}

	extra := time.Duration(batchSize-llmTagBatchSize) * 2 * time.Second
	if base+extra > 90*time.Second {
		return 90 * time.Second
	}

	return base + extra
}

// ─── enrichTagsWithLLM ──────────────────────────────────────────────────────
// enrichTagsWithLLM uses Apple FM to infer additional tags, in batches of files per call to speed up.
// Uses tag cache everywhere: project tags merged with top tag_frequency for LLM hints.
func enrichTagsWithLLM(ctx context.Context, fm FMProvider, discoveries []map[string]interface{}, tasks []Todo2Task, batchSizeParam int) []map[string]interface{} {
	batchSize := effectiveLLMBatchSize(batchSizeParam)
	canonicalTags := getCanonicalTagsList()

	projectTags := projectTagsWithCacheHints(tasks, 30)
	if len(projectTags) > 40 {
		projectTags = projectTags[:40]
	}

	totalCtx, totalCancel := context.WithTimeout(ctx, llmTotalTimeout)
	defer totalCancel()

	for start := 0; start < len(discoveries); start += batchSize {
		select {
		case <-totalCtx.Done():
			return discoveries
		default:
		}

		end := start + batchSize
		if end > len(discoveries) {
			end = len(discoveries)
		}

		batch := discoveries[start:end]

		prompt := buildBatchTagPrompt(batch, canonicalTags, projectTags)
		batchCtx, batchCancel := context.WithTimeout(totalCtx, batchTimeout(len(batch)))
		response, err := fm.Generate(batchCtx, prompt, 800, 0.2)

		batchCancel()

		if err != nil {
			if errors.Is(batchCtx.Err(), context.DeadlineExceeded) {
				for i := start; i < end; i++ {
					discoveries[i]["llm_timeout"] = true
				}
			}

			continue
		}

		suggestionsByFile := parseBatchTagResponse(response)

		for i, discovery := range batch {
			file, _ := discovery["file"].(string)
			existingTags, _ := discovery["tags"].([]string)

			suggestedTags := suggestionsByFile[file]
			if len(suggestedTags) == 0 {
				continue
			}

			if discoveries[start+i]["llm_suggestions"] == nil {
				discoveries[start+i]["llm_suggestions"] = suggestedTags
			}

			allTags := append(existingTags, suggestedTags...)
			seen := make(map[string]bool)
			uniqueTags := []string{}

			for _, tag := range allTags {
				if !seen[tag] {
					seen[tag] = true

					uniqueTags = append(uniqueTags, tag)
				}
			}

			discoveries[start+i]["tags"] = uniqueTags
		}
	}

	return discoveries
}

// buildBatchTagPrompt builds one concise prompt for a batch of files (smart prompt for speed).

// task_analysis_tags.go — Tag rules, NoiseTags, and tag-analysis action handlers.
// See also: task_analysis_tags_discover.go, task_analysis_tags_llm.go
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"
	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/proto"
	"github.com/davidl71/mcp-go-core/pkg/mcp/response"
)

// ─── Contents ───────────────────────────────────────────────────────────────
//   CanonicalTagRules
//   NoiseTags — NoiseTags returns tags that should be removed (status/meta tags with no filtering value).
//   handleTaskAnalysisTags — handleTaskAnalysisTags handles tag analysis and consolidation.
//   handleTaskAnalysisDiscoverTags — handleTaskAnalysisDiscoverTags discovers tags from markdown files and uses LLM for semantic inference.
// ────────────────────────────────────────────────────────────────────────────

// ─── CanonicalTagRules ──────────────────────────────────────────────────────
func CanonicalTagRules() map[string]string {
	return map[string]string{
		// Scorecard: testing
		"testing-validation": "testing",
		"test":               "testing",
		"tests":              "testing",
		"validation":         "testing",
		"integration":        "testing",

		// Scorecard: documentation
		"documentation": "docs",
		"doc":           "docs",
		"reporting":     "docs",
		"inventory":     "docs",

		// Scorecard: security
		"security-scan":   "security",
		"vulnerability":   "security",
		"vulnerabilities": "security",
		"audit":           "security",

		// Scorecard: build/CI
		"cicd":           "build",
		"ci":             "build",
		"cd":             "build",
		"ci-cd":          "build",
		"github-actions": "build",
		"automation":     "build",
		"uv":             "build",
		"setup":          "build",

		// Scorecard: performance
		"optimization": "performance",
		"profiling":    "performance",
		"pgo":          "performance",
		"gc":           "performance",
		"memory":       "performance",
		"caching":      "performance",
		"file-io":      "performance",
		"hot-reload":   "performance",
		"metrics":      "performance",

		// Scorecard: linting/code quality
		"spelling": "linting",
		"cspell":   "linting",

		// Type: bug
		"bug-fix":      "bug",
		"fix":          "bug",
		"bugfix":       "bug",
		"json-parsing": "bug",

		// Type: feature
		"enhancement":  "feature",
		"enhancements": "feature",

		// Type: refactor
		"refactoring":   "refactor",
		"code-quality":  "refactor",
		"cleanup":       "refactor",
		"consolidation": "refactor",
		"naming":        "refactor",

		// Domain: migration
		"bridge":         "migration",
		"legacy-code":    "migration",
		"python-cleanup": "migration",
		"native":         "migration",

		// Domain: config
		"configuration":        "config",
		"configuration-system": "config",
		"pyproject":            "config",

		// Domain: CLI
		"tui":        "cli",
		"bubble-tea": "cli",

		// Domain: MCP
		"mcp-config":             "mcp",
		"mcp-go-core":            "mcp",
		"mcp-go-core-extraction": "mcp",
		"go-sdk":                 "mcp",
		"framework-agnostic":     "mcp",
		"framework-improvements": "mcp",
		"todo-sync":              "mcp",
		"protobuf-integration":   "mcp",

		// Domain: LLM/AI
		"mlx":                     "llm",
		"apple-foundation-models": "llm",
		"apple-silicon":           "llm",
		"npu":                     "llm",
		"swift":                   "llm",
		"devwisdom-go":            "llm",
		"multi-agent":             "llm",

		// Domain: database
		"sqlite": "database",

		// Domain: concurrency
		"goroutines": "concurrency",

		// Domain: git
		"git-hooks": "git",

		// Domain: planning/workflow
		"strategy":       "planning",
		"coordination":   "planning",
		"task-breakdown": "planning",
		"crew-roles":     "planning",

		// Batch/phase normalization
		"batch-1": "batch",
		"batch-2": "batch",
		"batch1":  "batch",
		"batch2":  "batch",
		"batch3":  "batch",
		"phase-1": "phase",
		"phase-2": "phase",
		"phase-3": "phase",
		"phase-4": "phase",
		"phase-5": "phase",
		"phase-6": "phase",
	}
}

// ─── NoiseTags ──────────────────────────────────────────────────────────────
// NoiseTags returns tags that should be removed (status/meta tags with no filtering value).
func NoiseTags() []string {
	return []string{
		"complete",
		"blocking",
		"review",
		"development",
		"project-name",
		"catwalk",
	}
}

// ─── handleTaskAnalysisTags ─────────────────────────────────────────────────
// handleTaskAnalysisTags handles tag analysis and consolidation.
func handleTaskAnalysisTags(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	t0 := time.Now()

	store, err := getTaskStore(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get task store: %w", err)
	}

	list, err := store.ListTasks(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	tasks := tasksFromPtrs(list)
	profileLoadMs := time.Since(t0).Milliseconds()

	dryRun := true
	if run, ok := params["dry_run"].(bool); ok {
		dryRun = run
	}

	// Analyze tags
	t1 := time.Now()
	tagAnalysis := analyzeTags(tasks)
	profileKeywordMs := time.Since(t1).Milliseconds()

	// Optional batch: limit and/or prioritize tasks with no tags (tag_suggestions only)
	limit := 0
	if l, ok := params["limit"].(float64); ok && l > 0 {
		limit = int(l)
	} else if l, ok := params["limit"].(int); ok && l > 0 {
		limit = l
	}

	prioritizeUntagged := false
	if p, ok := params["prioritize_untagged"].(bool); ok {
		prioritizeUntagged = p
	}

	if (limit > 0 || prioritizeUntagged) && len(tagAnalysis.TagSuggestions) > 0 {
		taskByID := make(map[string]Todo2Task)
		for _, t := range tasks {
			taskByID[t.ID] = t
		}

		var order []string
		for taskID := range tagAnalysis.TagSuggestions {
			order = append(order, taskID)
		}

		if prioritizeUntagged {
			sort.Slice(order, func(i, j int) bool {
				untaggedI := len(taskByID[order[i]].Tags) == 0
				untaggedJ := len(taskByID[order[j]].Tags) == 0

				if untaggedI != untaggedJ {
					return untaggedI
				}

				return order[i] < order[j]
			})
		}

		if limit > 0 && len(order) > limit {
			order = order[:limit]
		}

		batch := make(map[string][]string)
		for _, taskID := range order {
			batch[taskID] = tagAnalysis.TagSuggestions[taskID]
		}

		tagAnalysis.TagSuggestions = batch
	}

	// Apply canonical rules if requested
	useCanonical := false
	if canonical, ok := params["use_canonical_rules"].(bool); ok {
		useCanonical = canonical
	}

	if useCanonical {
		tagAnalysis = applyTagRules(tasks, CanonicalTagRules(), tagAnalysis)
		// Also remove noise tags when using canonical rules
		tagAnalysis = removeTags(tasks, NoiseTags(), tagAnalysis)
	}

	// Apply custom rules if provided (in addition to canonical)
	customRulesJSON, _ := params["custom_rules"].(string)
	if customRulesJSON != "" {
		var rules map[string]string
		if err := json.Unmarshal([]byte(customRulesJSON), &rules); err == nil {
			tagAnalysis = applyTagRules(tasks, rules, tagAnalysis)
		}
	}

	// Remove additional tags if requested
	removeTagsJSON, _ := params["remove_tags"].(string)
	if removeTagsJSON != "" {
		var tagsToRemove []string
		if err := json.Unmarshal([]byte(removeTagsJSON), &tagsToRemove); err == nil {
			tagAnalysis = removeTags(tasks, tagsToRemove, tagAnalysis)
		}
	}

	// Optional: LLM semantic pass for quick tag addition from task title + content + existing tags
	useLLMSemantic := false
	if u, ok := params["use_llm_semantic"].(bool); ok {
		useLLMSemantic = u
	}

	matchExistingOnly := false
	if m, ok := params["match_existing_only"].(bool); ok {
		matchExistingOnly = m
	}

	useTinyTagModel := false
	if u, ok := params["use_tiny_tag_model"].(bool); ok {
		useTinyTagModel = u
	}

	llmSemanticMethod := ""
	llmSemanticProcessed := 0

	var llmProfile llmTagProfile

	if useLLMSemantic {
		llmBatchSize := 0
		if b, ok := params["llm_batch_size"].(float64); ok && b > 0 {
			llmBatchSize = int(b)
		} else if b, ok := params["llm_batch_size"].(int); ok && b > 0 {
			llmBatchSize = b
		}

		llmSemanticMethod, llmSemanticProcessed, llmProfile = enrichTaskTagSuggestionsWithLLM(ctx, tasks, &tagAnalysis, llmBatchSize, matchExistingOnly, useTinyTagModel)
	}

	// Apply changes if not dry run
	profileApplyMs := int64(0)

	if !dryRun {
		t2 := time.Now()

		tasks = applyTagChanges(tasks, tagAnalysis)
		for _, t := range tasks {
			taskPtr := &t
			if err := store.UpdateTask(ctx, taskPtr); err != nil {
				return nil, fmt.Errorf("failed to save task %s: %w", t.ID, err)
			}
		}
		// Update tag cache everywhere: frequency + task-level suggestions (for LLM hints)
		updateTagFrequencyCache(tasks)

		for i := range tasks {
			for _, tag := range tasks[i].Tags {
				_ = database.SaveTaskTagSuggestion(tasks[i].ID, tag, "tags", true)
			}
		}

		profileApplyMs = time.Since(t2).Milliseconds()
	}

	// Build profile summary (what's taking the most time)
	profileMinBatchMs, profileMaxBatchMs, profileAvgBatchMs := int64(0), int64(0), int64(0)
	if len(llmProfile.PerBatchMs) > 0 {
		profileMinBatchMs = llmProfile.PerBatchMs[0]
		for _, ms := range llmProfile.PerBatchMs {
			if ms < profileMinBatchMs {
				profileMinBatchMs = ms
			}

			if ms > profileMaxBatchMs {
				profileMaxBatchMs = ms
			}

			profileAvgBatchMs += ms
		}

		profileAvgBatchMs /= int64(len(llmProfile.PerBatchMs))
	}

	result := map[string]interface{}{
		"success":                true,
		"method":                 "native_go",
		"dry_run":                dryRun,
		"tag_analysis":           tagAnalysis,
		"total_tasks":            len(tasks),
		"recommendations":        buildTagRecommendations(tagAnalysis),
		"use_llm_semantic":       useLLMSemantic,
		"match_existing_only":    matchExistingOnly,
		"use_tiny_tag_model":     useTinyTagModel,
		"llm_semantic_method":    llmSemanticMethod,
		"llm_semantic_processed": llmSemanticProcessed,
		"profile_ms": map[string]interface{}{
			"load_tasks":       profileLoadMs,
			"keyword":          profileKeywordMs,
			"llm_total":        llmProfile.TotalMs,
			"llm_batches":      llmProfile.Batches,
			"llm_batch_min_ms": profileMinBatchMs,
			"llm_batch_max_ms": profileMaxBatchMs,
			"llm_batch_avg_ms": profileAvgBatchMs,
			"apply_and_cache":  profileApplyMs,
		},
	}

	outputPath, _ := params["output_path"].(string)
	if outputPath != "" {
		if err := saveAnalysisResult(outputPath, result); err != nil {
			return nil, fmt.Errorf("failed to save result: %w", err)
		}

		result["output_path"] = outputPath
	}

	resultJSON, _ := json.Marshal(result)
	resp := &proto.TaskAnalysisResponse{Action: "tags", OutputPath: outputPath, ResultJson: string(resultJSON)}

	return response.FormatResult(TaskAnalysisResponseToMap(resp), resp.GetOutputPath())
}

// ─── handleTaskAnalysisDiscoverTags ─────────────────────────────────────────
// handleTaskAnalysisDiscoverTags discovers tags from markdown files and uses LLM for semantic inference.
func handleTaskAnalysisDiscoverTags(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	startTime := time.Now()

	// Parse timeout parameter (default: 5 minutes)
	timeoutSec := 300 // 5 minutes default
	if t, ok := params["timeout_seconds"].(float64); ok && t > 0 {
		timeoutSec = int(t)
	} else if t, ok := params["timeout_seconds"].(int); ok && t > 0 {
		timeoutSec = t
	}

	// Create timeout context for entire operation
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSec)*time.Second)
	defer cancel()

	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	store, err := getTaskStore(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get task store: %w", err)
	}

	list, err := store.ListTasks(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	tasks := tasksFromPtrs(list)

	dryRun := true
	if run, ok := params["dry_run"].(bool); ok {
		dryRun = run
	}

	useLLM := true
	if llm, ok := params["use_llm"].(bool); ok {
		useLLM = llm
	}

	useCache := true
	if cache, ok := params["use_cache"].(bool); ok {
		useCache = cache
	}

	clearCache := false
	if clear, ok := params["clear_cache"].(bool); ok {
		clearCache = clear
	}

	docPath := "docs"
	if path, ok := params["doc_path"].(string); ok && path != "" {
		docPath = path
	}

	llmBatchSize := 0
	if b, ok := params["llm_batch_size"].(float64); ok && b > 0 {
		llmBatchSize = int(b)
	} else if b, ok := params["llm_batch_size"].(int); ok && b > 0 {
		llmBatchSize = b
	}

	// Clear cache if requested
	cacheCleared := false

	if clearCache {
		if err := database.ClearDiscoveredTagsCache(); err == nil {
			cacheCleared = true
		}
	}

	// Check timeout before expensive operations
	select {
	case <-timeoutCtx.Done():
		return nil, fmt.Errorf("operation timed out after %ds", timeoutSec)
	default:
	}

	// Discover tags from markdown files (with caching)
	discoveries, cacheHits, cacheMisses := discoverTagsFromMarkdownWithCache(projectRoot, docPath, useCache)

	// Track LLM stats
	llmTimeouts := 0
	llmProcessed := 0

	// Use LLM for semantic tag inference if available and requested
	llmMethod := "none"

	if useLLM {
		fm := DefaultFMProvider()
		if fm != nil && fm.Supported() {
			llmMethod = "apple_fm"
			discoveries = enrichTagsWithLLM(timeoutCtx, fm, discoveries, tasks, llmBatchSize)
		} else {
			ollama := DefaultOllama()
			if ollama != nil {
				llmMethod = "ollama"
				discoveries = enrichTagsWithOllama(timeoutCtx, ollama, discoveries, tasks, llmBatchSize)
			}
		}
	}

	// Count LLM stats from discoveries
	for _, d := range discoveries {
		if _, ok := d["llm_suggestions"]; ok {
			llmProcessed++
		}

		if timedOut, ok := d["llm_timeout"].(bool); ok && timedOut {
			llmTimeouts++
		}
	}

	// Save discoveries to cache (if cache is enabled and we have new discoveries)
	if useCache && cacheMisses > 0 {
		saveDiscoveriesToCache(discoveries)
	}

	// Match discovered tags to tasks
	tagMatches := matchDiscoveredTagsToTasks(discoveries, tasks)

	// Optional batch: limit and/or prioritize tasks with no tags
	limit := 0
	if l, ok := params["limit"].(float64); ok && l > 0 {
		limit = int(l)
	} else if l, ok := params["limit"].(int); ok && l > 0 {
		limit = l
	}

	prioritizeUntagged := false
	if p, ok := params["prioritize_untagged"].(bool); ok {
		prioritizeUntagged = p
	}

	if (limit > 0 || prioritizeUntagged) && len(tagMatches) > 0 {
		taskByID := make(map[string]Todo2Task)
		for _, t := range tasks {
			taskByID[t.ID] = t
		}

		var order []string
		for taskID := range tagMatches {
			order = append(order, taskID)
		}

		if prioritizeUntagged {
			sort.Slice(order, func(i, j int) bool {
				untaggedI := len(taskByID[order[i]].Tags) == 0
				untaggedJ := len(taskByID[order[j]].Tags) == 0

				if untaggedI != untaggedJ {
					return untaggedI // untagged first
				}

				return order[i] < order[j]
			})
		}

		if limit > 0 && len(order) > limit {
			order = order[:limit]
		}

		batch := make(map[string][]string)
		for _, taskID := range order {
			batch[taskID] = tagMatches[taskID]
		}

		tagMatches = batch
	}

	// Optional: only backlog (Todo + In Progress) — parse todo2 backlog and update tags for those only
	backlogOnly := false
	if b, ok := params["backlog_only"].(bool); ok {
		backlogOnly = b
	}

	if backlogOnly && len(tagMatches) > 0 {
		taskByID := make(map[string]Todo2Task)
		for _, t := range tasks {
			taskByID[t.ID] = t
		}

		backlogMatches := make(map[string][]string)

		for taskID, tags := range tagMatches {
			if t, ok := taskByID[taskID]; ok && IsBacklogStatus(t.Status) {
				backlogMatches[taskID] = tags
			}
		}

		tagMatches = backlogMatches
	}

	// Save file-task tag matches to cache
	if useCache && !dryRun {
		saveFileTaskTagsToCache(tagMatches, discoveries)
	}

	// Apply tags if not dry run
	appliedCount := 0

	if !dryRun && len(tagMatches) > 0 {
		for taskID, newTags := range tagMatches {
			for i := range tasks {
				if tasks[i].ID == taskID {
					// Merge tags (avoid duplicates)
					existingTags := make(map[string]bool)
					for _, t := range tasks[i].Tags {
						existingTags[t] = true
					}

					for _, newTag := range newTags {
						if !existingTags[newTag] {
							tasks[i].Tags = append(tasks[i].Tags, newTag)
							appliedCount++
						}
					}

					break
				}
			}
		}

		if appliedCount > 0 {
			for _, t := range tasks {
				taskPtr := &t
				if err := store.UpdateTask(ctx, taskPtr); err != nil {
					return nil, fmt.Errorf("failed to save task %s: %w", t.ID, err)
				}
			}
		}

		// Update tag frequency cache
		updateTagFrequencyCache(tasks)
	}

	elapsedMs := time.Since(startTime).Milliseconds()
	effectiveBatch := effectiveLLMBatchSize(llmBatchSize)
	suggestion := effectiveBatch

	if useLLM {
		suggestion = suggestNextLLMBatchSize(effectiveBatch, llmTimeouts, len(discoveries))
	}

	result := map[string]interface{}{
		"cache_hits":           cacheHits,
		"cache_misses":         cacheMisses,
		"cache_cleared":        cacheCleared,
		"success":              true,
		"method":               "native_go",
		"llm_method":           llmMethod,
		"llm_processed":        llmProcessed,
		"llm_timeouts":         llmTimeouts,
		"llm_batch_size":       effectiveBatch,
		"llm_batch_suggestion": suggestion,
		"timeout_seconds":      timeoutSec,
		"elapsed_ms":           elapsedMs,
		"dry_run":              dryRun,
		"discoveries":          discoveries,
		"tag_matches":          tagMatches,
		"total_discoveries":    len(discoveries),
		"total_matches":        len(tagMatches),
		"applied_count":        appliedCount,
	}

	resultJSON, _ := json.Marshal(result)
	resp := &proto.TaskAnalysisResponse{Action: "discover_tags", ResultJson: string(resultJSON)}

	return response.FormatResult(TaskAnalysisResponseToMap(resp), resp.GetOutputPath())
}

// discoverTagsFromMarkdown scans markdown files for tag patterns.

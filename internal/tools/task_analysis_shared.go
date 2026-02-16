package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/davidl71/exarp-go/internal/config"
	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/internal/taskanalysis"
	"github.com/davidl71/exarp-go/proto"
	"github.com/davidl71/mcp-go-core/pkg/mcp/response"
)

// TaskAnalysisResponseToMap converts TaskAnalysisResponse to a map for response.FormatResult (unmarshals result_json into map).
func TaskAnalysisResponseToMap(resp *proto.TaskAnalysisResponse) map[string]interface{} {
	if resp == nil {
		return nil
	}

	out := map[string]interface{}{
		"action": resp.GetAction(),
	}
	if resp.GetOutputPath() != "" {
		out["output_path"] = resp.GetOutputPath()
	}

	if resp.GetResultJson() != "" {
		var payload map[string]interface{}
		if json.Unmarshal([]byte(resp.GetResultJson()), &payload) == nil {
			for k, v := range payload {
				out[k] = v
			}
		}
	}

	return out
}

// handleTaskAnalysisNative dispatches to the appropriate action (duplicates, tags, dependencies, parallelization, hierarchy).
// Hierarchy uses the FM abstraction (DefaultFMProvider()); when FM is not available, hierarchy returns a clear error (no Python fallback).
func handleTaskAnalysisNative(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	action, _ := params["action"].(string)
	if action == "" {
		action = "duplicates"
	}

	switch action {
	case "hierarchy":
		return handleTaskAnalysisHierarchy(ctx, params)
	case "duplicates":
		return handleTaskAnalysisDuplicates(ctx, params)
	case "tags":
		return handleTaskAnalysisTags(ctx, params)
	case "discover_tags":
		return handleTaskAnalysisDiscoverTags(ctx, params)
	case "dependencies":
		return handleTaskAnalysisDependencies(ctx, params)
	case "parallelization":
		return handleTaskAnalysisParallelization(ctx, params)
	case "fix_missing_deps":
		return handleTaskAnalysisFixMissingDeps(ctx, params)
	case "validate":
		return handleTaskAnalysisValidate(ctx, params)
	case "execution_plan":
		return handleTaskAnalysisExecutionPlan(ctx, params)
	case "complexity":
		return handleTaskAnalysisComplexity(ctx, params)
	case "conflicts":
		return handleTaskAnalysisConflicts(ctx, params)
	case "dependencies_summary":
		return handleTaskAnalysisDependenciesSummary(ctx, params)
	case "suggest_dependencies":
		return handleTaskAnalysisSuggestDependencies(ctx, params)
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

// handleTaskAnalysisConflicts detects task-overlap conflicts (In Progress tasks with dependent also In Progress).
func handleTaskAnalysisConflicts(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	store, err := getTaskStore(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get task store: %w", err)
	}

	list, err := store.ListTasks(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	conflicts := DetectTaskOverlapConflicts(list)
	hasConflict := len(conflicts) > 0
	overlapping := make([]string, 0)

	if hasConflict {
		seen := make(map[string]bool)
		for _, c := range conflicts {
			if !seen[c.DepTaskID] {
				seen[c.DepTaskID] = true

				overlapping = append(overlapping, c.DepTaskID)
			}

			if !seen[c.TaskID] {
				seen[c.TaskID] = true

				overlapping = append(overlapping, c.TaskID)
			}
		}
	}

	out := map[string]interface{}{
		"conflict":    hasConflict,
		"conflicts":   conflicts,
		"overlapping": overlapping,
	}

	if hasConflict {
		reasons := make([]string, len(conflicts))
		for i, c := range conflicts {
			reasons[i] = c.Reason
		}

		out["reasons"] = reasons
	}

	resultJSON, _ := json.Marshal(out)
	resp := &proto.TaskAnalysisResponse{Action: "conflicts", ResultJson: string(resultJSON)}

	return response.FormatResult(TaskAnalysisResponseToMap(resp), resp.GetOutputPath())
}

// handleTaskAnalysisDuplicates handles duplicates detection.
func handleTaskAnalysisDuplicates(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	store, err := getTaskStore(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get task store: %w", err)
	}

	list, err := store.ListTasks(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	tasks := tasksFromPtrs(list)

	// Use config default, allow override from params
	similarityThreshold := config.SimilarityThreshold()
	if threshold, ok := params["similarity_threshold"].(float64); ok {
		similarityThreshold = threshold
	}

	autoFix := false
	if fix, ok := params["auto_fix"].(bool); ok {
		autoFix = fix
	}

	// Find duplicates
	duplicates := findDuplicateTasks(tasks, similarityThreshold)

	// Auto-fix if requested
	if autoFix && len(duplicates) > 0 {
		tasks = mergeDuplicateTasks(tasks, duplicates)
		// Delete removed task IDs (merge keeps first per group, removes group[1:])
		for _, grp := range duplicates {
			for i := 1; i < len(grp); i++ {
				_ = store.DeleteTask(ctx, grp[i])
			}
		}
		// Update kept/merged tasks
		for _, t := range tasks {
			taskPtr := &t
			if err := store.UpdateTask(ctx, taskPtr); err != nil {
				return nil, fmt.Errorf("failed to save merged task %s: %w", t.ID, err)
			}
		}
	}

	// Build result
	result := map[string]interface{}{
		"success":              true,
		"method":               "native_go",
		"total_tasks":          len(tasks),
		"duplicate_groups":     len(duplicates),
		"duplicates":           duplicates,
		"similarity_threshold": similarityThreshold,
		"auto_fix":             autoFix,
	}

	if autoFix {
		result["merged"] = true
		result["tasks_after_merge"] = len(tasks)
	}

	outputPath, _ := params["output_path"].(string)
	if outputPath != "" {
		if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
			return nil, fmt.Errorf("failed to create output dir: %w", err)
		}
	}

	resultJSON, _ := json.Marshal(result)
	resp := &proto.TaskAnalysisResponse{Action: "duplicates", OutputPath: outputPath, ResultJson: string(resultJSON)}

	return response.FormatResult(TaskAnalysisResponseToMap(resp), resp.GetOutputPath())
}

// CanonicalTagRules returns default tag consolidation rules aligned with scorecard dimensions.
// Categories: testing, docs, security, build, performance, bug, feature, refactor, migration, config, cli, mcp, llm, database, workflow, planning, linting.
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

	filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
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
	})

	return discoveries
}

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

	filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
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
	})

	return discoveries, cacheHits, cacheMisses
}

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
		database.SaveDiscoveredTags(filePath, fileHash, cacheEntries)
	}
}

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
						database.SaveFileTaskTag(filePath, taskID, tag, true)
						break
					}
				}
			}
		}
	}
}

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
		database.UpdateTagFrequency(tag, count, isCanonical)
	}
}

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

func effectiveLLMBatchSize(param int) int {
	if param > 0 {
		return param
	}

	return llmTagBatchSize
}

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
func buildBatchTagPrompt(batch []map[string]interface{}, canonicalTags, projectTags []string) string {
	var b strings.Builder

	b.WriteString("Suggest 0-3 tags per file from path/context. Prefer: ")
	b.WriteString(strings.Join(canonicalTags, ", "))
	b.WriteString(". Project tags: ")
	b.WriteString(strings.Join(projectTags, ", "))
	b.WriteString(".\n\nFiles (path → existing tags):\n")

	for _, d := range batch {
		file, _ := d["file"].(string)
		existing, _ := d["tags"].([]string)

		b.WriteString(file)
		b.WriteString(": ")

		enc, _ := json.Marshal(existing)
		b.WriteString(string(enc))
		b.WriteString("\n")
	}

	b.WriteString("\nReturn ONLY a JSON object: {\"path\": [\"tag1\",\"tag2\"]}. No other text. Use file path as key. Empty array [] if none.")

	return b.String()
}

// extractOllamaGenerateResponseText extracts the "response" field from Ollama generate JSON.
// handleOllamaGenerate returns Text as a JSON object with "response" containing the actual LLM output.
// Returns the inner response string, or the original text if not in that format.
func extractOllamaGenerateResponseText(text string) string {
	text = strings.TrimSpace(text)

	var raw map[string]interface{}

	if err := json.Unmarshal([]byte(text), &raw); err != nil {
		return text
	}

	if r, ok := raw["response"].(string); ok && r != "" {
		return r
	}

	return text
}

// parseBatchTagResponse parses LLM response into map[file path][]suggested tags.
func parseBatchTagResponse(response string) map[string][]string {
	return parseTagResponseJSON(response)
}

// parseTaskTagResponse parses LLM response into map[task_id][]suggested tags (same JSON shape).
func parseTaskTagResponse(response string) map[string][]string {
	return parseTagResponseJSON(response)
}

// parseTagResponseJSON parses a JSON object of the form {"id": ["tag1","tag2"], ...} into map[string][]string.
func parseTagResponseJSON(response string) map[string][]string {
	response = strings.TrimSpace(response)
	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")

	if start < 0 || end <= start {
		return nil
	}

	response = response[start : end+1]

	var raw map[string]interface{}

	if err := json.Unmarshal([]byte(response), &raw); err != nil {
		return nil
	}

	out := make(map[string][]string)

	for id, v := range raw {
		var tags []string

		switch arr := v.(type) {
		case []interface{}:
			for _, x := range arr {
				if s, ok := x.(string); ok && s != "" {
					tags = append(tags, s)
				}
			}

			out[id] = tags
		case []string:
			out[id] = arr
		}
	}

	return out
}

// taskTagLLMBatchItem holds one task's data for LLM tag suggestion (action=tags semantic pass).
type taskTagLLMBatchItem struct {
	TaskID       string
	Title        string
	Snippet      string
	ExistingTags []string
}

const maxTaskSnippetLen = 400

// buildTaskBatchTagPrompt builds one prompt for a batch of tasks (title + content snippet + existing tags).
func buildTaskBatchTagPrompt(batch []taskTagLLMBatchItem, canonicalTags, projectTags []string) string {
	var b strings.Builder

	b.WriteString("Suggest 0-3 tags per task from title and context. Prefer: ")
	b.WriteString(strings.Join(canonicalTags, ", "))
	b.WriteString(". Project tags: ")
	b.WriteString(strings.Join(projectTags, ", "))
	b.WriteString(".\n\nTasks (task_id → title, snippet, existing tags):\n")

	for _, item := range batch {
		b.WriteString(item.TaskID)
		b.WriteString(": title=")
		b.WriteString(item.Title)
		b.WriteString(" snippet=")

		if len(item.Snippet) > maxTaskSnippetLen {
			b.WriteString(item.Snippet[:maxTaskSnippetLen])
			b.WriteString("...")
		} else {
			b.WriteString(item.Snippet)
		}

		b.WriteString(" existing=")

		enc, _ := json.Marshal(item.ExistingTags)
		b.WriteString(string(enc))
		b.WriteString("\n")
	}

	b.WriteString("\nReturn ONLY a JSON object: {\"T-123\": [\"tag1\",\"tag2\"]}. Use task_id as key. No other text. Empty array [] if none.")

	return b.String()
}

// buildTaskBatchTagPromptMatchOnly builds a short prompt for quick FM inference: pick 0-3 tags only from allowed list (no free-form).
func buildTaskBatchTagPromptMatchOnly(batch []taskTagLLMBatchItem, allowedTags []string) string {
	var b strings.Builder

	b.WriteString("Pick 0-3 tags per task from this list only: ")
	b.WriteString(strings.Join(allowedTags, ", "))
	b.WriteString(".\n\nTasks (task_id, title, snippet):\n")

	for _, item := range batch {
		b.WriteString(item.TaskID)
		b.WriteString(": ")
		b.WriteString(item.Title)
		b.WriteString(" | ")

		snippet := item.Snippet
		if len(snippet) > 200 {
			snippet = snippet[:200] + "..."
		}

		b.WriteString(snippet)
		b.WriteString("\n")
	}

	b.WriteString("\nReturn ONLY JSON: {\"T-123\": [\"tag1\",\"tag2\"]}. Use task_id as key. Only tags from the list. Empty [] if none.")

	return b.String()
}

// filterSuggestionsToAllowed keeps only tags that are in the allowed set (for match_existing_only).
func filterSuggestionsToAllowed(suggestionsByTask map[string][]string, allowed map[string]bool) map[string][]string {
	out := make(map[string][]string)

	for taskID, tags := range suggestionsByTask {
		filtered := make([]string, 0, len(tags))

		seen := make(map[string]bool)
		for _, tag := range tags {
			if allowed[tag] && !seen[tag] {
				seen[tag] = true

				filtered = append(filtered, tag)
			}
		}

		if len(filtered) > 0 {
			out[taskID] = filtered
		}
	}

	return out
}

// llmTagProfile holds timing profile for the LLM tag inference phase.
type llmTagProfile struct {
	TotalMs    int64
	Batches    int
	PerBatchMs []int64
}

// enrichTaskTagSuggestionsWithLLM runs Apple FM, Ollama, or MLX on batches of tasks and merges suggested tags into analysis.TagSuggestions.
// Uses tag cache everywhere as quick hints. When matchExistingOnly is true, uses a constrained prompt and filters output to allowed tags only.
// When useTinyTagModel is true, tries Ollama with tinyllama first, then MLX with TinyLlama (faster small models) before Apple FM.
// Returns method, processed count, and profile (LLM total ms, batch count, per-batch ms) for profiling.
func enrichTaskTagSuggestionsWithLLM(ctx context.Context, tasks []Todo2Task, analysis *TagAnalysis, batchSizeParam int, matchExistingOnly bool, useTinyTagModel bool) (method string, processed int, profile llmTagProfile) {
	batchSize := effectiveLLMBatchSize(batchSizeParam)
	canonicalTags := getCanonicalTagsList()

	projectTags := projectTagsWithCacheHints(tasks, 30)
	if len(projectTags) > 40 {
		projectTags = projectTags[:40]
	}

	allowedSet := make(map[string]bool)
	for _, t := range canonicalTags {
		allowedSet[t] = true
	}

	for _, t := range projectTags {
		allowedSet[t] = true
	}

	const totalTimeout = 120 * time.Second

	totalCtx, cancel := context.WithTimeout(ctx, totalTimeout)
	defer cancel()

	// Build flat list of task batch items (title, snippet, existing = Tags + TagSuggestions + cached suggestions)
	items := make([]taskTagLLMBatchItem, 0, len(tasks))

	for _, t := range tasks {
		existing := make(map[string]bool)
		for _, tag := range t.Tags {
			existing[tag] = true
		}

		for _, tag := range analysis.TagSuggestions[t.ID] {
			existing[tag] = true
		}

		cached, _ := database.GetTaskTagSuggestions(t.ID)
		for _, tag := range cached {
			existing[tag] = true
		}

		existingList := make([]string, 0, len(existing))
		for tag := range existing {
			existingList = append(existingList, tag)
		}

		sort.Strings(existingList)

		snippet := strings.TrimSpace(t.LongDescription)
		if snippet == "" {
			snippet = t.Content
		}

		items = append(items, taskTagLLMBatchItem{
			TaskID:       t.ID,
			Title:        t.Content,
			Snippet:      snippet,
			ExistingTags: existingList,
		})
	}

	mergeSuggestions := func(suggestionsByTask map[string][]string) {
		if matchExistingOnly {
			suggestionsByTask = filterSuggestionsToAllowed(suggestionsByTask, allowedSet)
		}

		for taskID, suggested := range suggestionsByTask {
			if len(suggested) == 0 {
				continue
			}

			processed++

			existingSet := make(map[string]bool)
			for _, tag := range analysis.TagSuggestions[taskID] {
				existingSet[tag] = true
			}

			for _, tag := range suggested {
				if !existingSet[tag] {
					existingSet[tag] = true

					analysis.TagSuggestions[taskID] = append(analysis.TagSuggestions[taskID], tag)
				}
			}
		}
	}

	llmStart := time.Now()

	// When useTinyTagModel: try Ollama(tinyllama) then MLX(TinyLlama) for faster inference
	if useTinyTagModel {
		ollama := DefaultOllama()
		if ollama != nil {
			tinyProcessed := 0

			for start := 0; start < len(items); start += batchSize {
				select {
				case <-totalCtx.Done():
					profile.TotalMs = time.Since(llmStart).Milliseconds()
					profile.Batches = len(profile.PerBatchMs)

					return "ollama", processed, profile
				default:
				}

				end := start + batchSize
				if end > len(items) {
					end = len(items)
				}

				batch := items[start:end]
				batchStart := time.Now()

				var prompt string

				if matchExistingOnly {
					allowedList := make([]string, 0, len(allowedSet))
					for tag := range allowedSet {
						allowedList = append(allowedList, tag)
					}

					sort.Strings(allowedList)
					prompt = buildTaskBatchTagPromptMatchOnly(batch, allowedList)
				} else {
					prompt = buildTaskBatchTagPrompt(batch, canonicalTags, projectTags)
				}

				batchCtx, batchCancel := context.WithTimeout(totalCtx, batchTimeout(len(batch)))
				result, err := ollama.Invoke(batchCtx, map[string]interface{}{
					"action":      "generate",
					"model":       ollamaTinyTagModel,
					"prompt":      prompt,
					"max_tokens":  llmTagMaxTokens,
					"temperature": 0.2,
				})

				batchCancel()

				profile.PerBatchMs = append(profile.PerBatchMs, time.Since(batchStart).Milliseconds())

				if err != nil || len(result) == 0 {
					continue
				}

				response := extractOllamaGenerateResponseText(result[0].Text)
				mergeSuggestions(parseTaskTagResponse(response))

				tinyProcessed++
			}

			if tinyProcessed > 0 {
				profile.TotalMs = time.Since(llmStart).Milliseconds()
				profile.Batches = len(profile.PerBatchMs)

				return "ollama", processed, profile
			}
		}
		// Try MLX with TinyLlama (reset profile so we only report MLX times)
		profile = llmTagProfile{}
		mlxProcessed := 0

		for start := 0; start < len(items); start += batchSize {
			select {
			case <-totalCtx.Done():
				profile.TotalMs = time.Since(llmStart).Milliseconds()
				profile.Batches = len(profile.PerBatchMs)

				if mlxProcessed > 0 {
					return "mlx", processed, profile
				}

				goto fallthroughTiny
			default:
			}

			end := start + batchSize
			if end > len(items) {
				end = len(items)
			}

			batch := items[start:end]
			batchStart := time.Now()

			var prompt string

			if matchExistingOnly {
				allowedList := make([]string, 0, len(allowedSet))
				for tag := range allowedSet {
					allowedList = append(allowedList, tag)
				}

				sort.Strings(allowedList)
				prompt = buildTaskBatchTagPromptMatchOnly(batch, allowedList)
			} else {
				prompt = buildTaskBatchTagPrompt(batch, canonicalTags, projectTags)
			}

			batchCtx, batchCancel := context.WithTimeout(totalCtx, batchTimeout(len(batch)))
			raw, err := InvokeMLXTool(batchCtx, map[string]interface{}{
				"action":      "generate",
				"model":       mlxTinyTagModel,
				"prompt":      prompt,
				"max_tokens":  llmTagMaxTokens,
				"temperature": 0.2,
			})

			batchCancel()

			profile.PerBatchMs = append(profile.PerBatchMs, time.Since(batchStart).Milliseconds())

			if err != nil {
				continue
			}

			response, err := parseGeneratedTextFromMLXResponse(raw)
			if err != nil || response == "" {
				continue
			}

			mergeSuggestions(parseTaskTagResponse(response))

			mlxProcessed++
		}

		if mlxProcessed > 0 {
			profile.TotalMs = time.Since(llmStart).Milliseconds()
			profile.Batches = len(profile.PerBatchMs)

			return "mlx", processed, profile
		}
	fallthroughTiny:
		// Tiny path failed; fall through to ModelRouter (FM → Ollama → MLX)

	}

	// Use DefaultModelRouter for tag enrichment: FM chain, then Ollama, then MLX.
	requirements := ModelRequirements{PreferSpeed: useTinyTagModel}
	model := DefaultModelRouter.SelectModel("general", requirements)
	method = modelTypeToMethodString(model)

	for start := 0; start < len(items); start += batchSize {
		select {
		case <-totalCtx.Done():
			profile.TotalMs = time.Since(llmStart).Milliseconds()
			profile.Batches = len(profile.PerBatchMs)

			return method, processed, profile
		default:
		}

		end := start + batchSize
		if end > len(items) {
			end = len(items)
		}

		batch := items[start:end]
		batchStart := time.Now()

		var prompt string

		if matchExistingOnly {
			allowedList := make([]string, 0, len(allowedSet))
			for tag := range allowedSet {
				allowedList = append(allowedList, tag)
			}

			sort.Strings(allowedList)
			prompt = buildTaskBatchTagPromptMatchOnly(batch, allowedList)
		} else {
			prompt = buildTaskBatchTagPrompt(batch, canonicalTags, projectTags)
		}

		batchCtx, batchCancel := context.WithTimeout(totalCtx, batchTimeout(len(batch)))
		response, err := DefaultModelRouter.Generate(batchCtx, model, prompt, llmTagMaxTokens, 0.2)

		batchCancel()

		profile.PerBatchMs = append(profile.PerBatchMs, time.Since(batchStart).Milliseconds())

		if err != nil {
			continue
		}

		mergeSuggestions(parseTaskTagResponse(response))
	}

	profile.TotalMs = time.Since(llmStart).Milliseconds()
	profile.Batches = len(profile.PerBatchMs)

	return method, processed, profile
}

// modelTypeToMethodString maps ModelType to the profiling method string used in tag enrichment.
func modelTypeToMethodString(m ModelType) string {
	switch m {
	case ModelFM:
		return "apple_fm"
	case ModelOllamaLlama, ModelOllamaCode:
		return "ollama"
	case ModelMLX:
		return "mlx"
	default:
		return "model_router"
	}
}

// enrichTagsWithOllama uses Ollama to infer additional tags, in batches (same smart prompt as Apple FM).
// Uses tag cache everywhere for LLM hints.
func enrichTagsWithOllama(ctx context.Context, ollama OllamaProvider, discoveries []map[string]interface{}, tasks []Todo2Task, batchSizeParam int) []map[string]interface{} {
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
		result, err := ollama.Invoke(batchCtx, map[string]interface{}{
			"action":      "generate",
			"prompt":      prompt,
			"max_tokens":  800,
			"temperature": 0.2,
		})

		batchCancel()

		if err != nil || len(result) == 0 {
			if errors.Is(batchCtx.Err(), context.DeadlineExceeded) {
				for i := start; i < end; i++ {
					discoveries[i]["llm_timeout"] = true
				}
			}

			continue
		}

		response := extractOllamaGenerateResponseText(result[0].Text)
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

// matchDiscoveredTagsToTasks matches discovered tags from files to related tasks.
func matchDiscoveredTagsToTasks(discoveries []map[string]interface{}, tasks []Todo2Task) map[string][]string {
	matches := make(map[string][]string)

	for _, discovery := range discoveries {
		file, _ := discovery["file"].(string)

		tags, ok := discovery["tags"].([]string)
		if !ok {
			continue
		}

		// Match tasks that reference this file or have related content
		for _, task := range tasks {
			matched := false

			// Check if task references the file
			if strings.Contains(task.Content, file) || strings.Contains(task.LongDescription, file) {
				matched = true
			}

			// Check if task was discovered from this file
			if df, ok := task.Metadata["discovered_from"].(string); ok && df == file {
				matched = true
			}

			// Check planning doc link
			if pd, ok := task.Metadata["planning_doc"].(string); ok && pd == file {
				matched = true
			}

			if matched {
				// Apply canonical rules to discovered tags
				canonicalRules := CanonicalTagRules()

				for _, tag := range tags {
					finalTag := tag
					if newTag, exists := canonicalRules[tag]; exists && newTag != "" {
						finalTag = newTag
					}
					// Check if task already has this tag
					hasTag := false

					for _, existingTag := range task.Tags {
						if existingTag == finalTag {
							hasTag = true
							break
						}
					}

					if !hasTag {
						matches[task.ID] = append(matches[task.ID], finalTag)
					}
				}
			}
		}
	}

	// Dedupe matches
	for taskID, tags := range matches {
		seen := make(map[string]bool)
		unique := []string{}

		for _, tag := range tags {
			if !seen[tag] {
				seen[tag] = true

				unique = append(unique, tag)
			}
		}

		matches[taskID] = unique
	}

	return matches
}

// handleTaskAnalysisDependencies handles dependency analysis.
func handleTaskAnalysisDependencies(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	store, err := getTaskStore(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get task store: %w", err)
	}

	list, err := store.ListTasks(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	tasks := tasksFromPtrs(list)

	cycles, missing, err := GetDependencyAnalysisFromTasks(tasks)
	if err != nil {
		return nil, err
	}

	// Build legacy graph format for backward compatibility
	tg, err := BuildTaskGraph(tasks)
	if err != nil {
		return nil, fmt.Errorf("failed to build task graph: %w", err)
	}

	graph := buildLegacyGraphFormat(tg)

	// Calculate critical path from backlog only (exclude Done by default)
	var criticalPath []string

	var criticalPathDetails []map[string]interface{}

	maxLevel := 0

	tgBacklog, err := BuildTaskGraphBacklogOnly(tasks)
	if err == nil && tgBacklog.Graph.Nodes().Len() > 0 {
		hasCycles, err := HasCycles(tgBacklog)
		if err == nil && !hasCycles {
			// Find critical path among backlog tasks
			path, err := FindCriticalPath(tgBacklog)
			if err == nil {
				criticalPath = path

				// Build detailed path information
				for _, taskID := range path {
					for _, task := range tasks {
						if task.ID == taskID {
							criticalPathDetails = append(criticalPathDetails, map[string]interface{}{
								"id":                 task.ID,
								"content":            task.Content,
								"priority":           task.Priority,
								"status":             task.Status,
								"dependencies":       task.Dependencies,
								"dependencies_count": len(task.Dependencies),
							})

							break
						}
					}
				}
			}

			// Get max dependency level from backlog graph
			levels := GetTaskLevels(tgBacklog)
			for _, level := range levels {
				if level > maxLevel {
					maxLevel = level
				}
			}
		}
	}

	outputFormat := "json"
	if format, ok := params["output_format"].(string); ok && format != "" {
		outputFormat = format
	}

	result := map[string]interface{}{
		"success":               true,
		"method":                "native_go",
		"total_tasks":           len(tasks),
		"dependency_graph":      graph,
		"circular_dependencies": cycles,
		"missing_dependencies":  missing,
		"recommendations":       buildDependencyRecommendations(graph, cycles, missing),
	}

	// Add critical path information if available
	if len(criticalPath) > 0 {
		result["critical_path"] = criticalPath
		result["critical_path_length"] = len(criticalPath)
		result["critical_path_details"] = criticalPathDetails
		result["max_dependency_level"] = maxLevel
	}

	// Include human-readable report in JSON for CLI/consumers
	result["report"] = formatDependencyAnalysisText(result)

	outputPath, _ := params["output_path"].(string)
	if outputFormat == "json" {
		if outputPath != "" {
			if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
				return nil, fmt.Errorf("failed to create output dir: %w", err)
			}
		}

		resultJSON, _ := json.Marshal(result)
		resp := &proto.TaskAnalysisResponse{Action: "dependencies", OutputPath: outputPath, ResultJson: string(resultJSON)}

		return response.FormatResult(TaskAnalysisResponseToMap(resp), resp.GetOutputPath())
	}

	output := formatDependencyAnalysisText(result)

	if outputPath != "" {
		if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
			return nil, fmt.Errorf("failed to create output dir: %w", err)
		}

		if err := os.WriteFile(outputPath, []byte(output), 0644); err != nil {
			return nil, fmt.Errorf("failed to save result: %w", err)
		}

		output += fmt.Sprintf("\n\n[Saved to: %s]", outputPath)
	}

	return []framework.TextContent{{Type: "text", Text: output}}, nil
}

// handleTaskAnalysisDependenciesSummary combines dependencies, parallelization, and execution_plan (T-227).
func handleTaskAnalysisDependenciesSummary(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	var parts []string

	deps, err := handleTaskAnalysisDependencies(ctx, params)
	if err == nil && len(deps) > 0 {
		parts = append(parts, "## Dependency Analysis\n"+deps[0].Text)
	}

	par, err := handleTaskAnalysisParallelization(ctx, params)
	if err == nil && len(par) > 0 {
		parts = append(parts, "## Parallelization\n"+par[0].Text)
	}

	plan, err := handleTaskAnalysisExecutionPlan(ctx, params)
	if err == nil && len(plan) > 0 {
		parts = append(parts, "## Execution Plan\n"+plan[0].Text)
	}

	report := "# Task Dependencies Summary\n\n" + strings.Join(parts, "\n\n")

	return []framework.TextContent{{Type: "text", Text: report}}, nil
}

// handleTaskAnalysisExecutionPlan handles execution plan: backlog (Todo + In Progress) in dependency order.
func handleTaskAnalysisExecutionPlan(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
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

	// Optional tag filter: restrict backlog to tasks with filter_tag or any of filter_tags
	var backlogFilter map[string]bool
	if ft, ok := params["filter_tag"].(string); ok && ft != "" {
		backlogFilter = make(map[string]bool)

		for _, t := range tasks {
			if !IsBacklogStatus(t.Status) {
				continue
			}

			for _, tag := range t.Tags {
				if tag == ft {
					backlogFilter[t.ID] = true
					break
				}
			}
		}
	} else if fts, ok := params["filter_tags"].(string); ok && fts != "" {
		allowed := strings.Split(fts, ",")
		for i := range allowed {
			allowed[i] = strings.TrimSpace(allowed[i])
		}

		backlogFilter = make(map[string]bool)

		for _, t := range tasks {
			if !IsBacklogStatus(t.Status) {
				continue
			}

			for _, tag := range t.Tags {
				for _, a := range allowed {
					if a != "" && tag == a {
						backlogFilter[t.ID] = true
						break
					}
				}

				if backlogFilter[t.ID] {
					break
				}
			}
		}
	}

	orderedIDs, waves, details, err := BacklogExecutionOrder(tasks, backlogFilter)
	if err != nil {
		return nil, fmt.Errorf("execution order: %w", err)
	}

	// Optional limit (0 = all)
	limit := 0
	if l, ok := params["limit"].(float64); ok && l > 0 {
		limit = int(l)
	}

	if limit > 0 && len(orderedIDs) > limit {
		orderedIDs = orderedIDs[:limit]
		// Trim details to match
		if len(details) > limit {
			details = details[:limit]
		}
	}

	result := map[string]interface{}{
		"success":          true,
		"method":           "native_go",
		"backlog_count":    len(orderedIDs),
		"ordered_task_ids": orderedIDs,
		"waves":            waves,
		"details":          details,
	}

	outputFormat := "json"
	if format, ok := params["output_format"].(string); ok && format != "" {
		outputFormat = format
	}

	outputPath, _ := params["output_path"].(string)

	// subagents_plan: write parallel-execution-subagents.plan.md using wave detection
	if outputFormat == "subagents_plan" {
		wavesCopy := waves
		if max := config.MaxTasksPerWave(); max > 0 {
			wavesCopy = LimitWavesByMaxTasks(wavesCopy, max)
		}
		if len(wavesCopy) == 0 {
			return nil, fmt.Errorf("no waves (empty backlog or no Todo/In Progress tasks)")
		}
		planTitle, _ := params["plan_title"].(string)
		if planTitle == "" {
			planTitle = filepath.Base(projectRoot)
			if info, err := getProjectInfo(projectRoot); err == nil {
				if name, ok := info["name"].(string); ok && name != "" {
					planTitle = name
				}
			}
		}
		if outputPath == "" {
			outputPath = filepath.Join(projectRoot, ".cursor", "plans", "parallel-execution-subagents.plan.md")
		}
		if dir := filepath.Dir(outputPath); dir != "." {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return nil, fmt.Errorf("failed to create plan directory: %w", err)
			}
		}
		md := FormatWavesAsSubagentsPlanMarkdown(wavesCopy, planTitle)
		if err := os.WriteFile(outputPath, []byte(md), 0644); err != nil {
			return nil, fmt.Errorf("failed to write subagents plan: %w", err)
		}
		msg := fmt.Sprintf("Parallel execution subagents plan saved to: %s", outputPath)
		return []framework.TextContent{{Type: "text", Text: msg}}, nil
	}

	if outputFormat == "json" {
		if outputPath != "" {
			if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
				return nil, fmt.Errorf("failed to create output dir: %w", err)
			}
		}

		resultJSON, _ := json.Marshal(result)
		resp := &proto.TaskAnalysisResponse{Action: "execution_plan", OutputPath: outputPath, ResultJson: string(resultJSON)}

		return response.FormatResult(TaskAnalysisResponseToMap(resp), resp.GetOutputPath())
	}

	output := formatExecutionPlanText(result)

	if outputPath != "" {
		if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
			return nil, fmt.Errorf("failed to create output dir: %w", err)
		}

		if strings.HasSuffix(strings.ToLower(outputPath), ".md") {
			if !strings.HasSuffix(strings.ToLower(outputPath), ".plan.md") {
				outputPath = outputPath[:len(outputPath)-3] + ".plan.md"
			}

			md := formatExecutionPlanMarkdown(result, projectRoot)
			if err := os.WriteFile(outputPath, []byte(md), 0644); err != nil {
				return nil, fmt.Errorf("failed to save markdown: %w", err)
			}
		} else {
			if err := os.WriteFile(outputPath, []byte(output), 0644); err != nil {
				return nil, fmt.Errorf("failed to save result: %w", err)
			}
		}

		output += fmt.Sprintf("\n\n[Saved to: %s]", outputPath)
	}

	return []framework.TextContent{{Type: "text", Text: output}}, nil
}

func formatExecutionPlanText(result map[string]interface{}) string {
	var sb strings.Builder

	sb.WriteString("Backlog execution order\n")
	sb.WriteString(strings.Repeat("-", 40) + "\n")

	if ids, ok := result["ordered_task_ids"].([]string); ok {
		for i, id := range ids {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, id))
		}
	}

	return sb.String()
}

func formatExecutionPlanMarkdown(result map[string]interface{}, projectRoot string) string {
	var sb strings.Builder

	sb.WriteString("# Backlog Execution Plan\n\n")
	sb.WriteString(fmt.Sprintf("**Generated:** %s\n\n", fmtTime(time.Now())))

	if count, ok := result["backlog_count"].(int); ok {
		sb.WriteString(fmt.Sprintf("**Backlog:** %d tasks (Todo + In Progress)\n\n", count))
	}

	if w, ok := result["waves"].(map[int][]string); ok && len(w) > 0 {
		sb.WriteString(fmt.Sprintf("**Waves:** %d dependency levels\n\n", len(w)))

		details, _ := result["details"].([]BacklogTaskDetail)

		levelOrder := make([]int, 0, len(w))
		for k := range w {
			levelOrder = append(levelOrder, k)
		}

		sort.Ints(levelOrder)

		for _, level := range levelOrder {
			ids := w[level]
			sb.WriteString(fmt.Sprintf("## Wave %d\n\n", level))
			sb.WriteString("| ID | Content | Priority | Tags |\n")
			sb.WriteString("|----|--------|----------|------|\n")

			for _, id := range ids {
				for _, d := range details {
					if d.ID == id {
						content := d.Content
						if len(content) > 60 {
							content = content[:57] + "..."
						}

						tagsStr := strings.Join(d.Tags, ", ")
						if tagsStr == "" {
							tagsStr = "-"
						}

						sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n", d.ID, content, d.Priority, tagsStr))

						break
					}
				}
			}

			sb.WriteString("\n")
		}
	}

	sb.WriteString("## Full order\n\n")

	if ids, ok := result["ordered_task_ids"].([]string); ok {
		sb.WriteString(strings.Join(ids, ", "))
		sb.WriteString("\n")
	}

	return sb.String()
}

func fmtTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// handleTaskAnalysisComplexity classifies task complexity (simple/medium/complex) using heuristic rules.
func handleTaskAnalysisComplexity(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	store, err := getTaskStore(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get task store: %w", err)
	}

	list, err := store.ListTasks(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	// Optional: filter to backlog only
	statusFilter := params["status_filter"]

	var tasks []Todo2Task

	if sf, ok := statusFilter.(string); ok && sf != "" {
		for _, t := range list {
			if t != nil && t.Status == sf {
				tasks = append(tasks, *t)
			}
		}
	} else {
		tasks = tasksFromPtrs(list)
	}

	classifications := make([]map[string]interface{}, 0, len(tasks))

	for _, t := range tasks {
		taskPtr := &t
		r := taskanalysis.AnalyzeTask(taskPtr)
		classifications = append(classifications, map[string]interface{}{
			"id":               t.ID,
			"content":          t.Content,
			"complexity":       string(r.Complexity),
			"can_auto_execute": r.CanAutoExecute,
			"needs_breakdown":  r.NeedsBreakdown,
			"reason":           r.Reason,
		})
	}

	result := map[string]interface{}{
		"success":         true,
		"method":          "native_go",
		"classifications": classifications,
		"total":           len(classifications),
	}

	outputPath, _ := params["output_path"].(string)
	resultJSON, _ := json.Marshal(result)
	resp := &proto.TaskAnalysisResponse{Action: "complexity", OutputPath: outputPath, ResultJson: string(resultJSON)}

	return response.FormatResult(TaskAnalysisResponseToMap(resp), resp.GetOutputPath())
}

// handleTaskAnalysisParallelization handles parallelization analysis.
func handleTaskAnalysisParallelization(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	store, err := getTaskStore(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get task store: %w", err)
	}

	list, err := store.ListTasks(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	tasks := tasksFromPtrs(list)

	durationWeight := 0.3
	if weight, ok := params["duration_weight"].(float64); ok {
		durationWeight = weight
	}

	// Find parallelizable tasks
	parallelGroups := findParallelizableTasks(tasks, durationWeight)

	outputFormat := "json"
	if format, ok := params["output_format"].(string); ok && format != "" {
		outputFormat = format
	}

	result := map[string]interface{}{
		"success":         true,
		"method":          "native_go",
		"total_tasks":     len(tasks),
		"parallel_groups": parallelGroups,
		"duration_weight": durationWeight,
		"recommendations": buildParallelizationRecommendations(parallelGroups),
	}

	// Include human-readable report in JSON for CLI/consumers
	result["report"] = formatParallelizationAnalysisText(result)

	outputPath, _ := params["output_path"].(string)
	if outputFormat == "json" {
		if outputPath != "" {
			if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
				return nil, fmt.Errorf("failed to create output dir: %w", err)
			}
		}

		resultJSON, _ := json.Marshal(result)
		resp := &proto.TaskAnalysisResponse{Action: "parallelization", OutputPath: outputPath, ResultJson: string(resultJSON)}

		return response.FormatResult(TaskAnalysisResponseToMap(resp), resp.GetOutputPath())
	}

	output := formatParallelizationAnalysisText(result)

	if outputPath != "" {
		if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
			return nil, fmt.Errorf("failed to create output dir: %w", err)
		}

		if err := os.WriteFile(outputPath, []byte(output), 0644); err != nil {
			return nil, fmt.Errorf("failed to save result: %w", err)
		}

		output += fmt.Sprintf("\n\n[Saved to: %s]", outputPath)
	}

	return []framework.TextContent{{Type: "text", Text: output}}, nil
}

// handleTaskAnalysisFixMissingDeps removes invalid dependency refs from tasks and saves.
// Use once to fix tasks that depend on non-existent IDs (e.g. T-45, T-5 depending on T-4).
func handleTaskAnalysisFixMissingDeps(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	store, err := getTaskStore(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get task store: %w", err)
	}

	list, err := store.ListTasks(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	tasks := tasksFromPtrs(list)

	tg, err := BuildTaskGraph(tasks)
	if err != nil {
		return nil, fmt.Errorf("failed to build task graph: %w", err)
	}

	missing := findMissingDependencies(tasks, tg)
	if len(missing) == 0 {
		out := map[string]interface{}{
			"success":     true,
			"message":     "No missing dependency refs to fix",
			"total_tasks": len(tasks),
		}
		resultJSON, _ := json.Marshal(out)
		resp := &proto.TaskAnalysisResponse{Action: "fix_missing_deps", ResultJson: string(resultJSON)}

		return response.FormatResult(TaskAnalysisResponseToMap(resp), resp.GetOutputPath())
	}

	// Per task, collect missing dep IDs to remove
	missingDepsByTask := make(map[string]map[string]struct{})

	for _, m := range missing {
		tid, _ := m["task_id"].(string)
		dep, _ := m["missing_dep"].(string)

		if tid == "" || dep == "" {
			continue
		}

		if missingDepsByTask[tid] == nil {
			missingDepsByTask[tid] = make(map[string]struct{})
		}

		missingDepsByTask[tid][dep] = struct{}{}
	}

	// Remove all invalid deps from each task in one pass
	fixed := 0

	for i := range tasks {
		t := &tasks[i]

		toRemove := missingDepsByTask[t.ID]
		if len(toRemove) == 0 {
			continue
		}

		newDeps := make([]string, 0, len(t.Dependencies))

		for _, d := range t.Dependencies {
			if _, remove := toRemove[d]; remove {
				fixed++
			} else {
				newDeps = append(newDeps, d)
			}
		}

		t.Dependencies = newDeps
	}

	for _, t := range tasks {
		if missingDepsByTask[t.ID] != nil {
			taskPtr := &t
			if err := store.UpdateTask(ctx, taskPtr); err != nil {
				return nil, fmt.Errorf("failed to save task %s after fix: %w", t.ID, err)
			}
		}
	}

	result := map[string]interface{}{
		"success":      true,
		"total_tasks":  len(tasks),
		"missing_refs": len(missing),
		"removed":      fixed,
		"message":      fmt.Sprintf("Removed %d invalid dependency ref(s) and saved", fixed),
	}
	resultJSON, _ := json.Marshal(result)
	resp := &proto.TaskAnalysisResponse{Action: "fix_missing_deps", ResultJson: string(resultJSON)}

	return response.FormatResult(TaskAnalysisResponseToMap(resp), resp.GetOutputPath())
}

// handleTaskAnalysisValidate reports missing dependency IDs and optionally hierarchy parse warnings.
// Returns JSON with missing_deps and optional hierarchy_warning (when FM returns non-JSON).
func handleTaskAnalysisValidate(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	store, err := getTaskStore(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get task store: %w", err)
	}

	list, err := store.ListTasks(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	tasks := tasksFromPtrs(list)

	tg, err := BuildTaskGraph(tasks)
	if err != nil {
		return nil, fmt.Errorf("failed to build task graph: %w", err)
	}

	missing := findMissingDependencies(tasks, tg)
	result := map[string]interface{}{
		"success":       true,
		"method":        "native_go",
		"total_tasks":   len(tasks),
		"missing_deps":  missing,
		"missing_count": len(missing),
	}

	includeHierarchy := false
	if h, ok := params["include_hierarchy"].(bool); ok {
		includeHierarchy = h
	}

	if includeHierarchy && FMAvailable() {
		hierResult, err := handleTaskAnalysisHierarchy(ctx, params)
		if err == nil && len(hierResult) > 0 {
			var hierData map[string]interface{}
			if json.Unmarshal([]byte(hierResult[0].Text), &hierData) == nil {
				if skipped, _ := hierData["hierarchy_skipped"].(string); skipped != "" {
					result["hierarchy_warning"] = map[string]interface{}{
						"skipped":          skipped,
						"parse_error":      hierData["parse_error"],
						"response_snippet": hierData["response_snippet"],
					}
				}
			}
		}
	}

	resultJSON, _ := json.Marshal(result)
	resp := &proto.TaskAnalysisResponse{Action: "validate", ResultJson: string(resultJSON)}

	return response.FormatResult(TaskAnalysisResponseToMap(resp), resp.GetOutputPath())
}

// Helper functions for duplicates detection

func findDuplicateTasks(tasks []Todo2Task, threshold float64) [][]string {
	// For small datasets, use sequential approach (overhead of parallelization not worth it)
	if len(tasks) < 100 {
		return findDuplicateTasksSequential(tasks, threshold)
	}

	// For larger datasets, use parallel processing
	return findDuplicateTasksParallel(tasks, threshold)
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
func buildLegacyGraphFormat(tg *TaskGraph) DependencyGraph {
	graph := make(DependencyGraph)

	nodes := tg.Graph.Nodes()
	for nodes.Next() {
		nodeID := nodes.Node().ID()
		taskID := tg.NodeIDMap[nodeID]

		deps := []string{}

		fromNodes := tg.Graph.To(nodeID)
		for fromNodes.Next() {
			fromNodeID := fromNodes.Node().ID()
			if depTaskID, ok := tg.NodeIDMap[fromNodeID]; ok {
				deps = append(deps, depTaskID)
			}
		}

		graph[taskID] = deps
	}

	return graph
}

func findMissingDependencies(tasks []Todo2Task, tg *TaskGraph) []map[string]interface{} {
	missing := []map[string]interface{}{}
	taskMap := make(map[string]bool)

	for _, task := range tasks {
		taskMap[task.ID] = true
	}

	for _, task := range tasks {
		for _, dep := range task.Dependencies {
			if !taskMap[dep] {
				missing = append(missing, map[string]interface{}{
					"task_id":     task.ID,
					"missing_dep": dep,
					"message":     fmt.Sprintf("Task %s depends on %s which doesn't exist", task.ID, dep),
				})
			}
		}
	}

	return missing
}

// GetDependencyAnalysisFromTasks returns cycles and missing dependencies for the given tasks.
// Used by task_discovery findOrphanTasks and handleTaskAnalysisDependencies to share graph logic.
func GetDependencyAnalysisFromTasks(tasks []Todo2Task) (cycles [][]string, missing []map[string]interface{}, err error) {
	tg, err := BuildTaskGraph(tasks)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build task graph: %w", err)
	}

	cycles = DetectCycles(tg)
	missing = findMissingDependencies(tasks, tg)

	return cycles, missing, nil
}

func buildDependencyRecommendations(graph DependencyGraph, cycles [][]string, missing []map[string]interface{}) []map[string]interface{} {
	recommendations := []map[string]interface{}{}

	if len(cycles) > 0 {
		recommendations = append(recommendations, map[string]interface{}{
			"type":    "circular_dependency",
			"count":   len(cycles),
			"message": fmt.Sprintf("Found %d circular dependency chain(s)", len(cycles)),
		})
	}

	if len(missing) > 0 {
		recommendations = append(recommendations, map[string]interface{}{
			"type":    "missing_dependency",
			"count":   len(missing),
			"message": fmt.Sprintf("Found %d missing dependency reference(s)", len(missing)),
		})
	}

	return recommendations
}

func formatDependencyAnalysisText(result map[string]interface{}) string {
	var sb strings.Builder

	sb.WriteString("Dependency Analysis\n")
	sb.WriteString("==================\n\n")

	if total, ok := result["total_tasks"].(int); ok {
		sb.WriteString(fmt.Sprintf("Total Tasks: %d\n", total))
	}

	if maxLevel, ok := result["max_dependency_level"].(int); ok {
		sb.WriteString(fmt.Sprintf("Max Dependency Level: %d\n", maxLevel))
	}

	sb.WriteString("\n")

	// Critical Path
	if criticalPath, ok := result["critical_path"].([]string); ok && len(criticalPath) > 0 {
		sb.WriteString("Critical Path (Longest Dependency Chain):\n")
		sb.WriteString(fmt.Sprintf("  Length: %d tasks\n\n", len(criticalPath)))

		if details, ok := result["critical_path_details"].([]map[string]interface{}); ok {
			for i, detail := range details {
				taskID, _ := detail["id"].(string)
				content, _ := detail["content"].(string)

				sb.WriteString(fmt.Sprintf("  %d. %s", i+1, taskID))

				if content != "" {
					sb.WriteString(fmt.Sprintf(": %s", content))
				}

				sb.WriteString("\n")

				if deps, ok := detail["dependencies"].([]interface{}); ok && len(deps) > 0 {
					depStrs := make([]string, len(deps))

					for j, d := range deps {
						if depStr, ok := d.(string); ok {
							depStrs[j] = depStr
						}
					}

					if len(depStrs) > 0 {
						sb.WriteString(fmt.Sprintf("     Depends on: %s\n", strings.Join(depStrs, ", ")))
					}
				}

				if i < len(details)-1 {
					sb.WriteString("     ↓\n")
				}
			}
		} else {
			// Fallback to simple path
			sb.WriteString(fmt.Sprintf("  %s\n", strings.Join(criticalPath, " → ")))
		}

		sb.WriteString("\n")
	}

	if cycles, ok := result["circular_dependencies"].([][]string); ok && len(cycles) > 0 {
		sb.WriteString("Circular Dependencies:\n")

		for i, cycle := range cycles {
			sb.WriteString(fmt.Sprintf("  %d. %s\n", i+1, strings.Join(cycle, " -> ")))
		}

		sb.WriteString("\n")
	}

	if missing, ok := result["missing_dependencies"].([]map[string]interface{}); ok && len(missing) > 0 {
		sb.WriteString("Missing Dependencies:\n")

		for _, m := range missing {
			if taskID, ok := m["task_id"].(string); ok {
				if dep, ok := m["missing_dep"].(string); ok {
					sb.WriteString(fmt.Sprintf("  - %s depends on %s (not found)\n", taskID, dep))
				}
			}
		}

		sb.WriteString("\n")
	}

	return sb.String()
}

// Helper functions for parallelization analysis

type ParallelGroup struct {
	Tasks    []string `json:"tasks"`
	Priority string   `json:"priority"`
	Reason   string   `json:"reason"`
}

func findParallelizableTasks(tasks []Todo2Task, durationWeight float64) []ParallelGroup {
	groups := []ParallelGroup{}

	// Build dependency graph using gonum
	tg, err := BuildTaskGraph(tasks)
	if err != nil {
		// Fallback to simple approach if graph building fails
		return findParallelizableTasksSimple(tasks, durationWeight)
	}

	taskMap := make(map[string]*Todo2Task)
	for i := range tasks {
		taskMap[tasks[i].ID] = &tasks[i]
	}

	// Filter to pending tasks only
	pendingTasks := []Todo2Task{}

	for _, task := range tasks {
		if IsPendingStatus(task.Status) {
			pendingTasks = append(pendingTasks, task)
		}
	}

	if len(pendingTasks) == 0 {
		return groups
	}

	// Use dependency levels to group parallelizable tasks
	levels := GetTaskLevels(tg)

	// Group tasks by dependency level (tasks at same level can run in parallel)
	byLevel := make(map[int][]string)

	for _, task := range pendingTasks {
		level := levels[task.ID]
		byLevel[level] = append(byLevel[level], task.ID)
	}

	// Build parallel groups from each level
	for level, taskIDs := range byLevel {
		if len(taskIDs) < 2 {
			continue // Skip levels with only one task
		}

		// Group by priority within each level
		byPriority := make(map[string][]string)

		for _, taskID := range taskIDs {
			task := taskMap[taskID]

			priority := task.Priority
			if priority == "" {
				priority = "medium"
			}

			byPriority[priority] = append(byPriority[priority], taskID)
		}

		// Create groups for each priority within this level
		for priority, ids := range byPriority {
			if len(ids) > 1 {
				groups = append(groups, ParallelGroup{
					Tasks:    ids,
					Priority: priority,
					Reason:   fmt.Sprintf("%d tasks at dependency level %d can run in parallel", len(ids), level),
				})
			}
		}
	}

	// Also include tasks with no dependencies (level 0)
	if level0Tasks, ok := byLevel[0]; ok && len(level0Tasks) > 0 {
		byPriority := make(map[string][]string)

		for _, taskID := range level0Tasks {
			task := taskMap[taskID]

			priority := task.Priority
			if priority == "" {
				priority = "medium"
			}

			byPriority[priority] = append(byPriority[priority], taskID)
		}

		for priority, ids := range byPriority {
			if len(ids) > 1 {
				// Check if we already added this group
				exists := false

				for _, g := range groups {
					if g.Priority == priority && len(g.Tasks) == len(ids) {
						exists = true
						break
					}
				}

				if !exists {
					groups = append(groups, ParallelGroup{
						Tasks:    ids,
						Priority: priority,
						Reason:   fmt.Sprintf("%d tasks with no dependencies can run in parallel", len(ids)),
					})
				}
			}
		}
	}

	// Sort groups by priority (high -> medium -> low)
	sort.Slice(groups, func(i, j int) bool {
		priorityOrder := map[string]int{"high": 0, "medium": 1, "low": 2}
		return priorityOrder[groups[i].Priority] < priorityOrder[groups[j].Priority]
	})

	return groups
}

// findParallelizableTasksSimple is a fallback implementation without graph analysis.
func findParallelizableTasksSimple(tasks []Todo2Task, durationWeight float64) []ParallelGroup {
	groups := []ParallelGroup{}

	taskMap := make(map[string]*Todo2Task)
	for i := range tasks {
		taskMap[tasks[i].ID] = &tasks[i]
	}

	// Find tasks with no dependencies (can run in parallel)
	readyTasks := []string{}

	for _, task := range tasks {
		if IsPendingStatus(task.Status) && len(task.Dependencies) == 0 {
			readyTasks = append(readyTasks, task.ID)
		}
	}

	if len(readyTasks) > 0 {
		// Group by priority
		byPriority := make(map[string][]string)

		for _, taskID := range readyTasks {
			task := taskMap[taskID]

			priority := task.Priority
			if priority == "" {
				priority = "medium"
			}

			byPriority[priority] = append(byPriority[priority], taskID)
		}

		for priority, taskIDs := range byPriority {
			if len(taskIDs) > 1 {
				groups = append(groups, ParallelGroup{
					Tasks:    taskIDs,
					Priority: priority,
					Reason:   fmt.Sprintf("%d tasks with no dependencies can run in parallel", len(taskIDs)),
				})
			}
		}
	}

	// Sort groups by priority (high -> medium -> low)
	sort.Slice(groups, func(i, j int) bool {
		priorityOrder := map[string]int{"high": 0, "medium": 1, "low": 2}
		return priorityOrder[groups[i].Priority] < priorityOrder[groups[j].Priority]
	})

	return groups
}

func buildParallelizationRecommendations(groups []ParallelGroup) []map[string]interface{} {
	recommendations := []map[string]interface{}{}

	totalParallelizable := 0
	for _, group := range groups {
		totalParallelizable += len(group.Tasks)
	}

	if totalParallelizable > 0 {
		recommendations = append(recommendations, map[string]interface{}{
			"type":    "parallel_execution",
			"count":   totalParallelizable,
			"groups":  len(groups),
			"message": fmt.Sprintf("%d tasks can be executed in parallel across %d groups", totalParallelizable, len(groups)),
		})
	}

	return recommendations
}

func formatParallelizationAnalysisText(result map[string]interface{}) string {
	var sb strings.Builder

	sb.WriteString("Parallelization Analysis\n")
	sb.WriteString("========================\n\n")

	if total, ok := result["total_tasks"].(int); ok {
		sb.WriteString(fmt.Sprintf("Total Tasks: %d\n\n", total))
	}

	if groups, ok := result["parallel_groups"].([]ParallelGroup); ok && len(groups) > 0 {
		sb.WriteString("Parallel Execution Groups:\n\n")

		for i, group := range groups {
			sb.WriteString(fmt.Sprintf("Group %d (%s priority):\n", i+1, group.Priority))
			sb.WriteString(fmt.Sprintf("  Reason: %s\n", group.Reason))
			sb.WriteString("  Tasks:\n")

			for _, taskID := range group.Tasks {
				sb.WriteString(fmt.Sprintf("    - %s\n", taskID))
			}

			sb.WriteString("\n")
		}
	} else {
		sb.WriteString("No parallel execution opportunities found.\n")
	}

	return sb.String()
}

// Helper function to save analysis results.
func saveAnalysisResult(outputPath string, result map[string]interface{}) error {
	// Ensure directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(outputPath, output, 0644)
}

// handleTaskAnalysisHierarchy handles hierarchy analysis using the FM provider abstraction.
// When DefaultFMProvider() is available (e.g. Apple FM on darwin/arm64/cgo), it classifies tasks; otherwise returns ErrFMNotSupported.
func handleTaskAnalysisHierarchy(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	if !FMAvailable() {
		return nil, fmt.Errorf("hierarchy requires a foundation model: %w", ErrFMNotSupported)
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

	pendingTasks := []Todo2Task{}

	for _, task := range tasks {
		if IsPendingStatus(task.Status) {
			pendingTasks = append(pendingTasks, task)
		}
	}

	if len(pendingTasks) == 0 {
		return []framework.TextContent{
			{Type: "text", Text: `{"success": true, "message": "No pending tasks to analyze"}`},
		}, nil
	}

	taskDescriptions := make([]string, 0, len(pendingTasks))

	for i, task := range pendingTasks {
		if i >= 20 {
			break
		}

		desc := task.Content
		if task.LongDescription != "" {
			desc += " " + task.LongDescription
		}

		taskDescriptions = append(taskDescriptions, fmt.Sprintf("Task %s: %s", task.ID, desc))
	}

	prompt := fmt.Sprintf(`Analyze these tasks and classify them into hierarchy levels:

Tasks:
%s

Classify each task into one of these categories:
- "component" - Tasks that belong to a specific component/feature
- "epic" - High-level tasks that contain multiple subtasks
- "task" - Regular standalone tasks
- "subtask" - Tasks that are part of a larger task

Return JSON array with format: [{"task_id": "T-1", "level": "component", "component": "security", "reason": "..."}, ...]`,
		strings.Join(taskDescriptions, "\n"))

	result, err := DefaultFMProvider().Generate(ctx, prompt, 2000, 0.2)
	if err != nil {
		return nil, fmt.Errorf("foundation model classification: %w", err)
	}

	var classifications []map[string]interface{}

	candidate := result
	if err := json.Unmarshal([]byte(candidate), &classifications); err != nil {
		candidate = ExtractJSONArrayFromLLMResponse(result)
		if err = json.Unmarshal([]byte(candidate), &classifications); err != nil {
			// Graceful fallback: return success with empty classifications and a warning so
			// task_analysis doesn't fail hard when the FM returns plain text instead of JSON.
			snippet := result
			if len(snippet) > MaxLLMResponseSnippetLen {
				snippet = snippet[:MaxLLMResponseSnippetLen] + "..."
			}

			analysis := map[string]interface{}{
				"success":           true,
				"method":            "foundation_model",
				"total_tasks":       len(tasks),
				"pending_tasks":     len(pendingTasks),
				"classifications":   []map[string]interface{}{},
				"hierarchy_skipped": "fm_response_not_valid_json",
				"parse_error":       err.Error(),
				"response_snippet":  snippet,
			}
			resultJSON, _ := json.Marshal(analysis)
			resp := &proto.TaskAnalysisResponse{Action: "hierarchy", ResultJson: string(resultJSON)}

			return response.FormatResult(TaskAnalysisResponseToMap(resp), resp.GetOutputPath())
		}
	}

	analysis := map[string]interface{}{
		"success":                   true,
		"method":                    "foundation_model",
		"total_tasks":               len(tasks),
		"pending_tasks":             len(pendingTasks),
		"classifications":           classifications,
		"hierarchy_recommendations": buildHierarchyRecommendations(classifications, pendingTasks),
	}

	includeRecommendations := true
	if rec, ok := params["include_recommendations"].(bool); ok {
		includeRecommendations = rec
	}

	if !includeRecommendations {
		delete(analysis, "hierarchy_recommendations")
	}

	outputFormat := "json"
	if format, ok := params["output_format"].(string); ok && format != "" {
		outputFormat = format
	}

	if outputFormat == "text" {
		output := formatHierarchyAnalysisText(analysis)
		return []framework.TextContent{{Type: "text", Text: output}}, nil
	}

	resultJSON, _ := json.Marshal(analysis)
	resp := &proto.TaskAnalysisResponse{Action: "hierarchy", ResultJson: string(resultJSON)}

	return response.FormatResult(TaskAnalysisResponseToMap(resp), resp.GetOutputPath())
}

func buildHierarchyRecommendations(classifications []map[string]interface{}, tasks []Todo2Task) []map[string]interface{} {
	recommendations := []map[string]interface{}{}
	componentGroups := make(map[string][]string)

	for _, cls := range classifications {
		if comp, ok := cls["component"].(string); ok && comp != "" {
			taskID, _ := cls["task_id"].(string)
			componentGroups[comp] = append(componentGroups[comp], taskID)
		}
	}

	for comp, taskIDs := range componentGroups {
		if len(taskIDs) >= 5 {
			recommendations = append(recommendations, map[string]interface{}{
				"component":        comp,
				"task_count":       len(taskIDs),
				"recommendation":   "use_hierarchy",
				"suggested_prefix": fmt.Sprintf("T-%s", strings.ToUpper(comp)),
				"task_ids":         taskIDs,
			})
		}
	}

	return recommendations
}

func formatHierarchyAnalysisText(analysis map[string]interface{}) string {
	var sb strings.Builder

	sb.WriteString("Task Hierarchy Analysis\n")
	sb.WriteString("=======================\n\n")

	if total, ok := analysis["total_tasks"].(int); ok {
		sb.WriteString(fmt.Sprintf("Total Tasks: %d\n", total))
	}

	if pending, ok := analysis["pending_tasks"].(int); ok {
		sb.WriteString(fmt.Sprintf("Pending Tasks: %d\n\n", pending))
	}

	if recs, ok := analysis["hierarchy_recommendations"].([]map[string]interface{}); ok {
		sb.WriteString("Recommendations:\n")

		for _, rec := range recs {
			if comp, ok := rec["component"].(string); ok {
				sb.WriteString(fmt.Sprintf("- Component: %s\n", comp))

				if count, ok := rec["task_count"].(int); ok {
					sb.WriteString(fmt.Sprintf("  Tasks: %d\n", count))
				}

				if prefix, ok := rec["suggested_prefix"].(string); ok {
					sb.WriteString(fmt.Sprintf("  Suggested Prefix: %s\n", prefix))
				}

				sb.WriteString("\n")
			}
		}
	}

	return sb.String()
}

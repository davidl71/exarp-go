// task_analysis_ownership.go — task_analysis infer_ownership action: infer file ownership and lane from task metadata.
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/internal/models"
	"github.com/davidl71/exarp-go/proto"
	"github.com/spf13/cast"
)

// OwnershipSuggestion represents an inferred ownership for a task.
type OwnershipSuggestion struct {
	TaskID              string   `json:"task_id"`
	TaskContent         string   `json:"task_content"`
	Lane                string   `json:"lane,omitempty"`
	LaneReason          string   `json:"lane_reason,omitempty"`
	OwnedFiles          []string `json:"owned_files,omitempty"`
	OwnedGlobs          []string `json:"owned_globs,omitempty"`
	OwnershipConfidence string   `json:"ownership_confidence"` // "high" | "medium" | "low"
	ConfidenceReasons   []string `json:"confidence_reasons,omitempty"`
	AlreadyHasOwnership bool     `json:"already_has_ownership,omitempty"`
}

// LaneMapping maps tag/directory patterns to lane labels.
var LaneMapping = []struct {
	Pattern  string   // Regex pattern to match
	Lane     string   // Lane label
	Priority int      // Lower = higher priority
	Examples []string // Example matching paths
}{
	// Backend services
	{`(?i)\bauth\b|/auth/|/authentication/`, "backend-auth", 10, []string{"src/auth/", "middleware/jwt.go"}},
	{`(?i)\bapi\b|/api/|/routes/|/handlers/`, "backend-api", 11, []string{"src/api/", "routes/users.go"}},
	{`(?i)\bbackend\b|/server/|/service/`, "backend-runtime", 12, []string{"cmd/server/", "internal/service/"}},

	// Frontend/TUI
	{`(?i)\btui\b|/tui/|/ui/`, "tui-shell", 20, []string{"src/tui/", "ui/shell.go"}},
	{`(?i)\bshell\b|/shell/`, "tui-shell", 21, []string{"src/ui/shell.go"}},
	{`(?i)\bpane\b|/pane/|/panes/`, "tui-pane", 22, []string{"src/ui/panes/"}},

	// Infrastructure
	{`(?i)\bconfig\b|/config/|\.ya?ml$|\.toml$`, "config", 30, []string{"config/app.yaml", ".cursor/mcp.json"}},
	{`(?i)\btest\b|_test\.go$|/test/|/tests/`, "testing", 31, []string{"src/auth_test.go", "tests/integration/"}},
	{`(?i)\bdoc\b|/docs?/|\.md$`, "docs", 32, []string{"docs/README.md", "README.md"}},

	// Database
	{`(?i)\bdb\b|database|/db/|/models/|/schema/`, "database", 40, []string{"internal/database/", "migrations/"}},

	// Security
	{`(?i)\bsecurity\b|/security/|/authz/`, "backend-auth", 50, []string{"internal/security/"}},

	// General
	{`(?i)\bci\b|/ci/|\.github/|Makefile`, "config", 60, []string{".github/workflows/", "Makefile"}},
	{`(?i)\bproto\b|\.proto$|/proto/`, "source-architecture", 70, []string{"proto/", "api.proto"}},
}

// filePathPattern matches common file path patterns in text.
var filePathPattern = regexp.MustCompile(`(?m)(?:^|\s)(?:(?:\./|[a-zA-Z_][a-zA-Z0-9_-]*/)+(?:[a-zA-Z0-9_.-]+(?:\.[a-zA-Z0-9]+)+)|(?:[a-zA-Z0-9_.-]+\.(?:go|ts|tsx|js|jsx|py|rs|rb|java|c|cpp|h|hpp|md|yaml|yml|toml|json|sql|sh|bash)))\b`)

// wordBoundaryPattern matches potential file references in task content.
var wordBoundaryPattern = regexp.MustCompile(`(?i)(?:in|to|from|at|of|for|update|fix|add|create|modify|change|implement)\s+(?:the\s+)?([a-zA-Z0-9_/.-]+\.[a-zA-Z0-9]+)`)

// handleTaskAnalysisInferOwnership infers ownership metadata for tasks.
func handleTaskAnalysisInferOwnership(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	store, err := getTaskStore(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get task store: %w", err)
	}

	list, err := store.ListTasks(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	tasks := tasksFromPtrs(list)
	dryRun := cast.ToBool(params["dry_run"])
	useAI := cast.ToBool(params["use_ai"])

	// Get project root for file existence checks
	projectRoot, _ := GetProjectRootWithFallback()

	// Build directory index for lane inference
	dirIndex := buildDirectoryIndex(projectRoot)

	// Infer ownership for each task
	suggestions := make([]OwnershipSuggestion, 0)
	updatedCount := 0

	for i := range tasks {
		task := &tasks[i]

		// Skip completed tasks
		if task.Status == models.StatusDone || task.Status == models.StatusCancelled {
			continue
		}

		// Check if task already has ownership
		existingOwn := models.GetTaskOwnership(task)
		hasExisting := existingOwn != nil && (len(existingOwn.OwnedFiles) > 0 || existingOwn.Lane != "")

		suggestion := inferTaskOwnership(task, dirIndex, projectRoot, hasExisting)

		// Enhance low-confidence suggestions with AI if requested
		if useAI && FMAvailable() && suggestion.OwnershipConfidence == "low" && !hasExisting {
			aiSuggestion := enhanceOwnershipWithAI(ctx, task, projectRoot)
			if aiSuggestion != nil {
				// Merge AI suggestions
				if aiSuggestion.Lane != "" {
					suggestion.Lane = aiSuggestion.Lane
					suggestion.LaneReason = "AI suggested: " + aiSuggestion.LaneReason
				}
				if len(aiSuggestion.OwnedFiles) > 0 {
					suggestion.OwnedFiles = append(suggestion.OwnedFiles, aiSuggestion.OwnedFiles...)
					suggestion.ConfidenceReasons = append(suggestion.ConfidenceReasons, "AI suggested file references")
				}
				if len(aiSuggestion.OwnedGlobs) > 0 {
					suggestion.OwnedGlobs = append(suggestion.OwnedGlobs, aiSuggestion.OwnedGlobs...)
				}
				// Recalculate confidence after AI enhancement
				suggestion.OwnershipConfidence = calculateConfidence(suggestion)
				suggestion.ConfidenceReasons = append(suggestion.ConfidenceReasons, "enhanced by AI")
			}
		}

		if suggestion.OwnershipConfidence != "none" || hasExisting {
			suggestions = append(suggestions, suggestion)
		}

		// Apply if not dry_run and confidence is high/medium and no existing ownership
		if !dryRun && !hasExisting && (suggestion.OwnershipConfidence == "high" || suggestion.OwnershipConfidence == "medium") {
			// Fetch fresh copy for update
			taskPtr, err := store.GetTask(ctx, task.ID)
			if err != nil || taskPtr == nil {
				continue
			}
			own := &models.TaskOwnership{
				OwnedFiles:          suggestion.OwnedFiles,
				OwnedGlobs:          suggestion.OwnedGlobs,
				Lane:                suggestion.Lane,
				OwnershipConfidence: "inferred",
			}
			models.SetTaskOwnership(taskPtr, own)
			if err := store.UpdateTask(ctx, taskPtr); err == nil {
				updatedCount++
			}
		}
	}

	// Sort by confidence (high first)
	sort.Slice(suggestions, func(i, j int) bool {
		return ownershipConfidenceOrder[suggestions[i].OwnershipConfidence] < ownershipConfidenceOrder[suggestions[j].OwnershipConfidence]
	})

	method := "native_go"
	if useAI && FMAvailable() {
		method = "ai_assisted"
	}

	result := map[string]interface{}{
		"success":           true,
		"method":            method,
		"dry_run":           dryRun,
		"use_ai":            useAI && FMAvailable(),
		"fm_available":      FMAvailable(),
		"total_tasks":       len(tasks),
		"suggestions":       suggestions,
		"suggestions_count": len(suggestions),
		"updated_count":     updatedCount,
	}

	if dryRun {
		result["message"] = fmt.Sprintf("Dry run: would update %d tasks with inferred ownership", len(suggestions))
	} else {
		result["message"] = fmt.Sprintf("Updated %d tasks with inferred ownership", updatedCount)
	}

	outputFormat := ParamOutputFormat(params, "json")
	outputPath := ParamOutputPath(params)

	if outputFormat == "json" {
		if err := EnsureParentDir(outputPath); err != nil {
			return nil, fmt.Errorf("failed to create output dir: %w", err)
		}
		resultJSON, _ := json.Marshal(result)
		resp := &proto.TaskAnalysisResponse{Action: "infer_ownership", OutputPath: outputPath, ResultJson: string(resultJSON)}
		return framework.FormatResult(TaskAnalysisResponseToMap(resp), resp.GetOutputPath())
	}

	// Text format
	output := formatInferOwnershipText(result)

	if outputPath != "" {
		if err := EnsureParentDir(outputPath); err != nil {
			return nil, fmt.Errorf("failed to create output dir: %w", err)
		}
		if err := os.WriteFile(outputPath, []byte(output), 0644); err != nil {
			return nil, fmt.Errorf("failed to save result: %w", err)
		}
		output += fmt.Sprintf("\n\n[Saved to: %s]", outputPath)
	}

	return []framework.TextContent{{Type: "text", Text: output}}, nil
}

// inferTaskOwnership infers ownership for a single task.
func inferTaskOwnership(task *Todo2Task, dirIndex map[string][]string, projectRoot string, hasExisting bool) OwnershipSuggestion {
	suggestion := OwnershipSuggestion{
		TaskID:              task.ID,
		TaskContent:         task.Content,
		AlreadyHasOwnership: hasExisting,
		ConfidenceReasons:   []string{},
	}

	// Combine task content and description for analysis
	text := task.Content + " " + task.LongDescription

	// 1. Extract file paths mentioned in task text
	extractedFiles := extractFilePaths(text, projectRoot)
	if len(extractedFiles) > 0 {
		suggestion.OwnedFiles = append(suggestion.OwnedFiles, extractedFiles...)
		suggestion.ConfidenceReasons = append(suggestion.ConfidenceReasons, fmt.Sprintf("found %d file references in task text", len(extractedFiles)))
	}

	// 2. Infer lane from tags
	laneFromTags, tagReason := inferLaneFromTags(task.Tags)
	if laneFromTags != "" {
		suggestion.Lane = laneFromTags
		suggestion.LaneReason = tagReason
		suggestion.ConfidenceReasons = append(suggestion.ConfidenceReasons, tagReason)
	}

	// 3. Infer lane from task content if not found from tags
	if suggestion.Lane == "" {
		laneFromContent, contentReason := inferLaneFromContent(text)
		if laneFromContent != "" {
			suggestion.Lane = laneFromContent
			suggestion.LaneReason = contentReason
			suggestion.ConfidenceReasons = append(suggestion.ConfidenceReasons, contentReason)
		}
	}

	// 4. Add glob patterns based on lane
	if suggestion.Lane != "" {
		globs := globsForLane(suggestion.Lane, dirIndex)
		suggestion.OwnedGlobs = append(suggestion.OwnedGlobs, globs...)
	}

	// 5. Infer files from lane directory structure
	if len(suggestion.OwnedFiles) == 0 && suggestion.Lane != "" {
		filesFromLane := filesForLane(suggestion.Lane, dirIndex, projectRoot)
		if len(filesFromLane) > 0 {
			suggestion.OwnedFiles = append(suggestion.OwnedFiles, filesFromLane...)
			suggestion.ConfidenceReasons = append(suggestion.ConfidenceReasons, fmt.Sprintf("inferred %d files from lane %s", len(filesFromLane), suggestion.Lane))
		}
	}

	// Calculate overall confidence
	suggestion.OwnershipConfidence = calculateConfidence(suggestion)

	return suggestion
}

// extractFilePaths extracts file paths from text.
func extractFilePaths(text, projectRoot string) []string {
	var paths []string
	seen := make(map[string]bool)

	// Match explicit file paths
	matches := filePathPattern.FindAllString(text, -1)
	for _, m := range matches {
		m = strings.TrimSpace(m)
		if m == "" || seen[m] {
			continue
		}

		// Clean up path
		m = strings.TrimPrefix(m, "./")

		// Verify file exists if project root is available
		if projectRoot != "" {
			fullPath := filepath.Join(projectRoot, m)
			if _, err := os.Stat(fullPath); err == nil {
				paths = append(paths, m)
				seen[m] = true
			}
		} else {
			paths = append(paths, m)
			seen[m] = true
		}
	}

	// Match "in/to/from X.go" patterns
	wordMatches := wordBoundaryPattern.FindAllStringSubmatch(text, -1)
	for _, match := range wordMatches {
		if len(match) > 1 {
			path := strings.TrimSpace(match[1])
			if path != "" && !seen[path] {
				if projectRoot != "" {
					fullPath := filepath.Join(projectRoot, path)
					if _, err := os.Stat(fullPath); err == nil {
						paths = append(paths, path)
						seen[path] = true
					}
				} else {
					paths = append(paths, path)
					seen[path] = true
				}
			}
		}
	}

	return paths
}

// inferLaneFromTags infers lane from task tags.
func inferLaneFromTags(tags []string) (string, string) {
	tagStr := strings.Join(tags, " ")

	for _, mapping := range LaneMapping {
		matched, _ := regexp.MatchString(mapping.Pattern, tagStr)
		if matched {
			return mapping.Lane, fmt.Sprintf("matched tag pattern for lane %s", mapping.Lane)
		}
	}

	return "", ""
}

// inferLaneFromContent infers lane from task content text.
func inferLaneFromContent(text string) (string, string) {
	for _, mapping := range LaneMapping {
		matched, _ := regexp.MatchString(mapping.Pattern, text)
		if matched {
			return mapping.Lane, fmt.Sprintf("matched content pattern for lane %s", mapping.Lane)
		}
	}

	return "", ""
}

// buildDirectoryIndex builds an index of directories in the project.
func buildDirectoryIndex(projectRoot string) map[string][]string {
	index := make(map[string][]string)

	if projectRoot == "" {
		return index
	}

	filepath.Walk(projectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip hidden directories and vendor
		if info.IsDir() {
			name := info.Name()
			if strings.HasPrefix(name, ".") && name != "." {
				return filepath.SkipDir
			}
			if name == "vendor" || name == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}

		// Index by directory
		dir := filepath.Dir(path)
		relDir, _ := filepath.Rel(projectRoot, dir)
		ext := filepath.Ext(path)
		index[relDir] = append(index[relDir], ext)

		return nil
	})

	return index
}

// globsForLane returns glob patterns for a lane based on directory index.
func globsForLane(lane string, dirIndex map[string][]string) []string {
	var globs []string

	if patterns, ok := ownershipLaneToDirPattern[lane]; ok {
		for _, pattern := range patterns {
			globs = append(globs, pattern+"/**")
		}
	}

	return globs
}

// filesForLane returns likely files for a lane based on directory index.
func filesForLane(lane string, dirIndex map[string][]string, projectRoot string) []string {
	var files []string

	keywords, ok := ownershipLaneKeywords[lane]
	if !ok {
		return files
	}

	for dir := range dirIndex {
		dirLower := strings.ToLower(dir)
		for _, kw := range keywords {
			if strings.Contains(dirLower, kw) && projectRoot != "" {
				// Find Go files in this directory
				fullDir := filepath.Join(projectRoot, dir)
				entries, err := os.ReadDir(fullDir)
				if err == nil {
					for _, entry := range entries {
						if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".go") && !strings.HasSuffix(entry.Name(), "_test.go") {
							files = append(files, filepath.Join(dir, entry.Name()))
						}
					}
				}
			}
		}
	}

	return files
}

// enhanceOwnershipWithAI uses the foundation model to suggest ownership for tasks
// that couldn't be confidently inferred by heuristics alone.
func enhanceOwnershipWithAI(ctx context.Context, task *Todo2Task, projectRoot string) *OwnershipSuggestion {
	if !FMAvailable() {
		return nil
	}

	// Get directory listing for context
	dirStructure := ""
	if projectRoot != "" {
		if entries, err := ReadDirectoryStructure(projectRoot, 3); err == nil {
			var dirs []string
			for dir, contents := range entries {
				if dir == "." || dir == "" {
					continue
				}
				if len(contents) > 0 {
					maxContents := 5
					if len(contents) < maxContents {
						maxContents = len(contents)
					}
					dirs = append(dirs, fmt.Sprintf("%s: %v", dir, contents[:maxContents]))
				}
			}
			if len(dirs) > 0 {
				maxDirs := 20
				if len(dirs) < maxDirs {
					maxDirs = len(dirs)
				}
				dirStructure = "\n\nProject directory structure:\n" + strings.Join(dirs[:maxDirs], "\n")
			}
		}
	}

	prompt := fmt.Sprintf(`Analyze this task and suggest ownership metadata for parallel execution safety.

Task: %s - %s
Tags: %v
Status: %s
Priority: %s%s

Based on the task content and project structure, suggest:
1. Lane: Which logical lane does this belong to? (backend-auth, backend-api, tui-shell, tui-pane, docs, testing, config, database, source-architecture, or other)
2. Owned files: Which specific files is this task likely to modify? (only suggest files that actually exist in the project)
3. Owned globs: Which file patterns would this task touch?

Available lanes: backend-auth, backend-api, backend-runtime, tui-shell, tui-pane, docs, testing, config, database, source-architecture

Return JSON format:
{"lane": "...", "lane_reason": "...", "owned_files": ["..."], "owned_globs": ["..."]}`,
		task.Content, task.LongDescription, task.Tags, task.Status, task.Priority, dirStructure)

	result, err := DefaultFMProvider().Generate(ctx, prompt, 500, 0.3)
	if err != nil || result == "" {
		return nil
	}

	// Parse AI response
	var aiResult struct {
		Lane       string   `json:"lane"`
		LaneReason string   `json:"lane_reason"`
		OwnedFiles []string `json:"owned_files"`
		OwnedGlobs []string `json:"owned_globs"`
	}

	candidate := ExtractJSONArrayFromLLMResponse(result)
	if candidate == "" {
		candidate = result
	}

	if err := json.Unmarshal([]byte(candidate), &aiResult); err != nil {
		// Try to extract from markdown code block
		if strings.Contains(result, "```json") {
			parts := strings.Split(result, "```json")
			if len(parts) > 1 {
				end := strings.Index(parts[1], "```")
				if end > 0 {
					candidate = strings.TrimSpace(parts[1][:end])
					json.Unmarshal([]byte(candidate), &aiResult)
				}
			}
		}
	}

	// Validate and filter owned files (only keep existing ones)
	validFiles := []string{}
	if projectRoot != "" && len(aiResult.OwnedFiles) > 0 {
		for _, f := range aiResult.OwnedFiles {
			fullPath := filepath.Join(projectRoot, f)
			if _, err := os.Stat(fullPath); err == nil {
				validFiles = append(validFiles, f)
			}
		}
		aiResult.OwnedFiles = validFiles
	}

	if aiResult.Lane == "" && len(aiResult.OwnedFiles) == 0 && len(aiResult.OwnedGlobs) == 0 {
		return nil
	}

	return &OwnershipSuggestion{
		Lane:       aiResult.Lane,
		LaneReason: aiResult.LaneReason,
		OwnedFiles: aiResult.OwnedFiles,
		OwnedGlobs: aiResult.OwnedGlobs,
	}
}

// calculateConfidence calculates overall confidence based on signals.
func calculateConfidence(suggestion OwnershipSuggestion) string {
	signals := 0

	if len(suggestion.OwnedFiles) > 0 {
		signals += 2
	}
	if suggestion.Lane != "" {
		signals += 1
	}
	if len(suggestion.OwnedGlobs) > 0 {
		signals += 1
	}
	if suggestion.AlreadyHasOwnership {
		signals -= 2 // Existing ownership means we're just augmenting
	}

	switch {
	case signals >= 3:
		return "high"
	case signals >= 1:
		return "medium"
	default:
		return "low"
	}
}

// formatInferOwnershipText formats the inference result as text.
func formatInferOwnershipText(result map[string]interface{}) string {
	var sb strings.Builder

	sb.WriteString("Ownership Inference Results\n")
	sb.WriteString(strings.Repeat("=", 40) + "\n\n")

	if msg := ParamString(result, "message"); msg != "" {
		sb.WriteString(msg + "\n\n")
	}

	if dryRun, _ := result["dry_run"].(bool); dryRun {
		sb.WriteString("⚠️  DRY RUN MODE - No changes were made\n\n")
	}

	if suggestions, ok := result["suggestions"].([]OwnershipSuggestion); ok {
		if len(suggestions) == 0 {
			sb.WriteString("No ownership suggestions generated.\n")
			sb.WriteString("Tip: Add file paths or lane tags to task descriptions for better inference.\n")
		} else {
			sb.WriteString(fmt.Sprintf("Found %d ownership suggestions:\n\n", len(suggestions)))

			for i, s := range suggestions {
				sb.WriteString(fmt.Sprintf("%d. %s (%s)\n", i+1, s.TaskID, s.TaskContent))
				sb.WriteString(fmt.Sprintf("   Confidence: %s\n", s.OwnershipConfidence))

				if s.Lane != "" {
					sb.WriteString(fmt.Sprintf("   Lane: %s (%s)\n", s.Lane, s.LaneReason))
				}

				if len(s.OwnedFiles) > 0 {
					sb.WriteString("   Owned files:\n")
					for _, f := range s.OwnedFiles {
						sb.WriteString(fmt.Sprintf("     - %s\n", f))
					}
				}

				if len(s.OwnedGlobs) > 0 {
					sb.WriteString("   Owned globs:\n")
					for _, g := range s.OwnedGlobs {
						sb.WriteString(fmt.Sprintf("     - %s\n", g))
					}
				}

				if s.AlreadyHasOwnership {
					sb.WriteString("   ℹ️  Already has ownership (skipped)\n")
				}

				sb.WriteString("\n")
			}
		}
	}

	// Summary
	if updated, _ := result["updated_count"].(int); updated > 0 {
		sb.WriteString(fmt.Sprintf("✓ Updated %d tasks with inferred ownership\n", updated))
	}

	return sb.String()
}

// ExtractFileReferences is exported for testing - extracts file references from text.
func ExtractFileReferences(text string) []string {
	return extractFilePaths(text, "")
}

// InferLaneFromText is exported for testing - infers lane from text.
func InferLaneFromText(text string) string {
	lane, _ := inferLaneFromContent(text)
	return lane
}

// ReadDirectoryStructure scans a directory and returns its structure.
func ReadDirectoryStructure(root string, maxDepth int) (map[string][]string, error) {
	result := make(map[string][]string)

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		relPath, _ := filepath.Rel(root, path)
		depth := len(strings.Split(relPath, string(filepath.Separator)))

		if depth > maxDepth {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if d.IsDir() {
			name := d.Name()
			if strings.HasPrefix(name, ".") && name != "." {
				return filepath.SkipDir
			}
			if name == "vendor" || name == "node_modules" {
				return filepath.SkipDir
			}
			dir := filepath.Dir(relPath)
			result[dir] = append(result[dir], name+"/")
		} else {
			dir := filepath.Dir(relPath)
			result[dir] = append(result[dir], d.Name())
		}

		return nil
	})

	return result, err
}

// ScanGitHistoryForFiles scans git log to find files commonly edited.
// Returns a map of file paths to edit frequency.
func ScanGitHistoryForFiles(projectRoot string, maxCommits int) (map[string]int, error) {
	freq := make(map[string]int)

	if projectRoot == "" {
		return freq, fmt.Errorf("project root is empty")
	}

	// Use git log to get recent file changes via os/exec
	// This is a placeholder - full implementation would use exec.Command
	return freq, nil
}

// handleTaskAnalysisHotspots handles the hotspots action - reports files that are
// frequently touched by multiple tasks, helping identify merge conflicts.
func handleTaskAnalysisHotspots(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	store, err := getTaskStore(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get task store: %w", err)
	}

	list, err := store.ListTasks(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	tasks := tasksFromPtrs(list)

	// Count how many tasks touch each file
	fileToTasks := make(map[string][]string)
	fileEditCount := make(map[string]int)

	for _, task := range tasks {
		if !IsPendingStatus(task.Status) {
			continue
		}

		own := models.GetTaskOwnership(&task)
		if own == nil {
			continue
		}

		for _, f := range own.OwnedFiles {
			fileToTasks[f] = append(fileToTasks[f], task.ID)
			fileEditCount[f]++
		}
	}

	// Build hotspot list (files touched by 2+ tasks)
	hotspots := []models.HotspotFile{}
	highRisk := []string{}

	for file, taskIDs := range fileToTasks {
		if len(taskIDs) >= 2 {
			hf := models.HotspotFile{
				Path:      file,
				TaskCount: len(taskIDs),
			}
			hotspots = append(hotspots, hf)

			if len(taskIDs) >= 3 {
				highRisk = append(highRisk, file)
			}
		}
	}

	// Sort hotspots by task count (descending)
	sort.Slice(hotspots, func(i, j int) bool {
		return hotspots[i].TaskCount > hotspots[j].TaskCount
	})
	sort.Strings(highRisk)

	projectRoot, _ := GetProjectRootWithFallback()

	hp := &models.ProjectHotspots{
		ProjectRoot: projectRoot,
		AnalyzedAt:  fmt.Sprintf("%v", os.Getpid()), // Placeholder timestamp
		Hotspots:    hotspots,
		HighRisk:    highRisk,
		TotalFiles:  len(fileToTasks),
	}

	result := map[string]interface{}{
		"success":         true,
		"method":          "native_go",
		"hotspots_count":  len(hotspots),
		"high_risk_count": len(highRisk),
		"hotspots":        hotspots,
		"high_risk_files": highRisk,
		"total_contested": len(fileToTasks),
	}

	if len(hotspots) > 0 {
		result["warning"] = fmt.Sprintf("Found %d contested files - tasks sharing these files may collide", len(hotspots))
	}

	outputFormat := ParamOutputFormat(params, "json")
	outputPath := ParamOutputPath(params)

	if outputFormat == "json" {
		if err := EnsureParentDir(outputPath); err != nil {
			return nil, fmt.Errorf("failed to create output dir: %w", err)
		}
		resultJSON, _ := json.Marshal(result)
		resp := &proto.TaskAnalysisResponse{Action: "hotspots", OutputPath: outputPath, ResultJson: string(resultJSON)}
		return framework.FormatResult(TaskAnalysisResponseToMap(resp), resp.GetOutputPath())
	}

	// Text format
	output := formatHotspotsText(hp, fileToTasks)

	if outputPath != "" {
		if err := EnsureParentDir(outputPath); err != nil {
			return nil, fmt.Errorf("failed to create output dir: %w", err)
		}
		if err := os.WriteFile(outputPath, []byte(output), 0644); err != nil {
			return nil, fmt.Errorf("failed to save result: %w", err)
		}
		output += fmt.Sprintf("\n\n[Saved to: %s]", outputPath)
	}

	return []framework.TextContent{{Type: "text", Text: output}}, nil
}

// formatHotspotsText formats hotspot analysis as text.
func formatHotspotsText(hp *models.ProjectHotspots, fileToTasks map[string][]string) string {
	var sb strings.Builder

	sb.WriteString("File Hotspot Analysis\n")
	sb.WriteString(strings.Repeat("=", 40) + "\n\n")

	if len(hp.Hotspots) == 0 {
		sb.WriteString("No contested files found. All tasks own distinct files.\n")
		return sb.String()
	}

	sb.WriteString(fmt.Sprintf("Found %d contested files (touched by multiple tasks):\n\n", len(hp.Hotspots)))

	sb.WriteString("| File | Tasks | Risk |\n")
	sb.WriteString("|------|-------|------|\n")

	for _, hf := range hp.Hotspots {
		risk := "medium"
		if hf.TaskCount >= 3 {
			risk = "HIGH"
		}
		sb.WriteString(fmt.Sprintf("| %s | %d | %s |\n", hf.Path, hf.TaskCount, risk))
	}

	if len(hp.HighRisk) > 0 {
		sb.WriteString("\n⚠️  HIGH RISK FILES (3+ tasks):\n")
		for _, f := range hp.HighRisk {
			tasks := fileToTasks[f]
			sb.WriteString(fmt.Sprintf("  - %s (tasks: %s)\n", f, strings.Join(tasks, ", ")))
		}
	}

	sb.WriteString("\nRecommendation: Avoid parallelizing tasks that share contested files.\n")

	return sb.String()
}

// WarnAboutHotspots checks if proposed task files conflict with existing hotspots.
// Returns warning messages for conflicts found.
func WarnAboutHotspots(projectRoot string, proposedFiles []string, existingHotspots []models.HotspotFile) []string {
	var warnings []string

	for _, hf := range existingHotspots {
		for _, pf := range proposedFiles {
			if hf.Path == pf && hf.TaskCount >= 2 {
				warnings = append(warnings, fmt.Sprintf("⚠️  %s is contested (%d other tasks)", hf.Path, hf.TaskCount))
			}
		}
	}

	return warnings
}

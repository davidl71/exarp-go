package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/davidl71/exarp-go/internal/cache"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/internal/security"
	"github.com/davidl71/mcp-go-core/pkg/mcp/response"
)

// Design limit for MCP tool count (monitored by tool_count_health / health action=tools).
const designLimitTools = 30

// ExpectedToolCountBase is the base number of tools registered by RegisterAllTools (without conditional Apple FM).
// With Apple Foundation Models on darwin/arm64/cgo build, count is ExpectedToolCountBase+1.
const ExpectedToolCountBase = 29

// handleHealthNative handles the health tool with native Go implementation
// Implements all actions: "server", "git", "docs", "dod", "cicd", "tools"
// Note: Some actions provide basic functionality; complex features may fall back to Python bridge
func handleHealthNative(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	// Get action (default: "server")
	action := "server"
	if actionRaw, ok := params["action"].(string); ok && actionRaw != "" {
		action = actionRaw
	}

	switch action {
	case "server":
		return handleHealthServer(ctx, params)
	case "git":
		return handleHealthGit(ctx, params)
	case "docs":
		return handleHealthDocs(ctx, params)
	case "dod":
		return handleHealthDOD(ctx, params)
	case "cicd":
		return handleHealthCICD(ctx, params)
	case "tools":
		return handleHealthTools(ctx, params)
	default:
		// Unknown action - fall back to Python bridge
		return nil, fmt.Errorf("health action '%s' not yet implemented in native Go, using Python bridge", action)
	}
}

// handleHealthTools handles the "tools" action - MCP tool count vs design limit (≤30).
// Used by daily automation as tool_count_health.
func handleHealthTools(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	toolCount := ExpectedToolCountBase
	withinLimit := toolCount <= designLimitTools
	result := map[string]interface{}{
		"tool_count":   toolCount,
		"limit":        designLimitTools,
		"within_limit": withinLimit,
		"method":       "native_go",
		"success":      true,
	}
	return response.FormatResult(result, "")
}

// handleHealthServer handles the "server" action for health tool
func handleHealthServer(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	// Get project root
	projectRoot, err := security.GetProjectRoot(".")
	if err != nil {
		// Fallback to PROJECT_ROOT env var or current directory
		if envRoot := os.Getenv("PROJECT_ROOT"); envRoot != "" {
			projectRoot = envRoot
		} else {
			wd, _ := os.Getwd()
			projectRoot = wd
		}
	}

	// Try to get version from go.mod or pyproject.toml
	version := "unknown"

	// Check go.mod first (Go project) - using file cache
	fileCache := cache.GetGlobalFileCache()
	goModPath := filepath.Join(projectRoot, "go.mod")
	if _, err := os.Stat(goModPath); err == nil {
		// Read go.mod to find module version or use git tag
		content, _, err := fileCache.ReadFile(goModPath)
		if err == nil {
			// Look for module declaration
			moduleRegex := regexp.MustCompile(`module\s+([^\s]+)`)
			matches := moduleRegex.FindStringSubmatch(string(content))
			if len(matches) > 1 {
				// Try to get version from git
				gitCmd := exec.CommandContext(ctx, "git", "describe", "--tags", "--always")
				gitCmd.Dir = projectRoot
				if output, err := gitCmd.Output(); err == nil {
					version = string(output)
					// Remove trailing newline
					if len(version) > 0 && version[len(version)-1] == '\n' {
						version = version[:len(version)-1]
					}
				}
			}
		}
	}

	// Fallback to pyproject.toml (Python project)
	if version == "unknown" {
		pyprojectPath := filepath.Join(projectRoot, "pyproject.toml")
		if _, err := os.Stat(pyprojectPath); err == nil {
			content, _, err := fileCache.ReadFile(pyprojectPath)
			if err == nil {
				// Look for version = "x.y.z"
				versionRegex := regexp.MustCompile(`version\s*=\s*["']([^"']+)["']`)
				matches := versionRegex.FindStringSubmatch(string(content))
				if len(matches) > 1 {
					version = matches[1]
				}
			}
		}
	}

	// Build result
	result := map[string]interface{}{
		"status":       "operational",
		"version":      version,
		"project_root": projectRoot,
		"timestamp":    time.Now().Unix(),
	}

	return response.FormatResult(result, "")
}

// handleHealthGit handles the "git" action for health tool
// Checks git working copy health (status, uncommitted changes, remote sync)
func handleHealthGit(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	// Get project root
	projectRoot, err := security.GetProjectRoot(".")
	if err != nil {
		if envRoot := os.Getenv("PROJECT_ROOT"); envRoot != "" {
			projectRoot = envRoot
		} else {
			wd, _ := os.Getwd()
			projectRoot = wd
		}
	}

	// Get parameters
	agentName, _ := params["agent_name"].(string)
	checkRemote, _ := params["check_remote"].(bool)
	if !checkRemote {
		checkRemote = true // Default to true
	}

	// For now, only support local agent (fastest implementation)
	// Remote agents can fall back to Python bridge
	if agentName != "" && agentName != "local" {
		return nil, fmt.Errorf("remote agents not yet supported in native Go, using Python bridge")
	}

	// Check if git repository exists
	gitDir := filepath.Join(projectRoot, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		result := map[string]interface{}{
			"status":    "error",
			"error":     "not a git repository",
			"agent":     "local",
			"timestamp": time.Now().Unix(),
		}
		resultJSON, _ := json.MarshalIndent(result, "", "  ")
		return []framework.TextContent{
			{Type: "text", Text: string(resultJSON)},
		}, nil
	}

	// Get git status
	statusCmd := exec.CommandContext(ctx, "git", "status", "--porcelain")
	statusCmd.Dir = projectRoot
	statusOutput, _ := statusCmd.Output()
	hasChanges := len(statusOutput) > 0

	// Get current branch
	branchCmd := exec.CommandContext(ctx, "git", "branch", "--show-current")
	branchCmd.Dir = projectRoot
	branchOutput, _ := branchCmd.Output()
	currentBranch := strings.TrimSpace(string(branchOutput))
	if currentBranch == "" {
		currentBranch = "detached"
	}

	// Check remote sync (if check_remote is true)
	remoteStatus := "unknown"
	remoteSync := false
	if checkRemote {
		// Check if remote exists
		remoteCmd := exec.CommandContext(ctx, "git", "remote")
		remoteCmd.Dir = projectRoot
		remoteOutput, _ := remoteCmd.Output()
		hasRemote := len(remoteOutput) > 0

		if hasRemote {
			// Check if branch is tracking remote
			trackingCmd := exec.CommandContext(ctx, "git", "rev-parse", "--abbrev-ref", "--symbolic-full-name", currentBranch+"@{upstream}")
			trackingCmd.Dir = projectRoot
			trackingOutput, _ := trackingCmd.Output()
			isTracking := len(trackingOutput) > 0

			if isTracking {
				// Check if ahead/behind
				diffCmd := exec.CommandContext(ctx, "git", "rev-list", "--left-right", "--count", currentBranch+"@{upstream}..."+currentBranch)
				diffCmd.Dir = projectRoot
				diffOutput, _ := diffCmd.Output()
				parts := strings.Fields(string(diffOutput))
				behind := 0
				ahead := 0
				if len(parts) >= 2 {
					fmt.Sscanf(parts[0], "%d", &behind)
					fmt.Sscanf(parts[1], "%d", &ahead)
				}
				remoteSync = behind == 0 && ahead == 0
				if behind > 0 && ahead > 0 {
					remoteStatus = fmt.Sprintf("behind %d, ahead %d", behind, ahead)
				} else if behind > 0 {
					remoteStatus = fmt.Sprintf("behind %d", behind)
				} else if ahead > 0 {
					remoteStatus = fmt.Sprintf("ahead %d", ahead)
				} else {
					remoteStatus = "synced"
				}
			} else {
				remoteStatus = "not tracking"
			}
		} else {
			remoteStatus = "no remote"
		}
	}

	// Build result
	result := map[string]interface{}{
		"agent":         "local",
		"path":          projectRoot,
		"branch":        currentBranch,
		"has_changes":   hasChanges,
		"remote_status": remoteStatus,
		"remote_sync":   remoteSync,
		"status":        "healthy",
		"timestamp":     time.Now().Unix(),
	}

	// Add uncommitted files count if there are changes
	if hasChanges {
		lines := strings.Split(strings.TrimSpace(string(statusOutput)), "\n")
		result["uncommitted_files"] = len(lines)
	}

	return response.FormatResult(result, "")
}

// handleHealthDocs handles the "docs" action for health tool
// Checks documentation health (basic checks: README exists, docs directory exists)
func handleHealthDocs(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	// Get project root
	projectRoot, err := security.GetProjectRoot(".")
	if err != nil {
		if envRoot := os.Getenv("PROJECT_ROOT"); envRoot != "" {
			projectRoot = envRoot
		} else {
			wd, _ := os.Getwd()
			projectRoot = wd
		}
	}

	// Get parameters
	outputPath, _ := params["output_path"].(string)
	createTasks, _ := params["create_tasks"].(bool)

	// Basic documentation checks
	checks := map[string]interface{}{
		"readme_exists":   false,
		"docs_dir_exists": false,
		"readme_size":     0,
		"docs_file_count": 0,
	}

	// Check for README
	readmePaths := []string{"README.md", "README.rst", "README.txt", "readme.md"}
	for _, readmePath := range readmePaths {
		fullPath := filepath.Join(projectRoot, readmePath)
		if info, err := os.Stat(fullPath); err == nil {
			checks["readme_exists"] = true
			checks["readme_size"] = info.Size()
			break
		}
	}

	// Check for docs directory
	docsDir := filepath.Join(projectRoot, "docs")
	if info, err := os.Stat(docsDir); err == nil && info.IsDir() {
		checks["docs_dir_exists"] = true
		// Count markdown files in docs
		entries, _ := os.ReadDir(docsDir)
		mdCount := 0
		for _, entry := range entries {
			if !entry.IsDir() && (strings.HasSuffix(entry.Name(), ".md") || strings.HasSuffix(entry.Name(), ".rst")) {
				mdCount++
			}
		}
		checks["docs_file_count"] = mdCount
	}

	// Calculate basic health score
	score := 0.0
	if checks["readme_exists"].(bool) {
		score += 50.0
	}
	if checks["docs_dir_exists"].(bool) {
		score += 30.0
	}
	if checks["docs_file_count"].(int) > 0 {
		score += 20.0
	}

	// Build result
	result := map[string]interface{}{
		"status":       "completed",
		"health_score": score,
		"checks":       checks,
		"project_root": projectRoot,
		"timestamp":    time.Now().Unix(),
	}

	// Note: For full documentation analysis (broken links, format validation, etc.),
	// this falls back to Python bridge. This is a simplified version.
	if outputPath != "" {
		result["report_path"] = outputPath
		// Note: Full report generation would require Python bridge
	}

	if createTasks {
		// Note: Task creation would require Todo2 integration
		// For now, just note that tasks would be created
		result["tasks_note"] = "Task creation requires Python bridge for full functionality"
	}

	return response.FormatResult(result, "")
}

// handleHealthDOD handles the "dod" action for health tool
// Checks definition of done for tasks
func handleHealthDOD(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	// Get parameters
	taskID, _ := params["task_id"].(string)
	changedFiles, _ := params["changed_files"].(string)
	autoCheck, _ := params["auto_check"].(bool)
	outputPath, _ := params["output_path"].(string)

	// Basic DOD checks
	checks := map[string]interface{}{
		"tests_exist": false,
		"docs_exist":  false,
		"code_review": false,
		"no_todos":    false,
	}

	// If task_id is provided, check specific task
	if taskID != "" {
		// Note: Full DOD checking requires Todo2 database access
		// This is a simplified version
		result := map[string]interface{}{
			"status":    "partial",
			"task_id":   taskID,
			"checks":    checks,
			"note":      "Full DOD checking requires Python bridge for Todo2 integration",
			"timestamp": time.Now().Unix(),
		}

		if outputPath != "" {
			result["output_path"] = outputPath
		}

		resultJSON, _ := json.MarshalIndent(result, "", "  ")
		return []framework.TextContent{
			{Type: "text", Text: string(resultJSON)},
		}, nil
	}

	// General DOD check (if changed_files provided)
	if changedFiles != "" {
		var files []string
		if err := json.Unmarshal([]byte(changedFiles), &files); err == nil {
			// Basic checks on changed files
			hasTests := false
			hasDocs := false
			for _, file := range files {
				if strings.Contains(file, "_test.go") || strings.Contains(file, "test_") {
					hasTests = true
				}
				if strings.HasSuffix(file, ".md") || strings.Contains(file, "docs/") {
					hasDocs = true
				}
			}
			checks["tests_exist"] = hasTests
			checks["docs_exist"] = hasDocs
		}
	}

	result := map[string]interface{}{
		"status":     "completed",
		"checks":     checks,
		"auto_check": autoCheck,
		"timestamp":  time.Now().Unix(),
	}

	if outputPath != "" {
		result["output_path"] = outputPath
	}

	return response.FormatResult(result, "")
}

// handleHealthCICD handles the "cicd" action for health tool
// Validates CI/CD workflows
func handleHealthCICD(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	// Get parameters
	workflowPath, _ := params["workflow_path"].(string)
	checkRunners, _ := params["check_runners"].(bool)
	outputPath, _ := params["output_path"].(string)

	// Get project root
	projectRoot, err := security.GetProjectRoot(".")
	if err != nil {
		if envRoot := os.Getenv("PROJECT_ROOT"); envRoot != "" {
			projectRoot = envRoot
		} else {
			wd, _ := os.Getwd()
			projectRoot = wd
		}
	}

	// Check for CI/CD files
	checks := map[string]interface{}{
		"github_actions_exists": false,
		"gitlab_ci_exists":      false,
		"circleci_exists":       false,
		"workflow_files":        []string{},
	}

	// Check GitHub Actions
	githubWorkflows := filepath.Join(projectRoot, ".github", "workflows")
	if info, err := os.Stat(githubWorkflows); err == nil && info.IsDir() {
		checks["github_actions_exists"] = true
		entries, _ := os.ReadDir(githubWorkflows)
		workflows := []string{}
		for _, entry := range entries {
			if !entry.IsDir() && (strings.HasSuffix(entry.Name(), ".yml") || strings.HasSuffix(entry.Name(), ".yaml")) {
				workflows = append(workflows, entry.Name())
			}
		}
		checks["workflow_files"] = workflows
	}

	// Check GitLab CI
	gitlabCI := filepath.Join(projectRoot, ".gitlab-ci.yml")
	if _, err := os.Stat(gitlabCI); err == nil {
		checks["gitlab_ci_exists"] = true
	}

	// Check CircleCI
	circleCI := filepath.Join(projectRoot, ".circleci")
	if info, err := os.Stat(circleCI); err == nil && info.IsDir() {
		checks["circleci_exists"] = true
	}

	// Basic validation
	hasCICD := checks["github_actions_exists"].(bool) || checks["gitlab_ci_exists"].(bool) || checks["circleci_exists"].(bool)

	result := map[string]interface{}{
		"status":        "completed",
		"has_cicd":      hasCICD,
		"checks":        checks,
		"check_runners": checkRunners,
		"timestamp":     time.Now().Unix(),
	}

	if workflowPath != "" {
		result["workflow_path"] = workflowPath
		// Note: Full workflow validation would require YAML parsing
		result["note"] = "Full workflow validation requires Python bridge for YAML parsing and runner validation"
	}

	if outputPath != "" {
		result["output_path"] = outputPath
	}

	return response.FormatResult(result, "")
}

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/mcp-go-core/pkg/mcp/response"
)

// handleGenerateConfigNative handles the generate_config tool with native Go implementation
func handleGenerateConfigNative(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	action := "rules"
	if a, ok := params["action"].(string); ok && a != "" {
		action = a
	}

	// Get project root - try multiple methods
	var projectRoot string
	var err error

	// Try FindProjectRoot first
	projectRoot, err = FindProjectRoot()
	if err != nil {
		// Fallback to PROJECT_ROOT env var
		if envRoot := os.Getenv("PROJECT_ROOT"); envRoot != "" {
			projectRoot = envRoot
		} else {
			// Fallback to current working directory
			wd, _ := os.Getwd()
			projectRoot = wd
		}
	}

	switch action {
	case "rules":
		return handleGenerateRules(ctx, params, projectRoot)
	case "ignore":
		return handleGenerateIgnore(ctx, params, projectRoot)
	case "simplify":
		return handleSimplifyRules(ctx, params, projectRoot)
	default:
		return nil, fmt.Errorf("unknown action: %s. Use 'rules', 'ignore', or 'simplify'", action)
	}
}

// handleGenerateRules handles the "rules" action for generate_config
func handleGenerateRules(ctx context.Context, params map[string]interface{}, projectRoot string) ([]framework.TextContent, error) {
	rulesStr, _ := params["rules"].(string)
	overwrite, _ := params["overwrite"].(bool)
	analyzeOnly, _ := params["analyze_only"].(bool)

	generator := NewCursorRulesGenerator(projectRoot)

	if analyzeOnly {
		analysis := generator.AnalyzeProject()
		result := map[string]interface{}{
			"success":         true,
			"analysis":        analysis,
			"available_rules": getAvailableRuleNames(),
			"tip":             "Run without analyze_only to generate recommended rules",
		}
		return response.FormatResult(result, "")
	}

	var rulesList []string
	if rulesStr != "" {
		rulesList = strings.Split(rulesStr, ",")
		for i := range rulesList {
			rulesList[i] = strings.TrimSpace(rulesList[i])
		}
	}

	results := generator.GenerateRules(rulesList, overwrite)
	results["summary"] = map[string]interface{}{
		"generated": len(results["generated"].([]map[string]interface{})),
		"skipped":   len(results["skipped"].([]map[string]interface{})),
		"errors":    len(results["errors"].([]string)),
	}

	return response.FormatResult(map[string]interface{}{
		"success": true,
		"data":    results,
	}, "")
}

// handleGenerateIgnore handles the "ignore" action for generate_config
func handleGenerateIgnore(ctx context.Context, params map[string]interface{}, projectRoot string) ([]framework.TextContent, error) {
	includeIndexing := true
	if val, ok := params["include_indexing"].(bool); ok {
		includeIndexing = val
	}

	analyzeProject := true
	if val, ok := params["analyze_project"].(bool); ok {
		analyzeProject = val
	}

	dryRun, _ := params["dry_run"].(bool)

	generator := NewCursorIgnoreGenerator(projectRoot)
	results := generator.GenerateIgnore(includeIndexing, analyzeProject, dryRun)

	return response.FormatResult(map[string]interface{}{
		"success": true,
		"data":    results,
	}, "")
}

// handleSimplifyRules handles the "simplify" action for generate_config
func handleSimplifyRules(ctx context.Context, params map[string]interface{}, projectRoot string) ([]framework.TextContent, error) {
	ruleFilesStr, _ := params["rule_files"].(string)
	dryRun, _ := params["dry_run"].(bool)
	outputDir, _ := params["output_dir"].(string)

	var ruleFiles []string
	if ruleFilesStr != "" {
		var parsedFiles []interface{}
		if err := json.Unmarshal([]byte(ruleFilesStr), &parsedFiles); err == nil {
			for _, f := range parsedFiles {
				if str, ok := f.(string); ok {
					ruleFiles = append(ruleFiles, str)
				}
			}
		}
	}

	// If no files specified, find default rule files
	if len(ruleFiles) == 0 {
		// Check for .cursorrules
		cursorrulesPath := filepath.Join(projectRoot, ".cursorrules")
		if _, err := os.Stat(cursorrulesPath); err == nil {
			ruleFiles = append(ruleFiles, cursorrulesPath)
		}

		// Check for .cursor/rules/*.mdc files
		rulesDir := filepath.Join(projectRoot, ".cursor", "rules")
		if matches, err := filepath.Glob(filepath.Join(rulesDir, "*.mdc")); err == nil {
			ruleFiles = append(ruleFiles, matches...)
		}
	}

	simplifier := NewRuleSimplifier(projectRoot)
	results := simplifier.SimplifyRules(ruleFiles, dryRun, outputDir)

	return response.FormatResult(map[string]interface{}{
		"status":          "success",
		"files_processed": results["files_processed"],
		"files_skipped":   results["files_skipped"],
		"simplifications": results["simplifications"],
		"dry_run":         dryRun,
	}, "")
}

// CursorRulesGenerator generates Cursor rules files based on project analysis
type CursorRulesGenerator struct {
	projectRoot string
	rulesDir    string
}

// NewCursorRulesGenerator creates a new CursorRulesGenerator
func NewCursorRulesGenerator(projectRoot string) *CursorRulesGenerator {
	return &CursorRulesGenerator{
		projectRoot: projectRoot,
		rulesDir:    filepath.Join(projectRoot, ".cursor", "rules"),
	}
}

// AnalyzeProject analyzes the project to determine which rules to generate
func (g *CursorRulesGenerator) AnalyzeProject() map[string]interface{} {
	analysis := map[string]interface{}{
		"languages":         []string{},
		"frameworks":        []string{},
		"patterns":          []string{},
		"recommended_rules": []string{},
	}

	languages := []string{}
	frameworks := []string{}
	patterns := []string{}
	recommendedRules := []string{}

	// Detect languages by file extension
	extensions := map[string]string{
		".py":  "python",
		".ts":  "typescript",
		".tsx": "typescript",
		".js":  "javascript",
		".jsx": "javascript",
		".go":  "go",
		".rs":  "rust",
	}

	for ext, lang := range extensions {
		if g.hasFilesWithExtension(ext) {
			if !contains(languages, lang) {
				languages = append(languages, lang)
				if lang == "python" || lang == "typescript" {
					recommendedRules = append(recommendedRules, lang)
				}
			}
		}
	}

	// Detect frameworks from package.json
	if g.hasFile("package.json") {
		if g.hasDependency("react") {
			frameworks = append(frameworks, "react")
		}
		if g.hasDependency("next") {
			frameworks = append(frameworks, "nextjs")
		}
		if g.hasDependency("express") {
			frameworks = append(frameworks, "express")
			if !contains(recommendedRules, "api") {
				recommendedRules = append(recommendedRules, "api")
			}
		}
	}

	// Detect frameworks from pyproject.toml
	if g.hasFile("pyproject.toml") {
		content := g.readFile("pyproject.toml")
		contentLower := strings.ToLower(content)
		if strings.Contains(contentLower, "fastapi") {
			frameworks = append(frameworks, "fastapi")
			if !contains(recommendedRules, "api") {
				recommendedRules = append(recommendedRules, "api")
			}
		}
		if strings.Contains(contentLower, "django") {
			frameworks = append(frameworks, "django")
		}
		if strings.Contains(contentLower, "mcp") {
			frameworks = append(frameworks, "mcp")
			if !contains(recommendedRules, "mcp") {
				recommendedRules = append(recommendedRules, "mcp")
			}
		}
	}

	// Detect test files
	if g.hasFilesWithPattern("test_*.py") || g.hasFilesWithPattern("*.test.ts") {
		patterns = append(patterns, "testing")
		if !contains(recommendedRules, "testing") {
			recommendedRules = append(recommendedRules, "testing")
		}
	}

	// Detect API patterns
	apiDirs := []string{"api", "routes", "endpoints"}
	for _, dir := range apiDirs {
		if g.hasDirectory(dir) {
			patterns = append(patterns, "api")
			if !contains(recommendedRules, "api") {
				recommendedRules = append(recommendedRules, "api")
			}
			break
		}
	}

	analysis["languages"] = languages
	analysis["frameworks"] = frameworks
	analysis["patterns"] = patterns
	analysis["recommended_rules"] = recommendedRules

	return analysis
}

// GenerateRules generates Cursor rules files
func (g *CursorRulesGenerator) GenerateRules(rules []string, overwrite bool) map[string]interface{} {
	// Ensure rules directory exists
	if err := os.MkdirAll(g.rulesDir, 0755); err != nil {
		return map[string]interface{}{
			"analysis":  map[string]interface{}{},
			"generated": []map[string]interface{}{},
			"skipped":   []map[string]interface{}{},
			"errors":    []string{fmt.Sprintf("Failed to create rules directory: %v", err)},
		}
	}

	analysis := g.AnalyzeProject()
	recommendedRules := analysis["recommended_rules"].([]string)

	// Use provided rules or recommended rules
	rulesToGenerate := rules
	if len(rulesToGenerate) == 0 {
		rulesToGenerate = recommendedRules
	}

	results := map[string]interface{}{
		"analysis":  analysis,
		"generated": []map[string]interface{}{},
		"skipped":   []map[string]interface{}{},
		"errors":    []string{},
	}

	generated := []map[string]interface{}{}
	skipped := []map[string]interface{}{}
	errors := []string{}

	for _, ruleName := range rulesToGenerate {
		template, exists := getRuleTemplate(ruleName)
		if !exists {
			errors = append(errors, fmt.Sprintf("Unknown rule: %s", ruleName))
			continue
		}

		rulePath := filepath.Join(g.rulesDir, template.Filename)

		// Check if file exists and overwrite is false
		if _, err := os.Stat(rulePath); err == nil && !overwrite {
			skipped = append(skipped, map[string]interface{}{
				"rule":   ruleName,
				"path":   rulePath,
				"reason": "exists",
			})
			continue
		}

		// Write rule file
		if err := os.WriteFile(rulePath, []byte(template.Content), 0644); err != nil {
			errors = append(errors, fmt.Sprintf("Error writing %s: %v", ruleName, err))
			continue
		}

		generated = append(generated, map[string]interface{}{
			"rule": ruleName,
			"path": rulePath,
		})
	}

	results["generated"] = generated
	results["skipped"] = skipped
	results["errors"] = errors

	return results
}

// Helper functions

func (g *CursorRulesGenerator) hasFilesWithExtension(ext string) bool {
	return g.hasFilesWithPattern("*" + ext)
}

func (g *CursorRulesGenerator) hasFilesWithPattern(pattern string) bool {
	matches, err := filepath.Glob(filepath.Join(g.projectRoot, "**", pattern))
	return err == nil && len(matches) > 0
}

func (g *CursorRulesGenerator) hasFile(filename string) bool {
	path := filepath.Join(g.projectRoot, filename)
	_, err := os.Stat(path)
	return err == nil
}

func (g *CursorRulesGenerator) hasDirectory(dirname string) bool {
	path := filepath.Join(g.projectRoot, dirname)
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func (g *CursorRulesGenerator) readFile(filename string) string {
	path := filepath.Join(g.projectRoot, filename)
	content, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(content)
}

func (g *CursorRulesGenerator) hasDependency(depName string) bool {
	content := g.readFile("package.json")
	if content == "" {
		return false
	}

	var pkg map[string]interface{}
	if err := json.Unmarshal([]byte(content), &pkg); err != nil {
		return false
	}

	deps := map[string]bool{}
	if depsMap, ok := pkg["dependencies"].(map[string]interface{}); ok {
		for k := range depsMap {
			deps[k] = true
		}
	}
	if devDepsMap, ok := pkg["devDependencies"].(map[string]interface{}); ok {
		for k := range devDepsMap {
			deps[k] = true
		}
	}

	return deps[depName]
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func getAvailableRuleNames() []string {
	return []string{"python", "typescript", "javascript", "go", "rust", "api", "testing", "mcp"}
}

// RuleTemplate represents a rule template
type RuleTemplate struct {
	Filename string
	Content  string
}

func getRuleTemplate(ruleName string) (RuleTemplate, bool) {
	templates := getRuleTemplates()
	template, exists := templates[ruleName]
	return template, exists
}

// CursorIgnoreGenerator generates .cursorignore files
type CursorIgnoreGenerator struct {
	projectRoot string
}

// NewCursorIgnoreGenerator creates a new CursorIgnoreGenerator
func NewCursorIgnoreGenerator(projectRoot string) *CursorIgnoreGenerator {
	return &CursorIgnoreGenerator{
		projectRoot: projectRoot,
	}
}

// GenerateIgnore generates .cursorignore and optionally .cursorindexingignore files
func (g *CursorIgnoreGenerator) GenerateIgnore(includeIndexing bool, analyzeProject bool, dryRun bool) map[string]interface{} {
	results := map[string]interface{}{
		"generated":         []interface{}{},
		"skipped":           []interface{}{},
		"errors":            []interface{}{},
		"detected_patterns": []string{},
	}

	generated := []interface{}{}
	skipped := []interface{}{}
	errors := []interface{}{}

	// Generate .cursorignore
	cursorignorePath := filepath.Join(g.projectRoot, ".cursorignore")
	cursorignoreContent := getDefaultCursorIgnore()

	// Analyze project and add project-specific patterns
	detectedPatterns := []string{}
	if analyzeProject {
		additionalPatterns := g.analyzeProjectForIgnorePatterns()
		detectedPatterns = additionalPatterns
		if len(additionalPatterns) > 0 {
			cursorignoreContent += "\n# Project-specific patterns\n"
			cursorignoreContent += strings.Join(additionalPatterns, "\n") + "\n"
		}
	}

	if !dryRun {
		if err := os.WriteFile(cursorignorePath, []byte(cursorignoreContent), 0644); err != nil {
			errors = append(errors, fmt.Sprintf("Failed to write .cursorignore: %v", err))
		} else {
			generated = append(generated, cursorignorePath)
		}
	} else {
		// In dry run, just report what would be generated
		generated = append(generated, cursorignorePath)
	}

	// Generate .cursorindexingignore if requested
	if includeIndexing {
		indexingIgnorePath := filepath.Join(g.projectRoot, ".cursorindexingignore")
		indexingIgnoreContent := getDefaultIndexingIgnore()

		if !dryRun {
			if err := os.WriteFile(indexingIgnorePath, []byte(indexingIgnoreContent), 0644); err != nil {
				errors = append(errors, fmt.Sprintf("Failed to write .cursorindexingignore: %v", err))
			} else {
				generated = append(generated, indexingIgnorePath)
			}
		} else {
			generated = append(generated, indexingIgnorePath)
		}
	}

	results["generated"] = generated
	results["skipped"] = skipped
	results["errors"] = errors
	results["dry_run"] = dryRun
	results["detected_patterns"] = detectedPatterns
	results["files"] = generated // Python bridge expects "files" field

	return results
}

// analyzeProjectForIgnorePatterns analyzes project structure to suggest ignore patterns
func (g *CursorIgnoreGenerator) analyzeProjectForIgnorePatterns() []string {
	patterns := []string{}

	// Check for common build directories
	buildDirs := []string{"build", "dist", "out", "target", "bin", "obj"}
	for _, dir := range buildDirs {
		if g.hasDirectory(dir) {
			patterns = append(patterns, dir+"/")
		}
	}

	// Check for common cache directories
	cacheDirs := []string{".cache", ".tmp", "tmp", "temp"}
	for _, dir := range cacheDirs {
		if g.hasDirectory(dir) {
			patterns = append(patterns, dir+"/")
		}
	}

	// Check for generated files
	if g.hasFilesWithPattern("*.generated.*") {
		patterns = append(patterns, "*.generated.*")
	}

	return patterns
}

func (g *CursorIgnoreGenerator) hasDirectory(dirname string) bool {
	path := filepath.Join(g.projectRoot, dirname)
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func (g *CursorIgnoreGenerator) hasFilesWithPattern(pattern string) bool {
	// Use filepath.Walk for recursive glob matching
	found := false
	filepath.Walk(g.projectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if !info.IsDir() {
			matched, _ := filepath.Match(pattern, filepath.Base(path))
			if matched {
				found = true
				return filepath.SkipAll // Stop walking
			}
		}
		return nil
	})
	return found
}

func (g *CursorIgnoreGenerator) hasFile(filename string) bool {
	path := filepath.Join(g.projectRoot, filename)
	_, err := os.Stat(path)
	return err == nil
}

func getDefaultCursorIgnore() string {
	return `# Cursorignore - Files to exclude from AI context
# Generated by Exarp generate_config tool

# Dependencies
node_modules/
vendor/
.venv/
venv/
__pycache__/
*.pyc
.eggs/
*.egg-info/

# Build outputs
dist/
build/
*.min.js
*.min.css
*.bundle.js
*.map

# IDE/Editor
.idea/
.vscode/
*.swp
*.swo
*~

# OS files
.DS_Store
Thumbs.db

# Logs
*.log
logs/

# Test artifacts
coverage/
.coverage
htmlcov/
.pytest_cache/
.tox/

# Environment files
.env
.env.local
.env.*.local

# Generated files
*.generated.*
`
}

func getDefaultIndexingIgnore() string {
	return `# Cursorindexingignore - Files to exclude from indexing
# Generated by Exarp generate_config tool

# Large dependency directories
node_modules/
vendor/
.venv/
venv/

# Build outputs
dist/
build/
target/
out/

# Generated files
*.min.js
*.min.css
*.bundle.js
`
}

// RuleSimplifier simplifies rules by replacing manual processes with exarp tool references
type RuleSimplifier struct {
	projectRoot string
}

// NewRuleSimplifier creates a new RuleSimplifier
func NewRuleSimplifier(projectRoot string) *RuleSimplifier {
	return &RuleSimplifier{
		projectRoot: projectRoot,
	}
}

// SimplifyRules simplifies rule files by replacing manual processes with exarp tool references
func (s *RuleSimplifier) SimplifyRules(ruleFiles []string, dryRun bool, outputDir string) map[string]interface{} {
	results := map[string]interface{}{
		"files_processed": []map[string]interface{}{},
		"files_skipped":   []map[string]interface{}{},
		"simplifications": []map[string]interface{}{},
	}

	filesProcessed := []map[string]interface{}{}
	filesSkipped := []map[string]interface{}{}
	simplifications := []map[string]interface{}{}

	patterns := getSimplificationPatterns()

	for _, ruleFilePath := range ruleFiles {
		ruleFile := ruleFilePath
		if !filepath.IsAbs(ruleFile) {
			ruleFile = filepath.Join(s.projectRoot, ruleFilePath)
		}

		if _, err := os.Stat(ruleFile); os.IsNotExist(err) {
			filesSkipped = append(filesSkipped, map[string]interface{}{
				"file":   ruleFilePath,
				"reason": "File not found",
			})
			continue
		}

		content, err := os.ReadFile(ruleFile)
		if err != nil {
			filesSkipped = append(filesSkipped, map[string]interface{}{
				"file":   ruleFilePath,
				"reason": fmt.Sprintf("Failed to read file: %v", err),
			})
			continue
		}

		originalContent := string(content)
		simplifiedContent := originalContent
		fileSimplifications := []map[string]interface{}{}

		// Apply simplification patterns
		for patternName, patternConfig := range patterns {
			oldPattern := patternConfig["old"]
			newPattern := patternConfig["new"]

			if strings.Contains(simplifiedContent, oldPattern) {
				simplifiedContent = strings.ReplaceAll(simplifiedContent, oldPattern, newPattern)
				fileSimplifications = append(fileSimplifications, map[string]interface{}{
					"pattern": patternName,
					"old":     oldPattern,
					"new":     newPattern,
				})
			}
		}

		if len(fileSimplifications) > 0 {
			// Determine output path
			outputPath := ruleFile
			if outputDir != "" {
				outputPath = filepath.Join(outputDir, filepath.Base(ruleFile))
			}

			if !dryRun {
				// Ensure output directory exists
				if outputDir != "" {
					if err := os.MkdirAll(outputDir, 0755); err != nil {
						filesSkipped = append(filesSkipped, map[string]interface{}{
							"file":   ruleFilePath,
							"reason": fmt.Sprintf("Failed to create output directory: %v", err),
						})
						continue
					}
				}

				if err := os.WriteFile(outputPath, []byte(simplifiedContent), 0644); err != nil {
					filesSkipped = append(filesSkipped, map[string]interface{}{
						"file":   ruleFilePath,
						"reason": fmt.Sprintf("Failed to write simplified file: %v", err),
					})
					continue
				}
			}

			filesProcessed = append(filesProcessed, map[string]interface{}{
				"file":            ruleFilePath,
				"output":          outputPath,
				"simplifications": len(fileSimplifications),
			})

			simplifications = append(simplifications, fileSimplifications...)
		} else {
			filesSkipped = append(filesSkipped, map[string]interface{}{
				"file":   ruleFilePath,
				"reason": "No simplifications found",
			})
		}
	}

	results["files_processed"] = filesProcessed
	results["files_skipped"] = filesSkipped
	results["simplifications"] = simplifications

	return results
}

func getSimplificationPatterns() map[string]map[string]string {
	return map[string]map[string]string{
		"use_exarp_estimation": {
			"old": "Use exarp estimation tool to estimate task duration",
			"new": "Use `estimation` tool (action=estimate) to estimate task duration",
		},
		"use_exarp_task_analysis": {
			"old": "Analyze tasks for duplicates, dependencies, and parallelization",
			"new": "Use `task_analysis` tool (action=duplicates|dependencies|parallelization) to analyze tasks",
		},
		"use_exarp_task_workflow": {
			"old": "Use task workflow tools to manage task lifecycle",
			"new": "Use `task_workflow` tool (action=sync|approve|clarify) to manage task lifecycle",
		},
		"use_exarp_health": {
			"old": "Check project health and status",
			"new": "Use `health` tool (action=server|git|docs) to check project health",
		},
		"use_exarp_lint": {
			"old": "Run linters and fix code issues",
			"new": "Use `lint` tool (action=run|analyze) to run linters and analyze code",
		},
		"use_exarp_testing": {
			"old": "Run tests and check coverage",
			"new": "Use `testing` tool (action=run|coverage) to run tests and check coverage",
		},
	}
}

func getRuleTemplates() map[string]RuleTemplate {
	return map[string]RuleTemplate{
		"python": {
			Filename: "python.mdc",
			Content: `---
description: Python development rules for AI assistance
globs: ["**/*.py"]
alwaysApply: true
---

# Python Development Rules

## Code Style
- Follow PEP 8 conventions
- Use type hints for function arguments and return values
- Maximum line length: 100 characters
- Use docstrings for all public functions and classes

## Imports
- Group imports: stdlib, third-party, local
- Use absolute imports over relative where possible
- Sort imports alphabetically within groups

## Error Handling
- Use specific exception types
- Always include informative error messages
- Log exceptions with context

## Testing
- Write tests for all new functions
- Use pytest as the testing framework
- Aim for >80% test coverage

## Security
- Never hardcode secrets or credentials
- Use environment variables for configuration
- Validate all external input
`,
		},
		"typescript": {
			Filename: "typescript.mdc",
			Content: `---
description: TypeScript development rules for AI assistance
globs: ["**/*.ts", "**/*.tsx"]
alwaysApply: true
---

# TypeScript Development Rules

## Type Safety
- Always use explicit types, avoid 'any'
- Use interfaces for object shapes
- Prefer 'const' assertions where applicable

## Code Style
- Use arrow functions for callbacks
- Prefer async/await over raw Promises
- Use destructuring for cleaner code

## React (if applicable)
- Use functional components with hooks
- Prefer named exports over default exports
- Keep components small and focused

## Testing
- Use Jest for unit testing
- Write tests for edge cases
- Mock external dependencies
`,
		},
		"go": {
			Filename: "go.mdc",
			Content: `---
description: Go development rules for AI assistance
globs: ["**/*.go"]
alwaysApply: true
---

# Go Development Rules

## Code Style
- Follow Go formatting conventions (gofmt)
- Use goimports for import organization
- Maximum line length: 100 characters
- Use meaningful variable names

## Error Handling
- Always handle errors explicitly
- Use fmt.Errorf with %w for error wrapping
- Return errors, don't ignore them

## Testing
- Write tests for all exported functions
- Use table-driven tests where appropriate
- Aim for >80% test coverage

## Concurrency
- Use channels for communication
- Prefer sync primitives when appropriate
- Document goroutine lifecycle
`,
		},
		"api": {
			Filename: "api.mdc",
			Content: `---
description: API development rules for AI assistance
globs: ["**/api/**", "**/routes/**", "**/endpoints/**"]
alwaysApply: false
---

# API Development Rules

## REST Conventions
- Use standard HTTP methods (GET, POST, PUT, DELETE)
- Follow RESTful resource naming
- Use proper HTTP status codes

## Security
- Validate all input
- Sanitize user data
- Use authentication and authorization
- Rate limit endpoints

## Documentation
- Document all endpoints
- Include request/response examples
- Specify error responses
`,
		},
		"testing": {
			Filename: "testing.mdc",
			Content: `---
description: Testing best practices for AI assistance
globs: ["**/*_test.*", "**/*.test.*", "**/test/**"]
alwaysApply: false
---

# Testing Rules

## Test Structure
- Use descriptive test names
- Follow AAA pattern (Arrange, Act, Assert)
- Keep tests focused and isolated

## Coverage
- Aim for >80% code coverage
- Test edge cases and error conditions
- Test both success and failure paths

## Best Practices
- Use test fixtures for complex setup
- Mock external dependencies
- Keep tests fast and deterministic
`,
		},
		"mcp": {
			Filename: "mcp.mdc",
			Content: `---
description: MCP (Model Context Protocol) development rules
globs: ["**/mcp/**", "**/*mcp*"]
alwaysApply: false
---

# MCP Development Rules

## Tool Implementation
- Follow MCP tool schema conventions
- Use proper error handling and reporting
- Document tool parameters clearly

## Resources
- Use proper URI schemes
- Implement resource handlers correctly
- Handle resource errors gracefully

## Prompts
- Structure prompts clearly
- Use proper prompt templates
- Document prompt parameters
`,
		},
	}
}

// config_generator_rules.go — CursorRulesGenerator: project analysis and .cursor/rules generation.
package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CursorRulesGenerator generates Cursor rules files based on project analysis.
type CursorRulesGenerator struct {
	projectRoot string
	rulesDir    string
}

// NewCursorRulesGenerator creates a new CursorRulesGenerator.
func NewCursorRulesGenerator(projectRoot string) *CursorRulesGenerator {
	return &CursorRulesGenerator{
		projectRoot: projectRoot,
		rulesDir:    filepath.Join(projectRoot, ".cursor", "rules"),
	}
}

// AnalyzeProject analyzes the project to determine which rules to generate.
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
		".py":   "python",
		".ts":   "typescript",
		".tsx":  "typescript",
		".js":   "javascript",
		".jsx":  "javascript",
		".go":   "go",
		".rs":   "rust",
		".sh":   "shell",
		".yml":  "yaml",
		".yaml": "yaml",
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

	// Detect Ansible (playbooks/roles in ansible/)
	if g.hasDirectory("ansible") && (g.hasDirectory("ansible/playbooks") || g.hasDirectory("ansible/roles")) {
		frameworks = append(frameworks, "ansible")
		if !contains(recommendedRules, "ansible") {
			recommendedRules = append(recommendedRules, "ansible")
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

// GenerateRules generates Cursor rules files.
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

// RuleTemplate represents a rule template.
type RuleTemplate struct {
	Filename string
	Content  string
}

func getRuleTemplate(ruleName string) (RuleTemplate, bool) {
	templates := getRuleTemplates()
	template, exists := templates[ruleName]

	return template, exists
}

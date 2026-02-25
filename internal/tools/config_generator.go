// config_generator.go — MCP generate_config tool: handlers for rules, ignore, simplify actions.
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

// handleGenerateConfigNative handles the generate_config tool with native Go implementation.
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

// handleGenerateRules handles the "rules" action for generate_config.
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

// handleGenerateIgnore handles the "ignore" action for generate_config.
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

// handleSimplifyRules handles the "simplify" action for generate_config.
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


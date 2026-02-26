// config_generator.go — MCP generate_config tool: handlers for rules, ignore, simplify actions.
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/davidl71/exarp-go/internal/framework"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cast"
)

// handleGenerateConfigNative handles the generate_config tool with native Go implementation.
func handleGenerateConfigNative(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	action := strings.TrimSpace(cast.ToString(params["action"]))
	if action == "" {
		action = "rules"
	}

	// Get project root - try multiple methods
	projectRoot, err := GetProjectRootWithFallback()
	if err != nil {
		return nil, fmt.Errorf("project root: %w", err)
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
	rulesStr := cast.ToString(params["rules"])
	overwrite := cast.ToBool(params["overwrite"])
	analyzeOnly := cast.ToBool(params["analyze_only"])

	generator := NewCursorRulesGenerator(projectRoot)

	if analyzeOnly {
		analysis := generator.AnalyzeProject()
		result := map[string]interface{}{
			"success":         true,
			"analysis":        analysis,
			"available_rules": getAvailableRuleNames(),
			"tip":             "Run without analyze_only to generate recommended rules",
		}

		return framework.FormatResult(result, "")
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

	return framework.FormatResult(map[string]interface{}{
		"success": true,
		"data":    results,
	}, "")
}

// handleGenerateIgnore handles the "ignore" action for generate_config.
func handleGenerateIgnore(ctx context.Context, params map[string]interface{}, projectRoot string) ([]framework.TextContent, error) {
	includeIndexing := true
	if _, has := params["include_indexing"]; has {
		includeIndexing = cast.ToBool(params["include_indexing"])
	}

	analyzeProject := true
	if _, has := params["analyze_project"]; has {
		analyzeProject = cast.ToBool(params["analyze_project"])
	}

	dryRun := cast.ToBool(params["dry_run"])

	generator := NewCursorIgnoreGenerator(projectRoot)
	results := generator.GenerateIgnore(includeIndexing, analyzeProject, dryRun)

	return framework.FormatResult(map[string]interface{}{
		"success": true,
		"data":    results,
	}, "")
}

// handleSimplifyRules handles the "simplify" action for generate_config.
func handleSimplifyRules(ctx context.Context, params map[string]interface{}, projectRoot string) ([]framework.TextContent, error) {
	ruleFilesStr := cast.ToString(params["rule_files"])
	dryRun := cast.ToBool(params["dry_run"])
	outputDir := cast.ToString(params["output_dir"])

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

	return framework.FormatResult(map[string]interface{}{
		"status":          "success",
		"files_processed": results["files_processed"],
		"files_skipped":   results["files_skipped"],
		"simplifications": results["simplifications"],
		"dry_run":         dryRun,
	}, "")
}

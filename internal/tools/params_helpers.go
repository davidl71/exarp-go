// params_helpers.go — Shared helpers for MCP tool param parsing and output paths.
package tools

import (
	"path/filepath"
	"strings"

	"github.com/spf13/cast"
)

// ParamString returns params[key] as a trimmed string. Use instead of params["key"].(string) to avoid panics and trim whitespace.
func ParamString(params map[string]interface{}, key string) string {
	return strings.TrimSpace(cast.ToString(params[key]))
}

// DefaultReportOutputPath returns params["output_path"] if non-empty, else a default path under projectRoot.
// Noisy generated analysis artifacts default to out/, while user-facing docs default to docs/.
func DefaultReportOutputPath(projectRoot, defaultFilename string, params map[string]interface{}) string {
	if p := ParamString(params, "output_path"); p != "" {
		return p
	}

	if useOutDirForDefaultReport(defaultFilename) {
		return filepath.Join(projectRoot, "out", defaultFilename)
	}

	return filepath.Join(projectRoot, "docs", defaultFilename)
}

// DefaultPlanOutputPath returns params["output_path"] if non-empty, else projectRoot/.cursor/plans/defaultFilename.
// Use for tools that write plan files with an optional output_path and .cursor/plans default.
func DefaultPlanOutputPath(projectRoot, defaultFilename string, params map[string]interface{}) string {
	if p := ParamString(params, "output_path"); p != "" {
		return p
	}
	return filepath.Join(projectRoot, ".cursor", "plans", defaultFilename)
}

func useOutDirForDefaultReport(defaultFilename string) bool {
	switch defaultFilename {
	case "PROJECT_OVERVIEW.md",
		"TASK_ANALYSIS_DUPLICATES.md",
		"TAG_ANALYSIS_RESULT.json",
		"TASK_ANALYSIS_DEPENDENCIES.md",
		"TASK_ANALYSIS_DEPENDENCIES_SUMMARY.json",
		"TASK_ANALYSIS_COMPLEXITY.json",
		"TASK_ANALYSIS_PARALLELIZATION.md",
		"TASK_ANALYSIS_NOISE.json",
		"SUGGEST_DEPS_REPORT.json",
		"task_discovery_report.json":
		return true
	default:
		return false
	}
}

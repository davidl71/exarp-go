// params_helpers.go — Shared helpers for MCP tool param parsing and output paths.
package tools

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cast"
)

// ParamString returns params[key] as a trimmed string. Use instead of params["key"].(string) to avoid panics and trim whitespace.
func ParamString(params map[string]interface{}, key string) string {
	return strings.TrimSpace(cast.ToString(params[key]))
}

// ParamInt returns params[key] as an int. Returns defaultVal if key is missing or wrong type.
// Uses defaultVal for missing keys; to distinguish missing from explicit 0, use HasKey first.
func ParamInt(params map[string]interface{}, key string, defaultVal int) int {
	if val, err := cast.ToIntE(params[key]); err == nil {
		return val
	}
	return defaultVal
}

// ParamBool returns params[key] as a bool. Returns defaultVal if key is missing or wrong type.
// To distinguish between missing key and explicit false, use HasKey() first.
func ParamBool(params map[string]interface{}, key string, defaultVal bool) bool {
	if val, err := cast.ToBoolE(params[key]); err == nil {
		return val
	}
	return defaultVal
}

// HasKey returns true if the key exists in params (even if value is nil/empty).
func HasKey(params map[string]interface{}, key string) bool {
	_, exists := params[key]
	return exists
}

// RequireParam returns params[key] as a trimmed string. Returns an error if the
// key is missing or the value is empty after trimming.
func RequireParam(params map[string]interface{}, key string) (string, error) {
	v := strings.TrimSpace(cast.ToString(params[key]))
	if v == "" {
		return "", fmt.Errorf("missing required parameter %q", key)
	}
	return v, nil
}

// ParamEnum extracts params[key] as a trimmed string and validates it against
// the provided set of valid values. Returns an error listing valid options when
// the value is non-empty but not recognised. Returns defaultVal when the key is
// absent or empty.
func ParamEnum(params map[string]interface{}, key string, valid []string, defaultVal string) (string, error) {
	v := strings.TrimSpace(cast.ToString(params[key]))
	if v == "" {
		return defaultVal, nil
	}
	for _, ok := range valid {
		if strings.EqualFold(v, ok) {
			return ok, nil
		}
	}
	return "", fmt.Errorf("invalid value %q for %q: must be one of %s", v, key, strings.Join(valid, ", "))
}

// RequireEnum combines RequireParam and ParamEnum: the key must be present and
// its value must be in the valid set.
func RequireEnum(params map[string]interface{}, key string, valid []string) (string, error) {
	v, err := RequireParam(params, key)
	if err != nil {
		return "", err
	}
	for _, ok := range valid {
		if strings.EqualFold(v, ok) {
			return ok, nil
		}
	}
	return "", fmt.Errorf("invalid value %q for %q: must be one of %s", v, key, strings.Join(valid, ", "))
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

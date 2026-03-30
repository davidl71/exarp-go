// params_helpers.go — Shared helpers for MCP tool param parsing and output paths.
// Prefer ParamInt / ParamFloat64 / ParamFloat64OK / ParamStringSlice over raw type assertions on JSON-backed maps.
package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cast"
)

// ParamString returns params[key] as a trimmed string. Use instead of params["key"].(string) to avoid panics and trim whitespace.
func ParamString(params map[string]interface{}, key string) string {
	return strings.TrimSpace(cast.ToString(params[key]))
}

// ParamBool returns params[key] as a bool. Returns defaultVal if key is missing or wrong type.
func ParamBool(params map[string]interface{}, key string, defaultVal bool) bool {
	v, ok := params[key]
	if !ok || v == nil {
		return defaultVal
	}
	if val, err := cast.ToBoolE(v); err == nil {
		return val
	}
	return defaultVal
}

// ParamInt returns params[key] as int. Accepts JSON numbers (float64), integers, and numeric strings.
// Returns defaultVal if the key is missing, nil, or not convertible.
func ParamInt(params map[string]interface{}, key string, defaultVal int) int {
	if params == nil {
		return defaultVal
	}
	v, ok := params[key]
	if !ok || v == nil {
		return defaultVal
	}
	i, err := cast.ToInt64E(v)
	if err != nil {
		return defaultVal
	}
	return int(i)
}

// ParamIntOK returns (value, true) if key is present, non-nil, and converts to int.
func ParamIntOK(params map[string]interface{}, key string) (int, bool) {
	if params == nil {
		return 0, false
	}
	v, ok := params[key]
	if !ok || v == nil {
		return 0, false
	}
	i, err := cast.ToInt64E(v)
	if err != nil {
		return 0, false
	}
	return int(i), true
}

// ParamFloat64 returns params[key] as float64. Returns defaultVal if missing, nil, or not convertible.
func ParamFloat64(params map[string]interface{}, key string, defaultVal float64) float64 {
	if params == nil {
		return defaultVal
	}
	v, ok := params[key]
	if !ok || v == nil {
		return defaultVal
	}
	f, err := cast.ToFloat64E(v)
	if err != nil {
		return defaultVal
	}
	return f
}

// ParamFloat64OK returns (value, true) if key is present, non-nil, and converts to float64.
func ParamFloat64OK(params map[string]interface{}, key string) (float64, bool) {
	if params == nil {
		return 0, false
	}
	v, ok := params[key]
	if !ok || v == nil {
		return 0, false
	}
	f, err := cast.ToFloat64E(v)
	if err != nil {
		return 0, false
	}
	return f, true
}

// ParamStringSlice coerces params[key] to []string (JSON arrays and []interface{}; a plain string becomes one element — use ParamStringSliceTrimmedCommaSeparated when the string is comma-separated tokens).
// Returns nil when the key is missing or produces no elements.
func ParamStringSlice(params map[string]interface{}, key string) []string {
	if params == nil {
		return nil
	}
	out := cast.ToStringSlice(params[key])
	if len(out) == 0 {
		return nil
	}
	return out
}

// ParamStringSliceTrimmed is like ParamStringSlice but trims each element and drops empty strings.
// Returns nil when there are no non-empty elements after trimming.
func ParamStringSliceTrimmed(params map[string]interface{}, key string) []string {
	raw := ParamStringSlice(params, key)
	if len(raw) == 0 {
		return nil
	}
	out := make([]string, 0, len(raw))
	for _, s := range raw {
		if t := strings.TrimSpace(s); t != "" {
			out = append(out, t)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// ParamStringSliceTrimmedCommaSeparated trims and splits a string value on commas into separate elements.
// JSON arrays (and other non-string values) delegate to ParamStringSliceTrimmed so each array element stays one token.
// Use for tags, dependencies, recommended_tools, task_ids lists where clients pass "a,b" or ["a","b"].
func ParamStringSliceTrimmedCommaSeparated(params map[string]interface{}, key string) []string {
	if params == nil {
		return nil
	}
	v, ok := params[key]
	if !ok || v == nil {
		return nil
	}
	if s, ok := v.(string); ok {
		parts := strings.Split(s, ",")
		out := make([]string, 0, len(parts))
		for _, p := range parts {
			if t := strings.TrimSpace(p); t != "" {
				out = append(out, t)
			}
		}
		if len(out) == 0 {
			return nil
		}
		return out
	}
	return ParamStringSliceTrimmed(params, key)
}

// ParamOutputPath returns params["output_path"] as a trimmed string.
func ParamOutputPath(params map[string]interface{}) string {
	return ParamString(params, "output_path")
}

// ParamOutputFormat returns params["output_format"] as a trimmed string, defaulting to defaultVal.
func ParamOutputFormat(params map[string]interface{}, defaultVal string) string {
	if v := ParamString(params, "output_format"); v != "" {
		return v
	}
	return defaultVal
}

// EnsureParentDir creates the parent directory for a path when needed.
// If path is empty, "."-rooted, or has no parent directory, it does nothing.
func EnsureParentDir(path string) error {
	if strings.TrimSpace(path) == "" {
		return nil
	}
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}
	return os.MkdirAll(dir, 0o755)
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

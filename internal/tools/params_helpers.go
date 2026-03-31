// params_helpers.go — Shared helpers for MCP tool param parsing and output paths.
// Prefer ParamInt / ParamFloat64 / ParamFloat64OK / ParamStringSlice over raw type assertions on JSON-backed maps.
package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/davidl71/exarp-go/internal/models"
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

// parseStringAsJSONStringSlice reports ok when s unmarshals as a JSON array of strings (possibly empty after trimming elements).
// When ok is true, callers must not fall back to comma-splitting (e.g. "[]" is an empty tag list, not a literal tag).
func parseStringAsJSONStringSlice(s string) (out []string, ok bool) {
	s = strings.TrimSpace(s)
	if len(s) < 2 || s[0] != '[' {
		return nil, false
	}
	var arr []string
	if err := json.Unmarshal([]byte(s), &arr); err != nil {
		return nil, false
	}
	out = make([]string, 0, len(arr))
	for _, x := range arr {
		if t := strings.TrimSpace(x); t != "" {
			out = append(out, t)
		}
	}
	return out, true
}

var taskDependencyIDRe = regexp.MustCompile(`T-\d+`)

// dependencyTokensFromValue flattens dependency parameters from MCP/JSON (nested arrays, JSON-text
// blobs, comma-separated strings, quoted IDs) into raw string tokens before ID validation.
func dependencyTokensFromValue(v interface{}) []string {
	if v == nil {
		return nil
	}
	switch t := v.(type) {
	case string:
		return dependencyStringTokens(t)
	case []byte:
		return dependencyStringTokens(string(t))
	case []interface{}:
		var out []string
		for _, item := range t {
			out = append(out, dependencyTokensFromValue(item)...)
		}
		return out
	case []string:
		var out []string
		for _, item := range t {
			out = append(out, dependencyTokensFromValue(item)...)
		}
		return out
	default:
		return dependencyStringTokens(strings.TrimSpace(cast.ToString(t)))
	}
}

func dependencyStringTokens(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	if len(s) >= 2 && s[0] == '[' {
		var arr []json.RawMessage
		if err := json.Unmarshal([]byte(s), &arr); err == nil {
			if len(arr) == 0 {
				return nil
			}
			var out []string
			for _, raw := range arr {
				out = append(out, dependencyStringTokens(strings.TrimSpace(string(raw)))...)
			}
			return out
		}
	}
	if sl, ok := parseStringAsJSONStringSlice(s); ok {
		var out []string
		for _, x := range sl {
			out = append(out, dependencyTokensFromValue(x)...)
		}
		return out
	}
	if strings.Contains(s, ",") {
		parts := strings.Split(s, ",")
		var out []string
		for _, p := range parts {
			if p = strings.TrimSpace(p); p != "" {
				out = append(out, dependencyStringTokens(p)...)
			}
		}
		return out
	}
	if u, err := strconv.Unquote(s); err == nil {
		return dependencyStringTokens(u)
	}
	return []string{s}
}

// normalizeValidTaskDependencyIDs trims tokens, keeps unique Todo2-style IDs (T-<digits>), and
// recovers IDs embedded in malformed wrappers (e.g. a single token "['T-1']").
func normalizeValidTaskDependencyIDs(tokens []string) []string {
	seen := make(map[string]struct{}, len(tokens))
	out := make([]string, 0, len(tokens))
	for _, tok := range tokens {
		tok = strings.TrimSpace(tok)
		if tok == "" {
			continue
		}
		if models.IsValidTaskID(tok) {
			if _, ok := seen[tok]; ok {
				continue
			}
			seen[tok] = struct{}{}
			out = append(out, tok)
			continue
		}
		for _, m := range taskDependencyIDRe.FindAllString(tok, -1) {
			if !models.IsValidTaskID(m) {
				continue
			}
			if _, ok := seen[m]; ok {
				continue
			}
			seen[m] = struct{}{}
			out = append(out, m)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// ParamTaskDependencyIDs extracts task dependency IDs from params[key].
// It accepts comma-separated strings, JSON arrays (including nested), quoted JSON text, and
// native []interface{} slices from JSON-RPC — cases that plain ParamStringSliceTrimmedCommaSeparated
// can mis-parse as one invalid token (e.g. "[\"T-1\"]").
func ParamTaskDependencyIDs(params map[string]interface{}, key string) []string {
	if params == nil {
		return nil
	}
	v, ok := params[key]
	if !ok || v == nil {
		return nil
	}
	return normalizeValidTaskDependencyIDs(dependencyTokensFromValue(v))
}

// ParamStringSliceTrimmedCommaSeparated trims and splits a string value on commas into separate elements.
// String values that look like JSON arrays of strings are parsed as such (one element per array item).
// Native JSON arrays in params (and other non-string values) delegate to ParamStringSliceTrimmed.
// Use for tags, dependencies, recommended_tools, task_ids lists where clients pass "a,b", `["a","b"]`, or ["a","b"].
func ParamStringSliceTrimmedCommaSeparated(params map[string]interface{}, key string) []string {
	if params == nil {
		return nil
	}
	v, ok := params[key]
	if !ok || v == nil {
		return nil
	}
	if s, ok := v.(string); ok {
		s = strings.TrimSpace(s)
		if s == "" {
			return nil
		}
		if jsonSlice, ok := parseStringAsJSONStringSlice(s); ok {
			if len(jsonSlice) == 0 {
				return nil
			}
			return jsonSlice
		}
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

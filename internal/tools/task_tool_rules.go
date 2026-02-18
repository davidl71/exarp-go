package tools

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// TaskToolRulesFile is the optional config file for tag → recommended_tools mapping (under project root).
const TaskToolRulesFile = ".cursor/task_tool_rules.yaml"

// DefaultTagToolMap returns the built-in tag → tool IDs mapping per TASK_TOOL_ENRICHMENT_DESIGN.md.
// Tag keys are lowercase (e.g. "research"); tags may include or omit "#" when looking up.
func DefaultTagToolMap() map[string][]string {
	return map[string][]string{
		"research":        {"tractatus_thinking"},
		"design":          {"tractatus_thinking"},
		"planning":        {"tractatus_thinking"},
		"docs":            {"context7"},
		"task_workflow":   {"task_workflow", "task_analysis"},
		"task_management": {"task_workflow", "task_analysis"},
	}
}

// taskToolRulesYAML is the shape of .cursor/task_tool_rules.yaml.
type taskToolRulesYAML struct {
	TagTools map[string][]string `yaml:"tag_tools"`
}

// LoadTaskToolRules returns tag → tool IDs. It uses DefaultTagToolMap and overlays
// .cursor/task_tool_rules.yaml if present (keys from file override defaults for that tag).
func LoadTaskToolRules(projectRoot string) map[string][]string {
	out := make(map[string][]string)
	for k, v := range DefaultTagToolMap() {
		out[k] = append([]string(nil), v...)
	}

	path := filepath.Join(projectRoot, TaskToolRulesFile)
	data, err := os.ReadFile(path)
	if err != nil {
		return out
	}

	var f taskToolRulesYAML
	if err := yaml.Unmarshal(data, &f); err != nil || f.TagTools == nil {
		return out
	}

	for tag, tools := range f.TagTools {
		tag = strings.TrimSpace(strings.TrimPrefix(strings.ToLower(tag), "#"))
		if tag == "" {
			continue
		}
		clean := make([]string, 0, len(tools))
		for _, t := range tools {
			if s := strings.TrimSpace(t); s != "" {
				clean = append(clean, s)
			}
		}
		if len(clean) > 0 {
			out[tag] = clean
		}
	}

	return out
}

// ToolsForTags returns recommended tool IDs for the given tags using the provided tag→tools map.
// Tags may be with or without "#"; lookup is case-insensitive. Deduplicated.
func ToolsForTags(tagToTools map[string][]string, tags []string) []string {
	seen := make(map[string]bool)
	var out []string
	for _, t := range tags {
		tag := strings.TrimSpace(strings.TrimPrefix(strings.ToLower(t), "#"))
		if tag == "" {
			continue
		}
		tools := tagToTools[tag]
		for _, id := range tools {
			if !seen[id] {
				seen[id] = true
				out = append(out, id)
			}
		}
	}
	return out
}

// MergeRecommendedTools merges existing recommended_tools with new tools (no duplicates).
// Order: existing first, then new tools not already present.
func MergeRecommendedTools(existing []string, add []string) []string {
	seen := make(map[string]bool)
	var out []string
	for _, s := range existing {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	for _, s := range add {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}

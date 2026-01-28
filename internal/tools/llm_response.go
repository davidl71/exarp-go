package tools

import "strings"

// MaxLLMResponseSnippetLen is the max chars of LLM response to include in parse errors for debugging.
const MaxLLMResponseSnippetLen = 200

// stripMarkdownCodeBlocks removes ```json ... ``` or ``` ... ``` from s and returns the inner content trimmed.
func stripMarkdownCodeBlocks(s string) string {
	s = strings.TrimSpace(s)
	if idx := strings.Index(s, "```json"); idx >= 0 {
		start := idx + 7
		if end := strings.Index(s[start:], "```"); end >= 0 {
			return strings.TrimSpace(s[start : start+end])
		}
	}
	if idx := strings.Index(s, "```"); idx >= 0 {
		start := idx + 3
		if end := strings.Index(s[start:], "```"); end >= 0 {
			return strings.TrimSpace(s[start : start+end])
		}
	}
	return s
}

// ExtractJSONArrayFromLLMResponse extracts a JSON array from LLM output that may be
// wrapped in markdown code blocks or preceded by plain text (e.g. "Priority: high" or "Phase 1").
// Returns a string suitable for json.Unmarshal into []map[string]interface{} or similar; callers should handle parse errors.
func ExtractJSONArrayFromLLMResponse(result string) string {
	s := stripMarkdownCodeBlocks(result)
	if jsonStart := strings.Index(s, "["); jsonStart >= 0 {
		if jsonEnd := strings.LastIndex(s, "]"); jsonEnd > jsonStart {
			return s[jsonStart : jsonEnd+1]
		}
	}
	return s
}

// ExtractJSONObjectFromLLMResponse extracts a JSON object from LLM output that may be
// wrapped in markdown code blocks or preceded by plain text.
// Returns a string suitable for json.Unmarshal into map[string]interface{} or similar; callers should handle parse errors.
func ExtractJSONObjectFromLLMResponse(result string) string {
	s := stripMarkdownCodeBlocks(result)
	if jsonStart := strings.Index(s, "{"); jsonStart >= 0 {
		if jsonEnd := strings.LastIndex(s, "}"); jsonEnd > jsonStart {
			return s[jsonStart : jsonEnd+1]
		}
	}
	return s
}

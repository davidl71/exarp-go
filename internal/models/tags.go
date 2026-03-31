// tags.go — Tag parsing and normalization helpers for Todo2 tasks.
package models

import "strings"

// NormalizeTag converts a user-supplied tag token (with or without leading '#')
// into the canonical representation used across Todo2.
//
// Canonical form:
// - lower-case
// - leading '#'
// - only [a-z0-9_-] after the '#'
// - consecutive separators collapsed to single '-'
//
// Returns ("", false) when the token is empty or normalizes to nothing.
func NormalizeTag(token string) (string, bool) {
	t := strings.TrimSpace(token)
	if t == "" {
		return "", false
	}

	// Strip any number of leading '#'.
	t = strings.TrimLeft(t, "#")
	t = strings.TrimSpace(t)
	if t == "" {
		return "", false
	}

	t = strings.ToLower(t)

	var b strings.Builder
	b.Grow(len(t) + 1)

	// We keep only [a-z0-9_-]. Everything else becomes a separator.
	lastWasSep := true // avoid leading separators
	for i := 0; i < len(t); i++ {
		c := t[i]
		isAlphaNum := (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9')
		isKeep := isAlphaNum || c == '_' || c == '-'
		if isKeep {
			b.WriteByte(c)
			lastWasSep = false
			continue
		}
		// Treat all other characters as separators.
		if !lastWasSep {
			b.WriteByte('-')
			lastWasSep = true
		}
	}

	out := strings.Trim(b.String(), "-_")
	if out == "" {
		return "", false
	}
	return "#" + out, true
}

// NormalizeTags normalizes and de-dupes tags, preserving first occurrence order.
// It also splits any element that contains whitespace into multiple tokens so
// malformed inputs like "a b" don't silently become a single tag.
func NormalizeTags(raw []string) []string {
	if len(raw) == 0 {
		return nil
	}

	seen := make(map[string]bool, len(raw))
	out := make([]string, 0, len(raw))

	for _, item := range raw {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		parts := strings.Fields(item)
		if len(parts) == 0 {
			continue
		}
		for _, p := range parts {
			if norm, ok := NormalizeTag(p); ok {
				if seen[norm] {
					continue
				}
				seen[norm] = true
				out = append(out, norm)
			}
		}
	}

	if len(out) == 0 {
		return nil
	}
	return out
}

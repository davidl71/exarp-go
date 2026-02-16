// Package tools - execution_apply implements change application for MODEL_ASSISTED_WORKFLOW Phase 4 (T-214).
// Applies structured file changes (from task_execution template response) within project root with path validation.
package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/davidl71/exarp-go/internal/security"
)

// FileChange represents a single file edit from an execution response (task_execution template).
type FileChange struct {
	File       string `json:"file"`
	Location   string `json:"location"`
	OldContent string `json:"old_content"`
	NewContent string `json:"new_content"`
}

// ExecutionResult is the parsed response from a model using the task_execution template.
type ExecutionResult struct {
	Changes     []FileChange `json:"changes"`
	Explanation string       `json:"explanation"`
	Confidence  float64      `json:"confidence"`
}

// ParseExecutionResponse parses JSON from the task_execution template response.
// Handles optional whitespace and common LLM wrapping (e.g. markdown code blocks).
func ParseExecutionResponse(body []byte) (*ExecutionResult, error) {
	text := strings.TrimSpace(string(body))
	// Strip markdown code block if present
	if strings.HasPrefix(text, "```json") {
		text = strings.TrimPrefix(text, "```json")
	} else if strings.HasPrefix(text, "```") {
		text = strings.TrimPrefix(text, "```")
	}

	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)

	var result ExecutionResult
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		return nil, fmt.Errorf("parse execution response: %w", err)
	}

	return &result, nil
}

// ApplyChanges applies a list of file changes under projectRoot. All paths are validated
// to stay within project root. Returns the list of applied file paths and the first error if any.
func ApplyChanges(projectRoot string, changes []FileChange) (applied []string, err error) {
	if projectRoot == "" {
		return nil, fmt.Errorf("project root is required")
	}

	applied = make([]string, 0, len(changes))

	for i, c := range changes {
		if c.File == "" {
			continue
		}

		absPath, err := security.ValidatePath(c.File, projectRoot)
		if err != nil {
			return applied, fmt.Errorf("change %d: invalid path %q: %w", i+1, c.File, err)
		}

		content, readErr := os.ReadFile(absPath)
		if readErr != nil && !os.IsNotExist(readErr) {
			return applied, fmt.Errorf("change %d: read %q: %w", i+1, c.File, readErr)
		}

		var newContent string
		if os.IsNotExist(readErr) {
			// New file
			newContent = c.NewContent
		} else {
			current := string(content)
			if strings.TrimSpace(c.OldContent) != "" {
				// Replace first occurrence of old_content with new_content
				if !strings.Contains(current, c.OldContent) {
					return applied, fmt.Errorf("change %d: file %q does not contain the specified old_content", i+1, c.File)
				}

				newContent = strings.Replace(current, c.OldContent, c.NewContent, 1)
			} else {
				newContent = c.NewContent
			}
		}

		dir := filepath.Dir(absPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return applied, fmt.Errorf("change %d: mkdir %q: %w", i+1, dir, err)
		}

		if err := os.WriteFile(absPath, []byte(newContent), 0644); err != nil {
			return applied, fmt.Errorf("change %d: write %q: %w", i+1, c.File, err)
		}

		rel, _ := filepath.Rel(projectRoot, absPath)
		if rel == "" || rel == "." {
			rel = c.File
		}

		applied = append(applied, rel)
	}

	return applied, nil
}

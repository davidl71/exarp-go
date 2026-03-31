//go:build darwin && arm64 && cgo
// +build darwin,arm64,cgo

// task_discovery_native_scanners.go — Task discovery: Apple FM enhancement and planning doc scanner (git JSON: task_discovery_common.go).
// See also: task_discovery_native.go
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ─── Contents ───────────────────────────────────────────────────────────────
//   enhanceTaskWithAppleFM
//   enhancePlanningDocWithAppleFM — enhancePlanningDocWithAppleFM uses the default FM to extract task/epic references and structure from planning documents.
//   scanPlanningDocs — scanPlanningDocs scans markdown files for planning document structure and task/epic links.
// ────────────────────────────────────────────────────────────────────────────

// ─── enhanceTaskWithAppleFM ─────────────────────────────────────────────────
func enhanceTaskWithAppleFM(ctx context.Context, taskText string) map[string]interface{} {
	if !FMAvailable() {
		return nil
	}

	fmCtx, cancel := context.WithTimeout(ctx, fmDiscoveryTimeout)
	defer cancel()

	prompt := fmt.Sprintf(`Extract structured information from this task comment:

"%s"

Return JSON with: {"description": "cleaned task description", "priority": "low|medium|high", "category": "bug|feature|refactor|docs"}`,
		taskText)

	result, err := DefaultFMProvider().Generate(fmCtx, prompt, 200, 0.2)
	if err != nil {
		return nil
	}

	// Try to parse JSON from result
	jsonStart := strings.Index(result, "{")
	jsonEnd := strings.LastIndex(result, "}")

	if jsonStart >= 0 && jsonEnd > jsonStart {
		var enhanced map[string]interface{}
		if err := json.Unmarshal([]byte(result[jsonStart:jsonEnd+1]), &enhanced); err == nil {
			return enhanced
		}
	}

	return nil
}

// ─── enhancePlanningDocWithAppleFM ──────────────────────────────────────────
// enhancePlanningDocWithAppleFM uses the default FM to extract task/epic references and structure from planning documents.
// A per-call timeout prevents a hung FM from blocking the entire discovery scan.
func enhancePlanningDocWithAppleFM(ctx context.Context, content string, filePath string) map[string]interface{} {
	if !FMAvailable() {
		return nil
	}

	fmCtx, cancel := context.WithTimeout(ctx, fmDiscoveryTimeout)
	defer cancel()

	contentLimit := len(content)
	if contentLimit > 5000 {
		contentLimit = 5000
	}

	contentPreview := content[:contentLimit]

	prompt := fmt.Sprintf(`Analyze this planning document and extract structured information:

File: %s

Content:
%s

Extract:
1. Task IDs referenced (format: T-123 or T-1234567890)
2. Epic IDs referenced (tasks tagged with #epic)
3. Planning document type (epic_planning, feature_planning, migration_planning, architecture_planning, etc.)
4. Related planning documents mentioned (file paths)
5. Task/epic relationships described

Return JSON only (no other text):
{
  "task_refs": ["T-123", "T-456"],
  "epic_refs": ["T-789"],
  "doc_type": "epic_planning",
  "related_docs": ["docs/planning/related-plan.md"],
  "relationships": [{"from": "T-123", "to": "T-456", "type": "depends_on"}]
}`,
		filePath, contentPreview)

	result, err := DefaultFMProvider().Generate(fmCtx, prompt, 1000, 0.2)
	if err != nil {
		return nil
	}

	// Parse JSON from result
	jsonStart := strings.Index(result, "{")
	jsonEnd := strings.LastIndex(result, "}")

	if jsonStart >= 0 && jsonEnd > jsonStart {
		var enhanced map[string]interface{}
		if err := json.Unmarshal([]byte(result[jsonStart:jsonEnd+1]), &enhanced); err == nil {
			return enhanced
		}
	}

	return nil
}

// ─── scanPlanningDocs ───────────────────────────────────────────────────────
// scanPlanningDocs scans markdown files for planning document structure and task/epic links.
func scanPlanningDocs(ctx context.Context, projectRoot string, docPath string, ignorePaths []string, useAppleFM bool) []map[string]interface{} {
	discoveries := []map[string]interface{}{}

	searchPath := projectRoot
	if docPath != "" {
		searchPath = filepath.Join(projectRoot, docPath)
	}

	// Basic regex patterns (fallback if Apple FM unavailable or for validation)
	taskRefPattern := regexp.MustCompile(`(?:Epic|Task)\s+ID[:\s]+` + "`?T-(\\d+)`?")

	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			if shouldSkipDiscoveryDir(projectRoot, path, ignorePaths) {
				return filepath.SkipDir
			}

			return nil
		}

		if filepath.Ext(path) != ".md" && filepath.Ext(path) != ".markdown" {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		relativePath := strings.TrimPrefix(path, projectRoot+"/")
		contentStr := string(content)

		// Use default FM for semantic extraction if available
		if useAppleFM {
			enhanced := enhancePlanningDocWithAppleFM(ctx, contentStr, relativePath)
			if enhanced != nil {
				discoveries = append(discoveries, map[string]interface{}{
					"type":          "PLANNING_DOC",
					"file":          relativePath,
					"task_refs":     enhanced["task_refs"],
					"epic_refs":     enhanced["epic_refs"],
					"doc_type":      enhanced["doc_type"],
					"related_docs":  enhanced["related_docs"],
					"relationships": enhanced["relationships"],
					"source":        "planning_doc",
					"ai_enhanced":   true,
				})

				return nil
			}
		}

		// Fallback to regex-based extraction
		taskRefs := taskRefPattern.FindAllStringSubmatch(contentStr, -1)
		extractedRefs := []string{}

		for _, match := range taskRefs {
			if len(match) > 1 {
				extractedRefs = append(extractedRefs, "T-"+match[1])
			}
		}

		if len(extractedRefs) > 0 {
			discoveries = append(discoveries, map[string]interface{}{
				"type":      "PLANNING_DOC",
				"file":      relativePath,
				"task_refs": extractedRefs,
				"source":    "planning_doc",
			})
		}

		return nil
	})
	if err != nil {
		// Log error but continue
	}

	return discoveries
}

// task_analysis_suggest_deps.go — Dependency suggestion heuristics for task analysis.
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/proto"
	"github.com/davidl71/mcp-go-core/pkg/mcp/response"
)

// SuggestedDependency is a single suggestion: task_id should depend on suggested_dep_id (reason).
type SuggestedDependency struct {
	TaskID         string `json:"task_id"`
	SuggestedDepID string `json:"suggested_dep_id"`
	Reason         string `json:"reason"`
	Source         string `json:"source"` // "content" or "planning_doc"
}

// SuggestDependenciesFromContent infers dependency suggestions from task titles/content.
// Patterns: Phase N.M → Phase N.(M-1); T1.5.M → T1.5.(M-1); "Verify Wave N" / "Wave N verification" → Fix/Verify chain; "Fix tools/CLI" → Fix database.
// Only suggests if suggested_dep_id exists in tasks and is not already a dependency of task_id.
func SuggestDependenciesFromContent(tasks []Todo2Task) []SuggestedDependency {
	idToTask := make(map[string]*Todo2Task)
	phaseRe := regexp.MustCompile(`(?i)phase\s+(\d+)\.(\d+)`)
	t15Re := regexp.MustCompile(`(?i)T1\.5\.(\d+)`)
	// phaseOrT15ToID: e.g. "phase 1.1" -> first task ID containing that phrase (for Phase N.M / T1.5.M lookup)
	phaseOrT15ToID := make(map[string]string)

	for i := range tasks {
		t := &tasks[i]
		idToTask[t.ID] = t

		content := strings.ToLower(t.Content + " " + t.LongDescription)
		if ms := phaseRe.FindStringSubmatch(content); len(ms) >= 3 {
			key := "phase " + ms[1] + "." + ms[2]
			if phaseOrT15ToID[key] == "" {
				phaseOrT15ToID[key] = t.ID
			}
		}

		if ms := t15Re.FindStringSubmatch(content); len(ms) >= 2 {
			key := "t1.5." + ms[1]
			if phaseOrT15ToID[key] == "" {
				phaseOrT15ToID[key] = t.ID
			}
		}
	}

	hasDep := func(taskID, depID string) bool {
		t := idToTask[taskID]
		if t == nil {
			return false
		}

		for _, d := range t.Dependencies {
			if d == depID {
				return true
			}
		}

		return false
	}

	var out []SuggestedDependency

	// Phase N.M → Phase N.(M-1)
	for _, t := range tasks {
		content := strings.ToLower(t.Content + " " + t.LongDescription)
		if hasPhase := phaseRe.FindStringSubmatch(content); len(hasPhase) >= 3 {
			n, m := hasPhase[1], hasPhase[2]
			if m == "1" {
				continue
			}

			prevKey := "phase " + n + "." + prevNum(m)
			candID := phaseOrT15ToID[prevKey]

			if candID != "" && candID != t.ID && !hasDep(t.ID, candID) {
				out = append(out, SuggestedDependency{
					TaskID:         t.ID,
					SuggestedDepID: candID,
					Reason:         "Phase " + n + "." + m + " typically follows " + prevKey,
					Source:         "content",
				})
			}
		}
	}

	// T1.5.M → T1.5.(M-1)
	for _, t := range tasks {
		content := strings.ToLower(t.Content + " " + t.LongDescription)
		if ms := t15Re.FindStringSubmatch(content); len(ms) >= 2 {
			m := ms[1]
			if m == "1" {
				continue
			}

			prevKey := "t1.5." + prevNum(m)
			candID := phaseOrT15ToID[prevKey]

			if candID != "" && candID != t.ID && !hasDep(t.ID, candID) {
				out = append(out, SuggestedDependency{
					TaskID:         t.ID,
					SuggestedDepID: candID,
					Reason:         "T1.5." + m + " typically follows " + prevKey,
					Source:         "content",
				})
			}
		}
	}

	// "Verify Wave N gosdk" / "Wave N verification" → find "Fix gosdk" / "Verify Wave N gosdk"
	verifyWaveRe := regexp.MustCompile(`(?i)(?:verify\s+)?wave\s+(\d+)\s+(?:verification|gosdk)`)

	for _, t := range tasks {
		content := strings.ToLower(t.Content + " " + t.LongDescription)
		if verifyWaveRe.MatchString(content) {
			// "Wave 2 verification" → depend on "Verify Wave 2 gosdk"
			if strings.Contains(content, "wave 2 verification") && !strings.Contains(content, "verify wave 2 gosdk") {
				for _, cand := range tasks {
					cc := strings.ToLower(cand.Content)
					if strings.Contains(cc, "verify wave 2 gosdk") && cand.ID != t.ID && !hasDep(t.ID, cand.ID) {
						out = append(out, SuggestedDependency{
							TaskID:         t.ID,
							SuggestedDepID: cand.ID,
							Reason:         "Wave 2 verification typically follows Verify Wave 2 gosdk",
							Source:         "content",
						})

						break
					}
				}
			}
			// "Verify Wave 2 gosdk" → depend on task with "Fix gosdk" or "mcp-go-core" Fix
			if strings.Contains(content, "verify wave 2 gosdk") {
				for _, cand := range tasks {
					cc := strings.ToLower(cand.Content)
					if (strings.Contains(cc, "gosdk") && strings.Contains(cc, "fix")) && cand.ID != t.ID && !hasDep(t.ID, cand.ID) {
						out = append(out, SuggestedDependency{
							TaskID:         t.ID,
							SuggestedDepID: cand.ID,
							Reason:         "Verify Wave 2 gosdk typically follows mcp-go-core gosdk fix",
							Source:         "content",
						})

						break
					}
				}
			}
		}
	}

	// "Fix tools tests" / "Fix internal/tools" / "Fix CLI integration" → "Fix database package tests"
	fixDatabaseID := ""

	for _, t := range tasks {
		if strings.Contains(strings.ToLower(t.Content), "fix database package tests") {
			fixDatabaseID = t.ID
			break
		}
	}

	if fixDatabaseID != "" {
		for _, t := range tasks {
			if t.ID == fixDatabaseID {
				continue
			}

			cc := strings.ToLower(t.Content)
			if (strings.Contains(cc, "fix tools tests") || strings.Contains(cc, "fix internal/tools tests") || strings.Contains(cc, "fix cli integration")) && !hasDep(t.ID, fixDatabaseID) {
				out = append(out, SuggestedDependency{
					TaskID:         t.ID,
					SuggestedDepID: fixDatabaseID,
					Reason:         "Fix test tasks often depend on Fix database package tests first",
					Source:         "content",
				})
			}
		}
	}

	return out
}

func prevNum(s string) string {
	// simple decimal decrement for "1".."9" and "10"+
	switch s {
	case "1":
		return "0" // no Phase N.0
	case "2":
		return "1"
	case "3":
		return "2"
	case "4":
		return "3"
	case "5":
		return "4"
	case "6":
		return "5"
	case "7":
		return "6"
	case "8":
		return "7"
	case "9":
		return "8"
	case "10":
		return "9"
	default:
		return "1"
	}
}

// handleTaskAnalysisSuggestDependencies returns suggested dependencies from content and (optionally) planning docs.
// Does not apply changes; use task_workflow update with dependencies to apply.
func handleTaskAnalysisSuggestDependencies(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	store, err := getTaskStore(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get task store: %w", err)
	}

	list, err := store.ListTasks(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	tasks := tasksFromPtrs(list)
	suggestions := SuggestDependenciesFromContent(tasks)

	// Optionally include hints from planning docs
	includePlan, _ := params["include_planning_docs"].(bool)
	if includePlan {
		projectRoot, _ := FindProjectRoot()
		if projectRoot != "" {
			planHints := ExtractDependencyHintsFromPlanDir(projectRoot)
			for _, h := range planHints {
				// Only add if both IDs exist and not already a dependency
				existsTask := false
				existsDep := false
				alreadyDep := false

				for _, t := range tasks {
					if t.ID == h.TaskID {
						existsTask = true

						for _, d := range t.Dependencies {
							if d == h.DependsOnID {
								alreadyDep = true
								break
							}
						}
					}

					if t.ID == h.DependsOnID {
						existsDep = true
					}
				}

				if existsTask && existsDep && !alreadyDep {
					suggestions = append(suggestions, SuggestedDependency{
						TaskID:         h.TaskID,
						SuggestedDepID: h.DependsOnID,
						Reason:         h.Reason,
						Source:         "planning_doc",
					})
				}
			}
		}
	}

	if suggestions == nil {
		suggestions = []SuggestedDependency{}
	}

	result := map[string]interface{}{
		"success":     true,
		"suggestions": suggestions,
		"count":       len(suggestions),
		"apply_hint":  "Use task_workflow action=update with task_ids and dependencies to apply.",
	}

	resultJSON, _ := json.Marshal(result)

	outputPath, _ := params["output_path"].(string)
	if outputPath != "" {
		if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err == nil {
			_ = os.WriteFile(outputPath, resultJSON, 0644)
		}
	}

	resp := &proto.TaskAnalysisResponse{
		Action:     "suggest_dependencies",
		OutputPath: outputPath,
		ResultJson: string(resultJSON),
	}

	return response.FormatResult(TaskAnalysisResponseToMap(resp), resp.GetOutputPath())
}

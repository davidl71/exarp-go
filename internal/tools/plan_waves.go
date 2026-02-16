package tools

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"

	"github.com/davidl71/exarp-go/internal/config"
)

// DefaultPlanWavesPath is the default path for the parallel execution plan research markdown,
// relative to project root. The TUI uses this file for waves when available.
const DefaultPlanWavesPath = "docs/PARALLEL_EXECUTION_PLAN_RESEARCH.md"

// taskIDRegex matches **T-123** or **T-0** style task IDs in markdown.
var taskIDRegex = regexp.MustCompile(`\*\*T-\d+\*\*`)

// phaseHeadingRegex matches ## Phase N: or ## Phase N - style headings.
var phaseHeadingRegex = regexp.MustCompile(`^##\s+Phase\s+(\d+)\s*[:\-]`)

// ParseWavesFromPlanMarkdown parses a PARALLEL_EXECUTION_PLAN_RESEARCH-style markdown file
// and returns waves as map[int][]string (wave level -> task IDs).
// Each "## Phase N:" section becomes wave (N-1). Task IDs are extracted from lines
// containing **T-NNN** patterns (e.g., "1. **T-0** - description").
func ParseWavesFromPlanMarkdown(path string) (map[int][]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	waves := make(map[int][]string)
	var currentWave int = -1
	var currentIDs []string

	flushWave := func() {
		if currentWave >= 0 && len(currentIDs) > 0 {
			waves[currentWave] = currentIDs
		}
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()

		if m := phaseHeadingRegex.FindStringSubmatch(line); len(m) >= 2 {
			flushWave()
			var phaseNum int
			if _, err := fmt.Sscanf(m[1], "%d", &phaseNum); err == nil && phaseNum >= 1 {
				currentWave = phaseNum - 1 // Phase 1 -> wave 0
			} else {
				currentWave = -1
			}
			currentIDs = nil
			continue
		}

		// Within a phase, collect task IDs
		matches := taskIDRegex.FindAllString(line, -1)
		for _, m := range matches {
			// m is "**T-123**", extract "T-123"
			id := m[2 : len(m)-2] // strip **
			if id != "" {
				currentIDs = append(currentIDs, id)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	flushWave()

	if len(waves) == 0 {
		return nil, nil
	}

	// Ensure wave keys are contiguous 0..N for consistent display
	levels := make([]int, 0, len(waves))
	for k := range waves {
		levels = append(levels, k)
	}
	sort.Ints(levels)
	normalized := make(map[int][]string)
	for i, level := range levels {
		normalized[i] = waves[level]
	}
	return normalized, nil
}

// LoadWavesFromPlanFile loads waves from the default plan file at
// projectRoot/docs/PARALLEL_EXECUTION_PLAN_RESEARCH.md. Returns nil map and nil error
// if the file does not exist or is empty.
func LoadWavesFromPlanFile(projectRoot string) (map[int][]string, error) {
	if projectRoot == "" {
		return nil, nil
	}
	path := filepath.Join(projectRoot, DefaultPlanWavesPath)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, nil
	}
	return ParseWavesFromPlanMarkdown(path)
}

// filterWavesToBacklogOnly removes Done tasks from waves. Only Todo and In Progress remain.
// Drops empty waves and renormalizes indices to contiguous 0..N.
func filterWavesToBacklogOnly(waves map[int][]string, taskList []Todo2Task) map[int][]string {
	taskByID := make(map[string]Todo2Task)
	for _, t := range taskList {
		taskByID[t.ID] = t
	}
	levels := make([]int, 0, len(waves))
	for k := range waves {
		levels = append(levels, k)
	}
	sort.Ints(levels)
	filtered := make(map[int][]string)
	idx := 0
	for _, level := range levels {
		ids := waves[level]
		keep := make([]string, 0, len(ids))
		for _, id := range ids {
			if t, ok := taskByID[id]; ok && IsBacklogStatus(t.Status) {
				keep = append(keep, id)
			}
		}
		if len(keep) > 0 {
			filtered[idx] = keep
			idx++
		}
	}
	return filtered
}

// ComputeWavesForTUI returns waves for the TUI waves view. It prefers waves from
// docs/PARALLEL_EXECUTION_PLAN_RESEARCH.md when that file exists and has content;
// otherwise falls back to BacklogExecutionOrder with optional LimitWavesByMaxTasks.
// Waves include only Todo and In Progress tasks; Done tasks are excluded.
func ComputeWavesForTUI(projectRoot string, taskList []Todo2Task) (map[int][]string, error) {
	var waves map[int][]string
	var err error
	waves, err = LoadWavesFromPlanFile(projectRoot)
	if err == nil && waves != nil && len(waves) > 0 {
		waves = filterWavesToBacklogOnly(waves, taskList)
		return waves, nil
	}
	_, waves, _, err = BacklogExecutionOrder(taskList, nil)
	if err != nil {
		return nil, err
	}
	if max := config.MaxTasksPerWave(); max > 0 {
		waves = LimitWavesByMaxTasks(waves, max)
	}
	return waves, nil
}

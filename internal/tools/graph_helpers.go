package tools

import (
	"context"
	"fmt"
	"sort"

	"github.com/davidl71/exarp-go/internal/models"
	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/graph/topo"
)

// TaskNode represents a task as a graph node with metadata.
type TaskNode struct {
	NodeID   int64  // Internal node ID for graph
	TaskID   string // Task ID from Todo2
	Name     string
	Priority string
	Status   string
}

// ID returns the node ID (required for graph.Node interface).
func (n *TaskNode) ID() int64 {
	return n.NodeID
}

// TaskGraph wraps a gonum DirectedGraph with task ID mapping.
type TaskGraph struct {
	Graph     *simple.DirectedGraph
	TaskIDMap map[string]int64    // task ID -> node ID
	NodeIDMap map[int64]string    // node ID -> task ID
	NodeMap   map[int64]*TaskNode // node ID -> node
	nodeIDSeq int64
}

// NewTaskGraph creates a new task graph.
func NewTaskGraph() *TaskGraph {
	return &TaskGraph{
		Graph:     simple.NewDirectedGraph(),
		TaskIDMap: make(map[string]int64),
		NodeIDMap: make(map[int64]string),
		NodeMap:   make(map[int64]*TaskNode),
		nodeIDSeq: 0,
	}
}

// AddTask adds a task to the graph and returns its node ID.
func (tg *TaskGraph) AddTask(task Todo2Task) int64 {
	// Check if task already exists
	if nodeID, exists := tg.TaskIDMap[task.ID]; exists {
		return nodeID
	}

	// Create new node
	tg.nodeIDSeq++
	nodeID := tg.nodeIDSeq

	node := &TaskNode{
		NodeID:   nodeID,
		TaskID:   task.ID,
		Name:     task.Content,
		Priority: task.Priority,
		Status:   task.Status,
	}

	tg.Graph.AddNode(node)
	tg.TaskIDMap[task.ID] = nodeID
	tg.NodeIDMap[nodeID] = task.ID
	tg.NodeMap[nodeID] = node

	return nodeID
}

// AddDependency adds a dependency edge from depTaskID to taskID.
func (tg *TaskGraph) AddDependency(depTaskID, taskID string) error {
	depNodeID, depExists := tg.TaskIDMap[depTaskID]
	taskNodeID, taskExists := tg.TaskIDMap[taskID]

	if !depExists || !taskExists {
		return fmt.Errorf("task not found: dep=%s, task=%s", depTaskID, taskID)
	}

	depNode := tg.Graph.Node(depNodeID)
	taskNode := tg.Graph.Node(taskNodeID)

	if depNode == nil || taskNode == nil {
		return fmt.Errorf("node not found in graph")
	}

	edge := tg.Graph.NewEdge(depNode, taskNode)
	tg.Graph.SetEdge(edge)

	return nil
}

// BuildTaskGraph builds a gonum DirectedGraph from tasks.
func BuildTaskGraph(tasks []Todo2Task) (*TaskGraph, error) {
	tg := NewTaskGraph()

	// First pass: add all tasks as nodes
	taskMap := make(map[string]bool)
	for _, task := range tasks {
		taskMap[task.ID] = true

		tg.AddTask(task)
	}

	// Second pass: add dependency edges (Dependencies + parent_id for wave ordering)
	for _, task := range tasks {
		// Explicit blocking dependencies
		for _, depID := range task.Dependencies {
			if taskMap[depID] {
				if err := tg.AddDependency(depID, task.ID); err != nil {
					return nil, fmt.Errorf("failed to add dependency %s -> %s: %w", depID, task.ID, err)
				}
			}
		}
		// parent_id: treat as dependency for wave ordering (parent before child)
		if task.ParentID != "" && taskMap[task.ParentID] {
			if err := tg.AddDependency(task.ParentID, task.ID); err != nil {
				return nil, fmt.Errorf("failed to add parent dependency %s -> %s: %w", task.ParentID, task.ID, err)
			}
		}
	}

	return tg, nil
}

// BuildTaskGraphBacklogOnly builds a task graph from only Todo and In Progress tasks.
// Done tasks are excluded by default so critical path focuses on remaining work.
func BuildTaskGraphBacklogOnly(tasks []Todo2Task) (*TaskGraph, error) {
	var backlog []Todo2Task

	for _, t := range tasks {
		if IsBacklogStatus(t.Status) {
			backlog = append(backlog, t)
		}
	}

	return BuildTaskGraph(backlog)
}

// HasCycles checks if the graph has cycles using topological sort.
func HasCycles(tg *TaskGraph) (bool, error) {
	_, err := topo.Sort(tg.Graph)
	if err != nil {
		// If topological sort fails, graph has cycles
		return true, nil
	}

	return false, nil
}

// DetectCycles enumerates all cycles in the graph
// Gonum doesn't have simple_cycles, so we implement it using DFS.
func DetectCycles(tg *TaskGraph) [][]string {
	cycles := [][]string{}
	visited := make(map[int64]bool)
	recStack := make(map[int64]bool)

	var dfs func(nodeID int64, path []int64)

	dfs = func(nodeID int64, path []int64) {
		visited[nodeID] = true
		recStack[nodeID] = true

		path = append(path, nodeID)

		// Check all outgoing edges
		toNodes := tg.Graph.From(nodeID)
		for toNodes.Next() {
			neighborID := toNodes.Node().ID()

			if !visited[neighborID] {
				dfs(neighborID, path)
			} else if recStack[neighborID] {
				// Found cycle - find the start of the cycle in path
				cycleStart := -1

				for i, n := range path {
					if n == neighborID {
						cycleStart = i
						break
					}
				}

				if cycleStart >= 0 {
					// Build cycle with task IDs
					cycle := []string{}

					for _, n := range path[cycleStart:] {
						if taskID, ok := tg.NodeIDMap[n]; ok {
							cycle = append(cycle, taskID)
						}
					}
					// Add the neighbor to complete the cycle
					if taskID, ok := tg.NodeIDMap[neighborID]; ok {
						cycle = append(cycle, taskID)
					}

					if len(cycle) > 1 {
						cycles = append(cycles, cycle)
					}
				}
			}
		}

		recStack[nodeID] = false
	}

	// Visit all nodes
	allNodes := tg.Graph.Nodes()
	for allNodes.Next() {
		nodeID := allNodes.Node().ID()
		if !visited[nodeID] {
			dfs(nodeID, []int64{})
		}
	}

	// Deduplicate cycles (same cycle can be found starting from different nodes)
	return deduplicateCycles(cycles)
}

// deduplicateCycles removes duplicate cycles (same cycle, different starting point).
func deduplicateCycles(cycles [][]string) [][]string {
	seen := make(map[string]bool)
	result := [][]string{}

	for _, cycle := range cycles {
		if len(cycle) == 0 {
			continue
		}

		// Normalize cycle by rotating to start with smallest task ID
		normalized := normalizeCycle(cycle)
		key := cycleKey(normalized)

		if !seen[key] {
			seen[key] = true

			result = append(result, normalized)
		}
	}

	return result
}

// normalizeCycle rotates cycle to start with smallest task ID.
func normalizeCycle(cycle []string) []string {
	if len(cycle) <= 1 {
		return cycle
	}

	// Find smallest task ID
	minIdx := 0

	minID := cycle[0]
	for i, id := range cycle {
		if id < minID {
			minID = id
			minIdx = i
		}
	}

	// Rotate to start with min
	result := make([]string, len(cycle))
	copy(result, cycle[minIdx:])
	copy(result[len(cycle)-minIdx:], cycle[:minIdx])

	return result
}

// cycleKey creates a unique key for a cycle.
func cycleKey(cycle []string) string {
	return fmt.Sprintf("%v", cycle)
}

// TopoSortTasks returns tasks in topological order (dependency order).
func TopoSortTasks(tg *TaskGraph) ([]string, error) {
	sortedNodes, err := topo.Sort(tg.Graph)
	if err != nil {
		return nil, fmt.Errorf("graph has cycles, cannot sort: %w", err)
	}

	result := make([]string, len(sortedNodes))

	for i, node := range sortedNodes {
		if taskID, ok := tg.NodeIDMap[node.ID()]; ok {
			result[i] = taskID
		}
	}

	return result, nil
}

// FindCriticalPath finds the longest path in the DAG (critical path)
// Assumes graph is a DAG (no cycles).
func FindCriticalPath(tg *TaskGraph) ([]string, error) {
	// First check if it's a DAG
	hasCycles, err := HasCycles(tg)
	if err != nil {
		return nil, fmt.Errorf("failed to check for cycles: %w", err)
	}

	if hasCycles {
		return nil, fmt.Errorf("cannot find critical path in graph with cycles")
	}

	// Get topological sort
	sorted, err := TopoSortTasks(tg)
	if err != nil {
		return nil, err
	}

	// Calculate longest path using dynamic programming
	dist := make(map[string]int)    // distance from start
	prev := make(map[string]string) // previous node in longest path
	maxDist := 0
	endNode := ""

	// Initialize distances
	for _, taskID := range sorted {
		dist[taskID] = 0
	}

	// Process nodes in topological order
	for _, taskID := range sorted {
		nodeID, exists := tg.TaskIDMap[taskID]
		if !exists {
			continue
		}

		// Check all incoming edges
		fromNodes := tg.Graph.To(nodeID)
		for fromNodes.Next() {
			fromNodeID := fromNodes.Node().ID()
			fromTaskID := tg.NodeIDMap[fromNodeID]

			if dist[fromTaskID]+1 > dist[taskID] {
				dist[taskID] = dist[fromTaskID] + 1
				prev[taskID] = fromTaskID
			}
		}

		// Track maximum distance
		if dist[taskID] > maxDist {
			maxDist = dist[taskID]
			endNode = taskID
		}
	}

	// Reconstruct path from end to start
	path := []string{}
	current := endNode

	for current != "" {
		path = append([]string{current}, path...)
		current = prev[current]
	}

	return path, nil
}

// GetTaskLevel calculates the dependency level of each task (0 = no dependencies)
// Optimized version using topological sort for O(V + E) performance.
func GetTaskLevels(tg *TaskGraph) map[string]int {
	levels := make(map[string]int)

	// Check if graph has cycles - if so, use iterative approach
	hasCycles, err := HasCycles(tg)
	if err != nil || hasCycles {
		// Fall back to iterative approach for cyclic graphs
		return getTaskLevelsIterative(tg)
	}

	// Optimized: Use topological sort for acyclic graphs
	sortedNodes, err := topo.Sort(tg.Graph)
	if err != nil {
		// Fall back if topological sort fails
		return getTaskLevelsIterative(tg)
	}

	// Process nodes in topological order (dependencies before dependents)
	// This allows single-pass level calculation
	for _, node := range sortedNodes {
		nodeID := node.ID()
		taskID := tg.NodeIDMap[nodeID]

		// Check all incoming edges (dependencies)
		maxDepLevel := -1

		fromNodes := tg.Graph.To(nodeID)
		for fromNodes.Next() {
			fromNodeID := fromNodes.Node().ID()
			fromTaskID := tg.NodeIDMap[fromNodeID]

			if depLevel, ok := levels[fromTaskID]; ok && depLevel > maxDepLevel {
				maxDepLevel = depLevel
			}
		}

		// Level is one more than max dependency level
		levels[taskID] = maxDepLevel + 1
	}

	return levels
}

// getTaskLevelsIterative is the fallback iterative approach for cyclic graphs
// Optimized with iteration limit and tracking of changed nodes only.
func getTaskLevelsIterative(tg *TaskGraph) map[string]int {
	levels := make(map[string]int)

	const maxIterations = 1000 // Safety limit to prevent infinite loops

	// Initialize all tasks to level 0
	nodes := tg.Graph.Nodes()
	allTaskIDs := make([]string, 0)
	taskIDSet := make(map[string]bool)

	for nodes.Next() {
		nodeID := nodes.Node().ID()
		if taskID, ok := tg.NodeIDMap[nodeID]; ok {
			levels[taskID] = 0

			allTaskIDs = append(allTaskIDs, taskID)
			taskIDSet[taskID] = true
		}
	}

	// Calculate levels using iterative approach until convergence
	// Optimization: Track which nodes changed to only process affected nodes
	changedNodes := make(map[string]bool)
	for _, taskID := range allTaskIDs {
		changedNodes[taskID] = true // Start with all nodes
	}

	iteration := 0
	for len(changedNodes) > 0 && iteration < maxIterations {
		iteration++
		nextChanged := make(map[string]bool)

		// Only process nodes that changed or depend on changed nodes
		for taskID := range changedNodes {
			nodeID, exists := tg.TaskIDMap[taskID]
			if !exists {
				continue
			}

			// Check all incoming edges
			fromNodes := tg.Graph.To(nodeID)
			maxDepLevel := -1

			for fromNodes.Next() {
				fromNodeID := fromNodes.Node().ID()
				fromTaskID := tg.NodeIDMap[fromNodeID]

				if depLevel, ok := levels[fromTaskID]; ok && depLevel > maxDepLevel {
					maxDepLevel = depLevel
				}
			}

			newLevel := maxDepLevel + 1
			if newLevel > levels[taskID] {
				levels[taskID] = newLevel

				// Mark dependent nodes for next iteration
				toNodes := tg.Graph.From(nodeID)
				for toNodes.Next() {
					toNodeID := toNodes.Node().ID()
					if toTaskID, ok := tg.NodeIDMap[toNodeID]; ok && taskIDSet[toTaskID] {
						nextChanged[toTaskID] = true
					}
				}
			}
		}

		changedNodes = nextChanged
	}

	return levels
}

// AnalyzeCriticalPath analyzes and returns critical path information for tasks.
func AnalyzeCriticalPath(projectRoot string) (map[string]interface{}, error) {
	store := NewDefaultTaskStore(projectRoot)

	list, err := store.ListTasks(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	tasks := tasksFromPtrs(list)

	result := map[string]interface{}{
		"total_tasks":       len(tasks),
		"has_critical_path": false,
	}

	if len(tasks) == 0 {
		return result, nil
	}

	// Build graph from backlog only (Todo + In Progress); exclude Done by default
	tg, err := BuildTaskGraphBacklogOnly(tasks)
	if err != nil {
		return nil, fmt.Errorf("failed to build graph: %w", err)
	}

	// No backlog tasks — nothing on critical path
	if tg.Graph.Nodes().Len() == 0 {
		result["message"] = "No critical path: all tasks are Done"
		return result, nil
	}

	// Check for cycles
	hasCycles, err := HasCycles(tg)
	if err != nil {
		return nil, fmt.Errorf("failed to check cycles: %w", err)
	}

	if hasCycles {
		cycles := DetectCycles(tg)
		result["has_cycles"] = true
		result["cycle_count"] = len(cycles)
		result["cycles"] = cycles
		result["message"] = "Cannot compute critical path: graph has cycles"

		return result, nil
	}

	// Find critical path (longest chain among backlog tasks)
	criticalPath, err := FindCriticalPath(tg)
	if err != nil {
		return nil, fmt.Errorf("failed to find critical path: %w", err)
	}

	// Build detailed path information
	pathDetails := []map[string]interface{}{}

	for _, taskID := range criticalPath {
		for _, task := range tasks {
			if task.ID == taskID {
				pathDetails = append(pathDetails, map[string]interface{}{
					"id":                 task.ID,
					"content":            task.Content,
					"priority":           task.Priority,
					"status":             task.Status,
					"dependencies":       task.Dependencies,
					"dependencies_count": len(task.Dependencies),
				})

				break
			}
		}
	}

	// Get task levels for all tasks
	levels := GetTaskLevels(tg)

	result["has_critical_path"] = true
	result["critical_path"] = criticalPath
	result["critical_path_length"] = len(criticalPath)
	result["critical_path_details"] = pathDetails
	result["all_task_levels"] = levels

	// Calculate max level (should match critical path length - 1)
	maxLevel := 0
	for _, level := range levels {
		if level > maxLevel {
			maxLevel = level
		}
	}

	result["max_dependency_level"] = maxLevel

	return result, nil
}

// BacklogTaskDetail holds task info for execution-order output.
type BacklogTaskDetail struct {
	ID       string   `json:"id"`
	Content  string   `json:"content"`
	Priority string   `json:"priority"`
	Status   string   `json:"status"`
	Level    int      `json:"level"`
	Tags     []string `json:"tags,omitempty"`
}

// IsBacklogStatus returns true if status is Todo or In Progress.
func IsBacklogStatus(status string) bool {
	s := NormalizeStatusToTitleCase(status)
	return s == models.StatusTodo || s == models.StatusInProgress
}

// priorityOrderForSort returns sort key for priority (lower = higher priority).
func priorityOrderForSort(priority string) int {
	p := NormalizePriority(priority)
	switch p {
	case models.PriorityCritical:
		return 0
	case models.PriorityHigh:
		return 1
	case models.PriorityMedium:
		return 2
	case models.PriorityLow:
		return 3
	default:
		return 4
	}
}

// BacklogExecutionOrder returns backlog tasks (Todo + In Progress) in dependency order.
// Uses full task set for graph so levels are correct; sorts backlog by level asc, then priority, then ID.
// If backlogFilter is non-nil, only tasks whose ID is in backlogFilter are included in the backlog.
// Returns ordered IDs, waves (level -> task IDs), and details. If graph has cycles, levels are best-effort.
func BacklogExecutionOrder(tasks []Todo2Task, backlogFilter map[string]bool) (orderedIDs []string, waves map[int][]string, details []BacklogTaskDetail, err error) {
	orderedIDs = []string{}
	waves = make(map[int][]string)
	details = []BacklogTaskDetail{}

	backlog := make([]Todo2Task, 0)

	taskMap := make(map[string]Todo2Task)
	for _, t := range tasks {
		taskMap[t.ID] = t

		if !IsBacklogStatus(t.Status) {
			continue
		}

		if backlogFilter != nil && !backlogFilter[t.ID] {
			continue
		}

		backlog = append(backlog, t)
	}

	if len(backlog) == 0 {
		return orderedIDs, waves, details, nil
	}

	tg, err := BuildTaskGraph(tasks)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("build task graph: %w", err)
	}

	levels := GetTaskLevels(tg)

	// Sort backlog by level asc, then priority (critical first), then ID
	sort.Slice(backlog, func(i, j int) bool {
		li, lj := levels[backlog[i].ID], levels[backlog[j].ID]
		if li != lj {
			return li < lj
		}

		pi := priorityOrderForSort(backlog[i].Priority)
		pj := priorityOrderForSort(backlog[j].Priority)

		if pi != pj {
			return pi < pj
		}

		return backlog[i].ID < backlog[j].ID
	})

	for _, t := range backlog {
		orderedIDs = append(orderedIDs, t.ID)
		level := levels[t.ID]
		waves[level] = append(waves[level], t.ID)

		tags := t.Tags
		if tags == nil {
			tags = []string{}
		}

		details = append(details, BacklogTaskDetail{
			ID:       t.ID,
			Content:  t.Content,
			Priority: t.Priority,
			Status:   t.Status,
			Level:    level,
			Tags:     tags,
		})
	}

	return orderedIDs, waves, details, nil
}

// LimitWavesByMaxTasks splits waves so each has at most maxPerWave task IDs.
// Levels are preserved in order; waves with more than maxPerWave tasks are split into
// consecutive waves (sequential indices 0, 1, 2, ...). If maxPerWave <= 0, returns waves unchanged.
func LimitWavesByMaxTasks(waves map[int][]string, maxPerWave int) map[int][]string {
	if maxPerWave <= 0 || len(waves) == 0 {
		return waves
	}

	levels := make([]int, 0, len(waves))
	for k := range waves {
		levels = append(levels, k)
	}

	sort.Ints(levels)

	out := make(map[int][]string)
	idx := 0

	for _, level := range levels {
		ids := waves[level]
		for start := 0; start < len(ids); start += maxPerWave {
			end := start + maxPerWave
			if end > len(ids) {
				end = len(ids)
			}

			out[idx] = append([]string(nil), ids[start:end]...)
			idx++
		}
	}

	return out
}

// tui_sorting.go — Sorting, hierarchy, search/filter, and tree traversal for the Bubbletea TUI.
// Functions here compute visible task ordering, hierarchy depth, collapse state, and search matches.
package cli

import (
	"sort"
	"strings"

	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/models"
	"github.com/davidl71/exarp-go/internal/tools"
)

// sortTasksBy sorts tasks in place by order (id|status|priority|updated) and direction.
func sortTasksBy(tasks []*database.Todo2Task, order string, asc bool) {
	if len(tasks) == 0 {
		return
	}

	less := func(i, j int) bool {
		a, b := tasks[i], tasks[j]

		var cmp int

		switch order {
		case SortByStatus:
			cmp = strings.Compare(strings.ToLower(a.Status), strings.ToLower(b.Status))
		case SortByPriority:
			cmp = strings.Compare(strings.ToLower(a.Priority), strings.ToLower(b.Priority))
		case SortByUpdated:
			cmp = strings.Compare(a.LastModified, b.LastModified)
		default:
			cmp = strings.Compare(a.ID, b.ID)
		}

		if cmp == 0 {
			cmp = strings.Compare(a.ID, b.ID)
		}

		if asc {
			return cmp < 0
		}

		return cmp > 0
	}
	sort.Slice(tasks, less)
}

// priorityOrderForSort returns sort key for priority (lower = higher priority).
func priorityOrderForSort(p string) int {
	switch strings.ToLower(p) {
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

// computeHierarchyOrder builds hierarchyOrder and hierarchyDepth from task graph (dependencies + parent_id).
func (m *model) computeHierarchyOrder() {
	if len(m.tasks) == 0 {
		m.hierarchyOrder = nil
		m.hierarchyDepth = nil
		m.hierarchyDepthByID = nil

		return
	}

	taskList := make([]tools.Todo2Task, 0, len(m.tasks))

	for _, p := range m.tasks {
		if p != nil {
			taskList = append(taskList, *p)
		}
	}

	tg, err := tools.BuildTaskGraph(taskList)
	if err != nil {
		m.hierarchyOrder = nil
		m.hierarchyDepth = nil
		m.hierarchyDepthByID = nil

		return
	}

	levels := tools.GetTaskLevels(tg)
	m.hierarchyDepthByID = make(map[string]int)

	for id, level := range levels {
		m.hierarchyDepthByID[id] = level
	}

	type item struct {
		idx      int
		level    int
		priority string
		id       string
	}

	var items []item

	for i, p := range m.tasks {
		if p == nil {
			continue
		}

		level := levels[p.ID]
		items = append(items, item{i, level, p.Priority, p.ID})
	}

	asc := m.sortAsc

	sort.Slice(items, func(i, j int) bool {
		if items[i].level != items[j].level {
			if asc {
				return items[i].level < items[j].level
			}

			return items[i].level > items[j].level
		}

		pi := priorityOrderForSort(items[i].priority)
		pj := priorityOrderForSort(items[j].priority)

		if pi != pj {
			if asc {
				return pi < pj
			}

			return pi > pj
		}

		if asc {
			return items[i].id < items[j].id
		}

		return items[i].id > items[j].id
	})

	m.hierarchyOrder = make([]int, len(items))
	m.hierarchyDepth = make(map[int]int)

	for i, it := range items {
		m.hierarchyOrder[i] = it.idx
		m.hierarchyDepth[it.idx] = it.level
	}
}

// taskParentMap returns task ID -> parent task ID for all tasks that have a parent.
func (m model) taskParentMap() map[string]string {
	out := make(map[string]string, len(m.tasks))

	for _, t := range m.tasks {
		if t.ParentID != "" {
			out[t.ID] = t.ParentID
		}
	}

	return out
}

// taskAncestorIDs returns for each task ID the set of ancestor IDs (parent + dependencies).
// Used to hide both parent_id descendants and dependency dependents when a node is collapsed.
func (m model) taskAncestorIDs() map[string][]string {
	out := make(map[string][]string, len(m.tasks))
	taskIDs := make(map[string]struct{}, len(m.tasks))

	for _, t := range m.tasks {
		if t == nil {
			continue
		}

		taskIDs[t.ID] = struct{}{}
	}

	for _, t := range m.tasks {
		if t == nil {
			continue
		}

		var ancestors []string

		if t.ParentID != "" {
			if _, ok := taskIDs[t.ParentID]; ok {
				ancestors = append(ancestors, t.ParentID)
			}
		}

		for _, depID := range t.Dependencies {
			if _, ok := taskIDs[depID]; ok {
				ancestors = append(ancestors, depID)
			}
		}

		if len(ancestors) > 0 {
			out[t.ID] = ancestors
		}
	}

	return out
}

// isDescendantOfCollapsed returns true if taskID has an ancestor in m.collapsedTaskIDs,
// following both parent_id and dependency links (so collapsing hides parent children and dependents).
func (m model) isDescendantOfCollapsed(taskID string, ancestorIDs map[string][]string) bool {
	seen := make(map[string]struct{})

	var queue []string

	queue = append(queue, taskID)
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]

		if _, ok := seen[id]; ok {
			continue
		}

		seen[id] = struct{}{}

		for _, anc := range ancestorIDs[id] {
			if _, collapsed := m.collapsedTaskIDs[anc]; collapsed {
				return true
			}

			queue = append(queue, anc)
		}
	}

	return false
}

// taskHasChildren returns true if the task has real children (ParentID or Dependencies)
// or is a synthetic group head (first in a status/priority run with more tasks in that run).
// Synthetic grouping makes Tab collapse useful when no parent_id/dependencies exist.
func (m model) taskHasChildren(taskID string) bool {
	// Real hierarchy children
	for _, t := range m.tasks {
		if t == nil {
			continue
		}
		if t.ParentID == taskID {
			return true
		}
		for _, depID := range t.Dependencies {
			if depID == taskID {
				return true
			}
		}
	}

	// Synthetic group: when sorted by status or priority, first task in each run can "collapse" the rest
	var taskIdx int
	found := false
	for i, t := range m.tasks {
		if t != nil && t.ID == taskID {
			taskIdx = i
			found = true
			break
		}
	}
	if !found || len(m.tasks) == 0 {
		return false
	}

	task := m.tasks[taskIdx]
	if task == nil {
		return false
	}

	switch m.sortOrder {
	case SortByStatus:
		// First in status run? (previous task has different status)
		firstInRun := taskIdx == 0 || m.tasks[taskIdx-1].Status != task.Status
		// More in same run?
		hasMore := taskIdx+1 < len(m.tasks) && m.tasks[taskIdx+1].Status == task.Status
		return firstInRun && hasMore
	case SortByPriority:
		firstInRun := taskIdx == 0 || m.tasks[taskIdx-1].Priority != task.Priority
		hasMore := taskIdx+1 < len(m.tasks) && m.tasks[taskIdx+1].Priority == task.Priority
		return firstInRun && hasMore
	default:
		return false
	}
}

// syntheticGroupMemberHidden returns true if task at index i is hidden because a collapsed
// synthetic group head (same status or priority run) appears before it in display order.
func (m model) syntheticGroupMemberHidden(idx int, base []int) bool {
	if len(m.collapsedTaskIDs) == 0 || len(base) == 0 {
		return false
	}
	task := m.tasks[idx]
	if task == nil {
		return false
	}
	// Find position of idx in base
	var posInBase int
	for p, b := range base {
		if b == idx {
			posInBase = p
			break
		}
	}
	// Check if any collapsed task is the group head for this run
	switch m.sortOrder {
	case SortByStatus:
		// Walk backward in base to find group start
		groupStart := posInBase
		for groupStart > 0 && m.tasks[base[groupStart-1]] != nil && m.tasks[base[groupStart-1]].Status == task.Status {
			groupStart--
		}
		if groupStart == posInBase {
			return false // we are the group head
		}
		headIdx := base[groupStart]
		if m.tasks[headIdx] == nil {
			return false
		}
		_, collapsed := m.collapsedTaskIDs[m.tasks[headIdx].ID]
		return collapsed
	case SortByPriority:
		groupStart := posInBase
		for groupStart > 0 && m.tasks[base[groupStart-1]] != nil && m.tasks[base[groupStart-1]].Priority == task.Priority {
			groupStart--
		}
		if groupStart == posInBase {
			return false
		}
		headIdx := base[groupStart]
		if m.tasks[headIdx] == nil {
			return false
		}
		_, collapsed := m.collapsedTaskIDs[m.tasks[headIdx].ID]
		return collapsed
	default:
		return false
	}
}

// visibleIndices returns the indices of tasks to display (filtered by search, collapse, or all).
func (m model) visibleIndices() []int {
	if len(m.tasks) == 0 {
		return nil
	}

	var base []int
	if m.sortOrder == "hierarchy" && len(m.hierarchyOrder) > 0 {
		base = m.hierarchyOrder
	} else {
		// Non-hierarchy sort: show only root tasks so subtasks don't appear as independent rows.
		// When search is active, also include matching subtasks so search can find them.
		filterSet := make(map[int]struct{})
		if m.filteredIndices != nil {
			for _, i := range m.filteredIndices {
				filterSet[i] = struct{}{}
			}
		}
		for i := range m.tasks {
			if m.tasks[i] == nil {
				continue
			}
			isRoot := m.tasks[i].ParentID == ""
			_, inFilter := filterSet[i]
			// Show root tasks; when search is active also show matching subtasks.
			if (isRoot && len(filterSet) == 0) || inFilter {
				base = append(base, i)
			}
		}
	}

	if m.filteredIndices != nil {
		filterSet := make(map[int]struct{})
		for _, i := range m.filteredIndices {
			filterSet[i] = struct{}{}
		}

		var out []int

		for _, i := range base {
			if _, ok := filterSet[i]; ok {
				out = append(out, i)
			}
		}

		base = out
	}
	// Hide descendants of collapsed tasks (parent_id children and dependency dependents)
	// and hide synthetic group members (same status/priority run when collapsed)
	ancestorIDs := m.taskAncestorIDs()

	var final []int
	for _, i := range base {
		if m.tasks[i] == nil {
			continue
		}
		if m.isDescendantOfCollapsed(m.tasks[i].ID, ancestorIDs) {
			continue
		}
		if m.syntheticGroupMemberHidden(i, base) {
			continue
		}
		final = append(final, i)
	}

	return final
}

// realIndexAt returns the index into m.tasks for the current cursor position.
func (m model) realIndexAt(cursorPos int) int {
	vis := m.visibleIndices()
	if vis == nil || cursorPos < 0 || cursorPos >= len(vis) {
		return 0
	}

	return vis[cursorPos]
}

// computeFilteredIndices returns indices of m.tasks that match m.searchQuery (case-insensitive).
func (m model) computeFilteredIndices() []int {
	q := strings.TrimSpace(strings.ToLower(m.searchQuery))
	if q == "" {
		return nil
	}

	var out []int

	for i, t := range m.tasks {
		if taskMatchesSearch(t, q) {
			out = append(out, i)
		}
	}

	return out
}

func taskMatchesSearch(t *database.Todo2Task, q string) bool {
	if strings.Contains(strings.ToLower(t.ID), q) {
		return true
	}

	if strings.Contains(strings.ToLower(t.Content), q) {
		return true
	}

	if strings.Contains(strings.ToLower(t.LongDescription), q) {
		return true
	}

	if strings.Contains(strings.ToLower(t.Status), q) {
		return true
	}

	if strings.Contains(strings.ToLower(t.Priority), q) {
		return true
	}

	for _, tag := range t.Tags {
		if strings.Contains(strings.ToLower(tag), q) {
			return true
		}
	}

	return false
}

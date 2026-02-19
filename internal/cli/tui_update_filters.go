// tui_update_filters.go — Sort and filter key handlers for TUI Update().
// Handles cycle sort (o), sort direction (O), search (/), next/prev match (n/N), and collapse (tab).
package cli

import (
	tea "github.com/charmbracelet/bubbletea"
)

// handleSortFilterKeys handles keys that change sort order, sort direction, search, or collapse.
// Returns (model, cmd, true) when the key was handled.
func (m model) handleSortFilterKeys(key string) (model, tea.Cmd, bool) {
	switch key {
	case "o":
		// Cycle sort order (id → status → priority → updated → hierarchy → id)
		if m.mode == ModeTasks && len(m.tasks) > 0 {
			switch m.sortOrder {
			case SortByID:
				m.sortOrder = SortByStatus
			case SortByStatus:
				m.sortOrder = SortByPriority
			case SortByPriority:
				m.sortOrder = SortByUpdated
			case SortByUpdated:
				m.sortOrder = SortByHierarchy
			default:
				m.sortOrder = SortByID
			}

			if m.sortOrder == SortByHierarchy {
				m.computeHierarchyOrder()
			} else {
				sortTasksBy(m.tasks, m.sortOrder, m.sortAsc)
			}

			if m.cursor >= len(m.visibleIndices()) {
				m.cursor = len(m.visibleIndices()) - 1
			}
		}
		return m, nil, true

	case "O":
		// Toggle sort direction (asc ↔ desc)
		if m.mode == ModeTasks && len(m.tasks) > 0 {
			m.sortAsc = !m.sortAsc
			if m.sortOrder == SortByHierarchy {
				m.computeHierarchyOrder()
			} else {
				sortTasksBy(m.tasks, m.sortOrder, m.sortAsc)
			}
		}
		return m, nil, true

	case "/":
		// Start search/filter (vim-style)
		if m.mode == ModeTasks {
			m.searchMode = true
		}
		return m, nil, true

	case "n":
		// Next search match (vim-style)
		if m.mode == ModeTasks && m.searchQuery != "" {
			vis := m.visibleIndices()
			if len(vis) > 0 && m.cursor < len(vis)-1 {
				m.cursor++
			}
		}
		return m, nil, true

	case "N":
		// Previous search match (vim-style)
		if m.mode == ModeTasks && m.searchQuery != "" {
			if m.cursor > 0 {
				m.cursor--
			}
		}
		return m, nil, true

	case "tab", "\t":
		// In tasks mode: collapse/expand tree node under cursor (if it has children)
		if m.mode == ModeTasks {
			vis := m.visibleIndices()
			if len(vis) > 0 && m.cursor < len(vis) {
				realIdx := m.realIndexAt(m.cursor)

				task := m.tasks[realIdx]
				if task != nil && m.taskHasChildren(task.ID) {
					if _, ok := m.collapsedTaskIDs[task.ID]; ok {
						delete(m.collapsedTaskIDs, task.ID)
					} else {
						m.collapsedTaskIDs[task.ID] = struct{}{}
					}
				}
			}
		}
		return m, nil, true
	}

	return m, nil, false
}

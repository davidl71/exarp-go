package cli

import "fmt"

var allowedModeTransitions = map[string]map[string]struct{}{
	ModeTasks: {
		ModeConfig:       {},
		ModeScorecard:    {},
		ModeHandoffs:     {},
		ModeWaves:        {},
		ModeJobs:         {},
		ModeTaskDetail:   {},
		ModeTaskAnalysis: {},
	},
	ModeConfig: {
		ModeTasks:         {},
		ModeConfigSection: {},
	},
	ModeConfigSection: {
		ModeConfig: {},
	},
	ModeScorecard: {
		ModeTasks: {},
	},
	ModeHandoffs: {
		ModeTasks: {},
	},
	ModeWaves: {
		ModeTasks:        {},
		ModeTaskAnalysis: {},
	},
	ModeTaskAnalysis: {
		ModeTasks: {},
		ModeWaves: {},
	},
	ModeJobs: {
		ModeTasks: {},
	},
	ModeTaskDetail: {
		ModeTasks: {},
	},
}

func (m *model) transitionTo(next string) bool {
	if next == "" {
		return false
	}

	if m.mode == next {
		return true
	}

	if allowed, ok := allowedModeTransitions[m.mode]; ok {
		if _, ok := allowed[next]; ok {
			m.mode = next
			return true
		}
	}

	m.err = fmt.Errorf("invalid mode transition: %s -> %s", m.mode, next)
	return false
}

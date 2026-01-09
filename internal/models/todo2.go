package models

// Todo2Task represents a Todo2 task
type Todo2Task struct {
	ID              string                 `json:"id"`
	Content         string                 `json:"content"`
	LongDescription string                 `json:"long_description,omitempty"`
	Status          string                 `json:"status"`
	Priority        string                 `json:"priority,omitempty"`
	Tags            []string               `json:"tags,omitempty"`
	Dependencies    []string               `json:"dependencies,omitempty"`
	Completed       bool                   `json:"completed,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// Todo2State represents the Todo2 state file structure
type Todo2State struct {
	Todos []Todo2Task `json:"todos"`
}

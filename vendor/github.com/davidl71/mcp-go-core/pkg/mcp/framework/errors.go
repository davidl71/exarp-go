package framework

import "fmt"

// ErrInvalidTool represents an invalid tool error
type ErrInvalidTool struct {
	ToolName string
	Reason   string
}

func (e *ErrInvalidTool) Error() string {
	return fmt.Sprintf("invalid tool %q: %s", e.ToolName, e.Reason)
}

// ErrToolNotFound represents a tool not found error
type ErrToolNotFound struct {
	ToolName string
}

func (e *ErrToolNotFound) Error() string {
	return fmt.Sprintf("tool %q not found", e.ToolName)
}

// ErrInvalidPrompt represents an invalid prompt error
type ErrInvalidPrompt struct {
	PromptName string
	Reason     string
}

func (e *ErrInvalidPrompt) Error() string {
	return fmt.Sprintf("invalid prompt %q: %s", e.PromptName, e.Reason)
}

// ErrPromptNotFound represents a prompt not found error
type ErrPromptNotFound struct {
	PromptName string
}

func (e *ErrPromptNotFound) Error() string {
	return fmt.Sprintf("prompt %q not found", e.PromptName)
}

// ErrInvalidResource represents an invalid resource error
type ErrInvalidResource struct {
	URI    string
	Reason string
}

func (e *ErrInvalidResource) Error() string {
	return fmt.Sprintf("invalid resource %q: %s", e.URI, e.Reason)
}

// ErrResourceNotFound represents a resource not found error
type ErrResourceNotFound struct {
	URI string
}

func (e *ErrResourceNotFound) Error() string {
	return fmt.Sprintf("resource %q not found", e.URI)
}

// Helper functions for error checking

// IsToolNotFound checks if error is ErrToolNotFound
func IsToolNotFound(err error) bool {
	_, ok := err.(*ErrToolNotFound)
	return ok
}

// IsPromptNotFound checks if error is ErrPromptNotFound
func IsPromptNotFound(err error) bool {
	_, ok := err.(*ErrPromptNotFound)
	return ok
}

// IsResourceNotFound checks if error is ErrResourceNotFound
func IsResourceNotFound(err error) bool {
	_, ok := err.(*ErrResourceNotFound)
	return ok
}

// IsInvalidTool checks if error is ErrInvalidTool
func IsInvalidTool(err error) bool {
	_, ok := err.(*ErrInvalidTool)
	return ok
}

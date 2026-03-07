package framework

import (
	"fmt"
	"strings"
)

// ToolError is a base error type for tool-related errors
type ToolError struct {
	ToolName string
	Message  string
	Err      error
}

func (e *ToolError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("tool %q: %s: %v", e.ToolName, e.Message, e.Err)
	}
	return fmt.Sprintf("tool %q: %s", e.ToolName, e.Message)
}

func (e *ToolError) Unwrap() error {
	return e.Err
}

// WrapToolError wraps an error with tool name and message
func WrapToolError(toolName, message string, err error) error {
	return &ToolError{ToolName: toolName, Message: message, Err: err}
}

// ParseError represents a JSON parse error
type ParseError struct {
	Message string
	Err     error
}

func (e *ParseError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("parse error: %s: %v", e.Message, e.Err)
	}
	return fmt.Sprintf("parse error: %s", e.Message)
}

func (e *ParseError) Unwrap() error {
	return e.Err
}

// ActionError represents an invalid action error
type ActionError struct {
	Action  string
	Message string
	Err     error
}

func (e *ActionError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("action %q: %s: %v", e.Action, e.Message, e.Err)
	}
	return fmt.Sprintf("action %q: %s", e.Action, e.Message)
}

func (e *ActionError) Unwrap() error {
	return e.Err
}

// UnknownActionError represents an unknown action error
type UnknownActionError struct {
	Action       string
	ValidActions []string
}

func (e *UnknownActionError) Error() string {
	return fmt.Sprintf("unknown action %q, valid actions: %s", e.Action, strings.Join(e.ValidActions, ", "))
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
	Err     error
}

func (e *ValidationError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("validation error on %q: %s: %v", e.Field, e.Message, e.Err)
	}
	return fmt.Sprintf("validation error on %q: %s", e.Field, e.Message)
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

// FormatErrors collects multiple validation errors
type FormatErrors struct {
	Errors []error
}

func (e *FormatErrors) Error() string {
	if len(e.Errors) == 0 {
		return "no errors"
	}
	msgs := make([]string, len(e.Errors))
	for i, err := range e.Errors {
		msgs[i] = err.Error()
	}
	return strings.Join(msgs, "; ")
}

// Helper constructor functions

func NewParseError(message string, err error) error {
	return &ParseError{Message: message, Err: err}
}

func NewActionError(action, message string, err error) error {
	return &ActionError{Action: action, Message: message, Err: err}
}

func NewUnknownActionError(action string, validActions []string) error {
	return &UnknownActionError{Action: action, ValidActions: validActions}
}

func NewValidationError(field, message string, err error) error {
	return &ValidationError{Field: field, Message: message, Err: err}
}

func NewFormatErrors(errs []error) error {
	return &FormatErrors{Errors: errs}
}

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

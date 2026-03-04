package framework

import (
	"errors"
	"fmt"
)

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

// ErrInvalidRequest represents invalid request or params (e.g. nil request)
type ErrInvalidRequest struct {
	RequestType string
	Reason      string
}

func (e *ErrInvalidRequest) Error() string {
	return fmt.Sprintf("invalid %s request: %s", e.RequestType, e.Reason)
}

// ErrTransport represents transport lifecycle errors
type ErrTransport struct {
	Reason string
}

func (e *ErrTransport) Error() string {
	return fmt.Sprintf("transport: %s", e.Reason)
}

// ErrContextCancelled represents context cancelled or nil
type ErrContextCancelled struct {
	Err error
}

func (e *ErrContextCancelled) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("context cancelled: %v", e.Err)
	}
	return "context cannot be nil"
}

// Unwrap returns the underlying error for errors.As support
func (e *ErrContextCancelled) Unwrap() error {
	return e.Err
}

// Helper functions for error checking (use errors.As for wrapped error support)

// IsToolNotFound checks if error is ErrToolNotFound
func IsToolNotFound(err error) bool {
	var target *ErrToolNotFound
	return errors.As(err, &target)
}

// IsPromptNotFound checks if error is ErrPromptNotFound
func IsPromptNotFound(err error) bool {
	var target *ErrPromptNotFound
	return errors.As(err, &target)
}

// IsResourceNotFound checks if error is ErrResourceNotFound
func IsResourceNotFound(err error) bool {
	var target *ErrResourceNotFound
	return errors.As(err, &target)
}

// IsInvalidTool checks if error is ErrInvalidTool
func IsInvalidTool(err error) bool {
	var target *ErrInvalidTool
	return errors.As(err, &target)
}

// IsInvalidPrompt checks if error is ErrInvalidPrompt
func IsInvalidPrompt(err error) bool {
	var target *ErrInvalidPrompt
	return errors.As(err, &target)
}

// IsInvalidResource checks if error is ErrInvalidResource
func IsInvalidResource(err error) bool {
	var target *ErrInvalidResource
	return errors.As(err, &target)
}

// IsInvalidRequest checks if error is ErrInvalidRequest
func IsInvalidRequest(err error) bool {
	var target *ErrInvalidRequest
	return errors.As(err, &target)
}

// IsTransport checks if error is ErrTransport
func IsTransport(err error) bool {
	var target *ErrTransport
	return errors.As(err, &target)
}

// IsContextCancelled checks if error is ErrContextCancelled
func IsContextCancelled(err error) bool {
	var target *ErrContextCancelled
	return errors.As(err, &target)
}

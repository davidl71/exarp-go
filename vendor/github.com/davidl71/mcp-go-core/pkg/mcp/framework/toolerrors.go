package framework

import (
	"fmt"
	"strings"
)

// ToolError wraps an error with tool and action context.
type ToolError struct {
	Tool    string
	Action  string
	Message string
	Err     error
}

func (e *ToolError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s(%s): %s: %v", e.Tool, e.Action, e.Message, e.Err)
	}
	return fmt.Sprintf("%s(%s): %s", e.Tool, e.Action, e.Message)
}

func (e *ToolError) Unwrap() error { return e.Err }

// WrapToolError creates a standardized tool error.
func WrapToolError(tool, action, msg string, err error) error {
	return &ToolError{Tool: tool, Action: action, Message: msg, Err: err}
}

// ParseError creates a standardized parse error for tool arguments.
func ParseError(tool string, err error) error {
	return WrapToolError(tool, "parse", "failed to parse arguments", err)
}

// ActionError creates a standardized action error.
func ActionError(tool, action string, err error) error {
	return WrapToolError(tool, action, "action failed", err)
}

// UnknownActionError creates a standardized unknown action error.
func UnknownActionError(tool, action string) error {
	return WrapToolError(tool, action, "unknown action", nil)
}

// ValidationError creates a standardized validation error.
func ValidationError(tool, action, field, msg string) error {
	return WrapToolError(tool, action, fmt.Sprintf("validation: %s: %s", field, msg), nil)
}

// FormatErrors joins multiple errors into a single message.
func FormatErrors(errs []error) string {
	msgs := make([]string, 0, len(errs))
	for _, e := range errs {
		if e != nil {
			msgs = append(msgs, e.Error())
		}
	}
	return strings.Join(msgs, "; ")
}

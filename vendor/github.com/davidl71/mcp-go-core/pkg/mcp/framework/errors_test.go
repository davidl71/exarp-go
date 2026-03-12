package framework

import (
	"errors"
	"fmt"
	"testing"
)

func TestErrInvalidTool(t *testing.T) {
	err := &ErrInvalidTool{
		ToolName: "test_tool",
		Reason:   "name cannot be empty",
	}

	errorMsg := err.Error()
	expected := `invalid tool "test_tool": name cannot be empty`
	if errorMsg != expected {
		t.Errorf("ErrInvalidTool.Error() = %q, want %q", errorMsg, expected)
	}

	// Test error checking
	if !IsInvalidTool(err) {
		t.Error("IsInvalidTool() should return true for ErrInvalidTool")
	}
}

func TestErrToolNotFound(t *testing.T) {
	err := &ErrToolNotFound{
		ToolName: "missing_tool",
	}

	errorMsg := err.Error()
	expected := `tool "missing_tool" not found`
	if errorMsg != expected {
		t.Errorf("ErrToolNotFound.Error() = %q, want %q", errorMsg, expected)
	}

	// Test error checking
	if !IsToolNotFound(err) {
		t.Error("IsToolNotFound() should return true for ErrToolNotFound")
	}

	// Test with errors.Is (should work with Go 1.13+)
	var target *ErrToolNotFound
	if !errors.As(err, &target) {
		t.Error("errors.As() should work with ErrToolNotFound")
	}
}

func TestErrInvalidPrompt(t *testing.T) {
	err := &ErrInvalidPrompt{
		PromptName: "test_prompt",
		Reason:     "description cannot be empty",
	}

	errorMsg := err.Error()
	expected := `invalid prompt "test_prompt": description cannot be empty`
	if errorMsg != expected {
		t.Errorf("ErrInvalidPrompt.Error() = %q, want %q", errorMsg, expected)
	}
}

func TestErrPromptNotFound(t *testing.T) {
	err := &ErrPromptNotFound{
		PromptName: "missing_prompt",
	}

	errorMsg := err.Error()
	expected := `prompt "missing_prompt" not found`
	if errorMsg != expected {
		t.Errorf("ErrPromptNotFound.Error() = %q, want %q", errorMsg, expected)
	}

	// Test error checking
	if !IsPromptNotFound(err) {
		t.Error("IsPromptNotFound() should return true for ErrPromptNotFound")
	}
}

func TestErrInvalidResource(t *testing.T) {
	err := &ErrInvalidResource{
		URI:    "file:///test",
		Reason: "URI cannot be empty",
	}

	errorMsg := err.Error()
	expected := `invalid resource "file:///test": URI cannot be empty`
	if errorMsg != expected {
		t.Errorf("ErrInvalidResource.Error() = %q, want %q", errorMsg, expected)
	}
}

func TestErrResourceNotFound(t *testing.T) {
	err := &ErrResourceNotFound{
		URI: "file:///missing",
	}

	errorMsg := err.Error()
	expected := `resource "file:///missing" not found`
	if errorMsg != expected {
		t.Errorf("ErrResourceNotFound.Error() = %q, want %q", errorMsg, expected)
	}

	// Test error checking
	if !IsResourceNotFound(err) {
		t.Error("IsResourceNotFound() should return true for ErrResourceNotFound")
	}
}

func TestErrInvalidRequest(t *testing.T) {
	err := &ErrInvalidRequest{
		RequestType: "CallTool",
		Reason:      "request cannot be nil",
	}
	errorMsg := err.Error()
	expected := `invalid CallTool request: request cannot be nil`
	if errorMsg != expected {
		t.Errorf("ErrInvalidRequest.Error() = %q, want %q", errorMsg, expected)
	}
	if !IsInvalidRequest(err) {
		t.Error("IsInvalidRequest() should return true for ErrInvalidRequest")
	}
}

func TestErrTransport(t *testing.T) {
	err := &ErrTransport{Reason: "server is nil"}
	errorMsg := err.Error()
	expected := `transport: server is nil`
	if errorMsg != expected {
		t.Errorf("ErrTransport.Error() = %q, want %q", errorMsg, expected)
	}
	if !IsTransport(err) {
		t.Error("IsTransport() should return true for ErrTransport")
	}
}

func TestErrContextCancelled(t *testing.T) {
	t.Run("with nil Err", func(t *testing.T) {
		err := &ErrContextCancelled{Err: nil}
		errorMsg := err.Error()
		expected := "context cannot be nil"
		if errorMsg != expected {
			t.Errorf("ErrContextCancelled.Error() = %q, want %q", errorMsg, expected)
		}
		if !IsContextCancelled(err) {
			t.Error("IsContextCancelled() should return true for ErrContextCancelled")
		}
	})
	t.Run("with wrapped error", func(t *testing.T) {
		inner := errors.New("deadline exceeded")
		err := &ErrContextCancelled{Err: inner}
		if err.Unwrap() != inner {
			t.Error("Unwrap() should return the inner error")
		}
		var target *ErrContextCancelled
		if !errors.As(err, &target) {
			t.Error("errors.As() should work with ErrContextCancelled")
		}
	})
}

func TestErrorHelperFunctions(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		checkFn  func(error) bool
		expected bool
	}{
		{
			name:     "IsToolNotFound with ErrToolNotFound",
			err:      &ErrToolNotFound{ToolName: "test"},
			checkFn:  IsToolNotFound,
			expected: true,
		},
		{
			name:     "IsToolNotFound with other error",
			err:      &ErrInvalidTool{ToolName: "test", Reason: "test"},
			checkFn:  IsToolNotFound,
			expected: false,
		},
		{
			name:     "IsPromptNotFound with ErrPromptNotFound",
			err:      &ErrPromptNotFound{PromptName: "test"},
			checkFn:  IsPromptNotFound,
			expected: true,
		},
		{
			name:     "IsResourceNotFound with ErrResourceNotFound",
			err:      &ErrResourceNotFound{URI: "test"},
			checkFn:  IsResourceNotFound,
			expected: true,
		},
		{
			name:     "IsInvalidTool with ErrInvalidTool",
			err:      &ErrInvalidTool{ToolName: "test", Reason: "test"},
			checkFn:  IsInvalidTool,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.checkFn(tt.err)
			if result != tt.expected {
				t.Errorf("%s() = %v, want %v", tt.name, result, tt.expected)
			}
		})
	}
}

func TestIsHelpersWithWrappedErrors(t *testing.T) {
	// When errors are wrapped with fmt.Errorf("%w", err), Is* should still detect them via errors.As
	wrappedToolNotFound := fmt.Errorf("operation failed: %w", &ErrToolNotFound{ToolName: "x"})
	if !IsToolNotFound(wrappedToolNotFound) {
		t.Error("IsToolNotFound() should return true for wrapped ErrToolNotFound")
	}
	wrappedInvalidTool := fmt.Errorf("validation: %w", &ErrInvalidTool{ToolName: "y", Reason: "bad"})
	if !IsInvalidTool(wrappedInvalidTool) {
		t.Error("IsInvalidTool() should return true for wrapped ErrInvalidTool")
	}
	wrappedErrTransport := fmt.Errorf("run failed: %w", &ErrTransport{Reason: "not started"})
	if !IsTransport(wrappedErrTransport) {
		t.Error("IsTransport() should return true for wrapped ErrTransport")
	}
}

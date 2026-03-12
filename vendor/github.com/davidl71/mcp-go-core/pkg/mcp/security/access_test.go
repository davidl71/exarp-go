package security

import (
	"testing"
)

func TestAccessControl_AllowDeny(t *testing.T) {
	ac := NewAccessControl(PermissionAllow)

	// Default should allow
	if err := ac.CheckTool("test-tool"); err != nil {
		t.Errorf("Default should allow: %v", err)
	}

	// Explicitly deny
	ac.DenyTool("test-tool")
	if err := ac.CheckTool("test-tool"); err == nil {
		t.Error("Denied tool should return error")
	}

	// Explicitly allow
	ac.AllowTool("test-tool")
	if err := ac.CheckTool("test-tool"); err != nil {
		t.Errorf("Allowed tool should not return error: %v", err)
	}
}

func TestAccessControl_DefaultDeny(t *testing.T) {
	ac := NewAccessControl(PermissionDeny)

	// Default should deny
	if err := ac.CheckTool("test-tool"); err == nil {
		t.Error("Default deny should return error")
	}

	// Explicitly allow
	ac.AllowTool("test-tool")
	if err := ac.CheckTool("test-tool"); err != nil {
		t.Errorf("Allowed tool should not return error: %v", err)
	}
}

func TestAccessControl_Resource(t *testing.T) {
	ac := NewAccessControl(PermissionAllow)

	// Default should allow
	if err := ac.CheckResource("stdio://test"); err != nil {
		t.Errorf("Default should allow: %v", err)
	}

	// Explicitly deny
	ac.DenyResource("stdio://test")
	if err := ac.CheckResource("stdio://test"); err == nil {
		t.Error("Denied resource should return error")
	}

	// Explicitly allow
	ac.AllowResource("stdio://test")
	if err := ac.CheckResource("stdio://test"); err != nil {
		t.Errorf("Allowed resource should not return error: %v", err)
	}
}

func TestAccessDeniedError(t *testing.T) {
	err := &AccessDeniedError{
		Resource: "tool",
		Name:     "test-tool",
	}

	if err.Error() == "" {
		t.Error("Error message should not be empty")
	}

	if err.Resource != "tool" {
		t.Errorf("Expected resource 'tool', got %s", err.Resource)
	}

	if err.Name != "test-tool" {
		t.Errorf("Expected name 'test-tool', got %s", err.Name)
	}
}

func TestDefaultAccessControl(t *testing.T) {
	ac1 := GetDefaultAccessControl()
	ac2 := GetDefaultAccessControl()

	// Should return same instance
	if ac1 != ac2 {
		t.Error("GetDefaultAccessControl should return same instance")
	}

	// Default should allow (permissive)
	if err := ac1.CheckTool("any-tool"); err != nil {
		t.Errorf("Default should allow: %v", err)
	}
}

func TestCheckToolAccess(t *testing.T) {
	// Reset default
	defaultAccessControl = NewAccessControl(PermissionAllow)

	// Should allow by default
	if err := CheckToolAccess("test-tool"); err != nil {
		t.Errorf("Should allow by default: %v", err)
	}

	// Deny and check
	defaultAccessControl.DenyTool("test-tool")
	if err := CheckToolAccess("test-tool"); err == nil {
		t.Error("Should deny after explicit denial")
	}
}

func TestCheckResourceAccess(t *testing.T) {
	// Reset default
	defaultAccessControl = NewAccessControl(PermissionAllow)

	// Should allow by default
	if err := CheckResourceAccess("stdio://test"); err != nil {
		t.Errorf("Should allow by default: %v", err)
	}

	// Deny and check
	defaultAccessControl.DenyResource("stdio://test")
	if err := CheckResourceAccess("stdio://test"); err == nil {
		t.Error("Should deny after explicit denial")
	}
}

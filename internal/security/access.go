// access.go — Access control, permission checking, and rate limiting.
package security

import (
	"fmt"
	"sync"

	"github.com/davidl71/exarp-go/internal/config"
)

// Permission represents access permission levels.
type Permission int

const (
	// PermissionDeny denies access.
	PermissionDeny Permission = iota
	// PermissionAllow allows access.
	PermissionAllow
	// PermissionDefault uses default policy.
	PermissionDefault
)

// AccessControl manages tool and resource access permissions.
type AccessControl struct {
	mu            sync.RWMutex
	toolPerms     map[string]Permission // tool name -> permission
	resourcePerms map[string]Permission // resource URI -> permission
	defaultPolicy Permission            // default permission if not specified
	allowedTools  map[string]bool       // explicit allow list (if default is deny)
	deniedTools   map[string]bool       // explicit deny list (if default is allow)
}

// NewAccessControl creates a new access control manager
// defaultPolicy: default permission (Allow or Deny).
func NewAccessControl(defaultPolicy Permission) *AccessControl {
	return &AccessControl{
		toolPerms:     make(map[string]Permission),
		resourcePerms: make(map[string]Permission),
		defaultPolicy: defaultPolicy,
		allowedTools:  make(map[string]bool),
		deniedTools:   make(map[string]bool),
	}
}

// NewAccessControlFromConfig creates a new access control manager from centralized config.
func NewAccessControlFromConfig() *AccessControl {
	cfg := config.GetGlobalConfig()
	ac := NewAccessControl(PermissionAllow) // Default to allow

	// If access control is enabled, configure it
	if cfg.Security.AccessControl.Enabled {
		// Set default policy
		if cfg.Security.AccessControl.DefaultPolicy == "deny" {
			ac.defaultPolicy = PermissionDeny
		} else {
			ac.defaultPolicy = PermissionAllow
		}

		// Add restricted tools to deny list
		for _, toolName := range cfg.Security.AccessControl.RestrictedTools {
			ac.DenyTool(toolName)
		}
	}

	return ac
}

// AllowTool explicitly allows access to a tool.
func (ac *AccessControl) AllowTool(toolName string) {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	ac.toolPerms[toolName] = PermissionAllow
	ac.allowedTools[toolName] = true
	delete(ac.deniedTools, toolName)
}

// DenyTool explicitly denies access to a tool.
func (ac *AccessControl) DenyTool(toolName string) {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	ac.toolPerms[toolName] = PermissionDeny
	ac.deniedTools[toolName] = true
	delete(ac.allowedTools, toolName)
}

// AllowResource explicitly allows access to a resource.
func (ac *AccessControl) AllowResource(uri string) {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	ac.resourcePerms[uri] = PermissionAllow
}

// DenyResource explicitly denies access to a resource.
func (ac *AccessControl) DenyResource(uri string) {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	ac.resourcePerms[uri] = PermissionDeny
}

// CheckTool checks if a tool can be accessed.
func (ac *AccessControl) CheckTool(toolName string) error {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	// Check explicit permission
	if perm, exists := ac.toolPerms[toolName]; exists {
		if perm == PermissionDeny {
			return &AccessDeniedError{
				Resource: "tool",
				Name:     toolName,
			}
		}

		if perm == PermissionAllow {
			return nil
		}
	}

	// Check deny list
	if ac.deniedTools[toolName] {
		return &AccessDeniedError{
			Resource: "tool",
			Name:     toolName,
		}
	}

	// Check allow list (if default is deny)
	if ac.defaultPolicy == PermissionDeny {
		if !ac.allowedTools[toolName] {
			return &AccessDeniedError{
				Resource: "tool",
				Name:     toolName,
			}
		}
	}

	// Use default policy
	if ac.defaultPolicy == PermissionDeny {
		return &AccessDeniedError{
			Resource: "tool",
			Name:     toolName,
		}
	}

	return nil
}

// CheckResource checks if a resource can be accessed.
func (ac *AccessControl) CheckResource(uri string) error {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	// Check explicit permission
	if perm, exists := ac.resourcePerms[uri]; exists {
		if perm == PermissionDeny {
			return &AccessDeniedError{
				Resource: "resource",
				Name:     uri,
			}
		}

		if perm == PermissionAllow {
			return nil
		}
	}

	// Use default policy
	if ac.defaultPolicy == PermissionDeny {
		return &AccessDeniedError{
			Resource: "resource",
			Name:     uri,
		}
	}

	return nil
}

// AccessDeniedError represents an access denied error.
type AccessDeniedError struct {
	Resource string
	Name     string
}

func (e *AccessDeniedError) Error() string {
	return fmt.Sprintf("access denied to %s: %s", e.Resource, e.Name)
}

// DefaultAccessControl is the default access control instance
// Default: Allow all (permissive for local development).
var (
	defaultAccessControl *AccessControl
	accessOnce           sync.Once
)

// GetDefaultAccessControl returns the default access control manager
// Uses centralized config if available, otherwise defaults to Allow (permissive for local MCP server).
func GetDefaultAccessControl() *AccessControl {
	accessOnce.Do(func() {
		defaultAccessControl = NewAccessControlFromConfig()
	})

	return defaultAccessControl
}

// CheckToolAccess checks tool access using the default access control.
func CheckToolAccess(toolName string) error {
	return GetDefaultAccessControl().CheckTool(toolName)
}

// CheckResourceAccess checks resource access using the default access control.
func CheckResourceAccess(uri string) error {
	return GetDefaultAccessControl().CheckResource(uri)
}

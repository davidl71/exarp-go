//go:build !(darwin && arm64 && cgo)
// +build !darwin !arm64 !cgo

package tools

import (
	"github.com/davidl71/exarp-go/internal/framework"
)

// registerAppleFoundationModelsTool is a no-op on unsupported platforms
// This allows the code to compile on all platforms without the Swift bridge
func registerAppleFoundationModelsTool(server framework.MCPServer) error {
	// Tool not available on this platform - silently skip registration
	return nil
}


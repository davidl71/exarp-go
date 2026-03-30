// Package utils provides a small set of helpers used across exarp-go packages.
package utils

import (
	"time"

	corefilelock "github.com/davidl71/mcp-go-core/pkg/mcp/filelock"
)

// FileLock is the repo’s file lock type (re-exported from mcp-go-core).
type FileLock = corefilelock.FileLock

// NewFileLock creates a new file lock (delegates to mcp-go-core).
func NewFileLock(path string, timeout time.Duration) (*FileLock, error) {
	return corefilelock.New(path, timeout)
}

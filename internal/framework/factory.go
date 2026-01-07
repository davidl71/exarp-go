package framework

// Factory functions moved to internal/factory to avoid import cycles
// Import internal/factory to use NewServer and NewServerFromConfig

// NewStdioTransport creates a stdio transport
func NewStdioTransport() Transport {
	// This will be implemented by each adapter
	// For now, return a placeholder
	return &stdioTransport{}
}

type stdioTransport struct{}


package gosdk

// AdapterOption configures a GoSDKAdapter
type AdapterOption func(*GoSDKAdapter)

// WithLogger sets a custom logger for the adapter
// This is a placeholder for future logger integration
func WithLogger(logger interface{}) AdapterOption {
	return func(a *GoSDKAdapter) {
		// TODO: Implement logger integration
		// This would allow users to provide their own logger implementation
	}
}

// WithMiddleware adds middleware to the adapter
// This is a placeholder for future middleware support
func WithMiddleware(middleware interface{}) AdapterOption {
	return func(a *GoSDKAdapter) {
		// TODO: Implement middleware support
		// This would allow users to add request/response middleware
	}
}

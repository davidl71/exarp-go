// apple_foundation_helpers.go — FM-related param extraction without the Foundation Models API (all platforms).
package tools

// getTemperature extracts temperature from params with default.
func getTemperature(params map[string]interface{}) float32 {
	return float32(ParamFloat64(params, "temperature", 0.7))
}

// getMaxTokens extracts max_tokens from params with default.
func getMaxTokens(params map[string]interface{}) int {
	return ParamInt(params, "max_tokens", 512)
}

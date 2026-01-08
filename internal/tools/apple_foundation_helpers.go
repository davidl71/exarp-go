package tools

// Helper functions for Apple Foundation Models that don't require the Foundation Models API
// These can be used on all platforms for parameter extraction

// getTemperature extracts temperature from params with default
func getTemperature(params map[string]interface{}) float32 {
	if temp, ok := params["temperature"].(float64); ok {
		return float32(temp)
	}
	return 0.7 // default
}

// getMaxTokens extracts max_tokens from params with default
func getMaxTokens(params map[string]interface{}) int {
	// Handle both int and float64 (JSON unmarshaling can produce either)
	if maxTokens, ok := params["max_tokens"].(int); ok {
		return maxTokens
	}
	if maxTokens, ok := params["max_tokens"].(float64); ok {
		return int(maxTokens)
	}
	return 512 // default
}


// protobuf_helpers.go — Shared helpers for config protobuf conversion (duration, JSON, nil-safe ToProto).
package config

import (
	"encoding/json"
	"fmt"
	"time"
)

// durationToSeconds converts Go time.Duration to protobuf int64 (seconds).
func durationToSeconds(d time.Duration) int64 {
	return int64(d.Seconds())
}

// secondsToDuration converts protobuf int64 (seconds) to Go time.Duration.
func secondsToDuration(seconds int64) time.Duration {
	return time.Duration(seconds) * time.Second
}

// mapToJSON converts a map to JSON string.
func mapToJSON(m interface{}) (string, error) {
	if m == nil {
		return "", nil
	}
	data, err := json.Marshal(m)
	if err != nil {
		return "", fmt.Errorf("failed to marshal map to JSON: %w", err)
	}
	return string(data), nil
}

// jsonToMap converts JSON string to a map.
func jsonToMap(jsonStr string, target interface{}) error {
	if jsonStr == "" {
		return nil
	}
	if err := json.Unmarshal([]byte(jsonStr), target); err != nil {
		return fmt.Errorf("failed to unmarshal JSON to map: %w", err)
	}
	return nil
}

// ptrToProto returns f(t) if t != nil, otherwise the zero value of P (e.g. nil for pointer types).
// Deduplicates the repeated "if t == nil { return nil }; return &pb{...}" pattern in ToProtobuf conversions.
func ptrToProto[T any, P any](t *T, f func(*T) P) P {
	if t == nil {
		var zero P
		return zero
	}
	return f(t)
}

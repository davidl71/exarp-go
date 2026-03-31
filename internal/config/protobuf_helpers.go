// protobuf_helpers.go — Shared helpers for config protobuf conversion (duration, JSON, nil-safe ToProto).
package config

import (
	"encoding/json"
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/durationpb"
)

func durationToProto(d time.Duration) *durationpb.Duration {
	if d <= 0 {
		return nil
	}
	return durationpb.New(d)
}

func durationFromProto(pb *durationpb.Duration) time.Duration {
	if pb == nil {
		return 0
	}
	return pb.AsDuration()
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

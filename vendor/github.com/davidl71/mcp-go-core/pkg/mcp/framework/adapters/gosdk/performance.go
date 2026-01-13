package gosdk

import (
	"context"
	"sync"
)

// performanceTracker tracks performance metrics for adapter operations
type performanceTracker struct {
	mu                sync.RWMutex
	toolCallCounts    map[string]int64
	toolCallDurations map[string]int64 // Total microseconds
	toolCallErrors    map[string]int64
}

func newPerformanceTracker() *performanceTracker {
	return &performanceTracker{
		toolCallCounts:    make(map[string]int64),
		toolCallDurations: make(map[string]int64),
		toolCallErrors:    make(map[string]int64),
	}
}

func (pt *performanceTracker) recordToolCall(name string, durationMicroseconds int64, hadError bool) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	pt.toolCallCounts[name]++
	pt.toolCallDurations[name] += durationMicroseconds
	if hadError {
		pt.toolCallErrors[name]++
	}
}

func (pt *performanceTracker) getStats(name string) (count int64, avgDurationMicroseconds int64, errorCount int64) {
	pt.mu.RLock()
	defer pt.mu.RUnlock()

	count = pt.toolCallCounts[name]
	if count > 0 {
		avgDurationMicroseconds = pt.toolCallDurations[name] / count
	}
	errorCount = pt.toolCallErrors[name]
	return
}

// validateContext checks if context is valid (optimized version)
func validateContextFast(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

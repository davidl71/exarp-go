// process_memory.go — Process memory usage and optional soft limit for the server/runtime.
// Uses runtime.ReadMemStats and optionally RSS (Linux /proc). Limit via EXARP_MEMORY_LIMIT_MB.
package tools

import (
	"os"
	"runtime"
	"strconv"
	"strings"
)

// ProcessMemoryInfo holds current process memory metrics and optional limit/warning.
type ProcessMemoryInfo struct {
	HeapAllocMB float64 `json:"heap_alloc_mb"` // Go heap allocated (in use)
	HeapSysMB   float64 `json:"heap_sys_mb"`   // Go heap memory from OS
	RSSMB       float64 `json:"rss_mb"`        // Resident set size (Linux); 0 if unavailable
	LimitMB     int     `json:"limit_mb"`      // Soft limit from EXARP_MEMORY_LIMIT_MB; 0 = no limit
	Warning     string  `json:"warning,omitempty"`
}

const (
	envMemoryLimitMB = "EXARP_MEMORY_LIMIT_MB"
	nearLimitPct     = 90 // warn when usage >= 90% of limit
)

// GetProcessMemoryInfo returns current process memory usage and optional limit/warning.
// Reads EXARP_MEMORY_LIMIT_MB (positive integer = soft limit in MB). If set and usage
// (RSS when available, else HeapSys) is at or above limit, or above nearLimitPct% of limit,
// Warning is set so agents can check/respect memory.
func GetProcessMemoryInfo() ProcessMemoryInfo {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	heapAllocMB := float64(m.HeapAlloc) / (1024 * 1024)
	heapSysMB := float64(m.HeapSys) / (1024 * 1024)

	rssMB := getRSSMB()

	limitMB := 0
	if s := strings.TrimSpace(os.Getenv(envMemoryLimitMB)); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			limitMB = n
		}
	}

	usageMB := heapSysMB
	if rssMB > 0 {
		usageMB = rssMB
	}

	var warning string
	if limitMB > 0 {
		pct := 0.0
		if limitMB > 0 {
			pct = (usageMB / float64(limitMB)) * 100
		}
		if usageMB >= float64(limitMB) {
			warning = "process memory at or over soft limit"
		} else if pct >= nearLimitPct {
			warning = "process memory near soft limit"
		}
	}

	return ProcessMemoryInfo{
		HeapAllocMB: roundMB(heapAllocMB),
		HeapSysMB:   roundMB(heapSysMB),
		RSSMB:       roundMB(rssMB),
		LimitMB:     limitMB,
		Warning:     warning,
	}
}

func roundMB(x float64) float64 {
	const prec = 2
	return float64(int(x*prec*10+0.5)) / (prec * 10)
}

// getRSSMB returns process RSS in MB on Linux (/proc/self/statm); 0 otherwise.
func getRSSMB() float64 {
	// Linux: /proc/self/statm second field is RSS in pages
	if runtime.GOOS != "linux" {
		return 0
	}
	data, err := os.ReadFile("/proc/self/statm")
	if err != nil {
		return 0
	}
	fields := strings.Fields(string(data))
	if len(fields) < 2 {
		return 0
	}
	rssPages, err := strconv.ParseUint(fields[1], 10, 64)
	if err != nil {
		return 0
	}
	pageSize := int64(4096)
	if p := os.Getpagesize(); p > 0 {
		pageSize = int64(p)
	}
	rssBytes := rssPages * uint64(pageSize)
	return float64(rssBytes) / (1024 * 1024)
}

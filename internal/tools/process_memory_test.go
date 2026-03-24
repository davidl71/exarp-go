// process_memory_test.go — Tests for process memory monitoring and limit.
package tools

import (
	"os"
	"testing"
)

func TestGetProcessMemoryInfo(t *testing.T) {
	info := GetProcessMemoryInfo()

	if info.HeapAllocMB < 0 {
		t.Errorf("HeapAllocMB = %v, want >= 0", info.HeapAllocMB)
	}
	if info.HeapSysMB < 0 {
		t.Errorf("HeapSysMB = %v, want >= 0", info.HeapSysMB)
	}
	if info.HeapAllocMB > 0 && info.HeapSysMB > 0 {
		t.Logf("process memory: heap_alloc_mb=%.2f heap_sys_mb=%.2f rss_mb=%.2f limit_mb=%d",
			info.HeapAllocMB, info.HeapSysMB, info.RSSMB, info.LimitMB)
	}
}

func TestGetProcessMemoryInfoWithLimitEnv(t *testing.T) {
	// Set a high limit so we don't actually trigger warning
	const limit = "99999"
	os.Setenv(envMemoryLimitMB, limit)
	defer os.Unsetenv(envMemoryLimitMB)

	info := GetProcessMemoryInfo()
	if info.LimitMB != 99999 {
		t.Errorf("LimitMB = %d, want 99999 (from EXARP_MEMORY_LIMIT_MB)", info.LimitMB)
	}
}

func TestProcessMemoryInfoContainsExpectedFields(t *testing.T) {
	info := GetProcessMemoryInfo()
	if info.HeapAllocMB < 0 {
		t.Errorf("HeapAllocMB = %v, want >= 0", info.HeapAllocMB)
	}
	if info.HeapSysMB < 0 {
		t.Errorf("HeapSysMB = %v, want >= 0", info.HeapSysMB)
	}
	if info.RSSMB < 0 {
		t.Errorf("RSSMB = %v, want >= 0", info.RSSMB)
	}
}

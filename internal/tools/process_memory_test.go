// process_memory_test.go — Tests for process memory monitoring and limit.
package tools

import (
	"encoding/json"
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

func TestProcessMemoryExposedInLlamaCppStatus(t *testing.T) {
	result, err := handleLlamaCppStatus()
	if err != nil {
		t.Fatalf("handleLlamaCppStatus() error = %v", err)
	}
	if len(result) == 0 || result[0].Text == "" {
		t.Fatal("handleLlamaCppStatus() returned empty result")
	}

	var status map[string]interface{}
	if err := json.Unmarshal([]byte(result[0].Text), &status); err != nil {
		t.Fatalf("unmarshal status: %v", err)
	}
	pm, ok := status["process_memory"].(map[string]interface{})
	if !ok {
		t.Fatal("status missing process_memory")
	}
	if _, ok := pm["memory_mb"]; !ok {
		t.Error("process_memory missing memory_mb")
	}
	if _, ok := pm["heap_alloc_mb"]; !ok {
		t.Error("process_memory missing heap_alloc_mb")
	}
	t.Logf("llamacpp status process_memory: %+v", pm)
}

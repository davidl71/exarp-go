package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFileCache_ReadFile(t *testing.T) {
	// Create temporary directory and file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := []byte("test content")

	err := os.WriteFile(testFile, testContent, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cache := NewFileCache()

	// First read - should be cache miss
	content, hit, err := cache.ReadFile(testFile)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if hit {
		t.Error("Expected cache miss on first read")
	}
	if string(content) != string(testContent) {
		t.Errorf("Content mismatch: got %q, want %q", string(content), string(testContent))
	}

	// Second read - should be cache hit
	content2, hit2, err := cache.ReadFile(testFile)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if !hit2 {
		t.Error("Expected cache hit on second read")
	}
	if string(content2) != string(testContent) {
		t.Errorf("Content mismatch: got %q, want %q", string(content2), string(testContent))
	}
}

func TestFileCache_InvalidationOnModification(t *testing.T) {
	// Create temporary directory and file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := []byte("original content")

	err := os.WriteFile(testFile, testContent, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cache := NewFileCache()

	// First read - cache miss
	_, hit, _ := cache.ReadFile(testFile)
	if hit {
		t.Error("Expected cache miss on first read")
	}

	// Second read - cache hit
	_, hit2, _ := cache.ReadFile(testFile)
	if !hit2 {
		t.Error("Expected cache hit on second read")
	}

	// Modify file
	newContent := []byte("modified content")
	err = os.WriteFile(testFile, newContent, 0644)
	if err != nil {
		t.Fatalf("Failed to modify test file: %v", err)
	}

	// Small delay to ensure mtime changes
	time.Sleep(10 * time.Millisecond)

	// Third read - should be cache miss (file modified)
	content3, hit3, err := cache.ReadFile(testFile)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if hit3 {
		t.Error("Expected cache miss after file modification")
	}
	if string(content3) != string(newContent) {
		t.Errorf("Content mismatch: got %q, want %q", string(content3), string(newContent))
	}
}

func TestFileCache_ReadFileWithTTL(t *testing.T) {
	// Create temporary directory and file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := []byte("test content")

	err := os.WriteFile(testFile, testContent, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cache := NewFileCache()
	ttl := 100 * time.Millisecond

	// First read - cache miss
	_, hit, _ := cache.ReadFileWithTTL(testFile, ttl)
	if hit {
		t.Error("Expected cache miss on first read")
	}

	// Second read - cache hit (within TTL)
	_, hit2, _ := cache.ReadFileWithTTL(testFile, ttl)
	if !hit2 {
		t.Error("Expected cache hit on second read")
	}

	// Wait for TTL to expire
	time.Sleep(ttl + 10*time.Millisecond)

	// Third read - cache miss (TTL expired)
	_, hit3, _ := cache.ReadFileWithTTL(testFile, ttl)
	if hit3 {
		t.Error("Expected cache miss after TTL expiration")
	}
}

func TestFileCache_Invalidate(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := []byte("test content")

	err := os.WriteFile(testFile, testContent, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cache := NewFileCache()

	// Read file to populate cache
	_, _, _ = cache.ReadFile(testFile)

	// Invalidate cache
	cache.Invalidate(testFile)

	// Read again - should be cache miss
	_, hit, _ := cache.ReadFile(testFile)
	if hit {
		t.Error("Expected cache miss after invalidation")
	}
}

func TestFileCache_Clear(t *testing.T) {
	tmpDir := t.TempDir()
	testFile1 := filepath.Join(tmpDir, "test1.txt")
	testFile2 := filepath.Join(tmpDir, "test2.txt")

	os.WriteFile(testFile1, []byte("content1"), 0644)
	os.WriteFile(testFile2, []byte("content2"), 0644)

	cache := NewFileCache()

	// Read both files to populate cache
	_, _, _ = cache.ReadFile(testFile1)
	_, _, _ = cache.ReadFile(testFile2)

	// Clear cache
	cache.Clear()

	// Read again - should be cache miss
	_, hit1, _ := cache.ReadFile(testFile1)
	_, hit2, _ := cache.ReadFile(testFile2)

	if hit1 || hit2 {
		t.Error("Expected cache miss after clear")
	}
}

func TestFileCache_GetStats(t *testing.T) {
	tmpDir := t.TempDir()
	testFile1 := filepath.Join(tmpDir, "test1.txt")
	testFile2 := filepath.Join(tmpDir, "test2.txt")

	content1 := []byte("content1")
	content2 := []byte("content2")

	os.WriteFile(testFile1, content1, 0644)
	os.WriteFile(testFile2, content2, 0644)

	cache := NewFileCache()

	// Read files to populate cache
	_, _, _ = cache.ReadFile(testFile1)
	_, _, _ = cache.ReadFile(testFile2)

	stats := cache.GetStats()

	if stats["entries"].(int) != 2 {
		t.Errorf("Expected 2 entries, got %d", stats["entries"])
	}

	expectedSize := int64(len(content1) + len(content2))
	if stats["total_size_bytes"].(int64) != expectedSize {
		t.Errorf("Expected total size %d, got %d", expectedSize, stats["total_size_bytes"])
	}
}

func TestGetGlobalFileCache(t *testing.T) {
	cache1 := GetGlobalFileCache()
	cache2 := GetGlobalFileCache()

	if cache1 != cache2 {
		t.Error("GetGlobalFileCache should return the same instance")
	}
}

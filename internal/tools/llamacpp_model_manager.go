// llamacpp_model_manager.go — GGUF model lifecycle: loading, LRU eviction, reference counting.
// No build tags — defines pure-Go types and management logic used by llamacpp_provider.go (CGO).
// Actual llama.cpp C bindings live in llamacpp_provider.go behind the "llamacpp" build tag.
// exarp-tags: #feature #llamacpp
package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// ModelManager manages loaded GGUF models with LRU eviction and reference counting.
// Thread-safe for concurrent AcquireModel/ReleaseModel from multiple goroutines.
type ModelManager struct {
	mu          sync.RWMutex
	models      map[string]*LoadedModel
	accessOrder []string
	maxModels   int
	totalMemory int64
	maxMemory   int64
	config      ModelManagerConfig
}

// LoadedModel tracks a single GGUF model that has been loaded (or registered for loading).
type LoadedModel struct {
	Path       string
	Name       string
	RefCount   int32
	LoadedAt   time.Time
	LastAccess time.Time
	MemoryUsed int64
}

// ModelManagerConfig controls model cache sizing, preloading, and inference defaults.
type ModelManagerConfig struct {
	MaxLoadedModels int      `json:"max_loaded_models"`
	MaxMemoryBytes  int64    `json:"max_memory_bytes"`
	PreloadModels   []string `json:"preload_models"`
	GPULayers       int      `json:"gpu_layers"`
	ContextSize     int      `json:"context_size"`
}

// ModelManagerStats exposes cache metrics for monitoring and the tool status action.
type ModelManagerStats struct {
	LoadedModels int               `json:"loaded_models"`
	MaxModels    int               `json:"max_models"`
	TotalMemory  int64             `json:"total_memory_bytes"`
	MaxMemory    int64             `json:"max_memory_bytes"`
	Models       []LoadedModelStat `json:"models"`
}

// LoadedModelStat is a snapshot of one cached model for Stats().
type LoadedModelStat struct {
	Path       string    `json:"path"`
	Name       string    `json:"name"`
	RefCount   int32     `json:"ref_count"`
	LoadedAt   time.Time `json:"loaded_at"`
	LastAccess time.Time `json:"last_access"`
	MemoryUsed int64     `json:"memory_used_bytes"`
}

// DefaultModelManagerConfig returns sensible defaults for a workstation with 16–32 GB RAM.
func DefaultModelManagerConfig() ModelManagerConfig {
	return ModelManagerConfig{
		MaxLoadedModels: 3,
		MaxMemoryBytes:  0, // unlimited
		GPULayers:       -1,
		ContextSize:     2048,
	}
}

// NewModelManager creates a ModelManager with the given config.
func NewModelManager(config ModelManagerConfig) *ModelManager {
	if config.MaxLoadedModels <= 0 {
		config.MaxLoadedModels = 3
	}
	if config.ContextSize <= 0 {
		config.ContextSize = 2048
	}
	return &ModelManager{
		models:    make(map[string]*LoadedModel),
		maxModels: config.MaxLoadedModels,
		maxMemory: config.MaxMemoryBytes,
		config:    config,
	}
}

// ResolveModelPath resolves a model name or relative path to an absolute .gguf file path.
// Search order:
//  1. Absolute path — use as-is if it exists.
//  2. $LLAMACPP_MODEL_DIR/<name> — user-configured model directory.
//  3. ~/.cache/exarp/models/<name> — XDG-style default cache.
func ResolveModelPath(nameOrPath string) (string, error) {
	if filepath.IsAbs(nameOrPath) {
		if _, err := os.Stat(nameOrPath); err == nil {
			return nameOrPath, nil
		}
		return "", fmt.Errorf("model not found at absolute path: %s", nameOrPath)
	}

	dirs := []string{}

	if d := os.Getenv("LLAMACPP_MODEL_DIR"); d != "" {
		dirs = append(dirs, d)
	}

	home, err := os.UserHomeDir()
	if err == nil {
		dirs = append(dirs, filepath.Join(home, ".cache", "exarp", "models"))
	}

	for _, dir := range dirs {
		candidate := filepath.Join(dir, nameOrPath)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("model %q not found in search paths (set LLAMACPP_MODEL_DIR or place in ~/.cache/exarp/models/)", nameOrPath)
}

// AcquireModel returns a LoadedModel for the given path, loading it if necessary.
// Increments the reference count so the model is not evicted while in use.
// Callers MUST call ReleaseModel when done.
func (m *ModelManager) AcquireModel(path string) (*LoadedModel, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if lm, ok := m.models[path]; ok {
		atomic.AddInt32(&lm.RefCount, 1)
		lm.LastAccess = time.Now()
		m.touchLRU(path)
		return lm, nil
	}

	m.evictLRU()

	if len(m.models) >= m.maxModels {
		return nil, fmt.Errorf("model cache full (%d/%d) and no evictable models (all in use)", len(m.models), m.maxModels)
	}

	memEstimate := estimateModelMemory(path)
	if m.maxMemory > 0 && m.totalMemory+memEstimate > m.maxMemory {
		return nil, fmt.Errorf("loading %s would exceed memory limit (%d + %d > %d bytes)",
			filepath.Base(path), m.totalMemory, memEstimate, m.maxMemory)
	}

	now := time.Now()
	lm := &LoadedModel{
		Path:       path,
		Name:       filepath.Base(path),
		RefCount:   1,
		LoadedAt:   now,
		LastAccess: now,
		MemoryUsed: memEstimate,
	}

	m.models[path] = lm
	m.accessOrder = append(m.accessOrder, path)
	m.totalMemory += memEstimate

	return lm, nil
}

// ReleaseModel decrements the reference count for a model. When the count reaches
// zero the model becomes eligible for LRU eviction (but is NOT unloaded immediately).
func (m *ModelManager) ReleaseModel(path string) {
	m.mu.RLock()
	lm, ok := m.models[path]
	m.mu.RUnlock()

	if !ok {
		return
	}
	atomic.AddInt32(&lm.RefCount, -1)
}

// UnloadModel force-unloads a model. Returns an error if the model is still in use (refcount > 0).
func (m *ModelManager) UnloadModel(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	lm, ok := m.models[path]
	if !ok {
		return fmt.Errorf("model not loaded: %s", path)
	}

	if atomic.LoadInt32(&lm.RefCount) > 0 {
		return fmt.Errorf("cannot unload %s: still in use (refcount=%d)", filepath.Base(path), lm.RefCount)
	}

	m.removeLocked(path, lm)
	return nil
}

// Stats returns a point-in-time snapshot of the model cache.
func (m *ModelManager) Stats() ModelManagerStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := ModelManagerStats{
		LoadedModels: len(m.models),
		MaxModels:    m.maxModels,
		TotalMemory:  m.totalMemory,
		MaxMemory:    m.maxMemory,
	}

	for _, lm := range m.models {
		stats.Models = append(stats.Models, LoadedModelStat{
			Path:       lm.Path,
			Name:       lm.Name,
			RefCount:   atomic.LoadInt32(&lm.RefCount),
			LoadedAt:   lm.LoadedAt,
			LastAccess: lm.LastAccess,
			MemoryUsed: lm.MemoryUsed,
		})
	}

	sort.Slice(stats.Models, func(i, j int) bool {
		return stats.Models[i].LastAccess.After(stats.Models[j].LastAccess)
	})

	return stats
}

// Preload loads models listed in config.PreloadModels so first inference is fast.
func (m *ModelManager) Preload() []error {
	var errs []error
	for _, modelPath := range m.config.PreloadModels {
		resolved, err := ResolveModelPath(modelPath)
		if err != nil {
			errs = append(errs, fmt.Errorf("preload %s: %w", modelPath, err))
			continue
		}
		lm, err := m.AcquireModel(resolved)
		if err != nil {
			errs = append(errs, fmt.Errorf("preload %s: %w", modelPath, err))
			continue
		}
		// Release immediately — preload only warms the cache, doesn't hold a reference.
		m.ReleaseModel(lm.Path)
	}
	return errs
}

// evictLRU removes the least-recently-used model with refcount == 0.
// Must be called with m.mu held.
func (m *ModelManager) evictLRU() {
	if len(m.models) < m.maxModels && (m.maxMemory == 0 || m.totalMemory < m.maxMemory) {
		return
	}

	for _, path := range m.accessOrder {
		lm, ok := m.models[path]
		if !ok {
			continue
		}
		if atomic.LoadInt32(&lm.RefCount) == 0 {
			m.removeLocked(path, lm)
			return
		}
	}
}

// touchLRU moves path to the end of the access order (most recently used).
// Must be called with m.mu held.
func (m *ModelManager) touchLRU(path string) {
	for i, p := range m.accessOrder {
		if p == path {
			m.accessOrder = append(m.accessOrder[:i], m.accessOrder[i+1:]...)
			m.accessOrder = append(m.accessOrder, path)
			return
		}
	}
}

// removeLocked deletes a model from the cache. Must be called with m.mu held.
func (m *ModelManager) removeLocked(path string, lm *LoadedModel) {
	m.totalMemory -= lm.MemoryUsed
	delete(m.models, path)

	for i, p := range m.accessOrder {
		if p == path {
			m.accessOrder = append(m.accessOrder[:i], m.accessOrder[i+1:]...)
			break
		}
	}
}

// estimateModelMemory gives a rough memory estimate from the GGUF file size.
// Actual memory usage is ~1.1–1.3x file size depending on quantization and context.
func estimateModelMemory(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 512 * 1024 * 1024 // default 512 MB if we can't stat
	}
	return int64(float64(info.Size()) * 1.2)
}

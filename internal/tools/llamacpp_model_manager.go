//go:build llamacpp && cgo
// +build llamacpp,cgo

// llamacpp_model_manager.go — Thread-safe ModelManager with LRU eviction for GGUF models.
//
// Manages a pool of loaded llama.cpp models keyed by file path. When the pool
// exceeds maxModels, the least-recently-used model is unloaded. Reference counting
// prevents eviction of actively-generating models.
package tools

import (
	"container/list"
	"context"
	"fmt"
	"sync"
	"time"

	llama "github.com/go-skynet/go-llama.cpp"
)

// modelEntry holds a loaded model and its usage metadata.
type modelEntry struct {
	path     string
	model    *llama.LLama
	refs     int // active Generate calls holding this model
	lastUsed time.Time
	lruElem  *list.Element // pointer into mgr.lru
}

// ModelManager is a thread-safe pool of loaded GGUF models with LRU eviction.
type ModelManager struct {
	mu        sync.Mutex
	models    map[string]*modelEntry
	lru       *list.List // front = MRU, back = LRU; elements hold model paths
	maxModels int
	ctxSize   int
}

// NewModelManager creates a ModelManager that holds at most maxModels models.
func NewModelManager(maxModels, ctxSize int) *ModelManager {
	if maxModels <= 0 {
		maxModels = 3
	}
	if ctxSize <= 0 {
		ctxSize = 2048
	}
	return &ModelManager{
		models:    make(map[string]*modelEntry),
		lru:       list.New(),
		maxModels: maxModels,
		ctxSize:   ctxSize,
	}
}

// Acquire returns the loaded model for path, loading it if necessary. The caller
// must call Release(path) when done. Blocks callers if the model is being evicted.
func (m *ModelManager) Acquire(ctx context.Context, path string) (*llama.LLama, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if entry, ok := m.models[path]; ok {
		entry.refs++
		entry.lastUsed = time.Now()
		m.lru.MoveToFront(entry.lruElem)
		return entry.model, nil
	}

	// Need to load. Evict LRU if at capacity.
	if err := m.evictIfNeeded(); err != nil {
		return nil, err
	}

	gpu := DetectGPU()
	gpuLayers := 0
	if gpu.Available {
		gpuLayers = gpu.Layers
	}

	model, err := llama.New(
		path,
		llama.SetContext(m.ctxSize),
		llama.EnableF16Memory,
		llama.SetGPULayers(gpuLayers),
	)
	if err != nil {
		return nil, fmt.Errorf("load model %s: %w", path, err)
	}

	elem := m.lru.PushFront(path)
	entry := &modelEntry{
		path:     path,
		model:    model,
		refs:     1,
		lastUsed: time.Now(),
		lruElem:  elem,
	}
	m.models[path] = entry
	return model, nil
}

// Release decrements the reference count for the model at path.
func (m *ModelManager) Release(path string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if entry, ok := m.models[path]; ok && entry.refs > 0 {
		entry.refs--
	}
}

// Unload explicitly frees the model at path. No-op if not loaded or has active refs.
func (m *ModelManager) Unload(path string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	entry, ok := m.models[path]
	if !ok {
		return
	}
	if entry.refs > 0 {
		return // busy
	}
	m.evictEntry(entry)
}

// UnloadAll frees all loaded models regardless of ref counts.
func (m *ModelManager) UnloadAll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, entry := range m.models {
		entry.model.Free()
	}
	m.models = make(map[string]*modelEntry)
	m.lru.Init()
}

// Stats returns a snapshot of the manager's state.
func (m *ModelManager) Stats() map[string]interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()

	loaded := make([]map[string]interface{}, 0, len(m.models))
	for _, e := range m.models {
		loaded = append(loaded, map[string]interface{}{
			"path":      e.path,
			"refs":      e.refs,
			"last_used": e.lastUsed.Format(time.RFC3339),
		})
	}
	return map[string]interface{}{
		"loaded":     len(m.models),
		"max_models": m.maxModels,
		"models":     loaded,
	}
}

// evictIfNeeded removes the LRU unreferenced model if at capacity. Caller holds mu.
func (m *ModelManager) evictIfNeeded() error {
	if len(m.models) < m.maxModels {
		return nil
	}
	// Walk from back (LRU) looking for an unreferenced entry.
	for e := m.lru.Back(); e != nil; e = e.Prev() {
		path := e.Value.(string)
		entry := m.models[path]
		if entry.refs == 0 {
			m.evictEntry(entry)
			return nil
		}
	}
	return fmt.Errorf("all %d model slots are in use; try again later", m.maxModels)
}

// evictEntry frees and removes an entry. Caller holds mu.
func (m *ModelManager) evictEntry(entry *modelEntry) {
	entry.model.Free()
	m.lru.Remove(entry.lruElem)
	delete(m.models, entry.path)
}

// defaultModelManager is the package-level singleton.
var defaultModelManager = NewModelManager(3, 2048)

// DefaultModelManager returns the package-level ModelManager singleton.
func DefaultModelManager() *ModelManager {
	return defaultModelManager
}

// ModelManagerStats returns a snapshot of the default manager's state.
func ModelManagerStats() map[string]interface{} {
	return defaultModelManager.Stats()
}

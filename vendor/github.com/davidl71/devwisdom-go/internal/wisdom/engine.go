package wisdom

import (
	"fmt"
	"hash/fnv"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/davidl71/devwisdom-go/internal/config"
)

// Engine is the main wisdom engine managing sources, advisors, and consultations.
// It provides thread-safe access to wisdom sources and advisor consultations.
// The engine must be initialized before use by calling Initialize().
type Engine struct {
	sources     map[string]*Source
	loader      *SourceLoader
	advisors    *AdvisorRegistry
	config      *config.Config
	initialized bool
	mu          sync.RWMutex
	// Performance optimization: cached sorted source list
	sortedSources   []string
	sortedSourcesMu sync.RWMutex
	// Performance optimization: cached date hash for random source selection
	cachedDateHash int64
	cachedDate     string // YYYYMMDD format
	dateHashMu     sync.RWMutex
}

// NewEngine creates a new wisdom engine instance.
// The engine is not initialized by default; call Initialize() before use.
func NewEngine() *Engine {
	return &Engine{
		sources:  make(map[string]*Source),
		loader:   NewSourceLoader(),
		advisors: NewAdvisorRegistry(),
		config:   config.NewConfig(),
	}
}

// Initialize loads wisdom sources and configuration.
// This method is idempotent and can be called multiple times safely.
// It loads sources from configuration files or falls back to built-in sources.
func (e *Engine) Initialize() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.initialized {
		return nil
	}

	// Load configuration
	if err := e.config.Load(); err != nil {
		return fmt.Errorf("failed to load engine configuration: %w", err)
	}

	// Configure source loader
	e.loader = NewSourceLoader().
		WithConfigPaths(
			"sources.json",
			"wisdom/sources.json",
			".wisdom/sources.json",
		)

	// Try to load from default locations
	if err := e.loader.Load(); err != nil {
		// Fallback to minimal hard-coded source if config loading fails
		e.sources = make(map[string]*Source)
		e.sources["bofh"] = &Source{
			Name:   "BOFH (Bastard Operator From Hell)",
			Icon:   "😈",
			Quotes: make(map[string][]Quote),
		}
		e.sources["bofh"].Quotes["chaos"] = []Quote{
			{
				Quote:         "It's not a bug, it's a feature.",
				Source:        "BOFH Excuse Calendar",
				Encouragement: "Document it and ship it.",
			},
		}
	} else {
		// Use loaded sources
		e.sources = e.loader.GetAllSources()
	}

	// Pre-compute sorted source list for performance optimization
	e.updateSortedSources()

	// Initialize advisors
	e.advisors.Initialize()

	e.initialized = true
	return nil
}

// ReloadSources reloads sources from configuration files.
// This is useful when sources are updated externally and you want to refresh the engine.
func (e *Engine) ReloadSources() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.initialized {
		return fmt.Errorf("engine not initialized: call Initialize() before reloading sources")
	}

	if e.loader == nil {
		return fmt.Errorf("loader not initialized")
	}

	if err := e.loader.Reload(); err != nil {
		return fmt.Errorf("failed to reload sources: %w", err)
	}

	e.sources = e.loader.GetAllSources()
	// Update cached sorted source list
	e.updateSortedSources()
	return nil
}

// GetWisdom retrieves a wisdom quote based on score and source.
// The score determines the aeon level, which selects appropriate quotes from the source.
// If source is "random", a date-seeded random source is selected for consistency.
func (e *Engine) GetWisdom(score float64, source string) (*Quote, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if !e.initialized {
		return nil, fmt.Errorf("engine not initialized: call Initialize() before retrieving wisdom. This usually means sources.json could not be loaded")
	}

	// Handle "random" source selection
	if source == "random" {
		randomSource, err := e.getRandomSourceLocked(true)
		if err != nil {
			return nil, fmt.Errorf("failed to get random source (score: %.1f): %w", score, err)
		}
		source = randomSource
	}

	src, exists := e.sources[source]
	if !exists {
		// Get available sources for helpful error message
		availableSources := make([]string, 0, len(e.sources))
		for id := range e.sources {
			availableSources = append(availableSources, id)
		}
		if len(availableSources) > 0 {
			return nil, fmt.Errorf("unknown source %q (available sources: %v). Use 'devwisdom sources' to list all sources", source, availableSources)
		}
		return nil, fmt.Errorf("unknown source %q and no sources are available. Ensure sources.json is properly configured", source)
	}

	// Determine aeon level from score
	aeonLevel := GetAeonLevel(score)

	// Get quote from source based on aeon level
	quote := src.GetQuote(aeonLevel)
	return quote, nil
}

// GetRandomSource returns a random wisdom source ID.
// If seedDate is true, the same source will be returned for the entire day (date-seeded).
// This ensures consistent daily source selection across sessions.
func (e *Engine) GetRandomSource(seedDate bool) (string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.getRandomSourceLocked(seedDate)
}

// updateSortedSources updates the cached sorted source list.
// This should be called whenever sources are loaded or reloaded.
func (e *Engine) updateSortedSources() {
	e.sortedSourcesMu.Lock()
	defer e.sortedSourcesMu.Unlock()

	// Get all available source IDs
	allSources := make([]string, 0, len(e.sources))
	for id := range e.sources {
		allSources = append(allSources, id)
	}

	// Sort sources to ensure deterministic order (Go map iteration is non-deterministic)
	sort.Strings(allSources)

	e.sortedSources = allSources
}

// getDateHash computes and caches the date hash for random source selection.
// The hash is cached per day to avoid recomputation.
func (e *Engine) getDateHash() int64 {
	now := time.Now()
	dateStr := now.Format("20060102") // YYYYMMDD format

	// Check if we have a cached hash for today
	e.dateHashMu.RLock()
	if e.cachedDate == dateStr && e.cachedDateHash != 0 {
		hash := e.cachedDateHash
		e.dateHashMu.RUnlock()
		return hash
	}
	e.dateHashMu.RUnlock()

	// Compute hash (need write lock)
	e.dateHashMu.Lock()
	defer e.dateHashMu.Unlock()

	// Double-check after acquiring write lock (another goroutine might have computed it)
	if e.cachedDate == dateStr && e.cachedDateHash != 0 {
		return e.cachedDateHash
	}

	// Convert date string to int and add hash offset (matching Python implementation)
	var dateInt int64
	if _, err := fmt.Sscanf(dateStr, "%d", &dateInt); err != nil {
		// This should never fail since we just formatted the date, but handle it gracefully
		dateInt = int64(now.Unix())
	}

	// Hash "random_source" string for offset (matching Python hash("random_source"))
	h := fnv.New32a()
	h.Write([]byte("random_source"))
	hashOffset := int64(h.Sum32())

	seed := dateInt + hashOffset

	// Cache the result
	e.cachedDate = dateStr
	e.cachedDateHash = seed

	return seed
}

// getRandomSourceLocked is the internal implementation (assumes RLock is held)
func (e *Engine) getRandomSourceLocked(seedDate bool) (string, error) {
	if !e.initialized {
		return "", fmt.Errorf("engine not initialized: call Initialize() before use")
	}

	// Use cached sorted source list (performance optimization)
	e.sortedSourcesMu.RLock()
	allSources := e.sortedSources
	e.sortedSourcesMu.RUnlock()

	// If cache is empty or out of sync, update it
	if len(allSources) == 0 || len(allSources) != len(e.sources) {
		e.updateSortedSources()
		e.sortedSourcesMu.RLock()
		allSources = e.sortedSources
		e.sortedSourcesMu.RUnlock()
	}

	if len(allSources) == 0 {
		return "", fmt.Errorf("no sources available: ensure sources.json exists and contains valid source definitions")
	}

	// Date-seeded random selection for consistency
	var seed int64
	if seedDate {
		seed = e.getDateHash()
	} else {
		seed = time.Now().UnixNano()
	}

	// Create seeded random generator
	rng := rand.New(rand.NewSource(seed))

	// Select random source
	selectedIndex := rng.Intn(len(allSources))
	return allSources[selectedIndex], nil
}

// ListSources returns all available wisdom source IDs.
// Returns an empty slice if the engine is not initialized.
func (e *Engine) ListSources() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.loader != nil {
		return e.loader.ListSourceIDs()
	}

	sources := make([]string, 0, len(e.sources))
	for name := range e.sources {
		sources = append(sources, name)
	}
	return sources
}

// GetSource returns a specific source by ID.
// The second return value indicates whether the source was found.
func (e *Engine) GetSource(id string) (*Source, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.loader != nil {
		return e.loader.GetSource(id)
	}

	source, exists := e.sources[id]
	return source, exists
}

// GetLoader returns the source loader for advanced usage.
// This allows direct access to source loading and configuration management.
func (e *Engine) GetLoader() *SourceLoader {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.loader
}

// GetAdvisors returns the advisor registry.
// This provides access to advisor mappings and consultation functionality.
func (e *Engine) GetAdvisors() *AdvisorRegistry {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.advisors
}

// AddProjectSource adds a source and saves it to the project directory.
// The source is persisted to the project's sources.json file and immediately available.
func (e *Engine) AddProjectSource(config *SourceConfig) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.loader == nil {
		return fmt.Errorf("loader not initialized: engine must be initialized before adding project sources")
	}

	// Save to project directory
	if err := e.loader.SaveProjectSource(config); err != nil {
		return fmt.Errorf("failed to save project source %q to project directory: %w", config.ID, err)
	}

	// Add to loader (this will also update the loader's internal sources)
	if err := e.loader.AddSource(config); err != nil {
		return fmt.Errorf("failed to add source %q to loader: %w", config.ID, err)
	}

	// Reload to ensure consistency
	if err := e.loader.Reload(); err != nil {
		return fmt.Errorf("failed to reload sources after adding %q: %w", config.ID, err)
	}

	// Update engine's sources map from loader
	e.sources = e.loader.GetAllSources()

	return nil
}

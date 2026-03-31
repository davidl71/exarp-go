package tools

import (
	"context"
	"sync"
	"time"

	"github.com/davidl71/exarp-go/internal/config"
)

// ollamaTextGenerator implements TextGenerator using native Ollama generate (HTTP).
// Used in the FM chain (Apple → Ollama → stub) so DefaultFMProvider() can use Ollama
// when Apple FM is unavailable (e.g. Linux). Also used by text_generate (provider=ollama).
type ollamaTextGenerator struct{}

// DefaultOllamaGen is the shared Ollama TextGenerator for text_generate (provider=ollama).
var DefaultOllamaGen TextGenerator = &ollamaTextGenerator{}

// DefaultOllamaTextGenerator returns the Ollama TextGenerator (native HTTP, config-aware host/model).
func DefaultOllamaTextGenerator() TextGenerator {
	return DefaultOllamaGen
}

// ollamaFMProbeTTL is how long OllamaReachableForFM caches a successful or failed /api/tags probe.
const ollamaFMProbeTTL = 15 * time.Second

var (
	ollamaReachMu    sync.RWMutex
	ollamaReachHost  string
	ollamaReachUntil time.Time
	ollamaReachOK    bool
)

func resetOllamaReachabilityCacheForTest() {
	ollamaReachMu.Lock()
	defer ollamaReachMu.Unlock()
	ollamaReachHost = ""
	ollamaReachUntil = time.Time{}
}

// ollamaConfiguredHost returns the Ollama base URL from config (same source as Generate).
func ollamaConfiguredHost() string {
	host := "http://localhost:11434"
	if cfg := config.GetGlobalConfig(); cfg != nil && cfg.Tools.Ollama.DefaultHost != "" {
		host = cfg.Tools.Ollama.DefaultHost
	}
	return host
}

// OllamaReachableForFM reports whether the configured Ollama host recently returned HTTP 200
// for GET /api/tags. Results are cached for ollamaFMProbeTTL (host changes invalidate the cache entry).
// Use this and FMAvailable for discovery; ollamaTextGenerator.Supported remains true so chain Generate
// still attempts HTTP calls (handles races where the probe is stale).
func OllamaReachableForFM() bool {
	host := ollamaConfiguredHost()
	now := time.Now()

	ollamaReachMu.RLock()
	if host == ollamaReachHost && now.Before(ollamaReachUntil) {
		ok := ollamaReachOK
		ollamaReachMu.RUnlock()
		return ok
	}
	ollamaReachMu.RUnlock()

	ollamaReachMu.Lock()
	defer ollamaReachMu.Unlock()
	if host == ollamaReachHost && now.Before(ollamaReachUntil) {
		return ollamaReachOK
	}
	ctx, cancel := context.WithTimeout(context.Background(), ollamaFMProbeTimeout)
	defer cancel()
	ok := ollamaPingTagsAPI(ctx, host)
	ollamaReachHost = host
	ollamaReachOK = ok
	ollamaReachUntil = now.Add(ollamaFMProbeTTL)
	return ok
}

func (*ollamaTextGenerator) Supported() bool {
	// Always true so chain Generate still invokes Ollama; FMAvailable uses OllamaReachableForFM.
	return true
}

func (*ollamaTextGenerator) Generate(ctx context.Context, prompt string, maxTokens int, temperature float32) (string, error) {
	model := config.GetOllamaDefaultModel()
	if model == "" {
		model = "llama3.2"
	}

	return ollamaGenerateText(ctx, prompt, maxTokens, temperature, ollamaConfiguredHost(), model)
}

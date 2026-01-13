package tools

import (
	"sync"

	"github.com/davidl71/devwisdom-go/pkg/wisdom"
)

var (
	// wisdomEngine is a shared singleton instance of the wisdom engine
	wisdomEngine *wisdom.Engine
	wisdomOnce   sync.Once
)

// getWisdomEngine returns a shared singleton instance of the wisdom engine.
// The engine is initialized on first access.
func getWisdomEngine() (*wisdom.Engine, error) {
	wisdomOnce.Do(func() {
		wisdomEngine = wisdom.NewEngine()
	})

	if err := wisdomEngine.Initialize(); err != nil {
		return nil, err
	}

	return wisdomEngine, nil
}

package tools

import (
	"sync"
	"time"

	"github.com/davidl71/devwisdom-go/pkg/wisdom"
	"github.com/davidl71/exarp-go/proto"
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

// BuildBriefingDataProto builds *proto.BriefingData from the wisdom engine and score (for type-safe report briefing).
func BuildBriefingDataProto(engine *wisdom.Engine, score float64) *proto.BriefingData {
	if engine == nil {
		return nil
	}
	pb := &proto.BriefingData{
		Date:  time.Now().Format("2006-01-02"),
		Score: score,
	}
	sources := engine.ListSources()
	pb.Sources = sources
	selectedSources := []string{"pistis_sophia", "stoic", "tao"}
	if len(sources) > 0 {
		selectedSources = sources
		if len(selectedSources) > 3 {
			selectedSources = selectedSources[:3]
		}
	}
	for _, src := range selectedSources {
		quote, err := engine.GetWisdom(score, src)
		if err == nil && quote != nil {
			pb.Quotes = append(pb.Quotes, &proto.BriefingQuote{
				Quote:         quote.Quote,
				Source:        quote.Source,
				Encouragement: quote.Encouragement,
				WisdomSource:  quote.WisdomSource,
				WisdomIcon:    quote.WisdomIcon,
			})
		}
	}
	return pb
}

// Package wisdom provides wisdom quotes, sources, advisors, and consultation services.
// It includes functionality for retrieving quotes from various wisdom sources,
// managing advisor consultations, and handling project health-based guidance.
package wisdom

import (
	"fmt"
	"hash/fnv"
	"math/rand"
	"time"
)

// Quote represents a wisdom quote with metadata.
// It contains the quote text, source information, and optional encouragement message.
type Quote struct {
	Quote         string `json:"quote"`
	Source        string `json:"source"`
	Encouragement string `json:"encouragement"`
	WisdomSource  string `json:"wisdom_source,omitempty"`
	WisdomIcon    string `json:"wisdom_icon,omitempty"`
}

// Source represents a wisdom source with quotes organized by aeon level.
// Aeon levels correspond to project health scores (chaos, lower_aeons, middle_aeons, upper_aeons, treasury).
type Source struct {
	Name        string             `json:"name"`
	Icon        string             `json:"icon"`
	Quotes      map[string][]Quote `json:"quotes"` // Key: aeon level (chaos, lower_aeons, etc.)
	Description string             `json:"description,omitempty"`
}

// GetQuote retrieves a quote for the given aeon level.
// If no quotes exist for the specified level, it falls back to any available quotes.
// Returns a default quote if no quotes are available in the source.
func (s *Source) GetQuote(aeonLevel string) *Quote {
	quotes, exists := s.Quotes[aeonLevel]
	if !exists || len(quotes) == 0 {
		// Fallback to any available quotes
		for _, qs := range s.Quotes {
			if len(qs) > 0 {
				quotes = qs
				break
			}
		}
	}

	if len(quotes) == 0 {
		return &Quote{
			Quote:         "Silence is also wisdom.",
			Source:        "Unknown",
			Encouragement: "Sometimes reflection is the answer.",
		}
	}

	// If only one quote, return it directly
	if len(quotes) == 1 {
		return &quotes[0]
	}

	// Date-seeded random selection for daily consistency
	// Uses same pattern as getRandomSourceLocked() in engine.go
	now := time.Now()
	dateStr := now.Format("20060102") // YYYYMMDD format

	// Convert date string to int and add hash offset
	var dateInt int64
	if _, err := fmt.Sscanf(dateStr, "%d", &dateInt); err != nil {
		// This should never fail since we just formatted the date, but handle it gracefully
		dateInt = int64(now.Unix())
	}

	// Hash "random_quote" string for offset (different from "random_source" for source selection)
	h := fnv.New32a()
	h.Write([]byte("random_quote"))
	hashOffset := int64(h.Sum32())

	seed := dateInt + hashOffset

	// Create seeded random generator
	rng := rand.New(rand.NewSource(seed))

	// Select random quote index
	selectedIndex := rng.Intn(len(quotes))
	return &quotes[selectedIndex]
}

// Consultation represents an advisor consultation with full metadata.
// It includes advisor information, quote, rationale, score context, and mode guidance.
type Consultation struct {
	Timestamp        string  `json:"timestamp"`
	ConsultationType string  `json:"consultation_type"`
	Advisor          string  `json:"advisor"`
	AdvisorIcon      string  `json:"advisor_icon"`
	AdvisorName      string  `json:"advisor_name"`
	Rationale        string  `json:"rationale"`
	Metric           string  `json:"metric,omitempty"`
	Tool             string  `json:"tool,omitempty"`
	Stage            string  `json:"stage,omitempty"`
	ScoreAtTime      float64 `json:"score_at_time"`
	ConsultationMode string  `json:"consultation_mode"`
	ModeIcon         string  `json:"mode_icon"`
	ModeFrequency    string  `json:"mode_frequency"`
	Quote            string  `json:"quote"`
	QuoteSource      string  `json:"quote_source"`
	Encouragement    string  `json:"encouragement"`
	Context          string  `json:"context,omitempty"`
	SessionMode      string  `json:"session_mode,omitempty"`
	ModeGuidance     string  `json:"mode_guidance,omitempty"`
}

// AdvisorInfo represents advisor metadata including name, icon, rationale, and context.
type AdvisorInfo struct {
	Advisor   string `json:"advisor"`
	Icon      string `json:"icon"`
	Rationale string `json:"rationale"`
	HelpsWith string `json:"helps_with,omitempty"`
	Language  string `json:"language,omitempty"`
}

// AeonLevel represents project health stages based on score ranges.
// These levels determine which quotes are selected from wisdom sources.
type AeonLevel string

const (
	AeonChaos    AeonLevel = "chaos"        // < 30%
	AeonLower    AeonLevel = "lower_aeons"  // 30-50%
	AeonMiddle   AeonLevel = "middle_aeons" // 50-70%
	AeonUpper    AeonLevel = "upper_aeons"  // 70-85%
	AeonTreasury AeonLevel = "treasury"     // > 85%
)

// GetAeonLevel returns the aeon level based on score.
// Score ranges:
//   - < 30: chaos
//   - 30-50: lower_aeons
//   - 50-70: middle_aeons
//   - 70-85: upper_aeons
//   - >= 85: treasury
func GetAeonLevel(score float64) string {
	switch {
	case score < 30:
		return string(AeonChaos)
	case score < 50:
		return string(AeonLower)
	case score < 70:
		return string(AeonMiddle)
	case score < 85:
		return string(AeonUpper)
	default:
		return string(AeonTreasury)
	}
}

// ConsultationMode represents project health consultation modes based on score ranges.
// These modes determine consultation frequency and advisor selection.
type ConsultationMode string

const (
	ModeChaos    ConsultationMode = "chaos"    // < 30%
	ModeBuilding ConsultationMode = "building" // 30-60%
	ModeMaturing ConsultationMode = "maturing" // 60-80%
	ModeMastery  ConsultationMode = "mastery"  // > 80%
)

// ConsultationModeConfig represents configuration for a consultation mode.
// It defines score ranges, frequency, description, and icon for each mode.
type ConsultationModeConfig struct {
	Name        string  `json:"name"`
	MinScore    float64 `json:"min_score"`
	MaxScore    float64 `json:"max_score"`
	Frequency   string  `json:"frequency"`
	Description string  `json:"description"`
	Icon        string  `json:"icon"`
}

// SessionMode represents different session interaction modes.
// Used for mode-aware advisor selection (AGENT, ASK, MANUAL).
type SessionMode string

const (
	SessionModeAgent  SessionMode = "AGENT"
	SessionModeAsk    SessionMode = "ASK"
	SessionModeManual SessionMode = "MANUAL"
)

// ModeConfig represents configuration for session mode-aware advisor selection.
// It defines preferred advisors, tone, and focus for each session mode.
type ModeConfig struct {
	PreferredAdvisors []string `json:"preferred_advisors"`
	Tone              string   `json:"tone"`
	Focus             string   `json:"focus"`
}

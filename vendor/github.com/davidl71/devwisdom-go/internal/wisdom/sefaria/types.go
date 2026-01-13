// Package sefaria provides Sefaria API client for fetching Hebrew text sources.
package sefaria

// TextResponse represents a Sefaria API text response
type TextResponse struct {
	Ref      string    `json:"ref"`
	HeRef    string    `json:"heRef"`
	Text     []string  `json:"text"` // English verses
	He       []string  `json:"he"`   // Hebrew verses
	Versions []Version `json:"versions"`
	Metadata *Metadata `json:"-"`
}

// Version represents a translation version in the Sefaria API response
type Version struct {
	VersionTitle string   `json:"versionTitle"`
	Language     string   `json:"language"`
	Text         []string `json:"text"`
}

// Metadata represents metadata about the text (extracted from response)
type Metadata struct {
	Book         string   `json:"book"`
	HeTitle      string   `json:"heTitle"`
	Categories   []string `json:"categories"`
	IndexTitle   string   `json:"indexTitle"`
	HeIndexTitle string   `json:"heIndexTitle"`
}

// QuoteRequest represents a request for a specific quote from Sefaria
type QuoteRequest struct {
	Book    string // Sefaria book ID (e.g., "Pirkei_Avot", "Proverbs")
	Chapter int    // Chapter number (0 for full book)
	Verse   int    // Verse number (0 for full chapter)
}

// BookMapping maps our source IDs to Sefaria book IDs
var BookMapping = map[string]string{
	"pirkei_avot":  "Pirkei_Avot",
	"proverbs":     "Proverbs",
	"ecclesiastes": "Ecclesiastes",
	"psalms":       "Psalms",
}

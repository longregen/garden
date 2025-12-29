package entity

import "time"

// UnifiedSearchResult represents a unified search result from multiple tables
type UnifiedSearchResult struct {
	ItemType     string
	ItemID       string
	ItemTitle    string
	LastActivity time.Time
	SearchScore  float64
}

// SearchWeights defines the weights for different search scoring factors
type SearchWeights struct {
	ExactMatchWeight float64
	SimilarityWeight float64
	RecencyWeight    float64
}

// DefaultSearchWeights returns the default search weights
func DefaultSearchWeights() SearchWeights {
	return SearchWeights{
		ExactMatchWeight: 5.0,
		SimilarityWeight: 2.0,
		RecencyWeight:    1.0,
	}
}

// RetrievedItem represents a bookmark question/answer pair retrieved by similarity search
type RetrievedItem struct {
	ID            int     `json:"id"`
	Question      string  `json:"question"`
	Answer        string  `json:"answer"`
	BookmarkID    string  `json:"bookmarkId"`
	BookmarkTitle string  `json:"bookmarkTitle"`
	BookmarkURL   string  `json:"bookmarkUrl"`
	Title         string  `json:"title"` // backwards-compatible: same as BookmarkTitle
	URL           string  `json:"url"`   // backwards-compatible: same as BookmarkURL
	Summary       string  `json:"summary"`
	Similarity    float64 `json:"similarity"`
	Strategy      string  `json:"strategy"`
}

// AdvancedSearchResult contains the full result of an advanced LLM-powered search
type AdvancedSearchResult struct {
	Query            string          `json:"query"`
	QueryString      string          `json:"queryString"`
	SimilarQuestions []RetrievedItem `json:"similarQuestions"`
	RenderedPrompt   string          `json:"renderedPrompt"`
	ThinkingProcess  string          `json:"thinkingProcess"`
	FullResponse     string          `json:"fullResponse"`
	Answer           string          `json:"answer"`
}

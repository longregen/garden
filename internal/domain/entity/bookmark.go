package entity

import (
	"time"

	"github.com/google/uuid"
)

// Bookmark represents a basic bookmark
type Bookmark struct {
	BookmarkID   uuid.UUID
	URL          string
	CreationDate time.Time
}

// BookmarkWithTitle represents a bookmark with its title
type BookmarkWithTitle struct {
	BookmarkID   uuid.UUID `json:"bookmark_id"`
	URL          string    `json:"url"`
	CreationDate time.Time `json:"creation_date"`
	Title        *string   `json:"title,omitempty"`
	Summary      *string   `json:"summary,omitempty"`
}

// BookmarkDetails represents a complete bookmark with all relations
type BookmarkDetails struct {
	BookmarkID   uuid.UUID              `json:"bookmark_id"`
	URL          string                 `json:"url"`
	CreationDate time.Time              `json:"creation_date"`
	CategoryName *string                `json:"category_name,omitempty"`
	SourceURI    *string                `json:"source_uri,omitempty"`
	RawSource    *string                `json:"raw_source,omitempty"`
	Title        *string                `json:"title,omitempty"`
	LynxContent  *string                `json:"lynx,omitempty"`
	ReaderContent *string               `json:"reader,omitempty"`
	Summary      *string                `json:"summary,omitempty"`
	StatusCode   *int32                 `json:"status_code,omitempty"`
	Headers      *string                `json:"headers,omitempty"`
	HTTPContent  []byte                 `json:"content,omitempty"`
	FetchDate    *time.Time             `json:"fetch_date,omitempty"`
	Questions    []BookmarkQuestion     `json:"questions,omitempty"`
}

// BookmarkQuestion represents a Q&A pair for a bookmark
type BookmarkQuestion struct {
	ID      uuid.UUID `json:"id"`
	Content string    `json:"content"`
}

// BookmarkContentReference represents content reference with embeddings
type BookmarkContentReference struct {
	ID         uuid.UUID
	BookmarkID uuid.UUID
	Content    string
	Strategy   string
	Embedding  []float32
}

// BookmarkFilters represents filter criteria for listing bookmarks
type BookmarkFilters struct {
	CategoryID        *uuid.UUID
	SearchQuery       *string
	StartCreationDate *time.Time
	EndCreationDate   *time.Time
	Page              int32
	Limit             int32
}

// UpdateQuestionInput represents input for updating a Q&A
type UpdateQuestionInput struct {
	BookmarkID       uuid.UUID
	ReferenceID      uuid.UUID
	PreviousQuestion string
	PreviousAnswer   string
	NewQuestion      string
	NewAnswer        string
}

// DeleteQuestionInput represents input for deleting a Q&A
type DeleteQuestionInput struct {
	BookmarkID  uuid.UUID
	ReferenceID uuid.UUID
	Question    string
	Answer      string
}

// ProcessingResult represents the result of processing operations
type ProcessingResult struct {
	Message string  `json:"message"`
	Content *string `json:"content,omitempty"`
}

// FetchResult represents the result of fetching bookmark content
type FetchResult struct {
	StatusCode int32  `json:"status_code"`
	Headers    string `json:"headers"`
	Content    []byte `json:"content"`
	Message    string `json:"message"`
}

// EmbeddingResult represents the result of creating embeddings
type EmbeddingResult struct {
	IDs     []uuid.UUID `json:"ids"`
	Warning *string     `json:"warning,omitempty"`
}

// SummaryEmbeddingResult represents the result of creating a summary embedding
type SummaryEmbeddingResult struct {
	IDs     []uuid.UUID `json:"ids"`
	Summary string      `json:"summary"`
	Warning *string     `json:"warning,omitempty"`
}

// TitleExtractionResult represents the result of title extraction
type TitleExtractionResult struct {
	Data  BookmarkDetails
	Title *string
}

// Embedding represents a text embedding
type Embedding struct {
	Text      string
	Embedding []float32
}

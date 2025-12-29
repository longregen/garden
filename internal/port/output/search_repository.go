package output

import (
	"context"

	"garden3/internal/domain/entity"
)

// SearchRepository defines the interface for search data operations
type SearchRepository interface {
	// SearchAll performs a unified search across multiple tables
	SearchAll(ctx context.Context, query string, exactMatchWeight, similarityWeight, recencyWeight float64, limit int32) ([]entity.UnifiedSearchResult, error)

	// GetSimilarQuestions retrieves bookmarks with similar Q&A content using vector similarity
	GetSimilarQuestions(ctx context.Context, embedding []float32, limit int32) ([]entity.RetrievedItem, error)
}

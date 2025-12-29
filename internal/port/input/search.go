package input

import (
	"context"

	"garden3/internal/domain/entity"
)

// SearchUseCase defines the business operations for unified search
type SearchUseCase interface {
	// SearchAll performs a unified search across multiple tables
	SearchAll(ctx context.Context, query string, weights *entity.SearchWeights, limit int32) ([]entity.UnifiedSearchResult, error)

	// AdvancedSearch performs an LLM-powered search with context from similar bookmarks
	AdvancedSearch(ctx context.Context, query string) (*entity.AdvancedSearchResult, error)
}

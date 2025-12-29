package input

import (
	"context"

	"garden3/internal/domain/entity"
)

// BrowserHistoryUseCase defines the business operations for browser history
type BrowserHistoryUseCase interface {
	// ListHistory retrieves paginated and filtered browser history
	ListHistory(ctx context.Context, filters entity.BrowserHistoryFilters) (*PaginatedResponse[entity.BrowserHistory], error)

	// TopDomains retrieves the most visited domains
	TopDomains(ctx context.Context, limit int32) ([]entity.DomainVisitCount, error)

	// RecentHistory retrieves the N most recent history entries
	RecentHistory(ctx context.Context, limit int32) ([]entity.BrowserHistory, error)
}

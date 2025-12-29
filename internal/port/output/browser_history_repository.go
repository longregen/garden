package output

import (
	"context"

	"garden3/internal/domain/entity"
)

// BrowserHistoryRepository defines the interface for browser history data operations
type BrowserHistoryRepository interface {
	// ListHistory retrieves paginated and filtered browser history
	ListHistory(ctx context.Context, filters entity.BrowserHistoryFilters) ([]entity.BrowserHistory, int64, error)

	// TopDomains retrieves the most visited domains
	TopDomains(ctx context.Context, limit int32) ([]entity.DomainVisitCount, error)

	// RecentHistory retrieves the N most recent history entries
	RecentHistory(ctx context.Context, limit int32) ([]entity.BrowserHistory, error)
}

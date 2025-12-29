package output

import (
	"context"

	"garden3/internal/domain/entity"
)

// DashboardRepository defines the interface for dashboard data operations
type DashboardRepository interface {
	// GetContactStats retrieves contact statistics
	GetContactStats(ctx context.Context) (*entity.CategoryStats, error)

	// GetSessionStats retrieves session statistics
	GetSessionStats(ctx context.Context) (*entity.CategoryStats, error)

	// GetBookmarkStats retrieves bookmark statistics
	GetBookmarkStats(ctx context.Context) (*entity.CategoryStats, error)

	// GetHistoryStats retrieves browser history statistics
	GetHistoryStats(ctx context.Context) (*entity.CategoryStats, error)

	// GetRecentItems retrieves recent items from all categories
	GetRecentItems(ctx context.Context) ([]entity.RecentItem, error)
}

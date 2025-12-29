package service

import (
	"context"

	"garden3/internal/domain/entity"
	"garden3/internal/port/output"
)

// DashboardService implements the dashboard use case
type DashboardService struct {
	repo output.DashboardRepository
}

// NewDashboardService creates a new dashboard service
func NewDashboardService(repo output.DashboardRepository) *DashboardService {
	return &DashboardService{
		repo: repo,
	}
}

// GetStats retrieves comprehensive dashboard statistics
func (s *DashboardService) GetStats(ctx context.Context) (*entity.DashboardStats, error) {
	contactStats, err := s.repo.GetContactStats(ctx)
	if err != nil {
		return nil, err
	}

	sessionStats, err := s.repo.GetSessionStats(ctx)
	if err != nil {
		return nil, err
	}

	bookmarkStats, err := s.repo.GetBookmarkStats(ctx)
	if err != nil {
		return nil, err
	}

	historyStats, err := s.repo.GetHistoryStats(ctx)
	if err != nil {
		return nil, err
	}

	recentItems, err := s.repo.GetRecentItems(ctx)
	if err != nil {
		return nil, err
	}

	return &entity.DashboardStats{
		Contacts:    *contactStats,
		Sessions:    *sessionStats,
		Bookmarks:   *bookmarkStats,
		History:     *historyStats,
		RecentItems: recentItems,
	}, nil
}

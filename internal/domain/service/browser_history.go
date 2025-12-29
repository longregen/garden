package service

import (
	"context"
	"math"

	"garden3/internal/domain/entity"
	"garden3/internal/port/input"
	"garden3/internal/port/output"
)

// BrowserHistoryService implements the browser history use case
type BrowserHistoryService struct {
	repo output.BrowserHistoryRepository
}

// NewBrowserHistoryService creates a new browser history service
func NewBrowserHistoryService(repo output.BrowserHistoryRepository) *BrowserHistoryService {
	return &BrowserHistoryService{
		repo: repo,
	}
}

// ListHistory retrieves paginated and filtered browser history
func (s *BrowserHistoryService) ListHistory(ctx context.Context, filters entity.BrowserHistoryFilters) (*input.PaginatedResponse[entity.BrowserHistory], error) {
	if filters.Page <= 0 {
		filters.Page = 1
	}
	if filters.PageSize <= 0 {
		filters.PageSize = 10
	}

	offset := (filters.Page - 1) * filters.PageSize
	filters.Page = offset

	items, total, err := s.repo.ListHistory(ctx, filters)
	if err != nil {
		return nil, err
	}

	totalPages := int32(math.Ceil(float64(total) / float64(filters.PageSize)))

	return &input.PaginatedResponse[entity.BrowserHistory]{
		Data:       items,
		Total:      total,
		Page:       (offset / filters.PageSize) + 1,
		PageSize:   filters.PageSize,
		TotalPages: totalPages,
	}, nil
}

// TopDomains retrieves the most visited domains
func (s *BrowserHistoryService) TopDomains(ctx context.Context, limit int32) ([]entity.DomainVisitCount, error) {
	if limit <= 0 {
		limit = 10
	}

	return s.repo.TopDomains(ctx, limit)
}

// RecentHistory retrieves the N most recent history entries
func (s *BrowserHistoryService) RecentHistory(ctx context.Context, limit int32) ([]entity.BrowserHistory, error) {
	if limit <= 0 {
		limit = 20
	}

	return s.repo.RecentHistory(ctx, limit)
}

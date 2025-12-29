package repository

import (
	"context"
	"sort"
	"time"

	"garden3/internal/adapter/secondary/postgres/generated/db"
	"garden3/internal/domain/entity"
	"garden3/internal/port/output"

	"github.com/jackc/pgx/v5/pgxpool"
)

type dashboardRepository struct {
	queries *db.Queries
}

// NewDashboardRepository creates a new dashboard repository
func NewDashboardRepository(pool *pgxpool.Pool) output.DashboardRepository {
	return &dashboardRepository{
		queries: db.New(pool),
	}
}

// calculateMonthOverMonthChange calculates the percentage change between current and previous period
func calculateMonthOverMonthChange(current, previous int64) float64 {
	if previous == 0 {
		return 0
	}
	return ((float64(current) - float64(previous)) / float64(previous)) * 100
}

// GetContactStats retrieves contact statistics
func (r *dashboardRepository) GetContactStats(ctx context.Context) (*entity.CategoryStats, error) {
	row, err := r.queries.GetContactStats(ctx)
	if err != nil {
		return nil, err
	}

	return &entity.CategoryStats{
		Total:                row.Total,
		RecentlyActive:       row.RecentlyActive,
		MonthOverMonthChange: calculateMonthOverMonthChange(row.CurrentPeriodCount, row.PreviousPeriodCount),
	}, nil
}

// GetSessionStats retrieves session statistics
func (r *dashboardRepository) GetSessionStats(ctx context.Context) (*entity.CategoryStats, error) {
	row, err := r.queries.GetSessionStats(ctx)
	if err != nil {
		return nil, err
	}

	return &entity.CategoryStats{
		Total:                row.Total,
		RecentCount:          row.RecentCount,
		MonthOverMonthChange: calculateMonthOverMonthChange(row.CurrentPeriodCount, row.PreviousPeriodCount),
	}, nil
}

// GetBookmarkStats retrieves bookmark statistics
func (r *dashboardRepository) GetBookmarkStats(ctx context.Context) (*entity.CategoryStats, error) {
	row, err := r.queries.GetBookmarkStats(ctx)
	if err != nil {
		return nil, err
	}

	return &entity.CategoryStats{
		Total:                row.Total,
		RecentCount:          row.RecentCount,
		MonthOverMonthChange: calculateMonthOverMonthChange(row.CurrentPeriodCount, row.PreviousPeriodCount),
	}, nil
}

// GetHistoryStats retrieves browser history statistics
func (r *dashboardRepository) GetHistoryStats(ctx context.Context) (*entity.CategoryStats, error) {
	row, err := r.queries.GetHistoryStats(ctx)
	if err != nil {
		return nil, err
	}

	return &entity.CategoryStats{
		Total:                row.Total,
		RecentCount:          row.RecentCount,
		MonthOverMonthChange: calculateMonthOverMonthChange(row.CurrentPeriodCount, row.PreviousPeriodCount),
	}, nil
}

// GetRecentItems retrieves recent items from all categories
func (r *dashboardRepository) GetRecentItems(ctx context.Context) ([]entity.RecentItem, error) {
	contacts, err := r.queries.GetRecentContacts(ctx)
	if err != nil {
		return nil, err
	}

	sessions, err := r.queries.GetRecentSessions(ctx)
	if err != nil {
		return nil, err
	}

	bookmarks, err := r.queries.GetRecentBookmarks(ctx)
	if err != nil {
		return nil, err
	}

	history, err := r.queries.GetRecentHistory(ctx)
	if err != nil {
		return nil, err
	}

	notes, err := r.queries.GetRecentNotes(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]entity.RecentItem, 0)

	for _, c := range contacts {
		items = append(items, entity.RecentItem{
			ID:       c.ID,
			Category: c.Category,
			Name:     c.Name,
			Date:     c.Date.Time,
		})
	}

	for _, s := range sessions {
		items = append(items, entity.RecentItem{
			ID:       s.ID,
			Category: s.Category,
			Name:     s.Name,
			Date:     s.Date.Time,
		})
	}

	for _, b := range bookmarks {
		items = append(items, entity.RecentItem{
			ID:       b.ID,
			Category: b.Category,
			Name:     b.Name,
			Date:     b.Date.Time,
		})
	}

	for _, h := range history {
		items = append(items, entity.RecentItem{
			ID:       h.ID,
			Category: h.Category,
			Name:     h.Name,
			Date:     h.Date.Time,
		})
	}

	for _, n := range notes {
		if dateVal, ok := n.Date.(time.Time); ok {
			items = append(items, entity.RecentItem{
				ID:       n.ID,
				Category: n.Category,
				Name:     n.Name,
				Date:     dateVal,
			})
		}
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Date.After(items[j].Date)
	})

	if len(items) > 10 {
		items = items[:10]
	}

	return items, nil
}

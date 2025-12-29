package repository

import (
	"context"
	"time"

	"garden3/internal/adapter/secondary/postgres/generated/db"
	"garden3/internal/domain/entity"
	"garden3/internal/port/output"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type browserHistoryRepository struct {
	queries *db.Queries
}

// NewBrowserHistoryRepository creates a new browser history repository
func NewBrowserHistoryRepository(pool *pgxpool.Pool) output.BrowserHistoryRepository {
	return &browserHistoryRepository{
		queries: db.New(pool),
	}
}

// ListHistory retrieves paginated and filtered browser history
func (r *browserHistoryRepository) ListHistory(ctx context.Context, filters entity.BrowserHistoryFilters) ([]entity.BrowserHistory, int64, error) {
	var startDate, endDate pgtype.Timestamp
	if filters.StartDate != nil {
		startDate = pgtype.Timestamp{Time: *filters.StartDate, Valid: true}
	}
	if filters.EndDate != nil {
		endDate = pgtype.Timestamp{Time: *filters.EndDate, Valid: true}
	}

	var searchQuery, domain *string
	if filters.SearchQuery != nil {
		searchQuery = filters.SearchQuery
	}
	if filters.Domain != nil {
		domain = filters.Domain
	}

	rows, err := r.queries.ListBrowserHistory(ctx, db.ListBrowserHistoryParams{
		SearchQuery: searchQuery,
		StartDate:   startDate,
		EndDate:     endDate,
		Domain:      domain,
		PageSize:    filters.PageSize,
		PageOffset:  filters.Page,
	})
	if err != nil {
		return nil, 0, err
	}

	total, err := r.queries.CountBrowserHistory(ctx, db.CountBrowserHistoryParams{
		SearchQuery: searchQuery,
		StartDate:   startDate,
		EndDate:     endDate,
		Domain:      domain,
	})
	if err != nil {
		return nil, 0, err
	}

	items := make([]entity.BrowserHistory, 0, len(rows))
	for _, row := range rows {
		items = append(items, entity.BrowserHistory{
			ID:                         row.ID,
			URL:                        row.Url,
			Title:                      nullableString(row.Title),
			VisitDate:                  nullableTimestamp(&row.VisitDate),
			Typed:                      row.Typed,
			Hidden:                     row.Hidden,
			ImportedFromFirefoxPlaceID: row.ImportedFromFirefoxPlaceID,
			ImportedFromFirefoxVisitID: row.ImportedFromFirefoxVisitID,
			Domain:                     nullableString(row.Domain),
			CreatedAt:                  nullableTimestamptz(&row.CreatedAt),
		})
	}

	return items, total, nil
}

// TopDomains retrieves the most visited domains
func (r *browserHistoryRepository) TopDomains(ctx context.Context, limit int32) ([]entity.DomainVisitCount, error) {
	rows, err := r.queries.GetTopDomains(ctx, limit)
	if err != nil {
		return nil, err
	}

	results := make([]entity.DomainVisitCount, 0, len(rows))
	for _, row := range rows {
		if row.Domain != nil {
			results = append(results, entity.DomainVisitCount{
				Domain:     *row.Domain,
				VisitCount: row.VisitCount,
			})
		}
	}

	return results, nil
}

// RecentHistory retrieves the N most recent history entries
func (r *browserHistoryRepository) RecentHistory(ctx context.Context, limit int32) ([]entity.BrowserHistory, error) {
	rows, err := r.queries.GetRecentBrowserHistory(ctx, limit)
	if err != nil {
		return nil, err
	}

	items := make([]entity.BrowserHistory, 0, len(rows))
	for _, row := range rows {
		items = append(items, entity.BrowserHistory{
			ID:                         row.ID,
			URL:                        row.Url,
			Title:                      nullableString(row.Title),
			VisitDate:                  nullableTimestamp(&row.VisitDate),
			Typed:                      row.Typed,
			Hidden:                     row.Hidden,
			ImportedFromFirefoxPlaceID: row.ImportedFromFirefoxPlaceID,
			ImportedFromFirefoxVisitID: row.ImportedFromFirefoxVisitID,
			Domain:                     nullableString(row.Domain),
			CreatedAt:                  nullableTimestamptz(&row.CreatedAt),
		})
	}

	return items, nil
}

func stringOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func nullableString(s *string) *string {
	return s
}

func nullableInt32(i *int32) *int32 {
	return i
}

func nullableTimestamptz(t *pgtype.Timestamptz) *time.Time {
	if t == nil || !t.Valid {
		return nil
	}
	return &t.Time
}

func nullableTimestamp(t *pgtype.Timestamp) *time.Time {
	if t == nil || !t.Valid {
		return nil
	}
	return &t.Time
}

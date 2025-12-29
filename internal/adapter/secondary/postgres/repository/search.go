package repository

import (
	"context"
	"strings"

	"garden3/internal/adapter/secondary/postgres/generated/db"
	"garden3/internal/domain/entity"
	"garden3/internal/port/output"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
)

type searchRepository struct {
	queries *db.Queries
}

// NewSearchRepository creates a new search repository
func NewSearchRepository(pool *pgxpool.Pool) output.SearchRepository {
	return &searchRepository{
		queries: db.New(pool),
	}
}

// SearchAll performs a unified search across multiple tables
func (r *searchRepository) SearchAll(ctx context.Context, query string, exactMatchWeight, similarityWeight, recencyWeight float64, limit int32) ([]entity.UnifiedSearchResult, error) {
	rows, err := r.queries.SearchAll(ctx, db.SearchAllParams{
		Query:            query,
		ExactMatchWeight: exactMatchWeight,
		SimilarityWeight: similarityWeight,
		RecencyWeight:    recencyWeight,
		ResultLimit:      limit,
	})
	if err != nil {
		return nil, err
	}

	results := make([]entity.UnifiedSearchResult, 0, len(rows))
	for _, row := range rows {
		results = append(results, entity.UnifiedSearchResult{
			ItemType:     row.ItemType,
			ItemID:       row.ItemID,
			ItemTitle:    row.ItemTitle,
			LastActivity: row.LastActivity.Time,
			SearchScore:  float64(row.SearchScore),
		})
	}

	return results, nil
}

// GetSimilarQuestions retrieves bookmarks with similar Q&A content using vector similarity
func (r *searchRepository) GetSimilarQuestions(ctx context.Context, embedding []float32, limit int32) ([]entity.RetrievedItem, error) {
	// Convert embedding to pgvector format
	vec := pgvector.NewVector(embedding)

	rows, err := r.queries.GetSimilarQuestions(ctx, db.GetSimilarQuestionsParams{
		Embedding:   &vec,
		SearchLimit: limit,
	})
	if err != nil {
		return nil, err
	}

	results := make([]entity.RetrievedItem, 0, len(rows))
	for i, row := range rows {
		// Split question into question and answer parts
		question := ""
		answer := ""
		if row.Question != nil {
			parts := strings.SplitN(*row.Question, "\n", 2)
			question = parts[0]
			if len(parts) > 1 {
				answer = parts[1]
			}
		}

		bookmarkIDStr := row.BookmarkID.String()
		strategy := ""
		if row.Strategy != nil {
			strategy = *row.Strategy
		}

		results = append(results, entity.RetrievedItem{
			ID:            i + 1,
			Question:      question,
			Answer:        answer,
			BookmarkID:    bookmarkIDStr,
			BookmarkTitle: row.Title,
			BookmarkURL:   row.Url,
			Title:         row.Title, // backwards-compatible
			URL:           row.Url,   // backwards-compatible
			Summary:       row.Summary,
			Similarity:    float64(row.Similarity),
			Strategy:      strategy,
		})
	}

	return results, nil
}

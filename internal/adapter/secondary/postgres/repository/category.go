package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"garden3/internal/adapter/secondary/postgres/generated/db"
	"garden3/internal/domain/entity"
)

// CategoryRepository implements the output.CategoryRepository interface
type CategoryRepository struct {
	pool *pgxpool.Pool
}

// NewCategoryRepository creates a new category repository
func NewCategoryRepository(pool *pgxpool.Pool) *CategoryRepository {
	return &CategoryRepository{
		pool: pool,
	}
}

func (r *CategoryRepository) ListCategoriesWithSources(ctx context.Context) ([]entity.CategoryWithSources, error) {
	queries := db.New(r.pool)
	rows, err := queries.ListCategoriesWithSources(ctx)
	if err != nil {
		return nil, err
	}

	categories := make([]entity.CategoryWithSources, len(rows))
	for i, row := range rows {
		// Parse the JSON sources array
		var sources []entity.CategorySource
		if row.Sources != nil {
			sourcesJSON, err := json.Marshal(row.Sources)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal sources: %w", err)
			}

			// Parse into a slice of raw sources
			var rawSources []struct {
				ID         uuid.UUID       `json:"id"`
				CategoryID uuid.UUID       `json:"category_id"`
				SourceURI  *string         `json:"source_uri"`
				RawSource  json.RawMessage `json:"raw_source"`
			}
			if err := json.Unmarshal(sourcesJSON, &rawSources); err != nil {
				return nil, fmt.Errorf("failed to unmarshal sources: %w", err)
			}

			sources = make([]entity.CategorySource, len(rawSources))
			for j, rs := range rawSources {
				sources[j] = entity.CategorySource{
					ID:         rs.ID,
					CategoryID: rs.CategoryID,
					SourceURI:  rs.SourceURI,
					RawSource:  rs.RawSource,
				}
			}
		} else {
			sources = []entity.CategorySource{}
		}

		categories[i] = entity.CategoryWithSources{
			Category: entity.Category{
				CategoryID: row.CategoryID,
				Name:       row.Name,
			},
			Sources: sources,
		}
	}

	return categories, nil
}

func (r *CategoryRepository) GetCategory(ctx context.Context, categoryID uuid.UUID) (*entity.Category, error) {
	queries := db.New(r.pool)
	dbCategory, err := queries.GetCategory(ctx, categoryID)
	if err != nil {
		return nil, err
	}

	return &entity.Category{
		CategoryID: dbCategory.CategoryID,
		Name:       dbCategory.Name,
	}, nil
}

func (r *CategoryRepository) UpdateCategory(ctx context.Context, categoryID uuid.UUID, name string) error {
	queries := db.New(r.pool)
	return queries.UpdateCategory(ctx, db.UpdateCategoryParams{
		CategoryID: categoryID,
		Name:       name,
	})
}

func (r *CategoryRepository) MergeCategories(ctx context.Context, sourceID, targetID uuid.UUID) error {
	queries := db.New(r.pool)
	return queries.MergeCategories(ctx, db.MergeCategoriesParams{
		Column1: sourceID,
		Column2: targetID,
	})
}

func (r *CategoryRepository) CreateCategorySource(ctx context.Context, categoryID uuid.UUID, sourceURI *string, rawSource json.RawMessage) (*entity.CategorySource, error) {
	queries := db.New(r.pool)

	// Convert uuid.UUID to pgtype.UUID
	pgCategoryID := pgtype.UUID{
		Bytes: categoryID,
		Valid: true,
	}

	dbSource, err := queries.CreateCategorySource(ctx, db.CreateCategorySourceParams{
		CategoryID: pgCategoryID,
		SourceUri:  sourceURI,
		RawSource:  []byte(rawSource),
	})
	if err != nil {
		return nil, err
	}

	// Convert pgtype.UUID back to uuid.UUID
	var resultCategoryID uuid.UUID
	if dbSource.CategoryID.Valid {
		resultCategoryID = uuid.UUID(dbSource.CategoryID.Bytes)
	}

	return &entity.CategorySource{
		ID:         dbSource.ID,
		CategoryID: resultCategoryID,
		SourceURI:  dbSource.SourceUri,
		RawSource:  json.RawMessage(dbSource.RawSource),
	}, nil
}

func (r *CategoryRepository) UpdateCategorySource(ctx context.Context, sourceID uuid.UUID, sourceURI *string, rawSource json.RawMessage) error {
	queries := db.New(r.pool)
	return queries.UpdateCategorySource(ctx, db.UpdateCategorySourceParams{
		ID:        sourceID,
		SourceUri: sourceURI,
		RawSource: []byte(rawSource),
	})
}

func (r *CategoryRepository) DeleteCategorySource(ctx context.Context, sourceID uuid.UUID) error {
	queries := db.New(r.pool)
	return queries.DeleteCategorySource(ctx, sourceID)
}

func (r *CategoryRepository) GetCategorySource(ctx context.Context, sourceID uuid.UUID) (*entity.CategorySource, error) {
	queries := db.New(r.pool)
	dbSource, err := queries.GetCategorySource(ctx, sourceID)
	if err != nil {
		return nil, err
	}

	// Convert pgtype.UUID to uuid.UUID
	var categoryID uuid.UUID
	if dbSource.CategoryID.Valid {
		categoryID = uuid.UUID(dbSource.CategoryID.Bytes)
	}

	return &entity.CategorySource{
		ID:         dbSource.ID,
		CategoryID: categoryID,
		SourceURI:  dbSource.SourceUri,
		RawSource:  json.RawMessage(dbSource.RawSource),
	}, nil
}

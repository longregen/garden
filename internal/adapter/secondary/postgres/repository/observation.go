package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"garden3/internal/adapter/secondary/postgres/generated/db"
	"garden3/internal/domain/entity"
	"garden3/internal/port/output"
)

type observationRepository struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

func NewObservationRepository(pool *pgxpool.Pool) output.ObservationRepository {
	return &observationRepository{
		pool:    pool,
		queries: db.New(pool),
	}
}

func (r *observationRepository) Create(ctx context.Context, data []byte, obsType, source, tags string, ref uuid.UUID) (*entity.Observation, error) {
	row, err := r.queries.CreateObservationWithReturn(ctx, db.CreateObservationWithReturnParams{
		Data:   data,
		Type:   &obsType,
		Source: &source,
		Tags:   &tags,
		Ref:    convertUUIDToPgUUID(ref),
	})
	if err != nil {
		return nil, err
	}

	observation := toEntityObservation(row)
	return &observation, nil
}

func (r *observationRepository) GetFeedbackStats(ctx context.Context, bookmarkID uuid.UUID) (*entity.FeedbackStats, error) {
	pgUUID := convertUUIDToPgUUID(bookmarkID)
	row, err := r.queries.GetFeedbackStats(ctx, pgUUID)
	if err != nil {
		return nil, err
	}

	stats := &entity.FeedbackStats{
		Upvotes:   row.Upvotes,
		Downvotes: row.Downvotes,
		Trash:     row.Trash,
	}

	return stats, nil
}

func (r *observationRepository) DeleteBookmarkContentReference(ctx context.Context, referenceID, bookmarkID uuid.UUID) error {
	return r.queries.DeleteBookmarkContentReference(ctx, db.DeleteBookmarkContentReferenceParams{
		ID:         referenceID,
		BookmarkID: convertUUIDToPgUUID(bookmarkID),
	})
}

func toEntityObservation(row db.Observation) entity.Observation {
	obs := entity.Observation{
		ObservationID: row.ObservationID,
		Data:          row.Data,
		Type:          row.Type,
		Source:        row.Source,
		Tags:          row.Tags,
		CreationDate:  row.CreationDate,
	}

	obs.Parent = convertPgUUIDToUUIDPtr(row.Parent)
	obs.Ref = convertPgUUIDToUUIDPtr(row.Ref)

	return obs
}

package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"garden3/internal/adapter/secondary/postgres/generated/db"
	"garden3/internal/domain/entity"
	"garden3/internal/port/output"
)

type socialPostRepository struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

func NewSocialPostRepository(pool *pgxpool.Pool) output.SocialPostRepository {
	return &socialPostRepository{
		pool:    pool,
		queries: db.New(pool),
	}
}

func (r *socialPostRepository) List(ctx context.Context, filters entity.SocialPostFilters) ([]entity.SocialPost, int64, error) {
	offset := (filters.Page - 1) * filters.Limit

	var status string
	if filters.Status != nil {
		status = *filters.Status
	}

	rows, err := r.queries.ListSocialPosts(ctx, db.ListSocialPostsParams{
		Column1: status,
		Limit:   filters.Limit,
		Offset:  offset,
	})
	if err != nil {
		return nil, 0, err
	}

	count, err := r.queries.CountSocialPosts(ctx, status)
	if err != nil {
		return nil, 0, err
	}

	posts := make([]entity.SocialPost, len(rows))
	for i, row := range rows {
		posts[i] = toEntitySocialPost(row)
	}

	return posts, count, nil
}

func (r *socialPostRepository) GetByID(ctx context.Context, postID uuid.UUID) (*entity.SocialPost, error) {
	row, err := r.queries.GetSocialPost(ctx, postID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	post := toEntitySocialPost(row)
	return &post, nil
}

func (r *socialPostRepository) Create(ctx context.Context, content string, status string) (*entity.SocialPost, error) {
	row, err := r.queries.CreateSocialPost(ctx, db.CreateSocialPostParams{
		Content: content,
		Status:  status,
	})
	if err != nil {
		return nil, err
	}

	post := toEntitySocialPost(row)
	return &post, nil
}

func (r *socialPostRepository) UpdateStatus(ctx context.Context, postID uuid.UUID, status string, errorMessage *string) error {
	return r.queries.UpdateSocialPostStatus(ctx, db.UpdateSocialPostStatusParams{
		Status:       status,
		ErrorMessage: errorMessage,
		PostID:       postID,
	})
}

func (r *socialPostRepository) UpdateTwitterID(ctx context.Context, postID uuid.UUID, twitterPostID string) error {
	return r.queries.UpdateSocialPostTwitterID(ctx, db.UpdateSocialPostTwitterIDParams{
		TwitterPostID: &twitterPostID,
		PostID:        postID,
	})
}

func (r *socialPostRepository) UpdateBlueskyID(ctx context.Context, postID uuid.UUID, blueskyPostID string) error {
	return r.queries.UpdateSocialPostBlueskyID(ctx, db.UpdateSocialPostBlueskyIDParams{
		BlueskyPostID: &blueskyPostID,
		PostID:        postID,
	})
}

func (r *socialPostRepository) Delete(ctx context.Context, postID uuid.UUID) error {
	return r.queries.DeleteSocialPost(ctx, postID)
}

func toEntitySocialPost(row db.SocialPost) entity.SocialPost {
	post := entity.SocialPost{
		PostID:        row.PostID,
		Content:       row.Content,
		TwitterPostID: row.TwitterPostID,
		BlueskyPostID: row.BlueskyPostID,
		CreatedAt:     row.CreatedAt.Time,
		Status:        row.Status,
		ErrorMessage:  row.ErrorMessage,
	}

	if row.UpdatedAt.Valid {
		post.UpdatedAt = &row.UpdatedAt.Time
	}

	return post
}

package output

import (
	"context"

	"github.com/google/uuid"
	"garden3/internal/domain/entity"
)

// SocialPostRepository defines persistence operations for social posts
type SocialPostRepository interface {
	// List retrieves paginated social posts with optional status filter
	List(ctx context.Context, filters entity.SocialPostFilters) ([]entity.SocialPost, int64, error)

	// GetByID retrieves a single social post by ID
	GetByID(ctx context.Context, postID uuid.UUID) (*entity.SocialPost, error)

	// Create creates a new social post
	Create(ctx context.Context, content string, status string) (*entity.SocialPost, error)

	// UpdateStatus updates the status of a social post
	UpdateStatus(ctx context.Context, postID uuid.UUID, status string, errorMessage *string) error

	// UpdateTwitterID updates the Twitter post ID
	UpdateTwitterID(ctx context.Context, postID uuid.UUID, twitterPostID string) error

	// UpdateBlueskyID updates the Bluesky post ID
	UpdateBlueskyID(ctx context.Context, postID uuid.UUID, blueskyPostID string) error

	// Delete deletes a social post
	Delete(ctx context.Context, postID uuid.UUID) error
}

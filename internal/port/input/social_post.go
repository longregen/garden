package input

import (
	"context"

	"github.com/google/uuid"
	"garden3/internal/domain/entity"
)

// SocialPostUseCase defines the business operations for social posts
type SocialPostUseCase interface {
	// ListPosts retrieves paginated social posts with optional status filter
	ListPosts(ctx context.Context, filters entity.SocialPostFilters) (*PaginatedResponse[entity.SocialPost], error)

	// GetPost retrieves a single social post by ID
	GetPost(ctx context.Context, postID uuid.UUID) (*entity.SocialPost, error)

	// CreatePost creates and posts to social media platforms
	CreatePost(ctx context.Context, input entity.CreateSocialPostInput) (*entity.PostResult, error)

	// UpdateStatus updates the status of a social post
	UpdateStatus(ctx context.Context, postID uuid.UUID, input entity.UpdateStatusInput) error

	// DeletePost deletes a social post
	DeletePost(ctx context.Context, postID uuid.UUID) error

	// CheckCredentials checks the status of social media credentials
	CheckCredentials(ctx context.Context) (*entity.CredentialsStatus, error)

	// UpdateTwitterTokens updates Twitter OAuth tokens from authorization flow
	UpdateTwitterTokens(ctx context.Context, tokens entity.TwitterTokens) error

	// InitiateTwitterAuth generates the Twitter OAuth authorization URL
	InitiateTwitterAuth(ctx context.Context) (*entity.TwitterAuthURL, error)

	// HandleTwitterCallback processes the OAuth callback and exchanges code for tokens
	HandleTwitterCallback(ctx context.Context, input entity.TwitterCallbackInput) error
}

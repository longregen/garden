package output

import (
	"context"

	"garden3/internal/domain/entity"
)

// SocialMediaService defines operations for posting to social media platforms
type SocialMediaService interface {
	// PostToTwitter posts content to Twitter and returns the post ID
	PostToTwitter(ctx context.Context, content string) (string, error)

	// PostToBluesky posts content to Bluesky and returns the post ID
	PostToBluesky(ctx context.Context, content string) (string, error)

	// CheckTwitterCredentials checks if Twitter credentials are valid
	CheckTwitterCredentials(ctx context.Context) (*entity.TwitterProfile, error)

	// CheckBlueskyCredentials checks if Bluesky credentials are valid
	CheckBlueskyCredentials(ctx context.Context) (*entity.BlueskyProfile, error)

	// UpdateTwitterTokens updates Twitter OAuth tokens
	UpdateTwitterTokens(ctx context.Context, tokens entity.TwitterTokens) error

	// InitiateTwitterAuth generates the Twitter OAuth authorization URL
	InitiateTwitterAuth(ctx context.Context) (*entity.TwitterAuthURL, error)

	// HandleTwitterCallback processes the OAuth callback and exchanges code for tokens
	HandleTwitterCallback(ctx context.Context, input entity.TwitterCallbackInput) error
}

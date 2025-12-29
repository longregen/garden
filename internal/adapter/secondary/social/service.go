package social

import (
	"context"

	"garden3/internal/domain/entity"
	"garden3/internal/port/output"
)

// Service aggregates Twitter and Bluesky adapters
type Service struct {
	twitter *TwitterAdapter
	bluesky *BlueskyAdapter
}

// NewService creates a new social media service
func NewService(configRepo output.ConfigurationRepository) output.SocialMediaService {
	return &Service{
		twitter: NewTwitterAdapter(configRepo),
		bluesky: NewBlueskyAdapter(configRepo),
	}
}

func (s *Service) PostToTwitter(ctx context.Context, content string) (string, error) {
	return s.twitter.PostToTwitter(ctx, content)
}

func (s *Service) PostToBluesky(ctx context.Context, content string) (string, error) {
	return s.bluesky.PostToBluesky(ctx, content)
}

func (s *Service) CheckTwitterCredentials(ctx context.Context) (*entity.TwitterProfile, error) {
	return s.twitter.CheckCredentials(ctx)
}

func (s *Service) CheckBlueskyCredentials(ctx context.Context) (*entity.BlueskyProfile, error) {
	return s.bluesky.CheckCredentials(ctx)
}

func (s *Service) UpdateTwitterTokens(ctx context.Context, tokens entity.TwitterTokens) error {
	return s.twitter.UpdateTokens(ctx, tokens)
}

func (s *Service) InitiateTwitterAuth(ctx context.Context) (*entity.TwitterAuthURL, error) {
	return s.twitter.InitiateTwitterAuth(ctx)
}

func (s *Service) HandleTwitterCallback(ctx context.Context, input entity.TwitterCallbackInput) error {
	return s.twitter.HandleTwitterCallback(ctx, input)
}

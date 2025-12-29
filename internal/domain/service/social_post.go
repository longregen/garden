package service

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	"garden3/internal/domain/entity"
	"garden3/internal/port/input"
	"garden3/internal/port/output"
	"github.com/google/uuid"
)

const (
	// Default retry configuration
	defaultRetryCount = 3
	defaultRetryDelay = 2000 * time.Millisecond
)

// SocialPostService implements the SocialPostUseCase interface
type SocialPostService struct {
	repo        output.SocialPostRepository
	socialMedia output.SocialMediaService
	retryCount  int
	retryDelay  time.Duration
}

// NewSocialPostService creates a new social post service
func NewSocialPostService(
	repo output.SocialPostRepository,
	socialMedia output.SocialMediaService,
) *SocialPostService {
	return &SocialPostService{
		repo:        repo,
		socialMedia: socialMedia,
		retryCount:  defaultRetryCount,
		retryDelay:  defaultRetryDelay,
	}
}

func (s *SocialPostService) ListPosts(ctx context.Context, filters entity.SocialPostFilters) (*input.PaginatedResponse[entity.SocialPost], error) {
	page := filters.Page
	if page < 1 {
		page = 1
	}
	limit := filters.Limit
	if limit < 1 {
		limit = 10
	}

	posts, total, err := s.repo.List(ctx, entity.SocialPostFilters{
		Status: filters.Status,
		Page:   page,
		Limit:  limit,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list social posts: %w", err)
	}

	totalPages := int32((total + int64(limit) - 1) / int64(limit))

	return &input.PaginatedResponse[entity.SocialPost]{
		Data:       posts,
		Total:      total,
		Page:       page,
		PageSize:   limit,
		TotalPages: totalPages,
	}, nil
}

func (s *SocialPostService) GetPost(ctx context.Context, postID uuid.UUID) (*entity.SocialPost, error) {
	post, err := s.repo.GetByID(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("failed to get social post: %w", err)
	}
	return post, nil
}

func (s *SocialPostService) CreatePost(ctx context.Context, input entity.CreateSocialPostInput) (*entity.PostResult, error) {
	// Create the post in the database first with pending status
	post, err := s.repo.Create(ctx, input.Content, "pending")
	if err != nil {
		return nil, fmt.Errorf("failed to create social post: %w", err)
	}

	result := &entity.PostResult{
		PostID: post.PostID,
		Status: "pending",
	}

	// Track errors for both platforms
	var twitterErr, blueskyErr error
	hasSuccess := false

	// Attempt to post to Twitter with retry logic
	twitterPostID, twitterErr := s.retry(ctx, func() (string, error) {
		return s.socialMedia.PostToTwitter(ctx, input.Content)
	}, "Twitter post")

	if twitterErr == nil && twitterPostID != "" {
		if err := s.repo.UpdateTwitterID(ctx, post.PostID, twitterPostID); err == nil {
			result.TwitterPostID = &twitterPostID
			hasSuccess = true
		}
	}

	// Attempt to post to Bluesky with retry logic
	blueskyPostID, blueskyErr := s.retry(ctx, func() (string, error) {
		return s.socialMedia.PostToBluesky(ctx, input.Content)
	}, "Bluesky post")

	if blueskyErr == nil && blueskyPostID != "" {
		if err := s.repo.UpdateBlueskyID(ctx, post.PostID, blueskyPostID); err == nil {
			result.BlueskyPostID = &blueskyPostID
			hasSuccess = true
		}
	}

	// Determine final status
	var finalStatus string
	var errorMessage *string

	if !hasSuccess {
		finalStatus = "failed"
		errMsg := fmt.Sprintf("Twitter: %v; Bluesky: %v", twitterErr, blueskyErr)
		errorMessage = &errMsg
	} else if twitterErr != nil || blueskyErr != nil {
		finalStatus = "partial"
		errMsg := ""
		if twitterErr != nil {
			errMsg += fmt.Sprintf("Twitter: %v; ", twitterErr)
		}
		if blueskyErr != nil {
			errMsg += fmt.Sprintf("Bluesky: %v", blueskyErr)
		}
		errorMessage = &errMsg
	} else {
		finalStatus = "completed"
	}

	// Update the post status
	if err := s.repo.UpdateStatus(ctx, post.PostID, finalStatus, errorMessage); err != nil {
		return nil, fmt.Errorf("failed to update post status: %w", err)
	}

	result.Status = finalStatus
	result.ErrorMessage = errorMessage

	return result, nil
}

func (s *SocialPostService) UpdateStatus(ctx context.Context, postID uuid.UUID, input entity.UpdateStatusInput) error {
	if err := s.repo.UpdateStatus(ctx, postID, input.Status, input.ErrorMessage); err != nil {
		return fmt.Errorf("failed to update post status: %w", err)
	}
	return nil
}

func (s *SocialPostService) DeletePost(ctx context.Context, postID uuid.UUID) error {
	if err := s.repo.Delete(ctx, postID); err != nil {
		return fmt.Errorf("failed to delete post: %w", err)
	}
	return nil
}

func (s *SocialPostService) CheckCredentials(ctx context.Context) (*entity.CredentialsStatus, error) {
	status := &entity.CredentialsStatus{
		Twitter: entity.TwitterCredentials{
			Configured: false,
			Working:    false,
		},
		Bluesky: entity.BlueskyCredentials{
			Configured: false,
			Working:    false,
		},
	}

	// Check Twitter credentials
	twitterProfile, twitterErr := s.socialMedia.CheckTwitterCredentials(ctx)
	if twitterErr == nil {
		status.Twitter.Configured = true
		status.Twitter.Working = true
		status.Twitter.Profile = twitterProfile
	} else {
		status.Twitter.Configured = true // Assume configured if we got an error (not just missing)
		status.Twitter.Working = false
		errMsg := twitterErr.Error()
		status.Twitter.Error = &errMsg
	}

	// Check Bluesky credentials
	blueskyProfile, blueskyErr := s.socialMedia.CheckBlueskyCredentials(ctx)
	if blueskyErr == nil {
		status.Bluesky.Configured = true
		status.Bluesky.Working = true
		status.Bluesky.Profile = blueskyProfile
	} else {
		status.Bluesky.Configured = true // Assume configured if we got an error (not just missing)
		status.Bluesky.Working = false
		errMsg := blueskyErr.Error()
		status.Bluesky.Error = &errMsg
	}

	return status, nil
}

func (s *SocialPostService) UpdateTwitterTokens(ctx context.Context, tokens entity.TwitterTokens) error {
	if err := s.socialMedia.UpdateTwitterTokens(ctx, tokens); err != nil {
		return fmt.Errorf("failed to update Twitter tokens: %w", err)
	}
	return nil
}

func (s *SocialPostService) InitiateTwitterAuth(ctx context.Context) (*entity.TwitterAuthURL, error) {
	authURL, err := s.socialMedia.InitiateTwitterAuth(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initiate Twitter auth: %w", err)
	}
	return authURL, nil
}

func (s *SocialPostService) HandleTwitterCallback(ctx context.Context, input entity.TwitterCallbackInput) error {
	if err := s.socialMedia.HandleTwitterCallback(ctx, input); err != nil {
		return fmt.Errorf("failed to handle Twitter callback: %w", err)
	}
	return nil
}

// retry executes a function with exponential backoff and jitter
func (s *SocialPostService) retry(ctx context.Context, operation func() (string, error), name string) (string, error) {
	var lastErr error

	for attempt := 0; attempt < s.retryCount; attempt++ {
		result, err := operation()
		if err == nil {
			return result, nil
		}

		lastErr = err

		// Don't retry on the last attempt
		if attempt >= s.retryCount-1 {
			break
		}

		// Calculate exponential backoff with jitter (0.85-1.15 randomization)
		jitter := 0.85 + rand.Float64()*0.3 // 0.85-1.15
		backoffMultiplier := math.Pow(1.5, float64(attempt))
		sleepDuration := time.Duration(float64(s.retryDelay) * backoffMultiplier * jitter)

		fmt.Printf("%s attempt %d failed: %v. Retrying in %v...\n", name, attempt+1, err, sleepDuration)

		// Sleep with context awareness
		select {
		case <-time.After(sleepDuration):
			// Continue to next attempt
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}

	return "", lastErr
}

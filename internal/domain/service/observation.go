package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"garden3/internal/domain/entity"
	"garden3/internal/port/output"
)

// ObservationService implements the ObservationUseCase interface
type ObservationService struct {
	repo output.ObservationRepository
}

// NewObservationService creates a new observation service
func NewObservationService(repo output.ObservationRepository) *ObservationService {
	return &ObservationService{
		repo: repo,
	}
}

func (s *ObservationService) StoreFeedback(ctx context.Context, input entity.StoreFeedbackInput) (*entity.Observation, error) {
	// Create feedback data structure
	feedbackData := entity.FeedbackData{
		Question:     input.Question,
		Answer:       input.Answer,
		BookmarkID:   input.BookmarkID.String(),
		UserQuestion: input.UserQuestion,
		Similarity:   input.Similarity,
		FeedbackType: input.FeedbackType,
		Timestamp:    time.Now(),
	}

	// Marshal to JSON
	data, err := json.Marshal(feedbackData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal feedback data: %w", err)
	}

	// Store observation
	obsType := "qa-feedback"
	source := "user-feedback"
	tags := fmt.Sprintf("feedback,%s", input.FeedbackType)

	observation, err := s.repo.Create(ctx, data, obsType, source, tags, input.BookmarkID)
	if err != nil {
		return nil, fmt.Errorf("failed to create observation: %w", err)
	}

	// Delete content reference if requested
	if input.DeleteRef && input.ReferenceID != nil {
		if err := s.repo.DeleteBookmarkContentReference(ctx, *input.ReferenceID, input.BookmarkID); err != nil {
			// Log the error but don't fail the operation
			fmt.Printf("warning: failed to delete bookmark content reference: %v\n", err)
		}
	}

	return observation, nil
}

func (s *ObservationService) GetFeedbackStats(ctx context.Context, bookmarkID uuid.UUID) (*entity.FeedbackStats, error) {
	stats, err := s.repo.GetFeedbackStats(ctx, bookmarkID)
	if err != nil {
		return nil, fmt.Errorf("failed to get feedback stats: %w", err)
	}
	return stats, nil
}

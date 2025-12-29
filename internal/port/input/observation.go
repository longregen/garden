package input

import (
	"context"

	"github.com/google/uuid"
	"garden3/internal/domain/entity"
)

// ObservationUseCase defines the business operations for observations
type ObservationUseCase interface {
	// StoreFeedback stores Q&A feedback and optionally deletes the content reference
	StoreFeedback(ctx context.Context, input entity.StoreFeedbackInput) (*entity.Observation, error)

	// GetFeedbackStats retrieves feedback statistics for a bookmark
	GetFeedbackStats(ctx context.Context, bookmarkID uuid.UUID) (*entity.FeedbackStats, error)
}

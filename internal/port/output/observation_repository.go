package output

import (
	"context"

	"github.com/google/uuid"
	"garden3/internal/domain/entity"
)

// ObservationRepository defines persistence operations for observations
type ObservationRepository interface {
	// Create creates a new observation
	Create(ctx context.Context, data []byte, obsType, source, tags string, ref uuid.UUID) (*entity.Observation, error)

	// GetFeedbackStats retrieves feedback statistics for a bookmark
	GetFeedbackStats(ctx context.Context, bookmarkID uuid.UUID) (*entity.FeedbackStats, error)

	// DeleteBookmarkContentReference deletes a bookmark content reference
	DeleteBookmarkContentReference(ctx context.Context, referenceID, bookmarkID uuid.UUID) error
}

package input

import (
	"context"

	"github.com/google/uuid"
	"garden3/internal/domain/entity"
)

// SessionUseCase defines the business operations for sessions
type SessionUseCase interface {
	// SearchSessions performs semantic search using embeddings, falls back to text search
	SearchSessions(ctx context.Context, query string, limit int32) ([]entity.SessionSearchResult, error)

	// GetRoomSessions retrieves all session summaries for a room
	GetRoomSessions(ctx context.Context, roomID uuid.UUID) ([]entity.SessionSummary, error)

	// SearchContactSessions searches sessions where a contact participated
	SearchContactSessions(ctx context.Context, contactID uuid.UUID, searchTerm string) ([]entity.SessionSearchResult, error)

	// GetTimeline retrieves timeline visualization data for a room
	GetTimeline(ctx context.Context, roomID uuid.UUID) (*entity.TimelineData, error)

	// GetSessionMessages retrieves all messages in a session with contact info
	GetSessionMessages(ctx context.Context, sessionID uuid.UUID) (*entity.SessionMessagesResponse, error)
}

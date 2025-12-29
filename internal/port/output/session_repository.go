package output

import (
	"context"
	"time"

	"github.com/google/uuid"
	"garden3/internal/domain/entity"
)

// SessionRepository defines the data access operations for sessions
type SessionRepository interface {
	// SearchSessionsWithEmbeddings searches sessions using vector similarity
	SearchSessionsWithEmbeddings(ctx context.Context, embedding []float32, limit int32) ([]entity.SessionSearchResult, error)

	// SearchSessionsWithText searches sessions using text pattern matching
	SearchSessionsWithText(ctx context.Context, searchPattern string, limit int32) ([]entity.SessionSearchResult, error)

	// GetSessionSummaries retrieves all session summaries for a room
	GetSessionSummaries(ctx context.Context, roomID uuid.UUID) ([]entity.SessionSummary, error)

	// SearchContactSessionSummaries searches sessions where a contact participated
	SearchContactSessionSummaries(ctx context.Context, contactID uuid.UUID, searchPattern string) ([]entity.SessionSearchResult, error)

	// GetSessionsForRoom retrieves all sessions with message counts for a room
	GetSessionsForRoom(ctx context.Context, roomID uuid.UUID) ([]entity.SessionWithMessages, error)

	// GetSessionParticipantActivity retrieves participant activity for a room's sessions
	GetSessionParticipantActivity(ctx context.Context, roomID uuid.UUID) ([]entity.ParticipantActivity, error)

	// GetSessionMessages retrieves all messages in a session
	GetSessionMessages(ctx context.Context, sessionID uuid.UUID) ([]entity.SessionMessage, error)

	// GetContactsByIDs retrieves contacts by their IDs
	GetContactsByIDs(ctx context.Context, contactIDs []uuid.UUID) ([]entity.SessionMessageContact, error)

	// DeleteStaleConversations removes sessions older than the specified age with no messages
	DeleteStaleConversations(ctx context.Context, olderThan time.Time) (int64, error)
}

// EmbeddingService defines operations for generating embeddings
type EmbeddingService interface {
	// GetEmbedding generates an embedding vector for the given text
	GetEmbedding(ctx context.Context, text string) ([]float32, error)
}

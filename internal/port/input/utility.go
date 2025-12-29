package input

import (
	"context"

	"github.com/google/uuid"

	"garden3/internal/domain/entity"
)

// UtilityUseCase defines utility operations
type UtilityUseCase interface {
	// GetDebugInfo returns system debug information
	GetDebugInfo(ctx context.Context) (*entity.DebugInfo, error)

	// CleanupStaleConversations removes sessions older than 30 days with no messages
	CleanupStaleConversations(ctx context.Context) (int64, error)

	// GetMessagesByIDs retrieves messages by their IDs
	GetMessagesByIDs(ctx context.Context, messageIDs []uuid.UUID) ([]entity.Message, error)
}

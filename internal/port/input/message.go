package input

import (
	"context"
	"time"

	"garden3/internal/domain/entity"
	"github.com/google/uuid"
)

// MessageUseCase defines the business operations for messages
type MessageUseCase interface {
	// GetMessage retrieves a single message by ID
	GetMessage(ctx context.Context, messageID uuid.UUID) (*entity.Message, error)

	// GetMessagesByRoomID retrieves messages for a specific room with pagination
	GetMessagesByRoomID(ctx context.Context, roomID uuid.UUID, page, pageSize int32, beforeDatetime *time.Time) (*PaginatedResponse[entity.RoomMessage], error)

	// GetAllMessageContents retrieves all message body content as array of strings
	GetAllMessageContents(ctx context.Context) ([]string, error)

	// GetMessageTextRepresentations retrieves text representations for a message
	GetMessageTextRepresentations(ctx context.Context, messageID uuid.UUID) ([]entity.MessageTextRepresentation, error)

	// SearchMessages performs full-text search across all messages
	SearchMessages(ctx context.Context, query string, page, pageSize int32) (*PaginatedResponse[entity.RoomMessage], error)
}

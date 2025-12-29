package output

import (
	"context"
	"time"

	"github.com/google/uuid"
	"garden3/internal/domain/entity"
)

// MessageRepository defines the data access operations for messages
type MessageRepository interface {
	// GetMessage retrieves a single message by ID
	GetMessage(ctx context.Context, messageID uuid.UUID) (*entity.Message, error)

	// GetMessagesByRoomID retrieves messages for a specific room with pagination
	GetMessagesByRoomID(ctx context.Context, roomID uuid.UUID, limit, offset int32, beforeDatetime *time.Time) ([]entity.RoomMessage, error)

	// GetAllMessageContents retrieves all message body content as array of strings
	GetAllMessageContents(ctx context.Context) ([]string, error)

	// GetMessageTextRepresentations retrieves text representations for a message
	GetMessageTextRepresentations(ctx context.Context, messageID uuid.UUID) ([]entity.MessageTextRepresentation, error)

	// SearchMessages performs full-text search across all messages
	SearchMessages(ctx context.Context, query string, limit, offset int32) ([]entity.RoomMessage, error)

	// GetMessageContacts retrieves contacts by their IDs
	GetMessageContacts(ctx context.Context, contactIDs []uuid.UUID) ([]entity.Contact, error)

	// GetMessagesByIDs retrieves multiple messages by their IDs
	GetMessagesByIDs(ctx context.Context, messageIDs []uuid.UUID) ([]entity.Message, error)
}

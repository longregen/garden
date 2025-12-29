package input

import (
	"context"

	"github.com/google/uuid"
	"garden3/internal/domain/entity"
)

// RoomUseCase defines the business operations for rooms
type RoomUseCase interface {
	// ListRooms retrieves paginated rooms with optional search
	ListRooms(ctx context.Context, page, pageSize int32, searchText *string) (*PaginatedResponse[entity.Room], error)

	// GetRoomDetails retrieves a room with participants and recent messages
	GetRoomDetails(ctx context.Context, roomID uuid.UUID) (*entity.RoomDetails, error)

	// GetRoomMessages retrieves paginated messages for a room
	GetRoomMessages(ctx context.Context, roomID uuid.UUID, page, pageSize int32, beforeMessageID *uuid.UUID) (*entity.MessagesWithContacts, error)

	// SearchRoomMessages searches messages within a specific room
	SearchRoomMessages(ctx context.Context, roomID uuid.UUID, searchText string, page, pageSize int32) (*entity.MessagesWithContacts, error)

	// SetRoomName updates the user-defined name for a room
	SetRoomName(ctx context.Context, roomID uuid.UUID, name *string) error

	// GetSessionsCount retrieves the count of sessions in a room
	GetSessionsCount(ctx context.Context, roomID uuid.UUID) (int32, error)
}

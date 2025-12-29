package output

import (
	"context"
	"time"

	"github.com/google/uuid"
	"garden3/internal/domain/entity"
)

// RoomRepository defines the data access operations for rooms
type RoomRepository interface {
	// ListRooms retrieves a paginated list of rooms with optional search
	ListRooms(ctx context.Context, limit, offset int32, searchText *string) ([]entity.Room, error)

	// CountRooms returns the total count of rooms matching search criteria
	CountRooms(ctx context.Context, searchText *string) (int64, error)

	// GetRoom retrieves a single room by ID
	GetRoom(ctx context.Context, roomID uuid.UUID) (*entity.Room, error)

	// GetRoomParticipants retrieves all participants for a room
	GetRoomParticipants(ctx context.Context, roomID uuid.UUID) ([]entity.RoomParticipant, error)

	// GetRoomMessages retrieves paginated messages for a room
	GetRoomMessages(ctx context.Context, roomID uuid.UUID, limit, offset int32, beforeDatetime *time.Time) ([]entity.RoomMessage, error)

	// SearchRoomMessages searches messages within a room
	SearchRoomMessages(ctx context.Context, roomID uuid.UUID, searchText string, limit, offset int32) ([]entity.RoomMessage, error)

	// UpdateRoomName updates the user-defined name of a room
	UpdateRoomName(ctx context.Context, roomID uuid.UUID, name *string) error

	// GetSessionsCount retrieves the count of sessions in a room
	GetSessionsCount(ctx context.Context, roomID uuid.UUID) (int32, error)

	// GetMessageContacts retrieves contacts by their IDs
	GetMessageContacts(ctx context.Context, contactIDs []uuid.UUID) ([]entity.Contact, error)

	// GetBeforeMessageDatetime retrieves the datetime of a message
	GetBeforeMessageDatetime(ctx context.Context, messageID uuid.UUID) (*time.Time, error)

	// GetContextMessages retrieves N messages before and after each matched message
	GetContextMessages(ctx context.Context, roomID uuid.UUID, messageIDs []uuid.UUID, beforeCount, afterCount int32) ([]entity.RoomMessage, error)
}

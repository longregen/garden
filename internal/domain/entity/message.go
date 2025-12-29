package entity

import (
	"time"

	"github.com/google/uuid"
)

// Message represents a single message in the system
type Message struct {
	MessageID       uuid.UUID
	SenderContactID uuid.UUID
	RoomID          uuid.UUID
	EventID         string
	EventDatetime   time.Time
	Body            *string
	FormattedBody   *string
	MessageType     *string
	SenderName      *string
	SenderEmail     *string
}

// MessageSearchParams represents search parameters for messages
type MessageSearchParams struct {
	Query      string
	RoomID     *uuid.UUID
	Page       int32
	PageSize   int32
	BeforeDatetime *time.Time
}

// MessageTextRepresentation represents the searchable text of a message
type MessageTextRepresentation struct {
	ID           uuid.UUID
	MessageID    uuid.UUID
	TextContent  string
	SourceType   string
	SearchVector string
	CreatedAt    time.Time
}

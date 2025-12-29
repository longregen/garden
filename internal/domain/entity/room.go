package entity

import (
	"time"

	"github.com/google/uuid"
)

// Room represents a chat room or group
type Room struct {
	RoomID          uuid.UUID
	DisplayName     *string
	UserDefinedName *string
	SourceID        string
	LastActivity    *time.Time
	LastMessageTime *time.Time
	ParticipantCount int32
}

// RoomParticipant represents a participant in a room
type RoomParticipant struct {
	RoomID            uuid.UUID
	ContactID         uuid.UUID
	KnownLastPresence *time.Time
	KnownLastExit     *time.Time
	ContactName       string
	ContactEmail      *string
}

// RoomDetails represents a room with its participants
type RoomDetails struct {
	Room         Room
	Participants []RoomParticipant
	Messages     []RoomMessage
	Contacts     map[uuid.UUID]Contact
}

// RoomMessage represents an enriched message with related data
type RoomMessage struct {
	MessageID            uuid.UUID              `json:"message_id"`
	SenderContactID      uuid.UUID              `json:"sender_contact_id"`
	RoomID               uuid.UUID              `json:"room_id"`
	EventID              string                 `json:"event_id"`
	EventDatetime        time.Time              `json:"event_datetime"`
	Body                 *string                `json:"body"`
	FormattedBody        *string                `json:"formatted_body"`
	MessageType          *string                `json:"message_type"`
	MessageClassification *string                `json:"message_classification"`
	TranscriptionData    map[string]interface{} `json:"transcription_data"` // jsonb data
	BookmarkID           *uuid.UUID             `json:"bookmark_id,omitempty"`
	BookmarkURL          *string                `json:"url,omitempty"`
	BookmarkTitle        *string                `json:"bookmark_title,omitempty"`
	BookmarkSummary      *string                `json:"bookmark_summary,omitempty"`
}

// SearchResult represents a message search result with metadata
type SearchResult struct {
	Message RoomMessage
	Rank    float64
	Snippet string
}

// MessagesWithContacts bundles messages with their sender contact information
type MessagesWithContacts struct {
	Messages        []RoomMessage         `json:"messages"`
	Contacts        map[uuid.UUID]Contact `json:"contacts"`
	ContextMessages []RoomMessage         `json:"contextMessages,omitempty"`
}

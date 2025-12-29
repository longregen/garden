package entity

import (
	"time"

	"github.com/google/uuid"
)

// Session represents a conversation grouping
type Session struct {
	SessionID     uuid.UUID
	RoomID        uuid.UUID
	FirstDateTime time.Time
	FirstMessageID *uuid.UUID
	LastMessageID  *uuid.UUID
	LastDateTime   *time.Time
	CreatedAt      time.Time
	UpdatedAt      *time.Time
}

// SessionSummary represents a session with its summary
type SessionSummary struct {
	SessionID     uuid.UUID
	FirstDateTime time.Time
	LastDateTime  *time.Time
	Summary       *string
}

// SessionSearchResult represents a session found via search with similarity score
type SessionSearchResult struct {
	SessionID       uuid.UUID
	RoomID          uuid.UUID
	FirstDateTime   time.Time
	LastDateTime    *time.Time
	Summary         *string
	DisplayName     *string
	UserDefinedName *string
	Similarity      *float64 // Only populated for embedding search
}

// SessionMessage represents a message in a session with contact info
type SessionMessage struct {
	MessageID            uuid.UUID
	SenderContactID      uuid.UUID
	Body                 *string
	FormattedBody        *string
	MessageType          string
	MessageClassification *string
	EventDateTime        time.Time
	TranscriptionData    *string
}

// SessionMessageContact represents contact info for message senders
type SessionMessageContact struct {
	ContactID uuid.UUID
	Name      string
}

// SessionMessagesResponse combines messages with contact information
type SessionMessagesResponse struct {
	Messages []SessionMessage
	Contacts map[uuid.UUID]SessionMessageContact
}

// ParticipantActivity represents a participant's activity in a session
type ParticipantActivity struct {
	SenderContactID uuid.UUID
	SenderName      string
	SessionID       uuid.UUID
	FirstDateTime   time.Time
	MessageCount    int64
}

// SessionWithMessages represents a session with message count
type SessionWithMessages struct {
	SessionID     uuid.UUID
	FirstDateTime time.Time
	LastDateTime  *time.Time
	MessageCount  int64
}

// TimelineMonth represents aggregated session data for a month
type TimelineMonth struct {
	Month           string                  // Format: YYYY-MM
	Sessions        []TimelineSession
	TotalSessions   int
	TotalMessages   int64
	AverageDuration float64 // in minutes
}

// TimelineSession represents a session with participant data
type TimelineSession struct {
	SessionID     uuid.UUID
	FirstDateTime time.Time
	LastDateTime  *time.Time
	MessageCount  int64
	Duration      int64 // in minutes
	Participants  []TimelineParticipant
}

// TimelineParticipant represents a participant in a timeline session
type TimelineParticipant struct {
	ID           uuid.UUID
	Name         string
	MessageCount int64
}

// TimelineData represents complete timeline visualization data
type TimelineData struct {
	TimelineData       []TimelineMonth
	TotalSessions      int
	FirstSessionDate   *time.Time
	LastSessionDate    *time.Time
}

package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
	db "garden3/internal/adapter/secondary/postgres/generated/db"
	"garden3/internal/domain/entity"
)

// SessionRepository implements the output.SessionRepository interface
type SessionRepository struct {
	pool *pgxpool.Pool
}

// NewSessionRepository creates a new session repository
func NewSessionRepository(pool *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{
		pool: pool,
	}
}

// SearchSessionsWithEmbeddings searches sessions using vector similarity
func (r *SessionRepository) SearchSessionsWithEmbeddings(ctx context.Context, embedding []float32, limit int32) ([]entity.SessionSearchResult, error) {
	queries := db.New(r.pool)

	// Convert []float32 to pgvector.Vector
	vec := pgvector.NewVector(embedding)

	results, err := queries.SearchSessionsWithEmbeddings(ctx, db.SearchSessionsWithEmbeddingsParams{
		Column1: &vec,
		Limit:   limit,
	})
	if err != nil {
		return nil, err
	}

	searchResults := make([]entity.SessionSearchResult, len(results))
	for i, r := range results {
		similarity := float64(r.Similarity)
		searchResults[i] = entity.SessionSearchResult{
			SessionID:       r.SessionID,
			RoomID:          r.RoomID,
			FirstDateTime:   r.FirstDateTime.Time,
			LastDateTime:    pgTimestampToTimePtr(r.LastDateTime),
			Summary:         r.Summary,
			DisplayName:     r.DisplayName,
			UserDefinedName: r.UserDefinedName,
			Similarity:      &similarity,
		}
	}
	return searchResults, nil
}

// SearchSessionsWithText searches sessions using text pattern matching
func (r *SessionRepository) SearchSessionsWithText(ctx context.Context, searchPattern string, limit int32) ([]entity.SessionSearchResult, error) {
	queries := db.New(r.pool)

	results, err := queries.SearchSessionsWithText(ctx, db.SearchSessionsWithTextParams{
		Summary: &searchPattern,
		Limit:   limit,
	})
	if err != nil {
		return nil, err
	}

	searchResults := make([]entity.SessionSearchResult, len(results))
	for i, r := range results {
		searchResults[i] = entity.SessionSearchResult{
			SessionID:       r.SessionID,
			RoomID:          r.RoomID,
			FirstDateTime:   r.FirstDateTime.Time,
			LastDateTime:    pgTimestampToTimePtr(r.LastDateTime),
			Summary:         r.Summary,
			DisplayName:     r.DisplayName,
			UserDefinedName: r.UserDefinedName,
			Similarity:      nil,
		}
	}
	return searchResults, nil
}

// GetSessionSummaries retrieves all session summaries for a room
func (r *SessionRepository) GetSessionSummaries(ctx context.Context, roomID uuid.UUID) ([]entity.SessionSummary, error) {
	queries := db.New(r.pool)

	results, err := queries.GetSessionSummaries(ctx, roomID)
	if err != nil {
		return nil, err
	}

	summaries := make([]entity.SessionSummary, len(results))
	for i, r := range results {
		summaries[i] = entity.SessionSummary{
			SessionID:     r.SessionID,
			FirstDateTime: r.FirstDateTime.Time,
			LastDateTime:  pgTimestampToTimePtr(r.LastDateTime),
			Summary:       r.Summary,
		}
	}
	return summaries, nil
}

// SearchContactSessionSummaries searches sessions where a contact participated
func (r *SessionRepository) SearchContactSessionSummaries(ctx context.Context, contactID uuid.UUID, searchPattern string) ([]entity.SessionSearchResult, error) {
	queries := db.New(r.pool)

	results, err := queries.SearchContactSessionSummaries(ctx, db.SearchContactSessionSummariesParams{
		ContactID: contactID,
		Summary:   &searchPattern,
	})
	if err != nil {
		return nil, err
	}

	searchResults := make([]entity.SessionSearchResult, len(results))
	for i, r := range results {
		searchResults[i] = entity.SessionSearchResult{
			SessionID:    pgUUIDToUUID(r.SessionID),
			RoomID:       pgUUIDToUUID(r.RoomID),
			FirstDateTime: r.FirstDateTime.Time,
			LastDateTime: pgTimestampToTimePtr(r.LastDateTime),
			Summary:      r.Summary,
			Similarity:   nil,
		}
	}
	return searchResults, nil
}

// GetSessionsForRoom retrieves all sessions with message counts for a room
func (r *SessionRepository) GetSessionsForRoom(ctx context.Context, roomID uuid.UUID) ([]entity.SessionWithMessages, error) {
	queries := db.New(r.pool)

	results, err := queries.GetSessionsForRoom(ctx, roomID)
	if err != nil {
		return nil, err
	}

	sessions := make([]entity.SessionWithMessages, len(results))
	for i, r := range results {
		sessions[i] = entity.SessionWithMessages{
			SessionID:     r.SessionID,
			FirstDateTime: r.FirstDateTime.Time,
			LastDateTime:  pgTimestampToTimePtr(r.LastDateTime),
			MessageCount:  r.MessageCount,
		}
	}
	return sessions, nil
}

// GetSessionParticipantActivity retrieves participant activity for a room's sessions
func (r *SessionRepository) GetSessionParticipantActivity(ctx context.Context, roomID uuid.UUID) ([]entity.ParticipantActivity, error) {
	queries := db.New(r.pool)

	results, err := queries.GetSessionParticipantActivity(ctx, roomID)
	if err != nil {
		return nil, err
	}

	activities := make([]entity.ParticipantActivity, len(results))
	for i, r := range results {
		activities[i] = entity.ParticipantActivity{
			SenderContactID: r.SenderContactID,
			SenderName:      r.SenderName,
			SessionID:       r.SessionID,
			FirstDateTime:   r.FirstDateTime.Time,
			MessageCount:    r.MessageCount,
		}
	}
	return activities, nil
}

// GetSessionMessages retrieves all messages in a session
func (r *SessionRepository) GetSessionMessages(ctx context.Context, sessionID uuid.UUID) ([]entity.SessionMessage, error) {
	queries := db.New(r.pool)

	results, err := queries.GetSessionMessages(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	messages := make([]entity.SessionMessage, len(results))
	for i, r := range results {
		var transcriptionData *string
		if r.TranscriptionData != nil {
			str := string(r.TranscriptionData)
			transcriptionData = &str
		}

		messageType := ""
		if r.MessageType != nil {
			messageType = *r.MessageType
		}

		messages[i] = entity.SessionMessage{
			MessageID:             r.MessageID,
			SenderContactID:       r.SenderContactID,
			Body:                  r.Body,
			FormattedBody:         r.FormattedBody,
			MessageType:           messageType,
			MessageClassification: r.MessageClassification,
			EventDateTime:         r.EventDatetime.Time,
			TranscriptionData:     transcriptionData,
		}
	}
	return messages, nil
}

// GetContactsByIDs retrieves contacts by their IDs
func (r *SessionRepository) GetContactsByIDs(ctx context.Context, contactIDs []uuid.UUID) ([]entity.SessionMessageContact, error) {
	queries := db.New(r.pool)

	results, err := queries.GetContactsByIds(ctx, contactIDs)
	if err != nil {
		return nil, err
	}

	contacts := make([]entity.SessionMessageContact, len(results))
	for i, r := range results {
		contacts[i] = entity.SessionMessageContact{
			ContactID: r.ContactID,
			Name:      r.Name,
		}
	}
	return contacts, nil
}

// DeleteStaleConversations removes sessions older than the specified age with no messages
func (r *SessionRepository) DeleteStaleConversations(ctx context.Context, olderThan time.Time) (int64, error) {
	query := `
		DELETE FROM sessions
		WHERE first_datetime < $1
		AND session_id NOT IN (
			SELECT DISTINCT session_id
			FROM messages
			WHERE session_id IS NOT NULL
		)
	`

	result, err := r.pool.Exec(ctx, query, olderThan)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected(), nil
}

// Helper functions

func pgTimestampToTimePtr(ts pgtype.Timestamp) *time.Time {
	if !ts.Valid {
		return nil
	}
	return &ts.Time
}

func pgUUIDToUUID(u pgtype.UUID) uuid.UUID {
	if !u.Valid {
		return uuid.Nil
	}
	return u.Bytes
}

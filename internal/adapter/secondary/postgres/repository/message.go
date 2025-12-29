package repository

import (
	"context"
	"encoding/json"
	"time"

	db "garden3/internal/adapter/secondary/postgres/generated/db"
	"garden3/internal/domain/entity"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// MessageRepository implements the output.MessageRepository interface
type MessageRepository struct {
	pool *pgxpool.Pool
}

// NewMessageRepository creates a new message repository
func NewMessageRepository(pool *pgxpool.Pool) *MessageRepository {
	return &MessageRepository{
		pool: pool,
	}
}

func (r *MessageRepository) GetMessage(ctx context.Context, messageID uuid.UUID) (*entity.Message, error) {
	queries := db.New(r.pool)
	dbMessage, err := queries.GetMessage(ctx, messageID)
	if err != nil {
		return nil, err
	}

	return &entity.Message{
		MessageID:       dbMessage.MessageID,
		SenderContactID: dbMessage.SenderContactID,
		RoomID:          dbMessage.RoomID,
		EventID:         dbMessage.EventID,
		EventDatetime:   dbMessage.EventDatetime.Time,
		Body:            dbMessage.Body,
		FormattedBody:   dbMessage.FormattedBody,
		MessageType:     dbMessage.MessageType,
		SenderName:      dbMessage.SenderName,
		SenderEmail:     dbMessage.SenderEmail,
	}, nil
}

func (r *MessageRepository) SearchMessages(ctx context.Context, query string, limit, offset int32) ([]entity.RoomMessage, error) {
	queries := db.New(r.pool)
	dbMessages, err := queries.SearchMessages(ctx, db.SearchMessagesParams{
		ToTsquery: query,
		Limit:     limit,
		Offset:    offset,
	})
	if err != nil {
		return nil, err
	}

	messages := make([]entity.RoomMessage, len(dbMessages))
	for i, dbMsg := range dbMessages {
		messages[i] = entity.RoomMessage{
			MessageID:       dbMsg.MessageID,
			SenderContactID: dbMsg.SenderContactID,
			RoomID:          dbMsg.RoomID,
			EventID:         dbMsg.EventID,
			EventDatetime:   dbMsg.EventDatetime.Time,
			Body:            dbMsg.Body,
			FormattedBody:   dbMsg.FormattedBody,
			MessageType:     dbMsg.MessageType,
		}
	}
	return messages, nil
}

func (r *MessageRepository) GetMessageContacts(ctx context.Context, contactIDs []uuid.UUID) ([]entity.Contact, error) {
	queries := db.New(r.pool)
	dbContacts, err := queries.GetMessageContacts(ctx, contactIDs)
	if err != nil {
		return nil, err
	}

	contacts := make([]entity.Contact, len(dbContacts))
	for i, dbContact := range dbContacts {
		contacts[i] = entity.Contact{
			ContactID: dbContact.ContactID,
			Name:      dbContact.Name,
			Email:     dbContact.Email,
		}
	}
	return contacts, nil
}

func (r *MessageRepository) GetMessagesByIDs(ctx context.Context, messageIDs []uuid.UUID) ([]entity.Message, error) {
	if len(messageIDs) == 0 {
		return []entity.Message{}, nil
	}

	query := `
		SELECT
			m.message_id,
			m.sender_contact_id,
			m.room_id,
			m.event_id,
			m.event_datetime,
			m.body,
			m.formatted_body,
			m.message_type,
			c.name as sender_name,
			c.email as sender_email
		FROM messages m
		LEFT JOIN contacts c ON m.sender_contact_id = c.contact_id
		WHERE m.message_id = ANY($1)
		ORDER BY m.event_datetime
	`

	rows, err := r.pool.Query(ctx, query, messageIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []entity.Message
	for rows.Next() {
		var msg entity.Message
		var messageType *string

		err := rows.Scan(
			&msg.MessageID,
			&msg.SenderContactID,
			&msg.RoomID,
			&msg.EventID,
			&msg.EventDatetime,
			&msg.Body,
			&msg.FormattedBody,
			&messageType,
			&msg.SenderName,
			&msg.SenderEmail,
		)
		if err != nil {
			return nil, err
		}

		msg.MessageType = messageType
		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}

func (r *MessageRepository) GetMessagesByRoomID(ctx context.Context, roomID uuid.UUID, limit, offset int32, beforeDatetime *time.Time) ([]entity.RoomMessage, error) {
	queries := db.New(r.pool)

	var beforeTS pgtype.Timestamp
	if beforeDatetime != nil {
		beforeTS = pgtype.Timestamp{Time: *beforeDatetime, Valid: true}
	}

	dbMessages, err := queries.GetRoomMessages(ctx, db.GetRoomMessagesParams{
		RoomID:         roomID,
		Limit:          limit,
		Offset:         offset,
		BeforeDatetime: beforeTS,
	})
	if err != nil {
		return nil, err
	}

	messages := make([]entity.RoomMessage, len(dbMessages))
	for i, dbMsg := range dbMessages {
		var transcriptionData map[string]interface{}
		if len(dbMsg.TranscriptionData) > 0 {
			if err := json.Unmarshal(dbMsg.TranscriptionData, &transcriptionData); err != nil {
				transcriptionData = nil
			}
		}

		var bookmarkID *uuid.UUID
		if dbMsg.BookmarkID.Valid {
			id, err := uuid.FromBytes(dbMsg.BookmarkID.Bytes[:])
			if err == nil {
				bookmarkID = &id
			}
		}

		messages[i] = entity.RoomMessage{
			MessageID:             dbMsg.MessageID,
			SenderContactID:       dbMsg.SenderContactID,
			RoomID:                roomID,
			EventID:               dbMsg.EventID,
			EventDatetime:         dbMsg.EventDatetime.Time,
			Body:                  dbMsg.Body,
			FormattedBody:         dbMsg.FormattedBody,
			MessageType:           dbMsg.MessageType,
			MessageClassification: dbMsg.MessageClassification,
			TranscriptionData:     transcriptionData,
			BookmarkID:            bookmarkID,
			BookmarkURL:           dbMsg.BookmarkUrl,
			BookmarkTitle:         dbMsg.BookmarkTitle,
			BookmarkSummary:       dbMsg.BookmarkSummary,
		}
	}

	return messages, nil
}

func (r *MessageRepository) GetAllMessageContents(ctx context.Context) ([]string, error) {
	queries := db.New(r.pool)
	contents, err := queries.GetAllMessageContents(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]string, len(contents))
	for i, content := range contents {
		result[i] = string(content)
	}
	return result, nil
}

func (r *MessageRepository) GetMessageTextRepresentations(ctx context.Context, messageID uuid.UUID) ([]entity.MessageTextRepresentation, error) {
	queries := db.New(r.pool)
	dbReps, err := queries.GetMessageTextRepresentations(ctx, messageID)
	if err != nil {
		return nil, err
	}

	reps := make([]entity.MessageTextRepresentation, len(dbReps))
	for i, dbRep := range dbReps {
		reps[i] = entity.MessageTextRepresentation{
			ID:          dbRep.ID,
			MessageID:   dbRep.MessageID,
			TextContent: dbRep.TextContent,
			SourceType:  dbRep.SourceType,
			CreatedAt:   dbRep.CreatedAt.Time,
		}
	}

	return reps, nil
}

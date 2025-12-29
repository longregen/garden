package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	db "garden3/internal/adapter/secondary/postgres/generated/db"
	"garden3/internal/domain/entity"
)

// RoomRepository implements the output.RoomRepository interface
type RoomRepository struct {
	pool *pgxpool.Pool
}

// NewRoomRepository creates a new room repository
func NewRoomRepository(pool *pgxpool.Pool) *RoomRepository {
	return &RoomRepository{
		pool: pool,
	}
}

func (r *RoomRepository) ListRooms(ctx context.Context, limit, offset int32, searchText *string) ([]entity.Room, error) {
	queries := db.New(r.pool)

	var dbRooms []db.ListRoomsRow
	var err error

	if searchText != nil {
		dbRooms, err = queries.ListRooms(ctx, db.ListRoomsParams{
			Limit:      limit,
			Offset:     offset,
			SearchText: searchText,
		})
	} else {
		dbRooms, err = queries.ListRooms(ctx, db.ListRoomsParams{
			Limit:  limit,
			Offset: offset,
		})
	}

	if err != nil {
		return nil, err
	}

	rooms := make([]entity.Room, len(dbRooms))
	for i, dbRoom := range dbRooms {
		rooms[i] = entity.Room{
			RoomID:           dbRoom.RoomID,
			DisplayName:      dbRoom.DisplayName,
			UserDefinedName:  dbRoom.UserDefinedName,
			LastActivity:     toTimePtr(dbRoom.LastActivity),
			LastMessageTime:  convertInterfaceToTimePtr(dbRoom.LastMessageTime),
			ParticipantCount: dbRoom.ParticipantCount,
		}
	}
	return rooms, nil
}

func (r *RoomRepository) CountRooms(ctx context.Context, searchText *string) (int64, error) {
	queries := db.New(r.pool)

	var count int64
	var err error

	if searchText != nil {
		count, err = queries.CountRooms(ctx, searchText)
	} else {
		count, err = queries.CountRooms(ctx, nil)
	}

	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *RoomRepository) GetRoom(ctx context.Context, roomID uuid.UUID) (*entity.Room, error) {
	queries := db.New(r.pool)
	dbRoom, err := queries.GetRoom(ctx, roomID)
	if err != nil {
		return nil, err
	}

	return &entity.Room{
		RoomID:          dbRoom.RoomID,
		DisplayName:     dbRoom.DisplayName,
		UserDefinedName: dbRoom.UserDefinedName,
		SourceID:        dbRoom.SourceID,
		LastActivity:    toTimePtr(dbRoom.LastActivity),
		LastMessageTime: convertInterfaceToTimePtr(dbRoom.LastMessageTime),
	}, nil
}

func (r *RoomRepository) GetRoomParticipants(ctx context.Context, roomID uuid.UUID) ([]entity.RoomParticipant, error) {
	queries := db.New(r.pool)
	dbParticipants, err := queries.GetRoomParticipants(ctx, roomID)
	if err != nil {
		return nil, err
	}

	participants := make([]entity.RoomParticipant, len(dbParticipants))
	for i, dbParticipant := range dbParticipants {
		participants[i] = entity.RoomParticipant{
			RoomID:            dbParticipant.RoomID,
			ContactID:         dbParticipant.ContactID,
			KnownLastPresence: toTimePtr(dbParticipant.KnownLastPresence),
			KnownLastExit:     toTimePtr(dbParticipant.KnownLastExit),
			ContactName:       dbParticipant.ContactName,
			ContactEmail:      dbParticipant.ContactEmail,
		}
	}
	return participants, nil
}

func (r *RoomRepository) GetRoomMessages(ctx context.Context, roomID uuid.UUID, limit, offset int32, beforeDatetime *time.Time) ([]entity.RoomMessage, error) {
	queries := db.New(r.pool)

	params := db.GetRoomMessagesParams{
		RoomID: roomID,
		Limit:  limit,
		Offset: offset,
	}

	if beforeDatetime != nil {
		params.BeforeDatetime = *toPgTimestamp(*beforeDatetime)
	}

	dbMessages, err := queries.GetRoomMessages(ctx, params)
	if err != nil {
		return nil, err
	}

	messages := make([]entity.RoomMessage, len(dbMessages))
	for i, dbMsg := range dbMessages {
		var transcriptionData map[string]interface{}
		if dbMsg.TranscriptionData != nil {
			if err := json.Unmarshal(dbMsg.TranscriptionData, &transcriptionData); err == nil {
				// Successfully unmarshalled
			}
		}

		messages[i] = entity.RoomMessage{
			MessageID:            dbMsg.MessageID,
			SenderContactID:      dbMsg.SenderContactID,
			EventID:              dbMsg.EventID,
			EventDatetime:        dbMsg.EventDatetime.Time,
			Body:                 dbMsg.Body,
			FormattedBody:        dbMsg.FormattedBody,
			MessageType:          dbMsg.MessageType,
			MessageClassification: dbMsg.MessageClassification,
			TranscriptionData:    transcriptionData,
			BookmarkID:           toUUIDPtr(dbMsg.BookmarkID),
			BookmarkURL:          dbMsg.BookmarkUrl,
			BookmarkTitle:        dbMsg.BookmarkTitle,
			BookmarkSummary:      dbMsg.BookmarkSummary,
		}
	}
	return messages, nil
}

func (r *RoomRepository) SearchRoomMessages(ctx context.Context, roomID uuid.UUID, searchText string, limit, offset int32) ([]entity.RoomMessage, error) {
	queries := db.New(r.pool)

	dbMessages, err := queries.SearchRoomMessages(ctx, db.SearchRoomMessagesParams{
		RoomID:  roomID,
		Column2: &searchText,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		return nil, err
	}

	messages := make([]entity.RoomMessage, len(dbMessages))
	for i, dbMsg := range dbMessages {
		var transcriptionData map[string]interface{}
		if dbMsg.TranscriptionData != nil {
			json.Unmarshal(dbMsg.TranscriptionData, &transcriptionData)
		}

		messages[i] = entity.RoomMessage{
			MessageID:            dbMsg.MessageID,
			SenderContactID:      dbMsg.SenderContactID,
			EventID:              dbMsg.EventID,
			EventDatetime:        dbMsg.EventDatetime.Time,
			Body:                 dbMsg.Body,
			FormattedBody:        dbMsg.FormattedBody,
			MessageType:          dbMsg.MessageType,
			MessageClassification: dbMsg.MessageClassification,
			TranscriptionData:    transcriptionData,
			BookmarkID:           toUUIDPtr(dbMsg.BookmarkID),
			BookmarkURL:          dbMsg.BookmarkUrl,
			BookmarkTitle:        dbMsg.BookmarkTitle,
			BookmarkSummary:      dbMsg.BookmarkSummary,
		}
	}
	return messages, nil
}

func (r *RoomRepository) UpdateRoomName(ctx context.Context, roomID uuid.UUID, name *string) error {
	queries := db.New(r.pool)
	return queries.UpdateRoomName(ctx, db.UpdateRoomNameParams{
		RoomID:          roomID,
		UserDefinedName: name,
	})
}

func (r *RoomRepository) GetSessionsCount(ctx context.Context, roomID uuid.UUID) (int32, error) {
	queries := db.New(r.pool)
	count, err := queries.GetRoomSessionsCount(ctx, roomID)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *RoomRepository) GetMessageContacts(ctx context.Context, contactIDs []uuid.UUID) ([]entity.Contact, error) {
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

func (r *RoomRepository) GetBeforeMessageDatetime(ctx context.Context, messageID uuid.UUID) (*time.Time, error) {
	queries := db.New(r.pool)
	dt, err := queries.GetBeforeMessageDatetime(ctx, messageID)
	if err != nil {
		return nil, err
	}
	t := dt.Time
	return &t, nil
}

// Helper functions

func toTimePtr(t pgtype.Timestamp) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}

func toUUIDPtr(u pgtype.UUID) *uuid.UUID {
	if !u.Valid {
		return nil
	}
	id := uuid.UUID(u.Bytes)
	return &id
}

func toPgTimestamp(t time.Time) *pgtype.Timestamp {
	return &pgtype.Timestamp{
		Time:  t,
		Valid: true,
	}
}

func (r *RoomRepository) GetContextMessages(ctx context.Context, roomID uuid.UUID, messageIDs []uuid.UUID, beforeCount, afterCount int32) ([]entity.RoomMessage, error) {
	queries := db.New(r.pool)

	if len(messageIDs) == 0 {
		return []entity.RoomMessage{}, nil
	}

	dbMessages, err := queries.GetContextMessages(ctx, db.GetContextMessagesParams{
		RoomID:  roomID,
		Column2: messageIDs,
		Limit:   beforeCount,
		Limit_2: afterCount,
	})
	if err != nil {
		return nil, err
	}

	messages := make([]entity.RoomMessage, len(dbMessages))
	for i, dbMsg := range dbMessages {
		var transcriptionData map[string]interface{}
		if dbMsg.TranscriptionData != nil {
			json.Unmarshal(dbMsg.TranscriptionData, &transcriptionData)
		}

		messages[i] = entity.RoomMessage{
			MessageID:            dbMsg.MessageID,
			SenderContactID:      dbMsg.SenderContactID,
			EventID:              dbMsg.EventID,
			EventDatetime:        dbMsg.EventDatetime.Time,
			Body:                 dbMsg.Body,
			FormattedBody:        dbMsg.FormattedBody,
			MessageType:          dbMsg.MessageType,
			MessageClassification: dbMsg.MessageClassification,
			TranscriptionData:    transcriptionData,
			BookmarkID:           toUUIDPtr(dbMsg.BookmarkID),
			BookmarkURL:          dbMsg.BookmarkUrl,
			BookmarkTitle:        dbMsg.BookmarkTitle,
			BookmarkSummary:      dbMsg.BookmarkSummary,
		}
	}
	return messages, nil
}

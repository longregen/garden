package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"garden3/internal/domain/entity"
	"garden3/internal/port/input"
	"garden3/internal/port/output"
)

// RoomService implements the room use case
type RoomService struct {
	repo output.RoomRepository
}

// NewRoomService creates a new room service
func NewRoomService(repo output.RoomRepository) input.RoomUseCase {
	return &RoomService{repo: repo}
}

// ListRooms retrieves paginated rooms with optional search
func (s *RoomService) ListRooms(ctx context.Context, page, pageSize int32, searchText *string) (*input.PaginatedResponse[entity.Room], error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	rooms, err := s.repo.ListRooms(ctx, pageSize, offset, searchText)
	if err != nil {
		return nil, fmt.Errorf("failed to list rooms: %w", err)
	}

	total, err := s.repo.CountRooms(ctx, searchText)
	if err != nil {
		return nil, fmt.Errorf("failed to count rooms: %w", err)
	}

	totalPages := int32(total) / pageSize
	if int32(total)%pageSize > 0 {
		totalPages++
	}

	return &input.PaginatedResponse[entity.Room]{
		Data:       rooms,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// GetRoomDetails retrieves a room with participants and recent messages
func (s *RoomService) GetRoomDetails(ctx context.Context, roomID uuid.UUID) (*entity.RoomDetails, error) {
	room, err := s.repo.GetRoom(ctx, roomID)
	if err != nil {
		return nil, fmt.Errorf("failed to get room: %w", err)
	}
	if room == nil {
		return nil, nil
	}

	participants, err := s.repo.GetRoomParticipants(ctx, roomID)
	if err != nil {
		return nil, fmt.Errorf("failed to get room participants: %w", err)
	}

	messages, err := s.repo.GetRoomMessages(ctx, roomID, 50, 0, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get room messages: %w", err)
	}

	// Extract unique contact IDs from messages
	contactIDMap := make(map[uuid.UUID]bool)
	for _, msg := range messages {
		contactIDMap[msg.SenderContactID] = true
	}

	contactIDs := make([]uuid.UUID, 0, len(contactIDMap))
	for id := range contactIDMap {
		contactIDs = append(contactIDs, id)
	}

	var contacts []entity.Contact
	if len(contactIDs) > 0 {
		contacts, err = s.repo.GetMessageContacts(ctx, contactIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to get message contacts: %w", err)
		}
	}

	contactsMap := make(map[uuid.UUID]entity.Contact)
	for _, contact := range contacts {
		contactsMap[contact.ContactID] = contact
	}

	return &entity.RoomDetails{
		Room:         *room,
		Participants: participants,
		Messages:     messages,
		Contacts:     contactsMap,
	}, nil
}

// GetRoomMessages retrieves paginated messages for a room
func (s *RoomService) GetRoomMessages(ctx context.Context, roomID uuid.UUID, page, pageSize int32, beforeMessageID *uuid.UUID) (*entity.MessagesWithContacts, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	offset := (page - 1) * pageSize

	// Get before datetime if beforeMessageID is provided
	var beforeDatetime *time.Time
	if beforeMessageID != nil {
		dt, err := s.repo.GetBeforeMessageDatetime(ctx, *beforeMessageID)
		if err != nil {
			return nil, fmt.Errorf("failed to get before message datetime: %w", err)
		}
		beforeDatetime = dt
	}

	messages, err := s.repo.GetRoomMessages(ctx, roomID, pageSize, offset, beforeDatetime)
	if err != nil {
		return nil, fmt.Errorf("failed to get room messages: %w", err)
	}

	// Extract unique contact IDs
	contactIDMap := make(map[uuid.UUID]bool)
	for _, msg := range messages {
		contactIDMap[msg.SenderContactID] = true
	}

	contactIDs := make([]uuid.UUID, 0, len(contactIDMap))
	for id := range contactIDMap {
		contactIDs = append(contactIDs, id)
	}

	var contacts []entity.Contact
	if len(contactIDs) > 0 {
		contacts, err = s.repo.GetMessageContacts(ctx, contactIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to get message contacts: %w", err)
		}
	}

	contactsMap := make(map[uuid.UUID]entity.Contact)
	for _, contact := range contacts {
		contactsMap[contact.ContactID] = contact
	}

	return &entity.MessagesWithContacts{
		Messages: messages,
		Contacts: contactsMap,
	}, nil
}

// SearchRoomMessages searches messages within a specific room
func (s *RoomService) SearchRoomMessages(ctx context.Context, roomID uuid.UUID, searchText string, page, pageSize int32) (*entity.MessagesWithContacts, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	offset := (page - 1) * pageSize

	messages, err := s.repo.SearchRoomMessages(ctx, roomID, searchText, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search room messages: %w", err)
	}

	// Extract message IDs for context fetching
	messageIDs := make([]uuid.UUID, len(messages))
	for i, msg := range messages {
		messageIDs[i] = msg.MessageID
	}

	// Fetch context messages (3 before and 3 after each matched message)
	var contextMessages []entity.RoomMessage
	if len(messageIDs) > 0 {
		contextMessages, err = s.repo.GetContextMessages(ctx, roomID, messageIDs, 3, 3)
		if err != nil {
			return nil, fmt.Errorf("failed to get context messages: %w", err)
		}
	}

	// Extract unique contact IDs from both messages and context messages
	contactIDMap := make(map[uuid.UUID]bool)
	for _, msg := range messages {
		contactIDMap[msg.SenderContactID] = true
	}
	for _, msg := range contextMessages {
		contactIDMap[msg.SenderContactID] = true
	}

	contactIDs := make([]uuid.UUID, 0, len(contactIDMap))
	for id := range contactIDMap {
		contactIDs = append(contactIDs, id)
	}

	var contacts []entity.Contact
	if len(contactIDs) > 0 {
		contacts, err = s.repo.GetMessageContacts(ctx, contactIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to get message contacts: %w", err)
		}
	}

	contactsMap := make(map[uuid.UUID]entity.Contact)
	for _, contact := range contacts {
		contactsMap[contact.ContactID] = contact
	}

	return &entity.MessagesWithContacts{
		Messages:        messages,
		Contacts:        contactsMap,
		ContextMessages: contextMessages,
	}, nil
}

// SetRoomName updates the user-defined name for a room
func (s *RoomService) SetRoomName(ctx context.Context, roomID uuid.UUID, name *string) error {
	err := s.repo.UpdateRoomName(ctx, roomID, name)
	if err != nil {
		return fmt.Errorf("failed to update room name: %w", err)
	}
	return nil
}

// GetSessionsCount retrieves the count of sessions in a room
func (s *RoomService) GetSessionsCount(ctx context.Context, roomID uuid.UUID) (int32, error) {
	count, err := s.repo.GetSessionsCount(ctx, roomID)
	if err != nil {
		return 0, fmt.Errorf("failed to get sessions count: %w", err)
	}
	return count, nil
}

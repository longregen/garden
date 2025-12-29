package service

import (
	"context"
	"fmt"
	"time"

	"garden3/internal/domain/entity"
	"garden3/internal/port/input"
	"garden3/internal/port/output"
	"github.com/google/uuid"
)

// MessageService implements the message use case
type MessageService struct {
	repo output.MessageRepository
}

// NewMessageService creates a new message service
func NewMessageService(repo output.MessageRepository) input.MessageUseCase {
	return &MessageService{repo: repo}
}

// GetMessage retrieves a single message by ID
func (s *MessageService) GetMessage(ctx context.Context, messageID uuid.UUID) (*entity.Message, error) {
	message, err := s.repo.GetMessage(ctx, messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}
	return message, nil
}

// GetMessagesByRoomID retrieves messages for a specific room with pagination
func (s *MessageService) GetMessagesByRoomID(ctx context.Context, roomID uuid.UUID, page, pageSize int32, beforeDatetime *time.Time) (*input.PaginatedResponse[entity.RoomMessage], error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize

	messages, err := s.repo.GetMessagesByRoomID(ctx, roomID, pageSize, offset, beforeDatetime)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages by room ID: %w", err)
	}

	return &input.PaginatedResponse[entity.RoomMessage]{
		Data:       messages,
		Total:      -1,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: -1,
	}, nil
}

// GetAllMessageContents retrieves all message body content as array of strings
func (s *MessageService) GetAllMessageContents(ctx context.Context) ([]string, error) {
	contents, err := s.repo.GetAllMessageContents(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all message contents: %w", err)
	}
	return contents, nil
}

// GetMessageTextRepresentations retrieves text representations for a message
func (s *MessageService) GetMessageTextRepresentations(ctx context.Context, messageID uuid.UUID) ([]entity.MessageTextRepresentation, error) {
	reps, err := s.repo.GetMessageTextRepresentations(ctx, messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get message text representations: %w", err)
	}
	return reps, nil
}

// SearchMessages performs full-text search across all messages
func (s *MessageService) SearchMessages(ctx context.Context, query string, page, pageSize int32) (*input.PaginatedResponse[entity.RoomMessage], error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	offset := (page - 1) * pageSize

	messages, err := s.repo.SearchMessages(ctx, query, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search messages: %w", err)
	}

	// Note: Total count for FTS is not trivial to get efficiently
	// For now, we return -1 to indicate it's not calculated
	return &input.PaginatedResponse[entity.RoomMessage]{
		Data:       messages,
		Total:      -1,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: -1,
	}, nil
}

package input

import (
	"context"

	"github.com/google/uuid"
	"garden3/internal/domain/entity"
)

// NoteUseCase defines the business operations for notes
type NoteUseCase interface {
	// GetNote retrieves a single note with tags and processed content
	GetNote(ctx context.Context, noteID uuid.UUID) (*entity.FullNote, error)

	// ListNotes retrieves paginated notes with optional search
	ListNotes(ctx context.Context, page, pageSize int32, searchQuery *string) (*PaginatedResponse[entity.NoteListItem], error)

	// CreateNote creates a new note with tags and entity relationships
	CreateNote(ctx context.Context, input entity.CreateNoteInput) (*entity.FullNote, error)

	// UpdateNote updates a note's information and tags
	UpdateNote(ctx context.Context, noteID uuid.UUID, input entity.UpdateNoteInput) (*entity.FullNote, error)

	// DeleteNote deletes a note and all related data
	DeleteNote(ctx context.Context, noteID uuid.UUID) error

	// ListAllTags retrieves all tag names in the system
	ListAllTags(ctx context.Context) ([]entity.NoteTag, error)

	// SearchSimilarNotes performs vector similarity search
	SearchSimilarNotes(ctx context.Context, query string, strategy string) ([]entity.NoteListItem, error)
}

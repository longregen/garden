package output

import (
	"context"

	"github.com/google/uuid"
	"garden3/internal/domain/entity"
)

// NoteRepository defines the data access operations for notes
type NoteRepository interface {
	// Note operations
	GetNote(ctx context.Context, noteID uuid.UUID) (*entity.Note, error)
	GetNoteWithTags(ctx context.Context, noteID uuid.UUID) (*entity.Note, []string, error)
	ListNotes(ctx context.Context, limit, offset int32) ([]entity.NoteListItem, error)
	SearchNotes(ctx context.Context, searchPattern string, limit, offset int32) ([]entity.NoteListItem, error)
	CountNotes(ctx context.Context) (int64, error)
	CountSearchNotes(ctx context.Context, searchPattern string) (int64, error)
	CreateNote(ctx context.Context, title, slug, contents string) (*entity.Note, error)
	UpdateNote(ctx context.Context, noteID uuid.UUID, title, contents string) error
	DeleteNote(ctx context.Context, noteID uuid.UUID) error

	// Entity relationship operations
	GetNoteEntityRelationship(ctx context.Context, noteID uuid.UUID) (*uuid.UUID, error)
	CreateEntity(ctx context.Context, name, entityType string, description *string, properties map[string]interface{}) (*uuid.UUID, error)
	CreateEntityRelationship(ctx context.Context, entityID uuid.UUID, relatedType string, relatedID uuid.UUID, relationshipType string, metadata map[string]interface{}) error
	DeleteEntityRelationships(ctx context.Context, entityID uuid.UUID) error
	SoftDeleteEntity(ctx context.Context, entityID uuid.UUID) error
	GetEntityByID(ctx context.Context, entityID uuid.UUID) (*entity.Entity, error)
	GetEntityByName(ctx context.Context, name string) (*entity.Entity, error)

	// Entity reference operations
	CreateEntityReference(ctx context.Context, ref entity.EntityReference) error
	DeleteNoteEntityReferences(ctx context.Context, noteID uuid.UUID) error

	// Tag operations
	GetNoteTagByName(ctx context.Context, tagName string) (*entity.NoteTag, error)
	UpsertNoteTag(ctx context.Context, tagName string, lastActivity int64) (*entity.NoteTag, error)
	CreateItemTag(ctx context.Context, itemID, tagID uuid.UUID) error
	DeleteNoteItemTags(ctx context.Context, noteID uuid.UUID) error
	GetAllNoteTags(ctx context.Context) ([]entity.NoteTag, error)

	// Cleanup operations
	DeleteNoteSemanticIndex(ctx context.Context, noteID uuid.UUID) error

	// Search operations
	SearchSimilarNotes(ctx context.Context, embedding []float32, strategy string, limit int32) ([]entity.NoteListItem, error)

	// Transaction support
	BeginTx(ctx context.Context) (interface{}, error)
	CommitTx(ctx context.Context, tx interface{}) error
	RollbackTx(ctx context.Context, tx interface{}) error
	WithTx(ctx context.Context, tx interface{}) NoteRepository
}

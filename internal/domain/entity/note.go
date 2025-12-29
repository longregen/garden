package entity

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Note represents a note/item in the system
type Note struct {
	ID       uuid.UUID
	Title    *string
	Slug     *string
	Contents *string
	Created  int64
	Modified int64
}

// NoteTag represents a tag associated with a note
type NoteTag struct {
	ID           uuid.UUID
	Name         string
	Created      int64
	Modified     int64
	LastActivity *int64
}

// FullNote represents a note with all its relations
type FullNote struct {
	Note             Note
	Tags             []string
	EntityID         *uuid.UUID
	ProcessedContent *string // Content with entity references converted to markdown links
}

// NoteListItem represents a note in a list view
type NoteListItem struct {
	ID       uuid.UUID `json:"id"`
	Title    *string   `json:"title"`
	Tags     []string  `json:"tags"`
	Created  int64     `json:"created"`
	Modified int64     `json:"modified"`
}

// CreateNoteInput represents the input for creating a note
type CreateNoteInput struct {
	Title    string
	Contents string
	Tags     []string
}

// UpdateNoteInput represents the input for updating a note
type UpdateNoteInput struct {
	Title    *string
	Contents *string
	Tags     *[]string
}

// Entity represents an entity in the knowledge graph
type Entity struct {
	EntityID    uuid.UUID
	Name        string
	Type        string
	Description *string
	Properties  json.RawMessage
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

// CreateEntityInput represents the input for creating an entity
type CreateEntityInput struct {
	Name        string
	Type        string
	Description *string
	Properties  *json.RawMessage
}

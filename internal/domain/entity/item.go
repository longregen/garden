package entity

import (
	"github.com/google/uuid"
)

// Item represents an item in the system
type Item struct {
	ID       uuid.UUID
	Title    *string
	Slug     *string
	Contents *string
	Created  int64
	Modified int64
}

// ItemTag represents a tag associated with an item
type ItemTag struct {
	ID           uuid.UUID
	Name         string
	Created      int64
	Modified     int64
	LastActivity *int64
}

// FullItem represents an item with all its relations
type FullItem struct {
	Item Item
	Tags []string
}

// ItemListItem represents an item in a list view
type ItemListItem struct {
	ID       uuid.UUID `json:"id"`
	Title    *string   `json:"title"`
	Tags     []string  `json:"tags"`
	Created  int64     `json:"created"`
	Modified int64     `json:"modified"`
}

// CreateItemInput represents the input for creating an item
type CreateItemInput struct {
	Title    string
	Contents string
	Tags     []string
}

// UpdateItemInput represents the input for updating an item
type UpdateItemInput struct {
	Title    *string
	Contents *string
}

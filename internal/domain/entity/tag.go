package entity

import (
	"github.com/google/uuid"
)

// Tag represents a tag in the system
type Tag struct {
	ID           uuid.UUID
	Name         string
	Created      int64
	Modified     int64
	LastActivity *int64
}

// TagWithUsage represents a tag with usage count
type TagWithUsage struct {
	Tag
	UsageCount int64
}

// AddTagInput represents the input for adding a tag to an item
type AddTagInput struct {
	ItemID  uuid.UUID
	TagName string
}

// RemoveTagInput represents the input for removing a tag from an item
type RemoveTagInput struct {
	ItemID  uuid.UUID
	TagName string
}

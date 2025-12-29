package input

import (
	"context"

	"garden3/internal/domain/entity"
	"github.com/google/uuid"
)

// TagUseCase defines the business operations for tags
type TagUseCase interface {
	// AddTag adds a tag to an item
	AddTag(ctx context.Context, itemID uuid.UUID, tagName string) error

	// RemoveTag removes a tag from an item
	RemoveTag(ctx context.Context, itemID uuid.UUID, tagName string) error

	// GetTag retrieves a tag by name
	GetTag(ctx context.Context, tagName string) (*entity.Tag, error)

	// ListAllTags retrieves all tags with optional usage counts
	ListAllTags(ctx context.Context, includeUsage bool) ([]entity.TagWithUsage, error)

	// GetItemsByTag retrieves all items that have a specific tag
	GetItemsByTag(ctx context.Context, tagName string, page, pageSize int32) (*PaginatedResponse[entity.ItemListItem], error)

	// DeleteTag deletes a tag and all its associations
	DeleteTag(ctx context.Context, tagName string) error
}

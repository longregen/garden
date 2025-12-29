package output

import (
	"context"

	"garden3/internal/domain/entity"
	"github.com/google/uuid"
)

// TagRepository defines the data access operations for tags
type TagRepository interface {
	// GetTagByName retrieves a tag by its name
	GetTagByName(ctx context.Context, tagName string) (*entity.Tag, error)

	// UpsertTag creates a new tag or updates last_activity if it exists
	UpsertTag(ctx context.Context, tagName string, timestamp int64) (*entity.Tag, error)

	// CreateItemTagRelation creates a relationship between an item and a tag
	CreateItemTagRelation(ctx context.Context, itemID, tagID uuid.UUID) error

	// DeleteItemTagRelation removes the relationship between an item and a tag
	DeleteItemTagRelation(ctx context.Context, itemID, tagID uuid.UUID) error

	// ListAllTags retrieves all tags
	ListAllTags(ctx context.Context) ([]entity.Tag, error)

	// ListAllTagsWithUsage retrieves all tags with their usage counts
	ListAllTagsWithUsage(ctx context.Context) ([]entity.TagWithUsage, error)

	// GetItemsByTag retrieves all items with a specific tag
	GetItemsByTag(ctx context.Context, tagID uuid.UUID, limit, offset int32) ([]entity.ItemListItem, error)

	// CountItemsByTag counts items with a specific tag
	CountItemsByTag(ctx context.Context, tagID uuid.UUID) (int64, error)

	// DeleteTag deletes a tag
	DeleteTag(ctx context.Context, tagID uuid.UUID) error

	// DeleteAllTagRelations deletes all item-tag relations for a tag
	DeleteAllTagRelations(ctx context.Context, tagID uuid.UUID) error
}

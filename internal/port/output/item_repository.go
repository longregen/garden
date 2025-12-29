package output

import (
	"context"

	"github.com/google/uuid"
	"garden3/internal/domain/entity"
)

// ItemRepository defines the data access operations for items
type ItemRepository interface {
	// Item operations
	GetItem(ctx context.Context, itemID uuid.UUID) (*entity.Item, error)
	GetItemWithTags(ctx context.Context, itemID uuid.UUID) (*entity.Item, []string, error)
	ListItems(ctx context.Context, limit, offset int32) ([]entity.ItemListItem, error)
	CountItems(ctx context.Context) (int64, error)
	CreateItem(ctx context.Context, title, slug, contents string) (*entity.Item, error)
	UpdateItem(ctx context.Context, itemID uuid.UUID, title, contents string) error
	DeleteItem(ctx context.Context, itemID uuid.UUID) error

	// Tag operations
	GetItemTagByName(ctx context.Context, tagName string) (*entity.ItemTag, error)
	UpsertItemTag(ctx context.Context, tagName string, lastActivity int64) (*entity.ItemTag, error)
	CreateItemTagRelation(ctx context.Context, itemID, tagID uuid.UUID) error
	DeleteItemTags(ctx context.Context, itemID uuid.UUID) error
	GetItemTags(ctx context.Context, itemID uuid.UUID) ([]string, error)

	// Cleanup operations
	DeleteItemSemanticIndex(ctx context.Context, itemID uuid.UUID) error

	// Search operations
	SearchSimilarItems(ctx context.Context, embedding []float32, limit int32) ([]entity.ItemListItem, error)
}

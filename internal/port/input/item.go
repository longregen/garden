package input

import (
	"context"

	"github.com/google/uuid"
	"garden3/internal/domain/entity"
)

// ItemUseCase defines the business operations for items
type ItemUseCase interface {
	// GetItem retrieves a single item with tags
	GetItem(ctx context.Context, itemID uuid.UUID) (*entity.FullItem, error)

	// ListItems retrieves paginated items
	ListItems(ctx context.Context, page, pageSize int32) (*PaginatedResponse[entity.ItemListItem], error)

	// CreateItem creates a new item with tags
	CreateItem(ctx context.Context, input entity.CreateItemInput) (*entity.FullItem, error)

	// UpdateItem updates an item's information
	UpdateItem(ctx context.Context, itemID uuid.UUID, input entity.UpdateItemInput) (*entity.FullItem, error)

	// DeleteItem deletes an item and all related data
	DeleteItem(ctx context.Context, itemID uuid.UUID) error

	// GetItemTags retrieves tags for a specific item
	GetItemTags(ctx context.Context, itemID uuid.UUID) ([]string, error)

	// UpdateItemTags updates the tags for an item
	UpdateItemTags(ctx context.Context, itemID uuid.UUID, tags []string) error

	// SearchItems performs vector similarity search
	SearchItems(ctx context.Context, embedding []float32) ([]entity.ItemListItem, error)
}

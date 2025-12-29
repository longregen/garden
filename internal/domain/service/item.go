package service

import (
	"context"
	"fmt"

	"garden3/internal/domain/entity"
	"garden3/internal/port/input"
	"garden3/internal/port/output"
	"github.com/google/uuid"
)

// ItemService implements the ItemUseCase interface
type ItemService struct {
	repo output.ItemRepository
}

// NewItemService creates a new item service
func NewItemService(repo output.ItemRepository) *ItemService {
	return &ItemService{
		repo: repo,
	}
}

func (s *ItemService) GetItem(ctx context.Context, itemID uuid.UUID) (*entity.FullItem, error) {
	item, tags, err := s.repo.GetItemWithTags(ctx, itemID)
	if err != nil {
		return nil, fmt.Errorf("failed to get item: %w", err)
	}

	return &entity.FullItem{
		Item: *item,
		Tags: tags,
	}, nil
}

func (s *ItemService) ListItems(ctx context.Context, page, pageSize int32) (*input.PaginatedResponse[entity.ItemListItem], error) {
	offset := (page - 1) * pageSize

	items, err := s.repo.ListItems(ctx, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list items: %w", err)
	}

	total, err := s.repo.CountItems(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count items: %w", err)
	}

	totalPages := int32((total + int64(pageSize) - 1) / int64(pageSize))

	return &input.PaginatedResponse[entity.ItemListItem]{
		Data:       items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *ItemService) CreateItem(ctx context.Context, input entity.CreateItemInput) (*entity.FullItem, error) {
	// Create the item
	item, err := s.repo.CreateItem(ctx, input.Title, "", input.Contents)
	if err != nil {
		return nil, fmt.Errorf("failed to create item: %w", err)
	}

	// Handle tags
	currentTimestamp := item.Created
	if len(input.Tags) > 0 {
		for _, tagName := range input.Tags {
			tag, err := s.repo.UpsertItemTag(ctx, tagName, currentTimestamp)
			if err != nil {
				return nil, fmt.Errorf("failed to upsert tag: %w", err)
			}

			if err := s.repo.CreateItemTagRelation(ctx, item.ID, tag.ID); err != nil {
				return nil, fmt.Errorf("failed to create item tag: %w", err)
			}
		}
	}

	// Return the created item
	return s.GetItem(ctx, item.ID)
}

func (s *ItemService) UpdateItem(ctx context.Context, itemID uuid.UUID, input entity.UpdateItemInput) (*entity.FullItem, error) {
	// Get existing item
	item, err := s.repo.GetItem(ctx, itemID)
	if err != nil {
		return nil, fmt.Errorf("failed to get item: %w", err)
	}

	// Prepare update values
	title := ""
	if input.Title != nil {
		title = *input.Title
	} else if item.Title != nil {
		title = *item.Title
	}

	contents := ""
	if input.Contents != nil {
		contents = *input.Contents
	} else if item.Contents != nil {
		contents = *item.Contents
	}

	// Update the item
	if err := s.repo.UpdateItem(ctx, itemID, title, contents); err != nil {
		return nil, fmt.Errorf("failed to update item: %w", err)
	}

	return s.GetItem(ctx, itemID)
}

func (s *ItemService) DeleteItem(ctx context.Context, itemID uuid.UUID) error {
	// Delete item tags
	if err := s.repo.DeleteItemTags(ctx, itemID); err != nil {
		return fmt.Errorf("failed to delete item tags: %w", err)
	}

	// Delete semantic index
	if err := s.repo.DeleteItemSemanticIndex(ctx, itemID); err != nil {
		// Ignore error if semantic index doesn't exist
	}

	// Delete the item
	if err := s.repo.DeleteItem(ctx, itemID); err != nil {
		return fmt.Errorf("failed to delete item: %w", err)
	}

	return nil
}

func (s *ItemService) GetItemTags(ctx context.Context, itemID uuid.UUID) ([]string, error) {
	tags, err := s.repo.GetItemTags(ctx, itemID)
	if err != nil {
		return nil, fmt.Errorf("failed to get item tags: %w", err)
	}
	return tags, nil
}

func (s *ItemService) UpdateItemTags(ctx context.Context, itemID uuid.UUID, tags []string) error {
	// Get existing item to get current timestamp
	item, err := s.repo.GetItem(ctx, itemID)
	if err != nil {
		return fmt.Errorf("failed to get item: %w", err)
	}

	// Delete existing tags
	if err := s.repo.DeleteItemTags(ctx, itemID); err != nil {
		return fmt.Errorf("failed to delete tags: %w", err)
	}

	// Insert new tags
	currentTimestamp := item.Modified
	for _, tagName := range tags {
		tag, err := s.repo.UpsertItemTag(ctx, tagName, currentTimestamp)
		if err != nil {
			return fmt.Errorf("failed to upsert tag: %w", err)
		}

		if err := s.repo.CreateItemTagRelation(ctx, itemID, tag.ID); err != nil {
			return fmt.Errorf("failed to create item tag: %w", err)
		}
	}

	return nil
}

func (s *ItemService) SearchItems(ctx context.Context, embedding []float32) ([]entity.ItemListItem, error) {
	results, err := s.repo.SearchSimilarItems(ctx, embedding, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to search similar items: %w", err)
	}

	return results, nil
}

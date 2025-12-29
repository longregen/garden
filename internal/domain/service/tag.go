package service

import (
	"context"
	"fmt"
	"time"

	"garden3/internal/domain/entity"
	"garden3/internal/port/input"
	"garden3/internal/port/output"
	"github.com/google/uuid"
)

// TagService implements the TagUseCase interface
type TagService struct {
	repo output.TagRepository
}

// NewTagService creates a new tag service
func NewTagService(repo output.TagRepository) *TagService {
	return &TagService{
		repo: repo,
	}
}

func (s *TagService) AddTag(ctx context.Context, itemID uuid.UUID, tagName string) error {
	timestamp := time.Now().Unix()

	// Get or create the tag
	tag, err := s.repo.UpsertTag(ctx, tagName, timestamp)
	if err != nil {
		return fmt.Errorf("failed to upsert tag: %w", err)
	}

	// Create the item-tag relationship
	if err := s.repo.CreateItemTagRelation(ctx, itemID, tag.ID); err != nil {
		return fmt.Errorf("failed to create item-tag relation: %w", err)
	}

	return nil
}

func (s *TagService) RemoveTag(ctx context.Context, itemID uuid.UUID, tagName string) error {
	// Get the tag
	tag, err := s.repo.GetTagByName(ctx, tagName)
	if err != nil {
		return fmt.Errorf("tag not found: %w", err)
	}

	// Remove the item-tag relationship
	if err := s.repo.DeleteItemTagRelation(ctx, itemID, tag.ID); err != nil {
		return fmt.Errorf("failed to delete item-tag relation: %w", err)
	}

	return nil
}

func (s *TagService) GetTag(ctx context.Context, tagName string) (*entity.Tag, error) {
	tag, err := s.repo.GetTagByName(ctx, tagName)
	if err != nil {
		return nil, fmt.Errorf("failed to get tag: %w", err)
	}

	return tag, nil
}

func (s *TagService) ListAllTags(ctx context.Context, includeUsage bool) ([]entity.TagWithUsage, error) {
	if includeUsage {
		tags, err := s.repo.ListAllTagsWithUsage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list tags with usage: %w", err)
		}
		return tags, nil
	}

	tags, err := s.repo.ListAllTags(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list tags: %w", err)
	}

	// Convert to TagWithUsage with zero usage count
	result := make([]entity.TagWithUsage, len(tags))
	for i, tag := range tags {
		result[i] = entity.TagWithUsage{
			Tag:        tag,
			UsageCount: 0,
		}
	}

	return result, nil
}

func (s *TagService) GetItemsByTag(ctx context.Context, tagName string, page, pageSize int32) (*input.PaginatedResponse[entity.ItemListItem], error) {
	// Get the tag
	tag, err := s.repo.GetTagByName(ctx, tagName)
	if err != nil {
		return nil, fmt.Errorf("tag not found: %w", err)
	}

	offset := (page - 1) * pageSize

	// Get items
	items, err := s.repo.GetItemsByTag(ctx, tag.ID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get items by tag: %w", err)
	}

	// Get total count
	total, err := s.repo.CountItemsByTag(ctx, tag.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to count items by tag: %w", err)
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

func (s *TagService) DeleteTag(ctx context.Context, tagName string) error {
	// Get the tag
	tag, err := s.repo.GetTagByName(ctx, tagName)
	if err != nil {
		return fmt.Errorf("tag not found: %w", err)
	}

	// Delete all tag relations
	if err := s.repo.DeleteAllTagRelations(ctx, tag.ID); err != nil {
		return fmt.Errorf("failed to delete tag relations: %w", err)
	}

	// Delete the tag
	if err := s.repo.DeleteTag(ctx, tag.ID); err != nil {
		return fmt.Errorf("failed to delete tag: %w", err)
	}

	return nil
}

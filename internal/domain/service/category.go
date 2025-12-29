package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"garden3/internal/domain/entity"
	"garden3/internal/port/output"
)

// CategoryService implements the CategoryUseCase interface
type CategoryService struct {
	repo output.CategoryRepository
}

// NewCategoryService creates a new category service
func NewCategoryService(repo output.CategoryRepository) *CategoryService {
	return &CategoryService{
		repo: repo,
	}
}

func (s *CategoryService) ListCategories(ctx context.Context) ([]entity.CategoryWithSources, error) {
	categories, err := s.repo.ListCategoriesWithSources(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}
	return categories, nil
}

func (s *CategoryService) GetCategory(ctx context.Context, categoryID uuid.UUID) (*entity.Category, error) {
	category, err := s.repo.GetCategory(ctx, categoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get category: %w", err)
	}
	return category, nil
}

func (s *CategoryService) UpdateCategory(ctx context.Context, categoryID uuid.UUID, name string) error {
	if err := s.repo.UpdateCategory(ctx, categoryID, name); err != nil {
		return fmt.Errorf("failed to update category: %w", err)
	}
	return nil
}

func (s *CategoryService) MergeCategories(ctx context.Context, input entity.MergeCategoriesInput) error {
	if err := s.repo.MergeCategories(ctx, input.SourceID, input.TargetID); err != nil {
		return fmt.Errorf("failed to merge categories: %w", err)
	}
	return nil
}

func (s *CategoryService) CreateCategorySource(ctx context.Context, input entity.CreateCategorySourceInput) (*entity.CategorySource, error) {
	source, err := s.repo.CreateCategorySource(ctx, input.CategoryID, input.SourceURI, input.RawSource)
	if err != nil {
		return nil, fmt.Errorf("failed to create category source: %w", err)
	}
	return source, nil
}

func (s *CategoryService) UpdateCategorySource(ctx context.Context, sourceID uuid.UUID, input entity.UpdateCategorySourceInput) error {
	// Get existing source to preserve values that weren't updated
	existing, err := s.repo.GetCategorySource(ctx, sourceID)
	if err != nil {
		return fmt.Errorf("failed to get existing source: %w", err)
	}

	sourceURI := existing.SourceURI
	if input.SourceURI != nil {
		sourceURI = input.SourceURI
	}

	rawSource := existing.RawSource
	if input.RawSource != nil {
		rawSource = *input.RawSource
	}

	if err := s.repo.UpdateCategorySource(ctx, sourceID, sourceURI, rawSource); err != nil {
		return fmt.Errorf("failed to update category source: %w", err)
	}
	return nil
}

func (s *CategoryService) DeleteCategorySource(ctx context.Context, sourceID uuid.UUID) error {
	if err := s.repo.DeleteCategorySource(ctx, sourceID); err != nil {
		return fmt.Errorf("failed to delete category source: %w", err)
	}
	return nil
}

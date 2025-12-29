package input

import (
	"context"

	"github.com/google/uuid"
	"garden3/internal/domain/entity"
)

// CategoryUseCase defines the business operations for categories
type CategoryUseCase interface {
	// ListCategories retrieves all categories with their sources
	ListCategories(ctx context.Context) ([]entity.CategoryWithSources, error)

	// GetCategory retrieves a single category by ID
	GetCategory(ctx context.Context, categoryID uuid.UUID) (*entity.Category, error)

	// UpdateCategory updates a category's name
	UpdateCategory(ctx context.Context, categoryID uuid.UUID, name string) error

	// MergeCategories merges two categories
	MergeCategories(ctx context.Context, input entity.MergeCategoriesInput) error

	// CreateCategorySource adds a source to a category
	CreateCategorySource(ctx context.Context, input entity.CreateCategorySourceInput) (*entity.CategorySource, error)

	// UpdateCategorySource updates a category source
	UpdateCategorySource(ctx context.Context, sourceID uuid.UUID, input entity.UpdateCategorySourceInput) error

	// DeleteCategorySource deletes a category source
	DeleteCategorySource(ctx context.Context, sourceID uuid.UUID) error
}

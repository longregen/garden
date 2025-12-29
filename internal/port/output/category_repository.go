package output

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"garden3/internal/domain/entity"
)

// CategoryRepository defines the data access operations for categories
type CategoryRepository interface {
	// ListCategoriesWithSources retrieves all categories with their sources
	ListCategoriesWithSources(ctx context.Context) ([]entity.CategoryWithSources, error)

	// GetCategory retrieves a category by ID
	GetCategory(ctx context.Context, categoryID uuid.UUID) (*entity.Category, error)

	// UpdateCategory updates a category's name
	UpdateCategory(ctx context.Context, categoryID uuid.UUID, name string) error

	// MergeCategories calls the merge_categories stored procedure
	MergeCategories(ctx context.Context, sourceID, targetID uuid.UUID) error

	// CreateCategorySource creates a new category source
	CreateCategorySource(ctx context.Context, categoryID uuid.UUID, sourceURI *string, rawSource json.RawMessage) (*entity.CategorySource, error)

	// UpdateCategorySource updates a category source
	UpdateCategorySource(ctx context.Context, sourceID uuid.UUID, sourceURI *string, rawSource json.RawMessage) error

	// DeleteCategorySource deletes a category source
	DeleteCategorySource(ctx context.Context, sourceID uuid.UUID) error

	// GetCategorySource retrieves a category source by ID
	GetCategorySource(ctx context.Context, sourceID uuid.UUID) (*entity.CategorySource, error)
}

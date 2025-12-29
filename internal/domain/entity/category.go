package entity

import (
	"encoding/json"

	"github.com/google/uuid"
)

// Category represents a classification category
type Category struct {
	CategoryID uuid.UUID
	Name       string
}

// CategorySource represents a source for a category
type CategorySource struct {
	ID         uuid.UUID
	CategoryID uuid.UUID
	SourceURI  *string
	RawSource  json.RawMessage
}

// CategoryWithSources represents a category with all its sources
type CategoryWithSources struct {
	Category Category
	Sources  []CategorySource
}

// CreateCategorySourceInput represents the input for creating a category source
type CreateCategorySourceInput struct {
	CategoryID uuid.UUID
	SourceURI  *string
	RawSource  json.RawMessage
}

// UpdateCategorySourceInput represents the input for updating a category source
type UpdateCategorySourceInput struct {
	SourceURI *string
	RawSource *json.RawMessage
}

// MergeCategoriesInput represents the input for merging categories
type MergeCategoriesInput struct {
	SourceID uuid.UUID
	TargetID uuid.UUID
}

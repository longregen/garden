package input

import (
	"context"

	"garden3/internal/domain/entity"
	"github.com/google/uuid"
)

// EntityUseCase defines the business operations for entities
type EntityUseCase interface {
	// ListEntities retrieves all entities with optional filters
	ListEntities(ctx context.Context, entityType string, updatedSince *string) ([]entity.Entity, error)

	// GetEntity retrieves a single entity by ID
	GetEntity(ctx context.Context, entityID uuid.UUID) (*entity.Entity, error)

	// CreateEntity creates a new entity
	CreateEntity(ctx context.Context, input entity.CreateEntityInput) (*entity.Entity, error)

	// UpdateEntity updates an entity
	UpdateEntity(ctx context.Context, entityID uuid.UUID, input entity.UpdateEntityInput) (*entity.Entity, error)

	// DeleteEntity soft deletes an entity
	DeleteEntity(ctx context.Context, entityID uuid.UUID) error

	// SearchEntities searches entities by name with optional type filter
	SearchEntities(ctx context.Context, query string, entityType string) ([]entity.Entity, error)

	// ListDeletedEntities retrieves all soft-deleted entities
	ListDeletedEntities(ctx context.Context) ([]entity.Entity, error)

	// GetEntityRelationships retrieves relationships for an entity with optional filters
	GetEntityRelationships(ctx context.Context, entityID uuid.UUID, relatedType, relationshipType *string) ([]entity.EntityRelationship, error)

	// CreateEntityRelationship creates a new entity relationship
	CreateEntityRelationship(ctx context.Context, input entity.CreateEntityRelationshipInput) (*entity.EntityRelationship, error)

	// DeleteEntityRelationship deletes an entity relationship
	DeleteEntityRelationship(ctx context.Context, relationshipID uuid.UUID) error

	// GetEntityReferences retrieves entity references with optional filters
	GetEntityReferences(ctx context.Context, entityID uuid.UUID, sourceType *string) ([]entity.EntityReference, error)

	// ParseEntityReferences parses [[entity]] references from content
	ParseEntityReferences(ctx context.Context, content string) ([]entity.ParsedReference, error)
}

package output

import (
	"context"
	"time"

	"github.com/google/uuid"
	"garden3/internal/domain/entity"
)

// EntityRepository defines the data access operations for entities
type EntityRepository interface {
	// ListEntities retrieves all entities with optional filters
	ListEntities(ctx context.Context, entityType string, updatedSince *string) ([]entity.Entity, error)

	// GetEntity retrieves an entity by ID (excluding soft deleted)
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

	// GetEntityRelationships retrieves relationships for an entity
	GetEntityRelationships(ctx context.Context, entityID uuid.UUID) ([]entity.EntityRelationship, error)

	// GetEntityRelationshipsByTypeAndRelationship retrieves filtered relationships
	GetEntityRelationshipsByTypeAndRelationship(ctx context.Context, entityID uuid.UUID, relatedType, relationshipType string) ([]entity.EntityRelationship, error)

	// CreateEntityRelationship creates a new entity relationship
	CreateEntityRelationship(ctx context.Context, entityID uuid.UUID, relatedType string, relatedID uuid.UUID, relationshipType string, metadata []byte) (*entity.EntityRelationship, error)

	// DeleteEntityRelationship deletes an entity relationship by ID
	DeleteEntityRelationship(ctx context.Context, relationshipID uuid.UUID) error

	// UpdateContactNameByRelationship updates contact names for entities with identity relationships
	UpdateContactNameByRelationship(ctx context.Context, entityID uuid.UUID, relationshipType, name string) error

	// GetEntityReferences retrieves entity references with optional source type filter
	GetEntityReferences(ctx context.Context, entityID uuid.UUID, sourceType *string) ([]entity.EntityReference, error)

	// ListEntitiesUpdatedSince retrieves entities updated after a specific time
	ListEntitiesUpdatedSince(ctx context.Context, since time.Time) ([]entity.Entity, error)

	// ListEntitiesByProperty retrieves entities that have a specific property with a specific value
	ListEntitiesByProperty(ctx context.Context, propertyKey string, propertyValue string) ([]entity.Entity, error)
}

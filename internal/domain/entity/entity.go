package entity

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// EntityRelationship represents a relationship between an entity and another resource
type EntityRelationship struct {
	ID               uuid.UUID
	EntityID         uuid.UUID
	RelatedType      string
	RelatedID        uuid.UUID
	RelationshipType string
	Metadata         json.RawMessage
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// UpdateEntityInput represents the input for updating an entity
type UpdateEntityInput struct {
	Name        *string
	Type        *string
	Description *string
	Properties  *json.RawMessage
}

// CreateEntityRelationshipInput represents the input for creating a relationship
type CreateEntityRelationshipInput struct {
	EntityID         uuid.UUID
	RelatedType      string
	RelatedID        uuid.UUID
	RelationshipType string
	Metadata         json.RawMessage
}

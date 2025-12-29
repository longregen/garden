package entity

import (
	"time"

	"github.com/google/uuid"
)

// EntityReference represents a reference to an entity in content
type EntityReference struct {
	ID            uuid.UUID
	SourceType    string
	SourceID      uuid.UUID
	EntityID      uuid.UUID
	ReferenceText string
	Position      *int
	CreatedAt     time.Time
}

// ParsedReference represents a parsed [[entity]] reference from content
type ParsedReference struct {
	Original    string
	EntityName  string
	DisplayText string
}

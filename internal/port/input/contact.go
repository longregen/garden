package input

import (
	"context"

	"github.com/google/uuid"
	"garden3/internal/domain/entity"
)

// ContactUseCase defines the business operations for contacts
type ContactUseCase interface {
	// GetContact retrieves a single contact with all its relations
	GetContact(ctx context.Context, contactID uuid.UUID) (*entity.FullContact, error)

	// ListContacts retrieves paginated contacts with optional search
	ListContacts(ctx context.Context, page, pageSize int32, searchQuery *string) (*PaginatedResponse[entity.Contact], error)

	// CreateContact creates a new contact
	CreateContact(ctx context.Context, input entity.CreateContactInput) (*entity.Contact, error)

	// UpdateContact updates a contact's basic information
	UpdateContact(ctx context.Context, contactID uuid.UUID, input entity.UpdateContactInput) error

	// DeleteContact deletes a contact
	DeleteContact(ctx context.Context, contactID uuid.UUID) error

	// UpdateEvaluation updates or creates a contact's evaluation scores
	UpdateEvaluation(ctx context.Context, contactID uuid.UUID, input entity.UpdateEvaluationInput) error

	// MergeContacts merges two contacts together
	MergeContacts(ctx context.Context, sourceID, targetID uuid.UUID) error

	// BatchUpdate updates multiple contacts atomically
	BatchUpdate(ctx context.Context, updates []entity.BatchContactUpdate) ([]entity.BatchUpdateResult, error)

	// RefreshStats updates message statistics for all contacts
	RefreshStats(ctx context.Context) error

	// GetContactTags retrieves all tags for a contact
	GetContactTags(ctx context.Context, contactID uuid.UUID) ([]entity.ContactTag, error)

	// AddTag adds a tag to a contact (creates tag if it doesn't exist)
	AddTag(ctx context.Context, contactID uuid.UUID, tagName string) (*entity.ContactTag, error)

	// RemoveTag removes a tag from a contact
	RemoveTag(ctx context.Context, contactID uuid.UUID, tagID uuid.UUID) error

	// ListAllTags retrieves all tag names in the system
	ListAllTags(ctx context.Context) ([]entity.ContactTagName, error)

	// ListAllContactSources retrieves all contact sources
	ListAllContactSources(ctx context.Context) ([]entity.ContactSource, error)

	// CreateContactSource creates a new contact source
	CreateContactSource(ctx context.Context, input entity.CreateContactSourceInput) (*entity.ContactSource, error)

	// UpdateContactSource updates a contact source
	UpdateContactSource(ctx context.Context, id uuid.UUID, input entity.UpdateContactSourceInput) (*entity.ContactSource, error)

	// DeleteContactSource deletes a contact source
	DeleteContactSource(ctx context.Context, id uuid.UUID) error
}

// PaginatedResponse represents a paginated list of items
type PaginatedResponse[T any] struct {
	Data       []T   `json:"data"`
	Total      int64 `json:"total"`
	Page       int32 `json:"page"`
	PageSize   int32 `json:"pageSize"`
	TotalPages int32 `json:"totalPages"`
}

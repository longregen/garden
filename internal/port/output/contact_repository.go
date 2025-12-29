package output

import (
	"context"

	"github.com/google/uuid"
	"garden3/internal/domain/entity"
)

// ContactRepository defines the data access operations for contacts
type ContactRepository interface {
	// GetContact retrieves a contact by ID
	GetContact(ctx context.Context, contactID uuid.UUID) (*entity.Contact, error)

	// GetContactEvaluation retrieves evaluation for a contact
	GetContactEvaluation(ctx context.Context, contactID uuid.UUID) (*entity.ContactEvaluation, error)

	// GetContactTags retrieves all tags for a contact
	GetContactTags(ctx context.Context, contactID uuid.UUID) ([]entity.ContactTag, error)

	// GetContactKnownNames retrieves all known names for a contact
	GetContactKnownNames(ctx context.Context, contactID uuid.UUID) ([]entity.ContactKnownName, error)

	// GetContactRooms retrieves all rooms for a contact
	GetContactRooms(ctx context.Context, contactID uuid.UUID) ([]entity.ContactRoom, error)

	// GetContactSources retrieves all sources for a contact
	GetContactSources(ctx context.Context, contactID uuid.UUID) ([]entity.ContactSource, error)

	// ListContacts retrieves paginated contacts
	ListContacts(ctx context.Context, limit, offset int32) ([]entity.Contact, error)

	// SearchContacts retrieves paginated contacts matching search query
	SearchContacts(ctx context.Context, searchPattern string, limit, offset int32) ([]entity.Contact, error)

	// CountContacts returns the total number of contacts
	CountContacts(ctx context.Context) (int64, error)

	// GetEvaluationsByContactIDs retrieves evaluations for multiple contacts
	GetEvaluationsByContactIDs(ctx context.Context, contactIDs []uuid.UUID) (map[uuid.UUID]*entity.ContactEvaluation, error)

	// GetTagsByContactIDs retrieves tags for multiple contacts
	GetTagsByContactIDs(ctx context.Context, contactIDs []uuid.UUID) (map[uuid.UUID][]entity.ContactTag, error)

	// CreateContact creates a new contact
	CreateContact(ctx context.Context, input entity.CreateContactInput) (*entity.Contact, error)

	// UpdateContact updates a contact's fields
	UpdateContact(ctx context.Context, contactID uuid.UUID, input entity.UpdateContactInput) error

	// DeleteContact deletes a contact
	DeleteContact(ctx context.Context, contactID uuid.UUID) error

	// EvaluationExists checks if an evaluation exists for a contact
	EvaluationExists(ctx context.Context, contactID uuid.UUID) (bool, error)

	// CreateEvaluation creates a new evaluation
	CreateEvaluation(ctx context.Context, contactID uuid.UUID, input entity.UpdateEvaluationInput) error

	// UpdateEvaluation updates an existing evaluation
	UpdateEvaluation(ctx context.Context, contactID uuid.UUID, input entity.UpdateEvaluationInput) error

	// MergeContacts calls the merge_contacts stored procedure
	MergeContacts(ctx context.Context, sourceID, targetID uuid.UUID) error

	// GetAllTagNames retrieves all tag names
	GetAllTagNames(ctx context.Context) ([]entity.ContactTagName, error)

	// GetContactTagByName retrieves a contact tag by name
	GetContactTagByName(ctx context.Context, tagName string) (*entity.ContactTagName, error)

	// CreateTagName creates a new tag name
	CreateTagName(ctx context.Context, tagName string) (*entity.ContactTagName, error)

	// ContactTagExists checks if a contact already has a tag
	ContactTagExists(ctx context.Context, contactID, tagID uuid.UUID) (bool, error)

	// AddContactTag adds a tag to a contact
	AddContactTag(ctx context.Context, contactID, tagID uuid.UUID) error

	// RemoveContactTag removes a tag from a contact
	RemoveContactTag(ctx context.Context, contactID, tagID uuid.UUID) error

	// EnsureContactStatsExist ensures all contacts have stats records
	EnsureContactStatsExist(ctx context.Context) error

	// UpdateContactMessageStats updates message statistics for all contacts
	UpdateContactMessageStats(ctx context.Context) error

	// ListAllContactSources retrieves all contact sources
	ListAllContactSources(ctx context.Context) ([]entity.ContactSource, error)

	// CreateContactSource creates a new contact source
	CreateContactSource(ctx context.Context, input entity.CreateContactSourceInput) (*entity.ContactSource, error)

	// UpdateContactSource updates a contact source
	UpdateContactSource(ctx context.Context, id uuid.UUID, input entity.UpdateContactSourceInput) (*entity.ContactSource, error)

	// DeleteContactSource deletes a contact source
	DeleteContactSource(ctx context.Context, id uuid.UUID) error
}

package entity

import (
	"time"

	"github.com/google/uuid"
)

// Contact represents a person in the system
type Contact struct {
	ContactID        uuid.UUID
	Name             string
	Email            *string
	Phone            *string
	Birthday         *string
	Notes            *string
	Extras           map[string]interface{}
	CreationDate     time.Time
	LastUpdate       time.Time
	LastWeekMessages *int32
	GroupsInCommon   *int32
}

// ContactEvaluation represents evaluation metrics for a contact
type ContactEvaluation struct {
	ID         uuid.UUID
	ContactID  uuid.UUID
	Importance *int32
	Closeness  *int32
	Fondness   *int32
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// ContactTag represents a tag associated with a contact
type ContactTag struct {
	TagID uuid.UUID
	Name  string
}

// ContactKnownName represents an alternative name for a contact
type ContactKnownName struct {
	ID        uuid.UUID
	ContactID uuid.UUID
	Name      string
}

// ContactSource represents an external source for contact information
type ContactSource struct {
	ID         uuid.UUID
	ContactID  *uuid.UUID
	SourceID   string
	SourceName string
}

// ContactRoom represents a room/group the contact is in
type ContactRoom struct {
	RoomID          uuid.UUID
	DisplayName     *string
	UserDefinedName *string
}

// FullContact represents a contact with all its relations
type FullContact struct {
	Contact      Contact
	Evaluation   *ContactEvaluation
	Tags         []ContactTag
	KnownNames   []ContactKnownName
	Rooms        []ContactRoom
	Sources      []ContactSource
}

// ContactTagName represents a tag name in the system
type ContactTagName struct {
	TagID     uuid.UUID
	Name      string
	CreatedAt time.Time
}

// BatchUpdateResult represents the result of updating a single contact in a batch
type BatchUpdateResult struct {
	ContactID uuid.UUID `json:"contact_id"`
	Success   bool      `json:"success"`
	Error     *string   `json:"error,omitempty"`
}

// CreateContactInput represents the input for creating a contact
type CreateContactInput struct {
	Name     string
	Email    *string
	Phone    *string
	Birthday *string
	Notes    *string
	Extras   map[string]interface{}
}

// UpdateContactInput represents the input for updating a contact
type UpdateContactInput struct {
	Name     *string
	Email    *string
	Phone    *string
	Birthday *string
	Notes    *string
	Extras   map[string]interface{}
}

// UpdateEvaluationInput represents the input for updating a contact evaluation
type UpdateEvaluationInput struct {
	Importance *int32
	Closeness  *int32
	Fondness   *int32
}

// BatchContactUpdate represents a single contact update in a batch operation
type BatchContactUpdate struct {
	ContactID  uuid.UUID
	Name       *string
	Notes      *string
	Importance *int32
	Closeness  *int32
	Fondness   *int32
	Tags       []ContactTag
}

// CreateContactSourceInput represents the input for creating a contact source
type CreateContactSourceInput struct {
	ContactID  uuid.UUID
	SourceID   string
	SourceName string
}

// UpdateContactSourceInput represents the input for updating a contact source
type UpdateContactSourceInput struct {
	SourceID   string
	SourceName string
}

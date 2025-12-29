package repository

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"garden3/internal/adapter/secondary/postgres/generated/db"
	"garden3/internal/domain/entity"
)

// ContactRepository implements the output.ContactRepository interface
type ContactRepository struct {
	pool *pgxpool.Pool
}

// NewContactRepository creates a new contact repository
func NewContactRepository(pool *pgxpool.Pool) *ContactRepository {
	return &ContactRepository{
		pool: pool,
	}
}

func (r *ContactRepository) GetContact(ctx context.Context, contactID uuid.UUID) (*entity.Contact, error) {
	queries := db.New(r.pool)
	dbContact, err := queries.GetContact(ctx, contactID)
	if err != nil {
		return nil, err
	}

	var extras map[string]interface{}
	if len(dbContact.Extras) > 0 {
		if err := json.Unmarshal(dbContact.Extras, &extras); err != nil {
			extras = make(map[string]interface{})
		}
	} else {
		extras = make(map[string]interface{})
	}

	return &entity.Contact{
		ContactID:        dbContact.ContactID,
		Name:             dbContact.Name,
		Email:            dbContact.Email,
		Phone:            dbContact.Phone,
		Birthday:         dbContact.Birthday,
		Notes:            dbContact.Notes,
		Extras:           extras,
		CreationDate:     dbContact.CreationDate.Time,
		LastUpdate:       dbContact.LastUpdate.Time,
		LastWeekMessages: dbContact.LastWeekMessages,
		GroupsInCommon:   dbContact.GroupsInCommon,
	}, nil
}

func (r *ContactRepository) GetContactEvaluation(ctx context.Context, contactID uuid.UUID) (*entity.ContactEvaluation, error) {
	queries := db.New(r.pool)
	dbEval, err := queries.GetContactEvaluation(ctx, contactID)
	if err != nil {
		return nil, err
	}

	return &entity.ContactEvaluation{
		ID:         dbEval.ID,
		ContactID:  dbEval.ContactID,
		Importance: dbEval.Importance,
		Closeness:  dbEval.Closeness,
		Fondness:   dbEval.Fondness,
		CreatedAt:  dbEval.CreatedAt.Time,
		UpdatedAt:  dbEval.UpdatedAt.Time,
	}, nil
}

func (r *ContactRepository) GetContactTags(ctx context.Context, contactID uuid.UUID) ([]entity.ContactTag, error) {
	queries := db.New(r.pool)
	dbTags, err := queries.GetContactTags(ctx, contactID)
	if err != nil {
		return nil, err
	}

	tags := make([]entity.ContactTag, len(dbTags))
	for i, dbTag := range dbTags {
		tags[i] = entity.ContactTag{
			TagID: dbTag.TagID,
			Name:  dbTag.Name,
		}
	}
	return tags, nil
}

func (r *ContactRepository) GetContactKnownNames(ctx context.Context, contactID uuid.UUID) ([]entity.ContactKnownName, error) {
	queries := db.New(r.pool)
	dbNames, err := queries.GetContactKnownNames(ctx, contactID)
	if err != nil {
		return nil, err
	}

	names := make([]entity.ContactKnownName, len(dbNames))
	for i, dbName := range dbNames {
		names[i] = entity.ContactKnownName{
			ID:        dbName.ID,
			ContactID: dbName.ContactID,
			Name:      dbName.Name,
		}
	}
	return names, nil
}

func (r *ContactRepository) GetContactRooms(ctx context.Context, contactID uuid.UUID) ([]entity.ContactRoom, error) {
	queries := db.New(r.pool)
	dbRooms, err := queries.GetContactRooms(ctx, contactID)
	if err != nil {
		return nil, err
	}

	rooms := make([]entity.ContactRoom, len(dbRooms))
	for i, dbRoom := range dbRooms {
		rooms[i] = entity.ContactRoom{
			RoomID:          dbRoom.RoomID,
			DisplayName:     dbRoom.DisplayName,
			UserDefinedName: dbRoom.UserDefinedName,
		}
	}
	return rooms, nil
}

func (r *ContactRepository) GetContactSources(ctx context.Context, contactID uuid.UUID) ([]entity.ContactSource, error) {
	queries := db.New(r.pool)

	// Convert uuid.UUID to pgtype.UUID
	pgContactID := pgtype.UUID{
		Bytes: contactID,
		Valid: true,
	}

	dbSources, err := queries.GetContactSources(ctx, pgContactID)
	if err != nil {
		return nil, err
	}

	sources := make([]entity.ContactSource, len(dbSources))
	for i, dbSource := range dbSources {
		var contactIDPtr *uuid.UUID
		if dbSource.ContactID.Valid {
			cid := uuid.UUID(dbSource.ContactID.Bytes)
			contactIDPtr = &cid
		}
		sources[i] = entity.ContactSource{
			ID:         dbSource.ID,
			ContactID:  contactIDPtr,
			SourceID:   dbSource.SourceID,
			SourceName: dbSource.SourceName,
		}
	}
	return sources, nil
}

func (r *ContactRepository) ListContacts(ctx context.Context, limit, offset int32) ([]entity.Contact, error) {
	queries := db.New(r.pool)
	dbContacts, err := queries.ListContacts(ctx, db.ListContactsParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	contacts := make([]entity.Contact, len(dbContacts))
	for i, dbContact := range dbContacts {
		var extras map[string]interface{}
		if len(dbContact.Extras) > 0 {
			if err := json.Unmarshal(dbContact.Extras, &extras); err != nil {
				extras = make(map[string]interface{})
			}
		} else {
			extras = make(map[string]interface{})
		}

		contacts[i] = entity.Contact{
			ContactID:        dbContact.ContactID,
			Name:             dbContact.Name,
			Email:            dbContact.Email,
			Phone:            dbContact.Phone,
			Birthday:         dbContact.Birthday,
			Notes:            dbContact.Notes,
			Extras:           extras,
			CreationDate:     dbContact.CreationDate.Time,
			LastUpdate:       dbContact.LastUpdate.Time,
			LastWeekMessages: dbContact.LastWeekMessages,
			GroupsInCommon:   dbContact.GroupsInCommon,
		}
	}
	return contacts, nil
}

func (r *ContactRepository) SearchContacts(ctx context.Context, searchPattern string, limit, offset int32) ([]entity.Contact, error) {
	queries := db.New(r.pool)
	dbContacts, err := queries.SearchContacts(ctx, db.SearchContactsParams{
		Name:   searchPattern,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	contacts := make([]entity.Contact, len(dbContacts))
	for i, dbContact := range dbContacts {
		var extras map[string]interface{}
		if len(dbContact.Extras) > 0 {
			if err := json.Unmarshal(dbContact.Extras, &extras); err != nil {
				extras = make(map[string]interface{})
			}
		} else {
			extras = make(map[string]interface{})
		}

		contacts[i] = entity.Contact{
			ContactID:        dbContact.ContactID,
			Name:             dbContact.Name,
			Email:            dbContact.Email,
			Phone:            dbContact.Phone,
			Birthday:         dbContact.Birthday,
			Notes:            dbContact.Notes,
			Extras:           extras,
			CreationDate:     dbContact.CreationDate.Time,
			LastUpdate:       dbContact.LastUpdate.Time,
			LastWeekMessages: dbContact.LastWeekMessages,
			GroupsInCommon:   dbContact.GroupsInCommon,
		}
	}
	return contacts, nil
}

func (r *ContactRepository) CountContacts(ctx context.Context) (int64, error) {
	queries := db.New(r.pool)
	return queries.CountContacts(ctx)
}

func (r *ContactRepository) GetEvaluationsByContactIDs(ctx context.Context, contactIDs []uuid.UUID) (map[uuid.UUID]*entity.ContactEvaluation, error) {
	queries := db.New(r.pool)
	dbEvals, err := queries.GetContactEvaluationsByContactIds(ctx, contactIDs)
	if err != nil {
		return nil, err
	}

	evals := make(map[uuid.UUID]*entity.ContactEvaluation)
	for _, dbEval := range dbEvals {
		evals[dbEval.ContactID] = &entity.ContactEvaluation{
			ContactID:  dbEval.ContactID,
			Importance: dbEval.Importance,
			Closeness:  dbEval.Closeness,
			Fondness:   dbEval.Fondness,
		}
	}
	return evals, nil
}

func (r *ContactRepository) GetTagsByContactIDs(ctx context.Context, contactIDs []uuid.UUID) (map[uuid.UUID][]entity.ContactTag, error) {
	queries := db.New(r.pool)
	dbTags, err := queries.GetContactTagsByContactIds(ctx, contactIDs)
	if err != nil {
		return nil, err
	}

	tags := make(map[uuid.UUID][]entity.ContactTag)
	for _, dbTag := range dbTags {
		tags[dbTag.ContactID] = append(tags[dbTag.ContactID], entity.ContactTag{
			TagID: dbTag.TagID,
			Name:  dbTag.Name,
		})
	}
	return tags, nil
}

func (r *ContactRepository) CreateContact(ctx context.Context, input entity.CreateContactInput) (*entity.Contact, error) {
	queries := db.New(r.pool)

	var extrasJSON []byte
	if input.Extras != nil && len(input.Extras) > 0 {
		var err error
		extrasJSON, err = json.Marshal(input.Extras)
		if err != nil {
			return nil, err
		}
	} else {
		extrasJSON = []byte("{}")
	}

	dbContact, err := queries.CreateContact(ctx, db.CreateContactParams{
		Name:     input.Name,
		Email:    input.Email,
		Phone:    input.Phone,
		Birthday: input.Birthday,
		Notes:    input.Notes,
		Extras:   extrasJSON,
	})
	if err != nil {
		return nil, err
	}

	var extras map[string]interface{}
	if len(dbContact.Extras) > 0 {
		if err := json.Unmarshal(dbContact.Extras, &extras); err != nil {
			extras = make(map[string]interface{})
		}
	} else {
		extras = make(map[string]interface{})
	}

	return &entity.Contact{
		ContactID:    dbContact.ContactID,
		Name:         dbContact.Name,
		Email:        dbContact.Email,
		Phone:        dbContact.Phone,
		Birthday:     dbContact.Birthday,
		Notes:        dbContact.Notes,
		Extras:       extras,
		CreationDate: dbContact.CreationDate.Time,
		LastUpdate:   dbContact.LastUpdate.Time,
	}, nil
}

func (r *ContactRepository) UpdateContact(ctx context.Context, contactID uuid.UUID, input entity.UpdateContactInput) error {
	queries := db.New(r.pool)

	// COALESCE in SQL handles NULL properly, so we need to provide a non-nil value
	// If input.Name is nil, we pass empty string which will be ignored by COALESCE
	name := ""
	if input.Name != nil {
		name = *input.Name
	}

	var extrasJSON []byte
	if input.Extras != nil {
		var err error
		extrasJSON, err = json.Marshal(input.Extras)
		if err != nil {
			return err
		}
	}

	return queries.UpdateContact(ctx, db.UpdateContactParams{
		ContactID: contactID,
		Name:      name,
		Email:     input.Email,
		Phone:     input.Phone,
		Birthday:  input.Birthday,
		Notes:     input.Notes,
		Extras:    extrasJSON,
	})
}

func (r *ContactRepository) DeleteContact(ctx context.Context, contactID uuid.UUID) error {
	queries := db.New(r.pool)
	return queries.DeleteContact(ctx, contactID)
}

func (r *ContactRepository) EvaluationExists(ctx context.Context, contactID uuid.UUID) (bool, error) {
	queries := db.New(r.pool)
	return queries.GetContactEvaluationExists(ctx, contactID)
}

func (r *ContactRepository) CreateEvaluation(ctx context.Context, contactID uuid.UUID, input entity.UpdateEvaluationInput) error {
	queries := db.New(r.pool)
	return queries.CreateContactEvaluation(ctx, db.CreateContactEvaluationParams{
		ContactID:  contactID,
		Importance: input.Importance,
		Closeness:  input.Closeness,
		Fondness:   input.Fondness,
	})
}

func (r *ContactRepository) UpdateEvaluation(ctx context.Context, contactID uuid.UUID, input entity.UpdateEvaluationInput) error {
	queries := db.New(r.pool)
	return queries.UpdateContactEvaluation(ctx, db.UpdateContactEvaluationParams{
		ContactID:  contactID,
		Importance: input.Importance,
		Closeness:  input.Closeness,
		Fondness:   input.Fondness,
	})
}

func (r *ContactRepository) MergeContacts(ctx context.Context, sourceID, targetID uuid.UUID) error {
	queries := db.New(r.pool)
	return queries.MergeContacts(ctx, db.MergeContactsParams{
		Column1: sourceID,
		Column2: targetID,
	})
}

func (r *ContactRepository) GetAllTagNames(ctx context.Context) ([]entity.ContactTagName, error) {
	queries := db.New(r.pool)
	dbTags, err := queries.GetAllTagNames(ctx)
	if err != nil {
		return nil, err
	}

	tags := make([]entity.ContactTagName, len(dbTags))
	for i, dbTag := range dbTags {
		tags[i] = entity.ContactTagName{
			TagID: dbTag.TagID,
			Name:  dbTag.Name,
			// CreatedAt not returned by query
		}
	}
	return tags, nil
}

func (r *ContactRepository) GetContactTagByName(ctx context.Context, tagName string) (*entity.ContactTagName, error) {
	queries := db.New(r.pool)
	dbTag, err := queries.GetContactTagByName(ctx, tagName)
	if err != nil {
		return nil, err
	}

	return &entity.ContactTagName{
		TagID: dbTag.TagID,
		Name:  dbTag.Name,
		// CreatedAt not returned by query
	}, nil
}

func (r *ContactRepository) CreateTagName(ctx context.Context, tagName string) (*entity.ContactTagName, error) {
	queries := db.New(r.pool)
	dbTag, err := queries.CreateTagName(ctx, tagName)
	if err != nil {
		return nil, err
	}

	return &entity.ContactTagName{
		TagID: dbTag.TagID,
		Name:  dbTag.Name,
		// CreatedAt not returned by query
	}, nil
}

func (r *ContactRepository) ContactTagExists(ctx context.Context, contactID, tagID uuid.UUID) (bool, error) {
	queries := db.New(r.pool)
	return queries.GetContactTagExists(ctx, db.GetContactTagExistsParams{
		ContactID: contactID,
		TagID:     tagID,
	})
}

func (r *ContactRepository) AddContactTag(ctx context.Context, contactID, tagID uuid.UUID) error {
	queries := db.New(r.pool)
	return queries.AddContactTag(ctx, db.AddContactTagParams{
		ContactID: contactID,
		TagID:     tagID,
	})
}

func (r *ContactRepository) RemoveContactTag(ctx context.Context, contactID, tagID uuid.UUID) error {
	queries := db.New(r.pool)
	return queries.RemoveContactTag(ctx, db.RemoveContactTagParams{
		ContactID: contactID,
		TagID:     tagID,
	})
}

func (r *ContactRepository) EnsureContactStatsExist(ctx context.Context) error {
	queries := db.New(r.pool)
	return queries.EnsureContactStatsExist(ctx)
}

func (r *ContactRepository) UpdateContactMessageStats(ctx context.Context) error {
	queries := db.New(r.pool)
	return queries.UpdateContactMessageStats(ctx)
}

func (r *ContactRepository) ListAllContactSources(ctx context.Context) ([]entity.ContactSource, error) {
	queries := db.New(r.pool)
	dbSources, err := queries.ListAllContactSources(ctx)
	if err != nil {
		return nil, err
	}

	sources := make([]entity.ContactSource, len(dbSources))
	for i, dbSource := range dbSources {
		var contactIDPtr *uuid.UUID
		if dbSource.ContactID.Valid {
			id := uuid.UUID(dbSource.ContactID.Bytes)
			contactIDPtr = &id
		}
		sources[i] = entity.ContactSource{
			ID:         dbSource.ID,
			ContactID:  contactIDPtr,
			SourceID:   dbSource.SourceID,
			SourceName: dbSource.SourceName,
		}
	}
	return sources, nil
}

func (r *ContactRepository) CreateContactSource(ctx context.Context, input entity.CreateContactSourceInput) (*entity.ContactSource, error) {
	queries := db.New(r.pool)
	var contactID pgtype.UUID
	if input.ContactID != uuid.Nil {
		contactID = pgtype.UUID{Bytes: input.ContactID, Valid: true}
	}
	dbSource, err := queries.CreateContactSource(ctx, db.CreateContactSourceParams{
		ContactID:  contactID,
		SourceID:   input.SourceID,
		SourceName: input.SourceName,
	})
	if err != nil {
		return nil, err
	}

	var contactIDPtr *uuid.UUID
	if dbSource.ContactID.Valid {
		id := uuid.UUID(dbSource.ContactID.Bytes)
		contactIDPtr = &id
	}

	return &entity.ContactSource{
		ID:         dbSource.ID,
		ContactID:  contactIDPtr,
		SourceID:   dbSource.SourceID,
		SourceName: dbSource.SourceName,
	}, nil
}

func (r *ContactRepository) UpdateContactSource(ctx context.Context, id uuid.UUID, input entity.UpdateContactSourceInput) (*entity.ContactSource, error) {
	queries := db.New(r.pool)
	dbSource, err := queries.UpdateContactSource(ctx, db.UpdateContactSourceParams{
		ID:         id,
		SourceID:   input.SourceID,
		SourceName: input.SourceName,
	})
	if err != nil {
		return nil, err
	}

	var contactIDPtr *uuid.UUID
	if dbSource.ContactID.Valid {
		id := uuid.UUID(dbSource.ContactID.Bytes)
		contactIDPtr = &id
	}

	return &entity.ContactSource{
		ID:         dbSource.ID,
		ContactID:  contactIDPtr,
		SourceID:   dbSource.SourceID,
		SourceName: dbSource.SourceName,
	}, nil
}

func (r *ContactRepository) DeleteContactSource(ctx context.Context, id uuid.UUID) error {
	queries := db.New(r.pool)
	return queries.DeleteContactSource(ctx, id)
}

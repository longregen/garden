package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"garden3/internal/adapter/secondary/postgres/generated/db"
	"garden3/internal/domain/entity"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// EntityRepository implements the output.EntityRepository interface
type EntityRepository struct {
	pool *pgxpool.Pool
}

// NewEntityRepository creates a new entity repository
func NewEntityRepository(pool *pgxpool.Pool) *EntityRepository {
	return &EntityRepository{
		pool: pool,
	}
}

func (r *EntityRepository) GetEntity(ctx context.Context, entityID uuid.UUID) (*entity.Entity, error) {
	queries := db.New(r.pool)
	dbEntity, err := queries.GetEntity(ctx, entityID)
	if err != nil {
		return nil, err
	}

	return &entity.Entity{
		EntityID:    dbEntity.EntityID,
		Name:        dbEntity.Name,
		Type:        dbEntity.Type,
		Description: dbEntity.Description,
		Properties:  json.RawMessage(dbEntity.Properties),
		CreatedAt:   dbEntity.CreatedAt.Time,
		UpdatedAt:   dbEntity.UpdatedAt.Time,
		DeletedAt:   convertPgTimestampToTimePtr(dbEntity.DeletedAt),
	}, nil
}

func (r *EntityRepository) UpdateEntity(ctx context.Context, entityID uuid.UUID, input entity.UpdateEntityInput) (*entity.Entity, error) {
	queries := db.New(r.pool)

	// Prepare parameters - COALESCE in SQL expects non-nil values
	// Use empty values as defaults which will be ignored by COALESCE
	name := ""
	if input.Name != nil {
		name = *input.Name
	}

	entityType := ""
	if input.Type != nil {
		entityType = *input.Type
	}

	var properties []byte
	if input.Properties != nil {
		properties = []byte(*input.Properties)
	}

	dbEntity, err := queries.UpdateEntity(ctx, db.UpdateEntityParams{
		EntityID:    entityID,
		Name:        name,
		Type:        entityType,
		Description: input.Description,
		Properties:  properties,
	})
	if err != nil {
		return nil, err
	}

	return &entity.Entity{
		EntityID:    dbEntity.EntityID,
		Name:        dbEntity.Name,
		Type:        dbEntity.Type,
		Description: dbEntity.Description,
		Properties:  json.RawMessage(dbEntity.Properties),
		CreatedAt:   dbEntity.CreatedAt.Time,
		UpdatedAt:   dbEntity.UpdatedAt.Time,
		DeletedAt:   convertPgTimestampToTimePtr(dbEntity.DeletedAt),
	}, nil
}

func (r *EntityRepository) DeleteEntity(ctx context.Context, entityID uuid.UUID) error {
	queries := db.New(r.pool)
	return queries.DeleteEntity(ctx, entityID)
}

func (r *EntityRepository) GetEntityRelationships(ctx context.Context, entityID uuid.UUID) ([]entity.EntityRelationship, error) {
	queries := db.New(r.pool)
	dbRels, err := queries.GetEntityRelationshipsByEntity(ctx, entityID)
	if err != nil {
		return nil, err
	}

	rels := make([]entity.EntityRelationship, len(dbRels))
	for i, dbRel := range dbRels {
		rels[i] = entity.EntityRelationship{
			ID:               dbRel.ID,
			EntityID:         dbRel.EntityID,
			RelatedType:      dbRel.RelatedType,
			RelatedID:        dbRel.RelatedID,
			RelationshipType: dbRel.RelationshipType,
			Metadata:         json.RawMessage(dbRel.Metadata),
			CreatedAt:        dbRel.CreatedAt.Time,
			UpdatedAt:        dbRel.UpdatedAt.Time,
		}
	}
	return rels, nil
}

func (r *EntityRepository) GetEntityRelationshipsByTypeAndRelationship(ctx context.Context, entityID uuid.UUID, relatedType, relationshipType string) ([]entity.EntityRelationship, error) {
	queries := db.New(r.pool)
	dbRels, err := queries.GetEntityRelationshipsByTypeAndRelationship(ctx, db.GetEntityRelationshipsByTypeAndRelationshipParams{
		EntityID:         entityID,
		RelatedType:      relatedType,
		RelationshipType: relationshipType,
	})
	if err != nil {
		return nil, err
	}

	rels := make([]entity.EntityRelationship, len(dbRels))
	for i, dbRel := range dbRels {
		rels[i] = entity.EntityRelationship{
			ID:               dbRel.ID,
			EntityID:         dbRel.EntityID,
			RelatedType:      dbRel.RelatedType,
			RelatedID:        dbRel.RelatedID,
			RelationshipType: dbRel.RelationshipType,
			Metadata:         json.RawMessage(dbRel.Metadata),
			CreatedAt:        dbRel.CreatedAt.Time,
			UpdatedAt:        dbRel.UpdatedAt.Time,
		}
	}
	return rels, nil
}

func (r *EntityRepository) CreateEntityRelationship(ctx context.Context, entityID uuid.UUID, relatedType string, relatedID uuid.UUID, relationshipType string, metadata []byte) (*entity.EntityRelationship, error) {
	queries := db.New(r.pool)
	dbRel, err := queries.CreateEntityRelationship(ctx, db.CreateEntityRelationshipParams{
		EntityID:         entityID,
		RelatedType:      relatedType,
		RelatedID:        relatedID,
		RelationshipType: relationshipType,
		Metadata:         metadata,
	})
	if err != nil {
		return nil, err
	}

	return &entity.EntityRelationship{
		ID:               dbRel.ID,
		EntityID:         dbRel.EntityID,
		RelatedType:      dbRel.RelatedType,
		RelatedID:        dbRel.RelatedID,
		RelationshipType: dbRel.RelationshipType,
		Metadata:         json.RawMessage(dbRel.Metadata),
		CreatedAt:        dbRel.CreatedAt.Time,
		UpdatedAt:        dbRel.UpdatedAt.Time,
	}, nil
}

func (r *EntityRepository) DeleteEntityRelationship(ctx context.Context, relationshipID uuid.UUID) error {
	queries := db.New(r.pool)
	return queries.DeleteEntityRelationship(ctx, relationshipID)
}

func (r *EntityRepository) UpdateContactNameByRelationship(ctx context.Context, entityID uuid.UUID, relationshipType, name string) error {
	queries := db.New(r.pool)
	return queries.UpdateContactNameByRelationship(ctx, db.UpdateContactNameByRelationshipParams{
		EntityID:         entityID,
		RelationshipType: relationshipType,
		Name:             name,
	})
}

func (r *EntityRepository) GetEntityReferences(ctx context.Context, entityID uuid.UUID, sourceType *string) ([]entity.EntityReference, error) {
	queries := db.New(r.pool)

	var dbRefs []db.EntityReference
	var err error

	if sourceType != nil {
		dbRefs, err = queries.GetEntityReferencesBySourceType(ctx, db.GetEntityReferencesBySourceTypeParams{
			EntityID:   entityID,
			SourceType: *sourceType,
		})
	} else {
		dbRefs, err = queries.GetEntityReferences(ctx, entityID)
	}

	if err != nil {
		return nil, err
	}

	refs := make([]entity.EntityReference, len(dbRefs))
	for i, dbRef := range dbRefs {
		var position *int
		if dbRef.Position != nil {
			pos := int(*dbRef.Position)
			position = &pos
		}

		refs[i] = entity.EntityReference{
			ID:            dbRef.ID,
			SourceType:    dbRef.SourceType,
			SourceID:      dbRef.SourceID,
			EntityID:      dbRef.EntityID,
			ReferenceText: dbRef.ReferenceText,
			Position:      position,
			CreatedAt:     dbRef.CreatedAt.Time,
		}
	}

	return refs, nil
}

func (r *EntityRepository) ListEntities(ctx context.Context, entityType string, updatedSince *string) ([]entity.Entity, error) {
	queries := db.New(r.pool)

	var updatedSinceTime pgtype.Timestamp
	if updatedSince != nil && *updatedSince != "" {
		// Parse the timestamp string if needed
		// For now, leaving as invalid timestamp to pass NULL
	}

	dbEntities, err := queries.ListEntities(ctx, db.ListEntitiesParams{
		Column1: entityType,
		Column2: updatedSinceTime,
	})
	if err != nil {
		return nil, err
	}

	entities := make([]entity.Entity, len(dbEntities))
	for i, dbEntity := range dbEntities {
		entities[i] = entity.Entity{
			EntityID:    dbEntity.EntityID,
			Name:        dbEntity.Name,
			Type:        dbEntity.Type,
			Description: dbEntity.Description,
			Properties:  json.RawMessage(dbEntity.Properties),
			CreatedAt:   dbEntity.CreatedAt.Time,
			UpdatedAt:   dbEntity.UpdatedAt.Time,
			DeletedAt:   convertPgTimestampToTimePtr(dbEntity.DeletedAt),
		}
	}
	return entities, nil
}

func (r *EntityRepository) CreateEntity(ctx context.Context, input entity.CreateEntityInput) (*entity.Entity, error) {
	queries := db.New(r.pool)

	var properties []byte
	if input.Properties != nil {
		properties = []byte(*input.Properties)
	}

	dbEntity, err := queries.CreateEntity(ctx, db.CreateEntityParams{
		Name:        input.Name,
		Type:        input.Type,
		Description: input.Description,
		Properties:  properties,
	})
	if err != nil {
		return nil, err
	}

	return &entity.Entity{
		EntityID:    dbEntity.EntityID,
		Name:        dbEntity.Name,
		Type:        dbEntity.Type,
		Description: dbEntity.Description,
		Properties:  json.RawMessage(dbEntity.Properties),
		CreatedAt:   dbEntity.CreatedAt.Time,
		UpdatedAt:   dbEntity.UpdatedAt.Time,
		DeletedAt:   convertPgTimestampToTimePtr(dbEntity.DeletedAt),
	}, nil
}

func (r *EntityRepository) SearchEntities(ctx context.Context, query string, entityType string) ([]entity.Entity, error) {
	queries := db.New(r.pool)

	dbEntities, err := queries.SearchEntities(ctx, db.SearchEntitiesParams{
		Column1: &query,
		Column2: entityType,
	})
	if err != nil {
		return nil, err
	}

	entities := make([]entity.Entity, len(dbEntities))
	for i, dbEntity := range dbEntities {
		entities[i] = entity.Entity{
			EntityID:    dbEntity.EntityID,
			Name:        dbEntity.Name,
			Type:        dbEntity.Type,
			Description: dbEntity.Description,
			Properties:  json.RawMessage(dbEntity.Properties),
			CreatedAt:   dbEntity.CreatedAt.Time,
			UpdatedAt:   dbEntity.UpdatedAt.Time,
			DeletedAt:   convertPgTimestampToTimePtr(dbEntity.DeletedAt),
		}
	}
	return entities, nil
}

func (r *EntityRepository) ListDeletedEntities(ctx context.Context) ([]entity.Entity, error) {
	queries := db.New(r.pool)

	dbEntities, err := queries.ListDeletedEntities(ctx)
	if err != nil {
		return nil, err
	}

	entities := make([]entity.Entity, len(dbEntities))
	for i, dbEntity := range dbEntities {
		entities[i] = entity.Entity{
			EntityID:    dbEntity.EntityID,
			Name:        dbEntity.Name,
			Type:        dbEntity.Type,
			Description: dbEntity.Description,
			Properties:  json.RawMessage(dbEntity.Properties),
			CreatedAt:   dbEntity.CreatedAt.Time,
			UpdatedAt:   dbEntity.UpdatedAt.Time,
			DeletedAt:   convertPgTimestampToTimePtr(dbEntity.DeletedAt),
		}
	}
	return entities, nil
}

func (r *EntityRepository) ListEntitiesUpdatedSince(ctx context.Context, since time.Time) ([]entity.Entity, error) {
	query := `
		SELECT entity_id, name, type, description, properties, created_at, updated_at, deleted_at
		FROM entities
		WHERE deleted_at IS NULL AND updated_at > $1
		ORDER BY updated_at DESC
	`

	rows, err := r.pool.Query(ctx, query, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entities []entity.Entity
	for rows.Next() {
		var e entity.Entity
		var createdAt, updatedAt pgtype.Timestamp
		var deletedAt pgtype.Timestamp

		err := rows.Scan(
			&e.EntityID,
			&e.Name,
			&e.Type,
			&e.Description,
			&e.Properties,
			&createdAt,
			&updatedAt,
			&deletedAt,
		)
		if err != nil {
			return nil, err
		}

		e.CreatedAt = createdAt.Time
		e.UpdatedAt = updatedAt.Time
		e.DeletedAt = convertPgTimestampToTimePtr(deletedAt)

		entities = append(entities, e)
	}

	return entities, rows.Err()
}

func (r *EntityRepository) ListEntitiesByProperty(ctx context.Context, propertyKey string, propertyValue string) ([]entity.Entity, error) {
	var query string
	var args []interface{}

	if propertyValue == "" {
		// Query for entities that have the property key (regardless of value)
		query = `
			SELECT entity_id, name, type, description, properties, created_at, updated_at, deleted_at
			FROM entities
			WHERE deleted_at IS NULL AND properties::text LIKE $1
			ORDER BY updated_at DESC
		`
		args = append(args, "%\""+propertyKey+"\":%")
	} else {
		// Query for entities that have the property key with a specific value
		query = `
			SELECT entity_id, name, type, description, properties, created_at, updated_at, deleted_at
			FROM entities
			WHERE deleted_at IS NULL
			  AND properties::jsonb @> $1::jsonb
			ORDER BY updated_at DESC
		`
		jsonFilter := fmt.Sprintf(`{"%s": "%s"}`, propertyKey, propertyValue)
		args = append(args, jsonFilter)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entities []entity.Entity
	for rows.Next() {
		var e entity.Entity
		var createdAt, updatedAt pgtype.Timestamp
		var deletedAt pgtype.Timestamp

		err := rows.Scan(
			&e.EntityID,
			&e.Name,
			&e.Type,
			&e.Description,
			&e.Properties,
			&createdAt,
			&updatedAt,
			&deletedAt,
		)
		if err != nil {
			return nil, err
		}

		e.CreatedAt = createdAt.Time
		e.UpdatedAt = updatedAt.Time
		e.DeletedAt = convertPgTimestampToTimePtr(deletedAt)

		entities = append(entities, e)
	}

	return entities, rows.Err()
}

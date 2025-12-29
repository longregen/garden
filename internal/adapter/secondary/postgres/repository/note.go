package repository

import (
	"context"
	"encoding/json"
	"fmt"

	db "garden3/internal/adapter/secondary/postgres/generated/db"
	"garden3/internal/domain/entity"
	"garden3/internal/port/output"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
)

// NoteRepository implements the output.NoteRepository interface
type NoteRepository struct {
	pool *pgxpool.Pool
	tx   pgx.Tx
}

// NewNoteRepository creates a new note repository
func NewNoteRepository(pool *pgxpool.Pool) *NoteRepository {
	return &NoteRepository{
		pool: pool,
	}
}

// BeginTx starts a new transaction
func (r *NoteRepository) BeginTx(ctx context.Context) (interface{}, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// CommitTx commits a transaction
func (r *NoteRepository) CommitTx(ctx context.Context, tx interface{}) error {
	pgxTx, ok := tx.(pgx.Tx)
	if !ok {
		return fmt.Errorf("invalid transaction type")
	}
	return pgxTx.Commit(ctx)
}

// RollbackTx rolls back a transaction
func (r *NoteRepository) RollbackTx(ctx context.Context, tx interface{}) error {
	pgxTx, ok := tx.(pgx.Tx)
	if !ok {
		return fmt.Errorf("invalid transaction type")
	}
	return pgxTx.Rollback(ctx)
}

// WithTx returns a new repository with the given transaction
func (r *NoteRepository) WithTx(ctx context.Context, tx interface{}) output.NoteRepository {
	pgxTx, ok := tx.(pgx.Tx)
	if !ok {
		return r
	}
	return &NoteRepository{
		pool: r.pool,
		tx:   pgxTx,
	}
}

func (r *NoteRepository) getQuerier() db.DBTX {
	if r.tx != nil {
		return r.tx
	}
	return r.pool
}

func (r *NoteRepository) GetNote(ctx context.Context, noteID uuid.UUID) (*entity.Note, error) {
	queries := db.New(r.getQuerier())
	dbNote, err := queries.GetNote(ctx, noteID)
	if err != nil {
		return nil, err
	}

	created := int64(0)
	if dbNote.Created != nil {
		created = *dbNote.Created
	}
	modified := int64(0)
	if dbNote.Modified != nil {
		modified = *dbNote.Modified
	}

	return &entity.Note{
		ID:       dbNote.ID,
		Title:    dbNote.Title,
		Slug:     dbNote.Slug,
		Contents: dbNote.Contents,
		Created:  created,
		Modified: modified,
	}, nil
}

func (r *NoteRepository) GetNoteWithTags(ctx context.Context, noteID uuid.UUID) (*entity.Note, []string, error) {
	queries := db.New(r.getQuerier())
	dbNote, err := queries.GetNoteWithTags(ctx, noteID)
	if err != nil {
		return nil, nil, err
	}

	created := int64(0)
	if dbNote.Created != nil {
		created = *dbNote.Created
	}
	modified := int64(0)
	if dbNote.Modified != nil {
		modified = *dbNote.Modified
	}

	note := &entity.Note{
		ID:       dbNote.ID,
		Title:    dbNote.Title,
		Slug:     dbNote.Slug,
		Contents: dbNote.Contents,
		Created:  created,
		Modified: modified,
	}

	tags := convertInterfaceToStringSlice(dbNote.Tags)

	return note, tags, nil
}

func (r *NoteRepository) ListNotes(ctx context.Context, limit, offset int32) ([]entity.NoteListItem, error) {
	queries := db.New(r.getQuerier())
	dbNotes, err := queries.ListNotes(ctx, db.ListNotesParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	notes := make([]entity.NoteListItem, len(dbNotes))
	for i, dbNote := range dbNotes {
		created := int64(0)
		if dbNote.Created != nil {
			created = *dbNote.Created
		}
		modified := int64(0)
		if dbNote.Modified != nil {
			modified = *dbNote.Modified
		}

		tags := convertInterfaceToStringSlice(dbNote.Tags)

		notes[i] = entity.NoteListItem{
			ID:       dbNote.ID,
			Title:    dbNote.Title,
			Tags:     tags,
			Created:  created,
			Modified: modified,
		}
	}
	return notes, nil
}

func (r *NoteRepository) SearchNotes(ctx context.Context, searchPattern string, limit, offset int32) ([]entity.NoteListItem, error) {
	queries := db.New(r.getQuerier())
	dbNotes, err := queries.SearchNotes(ctx, db.SearchNotesParams{
		Title:  &searchPattern,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	notes := make([]entity.NoteListItem, len(dbNotes))
	for i, dbNote := range dbNotes {
		created := int64(0)
		if dbNote.Created != nil {
			created = *dbNote.Created
		}
		modified := int64(0)
		if dbNote.Modified != nil {
			modified = *dbNote.Modified
		}

		tags := convertInterfaceToStringSlice(dbNote.Tags)

		notes[i] = entity.NoteListItem{
			ID:       dbNote.ID,
			Title:    dbNote.Title,
			Tags:     tags,
			Created:  created,
			Modified: modified,
		}
	}
	return notes, nil
}

func (r *NoteRepository) CountNotes(ctx context.Context) (int64, error) {
	queries := db.New(r.getQuerier())
	return queries.CountNotes(ctx)
}

func (r *NoteRepository) CountSearchNotes(ctx context.Context, searchPattern string) (int64, error) {
	queries := db.New(r.getQuerier())
	return queries.CountSearchNotes(ctx, &searchPattern)
}

func (r *NoteRepository) CreateNote(ctx context.Context, title, slug, contents string) (*entity.Note, error) {
	queries := db.New(r.getQuerier())

	titlePtr := &title
	slugPtr := &slug
	contentsPtr := &contents

	dbNote, err := queries.CreateNote(ctx, db.CreateNoteParams{
		Title:    titlePtr,
		Slug:     slugPtr,
		Contents: contentsPtr,
	})
	if err != nil {
		return nil, err
	}

	created := int64(0)
	if dbNote.Created != nil {
		created = *dbNote.Created
	}
	modified := int64(0)
	if dbNote.Modified != nil {
		modified = *dbNote.Modified
	}

	return &entity.Note{
		ID:       dbNote.ID,
		Title:    dbNote.Title,
		Slug:     dbNote.Slug,
		Contents: dbNote.Contents,
		Created:  created,
		Modified: modified,
	}, nil
}

func (r *NoteRepository) UpdateNote(ctx context.Context, noteID uuid.UUID, title, contents string) error {
	queries := db.New(r.getQuerier())

	titlePtr := &title
	contentsPtr := &contents

	return queries.UpdateNote(ctx, db.UpdateNoteParams{
		ID:       noteID,
		Title:    titlePtr,
		Contents: contentsPtr,
	})
}

func (r *NoteRepository) DeleteNote(ctx context.Context, noteID uuid.UUID) error {
	queries := db.New(r.getQuerier())
	return queries.DeleteNote(ctx, noteID)
}

func (r *NoteRepository) GetNoteEntityRelationship(ctx context.Context, noteID uuid.UUID) (*uuid.UUID, error) {
	queries := db.New(r.getQuerier())
	entityID, err := queries.GetNoteEntityRelationship(ctx, noteID)
	if err != nil {
		return nil, err
	}
	return &entityID, nil
}

func (r *NoteRepository) CreateEntity(ctx context.Context, name, entityType string, description *string, properties map[string]interface{}) (*uuid.UUID, error) {
	queries := db.New(r.getQuerier())

	propertiesJSON, err := json.Marshal(properties)
	if err != nil {
		return nil, err
	}

	result, err := queries.CreateEntityForNote(ctx, db.CreateEntityForNoteParams{
		Name:        name,
		Type:        entityType,
		Description: description,
		Properties:  propertiesJSON,
	})
	if err != nil {
		return nil, err
	}

	return &result.EntityID, nil
}

func (r *NoteRepository) CreateEntityRelationship(ctx context.Context, entityID uuid.UUID, relatedType string, relatedID uuid.UUID, relationshipType string, metadata map[string]interface{}) error {
	queries := db.New(r.getQuerier())

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	return queries.CreateEntityRelationshipForNote(ctx, db.CreateEntityRelationshipForNoteParams{
		EntityID:         entityID,
		RelatedType:      relatedType,
		RelatedID:        relatedID,
		RelationshipType: relationshipType,
		Metadata:         metadataJSON,
	})
}

func (r *NoteRepository) DeleteEntityRelationships(ctx context.Context, entityID uuid.UUID) error {
	queries := db.New(r.getQuerier())
	return queries.DeleteEntityRelationships(ctx, entityID)
}

func (r *NoteRepository) SoftDeleteEntity(ctx context.Context, entityID uuid.UUID) error {
	queries := db.New(r.getQuerier())
	return queries.SoftDeleteEntity(ctx, entityID)
}

func (r *NoteRepository) GetEntityByID(ctx context.Context, entityID uuid.UUID) (*entity.Entity, error) {
	queries := db.New(r.getQuerier())
	dbEntity, err := queries.GetEntityByID(ctx, entityID)
	if err != nil {
		return nil, err
	}

	var properties map[string]interface{}
	if err := json.Unmarshal(dbEntity.Properties, &properties); err != nil {
		properties = make(map[string]interface{})
	}

	return &entity.Entity{
		EntityID:    dbEntity.EntityID,
		Name:        dbEntity.Name,
		Type:        dbEntity.Type,
		Description: dbEntity.Description,
		Properties:  convertMapToJSON(properties),
	}, nil
}

func (r *NoteRepository) GetEntityByName(ctx context.Context, name string) (*entity.Entity, error) {
	queries := db.New(r.getQuerier())
	dbEntity, err := queries.GetEntityByName(ctx, name)
	if err != nil {
		return nil, err
	}

	var properties map[string]interface{}
	if err := json.Unmarshal(dbEntity.Properties, &properties); err != nil {
		properties = make(map[string]interface{})
	}

	return &entity.Entity{
		EntityID:    dbEntity.EntityID,
		Name:        dbEntity.Name,
		Type:        dbEntity.Type,
		Description: dbEntity.Description,
		Properties:  convertMapToJSON(properties),
	}, nil
}

func (r *NoteRepository) CreateEntityReference(ctx context.Context, ref entity.EntityReference) error {
	queries := db.New(r.getQuerier())
	var position *int32
	if ref.Position != nil {
		pos := int32(*ref.Position)
		position = &pos
	}
	return queries.CreateEntityReference(ctx, db.CreateEntityReferenceParams{
		SourceType:    ref.SourceType,
		SourceID:      ref.SourceID,
		EntityID:      ref.EntityID,
		ReferenceText: ref.ReferenceText,
		Position:      position,
	})
}

func (r *NoteRepository) DeleteNoteEntityReferences(ctx context.Context, noteID uuid.UUID) error {
	queries := db.New(r.getQuerier())
	return queries.DeleteNoteEntityReferences(ctx, noteID)
}

func (r *NoteRepository) GetNoteTagByName(ctx context.Context, tagName string) (*entity.NoteTag, error) {
	queries := db.New(r.getQuerier())
	dbTag, err := queries.GetNoteTagByName(ctx, tagName)
	if err != nil {
		return nil, err
	}

	created := int64(0)
	if dbTag.Created != nil {
		created = *dbTag.Created
	}
	modified := int64(0)
	if dbTag.Modified != nil {
		modified = *dbTag.Modified
	}

	return &entity.NoteTag{
		ID:           dbTag.ID,
		Name:         dbTag.Name,
		Created:      created,
		Modified:     modified,
		LastActivity: dbTag.LastActivity,
	}, nil
}

func (r *NoteRepository) UpsertNoteTag(ctx context.Context, tagName string, lastActivity int64) (*entity.NoteTag, error) {
	queries := db.New(r.getQuerier())
	dbTag, err := queries.UpsertNoteTag(ctx, db.UpsertNoteTagParams{
		Name:         tagName,
		LastActivity: &lastActivity,
	})
	if err != nil {
		return nil, err
	}

	created := int64(0)
	if dbTag.Created != nil {
		created = *dbTag.Created
	}
	modified := int64(0)
	if dbTag.Modified != nil {
		modified = *dbTag.Modified
	}

	return &entity.NoteTag{
		ID:           dbTag.ID,
		Name:         dbTag.Name,
		Created:      created,
		Modified:     modified,
		LastActivity: dbTag.LastActivity,
	}, nil
}

func (r *NoteRepository) CreateItemTag(ctx context.Context, itemID, tagID uuid.UUID) error {
	queries := db.New(r.getQuerier())
	return queries.CreateItemTag(ctx, db.CreateItemTagParams{
		ItemID: itemID,
		TagID:  tagID,
	})
}

func (r *NoteRepository) DeleteNoteItemTags(ctx context.Context, noteID uuid.UUID) error {
	queries := db.New(r.getQuerier())
	return queries.DeleteNoteItemTags(ctx, noteID)
}

func (r *NoteRepository) GetAllNoteTags(ctx context.Context) ([]entity.NoteTag, error) {
	queries := db.New(r.getQuerier())
	dbTags, err := queries.GetAllNoteTags(ctx)
	if err != nil {
		return nil, err
	}

	tags := make([]entity.NoteTag, len(dbTags))
	for i, dbTag := range dbTags {
		created := int64(0)
		if dbTag.Created != nil {
			created = *dbTag.Created
		}
		modified := int64(0)
		if dbTag.Modified != nil {
			modified = *dbTag.Modified
		}

		tags[i] = entity.NoteTag{
			ID:           dbTag.ID,
			Name:         dbTag.Name,
			Created:      created,
			Modified:     modified,
			LastActivity: dbTag.LastActivity,
		}
	}
	return tags, nil
}

func (r *NoteRepository) DeleteNoteSemanticIndex(ctx context.Context, noteID uuid.UUID) error {
	queries := db.New(r.getQuerier())

	// Convert uuid.UUID to pgtype.UUID
	pgUUID := pgtype.UUID{
		Bytes: noteID,
		Valid: true,
	}

	return queries.DeleteNoteSemanticIndex(ctx, pgUUID)
}

func (r *NoteRepository) SearchSimilarNotes(ctx context.Context, embedding []float32, strategy string, limit int32) ([]entity.NoteListItem, error) {
	// Manual implementation since sqlc didn't generate this query
	// Using raw SQL with pgx to handle pgvector types

	query := `
		SELECT
			i.id,
			i.title,
			i.created,
			i.modified,
			array_agg(DISTINCT t.name) FILTER (WHERE t.name IS NOT NULL) as tags
		FROM items i
		INNER JOIN item_semantic_index isi ON i.id = isi.item_id
		LEFT JOIN item_tags it ON i.id = it.item_id
		LEFT JOIN tags t ON it.tag_id = t.id
		WHERE isi.strategy = $1
		GROUP BY i.id, i.title, i.created, i.modified, isi.embedding
		ORDER BY isi.embedding <=> $2::vector
		LIMIT $3
	`

	// Convert []float32 to pgvector.Vector
	vec := pgvector.NewVector(embedding)

	rows, err := r.pool.Query(ctx, query, strategy, vec, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search similar notes: %w", err)
	}
	defer rows.Close()

	var notes []entity.NoteListItem
	for rows.Next() {
		var note entity.NoteListItem
		var created, modified *int64

		err := rows.Scan(
			&note.ID,
			&note.Title,
			&created,
			&modified,
			&note.Tags,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan note row: %w", err)
		}

		if created != nil {
			note.Created = *created
		}
		if modified != nil {
			note.Modified = *modified
		}
		if note.Tags == nil {
			note.Tags = []string{}
		}

		notes = append(notes, note)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating note rows: %w", err)
	}

	return notes, nil
}

package repository

import (
	"context"

	db "garden3/internal/adapter/secondary/postgres/generated/db"
	"garden3/internal/domain/entity"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
)

// ItemRepository implements the output.ItemRepository interface
type ItemRepository struct {
	pool *pgxpool.Pool
}

// NewItemRepository creates a new item repository
func NewItemRepository(pool *pgxpool.Pool) *ItemRepository {
	return &ItemRepository{
		pool: pool,
	}
}

func (r *ItemRepository) GetItem(ctx context.Context, itemID uuid.UUID) (*entity.Item, error) {
	queries := db.New(r.pool)
	dbItem, err := queries.GetItemByID(ctx, itemID)
	if err != nil {
		return nil, err
	}

	created := int64(0)
	if dbItem.Created != nil {
		created = *dbItem.Created
	}
	modified := int64(0)
	if dbItem.Modified != nil {
		modified = *dbItem.Modified
	}

	return &entity.Item{
		ID:       dbItem.ID,
		Title:    dbItem.Title,
		Slug:     dbItem.Slug,
		Contents: dbItem.Contents,
		Created:  created,
		Modified: modified,
	}, nil
}

func (r *ItemRepository) GetItemWithTags(ctx context.Context, itemID uuid.UUID) (*entity.Item, []string, error) {
	queries := db.New(r.pool)
	dbItem, err := queries.GetItemWithTagsByID(ctx, itemID)
	if err != nil {
		return nil, nil, err
	}

	created := int64(0)
	if dbItem.Created != nil {
		created = *dbItem.Created
	}
	modified := int64(0)
	if dbItem.Modified != nil {
		modified = *dbItem.Modified
	}

	var tags []string
	if dbItem.Tags != nil {
		if tagSlice, ok := dbItem.Tags.([]interface{}); ok {
			tags = make([]string, len(tagSlice))
			for i, t := range tagSlice {
				if s, ok := t.(string); ok {
					tags[i] = s
				}
			}
		}
	}
	if tags == nil {
		tags = []string{}
	}

	item := &entity.Item{
		ID:       dbItem.ID,
		Title:    dbItem.Title,
		Slug:     dbItem.Slug,
		Contents: dbItem.Contents,
		Created:  created,
		Modified: modified,
	}

	return item, tags, nil
}

func (r *ItemRepository) ListItems(ctx context.Context, limit, offset int32) ([]entity.ItemListItem, error) {
	queries := db.New(r.pool)
	dbItems, err := queries.ListAllItems(ctx, db.ListAllItemsParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	items := make([]entity.ItemListItem, 0, len(dbItems))
	for _, dbItem := range dbItems {
		created := int64(0)
		if dbItem.Created != nil {
			created = *dbItem.Created
		}
		modified := int64(0)
		if dbItem.Modified != nil {
			modified = *dbItem.Modified
		}

		var tags []string
		if dbItem.Tags != nil {
			if tagSlice, ok := dbItem.Tags.([]interface{}); ok {
				tags = make([]string, len(tagSlice))
				for i, t := range tagSlice {
					if s, ok := t.(string); ok {
						tags[i] = s
					}
				}
			}
		}
		if tags == nil {
			tags = []string{}
		}

		items = append(items, entity.ItemListItem{
			ID:       dbItem.ID,
			Title:    dbItem.Title,
			Tags:     tags,
			Created:  created,
			Modified: modified,
		})
	}

	return items, nil
}

func (r *ItemRepository) CountItems(ctx context.Context) (int64, error) {
	queries := db.New(r.pool)
	return queries.CountAllItems(ctx)
}

func (r *ItemRepository) CreateItem(ctx context.Context, title, slug, contents string) (*entity.Item, error) {
	queries := db.New(r.pool)
	dbItem, err := queries.CreateItemRecord(ctx, db.CreateItemRecordParams{
		Title:    &title,
		Slug:     &slug,
		Contents: &contents,
	})
	if err != nil {
		return nil, err
	}

	created := int64(0)
	if dbItem.Created != nil {
		created = *dbItem.Created
	}
	modified := int64(0)
	if dbItem.Modified != nil {
		modified = *dbItem.Modified
	}

	return &entity.Item{
		ID:       dbItem.ID,
		Title:    dbItem.Title,
		Slug:     dbItem.Slug,
		Contents: dbItem.Contents,
		Created:  created,
		Modified: modified,
	}, nil
}

func (r *ItemRepository) UpdateItem(ctx context.Context, itemID uuid.UUID, title, contents string) error {
	queries := db.New(r.pool)
	return queries.UpdateItemRecord(ctx, db.UpdateItemRecordParams{
		ID:       itemID,
		Title:    &title,
		Contents: &contents,
	})
}

func (r *ItemRepository) DeleteItem(ctx context.Context, itemID uuid.UUID) error {
	queries := db.New(r.pool)
	return queries.DeleteItemRecord(ctx, itemID)
}

func (r *ItemRepository) GetItemTagByName(ctx context.Context, tagName string) (*entity.ItemTag, error) {
	queries := db.New(r.pool)
	dbTag, err := queries.GetItemTagByTagName(ctx, tagName)
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

	return &entity.ItemTag{
		ID:           dbTag.ID,
		Name:         dbTag.Name,
		Created:      created,
		Modified:     modified,
		LastActivity: dbTag.LastActivity,
	}, nil
}

func (r *ItemRepository) UpsertItemTag(ctx context.Context, tagName string, lastActivity int64) (*entity.ItemTag, error) {
	queries := db.New(r.pool)
	dbTag, err := queries.UpsertItemTagRecord(ctx, db.UpsertItemTagRecordParams{
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

	return &entity.ItemTag{
		ID:           dbTag.ID,
		Name:         dbTag.Name,
		Created:      created,
		Modified:     modified,
		LastActivity: dbTag.LastActivity,
	}, nil
}

func (r *ItemRepository) CreateItemTagRelation(ctx context.Context, itemID, tagID uuid.UUID) error {
	queries := db.New(r.pool)
	return queries.CreateItemTagLinkage(ctx, db.CreateItemTagLinkageParams{
		ItemID: itemID,
		TagID:  tagID,
	})
}

func (r *ItemRepository) DeleteItemTags(ctx context.Context, itemID uuid.UUID) error {
	queries := db.New(r.pool)
	return queries.DeleteItemTagsRelation(ctx, itemID)
}

func (r *ItemRepository) GetItemTags(ctx context.Context, itemID uuid.UUID) ([]string, error) {
	queries := db.New(r.pool)
	tagsInterface, err := queries.GetItemTagsList(ctx, itemID)
	if err != nil {
		return nil, err
	}
	return convertInterfaceToStringSlice(tagsInterface), nil
}

func (r *ItemRepository) DeleteItemSemanticIndex(ctx context.Context, itemID uuid.UUID) error {
	queries := db.New(r.pool)
	pgItemID := pgtype.UUID{
		Bytes: itemID,
		Valid: true,
	}
	return queries.DeleteItemSemanticIndexRecord(ctx, pgItemID)
}

func (r *ItemRepository) SearchSimilarItems(ctx context.Context, embedding []float32, limit int32) ([]entity.ItemListItem, error) {
	queries := db.New(r.pool)

	// Convert []float32 to pgvector.Vector
	vec := pgvector.NewVector(embedding)

	dbItems, err := queries.SearchSimilarItemsByEmbedding(ctx, db.SearchSimilarItemsByEmbeddingParams{
		Column1: &vec,
		Limit:   limit,
	})
	if err != nil {
		return nil, err
	}

	items := make([]entity.ItemListItem, 0, len(dbItems))
	for _, dbItem := range dbItems {
		created := int64(0)
		if dbItem.Created != nil {
			created = *dbItem.Created
		}
		modified := int64(0)
		if dbItem.Modified != nil {
			modified = *dbItem.Modified
		}

		var tags []string
		if dbItem.Tags != nil {
			if tagSlice, ok := dbItem.Tags.([]interface{}); ok {
				tags = make([]string, len(tagSlice))
				for i, t := range tagSlice {
					if s, ok := t.(string); ok {
						tags[i] = s
					}
				}
			}
		}
		if tags == nil {
			tags = []string{}
		}

		items = append(items, entity.ItemListItem{
			ID:       dbItem.ID,
			Title:    dbItem.Title,
			Tags:     tags,
			Created:  created,
			Modified: modified,
		})
	}

	return items, nil
}

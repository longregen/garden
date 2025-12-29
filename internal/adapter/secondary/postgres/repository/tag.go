package repository

import (
	"context"

	db "garden3/internal/adapter/secondary/postgres/generated/db"
	"garden3/internal/domain/entity"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TagRepository implements the output.TagRepository interface
type TagRepository struct {
	pool *pgxpool.Pool
}

// NewTagRepository creates a new tag repository
func NewTagRepository(pool *pgxpool.Pool) *TagRepository {
	return &TagRepository{
		pool: pool,
	}
}

func (r *TagRepository) GetTagByName(ctx context.Context, tagName string) (*entity.Tag, error) {
	queries := db.New(r.pool)
	dbTag, err := queries.GetTagByName(ctx, tagName)
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

	return &entity.Tag{
		ID:           dbTag.ID,
		Name:         dbTag.Name,
		Created:      created,
		Modified:     modified,
		LastActivity: dbTag.LastActivity,
	}, nil
}

func (r *TagRepository) UpsertTag(ctx context.Context, tagName string, timestamp int64) (*entity.Tag, error) {
	queries := db.New(r.pool)
	dbTag, err := queries.UpsertTagRecord(ctx, db.UpsertTagRecordParams{
		Name:    tagName,
		Created: &timestamp,
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

	return &entity.Tag{
		ID:           dbTag.ID,
		Name:         dbTag.Name,
		Created:      created,
		Modified:     modified,
		LastActivity: dbTag.LastActivity,
	}, nil
}

func (r *TagRepository) CreateItemTagRelation(ctx context.Context, itemID, tagID uuid.UUID) error {
	queries := db.New(r.pool)
	return queries.CreateItemTagRelationRecord(ctx, db.CreateItemTagRelationRecordParams{
		ItemID: itemID,
		TagID:  tagID,
	})
}

func (r *TagRepository) DeleteItemTagRelation(ctx context.Context, itemID, tagID uuid.UUID) error {
	queries := db.New(r.pool)
	return queries.DeleteItemTagRelationRecord(ctx, db.DeleteItemTagRelationRecordParams{
		ItemID: itemID,
		TagID:  tagID,
	})
}

func (r *TagRepository) ListAllTags(ctx context.Context) ([]entity.Tag, error) {
	queries := db.New(r.pool)
	dbTags, err := queries.ListAllTagsRecords(ctx)
	if err != nil {
		return nil, err
	}

	tags := make([]entity.Tag, 0, len(dbTags))
	for _, dbTag := range dbTags {
		created := int64(0)
		if dbTag.Created != nil {
			created = *dbTag.Created
		}
		modified := int64(0)
		if dbTag.Modified != nil {
			modified = *dbTag.Modified
		}

		tags = append(tags, entity.Tag{
			ID:           dbTag.ID,
			Name:         dbTag.Name,
			Created:      created,
			Modified:     modified,
			LastActivity: dbTag.LastActivity,
		})
	}

	return tags, nil
}

func (r *TagRepository) ListAllTagsWithUsage(ctx context.Context) ([]entity.TagWithUsage, error) {
	queries := db.New(r.pool)
	dbTags, err := queries.ListAllTagsWithUsageRecords(ctx)
	if err != nil {
		return nil, err
	}

	tags := make([]entity.TagWithUsage, 0, len(dbTags))
	for _, dbTag := range dbTags {
		created := int64(0)
		if dbTag.Created != nil {
			created = *dbTag.Created
		}
		modified := int64(0)
		if dbTag.Modified != nil {
			modified = *dbTag.Modified
		}

		tags = append(tags, entity.TagWithUsage{
			Tag: entity.Tag{
				ID:           dbTag.ID,
				Name:         dbTag.Name,
				Created:      created,
				Modified:     modified,
				LastActivity: dbTag.LastActivity,
			},
			UsageCount: dbTag.UsageCount,
		})
	}

	return tags, nil
}

func (r *TagRepository) GetItemsByTag(ctx context.Context, tagID uuid.UUID, limit, offset int32) ([]entity.ItemListItem, error) {
	queries := db.New(r.pool)
	dbItems, err := queries.GetItemsByTagID(ctx, db.GetItemsByTagIDParams{
		TagID:  tagID,
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

		tags := []string{}
		if dbItem.Tags != nil {
			if tagSlice, ok := dbItem.Tags.([]interface{}); ok {
				for _, t := range tagSlice {
					if tagStr, ok := t.(string); ok {
						tags = append(tags, tagStr)
					}
				}
			}
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

func (r *TagRepository) CountItemsByTag(ctx context.Context, tagID uuid.UUID) (int64, error) {
	queries := db.New(r.pool)
	return queries.CountItemsByTagID(ctx, tagID)
}

func (r *TagRepository) DeleteTag(ctx context.Context, tagID uuid.UUID) error {
	queries := db.New(r.pool)
	return queries.DeleteTagRecord(ctx, tagID)
}

func (r *TagRepository) DeleteAllTagRelations(ctx context.Context, tagID uuid.UUID) error {
	queries := db.New(r.pool)
	return queries.DeleteAllTagRelationsRecord(ctx, tagID)
}

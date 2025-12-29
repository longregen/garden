package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"garden3/internal/adapter/secondary/postgres/generated/db"
	"garden3/internal/domain/entity"
	"garden3/internal/port/output"
)

type configurationRepository struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

func NewConfigurationRepository(pool *pgxpool.Pool) output.ConfigurationRepository {
	return &configurationRepository{
		pool:    pool,
		queries: db.New(pool),
	}
}

func (r *configurationRepository) ListConfigurations(ctx context.Context, filter entity.ConfigurationFilter) ([]entity.Configuration, error) {
	params := db.ListConfigurationsParams{
		KeyPrefix:      filter.KeyPrefix,
		IncludeSecrets: &filter.IncludeSecrets,
	}

	rows, err := r.queries.ListConfigurations(ctx, params)
	if err != nil {
		return nil, err
	}

	configs := make([]entity.Configuration, len(rows))
	for i, row := range rows {
		configs[i] = toEntityConfiguration(row)
	}

	return configs, nil
}

func (r *configurationRepository) GetByKey(ctx context.Context, key string) (*entity.Configuration, error) {
	row, err := r.queries.GetConfigurationByKey(ctx, key)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	config := toEntityConfiguration(row)
	return &config, nil
}

func (r *configurationRepository) GetByPrefix(ctx context.Context, prefix string, includeSecrets bool) ([]entity.Configuration, error) {
	params := db.GetConfigurationsByPrefixParams{
		Column1:        &prefix,
		IncludeSecrets: &includeSecrets,
	}

	rows, err := r.queries.GetConfigurationsByPrefix(ctx, params)
	if err != nil {
		return nil, err
	}

	configs := make([]entity.Configuration, len(rows))
	for i, row := range rows {
		configs[i] = toEntityConfiguration(row)
	}

	return configs, nil
}

func (r *configurationRepository) Create(ctx context.Context, config entity.NewConfiguration) (*entity.Configuration, error) {
	params := db.CreateConfigurationParams{
		Key:       config.Key,
		Value:     config.Value,
		IsSecret:  config.IsSecret,
		UpdatedAt: convertTimeToPgTimestamp(config.UpdatedAt),
	}

	row, err := r.queries.CreateConfiguration(ctx, params)
	if err != nil {
		return nil, err
	}

	result := toEntityConfiguration(row)
	return &result, nil
}

func (r *configurationRepository) Update(ctx context.Context, key string, update entity.ConfigurationUpdate) (*entity.Configuration, error) {
	params := db.UpdateConfigurationParams{
		Key:       key,
		Value:     update.Value,
		IsSecret:  update.IsSecret,
		UpdatedAt: convertTimeToPgTimestamp(update.UpdatedAt),
	}

	row, err := r.queries.UpdateConfiguration(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	result := toEntityConfiguration(row)
	return &result, nil
}

func (r *configurationRepository) Delete(ctx context.Context, key string) error {
	return r.queries.DeleteConfiguration(ctx, key)
}

func (r *configurationRepository) Upsert(ctx context.Context, key, value string, isSecret bool, updatedAt interface{}) (*entity.Configuration, error) {
	var updateTime time.Time
	switch v := updatedAt.(type) {
	case time.Time:
		updateTime = v
	default:
		updateTime = time.Now()
	}

	params := db.UpsertConfigurationParams{
		Key:       key,
		Value:     value,
		IsSecret:  isSecret,
		UpdatedAt: convertTimeToPgTimestamp(updateTime),
	}

	row, err := r.queries.UpsertConfiguration(ctx, params)
	if err != nil {
		return nil, err
	}

	result := toEntityConfiguration(row)
	return &result, nil
}

func toEntityConfiguration(row db.Configuration) entity.Configuration {
	return entity.Configuration{
		ConfigID:  row.ConfigID,
		Key:       row.Key,
		Value:     row.Value,
		IsSecret:  row.IsSecret,
		CreatedAt: convertPgTimestampToTime(row.CreatedAt),
		UpdatedAt: convertPgTimestampToTime(row.UpdatedAt),
	}
}

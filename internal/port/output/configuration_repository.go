package output

import (
	"context"

	"garden3/internal/domain/entity"
)

type ConfigurationRepository interface {
	// ListConfigurations returns all configurations with optional filtering
	ListConfigurations(ctx context.Context, filter entity.ConfigurationFilter) ([]entity.Configuration, error)

	// GetByKey returns a single configuration by key
	GetByKey(ctx context.Context, key string) (*entity.Configuration, error)

	// GetByPrefix returns configurations with a given key prefix
	GetByPrefix(ctx context.Context, prefix string, includeSecrets bool) ([]entity.Configuration, error)

	// Create creates a new configuration
	Create(ctx context.Context, config entity.NewConfiguration) (*entity.Configuration, error)

	// Update updates an existing configuration
	Update(ctx context.Context, key string, update entity.ConfigurationUpdate) (*entity.Configuration, error)

	// Delete deletes a configuration
	Delete(ctx context.Context, key string) error

	// Upsert creates or updates a configuration
	Upsert(ctx context.Context, key, value string, isSecret bool, updatedAt interface{}) (*entity.Configuration, error)
}

package input

import (
	"context"

	"garden3/internal/domain/entity"
)

type ConfigurationUseCase interface {
	// ListConfigurations returns all configurations with optional filtering
	ListConfigurations(ctx context.Context, filter entity.ConfigurationFilter) ([]entity.Configuration, error)

	// GetConfigurationByKey returns a single configuration by key
	GetConfigurationByKey(ctx context.Context, key string) (*entity.Configuration, error)

	// GetConfigurationsByPrefix returns configurations with a given key prefix
	GetConfigurationsByPrefix(ctx context.Context, prefix string, includeSecrets bool) ([]entity.Configuration, error)

	// CreateConfiguration creates a new configuration
	CreateConfiguration(ctx context.Context, config entity.NewConfiguration) (*entity.Configuration, error)

	// UpdateConfiguration updates an existing configuration
	UpdateConfiguration(ctx context.Context, key string, update entity.ConfigurationUpdate) (*entity.Configuration, error)

	// DeleteConfiguration deletes a configuration
	DeleteConfiguration(ctx context.Context, key string) error

	// SetConfiguration creates or updates a configuration
	SetConfiguration(ctx context.Context, key, value string, isSecret bool) (*entity.Configuration, error)

	// GetValue returns just the value of a configuration
	GetValue(ctx context.Context, key string) (*string, error)

	// GetBoolValue returns a configuration value as a boolean
	GetBoolValue(ctx context.Context, key string, defaultValue bool) (bool, error)

	// GetNumberValue returns a configuration value as a float64
	GetNumberValue(ctx context.Context, key string, defaultValue float64) (float64, error)

	// GetJSONValue returns a configuration value as raw JSON
	GetJSONValue(ctx context.Context, key string) ([]byte, error)
}

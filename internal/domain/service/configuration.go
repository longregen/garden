package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"garden3/internal/domain/entity"
	"garden3/internal/port/input"
	"garden3/internal/port/output"
)

type configurationService struct {
	repo output.ConfigurationRepository
}

func NewConfigurationService(repo output.ConfigurationRepository) input.ConfigurationUseCase {
	return &configurationService{
		repo: repo,
	}
}

func (s *configurationService) ListConfigurations(ctx context.Context, filter entity.ConfigurationFilter) ([]entity.Configuration, error) {
	return s.repo.ListConfigurations(ctx, filter)
}

func (s *configurationService) GetConfigurationByKey(ctx context.Context, key string) (*entity.Configuration, error) {
	if key == "" {
		return nil, fmt.Errorf("configuration key is required")
	}
	return s.repo.GetByKey(ctx, key)
}

func (s *configurationService) GetConfigurationsByPrefix(ctx context.Context, prefix string, includeSecrets bool) ([]entity.Configuration, error) {
	if prefix == "" {
		return nil, fmt.Errorf("prefix is required")
	}
	return s.repo.GetByPrefix(ctx, prefix, includeSecrets)
}

func (s *configurationService) CreateConfiguration(ctx context.Context, config entity.NewConfiguration) (*entity.Configuration, error) {
	if config.Key == "" {
		return nil, fmt.Errorf("configuration key is required")
	}
	if config.Value == "" {
		return nil, fmt.Errorf("configuration value is required")
	}

	config.UpdatedAt = time.Now()
	return s.repo.Create(ctx, config)
}

func (s *configurationService) UpdateConfiguration(ctx context.Context, key string, update entity.ConfigurationUpdate) (*entity.Configuration, error) {
	if key == "" {
		return nil, fmt.Errorf("configuration key is required")
	}

	update.UpdatedAt = time.Now()
	return s.repo.Update(ctx, key, update)
}

func (s *configurationService) DeleteConfiguration(ctx context.Context, key string) error {
	if key == "" {
		return fmt.Errorf("configuration key is required")
	}
	return s.repo.Delete(ctx, key)
}

func (s *configurationService) SetConfiguration(ctx context.Context, key, value string, isSecret bool) (*entity.Configuration, error) {
	if key == "" {
		return nil, fmt.Errorf("configuration key is required")
	}
	if value == "" {
		return nil, fmt.Errorf("configuration value is required")
	}

	return s.repo.Upsert(ctx, key, value, isSecret, time.Now())
}

func (s *configurationService) GetValue(ctx context.Context, key string) (*string, error) {
	if key == "" {
		return nil, fmt.Errorf("configuration key is required")
	}

	config, err := s.repo.GetByKey(ctx, key)
	if err != nil {
		return nil, err
	}
	if config == nil {
		return nil, nil
	}

	return &config.Value, nil
}

func (s *configurationService) GetBoolValue(ctx context.Context, key string, defaultValue bool) (bool, error) {
	if key == "" {
		return defaultValue, fmt.Errorf("configuration key is required")
	}

	value, err := s.GetValue(ctx, key)
	if err != nil {
		return defaultValue, err
	}
	if value == nil {
		return defaultValue, nil
	}

	return strings.ToLower(*value) == "true", nil
}

func (s *configurationService) GetNumberValue(ctx context.Context, key string, defaultValue float64) (float64, error) {
	if key == "" {
		return defaultValue, fmt.Errorf("configuration key is required")
	}

	value, err := s.GetValue(ctx, key)
	if err != nil {
		return defaultValue, err
	}
	if value == nil {
		return defaultValue, nil
	}

	num, err := strconv.ParseFloat(*value, 64)
	if err != nil {
		return defaultValue, nil
	}

	return num, nil
}

func (s *configurationService) GetJSONValue(ctx context.Context, key string) ([]byte, error) {
	if key == "" {
		return nil, fmt.Errorf("configuration key is required")
	}

	value, err := s.GetValue(ctx, key)
	if err != nil {
		return nil, err
	}
	if value == nil {
		return nil, nil
	}

	// Validate it's valid JSON
	var js interface{}
	if err := json.Unmarshal([]byte(*value), &js); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return []byte(*value), nil
}

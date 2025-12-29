package entity

import (
	"time"

	"github.com/google/uuid"
)

type Configuration struct {
	ConfigID  uuid.UUID `json:"config_id"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	IsSecret  bool      `json:"is_secret"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type NewConfiguration struct {
	Key       string
	Value     string
	IsSecret  bool
	UpdatedAt time.Time
}

type ConfigurationUpdate struct {
	Value     string
	IsSecret  bool
	UpdatedAt time.Time
}

type ConfigurationFilter struct {
	KeyPrefix      *string
	IncludeSecrets bool
}

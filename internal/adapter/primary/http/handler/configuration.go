package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	httpAdapter "garden3/internal/adapter/primary/http"
	"garden3/internal/domain/entity"
	"garden3/internal/port/input"
)

type ConfigurationHandler struct {
	useCase input.ConfigurationUseCase
}

func NewConfigurationHandler(useCase input.ConfigurationUseCase) *ConfigurationHandler {
	return &ConfigurationHandler{
		useCase: useCase,
	}
}

func (h *ConfigurationHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/configurations", func(r chi.Router) {
		r.Get("/", h.ListConfigurations)
		r.Post("/", h.CreateConfiguration)
		r.Get("/{key}", h.GetConfiguration)
		r.Put("/{key}", h.UpdateConfiguration)
		r.Delete("/{key}", h.DeleteConfiguration)
		r.Put("/{key}/value", h.SetConfiguration)
		r.Get("/{key}/value", h.GetValue)
		r.Get("/{key}/bool", h.GetBoolValue)
		r.Get("/{key}/number", h.GetNumberValue)
		r.Get("/{key}/json", h.GetJsonValue)
		r.Get("/prefix/{prefix}", h.GetByPrefix)
	})

	// Backwards-compatible aliases for legacy endpoints
	r.Route("/api/configuration", func(r chi.Router) {
		r.Get("/get", h.LegacyGetConfiguration)
		r.Post("/get", h.LegacyGetConfigurationPost)
		r.Get("/set", h.LegacySetConfiguration)
		r.Post("/set", h.LegacySetConfigurationPost)
	})
}

// ListConfigurations godoc
// @Summary List all configurations
// @Description Get all configurations with optional filtering by prefix and secrets
// @Tags configurations
// @Param prefix query string false "Key prefix filter"
// @Param include_secrets query bool false "Include secret configurations" default(true)
// @Success 200 {array} entity.Configuration
// @Router /api/configurations [get]
func (h *ConfigurationHandler) ListConfigurations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	prefix := r.URL.Query().Get("prefix")
	includeSecretsStr := r.URL.Query().Get("include_secrets")
	includeSecrets := true
	if includeSecretsStr != "" {
		var err error
		includeSecrets, err = strconv.ParseBool(includeSecretsStr)
		if err != nil {
			httpAdapter.BadRequest(w, err)
			return
		}
	}

	filter := entity.ConfigurationFilter{
		IncludeSecrets: includeSecrets,
	}
	if prefix != "" {
		filter.KeyPrefix = &prefix
	}

	configs, err := h.useCase.ListConfigurations(ctx, filter)
	if err != nil {
		httpAdapter.InternalError(w, err)
		return
	}

	httpAdapter.JSON(w, http.StatusOK, configs)
}

// GetConfiguration godoc
// @Summary Get configuration by key
// @Description Get a single configuration by its key
// @Tags configurations
// @Param key path string true "Configuration key"
// @Success 200 {object} entity.Configuration
// @Router /api/configurations/{key} [get]
func (h *ConfigurationHandler) GetConfiguration(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	key := chi.URLParam(r, "key")

	config, err := h.useCase.GetConfigurationByKey(ctx, key)
	if err != nil {
		httpAdapter.InternalError(w, err)
		return
	}
	if config == nil {
		httpAdapter.NotFound(w)
		return
	}

	httpAdapter.JSON(w, http.StatusOK, config)
}

// GetByPrefix godoc
// @Summary Get configurations by prefix
// @Description Get all configurations with a given key prefix
// @Tags configurations
// @Param prefix path string true "Key prefix"
// @Param include_secrets query bool false "Include secret configurations" default(true)
// @Success 200 {array} entity.Configuration
// @Router /api/configurations/prefix/{prefix} [get]
func (h *ConfigurationHandler) GetByPrefix(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	prefix := chi.URLParam(r, "prefix")

	includeSecretsStr := r.URL.Query().Get("include_secrets")
	includeSecrets := true
	if includeSecretsStr != "" {
		var err error
		includeSecrets, err = strconv.ParseBool(includeSecretsStr)
		if err != nil {
			httpAdapter.BadRequest(w, err)
			return
		}
	}

	configs, err := h.useCase.GetConfigurationsByPrefix(ctx, prefix, includeSecrets)
	if err != nil {
		httpAdapter.InternalError(w, err)
		return
	}

	httpAdapter.JSON(w, http.StatusOK, configs)
}

type CreateConfigurationRequest struct {
	Key      string `json:"key"`
	Value    string `json:"value"`
	IsSecret bool   `json:"is_secret"`
}

// CreateConfiguration godoc
// @Summary Create a new configuration
// @Description Create a new configuration entry
// @Tags configurations
// @Accept json
// @Param body body CreateConfigurationRequest true "Configuration data"
// @Success 201 {object} entity.Configuration
// @Router /api/configurations [post]
func (h *ConfigurationHandler) CreateConfiguration(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req CreateConfigurationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpAdapter.BadRequest(w, err)
		return
	}

	config := entity.NewConfiguration{
		Key:      req.Key,
		Value:    req.Value,
		IsSecret: req.IsSecret,
	}

	created, err := h.useCase.CreateConfiguration(ctx, config)
	if err != nil {
		httpAdapter.InternalError(w, err)
		return
	}

	httpAdapter.JSON(w, http.StatusCreated, created)
}

type UpdateConfigurationRequest struct {
	Value    string `json:"value"`
	IsSecret bool   `json:"is_secret"`
}

// UpdateConfiguration godoc
// @Summary Update a configuration
// @Description Update an existing configuration
// @Tags configurations
// @Accept json
// @Param key path string true "Configuration key"
// @Param body body UpdateConfigurationRequest true "Update data"
// @Success 200 {object} entity.Configuration
// @Router /api/configurations/{key} [put]
func (h *ConfigurationHandler) UpdateConfiguration(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	key := chi.URLParam(r, "key")

	var req UpdateConfigurationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpAdapter.BadRequest(w, err)
		return
	}

	update := entity.ConfigurationUpdate{
		Value:    req.Value,
		IsSecret: req.IsSecret,
	}

	updated, err := h.useCase.UpdateConfiguration(ctx, key, update)
	if err != nil {
		httpAdapter.InternalError(w, err)
		return
	}
	if updated == nil {
		httpAdapter.NotFound(w)
		return
	}

	httpAdapter.JSON(w, http.StatusOK, updated)
}

// DeleteConfiguration godoc
// @Summary Delete a configuration
// @Description Delete a configuration by key
// @Tags configurations
// @Param key path string true "Configuration key"
// @Success 204
// @Router /api/configurations/{key} [delete]
func (h *ConfigurationHandler) DeleteConfiguration(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	key := chi.URLParam(r, "key")

	if err := h.useCase.DeleteConfiguration(ctx, key); err != nil {
		httpAdapter.InternalError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type SetConfigurationRequest struct {
	Value    string `json:"value"`
	IsSecret bool   `json:"is_secret"`
}

// SetConfiguration godoc
// @Summary Set configuration value (upsert)
// @Description Create or update a configuration value
// @Tags configurations
// @Accept json
// @Param key path string true "Configuration key"
// @Param body body SetConfigurationRequest true "Value and secret flag"
// @Success 200 {object} entity.Configuration
// @Router /api/configurations/{key}/value [put]
func (h *ConfigurationHandler) SetConfiguration(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	key := chi.URLParam(r, "key")

	var req SetConfigurationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpAdapter.BadRequest(w, err)
		return
	}

	config, err := h.useCase.SetConfiguration(ctx, key, req.Value, req.IsSecret)
	if err != nil {
		httpAdapter.InternalError(w, err)
		return
	}

	httpAdapter.JSON(w, http.StatusOK, config)
}

// GetValue godoc
// @Summary Get configuration value
// @Description Get just the value of a configuration
// @Tags configurations
// @Param key path string true "Configuration key"
// @Success 200 {object} map[string]string
// @Router /api/configurations/{key}/value [get]
func (h *ConfigurationHandler) GetValue(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	key := chi.URLParam(r, "key")

	value, err := h.useCase.GetValue(ctx, key)
	if err != nil {
		httpAdapter.InternalError(w, err)
		return
	}
	if value == nil {
		httpAdapter.NotFound(w)
		return
	}

	httpAdapter.JSON(w, http.StatusOK, map[string]string{"value": *value})
}

// GetBoolValue godoc
// @Summary Get configuration value as boolean
// @Description Get a configuration value parsed as a boolean
// @Tags configurations
// @Param key path string true "Configuration key"
// @Param default query bool false "Default value if not found" default(false)
// @Success 200 {object} map[string]bool
// @Router /api/configurations/{key}/bool [get]
func (h *ConfigurationHandler) GetBoolValue(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	key := chi.URLParam(r, "key")

	defaultValueStr := r.URL.Query().Get("default")
	defaultValue := false
	if defaultValueStr != "" {
		var err error
		defaultValue, err = strconv.ParseBool(defaultValueStr)
		if err != nil {
			httpAdapter.BadRequest(w, err)
			return
		}
	}

	value, err := h.useCase.GetBoolValue(ctx, key, defaultValue)
	if err != nil {
		httpAdapter.InternalError(w, err)
		return
	}

	httpAdapter.JSON(w, http.StatusOK, map[string]bool{"value": value})
}

// GetNumberValue godoc
// @Summary Get configuration value as number
// @Description Get a configuration value parsed as a number
// @Tags configurations
// @Param key path string true "Configuration key"
// @Param default query number false "Default value if not found" default(0)
// @Success 200 {object} map[string]float64
// @Router /api/configurations/{key}/number [get]
func (h *ConfigurationHandler) GetNumberValue(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	key := chi.URLParam(r, "key")

	defaultValueStr := r.URL.Query().Get("default")
	defaultValue := 0.0
	if defaultValueStr != "" {
		var err error
		defaultValue, err = strconv.ParseFloat(defaultValueStr, 64)
		if err != nil {
			httpAdapter.BadRequest(w, err)
			return
		}
	}

	value, err := h.useCase.GetNumberValue(ctx, key, defaultValue)
	if err != nil {
		httpAdapter.InternalError(w, err)
		return
	}

	httpAdapter.JSON(w, http.StatusOK, map[string]float64{"value": value})
}

// GetJsonValue godoc
// @Summary Get configuration value as JSON
// @Description Get a configuration value parsed as JSON
// @Tags configurations
// @Param key path string true "Configuration key"
// @Success 200 {object} map[string]interface{}
// @Router /api/configurations/{key}/json [get]
func (h *ConfigurationHandler) GetJsonValue(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	key := chi.URLParam(r, "key")

	value, err := h.useCase.GetJSONValue(ctx, key)
	if err != nil {
		httpAdapter.InternalError(w, err)
		return
	}
	if value == nil {
		httpAdapter.NotFound(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(value)
}

// Legacy endpoint compatibility wrappers

type LegacyGetRequest struct {
	Key string `json:"key"`
}

type LegacySetRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// LegacyGetConfiguration handles GET /api/configuration/get?key={key}
func (h *ConfigurationHandler) LegacyGetConfiguration(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	key := r.URL.Query().Get("key")

	if key == "" {
		httpAdapter.BadRequest(w, fmt.Errorf("key parameter is required"))
		return
	}

	config, err := h.useCase.GetConfigurationByKey(ctx, key)
	if err != nil {
		httpAdapter.InternalError(w, err)
		return
	}
	if config == nil {
		httpAdapter.NotFound(w)
		return
	}

	httpAdapter.JSON(w, http.StatusOK, config)
}

// LegacyGetConfigurationPost handles POST /api/configuration/get with {key} in body
func (h *ConfigurationHandler) LegacyGetConfigurationPost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req LegacyGetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpAdapter.BadRequest(w, err)
		return
	}

	if req.Key == "" {
		httpAdapter.BadRequest(w, fmt.Errorf("key field is required"))
		return
	}

	config, err := h.useCase.GetConfigurationByKey(ctx, req.Key)
	if err != nil {
		httpAdapter.InternalError(w, err)
		return
	}
	if config == nil {
		httpAdapter.NotFound(w)
		return
	}

	httpAdapter.JSON(w, http.StatusOK, config)
}

// LegacySetConfiguration handles GET /api/configuration/set?key={key}&value={value}
func (h *ConfigurationHandler) LegacySetConfiguration(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	key := r.URL.Query().Get("key")
	value := r.URL.Query().Get("value")

	if key == "" {
		httpAdapter.BadRequest(w, fmt.Errorf("key parameter is required"))
		return
	}
	if value == "" {
		httpAdapter.BadRequest(w, fmt.Errorf("value parameter is required"))
		return
	}

	config, err := h.useCase.SetConfiguration(ctx, key, value, false)
	if err != nil {
		httpAdapter.InternalError(w, err)
		return
	}

	httpAdapter.JSON(w, http.StatusOK, config)
}

// LegacySetConfigurationPost handles POST /api/configuration/set with {key, value} in body
func (h *ConfigurationHandler) LegacySetConfigurationPost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req LegacySetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpAdapter.BadRequest(w, err)
		return
	}

	if req.Key == "" {
		httpAdapter.BadRequest(w, fmt.Errorf("key field is required"))
		return
	}
	if req.Value == "" {
		httpAdapter.BadRequest(w, fmt.Errorf("value field is required"))
		return
	}

	config, err := h.useCase.SetConfiguration(ctx, req.Key, req.Value, false)
	if err != nil {
		httpAdapter.InternalError(w, err)
		return
	}

	httpAdapter.JSON(w, http.StatusOK, config)
}

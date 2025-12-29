package handler

import (
	"encoding/json"
	"net/http"

	httpAdapter "garden3/internal/adapter/primary/http"
	"garden3/internal/domain/entity"
	"garden3/internal/port/input"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type EntityHandler struct {
	useCase input.EntityUseCase
}

func NewEntityHandler(useCase input.EntityUseCase) *EntityHandler {
	return &EntityHandler{
		useCase: useCase,
	}
}

func (h *EntityHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/entities", func(r chi.Router) {
		r.Get("/", h.ListEntities)
		r.Post("/", h.CreateEntity)
		r.Get("/search", h.SearchEntities)
		r.Get("/deleted", h.ListDeletedEntities)
		r.Post("/relationship", h.CreateEntityRelationship)
		r.Route("/relationship/{id}", func(r chi.Router) {
			r.Delete("/", h.DeleteEntityRelationship)
		})
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.GetEntity)
			r.Put("/", h.UpdateEntity)
			r.Delete("/", h.DeleteEntity)
			r.Get("/relationships", h.GetEntityRelationships)
		})
	})

	r.Route("/api/entity-references", func(r chi.Router) {
		r.Get("/", h.GetEntityReferences)
		r.Post("/parse", h.ParseEntityReferences)
	})
}

// EntityResponse represents an entity in the API response
type EntityResponse struct {
	EntityID    uuid.UUID       `json:"entity_id"`
	Name        string          `json:"name"`
	Type        string          `json:"type"`
	Description *string         `json:"description"`
	Properties  json.RawMessage `json:"properties"`
	CreatedAt   string          `json:"created_at"`
	UpdatedAt   string          `json:"updated_at"`
}

// GetEntity godoc
// @Summary Get entity by ID
// @Description Get a single entity by ID
// @Tags entities
// @Param id path string true "Entity ID"
// @Success 200 {object} EntityResponse
// @Router /api/entities/{id} [get]
func (h *EntityHandler) GetEntity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	entityIDStr := chi.URLParam(r, "id")

	entityID, err := uuid.Parse(entityIDStr)
	if err != nil {
		http.Error(w, "Invalid entity ID", http.StatusBadRequest)
		return
	}

	ent, err := h.useCase.GetEntity(ctx, entityID)
	if err != nil {
		http.Error(w, "Entity not found", http.StatusNotFound)
		return
	}

	response := EntityResponse{
		EntityID:    ent.EntityID,
		Name:        ent.Name,
		Type:        ent.Type,
		Description: ent.Description,
		Properties:  ent.Properties,
		CreatedAt:   ent.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   ent.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	httpAdapter.JSON(w, http.StatusOK, response)
}

// UpdateEntityRequest represents the request body for updating an entity
type UpdateEntityRequest struct {
	Name        *string          `json:"name"`
	Type        *string          `json:"type"`
	Description *string          `json:"description"`
	Properties  *json.RawMessage `json:"properties"`
}

// UpdateEntity godoc
// @Summary Update entity
// @Description Update an entity's fields
// @Tags entities
// @Param id path string true "Entity ID"
// @Param body body UpdateEntityRequest true "Entity update data"
// @Success 200 {object} EntityResponse
// @Router /api/entities/{id} [put]
func (h *EntityHandler) UpdateEntity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	entityIDStr := chi.URLParam(r, "id")

	entityID, err := uuid.Parse(entityIDStr)
	if err != nil {
		http.Error(w, "Invalid entity ID", http.StatusBadRequest)
		return
	}

	var req UpdateEntityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	input := entity.UpdateEntityInput{
		Name:        req.Name,
		Type:        req.Type,
		Description: req.Description,
		Properties:  req.Properties,
	}

	updatedEntity, err := h.useCase.UpdateEntity(ctx, entityID, input)
	if err != nil {
		http.Error(w, "Failed to update entity", http.StatusInternalServerError)
		return
	}

	response := EntityResponse{
		EntityID:    updatedEntity.EntityID,
		Name:        updatedEntity.Name,
		Type:        updatedEntity.Type,
		Description: updatedEntity.Description,
		Properties:  updatedEntity.Properties,
		CreatedAt:   updatedEntity.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   updatedEntity.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	httpAdapter.JSON(w, http.StatusOK, response)
}

// DeleteEntity godoc
// @Summary Delete entity
// @Description Soft delete an entity
// @Tags entities
// @Param id path string true "Entity ID"
// @Success 204
// @Router /api/entities/{id} [delete]
func (h *EntityHandler) DeleteEntity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	entityIDStr := chi.URLParam(r, "id")

	entityID, err := uuid.Parse(entityIDStr)
	if err != nil {
		http.Error(w, "Invalid entity ID", http.StatusBadRequest)
		return
	}

	if err := h.useCase.DeleteEntity(ctx, entityID); err != nil {
		http.Error(w, "Failed to delete entity", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// EntityReferenceResponse represents an entity reference in the API response
type EntityReferenceResponse struct {
	ID            uuid.UUID `json:"id"`
	SourceType    string    `json:"source_type"`
	SourceID      uuid.UUID `json:"source_id"`
	EntityID      uuid.UUID `json:"entity_id"`
	ReferenceText string    `json:"reference_text"`
	Position      *int      `json:"position"`
	CreatedAt     string    `json:"created_at"`
}

// GetEntityReferences godoc
// @Summary Get entity references
// @Description Get references to an entity with optional source type filter
// @Tags entities
// @Param entity_id query string true "Entity ID"
// @Param source_type query string false "Source type filter"
// @Success 200 {array} EntityReferenceResponse
// @Router /api/entity-references [get]
func (h *EntityHandler) GetEntityReferences(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	entityIDStr := r.URL.Query().Get("entity_id")
	sourceTypeStr := r.URL.Query().Get("source_type")

	if entityIDStr == "" {
		http.Error(w, "entity_id query parameter is required", http.StatusBadRequest)
		return
	}

	entityID, err := uuid.Parse(entityIDStr)
	if err != nil {
		http.Error(w, "Invalid entity_id", http.StatusBadRequest)
		return
	}

	var sourceType *string
	if sourceTypeStr != "" {
		sourceType = &sourceTypeStr
	}

	refs, err := h.useCase.GetEntityReferences(ctx, entityID, sourceType)
	if err != nil {
		http.Error(w, "Failed to get entity references", http.StatusInternalServerError)
		return
	}

	response := make([]EntityReferenceResponse, len(refs))
	for i, ref := range refs {
		response[i] = EntityReferenceResponse{
			ID:            ref.ID,
			SourceType:    ref.SourceType,
			SourceID:      ref.SourceID,
			EntityID:      ref.EntityID,
			ReferenceText: ref.ReferenceText,
			Position:      ref.Position,
			CreatedAt:     ref.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	httpAdapter.JSON(w, http.StatusOK, response)
}

// ParsedReferenceResponse represents a parsed entity reference in the API response
type ParsedReferenceResponse struct {
	Original    string `json:"original"`
	EntityName  string `json:"entity_name"`
	DisplayText string `json:"display_text"`
}

// ParseEntityReferencesRequest represents the request body for parsing entity references
type ParseEntityReferencesRequest struct {
	Content string `json:"content"`
}

// ParseEntityReferences godoc
// @Summary Parse entity references from content
// @Description Parse [[entity]] references from content text
// @Tags entities
// @Param body body ParseEntityReferencesRequest true "Content to parse"
// @Success 200 {array} ParsedReferenceResponse
// @Router /api/entity-references/parse [post]
func (h *EntityHandler) ParseEntityReferences(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req ParseEntityReferencesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	refs, err := h.useCase.ParseEntityReferences(ctx, req.Content)
	if err != nil {
		http.Error(w, "Failed to parse entity references", http.StatusInternalServerError)
		return
	}

	response := make([]ParsedReferenceResponse, len(refs))
	for i, ref := range refs {
		response[i] = ParsedReferenceResponse{
			Original:    ref.Original,
			EntityName:  ref.EntityName,
			DisplayText: ref.DisplayText,
		}
	}

	httpAdapter.JSON(w, http.StatusOK, response)
}

// ListEntities godoc
// @Summary List entities
// @Description List all entities with optional filters
// @Tags entities
// @Param type query string false "Entity type filter"
// @Param updated_since query string false "Updated since timestamp"
// @Success 200 {array} EntityResponse
// @Router /api/entities [get]
func (h *EntityHandler) ListEntities(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	entityType := r.URL.Query().Get("type")
	updatedSinceStr := r.URL.Query().Get("updated_since")

	var updatedSince *string
	if updatedSinceStr != "" {
		updatedSince = &updatedSinceStr
	}

	entities, err := h.useCase.ListEntities(ctx, entityType, updatedSince)
	if err != nil {
		http.Error(w, "Failed to list entities", http.StatusInternalServerError)
		return
	}

	response := make([]EntityResponse, len(entities))
	for i, ent := range entities {
		response[i] = EntityResponse{
			EntityID:    ent.EntityID,
			Name:        ent.Name,
			Type:        ent.Type,
			Description: ent.Description,
			Properties:  ent.Properties,
			CreatedAt:   ent.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:   ent.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	httpAdapter.JSON(w, http.StatusOK, response)
}

// CreateEntityRequest represents the request body for creating an entity
type CreateEntityRequest struct {
	Name        string           `json:"name"`
	Type        string           `json:"type"`
	Description *string          `json:"description"`
	Properties  *json.RawMessage `json:"properties"`
}

// CreateEntity godoc
// @Summary Create entity
// @Description Create a new entity
// @Tags entities
// @Param body body CreateEntityRequest true "Entity data"
// @Success 201 {object} EntityResponse
// @Router /api/entities [post]
func (h *EntityHandler) CreateEntity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req CreateEntityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Type == "" {
		http.Error(w, "Name and type are required", http.StatusBadRequest)
		return
	}

	input := entity.CreateEntityInput{
		Name:        req.Name,
		Type:        req.Type,
		Description: req.Description,
		Properties:  req.Properties,
	}

	created, err := h.useCase.CreateEntity(ctx, input)
	if err != nil {
		http.Error(w, "Failed to create entity", http.StatusInternalServerError)
		return
	}

	response := EntityResponse{
		EntityID:    created.EntityID,
		Name:        created.Name,
		Type:        created.Type,
		Description: created.Description,
		Properties:  created.Properties,
		CreatedAt:   created.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   created.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	httpAdapter.JSON(w, http.StatusCreated, response)
}

// SearchEntities godoc
// @Summary Search entities
// @Description Search entities by name with optional type filter
// @Tags entities
// @Param q query string false "Search query (alternative to 'query')"
// @Param query query string false "Search query (alternative to 'q')"
// @Param type query string false "Entity type filter"
// @Success 200 {array} EntityResponse
// @Router /api/entities/search [get]
func (h *EntityHandler) SearchEntities(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Support both 'q' and 'query' for backwards compatibility
	query := r.URL.Query().Get("q")
	if query == "" {
		query = r.URL.Query().Get("query")
	}
	entityType := r.URL.Query().Get("type")

	if query == "" {
		http.Error(w, "Query parameter 'q' or 'query' is required", http.StatusBadRequest)
		return
	}

	entities, err := h.useCase.SearchEntities(ctx, query, entityType)
	if err != nil {
		http.Error(w, "Failed to search entities", http.StatusInternalServerError)
		return
	}

	response := make([]EntityResponse, len(entities))
	for i, ent := range entities {
		response[i] = EntityResponse{
			EntityID:    ent.EntityID,
			Name:        ent.Name,
			Type:        ent.Type,
			Description: ent.Description,
			Properties:  ent.Properties,
			CreatedAt:   ent.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:   ent.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	httpAdapter.JSON(w, http.StatusOK, response)
}

// ListDeletedEntities godoc
// @Summary List deleted entities
// @Description List all soft-deleted entities
// @Tags entities
// @Success 200 {array} EntityResponse
// @Router /api/entities/deleted [get]
func (h *EntityHandler) ListDeletedEntities(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	entities, err := h.useCase.ListDeletedEntities(ctx)
	if err != nil {
		http.Error(w, "Failed to list deleted entities", http.StatusInternalServerError)
		return
	}

	response := make([]EntityResponse, len(entities))
	for i, ent := range entities {
		response[i] = EntityResponse{
			EntityID:    ent.EntityID,
			Name:        ent.Name,
			Type:        ent.Type,
			Description: ent.Description,
			Properties:  ent.Properties,
			CreatedAt:   ent.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:   ent.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	httpAdapter.JSON(w, http.StatusOK, response)
}

// EntityRelationshipResponse represents an entity relationship in the API response
type EntityRelationshipResponse struct {
	ID               uuid.UUID       `json:"id"`
	EntityID         uuid.UUID       `json:"entity_id"`
	RelatedType      string          `json:"related_type"`
	RelatedID        uuid.UUID       `json:"related_id"`
	RelationshipType string          `json:"relationship_type"`
	Metadata         json.RawMessage `json:"metadata"`
	CreatedAt        string          `json:"created_at"`
	UpdatedAt        string          `json:"updated_at"`
}

// CreateEntityRelationshipRequest represents the request body for creating an entity relationship
type CreateEntityRelationshipRequest struct {
	EntityID         uuid.UUID       `json:"entity_id"`
	RelatedType      string          `json:"related_type"`
	RelatedID        uuid.UUID       `json:"related_id"`
	RelationshipType string          `json:"relationship_type"`
	Metadata         json.RawMessage `json:"metadata"`
}

// CreateEntityRelationship godoc
// @Summary Create entity relationship
// @Description Create a new relationship between an entity and another resource
// @Tags entities
// @Param body body CreateEntityRelationshipRequest true "Relationship data"
// @Success 201 {object} EntityRelationshipResponse
// @Router /api/entities/relationship [post]
func (h *EntityHandler) CreateEntityRelationship(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req CreateEntityRelationshipRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.EntityID == uuid.Nil || req.RelatedType == "" || req.RelatedID == uuid.Nil || req.RelationshipType == "" {
		http.Error(w, "Entity ID, related type, related ID, and relationship type are required", http.StatusBadRequest)
		return
	}

	// Check if entity exists
	_, err := h.useCase.GetEntity(ctx, req.EntityID)
	if err != nil {
		http.Error(w, "Entity not found", http.StatusNotFound)
		return
	}

	input := entity.CreateEntityRelationshipInput{
		EntityID:         req.EntityID,
		RelatedType:      req.RelatedType,
		RelatedID:        req.RelatedID,
		RelationshipType: req.RelationshipType,
		Metadata:         req.Metadata,
	}

	relationship, err := h.useCase.CreateEntityRelationship(ctx, input)
	if err != nil {
		http.Error(w, "Failed to create entity relationship", http.StatusInternalServerError)
		return
	}

	response := EntityRelationshipResponse{
		ID:               relationship.ID,
		EntityID:         relationship.EntityID,
		RelatedType:      relationship.RelatedType,
		RelatedID:        relationship.RelatedID,
		RelationshipType: relationship.RelationshipType,
		Metadata:         relationship.Metadata,
		CreatedAt:        relationship.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:        relationship.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	httpAdapter.JSON(w, http.StatusCreated, response)
}

// GetEntityRelationships godoc
// @Summary Get entity relationships
// @Description Get relationships for an entity with optional type filters
// @Tags entities
// @Param id path string true "Entity ID"
// @Param related_type query string false "Related type filter"
// @Param relationship_type query string false "Relationship type filter"
// @Success 200 {array} EntityRelationshipResponse
// @Router /api/entities/{id}/relationships [get]
func (h *EntityHandler) GetEntityRelationships(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	entityIDStr := chi.URLParam(r, "id")

	entityID, err := uuid.Parse(entityIDStr)
	if err != nil {
		http.Error(w, "Invalid entity ID", http.StatusBadRequest)
		return
	}

	relatedTypeStr := r.URL.Query().Get("related_type")
	relationshipTypeStr := r.URL.Query().Get("relationship_type")

	var relatedType *string
	if relatedTypeStr != "" {
		relatedType = &relatedTypeStr
	}

	var relationshipType *string
	if relationshipTypeStr != "" {
		relationshipType = &relationshipTypeStr
	}

	relationships, err := h.useCase.GetEntityRelationships(ctx, entityID, relatedType, relationshipType)
	if err != nil {
		http.Error(w, "Failed to get entity relationships", http.StatusInternalServerError)
		return
	}

	response := make([]EntityRelationshipResponse, len(relationships))
	for i, rel := range relationships {
		response[i] = EntityRelationshipResponse{
			ID:               rel.ID,
			EntityID:         rel.EntityID,
			RelatedType:      rel.RelatedType,
			RelatedID:        rel.RelatedID,
			RelationshipType: rel.RelationshipType,
			Metadata:         rel.Metadata,
			CreatedAt:        rel.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:        rel.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	httpAdapter.JSON(w, http.StatusOK, response)
}

// DeleteEntityRelationship godoc
// @Summary Delete entity relationship
// @Description Delete an entity relationship by ID
// @Tags entities
// @Param id path string true "Relationship ID"
// @Success 204
// @Router /api/entities/relationship/{id} [delete]
func (h *EntityHandler) DeleteEntityRelationship(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	relationshipIDStr := chi.URLParam(r, "id")

	relationshipID, err := uuid.Parse(relationshipIDStr)
	if err != nil {
		http.Error(w, "Invalid relationship ID", http.StatusBadRequest)
		return
	}

	if err := h.useCase.DeleteEntityRelationship(ctx, relationshipID); err != nil {
		http.Error(w, "Failed to delete entity relationship", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

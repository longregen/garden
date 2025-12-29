package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	httpAdapter "garden3/internal/adapter/primary/http"
	"garden3/internal/domain/entity"
	"garden3/internal/port/input"
	"garden3/internal/port/output"
)

type LogseqHandler struct {
	useCase    input.LogseqSyncUseCase
	entityRepo output.EntityRepository
}

func NewLogseqHandler(useCase input.LogseqSyncUseCase, entityRepo output.EntityRepository) *LogseqHandler {
	return &LogseqHandler{
		useCase:    useCase,
		entityRepo: entityRepo,
	}
}

func (h *LogseqHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/sync/logseq", func(r chi.Router) {
		r.Post("/", h.Synchronize)
		r.Get("/check", h.PerformHardSyncCheck)
		r.Post("/force-git", h.ForceUpdateFileFromDB)
		r.Post("/force-db", h.ForceUpdateDBFromFile)
		// Backwards-compatible route with UUID in path (base project format)
		r.Post("/force-db/{uuid}", h.ForceUpdateDBFromFileByUUID)
	})
}

// SyncStatsResponse represents sync statistics in the API response
type SyncStatsResponse struct {
	PagesProcessed    int      `json:"pages_processed"`
	PagesCreated      int      `json:"pages_created"`
	PagesUpdated      int      `json:"pages_updated"`
	PagesSkipped      int      `json:"pages_skipped"`
	EntitiesProcessed int      `json:"entities_processed"`
	EntitiesCreated   int      `json:"entities_created"`
	EntitiesUpdated   int      `json:"entities_updated"`
	EntitiesSkipped   int      `json:"entities_skipped"`
	Errors            []string `json:"errors"`
}

// SyncCheckResponse represents the sync check result in the API response
type SyncCheckResponse struct {
	MissingInDB  []EntitySummary `json:"missing_in_db"`
	MissingInGit []EntitySummary `json:"missing_in_git"`
	OutOfSync    []OutOfSyncItem `json:"out_of_sync"`
}

// EntitySummary represents a simplified entity in sync check results
type EntitySummary struct {
	EntityID    string  `json:"entity_id"`
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	Description *string `json:"description,omitempty"`
}

// OutOfSyncItem represents an out-of-sync item in the API response
type OutOfSyncItem struct {
	Entity      EntitySummary `json:"entity"`
	PagePath    string        `json:"page_path"`
	LastSyncDB  *string       `json:"last_sync_db,omitempty"`
	LastSyncGit *string       `json:"last_sync_git,omitempty"`
}

// Synchronize godoc
// @Summary Trigger full Logseq sync
// @Description Synchronizes the Logseq git repository with the database
// @Tags sync
// @Success 200 {object} SyncStatsResponse
// @Router /api/sync/logseq [post]
func (h *LogseqHandler) Synchronize(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	stats, err := h.useCase.Synchronize(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := SyncStatsResponse{
		PagesProcessed:    stats.PagesProcessed,
		PagesCreated:      stats.PagesCreated,
		PagesUpdated:      stats.PagesUpdated,
		PagesSkipped:      stats.PagesSkipped,
		EntitiesProcessed: stats.EntitiesProcessed,
		EntitiesCreated:   stats.EntitiesCreated,
		EntitiesUpdated:   stats.EntitiesUpdated,
		EntitiesSkipped:   stats.EntitiesSkipped,
		Errors:            stats.Errors,
	}

	httpAdapter.JSON(w, http.StatusOK, response)
}

// PerformHardSyncCheck godoc
// @Summary Perform hard sync check
// @Description Compares all files in the Logseq folder with all entries in the database
// @Tags sync
// @Success 200 {object} SyncCheckResponse
// @Router /api/sync/logseq/check [get]
func (h *LogseqHandler) PerformHardSyncCheck(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	result, err := h.useCase.PerformHardSyncCheck(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := SyncCheckResponse{
		MissingInDB:  make([]EntitySummary, len(result.MissingInDB)),
		MissingInGit: make([]EntitySummary, len(result.MissingInGit)),
		OutOfSync:    make([]OutOfSyncItem, len(result.OutOfSync)),
	}

	for i, ent := range result.MissingInDB {
		response.MissingInDB[i] = EntitySummary{
			EntityID:    ent.EntityID.String(),
			Name:        ent.Name,
			Type:        ent.Type,
			Description: ent.Description,
		}
	}

	for i, ent := range result.MissingInGit {
		response.MissingInGit[i] = EntitySummary{
			EntityID:    ent.EntityID.String(),
			Name:        ent.Name,
			Type:        ent.Type,
			Description: ent.Description,
		}
	}

	for i, item := range result.OutOfSync {
		var lastSyncDB, lastSyncGit *string
		if item.LastSyncDB != nil {
			s := item.LastSyncDB.Format("2006-01-02T15:04:05Z")
			lastSyncDB = &s
		}
		if item.LastSyncGit != nil {
			s := item.LastSyncGit.Format("2006-01-02T15:04:05Z")
			lastSyncGit = &s
		}

		response.OutOfSync[i] = OutOfSyncItem{
			Entity: EntitySummary{
				EntityID:    item.Entity.EntityID.String(),
				Name:        item.Entity.Name,
				Type:        item.Entity.Type,
				Description: item.Entity.Description,
			},
			PagePath:    item.PagePath,
			LastSyncDB:  lastSyncDB,
			LastSyncGit: lastSyncGit,
		}
	}

	httpAdapter.JSON(w, http.StatusOK, response)
}

// ForceUpdateFileFromDB godoc
// @Summary Force update git from DB
// @Description Forces update of a git file with data from the database for an entity
// @Tags sync
// @Param body body entity.ForceUpdateRequest true "Entity ID"
// @Success 200
// @Router /api/sync/logseq/force-git [post]
func (h *LogseqHandler) ForceUpdateFileFromDB(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req entity.ForceUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.useCase.ForceUpdateFileFromDB(ctx, req.EntityID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "File updated successfully"}`))
}

// ForceUpdateDBFromFileRequest represents request for force update from file
type ForceUpdateDBFromFileRequest struct {
	PagePath string `json:"page_path"`
}

// ForceUpdateDBFromFile godoc
// @Summary Force update DB from git file
// @Description Forces update of database entry with data from a git file
// @Tags sync
// @Param body body ForceUpdateDBFromFileRequest true "File path"
// @Success 200 {object} EntityResponse
// @Router /api/sync/logseq/force-db [post]
func (h *LogseqHandler) ForceUpdateDBFromFile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req ForceUpdateDBFromFileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.PagePath == "" {
		http.Error(w, "page_path is required", http.StatusBadRequest)
		return
	}

	pagePath := req.PagePath

	ent, err := h.useCase.ForceUpdateDBFromFile(ctx, pagePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

// ForceUpdateDBFromFileByUUID godoc
// @Summary Force update DB from git file by entity UUID (backwards-compatible)
// @Description Forces update of database entry with data from a git file using entity UUID (base project format)
// @Tags sync
// @Param uuid path string true "Entity UUID"
// @Success 200 {object} EntityResponse
// @Router /api/sync/logseq/force-db/{uuid} [post]
func (h *LogseqHandler) ForceUpdateDBFromFileByUUID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get UUID from path parameter
	uuidStr := chi.URLParam(r, "uuid")
	if uuidStr == "" {
		http.Error(w, "UUID is required", http.StatusBadRequest)
		return
	}

	// Parse UUID
	entityID, err := uuid.Parse(uuidStr)
	if err != nil {
		http.Error(w, "Invalid UUID format", http.StatusBadRequest)
		return
	}

	// Get entity to retrieve its page_path property
	ent, err := h.entityRepo.GetEntity(ctx, entityID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Entity not found: %v", err), http.StatusNotFound)
		return
	}

	// Extract page_path from entity properties
	var pagePath string
	if ent.Properties != nil {
		var props map[string]interface{}
		if err := json.Unmarshal(ent.Properties, &props); err == nil {
			if pp, ok := props["page_path"].(string); ok {
				pagePath = pp
			}
		}
	}

	if pagePath == "" {
		http.Error(w, "Entity does not have a page_path property", http.StatusBadRequest)
		return
	}

	// Call the use case with the page_path
	updatedEnt, err := h.useCase.ForceUpdateDBFromFile(ctx, pagePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := EntityResponse{
		EntityID:    updatedEnt.EntityID,
		Name:        updatedEnt.Name,
		Type:        updatedEnt.Type,
		Description: updatedEnt.Description,
		Properties:  updatedEnt.Properties,
		CreatedAt:   updatedEnt.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   updatedEnt.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	httpAdapter.JSON(w, http.StatusOK, response)
}

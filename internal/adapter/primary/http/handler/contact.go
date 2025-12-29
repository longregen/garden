package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	httpAdapter "garden3/internal/adapter/primary/http"
	"garden3/internal/domain/entity"
	"garden3/internal/port/input"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ContactHandler struct {
	useCase input.ContactUseCase
}

func NewContactHandler(useCase input.ContactUseCase) *ContactHandler {
	return &ContactHandler{
		useCase: useCase,
	}
}

func (h *ContactHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/contacts", func(r chi.Router) {
		r.Get("/", h.ListContacts)
		r.Post("/", h.CreateContact)
		r.Post("/batch", h.BatchUpdate)
		r.Post("/merge", h.MergeContacts)
		r.Post("/stats/refresh", h.RefreshStats)

		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.GetContact)
			r.Put("/", h.UpdateContact)
			r.Delete("/", h.DeleteContact)
			r.Put("/evaluation", h.UpdateEvaluation)
			r.Get("/tags", h.GetTags)
			r.Post("/tags", h.AddTag)
			r.Delete("/tags/{tagId}", h.RemoveTag)
		})
	})

	r.Route("/api/contact-tags", func(r chi.Router) {
		r.Get("/", h.ListAllTags)
	})

	r.Route("/api/contact-sources", func(r chi.Router) {
		r.Get("/", h.ListAllContactSources)
		r.Post("/", h.CreateContactSource)
		r.Put("/{id}", h.UpdateContactSource)
		r.Delete("/{id}", h.DeleteContactSource)
	})
}

// FullContactResponse represents the complete contact response with all relations
type FullContactResponse struct {
	ContactID        uuid.UUID                 `json:"contact_id"`
	Name             string                    `json:"name"`
	Email            *string                   `json:"email"`
	Phone            *string                   `json:"phone"`
	Birthday         *string                   `json:"birthday"`
	Notes            *string                   `json:"notes"`
	Extras           map[string]interface{}    `json:"extras"`
	CreationDate     string                    `json:"creation_date"`
	LastUpdate       string                    `json:"last_update"`
	LastWeekMessages *int32                    `json:"last_week_messages"`
	GroupsInCommon   *int32                    `json:"groups_in_common"`
	Closeness        *int32                    `json:"closeness"`
	Importance       *int32                    `json:"importance"`
	Fondness         *int32                    `json:"fondness"`
	Evaluation       *entity.ContactEvaluation `json:"evaluation"`
	Tags             []entity.ContactTag       `json:"tags"`
	KnownNames       []entity.ContactKnownName `json:"known_names"`
	AlternativeNames []entity.ContactKnownName `json:"alternativeNames"` // Backwards compatibility alias for known_names
	Rooms            []entity.ContactRoom      `json:"rooms"`
	Sources          []entity.ContactSource    `json:"sources"`
}

// ContactListItemResponse represents a contact in a list with tags and evaluation
type ContactListItemResponse struct {
	ContactID        uuid.UUID              `json:"contact_id"`
	Name             string                 `json:"name"`
	Email            *string                `json:"email"`
	Phone            *string                `json:"phone"`
	Birthday         *string                `json:"birthday"`
	Notes            *string                `json:"notes"`
	Extras           map[string]interface{} `json:"extras"`
	CreationDate     string                 `json:"creation_date"`
	LastUpdate       string                 `json:"last_update"`
	LastWeekMessages *int32                 `json:"last_week_messages"`
	GroupsInCommon   *int32                 `json:"groups_in_common"`
	Importance       *int32                 `json:"importance"`
	Closeness        *int32                 `json:"closeness"`
	Fondness         *int32                 `json:"fondness"`
	Tags             []entity.ContactTag    `json:"tags"`
}

// GetContact godoc
// @Summary Get contact by ID
// @Description Get a single contact with all its relations
// @Tags contacts
// @Param id path string true "Contact ID"
// @Success 200 {object} FullContactResponse
// @Router /api/contacts/{id} [get]
func (h *ContactHandler) GetContact(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	contactIDStr := chi.URLParam(r, "id")

	contactID, err := uuid.Parse(contactIDStr)
	if err != nil {
		http.Error(w, "Invalid contact ID", http.StatusBadRequest)
		return
	}

	fullContact, err := h.useCase.GetContact(ctx, contactID)
	if err != nil {
		http.Error(w, "Contact not found", http.StatusNotFound)
		return
	}

	var closeness, importance, fondness *int32
	if fullContact.Evaluation != nil {
		closeness = fullContact.Evaluation.Closeness
		importance = fullContact.Evaluation.Importance
		fondness = fullContact.Evaluation.Fondness
	}

	response := FullContactResponse{
		ContactID:        fullContact.Contact.ContactID,
		Name:             fullContact.Contact.Name,
		Email:            fullContact.Contact.Email,
		Phone:            fullContact.Contact.Phone,
		Birthday:         fullContact.Contact.Birthday,
		Notes:            fullContact.Contact.Notes,
		Extras:           fullContact.Contact.Extras,
		CreationDate:     fullContact.Contact.CreationDate.Format("2006-01-02T15:04:05Z"),
		LastUpdate:       fullContact.Contact.LastUpdate.Format("2006-01-02T15:04:05Z"),
		LastWeekMessages: fullContact.Contact.LastWeekMessages,
		GroupsInCommon:   fullContact.Contact.GroupsInCommon,
		Closeness:        closeness,
		Importance:       importance,
		Fondness:         fondness,
		Evaluation:       fullContact.Evaluation,
		Tags:             fullContact.Tags,
		KnownNames:       fullContact.KnownNames,
		AlternativeNames: fullContact.KnownNames,
		Rooms:            fullContact.Rooms,
		Sources:          fullContact.Sources,
	}

	httpAdapter.JSON(w, http.StatusOK, response)
}

// ListContacts godoc
// @Summary List contacts
// @Description Get paginated list of contacts with optional search
// @Tags contacts
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(20)
// @Param search query string false "Search query"
// @Success 200 {object} input.PaginatedResponse[ContactListItemResponse]
// @Router /api/contacts [get]
func (h *ContactHandler) ListContacts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	page := int32(1)
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = int32(p)
		}
	}

	pageSize := int32(20)
	if pageSizeStr := r.URL.Query().Get("pageSize"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
			pageSize = int32(ps)
		}
	}

	var searchQuery *string
	// Support both 'search' (new) and 'searchQuery' (old/backwards-compatible)
	if search := r.URL.Query().Get("search"); search != "" {
		searchQuery = &search
	} else if search := r.URL.Query().Get("searchQuery"); search != "" {
		searchQuery = &search
	}

	result, err := h.useCase.ListContacts(ctx, page, pageSize, searchQuery)
	if err != nil {
		http.Error(w, "Failed to list contacts", http.StatusInternalServerError)
		return
	}

	httpAdapter.JSON(w, http.StatusOK, result)
}

// CreateContact godoc
// @Summary Create contact
// @Description Create a new contact
// @Tags contacts
// @Param contact body entity.CreateContactInput true "Contact data"
// @Success 201 {object} entity.Contact
// @Router /api/contacts [post]
func (h *ContactHandler) CreateContact(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var input entity.CreateContactInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if input.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	contact, err := h.useCase.CreateContact(ctx, input)
	if err != nil {
		http.Error(w, "Failed to create contact", http.StatusInternalServerError)
		return
	}

	httpAdapter.JSON(w, http.StatusCreated, contact)
}

// UpdateContact godoc
// @Summary Update contact
// @Description Update a contact's basic information
// @Tags contacts
// @Param id path string true "Contact ID"
// @Param contact body entity.UpdateContactInput true "Contact data"
// @Success 200 {object} map[string]string
// @Router /api/contacts/{id} [put]
func (h *ContactHandler) UpdateContact(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	contactIDStr := chi.URLParam(r, "id")

	contactID, err := uuid.Parse(contactIDStr)
	if err != nil {
		http.Error(w, "Invalid contact ID", http.StatusBadRequest)
		return
	}

	var input entity.UpdateContactInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.useCase.UpdateContact(ctx, contactID, input); err != nil {
		http.Error(w, "Failed to update contact", http.StatusInternalServerError)
		return
	}

	httpAdapter.JSON(w, http.StatusOK, map[string]string{
		"message": "Contact updated successfully",
	})
}

// DeleteContact godoc
// @Summary Delete contact
// @Description Delete a contact
// @Tags contacts
// @Param id path string true "Contact ID"
// @Success 200 {object} map[string]string
// @Router /api/contacts/{id} [delete]
func (h *ContactHandler) DeleteContact(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	contactIDStr := chi.URLParam(r, "id")

	contactID, err := uuid.Parse(contactIDStr)
	if err != nil {
		http.Error(w, "Invalid contact ID", http.StatusBadRequest)
		return
	}

	if err := h.useCase.DeleteContact(ctx, contactID); err != nil {
		http.Error(w, "Failed to delete contact", http.StatusInternalServerError)
		return
	}

	httpAdapter.JSON(w, http.StatusOK, map[string]string{
		"message": "Contact deleted successfully",
	})
}

// UpdateEvaluation godoc
// @Summary Update contact evaluation
// @Description Update or create a contact's evaluation scores
// @Tags contacts
// @Param id path string true "Contact ID"
// @Param evaluation body entity.UpdateEvaluationInput true "Evaluation data"
// @Success 200 {object} map[string]string
// @Router /api/contacts/{id}/evaluation [put]
func (h *ContactHandler) UpdateEvaluation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	contactIDStr := chi.URLParam(r, "id")

	contactID, err := uuid.Parse(contactIDStr)
	if err != nil {
		http.Error(w, "Invalid contact ID", http.StatusBadRequest)
		return
	}

	var input entity.UpdateEvaluationInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.useCase.UpdateEvaluation(ctx, contactID, input); err != nil {
		http.Error(w, "Failed to update evaluation", http.StatusInternalServerError)
		return
	}

	httpAdapter.JSON(w, http.StatusOK, map[string]string{
		"message": "Contact evaluation updated successfully",
	})
}

// MergeContacts godoc
// @Summary Merge contacts
// @Description Merge two contacts together
// @Tags contacts
// @Param merge body map[string]string true "Source and target contact IDs"
// @Success 200 {object} map[string]string
// @Router /api/contacts/merge [post]
func (h *ContactHandler) MergeContacts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var input struct {
		SourceContactID string `json:"source_contact_id"`
		TargetContactID string `json:"target_contact_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	sourceID, err := uuid.Parse(input.SourceContactID)
	if err != nil {
		http.Error(w, "Invalid source contact ID", http.StatusBadRequest)
		return
	}

	targetID, err := uuid.Parse(input.TargetContactID)
	if err != nil {
		http.Error(w, "Invalid target contact ID", http.StatusBadRequest)
		return
	}

	if err := h.useCase.MergeContacts(ctx, sourceID, targetID); err != nil {
		http.Error(w, "Failed to merge contacts", http.StatusInternalServerError)
		return
	}

	httpAdapter.JSON(w, http.StatusOK, map[string]string{
		"message": "Contacts merged successfully",
	})
}

// BatchUpdate godoc
// @Summary Batch update contacts
// @Description Update multiple contacts atomically
// @Tags contacts
// @Param contacts body []entity.BatchContactUpdate true "Contact updates"
// @Success 200 {array} entity.BatchUpdateResult
// @Router /api/contacts/batch [post]
func (h *ContactHandler) BatchUpdate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var updates []entity.BatchContactUpdate
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	results, err := h.useCase.BatchUpdate(ctx, updates)
	if err != nil {
		http.Error(w, "Failed to batch update contacts", http.StatusInternalServerError)
		return
	}

	httpAdapter.JSON(w, http.StatusOK, results)
}

// RefreshStats godoc
// @Summary Refresh contact statistics
// @Description Update message statistics for all contacts
// @Tags contacts
// @Success 200 {object} map[string]string
// @Router /api/contacts/stats/refresh [post]
func (h *ContactHandler) RefreshStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := h.useCase.RefreshStats(ctx); err != nil {
		http.Error(w, "Failed to refresh stats", http.StatusInternalServerError)
		return
	}

	httpAdapter.JSON(w, http.StatusOK, map[string]string{
		"message": "Stats refreshed successfully",
	})
}

// GetTags godoc
// @Summary Get contact tags
// @Description Get all tags for a contact
// @Tags contacts
// @Param id path string true "Contact ID"
// @Success 200 {array} entity.ContactTag
// @Router /api/contacts/{id}/tags [get]
func (h *ContactHandler) GetTags(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	contactIDStr := chi.URLParam(r, "id")

	contactID, err := uuid.Parse(contactIDStr)
	if err != nil {
		http.Error(w, "Invalid contact ID", http.StatusBadRequest)
		return
	}

	tags, err := h.useCase.GetContactTags(ctx, contactID)
	if err != nil {
		http.Error(w, "Failed to get tags", http.StatusInternalServerError)
		return
	}

	httpAdapter.JSON(w, http.StatusOK, tags)
}

// AddTag godoc
// @Summary Add tag to contact
// @Description Add a tag to a contact (creates tag if it doesn't exist)
// @Tags contacts
// @Param id path string true "Contact ID"
// @Param tag body map[string]string true "Tag name"
// @Success 200 {object} entity.ContactTag
// @Router /api/contacts/{id}/tags [post]
func (h *ContactHandler) AddTag(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	contactIDStr := chi.URLParam(r, "id")

	contactID, err := uuid.Parse(contactIDStr)
	if err != nil {
		http.Error(w, "Invalid contact ID", http.StatusBadRequest)
		return
	}

	var input struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if input.Name == "" {
		http.Error(w, "Tag name is required", http.StatusBadRequest)
		return
	}

	tag, err := h.useCase.AddTag(ctx, contactID, input.Name)
	if err != nil {
		http.Error(w, "Failed to add tag", http.StatusInternalServerError)
		return
	}

	httpAdapter.JSON(w, http.StatusOK, tag)
}

// RemoveTag godoc
// @Summary Remove tag from contact
// @Description Remove a tag from a contact
// @Tags contacts
// @Param id path string true "Contact ID"
// @Param tagId path string true "Tag ID"
// @Success 200 {object} map[string]string
// @Router /api/contacts/{id}/tags/{tagId} [delete]
func (h *ContactHandler) RemoveTag(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	contactIDStr := chi.URLParam(r, "id")
	tagIDStr := chi.URLParam(r, "tagId")

	contactID, err := uuid.Parse(contactIDStr)
	if err != nil {
		http.Error(w, "Invalid contact ID", http.StatusBadRequest)
		return
	}

	tagID, err := uuid.Parse(tagIDStr)
	if err != nil {
		http.Error(w, "Invalid tag ID", http.StatusBadRequest)
		return
	}

	if err := h.useCase.RemoveTag(ctx, contactID, tagID); err != nil {
		http.Error(w, "Failed to remove tag", http.StatusInternalServerError)
		return
	}

	httpAdapter.JSON(w, http.StatusOK, map[string]string{
		"message": "Tag removed successfully",
	})
}

// ListAllTags godoc
// @Summary List all tags
// @Description Get all contact tag names in the system
// @Tags contacts
// @Success 200 {array} entity.ContactTagName
// @Router /api/contact-tags [get]
func (h *ContactHandler) ListAllTags(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tags, err := h.useCase.ListAllTags(ctx)
	if err != nil {
		http.Error(w, "Failed to list tags", http.StatusInternalServerError)
		return
	}

	httpAdapter.JSON(w, http.StatusOK, tags)
}

// ListAllContactSources godoc
// @Summary List all contact sources
// @Description Get all contact sources in the system
// @Tags contact-sources
// @Success 200 {array} entity.ContactSource
// @Router /api/contact-sources [get]
func (h *ContactHandler) ListAllContactSources(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	sources, err := h.useCase.ListAllContactSources(ctx)
	if err != nil {
		http.Error(w, "Failed to list contact sources", http.StatusInternalServerError)
		return
	}

	httpAdapter.JSON(w, http.StatusOK, sources)
}

// CreateContactSource godoc
// @Summary Create contact source
// @Description Create a new contact source
// @Tags contact-sources
// @Param source body entity.CreateContactSourceInput true "Contact source data"
// @Success 201 {object} entity.ContactSource
// @Router /api/contact-sources [post]
func (h *ContactHandler) CreateContactSource(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var input entity.CreateContactSourceInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if input.SourceID == "" || input.SourceName == "" {
		http.Error(w, "source_id and source_name are required", http.StatusBadRequest)
		return
	}

	source, err := h.useCase.CreateContactSource(ctx, input)
	if err != nil {
		http.Error(w, "Failed to create contact source", http.StatusInternalServerError)
		return
	}

	httpAdapter.JSON(w, http.StatusCreated, source)
}

// UpdateContactSource godoc
// @Summary Update contact source
// @Description Update a contact source
// @Tags contact-sources
// @Param id path string true "Contact Source ID"
// @Param source body entity.UpdateContactSourceInput true "Contact source data"
// @Success 200 {object} entity.ContactSource
// @Router /api/contact-sources/{id} [put]
func (h *ContactHandler) UpdateContactSource(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid contact source ID", http.StatusBadRequest)
		return
	}

	var input entity.UpdateContactSourceInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if input.SourceID == "" || input.SourceName == "" {
		http.Error(w, "source_id and source_name are required", http.StatusBadRequest)
		return
	}

	source, err := h.useCase.UpdateContactSource(ctx, id, input)
	if err != nil {
		http.Error(w, "Failed to update contact source", http.StatusInternalServerError)
		return
	}

	httpAdapter.JSON(w, http.StatusOK, source)
}

// DeleteContactSource godoc
// @Summary Delete contact source
// @Description Delete a contact source
// @Tags contact-sources
// @Param id path string true "Contact Source ID"
// @Success 200 {object} map[string]string
// @Router /api/contact-sources/{id} [delete]
func (h *ContactHandler) DeleteContactSource(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid contact source ID", http.StatusBadRequest)
		return
	}

	if err := h.useCase.DeleteContactSource(ctx, id); err != nil {
		http.Error(w, "Failed to delete contact source", http.StatusInternalServerError)
		return
	}

	httpAdapter.JSON(w, http.StatusOK, map[string]string{
		"message": "Contact source deleted successfully",
	})
}

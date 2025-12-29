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

type NoteHandler struct {
	useCase input.NoteUseCase
}

func NewNoteHandler(useCase input.NoteUseCase) *NoteHandler {
	return &NoteHandler{
		useCase: useCase,
	}
}

func (h *NoteHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/notes", func(r chi.Router) {
		r.Get("/", h.ListNotes)
		r.Get("/search", h.SearchNotes)
		r.Post("/", h.CreateNote)
		r.Get("/{id}", h.GetNote)
		r.Put("/{id}", h.UpdateNote)
		r.Delete("/{id}", h.DeleteNote)
	})

	r.Route("/api/tags", func(r chi.Router) {
		r.Get("/", h.ListTags)
	})
}

// NoteResponse represents the API response for a single note
type NoteResponse struct {
	ID               string   `json:"id"`
	Title            *string  `json:"title"`
	Contents         *string  `json:"contents"`
	ProcessedContent *string  `json:"processedContents"`
	Tags             []string `json:"tags"`
	Created          int64    `json:"created"`
	Modified         int64    `json:"modified"`
	EntityID         *string  `json:"entity_id,omitempty"`
}

// NotesListResponse represents the API response for note list
type NotesListResponse struct {
	Notes      []NoteListItemResponse `json:"notes"`
	TotalPages int32                  `json:"totalPages"`
}

// NoteListItemResponse represents a note in list view
type NoteListItemResponse struct {
	ID       string   `json:"id"`
	Title    *string  `json:"title"`
	Tags     []string `json:"tags"`
	Created  int64    `json:"created"`
	Modified int64    `json:"modified"`
}

// GetNote godoc
// @Summary Get note by ID
// @Description Get a single note with tags and processed content
// @Tags notes
// @Param id path string true "Note ID"
// @Success 200 {object} NoteResponse
// @Router /api/notes/{id} [get]
func (h *NoteHandler) GetNote(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	noteIDStr := chi.URLParam(r, "id")

	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		http.Error(w, "Invalid note ID", http.StatusBadRequest)
		return
	}

	fullNote, err := h.useCase.GetNote(ctx, noteID)
	if err != nil {
		http.Error(w, "Note not found", http.StatusNotFound)
		return
	}

	var entityIDStr *string
	if fullNote.EntityID != nil {
		idStr := fullNote.EntityID.String()
		entityIDStr = &idStr
	}

	response := NoteResponse{
		ID:               fullNote.Note.ID.String(),
		Title:            fullNote.Note.Title,
		Contents:         fullNote.Note.Contents,
		ProcessedContent: fullNote.ProcessedContent,
		Tags:             fullNote.Tags,
		Created:          fullNote.Note.Created,
		Modified:         fullNote.Note.Modified,
		EntityID:         entityIDStr,
	}

	httpAdapter.JSON(w, http.StatusOK, response)
}

// ListNotes godoc
// @Summary List notes
// @Description Get paginated list of notes with optional search
// @Tags notes
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(12)
// @Param searchQuery query string false "Search query"
// @Success 200 {object} NotesListResponse
// @Router /api/notes [get]
func (h *NoteHandler) ListNotes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	page := int32(1)
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = int32(p)
		}
	}

	pageSize := int32(12)
	if pageSizeStr := r.URL.Query().Get("pageSize"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
			pageSize = int32(ps)
		}
	}

	var searchQuery *string
	if search := r.URL.Query().Get("searchQuery"); search != "" {
		searchQuery = &search
	}

	result, err := h.useCase.ListNotes(ctx, page, pageSize, searchQuery)
	if err != nil {
		http.Error(w, "Failed to list notes", http.StatusInternalServerError)
		return
	}

	notes := make([]NoteListItemResponse, len(result.Data))
	for i, item := range result.Data {
		notes[i] = NoteListItemResponse{
			ID:       item.ID.String(),
			Title:    item.Title,
			Tags:     item.Tags,
			Created:  item.Created,
			Modified: item.Modified,
		}
	}

	response := NotesListResponse{
		Notes:      notes,
		TotalPages: result.TotalPages,
	}

	httpAdapter.JSON(w, http.StatusOK, response)
}

// CreateNote godoc
// @Summary Create note
// @Description Create a new note with tags and entity relationships
// @Tags notes
// @Param note body entity.CreateNoteInput true "Note data"
// @Success 201 {object} NoteResponse
// @Router /api/notes [post]
func (h *NoteHandler) CreateNote(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var input entity.CreateNoteInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if input.Title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	fullNote, err := h.useCase.CreateNote(ctx, input)
	if err != nil {
		http.Error(w, "Failed to create note", http.StatusInternalServerError)
		return
	}

	var entityIDStr *string
	if fullNote.EntityID != nil {
		idStr := fullNote.EntityID.String()
		entityIDStr = &idStr
	}

	response := NoteResponse{
		ID:               fullNote.Note.ID.String(),
		Title:            fullNote.Note.Title,
		Contents:         fullNote.Note.Contents,
		ProcessedContent: fullNote.ProcessedContent,
		Tags:             fullNote.Tags,
		Created:          fullNote.Note.Created,
		Modified:         fullNote.Note.Modified,
		EntityID:         entityIDStr,
	}

	httpAdapter.JSON(w, http.StatusCreated, response)
}

// UpdateNote godoc
// @Summary Update note
// @Description Update a note's information and tags
// @Tags notes
// @Param id path string true "Note ID"
// @Param note body entity.UpdateNoteInput true "Note data"
// @Success 200 {object} NoteResponse
// @Router /api/notes/{id} [put]
func (h *NoteHandler) UpdateNote(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	noteIDStr := chi.URLParam(r, "id")

	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		http.Error(w, "Invalid note ID", http.StatusBadRequest)
		return
	}

	var input entity.UpdateNoteInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	fullNote, err := h.useCase.UpdateNote(ctx, noteID, input)
	if err != nil {
		http.Error(w, "Failed to update note", http.StatusInternalServerError)
		return
	}

	var entityIDStr *string
	if fullNote.EntityID != nil {
		idStr := fullNote.EntityID.String()
		entityIDStr = &idStr
	}

	response := NoteResponse{
		ID:               fullNote.Note.ID.String(),
		Title:            fullNote.Note.Title,
		Contents:         fullNote.Note.Contents,
		ProcessedContent: fullNote.ProcessedContent,
		Tags:             fullNote.Tags,
		Created:          fullNote.Note.Created,
		Modified:         fullNote.Note.Modified,
		EntityID:         entityIDStr,
	}

	httpAdapter.JSON(w, http.StatusOK, response)
}

// DeleteNote godoc
// @Summary Delete note
// @Description Delete a note and all related data
// @Tags notes
// @Param id path string true "Note ID"
// @Success 200 {object} map[string]string
// @Router /api/notes/{id} [delete]
func (h *NoteHandler) DeleteNote(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	noteIDStr := chi.URLParam(r, "id")

	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		http.Error(w, "Invalid note ID", http.StatusBadRequest)
		return
	}

	if err := h.useCase.DeleteNote(ctx, noteID); err != nil {
		http.Error(w, "Failed to delete note", http.StatusInternalServerError)
		return
	}

	httpAdapter.JSON(w, http.StatusOK, map[string]string{
		"message": "Note deleted successfully",
	})
}

// ListTags godoc
// @Summary List all tags
// @Description Get all available tags in the system
// @Tags notes
// @Success 200 {array} entity.NoteTag
// @Router /api/tags [get]
func (h *NoteHandler) ListTags(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tags, err := h.useCase.ListAllTags(ctx)
	if err != nil {
		http.Error(w, "Failed to list tags", http.StatusInternalServerError)
		return
	}

	httpAdapter.JSON(w, http.StatusOK, tags)
}

// SearchNotes godoc
// @Summary Search notes
// @Description Perform vector similarity search on notes
// @Tags notes
// @Param q query string true "Search query"
// @Param strategy query string false "Search strategy" default(qa-v2-passage)
// @Success 200 {array} NoteListItemResponse
// @Router /api/notes/search [get]
func (h *NoteHandler) SearchNotes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	strategy := r.URL.Query().Get("strategy")
	if strategy == "" {
		strategy = "qa-v2-passage"
	}

	results, err := h.useCase.SearchSimilarNotes(ctx, query, strategy)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	notes := make([]NoteListItemResponse, len(results))
	for i, item := range results {
		notes[i] = NoteListItemResponse{
			ID:       item.ID.String(),
			Title:    item.Title,
			Tags:     item.Tags,
			Created:  item.Created,
			Modified: item.Modified,
		}
	}

	httpAdapter.JSON(w, http.StatusOK, notes)
}

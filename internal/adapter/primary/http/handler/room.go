package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	httpAdapter "garden3/internal/adapter/primary/http"
	"garden3/internal/port/input"
)

type RoomHandler struct {
	useCase input.RoomUseCase
}

func NewRoomHandler(useCase input.RoomUseCase) *RoomHandler {
	return &RoomHandler{
		useCase: useCase,
	}
}

func (h *RoomHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/rooms", func(r chi.Router) {
		r.Get("/", h.ListRooms)

		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.GetRoomDetails)
			r.Get("/messages", h.GetRoomMessages)
			r.Get("/messages/search", h.SearchRoomMessages)
			r.Put("/name", h.SetRoomName)
			r.Get("/sessions/count", h.GetSessionsCount)
		})
	})
}

// ListRooms handles GET /api/rooms
func (h *RoomHandler) ListRooms(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.ParseInt(r.URL.Query().Get("page"), 10, 32)
	pageSize, _ := strconv.ParseInt(r.URL.Query().Get("pageSize"), 10, 32)
	searchText := r.URL.Query().Get("searchText")

	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 10
	}

	var searchTextPtr *string
	if searchText != "" {
		searchTextPtr = &searchText
	}

	result, err := h.useCase.ListRooms(r.Context(), int32(page), int32(pageSize), searchTextPtr)
	if err != nil {
		httpAdapter.JSON(w, http.StatusInternalServerError, httpAdapter.ErrorResponse{Error: "Failed to list rooms"})
		return
	}

	httpAdapter.JSON(w, http.StatusOK, result)
}

// GetRoomDetails handles GET /api/rooms/:id
func (h *RoomHandler) GetRoomDetails(w http.ResponseWriter, r *http.Request) {
	roomIDStr := chi.URLParam(r, "id")
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		httpAdapter.JSON(w, http.StatusBadRequest, httpAdapter.ErrorResponse{Error: "Invalid room ID"})
		return
	}

	details, err := h.useCase.GetRoomDetails(r.Context(), roomID)
	if err != nil {
		httpAdapter.JSON(w, http.StatusInternalServerError, httpAdapter.ErrorResponse{Error: "Failed to get room details"})
		return
	}

	if details == nil {
		httpAdapter.JSON(w, http.StatusNotFound, httpAdapter.ErrorResponse{Error: "Room not found"})
		return
	}

	httpAdapter.JSON(w, http.StatusOK, details)
}

// GetRoomMessages handles GET /api/rooms/:id/messages
func (h *RoomHandler) GetRoomMessages(w http.ResponseWriter, r *http.Request) {
	roomIDStr := chi.URLParam(r, "id")
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		httpAdapter.JSON(w, http.StatusBadRequest, httpAdapter.ErrorResponse{Error: "Invalid room ID"})
		return
	}

	page, _ := strconv.ParseInt(r.URL.Query().Get("page"), 10, 32)
	pageSize, _ := strconv.ParseInt(r.URL.Query().Get("pageSize"), 10, 32)
	beforeMessageIDStr := r.URL.Query().Get("beforeMessageId")

	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 50
	}

	var beforeMessageID *uuid.UUID
	if beforeMessageIDStr != "" {
		id, err := uuid.Parse(beforeMessageIDStr)
		if err == nil {
			beforeMessageID = &id
		}
	}

	result, err := h.useCase.GetRoomMessages(r.Context(), roomID, int32(page), int32(pageSize), beforeMessageID)
	if err != nil {
		httpAdapter.JSON(w, http.StatusInternalServerError, httpAdapter.ErrorResponse{Error: "Failed to get room messages"})
		return
	}

	httpAdapter.JSON(w, http.StatusOK, result)
}

// SearchRoomMessages handles GET /api/rooms/:id/messages/search
func (h *RoomHandler) SearchRoomMessages(w http.ResponseWriter, r *http.Request) {
	roomIDStr := chi.URLParam(r, "id")
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		httpAdapter.JSON(w, http.StatusBadRequest, httpAdapter.ErrorResponse{Error: "Invalid room ID"})
		return
	}

	searchText := r.URL.Query().Get("searchText")
	if searchText == "" {
		httpAdapter.JSON(w, http.StatusBadRequest, httpAdapter.ErrorResponse{Error: "Search text is required"})
		return
	}

	page, _ := strconv.ParseInt(r.URL.Query().Get("page"), 10, 32)
	pageSize, _ := strconv.ParseInt(r.URL.Query().Get("pageSize"), 10, 32)

	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 50
	}

	result, err := h.useCase.SearchRoomMessages(r.Context(), roomID, searchText, int32(page), int32(pageSize))
	if err != nil {
		httpAdapter.JSON(w, http.StatusInternalServerError, httpAdapter.ErrorResponse{Error: "Failed to search room messages"})
		return
	}

	httpAdapter.JSON(w, http.StatusOK, result)
}

// SetRoomName handles PUT /api/rooms/:id/name
func (h *RoomHandler) SetRoomName(w http.ResponseWriter, r *http.Request) {
	roomIDStr := chi.URLParam(r, "id")
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		httpAdapter.JSON(w, http.StatusBadRequest, httpAdapter.ErrorResponse{Error: "Invalid room ID"})
		return
	}

	var req struct {
		Name *string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpAdapter.JSON(w, http.StatusBadRequest, httpAdapter.ErrorResponse{Error: "Invalid request body"})
		return
	}

	if err := h.useCase.SetRoomName(r.Context(), roomID, req.Name); err != nil {
		httpAdapter.JSON(w, http.StatusInternalServerError, httpAdapter.ErrorResponse{Error: "Failed to update room name"})
		return
	}

	httpAdapter.JSON(w, http.StatusOK, map[string]string{"status": "success"})
}

// GetSessionsCount handles GET /api/rooms/:id/sessions/count
func (h *RoomHandler) GetSessionsCount(w http.ResponseWriter, r *http.Request) {
	roomIDStr := chi.URLParam(r, "id")
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		httpAdapter.JSON(w, http.StatusBadRequest, httpAdapter.ErrorResponse{Error: "Invalid room ID"})
		return
	}

	count, err := h.useCase.GetSessionsCount(r.Context(), roomID)
	if err != nil {
		httpAdapter.JSON(w, http.StatusInternalServerError, httpAdapter.ErrorResponse{Error: "Failed to get sessions count"})
		return
	}

	httpAdapter.JSON(w, http.StatusOK, map[string]int32{"count": count})
}

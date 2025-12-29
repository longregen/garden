package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	httpAdapter "garden3/internal/adapter/primary/http"
	"garden3/internal/port/input"
)

type SessionHandler struct {
	useCase input.SessionUseCase
}

func NewSessionHandler(useCase input.SessionUseCase) *SessionHandler {
	return &SessionHandler{
		useCase: useCase,
	}
}

func (h *SessionHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/sessions", func(r chi.Router) {
		r.Get("/search", h.SearchSessions)

		r.Route("/{id}", func(r chi.Router) {
			r.Get("/messages", h.GetSessionMessages)
		})
	})

	r.Route("/api/rooms/{id}", func(r chi.Router) {
		r.Get("/sessions", h.GetRoomSessions)
		r.Get("/timeline", h.GetTimeline)
	})

	r.Route("/api/contacts/{id}", func(r chi.Router) {
		r.Get("/sessions/search", h.SearchContactSessions)
	})
}

// SearchSessions handles GET /api/sessions/search?q=query&limit=10
func (h *SessionHandler) SearchSessions(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		httpAdapter.BadRequest(w, errors.New("query parameter 'q' is required"))
		return
	}

	limit := int32(10)
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		parsedLimit, err := strconv.ParseInt(limitStr, 10, 32)
		if err != nil {
			httpAdapter.BadRequest(w, errors.New("invalid limit parameter"))
			return
		}
		limit = int32(parsedLimit)
	}

	results, err := h.useCase.SearchSessions(r.Context(), query, limit)
	if err != nil {
		httpAdapter.InternalError(w, err)
		return
	}

	httpAdapter.JSON(w, http.StatusOK, results)
}

// GetRoomSessions handles GET /api/rooms/:id/sessions
func (h *SessionHandler) GetRoomSessions(w http.ResponseWriter, r *http.Request) {
	roomIDStr := chi.URLParam(r, "id")
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		httpAdapter.BadRequest(w, errors.New("invalid room ID"))
		return
	}

	sessions, err := h.useCase.GetRoomSessions(r.Context(), roomID)
	if err != nil {
		httpAdapter.InternalError(w, err)
		return
	}

	httpAdapter.JSON(w, http.StatusOK, sessions)
}

// SearchContactSessions handles GET /api/contacts/:id/sessions/search?q=query
func (h *SessionHandler) SearchContactSessions(w http.ResponseWriter, r *http.Request) {
	contactIDStr := chi.URLParam(r, "id")
	contactID, err := uuid.Parse(contactIDStr)
	if err != nil {
		httpAdapter.BadRequest(w, errors.New("invalid contact ID"))
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		httpAdapter.BadRequest(w, errors.New("query parameter 'q' is required"))
		return
	}

	results, err := h.useCase.SearchContactSessions(r.Context(), contactID, query)
	if err != nil {
		httpAdapter.InternalError(w, err)
		return
	}

	httpAdapter.JSON(w, http.StatusOK, results)
}

// GetTimeline handles GET /api/rooms/:id/timeline
func (h *SessionHandler) GetTimeline(w http.ResponseWriter, r *http.Request) {
	roomIDStr := chi.URLParam(r, "id")
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		httpAdapter.BadRequest(w, errors.New("invalid room ID"))
		return
	}

	timeline, err := h.useCase.GetTimeline(r.Context(), roomID)
	if err != nil {
		httpAdapter.InternalError(w, err)
		return
	}

	httpAdapter.JSON(w, http.StatusOK, timeline)
}

// GetSessionMessages handles GET /api/sessions/:id/messages
func (h *SessionHandler) GetSessionMessages(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := chi.URLParam(r, "id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		httpAdapter.BadRequest(w, errors.New("invalid session ID"))
		return
	}

	response, err := h.useCase.GetSessionMessages(r.Context(), sessionID)
	if err != nil {
		httpAdapter.InternalError(w, err)
		return
	}

	httpAdapter.JSON(w, http.StatusOK, response)
}

package handler

import (
	"net/http"
	"strconv"

	httpAdapter "garden3/internal/adapter/primary/http"
	"garden3/internal/port/input"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type MessageHandler struct {
	useCase input.MessageUseCase
}

func NewMessageHandler(useCase input.MessageUseCase) *MessageHandler {
	return &MessageHandler{
		useCase: useCase,
	}
}

func (h *MessageHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/messages", func(r chi.Router) {
		r.Get("/content", h.GetAllMessageContents)
		r.Get("/search", h.SearchMessages)

		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.GetMessage)
			r.Get("/text-representations", h.GetMessageTextRepresentations)
		})
	})

	r.Route("/api/rooms/{roomId}/messages", func(r chi.Router) {
		r.Get("/", h.GetMessagesByRoomID)
	})
}

// GetMessage handles GET /api/messages/:id
func (h *MessageHandler) GetMessage(w http.ResponseWriter, r *http.Request) {
	messageIDStr := chi.URLParam(r, "id")
	messageID, err := uuid.Parse(messageIDStr)
	if err != nil {
		httpAdapter.BadRequest(w, err)
		return
	}

	message, err := h.useCase.GetMessage(r.Context(), messageID)
	if err != nil {
		httpAdapter.InternalError(w, err)
		return
	}

	if message == nil {
		httpAdapter.NotFound(w)
		return
	}

	httpAdapter.JSON(w, http.StatusOK, message)
}

// GetMessagesByRoomID handles GET /api/rooms/:roomId/messages
func (h *MessageHandler) GetMessagesByRoomID(w http.ResponseWriter, r *http.Request) {
	roomIDStr := chi.URLParam(r, "roomId")
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		httpAdapter.BadRequest(w, err)
		return
	}

	page, _ := strconv.ParseInt(r.URL.Query().Get("page"), 10, 32)
	pageSize, _ := strconv.ParseInt(r.URL.Query().Get("pageSize"), 10, 32)

	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 100
	}

	result, err := h.useCase.GetMessagesByRoomID(r.Context(), roomID, int32(page), int32(pageSize), nil)
	if err != nil {
		httpAdapter.InternalError(w, err)
		return
	}

	httpAdapter.JSON(w, http.StatusOK, result)
}

// GetAllMessageContents handles GET /api/messages/content
func (h *MessageHandler) GetAllMessageContents(w http.ResponseWriter, r *http.Request) {
	contents, err := h.useCase.GetAllMessageContents(r.Context())
	if err != nil {
		httpAdapter.InternalError(w, err)
		return
	}

	httpAdapter.JSON(w, http.StatusOK, map[string]interface{}{
		"data": contents,
	})
}

// GetMessageTextRepresentations handles GET /api/messages/:id/text-representations
func (h *MessageHandler) GetMessageTextRepresentations(w http.ResponseWriter, r *http.Request) {
	messageIDStr := chi.URLParam(r, "id")
	messageID, err := uuid.Parse(messageIDStr)
	if err != nil {
		httpAdapter.BadRequest(w, err)
		return
	}

	reps, err := h.useCase.GetMessageTextRepresentations(r.Context(), messageID)
	if err != nil {
		httpAdapter.InternalError(w, err)
		return
	}

	httpAdapter.JSON(w, http.StatusOK, reps)
}

// SearchMessages handles GET /api/messages/search
func (h *MessageHandler) SearchMessages(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		httpAdapter.JSON(w, http.StatusBadRequest, httpAdapter.ErrorResponse{Error: "Query parameter 'q' is required"})
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

	result, err := h.useCase.SearchMessages(r.Context(), query, int32(page), int32(pageSize))
	if err != nil {
		httpAdapter.InternalError(w, err)
		return
	}

	httpAdapter.JSON(w, http.StatusOK, result)
}

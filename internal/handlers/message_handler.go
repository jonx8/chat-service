package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/jonx8/chat-service/internal/dto"
	"github.com/jonx8/chat-service/internal/services"
)

type MessageHandler struct {
	messageService services.MessageService
}

func NewMessageHandler(messageService services.MessageService) *MessageHandler {
	return &MessageHandler{messageService: messageService}
}

func (h *MessageHandler) CreateMessage(w http.ResponseWriter, r *http.Request) {
	chatID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "BAD_REQUEST", "ID path param must be integer")
		return
	}

	var request dto.CreateMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeJSONError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid json")
		return
	}

	if len(request.Text) < 1 || len(request.Text) > 5000 {
		writeJSONError(w, http.StatusBadRequest, "BAD_REQUEST", "Message length must be between 1 and 5000")
		return
	}

	message, err := h.messageService.CreateMessage(r.Context(), chatID, &request)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrChatNotFound):
			writeJSONError(w, http.StatusNotFound, "NOT_FOUND", "Chat not found")
		default:
			slog.Error("Failed to create new message", "error", err)
			writeJSONError(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Internal Server Error")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(message); err != nil {
		slog.Error("Failed to serialize message", "error", err, "message", message)
	}

}

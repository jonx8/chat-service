package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/jonx8/chat-service/internal/dto"
	"github.com/jonx8/chat-service/internal/services"
)

type ChatHandler struct {
	chatService services.ChatService
}

func NewChatHandler(chatService services.ChatService) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
	}
}

func (h *ChatHandler) GetChat(w http.ResponseWriter, r *http.Request) {
	chatID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "BAD_REQUEST", "ID path param must be integer")
		return
	}

	limit := 20
	if limitParam := r.URL.Query().Get("limit"); limitParam != "" {
		if val, err := strconv.Atoi(limitParam); err == nil && val >= 1 && val <= 100 {
			limit = val
		}
	}
	chat, err := h.chatService.GetChat(r.Context(), chatID, limit)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrChatNotFound):
			writeJSONError(w, http.StatusNotFound, "NOT_FOUND", "Chat not found")
		default:
			slog.Error("Failed to get chat", "error", err, "chatID", chatID)
			writeJSONError(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Internal Server Error")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(chat)

}

func (h *ChatHandler) CreateChat(w http.ResponseWriter, r *http.Request) {
	var request dto.CreateChatRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeJSONError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid json")
		return
	}

	request.Title = strings.TrimSpace(request.Title)

	if len(request.Title) < 1 || len(request.Title) > 200 {
		writeJSONError(w, http.StatusBadRequest, "BAD_REQUEST", "Title length must be between 1 and 200")
		return
	}

	chat, err := h.chatService.CreateChat(r.Context(), &request)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrChatAlreadyExists):
			message := fmt.Sprintf("Chat with title %s already exists", request.Title)
			writeJSONError(w, http.StatusConflict, "CONFLICT", message)
		default:
			slog.Error("Failed to create chat", "error", err, "request", r)
			writeJSONError(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Internal Server Error")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(chat)

}

func (h *ChatHandler) DeleteChat(w http.ResponseWriter, r *http.Request) {
	chatID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "BAD_REQUEST", "ID path param must be integer")
		return
	}

	if err := h.chatService.DeleteChat(r.Context(), chatID); err != nil {
		switch {
		case errors.Is(err, services.ErrChatNotFound):
			writeJSONError(w, http.StatusNotFound, "NOT_FOUND", "Chat not found")
		default:
			slog.Error("Failed to delete chat", "error", err, "chatID", chatID)
			writeJSONError(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Internal Server Error")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)

}

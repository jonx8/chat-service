package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jonx8/chat-service/internal/dto"
	"github.com/jonx8/chat-service/internal/handlers"
	"github.com/jonx8/chat-service/internal/models"
	"github.com/jonx8/chat-service/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockChatService struct {
	mock.Mock
}

func (m *MockChatService) CreateChat(ctx context.Context, req *dto.CreateChatRequest) (*models.Chat, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Chat), args.Error(1)
}

func (m *MockChatService) GetChat(ctx context.Context, chatID int, limit int) (*models.Chat, error) {
	args := m.Called(ctx, chatID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Chat), args.Error(1)
}

func (m *MockChatService) DeleteChat(ctx context.Context, chatID int) error {
	args := m.Called(ctx, chatID)
	return args.Error(0)
}

func TestCreateChatHandler_Success(t *testing.T) {
	// Arrange
	mockService := new(MockChatService)
	handler := handlers.NewChatHandler(mockService)

	expectedChat := &models.Chat{
		ID:    1,
		Title: "Test Chat",
	}

	mockService.On("CreateChat", mock.Anything, mock.Anything).
		Return(expectedChat, nil)

	reqBody := `{"title": "Test Chat"}`
	req := httptest.NewRequest("POST", "/chats", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	// Act
	handler.CreateChat(w, req)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.Chat
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, expectedChat.ID, response.ID)
	assert.Equal(t, expectedChat.Title, response.Title)

	mockService.AssertExpectations(t)
}

func TestCreateChatHandler_InvalidJSON(t *testing.T) {
	// Arrange
	mockService := new(MockChatService)
	handler := handlers.NewChatHandler(mockService)

	req := httptest.NewRequest("POST", "/chats", bytes.NewBufferString(`{invalid json`))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	// Act
	handler.CreateChat(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "BAD_REQUEST", response["error"])
}

func TestCreateChatHandler_ChatAlreadyExists(t *testing.T) {
	// Arrange
	mockService := new(MockChatService)
	handler := handlers.NewChatHandler(mockService)

	mockService.On("CreateChat", mock.Anything, mock.Anything).
		Return(nil, services.ErrChatAlreadyExists)

	reqBody := `{"title": "Existing Chat"}`
	req := httptest.NewRequest("POST", "/chats", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	// Act
	handler.CreateChat(w, req)

	// Assert
	assert.Equal(t, http.StatusConflict, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "CONFLICT", response["error"])
}

func TestGetChatHandler_Success(t *testing.T) {
	// Arrange
	mockService := new(MockChatService)
	handler := handlers.NewChatHandler(mockService)

	expectedChat := &models.Chat{
		ID:    1,
		Title: "Test Chat",
		Messages: []models.Message{
			{ID: 1, ChatID: 1, Text: "Hello"},
		},
	}

	mockService.On("GetChat", mock.Anything, 1, 20).
		Return(expectedChat, nil)

	req := httptest.NewRequest("GET", "/chats/1", nil)
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()

	// Act
	handler.GetChat(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Chat
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, expectedChat.ID, response.ID)
	assert.Len(t, response.Messages, 1)

	mockService.AssertExpectations(t)
}

func TestGetChatHandler_NotFound(t *testing.T) {
	// Arrange
	mockService := new(MockChatService)
	handler := handlers.NewChatHandler(mockService)

	mockService.On("GetChat", mock.Anything, 999, 20).
		Return(nil, services.ErrChatNotFound)

	req := httptest.NewRequest("GET", "/chats/999", nil)
	req.SetPathValue("id", "999")
	w := httptest.NewRecorder()

	// Act
	handler.GetChat(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "NOT_FOUND", response["error"])

	mockService.AssertExpectations(t)
}

func TestDeleteChatHandler_Success(t *testing.T) {
	// Arrange
	mockService := new(MockChatService)
	handler := handlers.NewChatHandler(mockService)

	mockService.On("DeleteChat", mock.Anything, 1).
		Return(nil)

	req := httptest.NewRequest("DELETE", "/chats/1", nil)
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()

	// Act
	handler.DeleteChat(w, req)

	// Assert
	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, w.Body.String())

	mockService.AssertExpectations(t)
}

func TestGetChatHandler_WithLimit(t *testing.T) {
	// Arrange
	mockService := new(MockChatService)
	handler := handlers.NewChatHandler(mockService)

	expectedChat := &models.Chat{ID: 1, Title: "Test"}

	mockService.On("GetChat", mock.Anything, 1, 5).
		Return(expectedChat, nil)

	req := httptest.NewRequest("GET", "/chats/1?limit=5", nil)
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()

	// Act
	handler.GetChat(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestCreateChatHandler_EmptyTitle(t *testing.T) {
	// Arrange
	mockService := new(MockChatService)
	handler := handlers.NewChatHandler(mockService)

	reqBody := `{"title": "   "}`
	req := httptest.NewRequest("POST", "/chats", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	// Act
	handler.CreateChat(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "BAD_REQUEST", response["error"])
}

func TestCreateChatHandler_TitleTooLong(t *testing.T) {
	// Arrange
	mockService := new(MockChatService)
	handler := handlers.NewChatHandler(mockService)

	longTitle := strings.Repeat("a", 201)
	reqBody := fmt.Sprintf(`{"title": "%s"}`, longTitle)

	req := httptest.NewRequest("POST", "/chats", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	// Act
	handler.CreateChat(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "BAD_REQUEST", response["error"])
}

func TestGetChatHandler_InvalidID(t *testing.T) {
	// Arrange
	mockService := new(MockChatService)
	handler := handlers.NewChatHandler(mockService)

	req := httptest.NewRequest("GET", "/chats/abc", nil)
	req.SetPathValue("id", "abc")
	w := httptest.NewRecorder()

	// Act
	handler.GetChat(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "BAD_REQUEST", response["error"])
}

func TestGetChatHandler_LimitOutOfRange(t *testing.T) {
	// Arrange
	mockService := new(MockChatService)
	handler := handlers.NewChatHandler(mockService)

	req := httptest.NewRequest("GET", "/chats/1?limit=150", nil)
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()

	mockService.On("GetChat", mock.Anything, 1, 20).
		Return(&models.Chat{ID: 1}, nil)

	// Act
	handler.GetChat(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

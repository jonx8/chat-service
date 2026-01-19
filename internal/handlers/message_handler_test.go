package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
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

type MockMessageService struct {
	mock.Mock
}

func (m *MockMessageService) CreateMessage(ctx context.Context, chatID int, req *dto.CreateMessageRequest) (*models.Message, error) {
	args := m.Called(ctx, chatID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Message), args.Error(1)
}

func TestCreateMessageHandler_Success(t *testing.T) {
	// Arrange
	mockService := new(MockMessageService)
	handler := handlers.NewMessageHandler(mockService)

	expectedMessage := &models.Message{
		ID:     1,
		ChatID: 123,
		Text:   "Hello, world!",
	}

	mockService.On("CreateMessage", mock.Anything, 123, mock.MatchedBy(func(req *dto.CreateMessageRequest) bool {
		return req.Text == "Hello, world!"
	})).Return(expectedMessage, nil)

	reqBody := `{"text": "Hello, world!"}`
	req := httptest.NewRequest("POST", "/chats/123/messages", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "123")

	w := httptest.NewRecorder()

	// Act
	handler.CreateMessage(w, req)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.Message
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, expectedMessage.ID, response.ID)
	assert.Equal(t, expectedMessage.ChatID, response.ChatID)
	assert.Equal(t, expectedMessage.Text, response.Text)

	mockService.AssertExpectations(t)
}

func TestCreateMessageHandler_InvalidChatID(t *testing.T) {
	// Arrange
	mockService := new(MockMessageService)
	handler := handlers.NewMessageHandler(mockService)

	req := httptest.NewRequest("POST", "/chats/abc/messages", nil)
	req.SetPathValue("id", "abc")

	w := httptest.NewRecorder()

	// Act
	handler.CreateMessage(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "BAD_REQUEST", response["error"])
	assert.Contains(t, response["message"], "ID path param must be integer")
}

func TestCreateMessageHandler_InvalidJSON(t *testing.T) {
	// Arrange
	mockService := new(MockMessageService)
	handler := handlers.NewMessageHandler(mockService)

	reqBody := `{"text": "Hello"`
	req := httptest.NewRequest("POST", "/chats/123/messages", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "123")

	w := httptest.NewRecorder()

	// Act
	handler.CreateMessage(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "BAD_REQUEST", response["error"])
}

func TestCreateMessageHandler_EmptyText(t *testing.T) {
	// Arrange
	mockService := new(MockMessageService)
	handler := handlers.NewMessageHandler(mockService)

	reqBody := `{"text": ""}`
	req := httptest.NewRequest("POST", "/chats/123/messages", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "123")

	w := httptest.NewRecorder()

	// Act
	handler.CreateMessage(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "BAD_REQUEST", response["error"])
	assert.Contains(t, response["message"], "Message length must be between")
}

func TestCreateMessageHandler_TextTooLong(t *testing.T) {
	// Arrange
	mockService := new(MockMessageService)
	handler := handlers.NewMessageHandler(mockService)

	longText := strings.Repeat("a", 5001)

	reqBody := `{"text": "` + longText + `"}`
	req := httptest.NewRequest("POST", "/chats/123/messages", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "123")

	w := httptest.NewRecorder()

	// Act
	handler.CreateMessage(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "BAD_REQUEST", response["error"])
}

func TestCreateMessageHandler_ChatNotFound(t *testing.T) {
	// Arrange
	mockService := new(MockMessageService)
	handler := handlers.NewMessageHandler(mockService)

	mockService.On("CreateMessage", mock.Anything, 999, mock.Anything).
		Return(nil, services.ErrChatNotFound)

	reqBody := `{"text": "Hello"}`
	req := httptest.NewRequest("POST", "/chats/999/messages", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "999")

	w := httptest.NewRecorder()

	// Act
	handler.CreateMessage(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "NOT_FOUND", response["error"])
	assert.Equal(t, "Chat not found", response["message"])

	mockService.AssertExpectations(t)
}

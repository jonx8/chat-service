package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jonx8/chat-service/internal/dto"
	"github.com/jonx8/chat-service/internal/models"
	repo "github.com/jonx8/chat-service/internal/repositories"
)

var (
	ErrChatNotFound      = errors.New("chat not found")
	ErrChatAlreadyExists = errors.New("chat already exists")
)

type ChatService interface {
	CreateChat(ctx context.Context, request *dto.CreateChatRequest) (*models.Chat, error)
	GetChat(ctx context.Context, id int, limit int) (*models.Chat, error)
	DeleteChat(ctx context.Context, id int) error
}

type chatService struct {
	chatRepository repo.ChatRepository
}

func NewChatService(chatRepository repo.ChatRepository) ChatService {
	return &chatService{chatRepository: chatRepository}
}

func (service *chatService) CreateChat(ctx context.Context, req *dto.CreateChatRequest) (*models.Chat, error) {
	chat := &models.Chat{
		Title: req.Title,
	}
	if err := service.chatRepository.CreateIfNotExists(ctx, chat); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return nil, ErrChatAlreadyExists
		}
		return nil, fmt.Errorf("create chat: %w", err)
	}

	return chat, nil
}

func (service *chatService) GetChat(ctx context.Context, id int, limit int) (*models.Chat, error) {
	chat, err := service.chatRepository.GetByID(ctx, id, limit)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, ErrChatNotFound
		}
		return nil, fmt.Errorf("get chat: %w", err)
	}
	return chat, nil
}

func (service *chatService) DeleteChat(ctx context.Context, id int) error {
	if err := service.chatRepository.DeleteByID(ctx, id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return ErrChatNotFound
		}
		return fmt.Errorf("delete chat: %w", err)
	}
	return nil
}

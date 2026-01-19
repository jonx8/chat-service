package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/jonx8/chat-service/internal/dto"
	"github.com/jonx8/chat-service/internal/models"
	repo "github.com/jonx8/chat-service/internal/repositories"
)

type MessageService interface {
	CreateMessage(ctx context.Context, chatID int, req *dto.CreateMessageRequest) (*models.Message, error)
}

type messageService struct {
	messageRepository repo.MessageRepository
}

func NewMessageService(messageRepository repo.MessageRepository) MessageService {
	return &messageService{messageRepository: messageRepository}
}

func (service *messageService) CreateMessage(ctx context.Context, chatID int, req *dto.CreateMessageRequest) (*models.Message, error) {
	message := &models.Message{
		ChatID: chatID,
		Text:   req.Text,
	}
	if err := service.messageRepository.CreateMessage(ctx, message); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, ErrChatNotFound
		}
		return nil, fmt.Errorf("create message: %w", err)
	}
	return message, nil
}

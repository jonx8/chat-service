package repositories

import (
	"context"
	"fmt"

	"github.com/jonx8/chat-service/internal/models"
	"gorm.io/gorm"
)

type MessageRepository interface {
	CreateMessage(ctx context.Context, message *models.Message) error
}

type messageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) MessageRepository {
	return &messageRepository{db: db}
}

func (repo *messageRepository) CreateMessage(ctx context.Context, message *models.Message) error {
	return repo.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		err := tx.Model(&models.Chat{}).Where("id = ?", message.ChatID).Count(&count).Error

		if err != nil {
			return fmt.Errorf("check chat existence: %w", err)
		}

		if count == 0 {
			return fmt.Errorf("chat with id %d not found", message.ChatID)
		}

		return tx.Create(message).Error
	})
}

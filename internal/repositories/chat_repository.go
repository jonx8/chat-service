package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/jonx8/chat-service/internal/models"
	"gorm.io/gorm"
)

type ChatRepository interface {
	GetByID(ctx context.Context, id int, limit int) (*models.Chat, error)
	CreateIfNotExists(ctx context.Context, chat *models.Chat) error
	DeleteByID(ctx context.Context, id int) error
}

type chatRepository struct {
	db *gorm.DB
}

func NewChatRepository(db *gorm.DB) ChatRepository {
	return &chatRepository{db: db}
}

func (repo *chatRepository) CreateIfNotExists(ctx context.Context, chat *models.Chat) error {
	return repo.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64

		err := tx.
			Model(&models.Chat{}).
			Where("title = ?", chat.Title).
			Count(&count).Error

		if err != nil {
			return fmt.Errorf("check title existence: %w", err)
		}

		if count > 0 {
			return fmt.Errorf("chat with title '%s' already exists", chat.Title)
		}

		err = tx.Create(chat).Error
		if chat.Messages == nil {
			chat.Messages = []models.Message{}
		}

		return err

	})
}

func (repo *chatRepository) GetByID(ctx context.Context, id int, limit int) (*models.Chat, error) {

	tx := repo.db.WithContext(ctx).Model(&models.Chat{})

	if limit > 0 {
		tx = tx.Preload("Messages", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC").Limit(limit)
		})
	}

	var chat models.Chat
	if chat.Messages == nil {
		chat.Messages = []models.Message{}
	}

	if err := tx.First(&chat, id).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("chat with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to get chat: %w", err)
	}

	return &chat, nil
}

func (repo *chatRepository) DeleteByID(ctx context.Context, id int) error {
	result := repo.db.WithContext(ctx).Delete(&models.Chat{}, id)
	if err := result.Error; err != nil {
		return fmt.Errorf("failed to delete chat: %w", err)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("chat with id %d not found", id)
	}

	return nil
}

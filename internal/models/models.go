package models

import "time"

type Chat struct {
	ID        int       `gorm:"primaryKey" json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	Messages  []Message `json:"messages"`
}

type Message struct {
	ID        int       `gorm:"primaryKey" json:"id"`
	ChatID    int       `json:"chat_id"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`

	Chat *Chat `json:"-"`
}

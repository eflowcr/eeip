package models

import "time"

type Stakeholder struct {
	ID             string    `json:"id" db:"id"`
	Name           string    `json:"name" db:"name"`
	Email          string    `json:"email" db:"email"`
	TelegramChatID string    `json:"telegram_chat_id" db:"telegram_chat_id"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

package model

import (
	"database/sql"
	"time"
)

// AIBot represents an AI bot
type AIBot struct {
	ID          int64          `json:"id" db:"id"`
	BotID       int64          `json:"bot_id" db:"bot_id"`
	UserID      int64          `json:"user_id" db:"user_id"`
	Name        string         `json:"name" db:"name"`
	Personality string         `json:"personality" db:"personality"` // JSON string
	ModelName   string         `json:"model_name" db:"model_name"`
	Temperature float64        `json:"temperature" db:"temperature"`
	MaxTokens   int            `json:"max_tokens" db:"max_tokens"`
	CreatedAt   time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt   sql.NullTime   `json:"updated_at" db:"updated_at"`
}

// AIConversation represents a conversation with an AI bot
type AIConversation struct {
	ID        int64     `json:"id" db:"id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	BotID     int64     `json:"bot_id" db:"bot_id"`
	Title     string    `json:"title" db:"title"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// TableName returns the table name for AIBot
func (AIBot) TableName() string {
	return "ai_bots"
}

// TableName returns the table name for AIConversation
func (AIConversation) TableName() string {
	return "ai_conversations"
}

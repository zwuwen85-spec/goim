package model

import (
	"database/sql"
	"time"
)

// User represents a user account
type User struct {
	ID           int64          `json:"id" db:"id"`
	Username     string         `json:"username" db:"username"`
	PasswordHash string         `json:"-" db:"password_hash"`
	Nickname     string         `json:"nickname" db:"nickname"`
	AvatarURL    sql.NullString `json:"avatar_url" db:"avatar_url"`
	Status       int8           `json:"status" db:"status"` // 1=online, 2=offline, 3=busy
	Signature    sql.NullString `json:"signature" db:"signature"`
	CreatedAt    time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at" db:"updated_at"`
}

// UserToken represents an authentication token
type UserToken struct {
	ID        int64     `json:"id" db:"id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	Token     string    `json:"token" db:"token"`
	DeviceID  string    `json:"device_id" db:"device_id"`
	Platform  string    `json:"platform" db:"platform"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// TableName returns the table name for User
func (User) TableName() string {
	return "users"
}

// TableName returns the table name for UserToken
func (UserToken) TableName() string {
	return "user_tokens"
}

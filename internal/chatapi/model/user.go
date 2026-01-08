package model

import (
	"database/sql"
	"encoding/json"
	"time"
)

// User represents a user account
type User struct {
	ID           int64          `json:"id" db:"id"`
	Username     string         `json:"username" db:"username"`
	PasswordHash string         `json:"-" db:"password_hash"`
	Nickname     string         `json:"nickname" db:"nickname"`
	AvatarURL    sql.NullString `json:"-" db:"avatar_url"` // Custom marshaling
	Status       int8           `json:"status" db:"status"` // 1=online, 2=offline, 3=busy
	Signature    sql.NullString `json:"signature" db:"signature"`
	CreatedAt    time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at" db:"updated_at"`
}

// MarshalJSON custom JSON marshaling for User
func (u User) MarshalJSON() ([]byte, error) {
	type Alias User
	avatar := ""
	if u.AvatarURL.Valid {
		avatar = u.AvatarURL.String
	}
	signature := ""
	if u.Signature.Valid {
		signature = u.Signature.String
	}
	return json.Marshal(&struct {
		Avatar   string `json:"avatar,omitempty"`
		Signature string `json:"signature,omitempty"`
		*Alias
	}{
		Avatar:   avatar,
		Signature: signature,
		Alias:    (*Alias)(&u),
	})
}

// UnmarshalJSON custom JSON unmarshaling for User
func (u *User) UnmarshalJSON(data []byte) error {
	type Alias User
	aux := &struct {
		Avatar *string `json:"avatar"`
		*Alias
	}{
		Alias: (*Alias)(u),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if aux.Avatar != nil {
		u.AvatarURL = sql.NullString{String: *aux.Avatar, Valid: true}
	} else {
		u.AvatarURL = sql.NullString{Valid: false}
	}
	return nil
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

package model

import (
	"database/sql"
	"time"
)

// Friendship represents a friend relationship
type Friendship struct {
	ID        int64          `json:"id" db:"id"`
	UserID    int64          `json:"user_id" db:"user_id"`
	FriendID  int64          `json:"friend_id" db:"friend_id"`
	Remark    sql.NullString `json:"remark" db:"remark"`
	GroupName string         `json:"group_name" db:"group_name"`
	Status    int8           `json:"status" db:"status"` // 1=normal, 2=deleted, 3=blocked
	CreatedAt time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt time.Time      `json:"updated_at" db:"updated_at"`

	// Join fields (populated by queries)
	FriendUser *User `json:"friend_user,omitempty"`
}

// FriendRequest represents a friend request
type FriendRequest struct {
	ID         int64          `json:"id" db:"id"`
	FromUserID int64          `json:"from_user_id" db:"from_user_id"`
	ToUserID   int64          `json:"to_user_id" db:"to_user_id"`
	Message    sql.NullString `json:"message" db:"message"`
	Status     int8           `json:"status" db:"status"` // 1=pending, 2=accepted, 3=rejected
	CreatedAt  time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at" db:"updated_at"`

	// Join fields
	FromUser *User `json:"from_user,omitempty"`
	ToUser   *User `json:"to_user,omitempty"`
}

// TableName returns the table name for Friendship
func (Friendship) TableName() string {
	return "friendships"
}

// TableName returns the table name for FriendRequest
func (FriendRequest) TableName() string {
	return "friend_requests"
}

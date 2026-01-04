package model

import (
	"database/sql"
	"time"
)

// Message represents a chat message
type Message struct {
	ID               int64          `json:"id" db:"id"`
	MsgID            string         `json:"msg_id" db:"msg_id"`
	FromUserID       int64          `json:"from_user_id" db:"from_user_id"`
	ConversationID   int64          `json:"conversation_id" db:"conversation_id"`
	ConversationType int8           `json:"conversation_type" db:"conversation_type"` // 1=single, 2=group, 3=AI
	MsgType          int8           `json:"msg_type" db:"msg_type"`                   // 1=text, 2=image, 3=voice, 4=video, 5=file, 6=system
	Content          string         `json:"content" db:"content"`                     // JSON string
	Seq              int64          `json:"seq" db:"seq"`
	IsDeleted        int8           `json:"is_deleted" db:"is_deleted"`
	CreatedAt        time.Time      `json:"created_at" db:"created_at"`

	// Join fields
	FromUser *User `json:"from_user,omitempty"`
}

// MessageReadStatus represents message read status
type MessageReadStatus struct {
	ID     int64     `json:"id" db:"id"`
	MsgID  int64     `json:"msg_id" db:"msg_id"`
	UserID int64     `json:"user_id" db:"user_id"`
	ReadAt time.Time `json:"read_at" db:"read_at"`
}

// Conversation represents a user's conversation
type Conversation struct {
	ID              int64          `json:"id" db:"id"`
	UserID          int64          `json:"user_id" db:"user_id"`
	TargetID        int64          `json:"target_id" db:"target_id"`
	ConversationType int8          `json:"conversation_type" db:"conversation_type"` // 1=single, 2=group, 3=AI
	UnreadCount     int            `json:"unread_count" db:"unread_count"`
	LastMsgID       sql.NullInt64  `json:"last_msg_id" db:"last_msg_id"`
	LastMsgContent  sql.NullString `json:"last_msg_content" db:"last_msg_content"`
	LastMsgTime     sql.NullTime   `json:"last_msg_time" db:"last_msg_time"`
	IsPinned        int8           `json:"is_pinned" db:"is_pinned"`
	IsMuted         int8           `json:"is_muted" db:"is_muted"`
	UpdatedAt       time.Time      `json:"updated_at" db:"updated_at"`

	// Join fields
	TargetUser *User  `json:"target_user,omitempty"` // For single chat
	TargetGroup *Group `json:"target_group,omitempty"` // For group chat
	LastMessage *Message `json:"last_message,omitempty"`
}

// TableName returns the table name for Message
func (Message) TableName() string {
	return "messages"
}

// TableName returns the table name for MessageReadStatus
func (MessageReadStatus) TableName() string {
	return "message_read_status"
}

// TableName returns the table name for Conversation
func (Conversation) TableName() string {
	return "conversations"
}

// Message type constants
const (
	MsgTypeText   = 1
	MsgTypeImage  = 2
	MsgTypeVoice  = 3
	MsgTypeVideo  = 4
	MsgTypeFile   = 5
	MsgTypeSystem = 6
)

// Conversation type constants
const (
	ConversationTypeSingle = 1
	ConversationTypeGroup  = 2
	ConversationTypeAI     = 3
)

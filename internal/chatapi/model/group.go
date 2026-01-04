package model

import (
	"database/sql"
	"time"
)

// Group represents a chat group
type Group struct {
	ID         int64          `json:"id" db:"id"`
	GroupNo    string         `json:"group_no" db:"group_no"`
	Name       string         `json:"name" db:"name"`
	AvatarURL  sql.NullString `json:"avatar_url" db:"avatar_url"`
	OwnerID    int64          `json:"owner_id" db:"owner_id"`
	MaxMembers int            `json:"max_members" db:"max_members"`
	JoinType   int8           `json:"join_type" db:"join_type"` // 1=open, 2=need approval, 3=closed
	MuteAll    int8           `json:"mute_all" db:"mute_all"`   // 0=no, 1=yes
	CreatedAt  time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at" db:"updated_at"`
}

// GroupMember represents a group member
type GroupMember struct {
	ID        int64          `json:"id" db:"id"`
	GroupID   int64          `json:"group_id" db:"group_id"`
	UserID    int64          `json:"user_id" db:"user_id"`
	Role      int8           `json:"role" db:"role"` // 1=member, 2=admin, 3=owner
	Nickname  sql.NullString `json:"nickname" db:"nickname"`
	MuteUntil sql.NullTime   `json:"mute_until" db:"mute_until"`
	JoinedAt  time.Time      `json:"joined_at" db:"joined_at"`

	// Join fields
	User  *User  `json:"user,omitempty"`
	Group *Group `json:"group,omitempty"`
}

// GroupJoinRequest represents a group join request
type GroupJoinRequest struct {
	ID        int64          `json:"id" db:"id"`
	GroupID   int64          `json:"group_id" db:"group_id"`
	UserID    int64          `json:"user_id" db:"user_id"`
	Message   sql.NullString `json:"message" db:"message"`
	Status    int8           `json:"status" db:"status"` // 1=pending, 2=accepted, 3=rejected
	CreatedAt time.Time      `json:"created_at" db:"created_at"`

	// Join fields
	Group *Group `json:"group,omitempty"`
	User  *User  `json:"user,omitempty"`
}

// TableName returns the table name for Group
func (Group) TableName() string {
	return "groups"
}

// TableName returns the table name for GroupMember
func (GroupMember) TableName() string {
	return "group_members"
}

// TableName returns the table name for GroupJoinRequest
func (GroupJoinRequest) TableName() string {
	return "group_join_requests"
}

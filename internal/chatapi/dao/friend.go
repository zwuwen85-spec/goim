package dao

import (
	"context"
	"database/sql"

	"github.com/Terry-Mao/goim/internal/chatapi/model"
)

// FriendDAO handles friend-related database operations
type FriendDAO struct {
	mysql *MySQL
}

// NewFriendDAO creates a new FriendDAO
func NewFriendDAO(mysql *MySQL) *FriendDAO {
	return &FriendDAO{mysql: mysql}
}

// CreateFriendRequest creates a new friend request
func (d *FriendDAO) CreateFriendRequest(ctx context.Context, req *model.FriendRequest) error {
	query := `
		INSERT INTO friend_requests (from_user_id, to_user_id, message, status)
		VALUES (?, ?, ?, 1)
	`
	_, err := d.mysql.Exec(ctx, query, req.FromUserID, req.ToUserID, req.Message)
	return err
}

// FindFriendRequest finds a friend request by ID
func (d *FriendDAO) FindFriendRequest(ctx context.Context, id int64) (*model.FriendRequest, error) {
	query := `
		SELECT id, from_user_id, to_user_id, message, status, created_at, updated_at
		FROM friend_requests
		WHERE id = ?
	`
	return d.scanFriendRequest(d.mysql.QueryRow(ctx, query, id))
}

// FindPendingFriendRequest finds a pending friend request between two users
func (d *FriendDAO) FindPendingFriendRequest(ctx context.Context, fromUserID, toUserID int64) (*model.FriendRequest, error) {
	query := `
		SELECT id, from_user_id, to_user_id, message, status, created_at, updated_at
		FROM friend_requests
		WHERE from_user_id = ? AND to_user_id = ? AND status = 1
	`
	return d.scanFriendRequest(d.mysql.QueryRow(ctx, query, fromUserID, toUserID))
}

// UpdateFriendRequestStatus updates the status of a friend request
func (d *FriendDAO) UpdateFriendRequestStatus(ctx context.Context, id int64, status int8) error {
	query := `UPDATE friend_requests SET status = ? WHERE id = ?`
	_, err := d.mysql.Exec(ctx, query, status, id)
	return err
}

// GetFriendRequestsByUser gets all friend requests for a user (both sent and received)
func (d *FriendDAO) GetFriendRequestsByUser(ctx context.Context, userID int64) ([]*model.FriendRequest, error) {
	query := `
		SELECT id, from_user_id, to_user_id, message, status, created_at, updated_at
		FROM friend_requests
		WHERE to_user_id = ? AND status = 1
		ORDER BY created_at DESC
	`
	rows, err := d.mysql.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []*model.FriendRequest
	for rows.Next() {
		req, err := d.scanFriendRequest(rows)
		if err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}
	return requests, nil
}

// CreateFriendship creates a new friendship
func (d *FriendDAO) CreateFriendship(ctx context.Context, fs *model.Friendship) error {
	query := `
		INSERT INTO friendships (user_id, friend_id, remark, group_name, status)
		VALUES (?, ?, ?, ?, 1)
	`
	_, err := d.mysql.Exec(ctx, query, fs.UserID, fs.FriendID, fs.Remark, fs.GroupName)
	return err
}

// FindFriendship finds a friendship between two users
func (d *FriendDAO) FindFriendship(ctx context.Context, userID, friendID int64) (*model.Friendship, error) {
	query := `
		SELECT id, user_id, friend_id, remark, group_name, status, created_at, updated_at
		FROM friendships
		WHERE user_id = ? AND friend_id = ?
	`
	return d.scanFriendship(d.mysql.QueryRow(ctx, query, userID, friendID))
}

// GetFriends gets all friends for a user
func (d *FriendDAO) GetFriends(ctx context.Context, userID int64) ([]*model.Friendship, error) {
	query := `
		SELECT f.id, f.user_id, f.friend_id, f.remark, f.group_name, f.status, f.created_at, f.updated_at,
		       u.id, u.username, u.password_hash, u.nickname, u.avatar_url, u.status, u.signature, u.created_at, u.updated_at
		FROM friendships f
		INNER JOIN users u ON f.friend_id = u.id
		WHERE f.user_id = ? AND f.status = 1
		ORDER BY f.updated_at DESC
	`
	rows, err := d.mysql.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var friendships []*model.Friendship
	for rows.Next() {
		fs, err := d.scanFriendshipWithUser(rows)
		if err != nil {
			return nil, err
		}
		friendships = append(friendships, fs)
	}
	return friendships, nil
}

// DeleteFriendship deletes a friendship
func (d *FriendDAO) DeleteFriendship(ctx context.Context, userID, friendID int64) error {
	query := `DELETE FROM friendships WHERE user_id = ? AND friend_id = ?`
	_, err := d.mysql.Exec(ctx, query, userID, friendID)
	return err
}

// UpdateFriendship updates friendship remark or group
func (d *FriendDAO) UpdateFriendship(ctx context.Context, fs *model.Friendship) error {
	query := `UPDATE friendships SET remark = ?, group_name = ? WHERE user_id = ? AND friend_id = ?`
	_, err := d.mysql.Exec(ctx, query, fs.Remark, fs.GroupName, fs.UserID, fs.FriendID)
	return err
}

// scanFriendship scans a friendship from a row
func (d *FriendDAO) scanFriendship(scanner interface{ Scan(...interface{}) error }) (*model.Friendship, error) {
	var fs model.Friendship
	err := scanner.Scan(
		&fs.ID, &fs.UserID, &fs.FriendID, &fs.Remark, &fs.GroupName,
		&fs.Status, &fs.CreatedAt, &fs.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &fs, nil
}

// scanFriendshipWithUser scans a friendship with user info from a row
func (d *FriendDAO) scanFriendshipWithUser(rows *sql.Rows) (*model.Friendship, error) {
	var fs model.Friendship
	var user model.User

	err := rows.Scan(
		&fs.ID, &fs.UserID, &fs.FriendID, &fs.Remark, &fs.GroupName,
		&fs.Status, &fs.CreatedAt, &fs.UpdatedAt,
		&user.ID, &user.Username, &user.PasswordHash, &user.Nickname,
		&user.AvatarURL, &user.Status, &user.Signature,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	user.PasswordHash = "" // Don't return password hash
	fs.FriendUser = &user
	return &fs, nil
}

// scanFriendRequest scans a friend request from a row
func (d *FriendDAO) scanFriendRequest(scanner interface{ Scan(...interface{}) error }) (*model.FriendRequest, error) {
	var req model.FriendRequest
	err := scanner.Scan(
		&req.ID, &req.FromUserID, &req.ToUserID, &req.Message,
		&req.Status, &req.CreatedAt, &req.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &req, nil
}

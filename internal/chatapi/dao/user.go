package dao

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Terry-Mao/goim/internal/chatapi/model"
	"golang.org/x/crypto/bcrypt"
)

// UserDAO handles user-related database operations
type UserDAO struct {
	mysql *MySQL
}

// NewUserDAO creates a new UserDAO
func NewUserDAO(mysql *MySQL) *UserDAO {
	return &UserDAO{mysql: mysql}
}

// Create creates a new user
func (d *UserDAO) Create(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (username, password_hash, nickname, avatar_url, status, signature)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err := d.mysql.Exec(ctx, query,
		user.Username, user.PasswordHash, user.Nickname,
		user.AvatarURL, user.Status, user.Signature,
	)
	return err
}

// FindByID finds a user by ID
func (d *UserDAO) FindByID(ctx context.Context, id int64) (*model.User, error) {
	query := `
		SELECT id, username, password_hash, nickname, avatar_url, status, signature, created_at, updated_at
		FROM users
		WHERE id = ?
	`
	return d.scanOne(d.mysql.QueryRow(ctx, query, id))
}

// FindByUsername finds a user by username
func (d *UserDAO) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	query := `
		SELECT id, username, password_hash, nickname, avatar_url, status, signature, created_at, updated_at
		FROM users
		WHERE username = ?
	`
	return d.scanOne(d.mysql.QueryRow(ctx, query, username))
}

// Search searches users by nickname or username
func (d *UserDAO) Search(ctx context.Context, keyword string, limit int) ([]*model.User, error) {
	query := `
		SELECT id, username, password_hash, nickname, avatar_url, status, signature, created_at, updated_at
		FROM users
		WHERE username LIKE ? OR nickname LIKE ?
		LIMIT ?
	`
	pattern := "%" + keyword + "%"
	rows, err := d.mysql.Query(ctx, query, pattern, pattern, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		user, err := d.scanOne(rows)
		if err != nil {
			return nil, err
		}
		user.PasswordHash = "" // Don't return password hash
		users = append(users, user)
	}
	return users, nil
}

// Update updates a user's profile
func (d *UserDAO) Update(ctx context.Context, user *model.User) error {
	query := `
		UPDATE users
		SET nickname = ?, avatar_url = ?, signature = ?, status = ?
		WHERE id = ?
	`
	_, err := d.mysql.Exec(ctx, query,
		user.Nickname, user.AvatarURL, user.Signature, user.Status, user.ID,
	)
	return err
}

// UpdatePassword updates user's password
func (d *UserDAO) UpdatePassword(ctx context.Context, userID int64, passwordHash string) error {
	query := `UPDATE users SET password_hash = ? WHERE id = ?`
	_, err := d.mysql.Exec(ctx, query, passwordHash, userID)
	return err
}

// UpdateAvatar updates user's avatar URL
func (d *UserDAO) UpdateAvatar(ctx context.Context, userID int64, avatarURL string) error {
	query := `UPDATE users SET avatar_url = ? WHERE id = ?`
	_, err := d.mysql.Exec(ctx, query, avatarURL, userID)
	return err
}

// scanOne scans a user from a single row
func (d *UserDAO) scanOne(scanner interface{ Scan(...interface{}) error }) (*model.User, error) {
	var user model.User
	err := scanner.Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.Nickname,
		&user.AvatarURL, &user.Status, &user.Signature,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// TokenDAO handles token-related database operations
type TokenDAO struct {
	mysql *MySQL
}

// NewTokenDAO creates a new TokenDAO
func NewTokenDAO(mysql *MySQL) *TokenDAO {
	return &TokenDAO{mysql: mysql}
}

// Create creates a new token
func (d *TokenDAO) Create(ctx context.Context, token *model.UserToken) error {
	query := `
		INSERT INTO user_tokens (user_id, token, device_id, platform, expires_at)
		VALUES (?, ?, ?, ?, ?)
	`
	_, err := d.mysql.Exec(ctx, query,
		token.UserID, token.Token, token.DeviceID, token.Platform, token.ExpiresAt,
	)
	return err
}

// FindByToken finds a token by token string
func (d *TokenDAO) FindByToken(ctx context.Context, tokenStr string) (*model.UserToken, error) {
	query := `
		SELECT id, user_id, token, device_id, platform, expires_at, created_at
		FROM user_tokens
		WHERE token = ? AND expires_at > NOW()
	`
	var token model.UserToken
	err := d.mysql.QueryRow(ctx, query, tokenStr).Scan(
		&token.ID, &token.UserID, &token.Token, &token.DeviceID,
		&token.Platform, &token.ExpiresAt, &token.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("token not found or expired")
	}
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// DeleteByUserID deletes all tokens for a user
func (d *TokenDAO) DeleteByUserID(ctx context.Context, userID int64) error {
	query := `DELETE FROM user_tokens WHERE user_id = ?`
	_, err := d.mysql.Exec(ctx, query, userID)
	return err
}

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyPassword verifies a password against a hash
func VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

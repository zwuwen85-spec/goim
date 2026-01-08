package dao

import (
	"context"
	"database/sql"

	"github.com/Terry-Mao/goim/internal/chatapi/model"
	"github.com/google/uuid"
)

// MessageDAO handles message-related database operations
type MessageDAO struct {
	mysql *MySQL
}

// NewMessageDAO creates a new MessageDAO
func NewMessageDAO(mysql *MySQL) *MessageDAO {
	return &MessageDAO{mysql: mysql}
}

// CreateMessage creates a new message
func (d *MessageDAO) CreateMessage(ctx context.Context, msg *model.Message) error {
	// Generate msg_id if not provided
	if msg.MsgID == "" {
		msg.MsgID = generateMsgID()
	}
	query := `
		INSERT INTO messages (msg_id, from_user_id, conversation_id, conversation_type, msg_type, content, seq)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	_, err := d.mysql.Exec(ctx, query,
		msg.MsgID, msg.FromUserID, msg.ConversationID, msg.ConversationType,
		msg.MsgType, msg.Content, msg.Seq,
	)
	return err
}

// FindByMsgID finds a message by msg_id
func (d *MessageDAO) FindByMsgID(ctx context.Context, msgID string) (*model.Message, error) {
	query := `
		SELECT id, msg_id, from_user_id, conversation_id, conversation_type, msg_type, content, seq, is_deleted, created_at
		FROM messages
		WHERE msg_id = ?
	`
	return d.scanMessage(d.mysql.QueryRow(ctx, query, msgID))
}

// GetMessages gets messages for a conversation with pagination
func (d *MessageDAO) GetMessages(ctx context.Context, conversationID int64, conversationType int8, limit int, lastSeq int64) ([]*model.Message, error) {
	query := `
		SELECT id, msg_id, from_user_id, conversation_id, conversation_type, msg_type, content, seq, is_deleted, created_at
		FROM messages
		WHERE conversation_id = ? AND conversation_type = ? AND is_deleted = 0
	`
	args := []interface{}{conversationID, conversationType}

	if lastSeq > 0 {
		query += " AND seq < ?"
		args = append(args, lastSeq)
	}

	query += " ORDER BY seq DESC LIMIT ?"
	args = append(args, limit)

	rows, err := d.mysql.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*model.Message
	for rows.Next() {
		msg, err := d.scanMessage(rows)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	// Reverse to get chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// GetNextSeq gets the next sequence number for a conversation
func (d *MessageDAO) GetNextSeq(ctx context.Context, conversationID int64, conversationType int8) (int64, error) {
	query := `
		SELECT COALESCE(MAX(seq), 0) + 1
		FROM messages
		WHERE conversation_id = ? AND conversation_type = ?
	`
	var seq int64
	err := d.mysql.QueryRow(ctx, query, conversationID, conversationType).Scan(&seq)
	if err != nil && err != sql.ErrNoRows {
		return 1, err
	}
	if seq == 0 {
		return 1, nil
	}
	return seq, nil
}

// MarkMessagesRead marks messages as read for a user
func (d *MessageDAO) MarkMessagesRead(ctx context.Context, msgID int64, userID int64) error {
	query := `
		INSERT INTO message_read_status (msg_id, user_id, read_at)
		VALUES (?, ?, NOW())
		ON DUPLICATE KEY UPDATE read_at = NOW()
	`
	_, err := d.mysql.Exec(ctx, query, msgID, userID)
	return err
}

// scanMessage scans a message from a database row
func (d *MessageDAO) scanMessage(scanner interface{ Scan(...interface{}) error }) (*model.Message, error) {
	var msg model.Message
	err := scanner.Scan(
		&msg.ID, &msg.MsgID, &msg.FromUserID, &msg.ConversationID,
		&msg.ConversationType, &msg.MsgType, &msg.Content, &msg.Seq,
		&msg.IsDeleted, &msg.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

// ConversationDAO handles conversation-related database operations
type ConversationDAO struct {
	mysql *MySQL
}

// NewConversationDAO creates a new ConversationDAO
func NewConversationDAO(mysql *MySQL) *ConversationDAO {
	return &ConversationDAO{mysql: mysql}
}

// GetConversations gets all conversations for a user
func (d *ConversationDAO) GetConversations(ctx context.Context, userID int64) ([]*model.Conversation, error) {
	query := `
		SELECT c.id, c.user_id, c.target_id, c.conversation_type, c.unread_count,
		       c.last_msg_id, c.last_msg_content, c.last_msg_time, c.is_pinned, c.is_muted, c.updated_at,
		       u.id, u.username, u.password_hash, u.nickname, u.avatar_url, u.status, u.signature, u.created_at, u.updated_at,
		       g.id, g.group_no, g.name, g.avatar_url, g.owner_id, g.max_members, g.join_type, g.mute_all, g.created_at, g.updated_at
		FROM conversations c
		LEFT JOIN users u ON (c.conversation_type = 1 AND c.target_id = u.id)
		LEFT JOIN ` + "`groups`" + ` g ON (c.conversation_type = 2 AND c.target_id = g.id)
		WHERE c.user_id = ?
		ORDER BY c.is_pinned DESC, c.updated_at DESC
	`
	rows, err := d.mysql.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conversations []*model.Conversation
	for rows.Next() {
		conv, err := d.scanConversationWithDetails(rows)
		if err != nil {
			return nil, err
		}
		conversations = append(conversations, conv)
	}

	return conversations, nil
}

// GetConversation gets a specific conversation
func (d *ConversationDAO) GetConversation(ctx context.Context, userID, targetID int64, convType int8) (*model.Conversation, error) {
	query := `
		SELECT id, user_id, target_id, conversation_type, unread_count,
		       last_msg_id, last_msg_content, last_msg_time, is_pinned, is_muted, updated_at
		FROM conversations
		WHERE user_id = ? AND target_id = ? AND conversation_type = ?
	`
	return d.scanConversation(d.mysql.QueryRow(ctx, query, userID, targetID, convType))
}

// UpsertConversation creates or updates a conversation
func (d *ConversationDAO) UpsertConversation(ctx context.Context, conv *model.Conversation) error {
	query := `
		INSERT INTO conversations (user_id, target_id, conversation_type, unread_count, last_msg_id, last_msg_content, last_msg_time, is_pinned, is_muted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			unread_count = VALUES(unread_count),
			last_msg_id = VALUES(last_msg_id),
			last_msg_content = VALUES(last_msg_content),
			last_msg_time = VALUES(last_msg_time),
			updated_at = NOW()
	`
	_, err := d.mysql.Exec(ctx, query,
		conv.UserID, conv.TargetID, conv.ConversationType,
		conv.UnreadCount, conv.LastMsgID, conv.LastMsgContent,
		conv.LastMsgTime, conv.IsPinned, conv.IsMuted,
	)
	return err
}

// IncrementUnread increments unread count for a conversation
func (d *ConversationDAO) IncrementUnread(ctx context.Context, userID, targetID int64, convType int8) error {
	query := `
		INSERT INTO conversations (user_id, target_id, conversation_type, unread_count, last_msg_time)
		VALUES (?, ?, ?, 1, NOW())
		ON DUPLICATE KEY UPDATE
			unread_count = unread_count + 1,
			updated_at = NOW()
	`
	_, err := d.mysql.Exec(ctx, query, userID, targetID, convType)
	return err
}

// ClearUnread clears unread count for a conversation
func (d *ConversationDAO) ClearUnread(ctx context.Context, userID, targetID int64, convType int8) error {
	query := `
		UPDATE conversations
		SET unread_count = 0, updated_at = NOW()
		WHERE user_id = ? AND target_id = ? AND conversation_type = ?
	`
	_, err := d.mysql.Exec(ctx, query, userID, targetID, convType)
	return err
}

// scanConversation scans a conversation from a row
func (d *ConversationDAO) scanConversation(scanner interface{ Scan(...interface{}) error }) (*model.Conversation, error) {
	var conv model.Conversation
	err := scanner.Scan(
		&conv.ID, &conv.UserID, &conv.TargetID, &conv.ConversationType,
		&conv.UnreadCount, &conv.LastMsgID, &conv.LastMsgContent,
		&conv.LastMsgTime, &conv.IsPinned, &conv.IsMuted, &conv.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

// scanConversationWithDetails scans a conversation with user/group details
func (d *ConversationDAO) scanConversationWithDetails(rows *sql.Rows) (*model.Conversation, error) {
	var conv model.Conversation
	var user model.User
	var group model.Group

	// Scan all columns
	err := rows.Scan(
		&conv.ID, &conv.UserID, &conv.TargetID, &conv.ConversationType,
		&conv.UnreadCount, &conv.LastMsgID, &conv.LastMsgContent,
		&conv.LastMsgTime, &conv.IsPinned, &conv.IsMuted, &conv.UpdatedAt,
		// User fields
		&user.ID, &user.Username, &user.PasswordHash, &user.Nickname,
		&user.AvatarURL, &user.Status, &user.Signature,
		&user.CreatedAt, &user.UpdatedAt,
		// Group fields
		&group.ID, &group.GroupNo, &group.Name, &group.AvatarURL,
		&group.OwnerID, &group.MaxMembers, &group.JoinType, &group.MuteAll,
		&group.CreatedAt, &group.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Set TargetUser for single chat
	if conv.ConversationType == 1 && user.ID > 0 {
		user.PasswordHash = "" // Don't return password hash
		conv.TargetUser = &user
	}

	// Set TargetGroup for group chat
	if conv.ConversationType == 2 && group.ID > 0 {
		conv.TargetGroup = &group
	}

	return &conv, nil
}

// generateMsgID generates a unique message ID
func generateMsgID() string {
	return uuid.New().String()
}

// GetConversationPairID generates a unique pair ID for single chat conversation
// Ensures that two users always have the same conversation ID regardless of who initiates
func GetConversationPairID(userID1, userID2 int64) int64 {
	if userID1 > userID2 {
		userID1, userID2 = userID2, userID1
	}
	return userID1*1000000000 + userID2
}

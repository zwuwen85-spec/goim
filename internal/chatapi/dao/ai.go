package dao

import (
	"context"
	"database/sql"

	"github.com/Terry-Mao/goim/internal/chatapi/model"
)

// AIDAO handles AI-related database operations
type AIDAO struct {
	mysql *MySQL
}

// NewAIDAO creates a new AIDAO
func NewAIDAO(mysql *MySQL) *AIDAO {
	return &AIDAO{mysql: mysql}
}

// CreateBot creates a new AI bot
func (d *AIDAO) CreateBot(ctx context.Context, bot *model.AIBot) error {
	query := `INSERT INTO ai_bots (bot_id, user_id, name, personality, model_name, temperature, max_tokens)
		VALUES (?, ?, ?, ?, ?, ?, ?)`
	result, err := d.mysql.Exec(ctx, query,
		bot.BotID, bot.UserID, bot.Name, bot.Personality, bot.ModelName, bot.Temperature, bot.MaxTokens,
	)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	bot.ID = id
	return nil
}

// FindBotByID finds a bot by ID
func (d *AIDAO) FindBotByID(ctx context.Context, botID int64) (*model.AIBot, error) {
	query := `SELECT id, bot_id, user_id, name, personality, model_name, temperature, max_tokens, created_at, updated_at
		FROM ai_bots WHERE bot_id = ?`
	return d.scanBot(d.mysql.QueryRow(ctx, query, botID))
}

// FindBotsByUser finds all bots for a user
func (d *AIDAO) FindBotsByUser(ctx context.Context, userID int64) ([]*model.AIBot, error) {
	query := `SELECT id, bot_id, user_id, name, personality, model_name, temperature, max_tokens, created_at, updated_at
		FROM ai_bots WHERE user_id = ? ORDER BY created_at DESC`
	rows, err := d.mysql.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bots []*model.AIBot
	for rows.Next() {
		bot, err := d.scanBotRows(rows)
		if err != nil {
			return nil, err
		}
		bots = append(bots, bot)
	}
	return bots, nil
}

// UpdateBot updates a bot
func (d *AIDAO) UpdateBot(ctx context.Context, bot *model.AIBot) error {
	query := `UPDATE ai_bots SET name = ?, personality = ?, model_name = ?, temperature = ?, max_tokens = ? WHERE id = ?`
	_, err := d.mysql.Exec(ctx, query,
		bot.Name, bot.Personality, bot.ModelName, bot.Temperature, bot.MaxTokens, bot.ID,
	)
	return err
}

// DeleteBot deletes a bot
func (d *AIDAO) DeleteBot(ctx context.Context, id int64) error {
	query := `DELETE FROM ai_bots WHERE id = ?`
	_, err := d.mysql.Exec(ctx, query, id)
	return err
}

// CreateConversation creates an AI conversation
func (d *AIDAO) CreateConversation(ctx context.Context, conv *model.AIConversation) error {
	query := `INSERT INTO ai_conversations (user_id, bot_id, title) VALUES (?, ?, ?)`
	result, err := d.mysql.Exec(ctx, query, conv.UserID, conv.BotID, conv.Title)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	conv.ID = id
	return nil
}

// FindConversationsByUser finds all conversations for a user
func (d *AIDAO) FindConversationsByUser(ctx context.Context, userID int64) ([]*model.AIConversation, error) {
	query := `SELECT id, user_id, bot_id, title, created_at, updated_at
		FROM ai_conversations WHERE user_id = ? ORDER BY updated_at DESC`
	rows, err := d.mysql.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var convs []*model.AIConversation
	for rows.Next() {
		conv, err := d.scanConversationRows(rows)
		if err != nil {
			return nil, err
		}
		convs = append(convs, conv)
	}
	return convs, nil
}

// scanBot scans a bot from a row
func (d *AIDAO) scanBot(scanner interface{ Scan(...interface{}) error }) (*model.AIBot, error) {
	var bot model.AIBot
	err := scanner.Scan(
		&bot.ID, &bot.BotID, &bot.UserID, &bot.Name, &bot.Personality,
		&bot.ModelName, &bot.Temperature, &bot.MaxTokens, &bot.CreatedAt, &bot.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &bot, nil
}

// scanBotRows scans a bot from rows
func (d *AIDAO) scanBotRows(rows *sql.Rows) (*model.AIBot, error) {
	var bot model.AIBot
	err := rows.Scan(
		&bot.ID, &bot.BotID, &bot.UserID, &bot.Name, &bot.Personality,
		&bot.ModelName, &bot.Temperature, &bot.MaxTokens, &bot.CreatedAt, &bot.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &bot, nil
}

// scanConversationRows scans a conversation from rows
func (d *AIDAO) scanConversationRows(rows *sql.Rows) (*model.AIConversation, error) {
	var conv model.AIConversation
	err := rows.Scan(
		&conv.ID, &conv.UserID, &conv.BotID, &conv.Title, &conv.CreatedAt, &conv.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

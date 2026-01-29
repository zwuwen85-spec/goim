package ai

import (
	"context"
	"fmt"
	"sync"

	"github.com/Terry-Mao/goim/internal/ai/conf"
)

// Bot represents an AI bot
type Bot struct {
	ID         int64
	UserID     int64
	Name       string
	Personality *Personality
	Model      string
	Temperature float64
}

// BotManager manages AI bots
type BotManager struct {
	service Service
	bots    map[int64]*Bot
	mu      sync.RWMutex
}

// NewBotManager creates a new bot manager
func NewBotManager(cfg *conf.AI) *BotManager {
	var service Service
	switch cfg.Provider {
	case "openai":
		service = NewOpenAI(cfg)
	default:
		service = NewOpenAI(cfg) // Default to OpenAI
	}

	return &BotManager{
		service: service,
		bots:    make(map[int64]*Bot),
	}
}

// RegisterBot registers a new AI bot
func (bm *BotManager) RegisterBot(bot *Bot) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	bm.bots[bot.ID] = bot
	return nil
}

// GetBot gets a bot by ID
func (bm *BotManager) GetBot(botID int64) (*Bot, error) {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	bot, ok := bm.bots[botID]
	if !ok {
		return nil, fmt.Errorf("bot not found: %d", botID)
	}
	return bot, nil
}

// Chat sends a message to a bot and returns the response
func (bm *BotManager) Chat(ctx context.Context, botID int64, history []Message, userMessage string) (string, error) {
	bot, err := bm.GetBot(botID)
	if err != nil {
		return "", err
	}

	return bm.service.Chat(ctx, botID, bot.Personality, history, userMessage)
}

// MultimodalChat sends a message with images to a bot and returns the response
func (bm *BotManager) MultimodalChat(ctx context.Context, botID int64, history []Message, userMessage string, imageUrls []string) (string, error) {
	bot, err := bm.GetBot(botID)
	if err != nil {
		return "", err
	}

	return bm.service.MultimodalChat(ctx, botID, bot.Personality, history, userMessage, imageUrls)
}

// StreamChat sends a message to a bot and streams the response via callback
func (bm *BotManager) StreamChat(ctx context.Context, botID int64, history []Message, userMessage string, callback func(chunk string)) error {
	bot, err := bm.GetBot(botID)
	if err != nil {
		return err
	}

	return bm.service.StreamChat(ctx, botID, bot.Personality, history, userMessage, callback)
}

// StreamMultimodalChat sends a message with images to a bot and streams the response via callback
func (bm *BotManager) StreamMultimodalChat(ctx context.Context, botID int64, history []Message, userMessage string, imageUrls []string, callback func(chunk string)) error {
	bot, err := bm.GetBot(botID)
	if err != nil {
		return err
	}

	return bm.service.StreamMultimodalChat(ctx, botID, bot.Personality, history, userMessage, imageUrls, callback)
}

// SetService sets the AI service
func (bm *BotManager) SetService(service Service) {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	bm.service = service
}

// GetService returns the AI service
func (bm *BotManager) GetService() Service {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	return bm.service
}

// CreateBotFromDefault creates a bot from a default personality
func CreateBotFromDefault(botID int64, userID int64, personalityType string) (*Bot, error) {
	personalities := DefaultPersonalities()
	personality, ok := personalities[personalityType]
	if !ok {
		return nil, fmt.Errorf("unknown personality type: %s", personalityType)
	}

	return &Bot{
		ID:         botID,
		UserID:     userID,
		Name:       personality.Name,
		Personality: personality,
		Model:      "gpt-3.5-turbo",
		Temperature: 0.7,
	}, nil
}

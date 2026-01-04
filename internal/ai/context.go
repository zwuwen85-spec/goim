package ai

import (
	"fmt"
	"sync"
	"time"
)

// ConversationContext holds the conversation history for a user with a bot
type ConversationContext struct {
	BotID      int64
	UserID     int64
	Messages   []Message
	LastActive time.Time
	mu         sync.RWMutex
}

// ContextManager manages AI conversation contexts
type ContextManager struct {
	contexts map[string]*ConversationContext // key: botID:userID
	mu       sync.RWMutex
	ttl      time.Duration // Time to live for inactive conversations
}

// NewContextManager creates a new context manager
func NewContextManager(ttl time.Duration) *ContextManager {
	cm := &ContextManager{
		contexts: make(map[string]*ConversationContext),
		ttl:      ttl,
	}

	// Start cleanup goroutine
	go cm.cleanup()

	return cm
}

// GetContext gets or creates a conversation context
func (cm *ContextManager) GetContext(botID, userID int64) *ConversationContext {
	key := cm.key(botID, userID)

	cm.mu.RLock()
	ctx, ok := cm.contexts[key]
	cm.mu.RUnlock()

	if !ok {
		cm.mu.Lock()
		// Double-check
		ctx, ok = cm.contexts[key]
		if !ok {
			ctx = &ConversationContext{
				BotID:      botID,
				UserID:     userID,
				Messages:   make([]Message, 0, 20),
				LastActive: time.Now(),
			}
			cm.contexts[key] = ctx
		}
		cm.mu.Unlock()
	}

	return ctx
}

// AddMessage adds a message to the conversation
func (cc *ConversationContext) AddMessage(role, content string) {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	cc.Messages = append(cc.Messages, Message{
		Role:    role,
		Content: content,
	})

	// Limit history to last 20 messages
	if len(cc.Messages) > 20 {
		cc.Messages = cc.Messages[len(cc.Messages)-20:]
	}

	cc.LastActive = time.Now()
}

// GetMessages returns the message history
func (cc *ConversationContext) GetMessages() []Message {
	cc.mu.RLock()
	defer cc.mu.RUnlock()

	// Return a copy to avoid race conditions
	msgs := make([]Message, len(cc.Messages))
	copy(msgs, cc.Messages)
	return msgs
}

// Clear clears the conversation history
func (cc *ConversationContext) Clear() {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	cc.Messages = make([]Message, 0, 20)
	cc.LastActive = time.Now()
}

// ClearContext clears a conversation context
func (cm *ContextManager) ClearContext(botID, userID int64) {
	key := cm.key(botID, userID)

	cm.mu.Lock()
	defer cm.mu.Unlock()

	delete(cm.contexts, key)
}

// cleanup removes inactive contexts
func (cm *ContextManager) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		cm.mu.Lock()
		now := time.Now()
		for key, ctx := range cm.contexts {
			if now.Sub(ctx.LastActive) > cm.ttl {
				delete(cm.contexts, key)
			}
		}
		cm.mu.Unlock()
	}
}

// key generates a unique key for a bot-user pair
func (cm *ContextManager) key(botID, userID int64) string {
	return fmt.Sprintf("%d:%d", botID, userID)
}

// GetRecentMessages returns the last n messages
func (cc *ConversationContext) GetRecentMessages(n int) []Message {
	cc.mu.RLock()
	defer cc.mu.RUnlock()

	messages := cc.Messages
	if len(messages) > n {
		messages = messages[len(messages)-n:]
	}

	// Return a copy
	msgs := make([]Message, len(messages))
	copy(msgs, messages)
	return msgs
}

// SetMessages sets the conversation history (for loading from storage)
func (cc *ConversationContext) SetMessages(messages []Message) {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	cc.Messages = make([]Message, len(messages))
	copy(cc.Messages, messages)
	cc.LastActive = time.Now()
}

package ai

import (
	"context"
	"encoding/json"
)

// Personality defines the personality traits for an AI bot
type Personality struct {
	Name        string   `json:"name"`
	Tone        string   `json:"tone"`        // friendly, professional, humorous, etc.
	Role        string   `json:"role"`        // assistant, tutor, companion, etc.
	Traits      []string `json:"traits"`      // personality traits
	SystemPrompt string  `json:"system_prompt"` // custom system prompt
}

// Message represents a message in the conversation
type Message struct {
	Role    string `json:"role"`    // user, assistant, system
	Content string `json:"content"`
}

// Service is the AI service interface
type Service interface {
	// Chat sends a message to the AI and returns the response
	Chat(ctx context.Context, botID int64, personality *Personality, history []Message, userMessage string) (string, error)

	// StreamChat streams the AI response
	StreamChat(ctx context.Context, botID int64, personality *Personality, history []Message, userMessage string, callback func(chunk string)) error

	// SetModel sets the AI model to use
	SetModel(model string)

	// SetTemperature sets the temperature for responses (0.0 - 1.0)
	SetTemperature(temp float64)
}

// Config holds the AI service configuration
type Config struct {
	Provider     string  `json:"provider"`      // openai, anthropic, local
	APIKey       string  `json:"api_key"`
	Model        string  `json:"model"`         // gpt-3.5-turbo, gpt-4, etc.
	Temperature  float64 `json:"temperature"`
	MaxTokens    int     `json:"max_tokens"`
	BaseURL      string  `json:"base_url"`      // for custom endpoints
}

// DefaultConfig returns the default AI configuration
func DefaultConfig() *Config {
	return &Config{
		Provider:    "openai",
		Model:       "gpt-3.5-turbo",
		Temperature: 0.7,
		MaxTokens:   1000,
	}
}

// DefaultPersonalities returns predefined AI personalities
func DefaultPersonalities() map[string]*Personality {
	return map[string]*Personality{
		"assistant": {
			Name: "智能助手",
			Tone: "friendly",
			Role: "assistant",
			Traits: []string{"helpful", "knowledgeable", "polite"},
			SystemPrompt: "You are a helpful AI assistant. Answer questions clearly and concisely.",
		},
		"companion": {
			Name: "聊天伙伴",
			Tone: "casual",
			Role: "companion",
			Traits: []string{"friendly", "empathetic", "fun"},
			SystemPrompt: "You are a friendly chat companion. Be engaging and supportive in casual conversation.",
		},
		"tutor": {
			Name: "学习导师",
			Tone: "professional",
			Role: "tutor",
			Traits: []string{"knowledgeable", "patient", "encouraging"},
			SystemPrompt: "You are a patient tutor. Explain concepts clearly and encourage learning.",
		},
		"creative": {
			Name: "创意助手",
			Tone: "imaginative",
			Role: "creative",
			Traits: []string{"creative", "inspiring", "original"},
			SystemPrompt: "You are a creative assistant. Help with brainstorming and creative thinking.",
		},
	}
}

// BuildSystemPrompt builds a system prompt from personality config
func BuildSystemPrompt(personality *Personality) string {
	if personality.SystemPrompt != "" {
		return personality.SystemPrompt
	}

	prompt := "You are a " + personality.Role + " with a " + personality.Tone + " tone.\n"
	if len(personality.Traits) > 0 {
		prompt += "Your personality traits are: "
		for i, trait := range personality.Traits {
			if i > 0 {
				prompt += ", "
			}
			prompt += trait
		}
		prompt += ".\n"
	}
	return prompt
}

// ParsePersonality parses personality from JSON string
func ParsePersonality(data string) (*Personality, error) {
	var p Personality
	if err := json.Unmarshal([]byte(data), &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// PersonalityToString converts personality to JSON string
func PersonalityToString(p *Personality) (string, error) {
	data, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
